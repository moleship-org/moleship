package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipRateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimit(limitFreq rate.Limit, burst int) Middleware {
	limiter := rate.NewLimiter(limitFreq, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowRequest(w, limiter) {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitByIP(limitFreq rate.Limit, burst int) Middleware {
	const staleClientTTL = 10 * time.Minute
	const cleanupInterval = 1 * time.Minute

	limiters := make(map[string]*ipRateLimiter)
	var mu sync.Mutex
	nextCleanupAt := time.Now().Add(cleanupInterval)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			mu.Lock()
			ipLimiter, ok := limiters[clientIP]
			if !ok {
				ipLimiter = &ipRateLimiter{limiter: rate.NewLimiter(limitFreq, burst)}
				limiters[clientIP] = ipLimiter
			}
			ipLimiter.lastSeen = now

			if !now.Before(nextCleanupAt) {
				cleanupStaleLimiters(limiters, now, staleClientTTL)
				nextCleanupAt = now.Add(cleanupInterval)
			}

			limiter := ipLimiter.limiter
			mu.Unlock()

			if !allowRequest(w, limiter) {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func cleanupStaleLimiters(limiters map[string]*ipRateLimiter, now time.Time, ttl time.Duration) {
	for ip, limiter := range limiters {
		if now.Sub(limiter.lastSeen) > ttl {
			delete(limiters, ip)
		}
	}
}

func setRateLimitHeaders(w http.ResponseWriter, limiter *rate.Limiter) {
	// "X-RateLimit-Limit": The maximum number of requests allowed in a time window.
	windowBurst := limiter.Burst()
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(windowBurst))

	// "X-RateLimit-Remaining": Remaining requests before the current limiter rejects.
	remainingTokens := int(limiter.Tokens())
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remainingTokens))

	// "X-RateLimit-Reset": UTC epoch seconds when the bucket refills.
	windowResetTime := computeWindowResetTime(time.Now(), limiter)
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(windowResetTime, 10))
}

func allowRequest(w http.ResponseWriter, limiter *rate.Limiter) bool {
	if limiter.Allow() {
		setRateLimitHeaders(w, limiter)
		return true
	}

	// RFC 7231#section-7.1.3
	reservation := limiter.Reserve()
	if !reservation.OK() {
		w.WriteHeader(http.StatusTooManyRequests)
		return false
	}

	delay := reservation.Delay()
	reservation.Cancel()

	// "non-negative decimal integer"
	secondsToWait := int(math.Ceil(delay.Seconds()))
	if secondsToWait <= 0 {
		secondsToWait = 1
	}

	w.Header().Set("Retry-After", strconv.Itoa(secondsToWait))
	setRateLimitHeaders(w, limiter)
	w.WriteHeader(http.StatusTooManyRequests)
	return false
}

func computeWindowResetTime(now time.Time, limiter *rate.Limiter) int64 {
	limit := float64(limiter.Limit())
	if limit <= 0 {
		return now.Unix()
	}

	tokensMissing := float64(limiter.Burst()) - limiter.Tokens()
	if tokensMissing < 0 {
		tokensMissing = 0
	}

	secondsUntilFull := tokensMissing / limit
	if math.IsInf(secondsUntilFull, 0) || math.IsNaN(secondsUntilFull) {
		return now.Unix()
	}

	return now.Add(time.Duration(secondsUntilFull * float64(time.Second))).Unix()
}

func getClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}

	remoteAddr := strings.TrimSpace(r.RemoteAddr)
	if remoteAddr == "" {
		return "unknown"
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil || host == "" {
		return remoteAddr
	}

	return host
}
