package main

import (
	"net/http"
	"time"

	"github.com/joaosczip/go-rate-limiter/configs"
	"github.com/joaosczip/go-rate-limiter/internal/http/middlewares"
	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
)

func listOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("list of orders"))
}

func main() {
	envConf, err := configs.LoadConfig(".")

	if err != nil {
		panic(err)
	}

	rateLimiterConf := ratelimiter.NewRateLimiterConfig(
		ratelimiter.NewRateLimiterConfigByIP(envConf.MaxRequestsByIP, time.Duration(envConf.BlockUserForByIP)*time.Second),
		ratelimiter.NewRateLimiterConfigByToken(envConf.MaxRequestsByToken, time.Duration(envConf.BlockUserForByToken)*time.Second, "API_KEY"),
	)
	http.Handle("/", middlewares.RateLimiter(listOrders, rateLimiterConf))
	http.ListenAndServe(":8080", nil)
}
