package ratelimiter

import (
	"errors"
	"sync"
	"time"
)

var (
	errMaxRequests = errors.New("you have reached the maximum number of requests or actions allowed within a certain time frame")
)

type RateLimiter struct {
	clients map[string]*clientRateLimiter
}

type RateLimiterConfig struct {
	RequestesPerSecond int
	BlockUserFor       time.Duration
}

func NewRateLimiter() *RateLimiter {
	manager := &RateLimiter{clients: make(map[string]*clientRateLimiter)}

	go manager.clearRequests()

	return manager
}

func (r *RateLimiter) clearRequests() {
	for {
		time.Sleep(1 * time.Second)
		for _, client := range r.clients {
			client.clearRequests()
		}
	}
}

func (r *RateLimiter) HandleRequest(ip string, config RateLimiterConfig) error {
	clients := r.clients
	mux := sync.Mutex{}

	if _, found := clients[ip]; !found {
		clients[ip] = newClientLimiter(config.RequestesPerSecond, config.BlockUserFor)
	}

	mux.Lock()
	defer mux.Unlock()

	if clients[ip].isBlocked() {
		if clients[ip].hasBlockingExpired() {
			clients[ip].resetBlock()
		} else {
			return errMaxRequests
		}
	}

	clients[ip].totalRequests += 1

	if clients[ip].shouldBlock() {
		clients[ip].block()
		return errMaxRequests
	}

	return nil
}

type clientRateLimiter struct {
	requestsPerSecond int
	blockUserFor      time.Duration
	blocked           bool
	blockedAt         time.Time
	totalRequests     int
}

func newClientLimiter(rps int, blockDuration time.Duration) *clientRateLimiter {
	return &clientRateLimiter{
		requestsPerSecond: rps,
		blockUserFor:      blockDuration,
	}
}

func (c *clientRateLimiter) clearRequests() {
	if c.totalRequests <= c.requestsPerSecond {
		c.totalRequests = 0
	}
}

func (c *clientRateLimiter) isBlocked() bool {
	return c.blocked
}

func (c *clientRateLimiter) hasBlockingExpired() bool {
	return time.Since(c.blockedAt) > c.blockUserFor
}

func (c *clientRateLimiter) resetBlock() {
	c.blocked = false
	c.totalRequests = 0
}

func (c *clientRateLimiter) shouldBlock() bool {
	return c.totalRequests > c.requestsPerSecond
}

func (c *clientRateLimiter) block() {
	c.blocked = true
	c.blockedAt = time.Now()
}
