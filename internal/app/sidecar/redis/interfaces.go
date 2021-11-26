package redis

type RedisPrivateClient interface {
	Set(key string, value []byte) error
	Delete(key string) error
	Get(key string) ([]byte, error)
}
