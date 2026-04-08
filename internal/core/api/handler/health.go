package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/core/api/middleware"
	"golang.org/x/time/rate"
)

// Health godoc
//
//	@Summary		Health check
//	@Description	Check server health
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{string}	string	"OK"
//	@Router			/health [get]
type Health struct {
	allows string
}

func NewHealth() *Health {
	r := new(Health)
	r.allows = "OPTIONS, GET"
	return r
}

func (ht *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Allow", ht.allows)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		w.WriteHeader(http.StatusOK)

	default:
		w.Header().Set("Allow", ht.allows)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ht *Health) Mux(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitByIP(rate.Every(time.Second), 60))
		r.Handle("/health", ht)
	})
}
