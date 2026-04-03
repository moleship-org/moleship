package model

import (
	"github.com/containers/podman/v5/pkg/domain/entities"
)

type PodmanSystemVersion struct {
	Data entities.ComponentVersion `json:"data,omitempty"`
}
