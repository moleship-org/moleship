package handler

import (
	"log/slog"
	"net/http"

	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/moleship-org/moleship/internal/api/apiutil"
	"github.com/moleship-org/moleship/internal/services/auth/authtoken"
)

func auditAttrs(ctx apiutil.Context, r *http.Request, action string, attrs ...slog.Attr) []any {
	requestID := chi_middleware.GetReqID(r.Context())
	user, ok := authtoken.UserFromContext(r.Context())
	if !ok || user == "" {
		user = "anonymous"
	}

	base := []any{
		slog.String("event", "audit"),
		slog.String("action", action),
		slog.String("user", user),
		slog.String("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	}

	for _, attr := range attrs {
		base = append(base, attr)
	}

	return base
}

func auditSuccess(ctx apiutil.Context, r *http.Request, action string, attrs ...slog.Attr) {
	args := append(auditAttrs(ctx, r, action, attrs...), slog.Bool("success", true))
	ctx.Logger().Info("audit", args...)
}

func auditFailure(ctx apiutil.Context, r *http.Request, action string, err error, attrs ...slog.Attr) {
	args := append(auditAttrs(ctx, r, action, attrs...),
		slog.Bool("success", false),
		slog.String("error", err.Error()),
	)
	ctx.Logger().Warn("audit", args...)
}
