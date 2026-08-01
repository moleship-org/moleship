package serializer

import (
	"github.com/moleship-org/moleship/internal/domain/podman"
)

type ListContainer struct {
	Data []podman.ContainerEntity `json:"data"`
}

type GetContainer struct {
	Data *podman.ContainerEntity `json:"data"`
}
