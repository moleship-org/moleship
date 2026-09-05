package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLibpodStatusReportsEnabledFlag(t *testing.T) {
	h := &Libpod{enabled: false}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libpod/", nil)
	rr := httptest.NewRecorder()

	h.Status(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if body == "" || body == "null" {
		t.Fatalf("expected non-empty json body, got %q", body)
	}
	if want := `{"enabled":false}`; rr.Body.String() != want && rr.Body.String() != want+"\n" {
		t.Fatalf("expected body %q, got %q", want, rr.Body.String())
	}
}

func TestLibpodProxyReturnsForbiddenWhenDisabled(t *testing.T) {
	h := &Libpod{enabled: false}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/libpod/containers/json", nil)
	rr := httptest.NewRecorder()

	h.SocketAPI(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, rr.Code)
	}
}
