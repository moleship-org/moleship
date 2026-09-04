package serializer

import "github.com/moleship-org/moleship/internal/service/quadlet"

type ListQuadlets struct {
	Data []quadlet.ContainerFile `json:"data"`
}
