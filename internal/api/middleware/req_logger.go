package middleware

import (
	"net/http"
	"time"

	"github.com/moleship-org/moleship/internal/api/apiutil"
)

func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := apiutil.From(w, r)

			start := time.Now()
			next.ServeHTTP(w, r)

			c.Logger().Debug(
				"Request",
				"method", r.Method,
				"path", r.URL.Path,
				"since", time.Since(start),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}
