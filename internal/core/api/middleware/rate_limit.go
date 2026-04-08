package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func RateLimit(limitFreq rate.Limit, burst int) Middleware {
	limiter := rate.NewLimiter(limitFreq, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// API specific rate limit headers

			// "X-RateLimit-Limit": The maximum number of requests that the client is allowed to make in a given time window.
			windowBurst := limiter.Burst()
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(windowBurst))

			// "X-RateLimit-Remaining": The number of requests remaining in the current time window before the client is rate limited.
			remainingTokens := int(limiter.Tokens())
			if remainingTokens < 0 {
				remainingTokens = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remainingTokens))

			// "X-RateLimit-Reset": The time at which the current rate limit window resets in UTC epoch seconds.
			tokensMissing := float64(limiter.Burst()) - float64(limiter.Tokens())
			secondsUntilFull := tokensMissing / float64(limiter.Limit())
			windowResetTime := time.Now().Add(time.Duration(secondsUntilFull * float64(time.Second))).Unix()

			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(windowResetTime, 10))

			// RFC 7231#section-7.1.3
			if !limiter.Allow() {
				reservation := limiter.Reserve()
				if !reservation.OK() {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}

				delay := reservation.Delay()
				reservation.Cancel()

				// "non-negative decimal integer"
				secondsToWait := int(math.Ceil(delay.Seconds()))
				if secondsToWait <= 0 {
					secondsToWait = 1
				}

				w.Header().Add("Retry-After", strconv.Itoa(secondsToWait))
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

var (
	clients map[string]*client
	mu      sync.Mutex
	once    sync.Once
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimitByIP(limitFreq rate.Limit, burst int) Middleware {
	once.Do(func() {
		clients = make(map[string]*client)

		go func() {
			freq := time.Duration(limitFreq)
			for {
				mu.Lock()
				for ip, c := range clients {
					if time.Since(c.lastSeen) > freq {
						delete(clients, ip)
					}
				}
				mu.Unlock()
				time.Sleep(freq)
			}
		}()
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rStr *http.Request) {
			ip, _, err := net.SplitHostPort(rStr.RemoteAddr)
			if err != nil {
				ip = rStr.RemoteAddr
			}

			mu.Lock()
			if _, v := clients[ip]; !v {
				clients[ip] = &client{limiter: rate.NewLimiter(limitFreq, burst)}
			}
			clients[ip].lastSeen = time.Now()
			limiter := clients[ip].limiter
			mu.Unlock()

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.Burst()))

			remainingTokens := int(limiter.Tokens())
			if remainingTokens < 0 {
				remainingTokens = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remainingTokens))

			// Calculate when the rate limit will reset
			tokensMissing := float64(limiter.Burst()) - float64(limiter.Tokens())
			secondsUntilFull := tokensMissing / float64(limiter.Limit())
			windowResetTime := time.Now().Add(time.Duration(secondsUntilFull * float64(time.Second))).Unix()

			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(windowResetTime, 10))

			if !limiter.Allow() {
				reservation := limiter.Reserve()
				if !reservation.OK() {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
				delay := reservation.Delay()
				reservation.Cancel()

				secondsToWait := int(math.Ceil(delay.Seconds()))
				if secondsToWait <= 0 {
					secondsToWait = 1
				}

				w.Header().Set("Retry-After", strconv.Itoa(secondsToWait))
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, rStr)
		})
	}
}
