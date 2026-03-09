package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/luis13005/ratelimiter/internal/limiter"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func RateLimiter(rl *limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()

			fmt.Printf("[Request] method=%s path=%s\n", r.Method, r.URL.Path)

			if r.URL.Path == "/favicon.ico" || r.URL.Path == "/.well-known/appspecific/com.chrome.devtools.json" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			key := ip
			isToken := false

			if token := r.Header.Get("API_KEY"); token != "" {
				key = token
				isToken = true
			}

			result, err := rl.Allow(ctx, key, isToken)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining(result)))

			if !result.Allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(errorResponse{
					Error:   "rate_limit_exceeded",
					Message: fmt.Sprintf("you have reached the maximum number of requests try again in: %.0fs", result.RetryAfter.Seconds()),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func remaining(r *limiter.AllowResult) int64 {
	rem := int64(r.Limit) - r.Current
	if rem < 0 {
		return 0
	}
	return rem
}
