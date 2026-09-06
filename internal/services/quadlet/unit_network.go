package quadlet

import (
	"fmt"
	"net"
)

type NetworkUnit struct {
	UnitName string `json:"name"`

	Description string `json:"description,omitempty"`

	After    []string `json:"after,omitempty"`
	Requires []string `json:"requires,omitempty"`

	// --- [Network] ---

	ContainersConfModules []string `json:"containers_conf_modules,omitempty"`

	DisableDNS *bool    `json:"disable_dns,omitempty"`
	DNS        []string `json:"dns,omitempty"`
	Driver     string   `json:"driver,omitempty"`
	Gateway    []string `json:"gateway,omitempty"`
	GlobalArgs []string `json:"global_args,omitempty"`

	InterfaceName string   `json:"interface_name,omitempty"`
	Internal      *bool    `json:"internal,omitempty"`
	IPAMDriver    string   `json:"ipam_driver,omitempty"`
	IPRange       []string `json:"ip_range,omitempty"`
	IPv6          *bool    `json:"ipv6,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	NetworkDeleteOnStop *bool    `json:"network_delete_on_stop,omitempty"`
	NetworkName         string   `json:"network_name,omitempty"`
	Options             []string `json:"options,omitempty"`
	PodmanArgs          []string `json:"podman_args,omitempty"`
	Subnet              []string `json:"subnet,omitempty"`

	// --- [Install] ---

	WantedBy []string `json:"wanted_by,omitempty"`
}

var _ Unit = (*NetworkUnit)(nil)

func (n *NetworkUnit) Name() string {
	return n.UnitName
}

func (n *NetworkUnit) Kind() Kind {
	return KindNetwork
}

func (n *NetworkUnit) Validate() error {
	if n.UnitName == "" {
		return fmt.Errorf("%w: UnitName is required", ErrInvalidUnit)
	}
	if err := validateName(n.UnitName); err != nil {
		return fmt.Errorf("%w: UnitName %q is not valid: %v", ErrInvalidUnit, n.UnitName, err)
	}

	for _, dns := range n.DNS {
		if err := validateNetworkIPOrNone(dns); err != nil {
			return fmt.Errorf("%w: invalid DNS %q: %v", ErrInvalidUnit, dns, err)
		}
	}
	for _, gateway := range n.Gateway {
		if err := validateNetworkIP(gateway); err != nil {
			return fmt.Errorf("%w: invalid Gateway %q: %v", ErrInvalidUnit, gateway, err)
		}
	}
	for _, subnet := range n.Subnet {
		if err := validateNetworkCIDR(subnet); err != nil {
			return fmt.Errorf("%w: invalid Subnet %q: %v", ErrInvalidUnit, subnet, err)
		}
	}
	for _, ipRange := range n.IPRange {
		if err := validateNetworkRange(ipRange); err != nil {
			return fmt.Errorf("%w: invalid IPRange %q: %v", ErrInvalidUnit, ipRange, err)
		}
	}

	if len(n.Gateway) > 0 && len(n.Subnet) == 0 {
		return fmt.Errorf("%w: Subnet is required when Gateway is set", ErrInvalidUnit)
	}
	if len(n.IPRange) > 0 && len(n.Subnet) == 0 {
		return fmt.Errorf("%w: Subnet is required when IPRange is set", ErrInvalidUnit)
	}

	return nil
}

func (n *NetworkUnit) Render() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}

	wantedBy := n.WantedBy
	if len(wantedBy) == 0 {
		wantedBy = []string{"default.target"}
	}

	labels := labelLines(n.Labels)

	b := newIniBuilder()

	b.section("Unit").
		kv("Description", n.Description).
		kvSpaceJoined("After", n.After).
		kvSpaceJoined("Requires", n.Requires)

	b.section("Network").
		kvList("ContainersConfModule", n.ContainersConfModules).
		kv("DisableDNS", boolValue(n.DisableDNS)).
		kvList("DNS", n.DNS).
		kv("Driver", n.Driver).
		kvList("Gateway", n.Gateway).
		kvList("GlobalArgs", n.GlobalArgs).
		kv("InterfaceName", n.InterfaceName).
		kv("Internal", boolValue(n.Internal)).
		kv("IPAMDriver", n.IPAMDriver).
		kvList("IPRange", n.IPRange).
		kv("IPv6", boolValue(n.IPv6)).
		kvList("Label", labels).
		kv("NetworkDeleteOnStop", boolValue(n.NetworkDeleteOnStop)).
		kv("NetworkName", n.NetworkName).
		kvList("Options", n.Options).
		kvList("PodmanArgs", n.PodmanArgs).
		kvList("Subnet", n.Subnet)

	b.section("Install").
		kvSpaceJoined("WantedBy", wantedBy)

	return b.String(), nil
}

func boolValue(value *bool) string {
	if value == nil {
		return ""
	}
	if *value {
		return "true"
	}
	return "false"
}

func validateNetworkIPOrNone(value string) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}
	if value == "none" {
		return nil
	}
	return validateNetworkIP(value)
}

func validateNetworkIP(value string) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}
	if ip := net.ParseIP(value); ip == nil {
		return fmt.Errorf("must be a valid IP address")
	}
	return nil
}

func validateNetworkCIDR(value string) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}
	if _, _, err := net.ParseCIDR(value); err != nil {
		return fmt.Errorf("must be a valid CIDR")
	}
	return nil
}

func validateNetworkRange(value string) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}
	if _, _, err := net.ParseCIDR(value); err == nil {
		return nil
	}

	start, end, ok := splitRange(value)
	if !ok {
		return fmt.Errorf("must be a valid CIDR or startIP-endIP range")
	}
	if net.ParseIP(start) == nil || net.ParseIP(end) == nil {
		return fmt.Errorf("must be a valid CIDR or startIP-endIP range")
	}
	return nil
}

func splitRange(value string) (string, string, bool) {
	for i := 0; i < len(value); i++ {
		if value[i] != '-' {
			continue
		}
		return value[:i], value[i+1:], true
	}
	return "", "", false
}
