package middleware

import "net/http"

func Auth(apiKey string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Auth handling
			next.ServeHTTP(w, r)
		})
	}
}
