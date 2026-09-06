package apiutil

import (
	"errors"
	"log/slog"
	"net/http"
)

// ErrorMapping maps an error (checked via errors.Is) to an HTTP status and an
// optional response message. If Message is empty, err.Error() is used instead.
type ErrorMapping struct {
	Target  error
	Status  int
	Message string
}

// WriteMappedError logs err under logMsg and writes the response for the first
// mapping whose Target matches err (checked in order via errors.Is). If no
// mapping matches, it responds with 500 and fallbackMsg.
//
// This centralizes the repetitive "log then switch on errors.Is" pattern used
// by handlers to translate domain errors into HTTP responses.
func WriteMappedError(ctx Context, logMsg string, err error, fallbackMsg string, mappings ...ErrorMapping) error {
	ctx.Logger().Error(logMsg, slog.String("error", err.Error()))

	for _, m := range mappings {
		if errors.Is(err, m.Target) {
			msg := m.Message
			if msg == "" {
				msg = err.Error()
			}
			return ctx.Error(m.Status, msg)
		}
	}

	return ctx.Error(http.StatusInternalServerError, fallbackMsg)
}
