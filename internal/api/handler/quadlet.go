package handler

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/domain/systemd"
	"github.com/moleship-org/moleship/internal/services/quadlet"
)

type Quadlet struct {
	lg  *slog.Logger
	svc *quadlet.QuadletService
}

func NewQuadlet(lg *slog.Logger, svc *quadlet.QuadletService) *Quadlet {
	return &Quadlet{
		lg:  lg,
		svc: svc,
	}
}

func (h *Quadlet) Mount(r chi.Router) {
	h.lg.Info("Mounting /quadlet endpoint")

	r.Group(func(r chi.Router) {
		r.Route("/quadlet", func(r chi.Router) {
			r.Get("/", h.List)

			r.Route("/containers", func(r chi.Router) {
				r.Get("/", h.ListContainers)
				r.Post("/", h.CreateContainer)

				r.Route("/{name}", func(r chi.Router) {
					r.Get("/", h.readHandler(quadlet.KindContainer))
					r.Get("/status", h.statusHandler(quadlet.KindContainer))
					r.Post("/start", h.startHandler(quadlet.KindContainer))
					r.Post("/stop", h.stopHandler(quadlet.KindContainer))
					r.Post("/restart", h.restartHandler(quadlet.KindContainer))
					r.Delete("/", h.deleteHandler(quadlet.KindContainer))
				})
			})

			// TODO: r.Route("/volumes", ...) / r.Route("/networks", ...) ...
		})
	})
}

// POST/quadlet/containers?fail_if_exists=<bool>
func (h *Quadlet) CreateContainer(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)

	unit := &quadlet.ContainerUnit{}
	if err := json.UnmarshalRead(r.Body, &unit); err != nil {
		apiutil.AuditFailure(ctx, r, "quadlet.container.create", err)
		ctx.Error(http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := unit.Validate(); err != nil {
		apiutil.AuditFailure(ctx, r, "quadlet.container.create", err, slog.String("name", unit.Name()), slog.String("kind", string(unit.Kind())))
		writeQuadletError(ctx, err)
		return
	}

	opts := quadlet.CreateOptions{
		Start:        r.URL.Query().Get("start") == "true",
		FailIfExists: r.URL.Query().Get("fail_if_exists") == "true",
	}

	if err := h.svc.Create(r.Context(), unit, opts); err != nil {
		apiutil.AuditFailure(ctx, r, "quadlet.container.create", err, slog.String("name", unit.Name()), slog.String("kind", string(unit.Kind())), slog.Bool("start", opts.Start), slog.Bool("fail_if_exists", opts.FailIfExists))
		writeQuadletError(ctx, err)
		return
	}

	body, err := json.Marshal(map[string]string{
		"name":         unit.Name(),
		"kind":         string(unit.Kind()),
		"service_name": quadlet.ServiceName(unit),
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "error encoding response")
		return
	}

	apiutil.AuditSuccess(ctx, r, "quadlet.container.create", slog.String("name", unit.Name()), slog.String("kind", string(unit.Kind())), slog.String("service_name", quadlet.ServiceName(unit)), slog.Bool("start", opts.Start), slog.Bool("fail_if_exists", opts.FailIfExists))
	ctx.JSONBlob(http.StatusCreated, body)
}

// --- Generic operations by Kind ---

func (h *Quadlet) readHandler(kind quadlet.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := apiutil.From(w, r)
		name := chi.URLParam(r, "name")

		content, err := h.svc.Read(r.Context(), kind, name)
		if err != nil {
			apiutil.AuditFailure(ctx, r, "quadlet.read", err, slog.String("name", name), slog.String("kind", string(kind)))
			writeQuadletError(ctx, err)
			return
		}

		apiutil.AuditSuccess(ctx, r, "quadlet.read", slog.String("name", name), slog.String("kind", string(kind)))
		ctx.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctx.Bytes(http.StatusOK, []byte(content))
	}
}

func (h *Quadlet) deleteHandler(kind quadlet.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := apiutil.From(w, r)
		name := chi.URLParam(r, "name")

		if err := h.svc.Remove(r.Context(), kind, name); err != nil {
			apiutil.AuditFailure(ctx, r, "quadlet.delete", err, slog.String("name", name), slog.String("kind", string(kind)))
			writeQuadletError(ctx, err)
			return
		}

		apiutil.AuditSuccess(ctx, r, "quadlet.delete", slog.String("name", name), slog.String("kind", string(kind)))
		ctx.Status(http.StatusNoContent)
	}
}

func (h *Quadlet) statusHandler(kind quadlet.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := apiutil.From(w, r)
		name := chi.URLParam(r, "name")

		status, err := h.svc.Status(r.Context(), kind, name)
		if err != nil {
			apiutil.AuditFailure(ctx, r, "quadlet.status", err, slog.String("name", name), slog.String("kind", string(kind)))
			writeQuadletError(ctx, err)
			return
		}

		body, err := json.Marshal(map[string]string{
			"name":   name,
			"kind":   string(kind),
			"status": status,
		})
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "error encoding response")
			return
		}

		apiutil.AuditSuccess(ctx, r, "quadlet.status", slog.String("name", name), slog.String("kind", string(kind)), slog.String("status", status))
		ctx.JSONBlob(http.StatusOK, body)
	}
}

