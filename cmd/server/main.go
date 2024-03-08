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
	config := ratelimiter.NewRateLimiterConfig(
		ratelimiter.NewRateLimiterConfigByIP(5, 10*time.Second),
		ratelimiter.NewRateLimiterConfigByToken(10, 20*time.Second, "API_KEY"),
	)
	http.Handle("/", middlewares.RateLimiter(listOrders, config))
	http.ListenAndServe(":8080", nil)
}
