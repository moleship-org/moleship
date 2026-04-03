package serializer

import "github.com/moleship-org/moleship/internal/domain/model"

type ListQuadlets struct {
	Data []model.QuadletFile `json:"data"`
}
