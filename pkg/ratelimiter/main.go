package ratelimiter

import (
	"time"
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
	c.TotalRequests = 1
}

func (c *ClientLimiter) ShouldBlock() bool {
	return c.TotalRequests > c.Limiter.RequestsPerSecond
}

func (c *ClientLimiter) Block() {
	c.Blocked = true
	c.BlockedAt = time.Now()
}
