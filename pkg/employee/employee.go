package employee

type Employee struct {
	ID      string `json:"_id"`
	name    string `json:"name"`
	age     int    `json:"age"`
	DocType string `json:"doc_type"`
}
