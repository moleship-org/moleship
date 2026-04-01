package model

import "time"

type Quadlet struct {
	/* Systemd service */

	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Status    string    `json:"status"` // active, inactive, failed, etc.
	Content   string    `json:"content,omitempty"`
	StartedAt time.Time `json:"started_at,omitzero"`

	/* Container Runtime Information */

	ContainerID string    `json:"container_id,omitempty"`
	Image       string    `json:"image,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitzero"`

	/* Network & Connectivity */

	// MacAddress es útil para debuguear conflictos de red L2.
	MacAddress string `json:"mac_address,omitempty"`
	// Networks es la lista de nombres de redes (esto sí viene en ListContainer)
	Networks []string `json:"networks,omitempty"`
	// IPAddress (Podman lo mete en un campo llamado Namespaces o lo requiere vía Inspect)
	IPAddress string `json:"ip_address,omitempty"`
	// Ports viene como un []entities.PortMapping
	Ports []string `json:"ports,omitempty"`
}
