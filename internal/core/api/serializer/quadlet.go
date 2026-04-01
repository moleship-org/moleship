package serializer

import "github.com/moleship-org/moleship/internal/domain/model"

type ListQuadlet struct {
	Data []model.Quadlet `json:"data"`
}

type GetQuadlet struct {
	Data model.Quadlet `json:"data"`
}
