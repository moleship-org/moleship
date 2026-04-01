package model

import "github.com/containers/podman/v5/pkg/domain/entities"

type Quadlet struct {
	/* Systemd service */

	Name    string `json:"name"`
	Path    string `json:"path"`
	Status  string `json:"status"` // active, inactive, failed, etc.
	Content string `json:"content,omitempty"`

	/* Container Runtime Information */

	Container entities.ListContainer `json:"container,omitempty"`
}
