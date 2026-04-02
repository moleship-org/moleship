package serializer

import "github.com/moleship-org/moleship/internal/domain/model"

type ListContainer struct {
	Data []model.ContainerEntity `json:"data"`
}

type GetContainer struct {
	Data *model.ContainerEntity `json:"data"`
}
