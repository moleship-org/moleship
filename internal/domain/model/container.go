package model

import "github.com/containers/podman/v5/pkg/domain/entities"

type QuadletEntity struct {
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Status    string                 `json:"status"` // active, inactive, failed, etc.
	Container entities.ListContainer `json:"container,omitzero"`
}
