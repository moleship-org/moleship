package quadlet

import (
	"errors"
	"strings"
	"testing"
)

func TestNetworkUnitValidate(t *testing.T) {
	t.Run("requires subnet when gateway is set", func(t *testing.T) {
		unit := &NetworkUnit{
			UnitName: "private0",
			Gateway:  []string{"10.10.0.1"},
		}

		err := unit.Validate()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Fatalf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("requires subnet when ip range is set", func(t *testing.T) {
		unit := &NetworkUnit{
			UnitName: "private0",
			IPRange:  []string{"10.10.0.128/25"},
		}

		err := unit.Validate()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Fatalf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("rejects invalid dns", func(t *testing.T) {
		unit := &NetworkUnit{
			UnitName: "private0",
			DNS:      []string{"not-an-ip"},
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
		unit := &NetworkUnit{UnitName: "private0"}
		if err := unit.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("accepts valid network configuration", func(t *testing.T) {
		disableDNS := true
		internal := true
		ipv6 := true
		deleteOnStop := true
		unit := &NetworkUnit{
			UnitName:              "private0",
			ContainersConfModules: []string{"/etc/nvd.conf"},
			DisableDNS:            &disableDNS,
			DNS:                   []string{"192.168.55.1", "none"},
			Driver:                "bridge",
			Gateway:               []string{"10.10.0.1"},
			GlobalArgs:            []string{"--log-level=debug"},
			InterfaceName:         "enp1",
			Internal:              &internal,
			IPAMDriver:            "host-local",
			IPRange:               []string{"10.10.0.128/25", "10.10.1.10-10.10.1.50"},
			IPv6:                  &ipv6,
			Labels:                map[string]string{"app": "private", "display name": "private net"},
			NetworkDeleteOnStop:   &deleteOnStop,
			NetworkName:           "private0",
			Options:               []string{"isolate=true", "parent=enp1"},
			PodmanArgs:            []string{"--dns=192.168.55.1"},
			Subnet:                []string{"10.10.0.0/16", "fd00:dead:beef::/64"},
		}
		if err := unit.Validate(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestNetworkUnitRender(t *testing.T) {
	disableDNS := true
	internal := true
	ipv6 := true
	deleteOnStop := true
	unit := &NetworkUnit{
		UnitName:              "private0",
		Description:           "Private bridge network",
		After:                 []string{"network-online.target"},
		Requires:              []string{"container-a.service"},
		ContainersConfModules: []string{"/etc/nvd.conf"},
		DisableDNS:            &disableDNS,
		DNS:                   []string{"192.168.55.1", "none"},
		Driver:                "bridge",
		Gateway:               []string{"10.10.0.1"},
		GlobalArgs:            []string{"--log-level=debug"},
		InterfaceName:         "enp1",
		Internal:              &internal,
		IPAMDriver:            "host-local",
		IPRange:               []string{"10.10.0.128/25", "10.10.1.10-10.10.1.50"},
		IPv6:                  &ipv6,
		Labels:                map[string]string{"app": "private", "display name": "private net"},
		NetworkDeleteOnStop:   &deleteOnStop,
		NetworkName:           "private0",
		Options:               []string{"isolate=true", "parent=enp1"},
		PodmanArgs:            []string{"--dns=192.168.55.1"},
		Subnet:                []string{"10.10.0.0/16", "fd00:dead:beef::/64"},
		WantedBy:              []string{"multi-user.target"},
	}

	rendered, err := unit.Render()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	checks := []string{
		"[Unit]\n",
		"Description=Private bridge network\n",
		"After=network-online.target\n",
		"Requires=container-a.service\n",
		"\n[Network]\n",
		"ContainersConfModule=/etc/nvd.conf\n",
		"DisableDNS=true\n",
		"DNS=192.168.55.1\n",
		"DNS=none\n",
		"Driver=bridge\n",
		"Gateway=10.10.0.1\n",
		"GlobalArgs=--log-level=debug\n",
		"InterfaceName=enp1\n",
		"Internal=true\n",
		"IPAMDriver=host-local\n",
		"IPRange=10.10.0.128/25\n",
		"IPRange=10.10.1.10-10.10.1.50\n",
		"IPv6=true\n",
		"Label=app=private\n",
		"Label=display name=\"private net\"\n",
		"NetworkDeleteOnStop=true\n",
		"NetworkName=private0\n",
		"Options=isolate=true\n",
		"Options=parent=enp1\n",
		"PodmanArgs=--dns=192.168.55.1\n",
		"Subnet=10.10.0.0/16\n",
		"Subnet=fd00:dead:beef::/64\n",
		"\n[Install]\n",
		"WantedBy=multi-user.target\n",
	}

	for _, check := range checks {
		if !strings.Contains(rendered, check) {
			t.Fatalf("rendered unit missing %q\n%s", check, rendered)
		}
	}
}

func TestNetworkUnitRenderDefaultWantedByAndOmitBoolFields(t *testing.T) {
	unit := &NetworkUnit{UnitName: "private0"}

	rendered, err := unit.Render()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(rendered, "WantedBy=default.target\n") {
		t.Fatalf("expected default WantedBy, got:\n%s", rendered)
	}
	for _, key := range []string{"DisableDNS=", "Internal=", "IPv6=", "NetworkDeleteOnStop="} {
		if strings.Contains(rendered, key) {
			t.Fatalf("did not expect %s when omitted, got:\n%s", key, rendered)
		}
	}
}
