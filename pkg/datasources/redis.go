package datasources

import (
	"context"
	"encoding/json"
	"time"

	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
	"github.com/redis/go-redis/v9"
)

type RedisDatasource struct {
	client *redis.Client
}

func NewRedisDatasource(client *redis.Client) *RedisDatasource {
	return &RedisDatasource{client: client}
}

func (d *RedisDatasource) Add(key string, data *ratelimiter.ClientRateLimiter) {
	d.client.Set(context.Background(), key, data, 60*time.Minute)
}

func (d *RedisDatasource) Get(key string) *ratelimiter.ClientRateLimiter {
	data, err := d.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	var client *ratelimiter.ClientRateLimiter
	json.Unmarshal([]byte(data), &client)

	return client
}

func (d *RedisDatasource) Has(key string) bool {
	_, err := d.client.Get(context.Background(), key).Result()
	return err == nil
}

func (d *RedisDatasource) All() map[string]*ratelimiter.ClientRateLimiter {
	keys, err := d.client.Keys(context.Background(), "*").Result()
	if err != nil {
		return nil
	}

	clients := make(map[string]*ratelimiter.ClientRateLimiter)

	for _, key := range keys {
		client := d.Get(key)
		clients[key] = client
	}

	return clients
}
