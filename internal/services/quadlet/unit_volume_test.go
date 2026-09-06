package quadlet

import (
	"errors"
	"strings"
	"testing"
)

func TestVolumeUnitValidate(t *testing.T) {
	t.Run("requires image driver image backing", func(t *testing.T) {
		unit := &VolumeUnit{
			UnitName: "cache",
			Driver:   "image",
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
		unit := &VolumeUnit{UnitName: "cache"}
		if err := unit.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestVolumeUnitRender(t *testing.T) {
	copyValue := false
	unit := &VolumeUnit{
		UnitName:              "html",
		Description:           "App data volume",
		After:                 []string{"network-online.target"},
		Requires:              []string{"container-a.service"},
		ContainersConfModules: []string{"/etc/nvd.conf"},
		Copy:                  &copyValue,
		Device:                "tmpfs",
		Driver:                "image",
		GlobalArgs:            []string{"--log-level=debug"},
		Group:                 "1000",
		Image:                 "quay.io/centos/centos:latest",
		Labels:                map[string]string{"app": "data", "display name": "app data"},
		Options:               []string{"o=nodev"},
		PodmanArgs:            []string{"--driver=image"},
		Type:                  "tmpfs",
		User:                  "1000",
		VolumeName:            "mydata",
		WantedBy:              []string{"multi-user.target"},
	}

	rendered, err := unit.Render()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	checks := []string{
		"[Unit]\n",
		"Description=App data volume\n",
		"After=network-online.target\n",
		"Requires=container-a.service\n",
		"\n[Volume]\n",
		"ContainersConfModule=/etc/nvd.conf\n",
		"Copy=false\n",
		"Device=tmpfs\n",
		"Driver=image\n",
		"GlobalArgs=--log-level=debug\n",
		"Group=1000\n",
		"Image=quay.io/centos/centos:latest\n",
		"Label=app=data\n",
		"Label=display name=\"app data\"\n",
		"Options=o=nodev\n",
		"PodmanArgs=--driver=image\n",
		"Type=tmpfs\n",
		"User=1000\n",
		"VolumeName=mydata\n",
		"\n[Install]\n",
		"WantedBy=multi-user.target\n",
	}

	for _, check := range checks {
		if !strings.Contains(rendered, check) {
			t.Fatalf("rendered unit missing %q\n%s", check, rendered)
		}
	}
}

func TestVolumeUnitRenderDefaultWantedBy(t *testing.T) {
	unit := &VolumeUnit{UnitName: "data"}

	rendered, err := unit.Render()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(rendered, "WantedBy=default.target\n") {
		t.Fatalf("expected default WantedBy, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "Copy=") {
		t.Fatalf("did not expect Copy line when omitted, got:\n%s", rendered)
	}
}
