package middleware

import (
	"net/http"

	"github.com/moleship-org/moleship/internal/core/env"
)

func CORS(env *env.Env) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", env.CORSAllowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-CRSF-Token, X-Requested-With")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			if env.CORSAllowedOrigins != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
