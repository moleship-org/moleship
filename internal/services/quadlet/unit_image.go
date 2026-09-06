package quadlet

import (
	"fmt"
	"time"
)

type ImageUnit struct {
	UnitName string `json:"name"`

	Description string `json:"description,omitempty"`

	After    []string `json:"after,omitempty"`
	Requires []string `json:"requires,omitempty"`

	// --- [Image] ---

	AllTags *bool `json:"all_tags,omitempty"`

	Arch     string `json:"arch,omitempty"`
	AuthFile string `json:"auth_file,omitempty"`
	CertDir  string `json:"cert_dir,omitempty"`

	ContainersConfModules []string `json:"containers_conf_modules,omitempty"`

	Creds         string `json:"creds,omitempty"`
	DecryptionKey string `json:"decryption_key,omitempty"`

	GlobalArgs []string `json:"global_args,omitempty"`

	Image    string `json:"image"`
	ImageTag string `json:"image_tag,omitempty"`

	OS string `json:"os,omitempty"`

	PodmanArgs []string `json:"podman_args,omitempty"`

	Policy string `json:"policy,omitempty"`

	Retry      *int   `json:"retry,omitempty"`
	RetryDelay string `json:"retry_delay,omitempty"`

	TLSVerify *bool `json:"tls_verify,omitempty"`

	Variant string `json:"variant,omitempty"`

	// --- [Install] ---

	WantedBy []string `json:"wanted_by,omitempty"`
}

var _ Unit = (*ImageUnit)(nil)

func (i *ImageUnit) Name() string {
	return i.UnitName
}

func (i *ImageUnit) Kind() Kind {
	return KindImage
}

func (i *ImageUnit) Validate() error {
	if i.UnitName == "" {
		return fmt.Errorf("%w: UnitName is required", ErrInvalidUnit)
	}
	if err := validateName(i.UnitName); err != nil {
		return fmt.Errorf("%w: UnitName %q is not valid: %v", ErrInvalidUnit, i.UnitName, err)
	}
	if i.Image == "" {
		return fmt.Errorf("%w: Image is required", ErrInvalidUnit)
	}
	if i.Policy != "" {
		switch i.Policy {
		case "always", "missing", "never", "newer":
		default:
			return fmt.Errorf("%w: invalid Policy %q", ErrInvalidUnit, i.Policy)
		}
	}
	if i.Retry != nil && *i.Retry < 0 {
		return fmt.Errorf("%w: Retry must be greater than or equal to 0", ErrInvalidUnit)
	}
	if i.RetryDelay != "" {
		if _, err := time.ParseDuration(i.RetryDelay); err != nil {
			return fmt.Errorf("%w: invalid RetryDelay %q: %v", ErrInvalidUnit, i.RetryDelay, err)
		}
	}
	return nil
}

func (i *ImageUnit) Render() (string, error) {
	if err := i.Validate(); err != nil {
		return "", err
	}

	wantedBy := i.WantedBy
	if len(wantedBy) == 0 {
		wantedBy = []string{"default.target"}
	}

	b := newIniBuilder()

	b.section("Unit").
		kv("Description", i.Description).
		kvSpaceJoined("After", i.After).
		kvSpaceJoined("Requires", i.Requires)

	b.section("Image").
		kv("AllTags", boolValue(i.AllTags)).
		kv("Arch", i.Arch).
		kv("AuthFile", i.AuthFile).
		kv("CertDir", i.CertDir).
		kvList("ContainersConfModule", i.ContainersConfModules).
		kv("Creds", i.Creds).
		kv("DecryptionKey", i.DecryptionKey).
		kvList("GlobalArgs", i.GlobalArgs).
		kv("Image", i.Image).
		kv("ImageTag", i.ImageTag).
		kv("OS", i.OS).
		kvList("PodmanArgs", i.PodmanArgs).
		kv("Policy", i.Policy).
		kv("Retry", intValue(i.Retry)).
		kv("RetryDelay", i.RetryDelay).
		kv("TLSVerify", boolValue(i.TLSVerify)).
		kv("Variant", i.Variant)

	b.section("Install").
		kvSpaceJoined("WantedBy", wantedBy)

	return b.String(), nil
}

func intValue(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}
