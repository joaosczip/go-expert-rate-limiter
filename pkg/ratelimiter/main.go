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
	RequestsPerSecond int
	BlockUserFor      time.Duration
}

func NewRateLimiter(rps int, blockDuration time.Duration) *RateLimiter {
	return &RateLimiter{RequestsPerSecond: rps, BlockUserFor: blockDuration}
}

type RateLimiterManager struct {
	Clients map[string]*ClientLimiter
}

type RateLimiterConfig struct {
	RequestesPerSecond int
	BlockUserFor       time.Duration
}

func NewRateLimiterManager() *RateLimiterManager {
	manager := &RateLimiterManager{Clients: make(map[string]*ClientLimiter)}

	go manager.clearRequests()

	return manager
}

func (r *RateLimiterManager) clearRequests() {
	for {
		time.Sleep(1 * time.Second)
		for _, client := range r.Clients {
			client.clearRequests()
		}
	}
}

func (r *RateLimiterManager) HandleRequest(ip string, config RateLimiterConfig) error {
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

type ClientLimiter struct {
	Limiter       *RateLimiter
	LastSeen      time.Time
	Blocked       bool
	BlockedAt     time.Time
	TotalRequests int
}

func NewClientLimiter(rps int, blockDuration time.Duration) *ClientLimiter {
	return &ClientLimiter{Limiter: NewRateLimiter(rps, blockDuration)}
}

func (c *ClientLimiter) clearRequests() {
	if c.TotalRequests <= c.Limiter.RequestsPerSecond {
		c.TotalRequests = 0
	}
}

func (c *ClientLimiter) IsBlocked() bool {
	return c.Blocked
}

func (c *ClientLimiter) HasBlockingExpired() bool {
	return time.Since(c.BlockedAt) > c.Limiter.BlockUserFor
}

func (c *ClientLimiter) ResetBlock() {
	c.Blocked = false
	c.TotalRequests = 0
}

func (c *ClientLimiter) ShouldBlock() bool {
	return c.TotalRequests > c.Limiter.RequestsPerSecond
}

func (c *ClientLimiter) Block() {
	c.Blocked = true
	c.BlockedAt = time.Now()
}
