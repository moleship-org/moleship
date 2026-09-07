package systemd

import (
	"errors"
	"testing"
)

func TestClassifyUnitError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   error
	}{
		{
			name:   "unit not found",
			stderr: "Failed to stop foo.service: Unit foo.service not found.",
			want:   ErrUnitNotFound,
		},
		{
			name:   "unit does not exist",
			stderr: "Unit foo.service does not exist.",
			want:   ErrUnitNotFound,
		},
		{
			name:   "unit not loaded",
			stderr: "Failed to stop foo.service: Unit foo.service not loaded.",
			want:   ErrUnitNotFound,
		},
		{
			name:   "permission denied",
			stderr: "Failed to stop foo.service: Access denied.",
			want:   ErrPermissionDenied,
		},
		{
			name:   "transient or generated unit",
			stderr: "Failed to enable unit: Unit /run/user/1000/systemd/generator/server.service is transient or generated.",
			want:   ErrUnitTransientOrGenerated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyUnitError(tt.stderr, errors.New("exit status 5"))
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, err)
			}
		})
	}
}

func TestClassifyReloadDaemonError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   error
	}{
		{
			name:   "permission denied",
			stderr: "Failed to reload daemon: Permission denied",
			want:   ErrPermissionDenied,
		},
		{
			name:   "access denied",
			stderr: "Failed to reload daemon: Access denied",
			want:   ErrPermissionDenied,
		},
		{
			name:   "generic failure",
			stderr: "Failed to reload daemon: transport endpoint closed",
			want:   ErrDaemonReloadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyReloadDaemonError(tt.stderr, errors.New("exit status 1"))
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, err)
			}
		})
	}
}
