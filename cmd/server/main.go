package main

import (
	"net/http"
	"time"

	"github.com/joaosczip/go-rate-limiter/internal/http/middlewares"
	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
)

func listOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("list of orders"))
}

func main() {
	http.Handle("/", middlewares.RateLimiter(listOrders, &ratelimiter.RateLimiterConfig{RequestesPerSecond: 5, BlockUserFor: time.Duration(10 * time.Second)}))
	http.ListenAndServe(":8080", nil)
}
