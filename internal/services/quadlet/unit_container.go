package quadlet

import (
	"fmt"
	"sort"
)

type ContainerUnit struct {
	UnitName string `json:"name"`

	Description string `json:"description,omitempty"`

	After    []string `json:"after,omitempty"`
	Requires []string `json:"requires,omitempty"`

	// --- [Container] ---

	// Required container image
	Image string `json:"image"`

	// Optional execute commands
	Exec []string `json:"exec,omitempty"`

	// Overwrites the entrypoint of the image
	Entrypoint string `json:"entrypoint,omitempty"`

	// Overwrites the working directory of the image
	WorkingDir string `json:"working_dir,omitempty"`

	// Container user
	User string `json:"user,omitempty"`

	// Environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// EnvironmentFile list of files for the variables
	EnvironmentFile []string `json:"environment_file,omitempty"`

	// Mounting list of volumes
	Volumes []string `json:"volumes,omitempty"`

	// PublishPorts published ports in the format host:container
	PublishPorts []string `json:"publish_ports,omitempty"`

	// Networks
	Networks []string `json:"networks,omitempty"`

	// OCI labels
	Labels map[string]string `json:"labels,omitempty"`

	// "registry" or "local"
	AutoUpdate string `json:"auto_update,omitempty"`

	// Raw podman arguments for "podman run"
	PodmanArgs []string `json:"podman_args,omitempty"`

	// --- [Service] ---

	Restart string `json:"restart,omitempty"`

	// --- [Install] ---

	WantedBy []string `json:"wanted_by,omitempty"`
}

var _ Unit = (*ContainerUnit)(nil)

func (c *ContainerUnit) Name() string {
	return c.UnitName
}

func (c *ContainerUnit) Kind() Kind {
	return KindContainer
}

func (c *ContainerUnit) Validate() error {
	if c.UnitName == "" {
		return fmt.Errorf("%w: UnitName is required", ErrInvalidUnit)
	}
	if err := validateName(c.UnitName); err != nil {
		return fmt.Errorf("%w: UnitName %q is not valid: %v", ErrInvalidUnit, c.UnitName, err)
	}
	if c.Image == "" {
		return fmt.Errorf("%w: Image is required", ErrInvalidUnit)
	}
	return nil
}

func (c *ContainerUnit) Render() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	restart := c.Restart
	if restart == "" {
		restart = "always"
	}

	wantedBy := c.WantedBy
	if len(wantedBy) == 0 {
		wantedBy = []string{"default.target"}
	}

	environment := envLines(c.Environment)

	b := newIniBuilder()

	b.section("Unit").
		kv("Description", c.Description).
		kvSpaceJoined("After", c.After).
		kvSpaceJoined("Requires", c.Requires)

	b.section("Container").
		kv("Image", c.Image)

	if len(c.Exec) > 0 {
		b.kv("Exec", quoteArgs(c.Exec))
	}

	b.kv("Entrypoint", c.Entrypoint).
		kv("WorkingDir", c.WorkingDir).
		kv("User", c.User).
		kvList("Environment", environment).
		kvList("EnvironmentFile", c.EnvironmentFile).
		kvList("Volume", c.Volumes).
		kvList("PublishPort", c.PublishPorts).
		kvList("Network", c.Networks).
		kvMap("Label", c.Labels).
		kv("AutoUpdate", c.AutoUpdate).
		kvList("PodmanArgs", c.PodmanArgs)

	b.section("Service").
		kv("Restart", restart)

	b.section("Install").
		kvSpaceJoined("WantedBy", wantedBy)

	return b.String(), nil
}

func envLines(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}

	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("%s=%s", name, quoteIfNeeded(env[name])))
	}
	return lines
}