func (h *Quadlet) startHandler(kind quadlet.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := apiutil.From(w, r)
		name := chi.URLParam(r, "name")

		if err := h.svc.Start(r.Context(), kind, name); err != nil {
			apiutil.AuditFailure(ctx, r, "quadlet.start", err, slog.String("name", name), slog.String("kind", string(kind)))
			writeQuadletError(ctx, err)
			return
		}

		apiutil.AuditSuccess(ctx, r, "quadlet.start", slog.String("name", name), slog.String("kind", string(kind)))
		ctx.Status(http.StatusNoContent)
	}
}

func (h *Quadlet) stopHandler(kind quadlet.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := apiutil.From(w, r)
		name := chi.URLParam(r, "name")

		if err := h.svc.Stop(r.Context(), kind, name); err != nil {
			apiutil.AuditFailure(ctx, r, "quadlet.stop", err, slog.String("name", name), slog.String("kind", string(kind)))
			writeQuadletError(ctx, err)
			return
		}

		apiutil.AuditSuccess(ctx, r, "quadlet.stop", slog.String("name", name), slog.String("kind", string(kind)))
		ctx.Status(http.StatusNoContent)
	}
}

func (h *Quadlet) restartHandler(kind quadlet.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := apiutil.From(w, r)
		name := chi.URLParam(r, "name")

		if err := h.svc.Restart(r.Context(), kind, name); err != nil {
			apiutil.AuditFailure(ctx, r, "quadlet.restart", err, slog.String("name", name), slog.String("kind", string(kind)))
			writeQuadletError(ctx, err)
			return
		}

		apiutil.AuditSuccess(ctx, r, "quadlet.restart", slog.String("name", name), slog.String("kind", string(kind)))
		ctx.Status(http.StatusNoContent)
	}
}

// --- List ---

type unitInfoResponse struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	ServiceName string `json:"service_name"`
	Status      string `json:"status,omitempty"`
	StatusError string `json:"status_error,omitempty"`
}

func (h *Quadlet) List(w http.ResponseWriter, r *http.Request) {
	h.listByKind(w, r, nil, "quadlet.list")
}

func (h *Quadlet) ListContainers(w http.ResponseWriter, r *http.Request) {
	kind := quadlet.KindContainer
	h.listByKind(w, r, &kind, "quadlet.container.list")
}

func (h *Quadlet) listByKind(w http.ResponseWriter, r *http.Request, kind *quadlet.Kind, action string) {
	ctx := apiutil.From(w, r)

	entries, err := h.svc.List(r.Context())
	if err != nil {
		apiutil.AuditFailure(ctx, r, action, err)
		writeQuadletError(ctx, err)
		return
	}

	resp := make([]unitInfoResponse, 0, len(entries))
	for _, e := range entries {
		if kind != nil && e.Kind != *kind {
			continue
		}

		item := unitInfoResponse{
			Name:        e.Name,
			Kind:        string(e.Kind),
			ServiceName: e.ServiceName,
			Status:      e.Status,
		}
		if e.StatusError != nil {
			item.StatusError = e.StatusError.Error()
		}
		resp = append(resp, item)
	}

	body, err := json.Marshal(resp)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "error encoding response")
		return
	}

	apiutil.AuditSuccess(ctx, r, action, slog.Int("count", len(resp)))
	ctx.Header().Set("Content-Type", "application/json")
	ctx.Bytes(http.StatusOK, body)
}

// --- Error mapping ---

func writeQuadletError(ctx apiutil.Context, err error) {
	ctx.Logger().Error("Quadlet endpoint error", slog.String("error", err.Error()))

	switch {
	case errors.Is(err, quadlet.ErrInvalidUnit),
		errors.Is(err, quadlet.ErrInvalidName),
		errors.Is(err, quadlet.ErrInvalidKind):
		ctx.Error(http.StatusBadRequest, err.Error())
	case errors.Is(err, quadlet.ErrUnitAlreadyExists):
		ctx.Error(http.StatusConflict, "unit already exists")
	case errors.Is(err, quadlet.ErrUnitNotFound),
		errors.Is(err, systemd.ErrUnitNotFound):
		ctx.Error(http.StatusNotFound, "unit not found")
	case errors.Is(err, systemd.ErrPermissionDenied):
		ctx.Error(http.StatusForbidden, "permission denied")
	case errors.Is(err, systemd.ErrCommandFailed):
		ctx.Error(http.StatusBadGateway, "systemd failed to apply the unit change; check `systemctl --user status` and `journalctl --user -xeu` for the service")
	default:
		ctx.Error(http.StatusInternalServerError, "unexpected error")
	}
}
