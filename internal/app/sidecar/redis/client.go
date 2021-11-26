package redis

import (
	"context"
)

func NewRedisClient(endpoint string, ctx context.Context) (*RedisPrivate, error) {
	if endpoint == "" {
		endpoint = ":6379"
	}
	pool := newPool(endpoint)
	re := RedisPrivate{
		Endpoint: endpoint,
		Pool:     pool,
		PoolSize: 500,
	}

	return &re, nil
}
