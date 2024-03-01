package main

import (
	"net/http"
	"time"

	"github.com/joaosczip/go-rate-limiter/internal/http/middlewares"
)

func listOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("list of orders"))
}

func main() {
	http.Handle("/", middlewares.RateLimiter(listOrders, &middlewares.RateLimiterConfig{RequestesPerSecond: 5, BlockUserFor: time.Duration(10 * time.Second)}))
	http.ListenAndServe(":8080", nil)
}
