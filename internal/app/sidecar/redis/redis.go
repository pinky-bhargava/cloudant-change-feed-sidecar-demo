package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

type RedisPrivate struct {
	Endpoint string
	Pool     *redis.Pool
	PoolSize int
	// Should be a value in bytes
	MaxStringValueLength int
}

/*func init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = ":6379"
	}
	Pool = newPool(redisHost)
	//cleanupHook()
}*/

func newPool(server string) *redis.Pool {

	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

/*
func (r *RedisPrivate) Ping() error {

	conn := r.Pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		return fmt.Errorf("cannot 'PING' db: %v", err)
	}
	return nil
}
func (r *RedisPrivate) Exists(key string) (bool, error) {
	conn := r.Pool.Get()
	defer conn.Close()
	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}
	return ok, err
}
func (r *RedisPrivate) GetKeys(pattern string) ([]string, error) {
	conn := r.Pool.Get()
	defer conn.Close()
	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)
		if iter == 0 {
			break
		}
	}
	return keys, nil
}

func (r *RedisPrivate) Incr(counterKey string) (int, error) {
	conn := r.Pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("INCR", counterKey))
}
*/

func (r *RedisPrivate) Get(key string) ([]byte, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	var data []byte
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		fmt.Errorf("error getting key %s: %v", key, err)
		return data, err
	}
	return data, nil
}

func (r *RedisPrivate) Set(key string, value []byte) error {

	conn := r.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}
	return err
}

func (r *RedisPrivate) Delete(key string) error {
	conn := r.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}
