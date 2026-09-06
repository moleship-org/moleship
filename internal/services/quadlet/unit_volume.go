package quadlet

import (
	"fmt"
	"sort"
)

type VolumeUnit struct {
	UnitName string `json:"name"`

	Description string `json:"description,omitempty"`

	After    []string `json:"after,omitempty"`
	Requires []string `json:"requires,omitempty"`

	// --- [Volume] ---

	ContainersConfModules []string `json:"containers_conf_modules,omitempty"`

	// Defaults to true when omitted.
	Copy *bool `json:"copy,omitempty"`

	Device string `json:"device,omitempty"`
	Driver string `json:"driver,omitempty"`
	Group  string `json:"group,omitempty"`
	Image  string `json:"image,omitempty"`
	Type   string `json:"type,omitempty"`
	User   string `json:"user,omitempty"`

	GlobalArgs []string `json:"global_args,omitempty"`
	Options    []string `json:"options,omitempty"`
	PodmanArgs []string `json:"podman_args,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	VolumeName string `json:"volume_name,omitempty"`

	// --- [Install] ---

	WantedBy []string `json:"wanted_by,omitempty"`
}

var _ Unit = (*VolumeUnit)(nil)

func (v *VolumeUnit) Name() string {
	return v.UnitName
}

func (v *VolumeUnit) Kind() Kind {
	return KindVolume
}

func (v *VolumeUnit) Validate() error {
	if v.UnitName == "" {
		return fmt.Errorf("%w: UnitName is required", ErrInvalidUnit)
	}
	if err := validateName(v.UnitName); err != nil {
		return fmt.Errorf("%w: UnitName %q is not valid: %v", ErrInvalidUnit, v.UnitName, err)
	}
	if v.Driver == "image" && v.Image == "" {
		return fmt.Errorf("%w: Image is required when Driver is %q", ErrInvalidUnit, v.Driver)
	}
	return nil
}

func (v *VolumeUnit) Render() (string, error) {
	if err := v.Validate(); err != nil {
		return "", err
	}

	wantedBy := v.WantedBy
	if len(wantedBy) == 0 {
		wantedBy = []string{"default.target"}
	}

	labels := labelLines(v.Labels)

	b := newIniBuilder()

	b.section("Unit").
		kv("Description", v.Description).
		kvSpaceJoined("After", v.After).
		kvSpaceJoined("Requires", v.Requires)

	b.section("Volume").
		kvList("ContainersConfModule", v.ContainersConfModules).
		kv("Copy", volumeCopyValue(v.Copy)).
		kv("Device", v.Device).
		kv("Driver", v.Driver).
		kvList("GlobalArgs", v.GlobalArgs).
		kv("Group", v.Group).
		kv("Image", v.Image).
		kvList("Label", labels).
		kvList("Options", v.Options).
		kvList("PodmanArgs", v.PodmanArgs).
		kv("Type", v.Type).
		kv("User", v.User).
		kv("VolumeName", v.VolumeName)

	b.section("Install").
		kvSpaceJoined("WantedBy", wantedBy)

	return b.String(), nil
}

func volumeCopyValue(copy *bool) string {
	if copy == nil {
		return ""
	}
	if *copy {
		return "true"
	}
	return "false"
}

func labelLines(labels map[string]string) []string {
	if len(labels) == 0 {
		return nil
	}

	names := make([]string, 0, len(labels))
	for name := range labels {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("%s=%s", name, quoteIfNeeded(labels[name])))
	}
	return lines
}
