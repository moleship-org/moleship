package model

import (
	"strings"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

type Filters map[string][]string

func (f Filters) Query() string {
	first := true
	q := "?"

	for key, value := range f {
		if !first {
			q += "&"
		} else {
			first = false
		}
		q += key + "=" + strings.Join(value, ",")
	}

	return q
}

type PodmanSystemVersion struct {
	Data entities.ComponentVersion `json:"data,omitempty"`
}
