package quadlet

import (
	"errors"
	"testing"
)

func TestContainerUnitValidatePublishPorts(t *testing.T) {
	tests := []struct {
		name         string
		publishPorts []string
		wantErr      bool
	}{
		{
			name:         "accepts host to container mapping",
			publishPorts: []string{"8080:80"},
		},
		{
			name:         "accepts ip host to container mapping",
			publishPorts: []string{"127.0.0.1:8080:80"},
		},
		{
			name:         "rejects host port above max",
			publishPorts: []string{"69696:80"},
			wantErr:      true,
		},
		{
			name:         "rejects container port above max",
			publishPorts: []string{"8080:69696"},
			wantErr:      true,
		},
		{
			name:         "rejects non numeric port",
			publishPorts: []string{"abc:80"},
			wantErr:      true,
		},
		{
			name:         "rejects malformed format",
			publishPorts: []string{"8080"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := &ContainerUnit{
				UnitName:     "nginx",
				Image:        "docker.io/library/nginx:latest",
				PublishPorts: tt.publishPorts,
			}

			err := unit.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, ErrInvalidUnit) {
					t.Fatalf("expected ErrInvalidUnit, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
