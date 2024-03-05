package datasources

import (
	"sync"

	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
)

type InMemoryDatasource struct {
	clients map[string]*ratelimiter.ClientRateLimiter
	mux     sync.Mutex
}

func NewInMemoryDatasource() *InMemoryDatasource {
	return &InMemoryDatasource{clients: make(map[string]*ratelimiter.ClientRateLimiter), mux: sync.Mutex{}}
}

func (d *InMemoryDatasource) Add(key string, data *ratelimiter.ClientRateLimiter) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.clients[key] = data
}

func (d *InMemoryDatasource) Get(key string) *ratelimiter.ClientRateLimiter {
	if data, found := d.clients[key]; found {
		return data
	}
	return nil
}

func (d *InMemoryDatasource) Has(key string) bool {
	_, found := d.clients[key]
	return found
}

func (d *InMemoryDatasource) All() map[string]*ratelimiter.ClientRateLimiter {
	return d.clients
}
