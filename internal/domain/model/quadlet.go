package model

type QuadletFile struct {
	Unit      UnitOptions      `json:"unit,omitempty"`
	Service   ServiceOptions   `json:"service,omitempty"`
	Container ContainerOptions `json:"container,omitempty"`
	Install   InstallOptions   `json:"install,omitempty"`
}

// [Unit]
type UnitOptions struct {
	Description string   `json:"description,omitempty"`
	Requires    []string `json:"requires,omitempty"`
	Wants       []string `json:"wants,omitempty"`
	After       []string `json:"after,omitempty"`
	Before      []string `json:"before,omitempty"`
}

// [Service]
type ServiceOptions struct {
	Restart         string   `json:"restart,omitempty"`
	TimeoutStartSec string   `json:"timeout_start_sec,omitempty"`
	TimeoutStopSec  string   `json:"timeout_stop_sec,omitempty"`
	Environment     []string `json:"environment,omitempty"`
	ExecStartPre    []string `json:"exec_start_pre,omitempty"`
	ExecStartPost   []string `json:"exec_start_post,omitempty"`
}

// [Container]
// Ref: https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html
type ContainerOptions struct {
	// --- Base ---

	Image         string `json:"image,omitempty"`
	ContainerName string `json:"container_name,omitempty"`

	// --- Network ---

	Network        []string `json:"network,omitempty"`
	PublishPort    []string `json:"publish_port,omitempty"`
	ExposeHostPort []string `json:"expose_host_port,omitempty"`

	// --- Volumes ---

	Volume []string `json:"volume,omitempty"`
	Mount  []string `json:"mount,omitempty"`

	// --- Environment and Secrets ---

	Environment     []string `json:"environment,omitempty"`
	EnvironmentFile []string `json:"environment_file,omitempty"`
	Secret          []string `json:"secret,omitempty"`

	// --- Binary entry point ---

	Exec       string `json:"exec,omitempty"`
	Args       string `json:"args,omitempty"`
	Entrypoint string `json:"entrypoint,omitempty"`

	// --- Life cycle ---

	AutoUpdate string `json:"auto_update,omitempty"`
	Removable  *bool  `json:"removable,omitempty"`

	// --- Metadata ---

	Label      []string `json:"label,omitempty"`
	Annotation []string `json:"annotation,omitempty"`

	// --- Security ---

	User                 string   `json:"user,omitempty"`
	UserNS               string   `json:"userns,omitempty"`
	DropCapability       []string `json:"drop_capability,omitempty"`
	AddCapability        []string `json:"add_capability,omitempty"`
	SecurityLabelDisable *bool    `json:"security_label_disable,omitempty"`

	// --- Health ---

	HealthCmd string `json:"health_cmd,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
	Pod       string `json:"pod,omitempty"`
}

// [Install]
type InstallOptions struct {
	WantedBy   []string `json:"wanted_by,omitempty"`
	RequiredBy []string `json:"required_by,omitempty"`
}
