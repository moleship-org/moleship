package quadlet

import (
	"errors"
	"strings"
	"testing"
)

func TestImageUnitValidate(t *testing.T) {
	t.Run("requires image", func(t *testing.T) {
		unit := &ImageUnit{UnitName: "base-image"}

		err := unit.Validate()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Fatalf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("rejects invalid policy", func(t *testing.T) {
		unit := &ImageUnit{
			UnitName: "base-image",
			Image:    "quay.io/centos/centos:latest",
			Policy:   "sometimes",
		}

		err := unit.Validate()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Fatalf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("rejects invalid retry delay", func(t *testing.T) {
		unit := &ImageUnit{
			UnitName:   "base-image",
			Image:      "quay.io/centos/centos:latest",
			RetryDelay: "later",
		}

		err := unit.Validate()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Fatalf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("rejects negative retry", func(t *testing.T) {
		retry := -1
		unit := &ImageUnit{
			UnitName: "base-image",
			Image:    "quay.io/centos/centos:latest",
			Retry:    &retry,
		}

		err := unit.Validate()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Fatalf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("accepts minimal unit", func(t *testing.T) {
		unit := &ImageUnit{
			UnitName: "base-image",
			Image:    "quay.io/centos/centos:latest",
		}
		if err := unit.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestImageUnitRender(t *testing.T) {
	allTags := true
	retry := 5
	tlsVerify := false
	unit := &ImageUnit{
		UnitName:              "base-image",
		Description:           "CentOS base image",
		After:                 []string{"network-online.target"},
		Requires:              []string{"registry.service"},
		AllTags:               &allTags,
		Arch:                  "aarch64",
		AuthFile:              "/etc/registry/auth.json",
		CertDir:               "/etc/registry/certs",
		ContainersConfModules: []string{"/etc/nvd.conf"},
		Creds:                 "myuser:mypassword",
		DecryptionKey:         "/etc/registry.key",
		GlobalArgs:            []string{"--log-level=debug"},
		Image:                 "docker-archive:/tmp/centos.tar",
		ImageTag:              "quay.io/centos/centos:latest",
		OS:                    "linux",
		PodmanArgs:            []string{"--platform=linux/arm64"},
		Policy:                "always",
		Retry:                 &retry,
		RetryDelay:            "10s",
		TLSVerify:             &tlsVerify,
		Variant:               "arm/v7",
		WantedBy:              []string{"multi-user.target"},
	}

	rendered, err := unit.Render()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	checks := []string{
		"[Unit]\n",
		"Description=CentOS base image\n",
		"After=network-online.target\n",
		"Requires=registry.service\n",
		"\n[Image]\n",
		"AllTags=true\n",
		"Arch=aarch64\n",
		"AuthFile=/etc/registry/auth.json\n",
		"CertDir=/etc/registry/certs\n",
		"ContainersConfModule=/etc/nvd.conf\n",
		"Creds=myuser:mypassword\n",
		"DecryptionKey=/etc/registry.key\n",
		"GlobalArgs=--log-level=debug\n",
		"Image=docker-archive:/tmp/centos.tar\n",
		"ImageTag=quay.io/centos/centos:latest\n",
		"OS=linux\n",
		"PodmanArgs=--platform=linux/arm64\n",
		"Policy=always\n",
		"Retry=5\n",
		"RetryDelay=10s\n",
		"TLSVerify=false\n",
		"Variant=arm/v7\n",
		"\n[Install]\n",
		"WantedBy=multi-user.target\n",
	}

	for _, check := range checks {
		if !strings.Contains(rendered, check) {
			t.Fatalf("rendered unit missing %q\n%s", check, rendered)
		}
	}
}

func TestImageUnitRenderDefaultWantedByAndOmitZeroValueFields(t *testing.T) {
	unit := &ImageUnit{
		UnitName: "base-image",
		Image:    "quay.io/centos/centos:latest",
	}

	rendered, err := unit.Render()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(rendered, "WantedBy=default.target\n") {
		t.Fatalf("expected default WantedBy, got:\n%s", rendered)
	}
	for _, key := range []string{"AllTags=", "Retry=", "TLSVerify="} {
		if strings.Contains(rendered, key) {
			t.Fatalf("did not expect %s when omitted, got:\n%s", key, rendered)
		}
	}
}
