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
	Clients map[string]*ClientRateLimiter
}

type RateLimiterConfig struct {
	RequestesPerSecond int
	BlockUserFor       time.Duration
}

func NewRateLimiter() *RateLimiter {
	manager := &RateLimiter{Clients: make(map[string]*ClientRateLimiter)}

	go manager.clearRequests()

	return manager
}

func (r *RateLimiter) clearRequests() {
	for {
		time.Sleep(1 * time.Second)
		for _, client := range r.Clients {
			client.clearRequests()
		}
	}
}

func (r *RateLimiter) HandleRequest(ip string, config RateLimiterConfig) error {
	clients := r.Clients
	mux := sync.Mutex{}

	if _, found := clients[ip]; !found {
		clients[ip] = NewClientLimiter(config.RequestesPerSecond, config.BlockUserFor)
	}

	mux.Lock()
	defer mux.Unlock()
	clients[ip].LastSeen = time.Now()

	if clients[ip].IsBlocked() {
		if clients[ip].HasBlockingExpired() {
			clients[ip].ResetBlock()
		} else {
			return errMaxRequests
		}
	}

	clients[ip].TotalRequests += 1

	if clients[ip].ShouldBlock() {
		clients[ip].Block()
		return errMaxRequests
	}

	return nil
}

type ClientRateLimiter struct {
	RequestsPerSecond int
	BlockUserFor      time.Duration
	LastSeen          time.Time
	Blocked           bool
	BlockedAt         time.Time
	TotalRequests     int
}

func NewClientLimiter(rps int, blockDuration time.Duration) *ClientRateLimiter {
	return &ClientRateLimiter{
		RequestsPerSecond: rps,
		BlockUserFor:      blockDuration,
	}
}

func (c *ClientRateLimiter) clearRequests() {
	if c.TotalRequests <= c.RequestsPerSecond {
		c.TotalRequests = 0
	}
}

func (c *ClientRateLimiter) IsBlocked() bool {
	return c.Blocked
}

func (c *ClientRateLimiter) HasBlockingExpired() bool {
	return time.Since(c.BlockedAt) > c.BlockUserFor
}

func (c *ClientRateLimiter) ResetBlock() {
	c.Blocked = false
	c.TotalRequests = 0
}

func (c *ClientRateLimiter) ShouldBlock() bool {
	return c.TotalRequests > c.RequestsPerSecond
}

func (c *ClientRateLimiter) Block() {
	c.Blocked = true
	c.BlockedAt = time.Now()
}
