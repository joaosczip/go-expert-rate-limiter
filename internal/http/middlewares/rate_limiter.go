package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
	"github.com/redis/go-redis/v9"
)

type Response struct {
	Message string `json:"message"`
}

func RateLimiter(next func(w http.ResponseWriter, r *http.Request), config *ratelimiter.RateLimiterConfig, redisClient *redis.Client) http.Handler {
	rateLimiter := ratelimiter.NewRateLimiter(
		ratelimiter.NewRedisDatasource(redisClient),
		ratelimiter.NewTimeSleeper(),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			fmt.Printf("error extracting the ip address from the request: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var token = ""

		if config.ConfigByToken != nil {
			token = r.Header.Get(config.ConfigByToken.Key)
		}

		err = rateLimiter.HandleRequest(ip, token, config)

		if err == nil {
			next(w, r)
		} else {
			var statusCode int
			var errMessage string

			if !errors.Is(err, ratelimiter.ErrMaxRequests) {
				errMessage = "Internal Server Error"
				statusCode = http.StatusInternalServerError
			} else {
				errMessage = err.Error()
				statusCode = http.StatusTooManyRequests
			}

			response := Response{
				Message: errMessage,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(&response)
		}
	})
}
