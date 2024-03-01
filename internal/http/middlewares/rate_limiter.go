package middlewares

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/joaosczip/go-rate-limiter/pkg/ratelimiter"
)

type Response struct {
	Message string `json:"message"`
}

type RateLimiterConfig struct {
	RequestesPerSecond int
	BlockUserFor       time.Duration
}

func RateLimiter(next func(w http.ResponseWriter, r *http.Request), config *RateLimiterConfig) http.Handler {
	rateLimiterManger := ratelimiter.NewRateLimiterManager()
	mux := sync.Mutex{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			fmt.Printf("error extracting the ip address from the request: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		clients := rateLimiterManger.Clients

		if _, found := clients[ip]; !found {
			clients[ip] = ratelimiter.NewClientLimiter(config.RequestesPerSecond, config.BlockUserFor)
		}

		mux.Lock()
		defer mux.Unlock()
		clients[ip].LastSeen = time.Now()

		if clients[ip].IsBlocked() {
			if clients[ip].HasBlockingExpired() {
				clients[ip].ResetBlock()
				next(w, r)
				return
			} else {
				response := Response{
					Message: "you have reached the maximum number of requests or actions allowed within a certain time frame",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(&response)
				return
			}
		}

		clients[ip].TotalRequests += 1

		if clients[ip].ShouldBlock() {
			clients[ip].Block()
			response := Response{
				Message: "you have reached the maximum number of requests or actions allowed within a certain time frame",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&response)
		} else {
			next(w, r)
		}
	})
}
