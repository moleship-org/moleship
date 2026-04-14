package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/adapter/podman"
	"github.com/moleship-org/moleship/internal/core/api/apiutil"
	"github.com/moleship-org/moleship/internal/domain/port"
)

type Libpod struct {
	podmanProv port.PodmanProvider
}

func NewLibpod(s port.PodmanProvider) *Libpod {
	return &Libpod{
		podmanProv: s,
	}
}

func (p *Libpod) Libpod(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.FromRequest(w, r)
	path := ctx.PathValue("*")

	libpodPath := strings.Split(path, "/")
	libpodPath = append(libpodPath, "?", r.URL.Query().Encode())

	res, err := p.podmanProv.RawCall(r.Context(), r.Method, libpodPath...)
	if errors.Is(err, podman.ErrContainerNotFound) {
		ctx.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "error trying to call podman socket")
		return
	}
	defer res.Body.Close()

	if res.Body != nil {
		b, err := io.ReadAll(res.Body)
		if err != nil && err != io.EOF {
			ctx.Error(http.StatusInternalServerError, "error when trying to read request body")
			return
		}

		ctx.Bytes(res.StatusCode, b)
		return
	}

	for key, value := range res.Header {
		ctx.Header().Set(key, strings.Join(value, ","))
	}
	ctx.Status(res.StatusCode)
}

func (p *Libpod) Mux(r chi.Router) {
	r.Route("/libpod", func(r chi.Router) {
		r.HandleFunc("/*", p.Libpod)
	})
}
