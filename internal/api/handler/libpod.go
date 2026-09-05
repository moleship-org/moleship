package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/config"
	"github.com/moleship-org/moleship/internal/domain/podman"
)

type Libpod struct {
	lg      *slog.Logger
	pdm     podman.Port
	enabled bool
}

func NewLibpod(lg *slog.Logger, s podman.Port) *Libpod {
	return &Libpod{
		lg:      lg,
		pdm:     s,
		enabled: config.ENABLE_LIBPOD_PROXY,
	}
}

func (h *Libpod) Mount(r chi.Router) {
	h.lg.Info("Mounting /libpod endpoint", slog.Bool("enabled", h.enabled))

	r.Route("/libpod", func(r chi.Router) {
		r.Get("/", h.Status)
		r.HandleFunc("/*", h.SocketAPI)
	})
}

func (h *Libpod) Status(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	ctx.JSON(http.StatusOK, map[string]any{
		"enabled": h.enabled,
	})
}

func (h *Libpod) SocketAPI(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	path := ctx.PathValue("*")

	if !h.enabled {
		err := http.ErrNotSupported
		auditFailure(ctx, r, "libpod.proxy", err, slog.String("proxy_path", path))
		ctx.Error(http.StatusForbidden, "libpod proxy is disabled")
		return
	}

	libpodPath := strings.Split(path, "/")
	libpodPath = append(libpodPath, "?", r.URL.Query().Encode())

	res, err := h.pdm.RawCall(r.Context(), r.Method, libpodPath...)
	if err != nil {
		auditFailure(ctx, r, "libpod.proxy", err, slog.String("proxy_path", path))
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
			auditFailure(ctx, r, "libpod.proxy", err, slog.String("proxy_path", path))
			ctx.Error(http.StatusInternalServerError, "error when trying to read request body")
			return
		}

		auditSuccess(ctx, r, "libpod.proxy", slog.String("proxy_path", path), slog.Int("status_code", res.StatusCode))
		ctx.Bytes(res.StatusCode, b)
		return
	}

	auditSuccess(ctx, r, "libpod.proxy", slog.String("proxy_path", path), slog.Int("status_code", res.StatusCode))
	ctx.Status(res.StatusCode)
}
