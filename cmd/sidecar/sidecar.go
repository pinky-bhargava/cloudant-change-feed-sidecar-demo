package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"sidecar-demo/internal/app/sidecar/redis"
	pkgdocument "sidecar-demo/pkg/employee"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/IBM/go-sdk-core/v5/core"
	"golang.org/x/net/context"
)

//
// this done is a loop to give us recovery from fatal errors in the db listenign loop
//
func main() {
	ctx := context.Background()

	// Main loop
	fmt.Println("\n Starting  sidecar service ...")
	var mainLoopErrCount = 0 // used to keep track of how many continous errors mainLoopBody has hit
	var since = "0"
	for {
		if !mainLoopBody(&mainLoopErrCount, &since, ctx) {
			msg := fmt.Sprintf("mainLoopBody bailing out after %d failures\n", mainLoopErrCount)
			fmt.Println(msg)
			break
		}
		time.Sleep(5 * time.Second) // make sure we don't start hammering something too hard in a failure loop
	}
}

const ErrLimit = 0 // how many continous errors until we give up

//
// This is the main loop of the DB listeners.  It's done as it's ownb function so we can
// In: *err_count - pointer to variable holding continous error count.  This value is set to zero
//
func mainLoopBody(err_count *int, since *string, ctx context.Context) (keep_going bool) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("Resolver main loop body failed.  err_count: %d Err: %v", *err_count, err)
			if *err_count > ErrLimit {
				keep_going = false // we've err'd too many times - tell caller time to bail out of loop
			}
		}
	}()

	keep_going = true
	*err_count++
	cloudantEndpoint := "https://" + os.Getenv("CLOUDANT_HOST")
	cloudantDB := os.Getenv("CLOUDANT_DB_NAME")
	fmt.Printf("get Cloudant client for db %s:%s", cloudantDB, cloudantEndpoint)
	client, err := getCloudantClient(os.Getenv("CLOUDANT_APIKEY"), cloudantEndpoint, cloudantDB, ctx)
	if err != nil {
		fmt.Errorf("\n unable to connect to cloudant using  iam authenticator  : %v: ", err)
		return
	}
	fmt.Println("Set cloudant PostChanges option ...")
	// Watch for new/updated zones
	postChangesOptions := client.NewPostChangesOptions(cloudantDB)
	postChangesOptions.SetFeed("longpoll")
	selector := map[string]interface{}{
		"doc_type": map[string]string{
			"$eq": "employee_info",
		},
	}

	postChangesOptions.SetSelector(selector)
	postChangesOptions.SetFilter("_selector")
	postChangesOptions.SetIncludeDocs(true)
	postChangesOptions.SetSince(*since)
	//postChangesOptions.SetTimeout(int64(timeout)) // 20 seconds

	for {

		changesResult, response, err := client.PostChanges(postChangesOptions)
		fmt.Printf("response %v: ", response)
		fmt.Printf("changesResult : %v", changesResult)
		if err != nil {
			fmt.Printf("failed to read changes from cloudant : %v", err)
			timeoutErrMsg := "Client.Timeout exceeded while awaiting headers"
			if strings.Contains(err.Error(), timeoutErrMsg) {
				fmt.Println("resetting the cloudant connection")
			}
			fmt.Printf("set since:%s", *since)
			return
		}

		updatedEmployees := map[string]pkgdocument.Employee{}
		for _, r := range changesResult.Results {
			employee := pkgdocument.Employee{}
			b, err := json.MarshalIndent(r.Doc, "", "  ")
			if err != nil {
				fmt.Errorf("Failed to Marshal the document%v:", err)
				continue
			}
			err = json.Unmarshal(b, &employee)
			if err != nil {
				fmt.Errorf("Failed to unmarshal the document %v:", err)
				continue
			}

			updatedEmployees[employee.ID] = employee
		} //for
		for employee := range updatedEmployees {
			delete_err := deleteCache(employee, ctx)

			if delete_err != nil {
				fmt.Errorf("\n Delete error for employee id  [%s] : %v", employee, delete_err)
			} else {
				fmt.Printf("\n Deleted requestor-ip-to-zone-link  [%s] from cache successfully ", employee)
			}
		}
		*since = *changesResult.LastSeq
		postChangesOptions.SetSince(*since)
		*err_count = 0 // if we get here, we've run succesfully, so reset the error count
	}
}
func deleteCache(key string, ctx context.Context) error {
	redisClient, err := redis.NewRedisClient("", ctx)
	if err != nil {
		fmt.Printf("Error when getting redis client : %v", err)
		return err
	}
	err = redisClient.Delete(key)
	return err
}

func getCloudantClient(apiKey, cloudantEndpoint, cloudantDB string, ctx context.Context) (cloudantClient *cloudantv1.CloudantV1, err error) {
	fmt.Println("Inside getCloudantClient")
	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
	}
	cloudantClient, err = cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			URL:           cloudantEndpoint,
			ServiceName:   "sidecar-demo",
			Authenticator: authenticator,
		},
	)

	if err != nil {
		fmt.Printf("error getting cloudant client %v", err)
		return nil, err
	}

	cloudantClient.SetServiceURL(cloudantEndpoint)

	// Try to connect to CouchDB server
	connectAttempts := 3
	for i := 0; i < connectAttempts; i++ {

		_, _, err = cloudantClient.GetDatabaseInformation(
			cloudantClient.NewGetDatabaseInformationOptions(cloudantDB),
		)
		if err != nil {
			fmt.Printf("failed to connect to cloudant.retrying in %v seconds ,Error - %v", time.Second, err)
			// Failed to connect to DB, maybe it's still starting
			time.Sleep(time.Second)
			continue
		}
		// Successfully connected to the DB, stop trying..
		break
	}

	return cloudantClient, err
}
