package serializer

import "github.com/moleship-org/moleship/internal/domain/model"

type ListQuadlet struct {
	Data []model.QuadletEntity `json:"data"`
}

type GetQuadlet struct {
	Data *model.QuadletEntity `json:"data"`
}
