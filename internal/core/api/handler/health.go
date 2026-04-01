package handler

import "net/http"

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

func (ht *Health) Mux(m *http.ServeMux) {
	m.Handle("/api/health", ht)
}
