package middleware

import (
	"context"
	"net/http"

	"github.com/moleship-org/moleship/internal/core/api/apiutil"
)

func ContextInjector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mCtx := apiutil.NewContext(w, r)
		goCtx := context.WithValue(r.Context(), apiutil.CtxKey, mCtx)
		newR := r.WithContext(goCtx)

		mCtx.SetRequest(newR)

		next.ServeHTTP(w, newR)
	})
}
