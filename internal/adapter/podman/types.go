package podman

type podmanVersionResponse struct {
	Version    string `json:"Version"`
	APIVersion string `json:"ApiVersion"`
}
