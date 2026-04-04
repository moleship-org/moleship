package model

import (
	"fmt"
	"strings"
)

type RawQuadletFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

type QuadletFile struct {
	Name      string           `json:"name"`
	Path      string           `json:"path"`
	Unit      UnitOptions      `json:"unit"`
	Service   ServiceOptions   `json:"service"`
	Container ContainerOptions `json:"container"`
	Install   InstallOptions   `json:"install"`
}

func (q *QuadletFile) String() string {
	var sb strings.Builder

	// --- [Unit] ---
	if !isUnitEmpty(q.Unit) {
		sb.WriteString("[Unit]\n")
		writeString(&sb, "Description", q.Unit.Description)
		writeSlice(&sb, "Requires", q.Unit.Requires)
		writeSlice(&sb, "Wants", q.Unit.Wants)
		writeSlice(&sb, "After", q.Unit.After)
		writeSlice(&sb, "Before", q.Unit.Before)
		sb.WriteString("\n")
	}

	// --- [Service] ---
	if !isServiceEmpty(q.Service) {
		sb.WriteString("[Service]\n")
		writeString(&sb, "Restart", q.Service.Restart)
		writeString(&sb, "TimeoutStartSec", q.Service.TimeoutStartSec)
		writeString(&sb, "TimeoutStopSec", q.Service.TimeoutStopSec)
		writeSlice(&sb, "Environment", q.Service.Environment)
		writeSlice(&sb, "ExecStartPre", q.Service.ExecStartPre)
		writeSlice(&sb, "ExecStartPost", q.Service.ExecStartPost)
		sb.WriteString("\n")
	}

	// --- [Container] ---
	sb.WriteString("[Container]\n")
	writeString(&sb, "Image", q.Container.Image)
	writeString(&sb, "ContainerName", q.Container.ContainerName)

	// Network & Ports
	writeSlice(&sb, "Network", q.Container.Network)
	writeSlice(&sb, "PublishPort", q.Container.PublishPort)
	writeSlice(&sb, "ExposeHostPort", q.Container.ExposeHostPort)

	// Volumes
	writeSlice(&sb, "Volume", q.Container.Volume)
	writeSlice(&sb, "Mount", q.Container.Mount)

	// Env & Secrets
	writeSlice(&sb, "Environment", q.Container.Environment)
	writeSlice(&sb, "EnvironmentFile", q.Container.EnvironmentFile)
	writeSlice(&sb, "Secret", q.Container.Secret)

	// Execution
	writeString(&sb, "Exec", q.Container.Exec)
	writeString(&sb, "Args", q.Container.Args)
	writeString(&sb, "Entrypoint", q.Container.Entrypoint)

	// Lifecycle
	writeString(&sb, "AutoUpdate", q.Container.AutoUpdate)
	writeBool(&sb, "Removable", q.Container.Removable)

	// Metadata
	writeSlice(&sb, "Label", q.Container.Label)
	writeSlice(&sb, "Annotation", q.Container.Annotation)

	// Security
	writeString(&sb, "User", q.Container.User)
	writeString(&sb, "UserNS", q.Container.UserNS)
	writeSlice(&sb, "DropCapability", q.Container.DropCapability)
	writeSlice(&sb, "AddCapability", q.Container.AddCapability)
	writeBool(&sb, "SecurityLabelDisable", q.Container.SecurityLabelDisable)

	// Health & Misc
	writeString(&sb, "HealthCmd", q.Container.HealthCmd)
	writeString(&sb, "Timezone", q.Container.Timezone)
	writeString(&sb, "Pod", q.Container.Pod)
	sb.WriteString("\n")

	// --- [Install] ---
	if !isInstallEmpty(q.Install) {
		sb.WriteString("[Install]\n")
		writeSlice(&sb, "WantedBy", q.Install.WantedBy)
		writeSlice(&sb, "RequiredBy", q.Install.RequiredBy)
	}

	return sb.String()
}

func writeString(sb *strings.Builder, key, value string) {
	if value != "" {
		sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
}

func writeSlice(sb *strings.Builder, key string, values []string) {
	for _, v := range values {
		sb.WriteString(fmt.Sprintf("%s=%s\n", key, v))
	}
}

func writeBool(sb *strings.Builder, key string, value *bool) {
	if value != nil {
		sb.WriteString(fmt.Sprintf("%s=%v\n", key, *value))
	}
}

func isUnitEmpty(u UnitOptions) bool {
	return u.Description == "" && len(u.Requires) == 0 && len(u.Wants) == 0 && len(u.After) == 0 && len(u.Before) == 0
}

func isServiceEmpty(s ServiceOptions) bool {
	return s.Restart == "" && len(s.Environment) == 0 && len(s.ExecStartPre) == 0
}

func isInstallEmpty(i InstallOptions) bool {
	return len(i.WantedBy) == 0 && len(i.RequiredBy) == 0
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
