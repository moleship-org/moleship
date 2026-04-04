package middleware

import "net/http"

func Auth() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Auth handling
			next.ServeHTTP(w, r)
		})
	}
}
