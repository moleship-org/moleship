package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// Env variable names for configuration
	EnvHostUser             = "MOLESHIP_HOST_USER"
	EnvHostUID              = "MOLESHIP_HOST_UID"
	EnvPort                 = "MOLESHIP_PORT"
	EnvLogLevel             = "MOLESHIP_LOG_LEVEL"
	EnvMode                 = "MOLESHIP_MODE"
	EnvConfigHome           = "MOLESHIP_CONFIG_HOME"
	EnvCacheHome            = "MOLESHIP_CACHE_HOME"
	EnvDataHome             = "MOLESHIP_DATA_HOME"
	EnvPodmanSocket         = "MOLESHIP_PODMAN_SOCKET"
	EnvPodmanPath           = "MOLESHIP_PODMAN_PATH"
	EnvEnableLibpodProxy    = "MOLESHIP_ENABLE_LIBPOD_PROXY"
	EnvSystemctlPath        = "MOLESHIP_SYSTEMCTL_PATH"
	EnvQuadletHome          = "MOLESHIP_QUADLET_HOME"
	EnvJWTSecret            = "MOLESHIP_JWT_SECRET"
	EnvAllowedOrigins       = "MOLESHIP_ALLOWED_ORIGINS"
	EnvPublicRateLimit      = "MOLESHIP_PUBLIC_RATE_LIMIT"
	EnvPublicBurstLimit     = "MOLESHIP_PUBLIC_BURST_LIMIT"
	EnvPublicIPHeaderLookup = "MOLESHIP_PUBLIC_IP_HEADER_LOOKUP"

	// Rootless defaults
	DefaultRootlessConfigHome  = ".config/moleship"
	DefaultRootlessCacheHome   = ".cache/moleship"
	DefaultRootlessDataHome    = ".local/share/moleship"
	DefaultRootlessQuadletHome = ".config/containers/systemd"

	// Rootful defaults

	DefaultRootfulConfigHome  = "/etc/moleship"
	DefaultRootfulCacheHome   = "/var/cache/moleship"
	DefaultRootfulDataHome    = "/var/lib/moleship"
	DefaultRootfulQuadletHome = "/etc/containers/systemd"
)

var (
	// Default listening port for the application
	PORT uint16 = 5000

	// Default host username
	HOST_USER string = ""

	// Default host identifier
	HOST_UID string = ""

	// Default log level for the application used by slog package. Possible values: "debug", "info", "warn", "error"
	LOG_LEVEL string = "info"

	// Default mode for the application. Possible values: "debug", "release"
	MODE string = "debug"

	// Default directory path for storing configuration files
	CONFIG_HOME string = ""

	// Default directory path for storing cache files
	CACHE_HOME string = ""

	// Default directory path for storing data files
	DATA_HOME string = ""

	// Default path to the Podman unix socket for communication with the Podman service (libpod)
	PODMAN_SOCKET string = "/var/run/podman.sock"

	// Default path to the Podman executable
	PODMAN_PATH string = "/usr/bin/podman"

	// Default flag indicating whether the raw libpod proxy endpoint is enabled
	ENABLE_LIBPOD_PROXY bool = false

	// Default path to the systemctl executable
	SYSTEMCTL_PATH string = "/usr/bin/systemctl"

	// Default directory path for storing quadlet files
	QUADLET_HOME string = ""

	// Default flag indicating whether the application is running in rootful mode
	ROOTFUL bool = false

	// Default podman version for the connection
	PODMAN_VERSION string = "5.0.0"

	// Default json web token secret
	JWT_SECRET []byte = []byte("mysecret")

	// Default allowed origins. Empty by default; configure explicitly for browser access.
	ALLOWED_ORIGINS []string = []string{}

	// Default rate limit for the public endpoints
	PUBLIC_RATE_LIMIT float64 = 15.0 // req/sec

	// Default base burst for the public endpoints
	PUBLIC_BURST_LIMIT int = 3

	// Default IP header lookup for the public endpoint ratelimit
	// "RemoteAddr", "X-Forwarded-For", "CF-Connecting-IP"
	PUBLIC_IP_HEADER_LOOKUP string = "RemoteAddr"
)

func IsDebugMode() bool {
	return strings.ToLower(MODE) == "debug" || MODE == ""
}

func IsReleaseMode() bool {
	return strings.ToLower(MODE) == "release"
}

func getEnvOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func init() {
	rootfulEnv := getEnvOrDefault("MOLESHIP_ROOTFUL", "false")
	if rootful, err := strconv.ParseBool(rootfulEnv); err == nil {
		ROOTFUL = rootful
	}

	if ROOTFUL {
		handleRootfulConfig()
	} else {
		handleRootlessConfig()
	}

	p, err := strconv.ParseUint(getEnvOrDefault(EnvPort, "5000"), 10, 16)
	if err == nil {
		PORT = uint16(p)
	}

	HOST_USER = getEnvOrDefault(EnvHostUser, "unknown")
	HOST_UID = getEnvOrDefault(EnvHostUID, "")
	LOG_LEVEL = getEnvOrDefault(EnvLogLevel, "info")
	MODE = getEnvOrDefault(EnvMode, "debug")
	PODMAN_SOCKET = getEnvOrDefault(EnvPodmanSocket, "/run/user/1000/podman/podman.sock")
	PODMAN_PATH = getEnvOrDefault(EnvPodmanPath, "/usr/bin/podman")
	if enabled, err := strconv.ParseBool(getEnvOrDefault(EnvEnableLibpodProxy, "false")); err == nil {
		ENABLE_LIBPOD_PROXY = enabled
	}
	SYSTEMCTL_PATH = getEnvOrDefault(EnvSystemctlPath, "/usr/bin/systemctl")

	secret, err := getOrCreateJWTSecret(DATA_HOME)
	if err != nil {
		panic("provide a MOLESHIP_JWT_SECRET or create a jwt_secret.key file at MOLESHIP_DATA_HOME")
	}
	JWT_SECRET = secret

	allowedOrg := strings.TrimSpace(getEnvOrDefault(EnvAllowedOrigins, ""))
	if allowedOrg != "" {
		ALLOWED_ORIGINS = strings.Split(allowedOrg, ",")
	}

	rateLimit := getEnvOrDefault(EnvPublicRateLimit, "")
	if rateLimit != "" {
		rl, err := strconv.ParseFloat(rateLimit, 64)
		if err != nil {
			panic("MOLESHIP_PUBLIC_RATE_LIMIT must be a valid float")
		}
		PUBLIC_RATE_LIMIT = rl
	}

	burstLimit := getEnvOrDefault(EnvPublicBurstLimit, "")
	if burstLimit != "" {
		bl, err := strconv.Atoi(burstLimit)
		if err != nil {
			panic("MOLESHIP_PUBLIC_BURST_LIMIT must be a valid integer")
		}
		PUBLIC_BURST_LIMIT = bl
	}

	PUBLIC_IP_HEADER_LOOKUP = getEnvOrDefault(EnvPublicIPHeaderLookup, PUBLIC_IP_HEADER_LOOKUP)

	dirs := []string{CONFIG_HOME, CACHE_HOME, DATA_HOME, QUADLET_HOME}
	for _, dir := range dirs {
		if err := makeDirIfNotExists(dir); err != nil {
			panic("Failed to create directory: " + dir + " - " + err.Error())
		}
	}
}

func handleRootfulConfig() {
	CONFIG_HOME = getEnvOrDefault(EnvConfigHome, DefaultRootfulConfigHome)
	CACHE_HOME = getEnvOrDefault(EnvCacheHome, DefaultRootfulCacheHome)
	DATA_HOME = getEnvOrDefault(EnvDataHome, DefaultRootfulDataHome)
	QUADLET_HOME = getEnvOrDefault(EnvQuadletHome, DefaultRootfulQuadletHome)
}

func handleRootlessConfig() {
	homeDir := filepath.Join("home", HOST_USER)
	if err := makeDirIfNotExists(homeDir); err != nil {
		panic("Failed to create the home directory: " + homeDir + " - " + err.Error())
	}

	CONFIG_HOME = getEnvOrDefault(EnvConfigHome, filepath.Join(homeDir, ".config", "moleship"))
	CACHE_HOME = getEnvOrDefault(EnvCacheHome, filepath.Join(homeDir, ".cache", "moleship"))
	DATA_HOME = getEnvOrDefault(EnvDataHome, filepath.Join(homeDir, ".local", "share", "moleship"))
	QUADLET_HOME = getEnvOrDefault(EnvQuadletHome, filepath.Join(homeDir, ".config", "containers", "systemd"))
}

func makeDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func getOrCreateJWTSecret(dataHome string) ([]byte, error) {
	secretPath := filepath.Join(dataHome, "jwt_secret.key")

	if secret, err := os.ReadFile(secretPath); err == nil {
		return secret, nil
	}

	if err := os.MkdirAll(dataHome, 0700); err != nil {
		return nil, err
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	key := base64.RawStdEncoding.EncodeToString(secret)

	err := os.WriteFile(secretPath, []byte(key), 0600)
	if err != nil {
		return nil, err
	}

	return secret, nil
}
