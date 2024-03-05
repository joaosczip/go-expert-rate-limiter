package ratelimiter

import (
	"errors"
	"sync"
	"time"
)

var (
	errMaxRequests = errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
)

type Datasource interface {
	Add(key string, data *ClientRateLimiter)
	Get(key string) *ClientRateLimiter
	Has(key string) bool
	All() map[string]*ClientRateLimiter
}

type InMemoryDatasource struct {
	clients map[string]*ClientRateLimiter
	mux     sync.Mutex
}

func NewInMemoryDatasource() *InMemoryDatasource {
	return &InMemoryDatasource{clients: make(map[string]*ClientRateLimiter), mux: sync.Mutex{}}
}

func (d *InMemoryDatasource) Add(key string, data *ClientRateLimiter) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.clients[key] = data
}

func (d *InMemoryDatasource) Get(key string) *ClientRateLimiter {
	if data, found := d.clients[key]; found {
		return data
	}
	return nil
}

func (d *InMemoryDatasource) Has(key string) bool {
	_, found := d.clients[key]
	return found
}

func (d *InMemoryDatasource) All() map[string]*ClientRateLimiter {
	return d.clients
}

type RateLimiter struct {
	datasource Datasource
}

type RateLimiterConfigByIP struct {
	RequestesPerSecond int
	BlockUserFor       time.Duration
}

type RateLimiterConfigByToken struct {
	RequestesPerSecond int
	BlockUserFor       time.Duration
	Key                string
}

type RateLimiterConfig struct {
	ConfigByIP    *RateLimiterConfigByIP
	ConfigByToken *RateLimiterConfigByToken
}

func NewRateLimiter(datasource Datasource) *RateLimiter {
	limiter := &RateLimiter{datasource: datasource}

	go limiter.clearRequests()

	return limiter
}

func (r *RateLimiter) clearRequests() {
	for {
		time.Sleep(1 * time.Second)
		for _, client := range r.datasource.All() {
			client.clearRequests()
		}
	}
}

func (r *RateLimiter) getClient(ip, token string, config RateLimiterConfig) *ClientRateLimiter {
	ipConfig, tokenConfig := config.ConfigByIP, config.ConfigByToken

	var client *ClientRateLimiter

	if found := r.datasource.Has(ip); !found {
		client := newClientLimiter(ipConfig.RequestesPerSecond, ipConfig.BlockUserFor)
		r.datasource.Add(ip, client)
	}

	client = r.datasource.Get(ip)

	if token != "" && config.ConfigByToken != nil {
		if found := r.datasource.Has(token); !found {
			tokenClient := newClientLimiter(tokenConfig.RequestesPerSecond, tokenConfig.BlockUserFor)
			r.datasource.Add(token, tokenClient)
		}
		client = r.datasource.Get(token)
	}

	return client
}

func (r *RateLimiter) HandleRequest(ip, token string, config RateLimiterConfig) error {
	client := r.getClient(ip, token, config)

	if err := client.verifyAndBlockUser(); err != nil {
		return err
	}

	return nil
}

type ClientRateLimiter struct {
	requestsPerSecond int
	blockUserFor      time.Duration
	blocked           bool
	blockedAt         time.Time
	totalRequests     int
	mux               sync.Mutex
}

func newClientLimiter(rps int, blockDuration time.Duration) *ClientRateLimiter {
	return &ClientRateLimiter{
		requestsPerSecond: rps,
		blockUserFor:      blockDuration,
		mux:               sync.Mutex{},
	}
}

func (c *ClientRateLimiter) verifyAndBlockUser() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.isBlocked() {
		if c.hasBlockingExpired() {
			c.resetBlock()
		} else {
			return errMaxRequests
		}
	}

	c.totalRequests += 1

	if c.shouldBlock() {
		c.block()
		return errMaxRequests
	}

	return nil
}

func (c *ClientRateLimiter) clearRequests() {
	if c.totalRequests <= c.requestsPerSecond {
		c.totalRequests = 0
	}
}

func (c *ClientRateLimiter) isBlocked() bool {
	return c.blocked
}

func (c *ClientRateLimiter) hasBlockingExpired() bool {
	return time.Since(c.blockedAt) > c.blockUserFor
}

func (c *ClientRateLimiter) resetBlock() {
	c.blocked = false
	c.totalRequests = 0
}

func (c *ClientRateLimiter) shouldBlock() bool {
	return c.totalRequests > c.requestsPerSecond
}

func (c *ClientRateLimiter) block() {
	c.blocked = true
	c.blockedAt = time.Now()
}
