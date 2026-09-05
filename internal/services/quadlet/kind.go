package quadlet

// Ref: https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html
type Kind string

const (
	KindContainer Kind = "container"
	KindVolume    Kind = "volume"
	KindNetwork   Kind = "network"
	KindKube      Kind = "kube"
	KindPod       Kind = "pod"
	KindBuild     Kind = "build"
	KindImage     Kind = "image"
)

var AllKinds = []Kind{
	KindContainer,
	KindVolume,
	KindNetwork,
	KindKube,
	KindPod,
	KindBuild,
	KindImage,
}

func (k Kind) Extension() string {
	return string(k)
}

func (k Kind) Valid() bool {
	for _, known := range AllKinds {
		if k == known {
			return true
		}
	}
	return false
}

func (k Kind) ServiceSuffix() string {
	switch k {
	case KindVolume:
		return "-volume"
	case KindNetwork:
		return "-network"
	default:
		return ""
	}
}

func (k Kind) ServiceName(name string) string {
	return name + k.ServiceSuffix() + ".service"
}
