package middlewares

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
)

type Response struct {
	Message string `json:"message"`
}

func RateLimiter(next func(w http.ResponseWriter, r *http.Request), config *ratelimiter.RateLimiterConfig) http.Handler {
	rateLimiter := ratelimiter.NewRateLimiter()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			fmt.Printf("error extracting the ip address from the request: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = rateLimiter.HandleRequest(ip, *config)

		if err == nil {
			next(w, r)
		} else {
			response := Response{
				Message: err.Error(),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&response)
		}
	})
}
