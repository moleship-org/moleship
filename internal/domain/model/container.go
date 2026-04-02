package model

import "github.com/containers/podman/v5/pkg/domain/entities"

type ContainerEntity struct {
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Status    string                 `json:"status"` // active, inactive, failed, etc.
	Container entities.ListContainer `json:"container,omitzero"`
}

type ContainerStats struct {
	Name        string `json:"name"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
	} `json:"cpu_stats"`
}
