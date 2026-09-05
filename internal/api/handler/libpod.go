package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/domain/podman"
)

type Libpod struct {
	lg  *slog.Logger
	pdm podman.Port
}

func NewLibpod(lg *slog.Logger, s podman.Port) *Libpod {
	return &Libpod{
		lg:  lg,
		pdm: s,
	}
}

func (h *Libpod) Mount(r chi.Router) {
	h.lg.Info("Mounting /libpod endpoint")

	r.Route("/libpod", func(r chi.Router) {
		r.HandleFunc("/*", h.SocketAPI)
	})
}

func (h *Libpod) SocketAPI(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	path := ctx.PathValue("*")

	libpodPath := strings.Split(path, "/")
	libpodPath = append(libpodPath, "?", r.URL.Query().Encode())

	res, err := h.pdm.RawCall(r.Context(), r.Method, libpodPath...)
	if err != nil {
		h.lg.Error("error trying to call podman socket", slog.String("error", err.Error()))
		ctx.Error(http.StatusInternalServerError, "error trying to call podman socket")
		return
	}
	defer res.Body.Close()

	for key, value := range res.Header {
		ctx.Header().Set(key, strings.Join(value, ","))
	}

	if res.Body != nil {
		b, err := io.ReadAll(res.Body)
		if err != nil && err != io.EOF {
			ctx.Error(http.StatusInternalServerError, "error when trying to read request body")
			return
		}

		ctx.Bytes(res.StatusCode, b)
		return
	}

	ctx.Status(res.StatusCode)
}
