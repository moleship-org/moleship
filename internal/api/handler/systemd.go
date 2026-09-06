package handler

import (
	"encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/domain/systemd"
)

type Systemd struct {
	lg  *slog.Logger
	sys systemd.Port
}

func NewSystemd(lg *slog.Logger, s systemd.Port) *Systemd {
	return &Systemd{
		lg:  lg,
		sys: s,
	}
}

func (h *Systemd) Mount(r chi.Router) {
	h.lg.Info("Mounting /systemd endpoint")

	r.Group(func(r chi.Router) {
		r.Route("/systemd", func(r chi.Router) {
			r.Post("/daemon-reload", h.ReloadDaemon)

			r.Route("/units/{unit}", func(r chi.Router) {
				r.Get("/status", h.UnitStatus)
				r.Post("/start", h.StartUnit)
				r.Post("/stop", h.StopUnit)
				r.Post("/restart", h.RestartUnit)
			})
		})
	})
}

func (h *Systemd) UnitStatus(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	unit := chi.URLParam(r, "unit")

	status, err := h.sys.UnitStatus(r.Context(), unit)
	if err != nil {
		apiutil.AuditFailure(ctx, r, "systemd.unit.status", err, slog.String("unit", unit))
		writeSystemdError(ctx, err)
		return
	}

	body, err := json.Marshal(map[string]string{
		"unit":   unit,
		"status": status,
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "error encoding response")
		return
	}

	apiutil.AuditSuccess(ctx, r, "systemd.unit.status", slog.String("unit", unit), slog.String("status", status))
	ctx.JSONBlob(http.StatusOK, body)
}

func (h *Systemd) StartUnit(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	unit := chi.URLParam(r, "unit")

	if err := h.sys.StartUnit(r.Context(), unit); err != nil {
		apiutil.AuditFailure(ctx, r, "systemd.unit.start", err, slog.String("unit", unit))
		writeSystemdError(ctx, err)
		return
	}

	apiutil.AuditSuccess(ctx, r, "systemd.unit.start", slog.String("unit", unit))
	ctx.Status(http.StatusNoContent)
}

func (h *Systemd) StopUnit(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	unit := chi.URLParam(r, "unit")

	if err := h.sys.StopUnit(r.Context(), unit); err != nil {
		apiutil.AuditFailure(ctx, r, "systemd.unit.stop", err, slog.String("unit", unit))
		writeSystemdError(ctx, err)
		return
	}

	apiutil.AuditSuccess(ctx, r, "systemd.unit.stop", slog.String("unit", unit))
	ctx.Status(http.StatusNoContent)
}

func (h *Systemd) RestartUnit(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)
	unit := chi.URLParam(r, "unit")

	if err := h.sys.RestartUnit(r.Context(), unit); err != nil {
		apiutil.AuditFailure(ctx, r, "systemd.unit.restart", err, slog.String("unit", unit))
		writeSystemdError(ctx, err)
		return
	}

	apiutil.AuditSuccess(ctx, r, "systemd.unit.restart", slog.String("unit", unit))
	ctx.Status(http.StatusNoContent)
}

func (h *Systemd) ReloadDaemon(w http.ResponseWriter, r *http.Request) {
	ctx := apiutil.From(w, r)

	if err := h.sys.ReloadDaemon(r.Context()); err != nil {
		apiutil.AuditFailure(ctx, r, "systemd.daemon_reload", err)
		writeSystemdError(ctx, err)
		return
	}

	apiutil.AuditSuccess(ctx, r, "systemd.daemon_reload")
	ctx.Status(http.StatusNoContent)
}

func writeSystemdError(ctx apiutil.Context, err error) {
	apiutil.WriteMappedError(ctx, "Systemd endpoint error", err, "unexpected error calling systemd",
		apiutil.ErrorMapping{Target: systemd.ErrUnitNotFound, Status: http.StatusNotFound, Message: "unit not found"},
		apiutil.ErrorMapping{Target: systemd.ErrPermissionDenied, Status: http.StatusForbidden, Message: "permission denied"},
		apiutil.ErrorMapping{Target: systemd.ErrDaemonReloadFailed, Status: http.StatusInternalServerError, Message: "daemon-reload failed"},
		apiutil.ErrorMapping{Target: systemd.ErrCommandFailed, Status: http.StatusBadGateway, Message: "systemd command failed; check `systemctl --user status <unit>` and `journalctl --user -xeu <unit>` for details"},
	)
}
