package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(lg *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			if lg != nil {
				lg.Debug(
					"Request",
					"method", r.Method,
					"path", r.URL.Path,
					"since", time.Since(start),
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
				)
			}
		})
	}
}
