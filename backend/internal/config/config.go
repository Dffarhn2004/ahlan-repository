package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	DataRoot        string
	Environment     string
	SessionHours    int
	HermesBaseURL   string
	HermesAPIKey    string
	HermesContainer string
	// HermesExecMode is how profile management reaches the Hermes container: "docker" (local CLI, development) or
	// "bridge" (Unix socket to the privileged bridge, production).
	HermesExecMode string
	// HermesBridgeSocket is the bridge's Unix socket path (bridge mode).
	HermesBridgeSocket string
	// HermesEventsURL is where the Hermes container posts conversation events; it must be reachable from inside it.
	HermesEventsURL string
}

// DefaultHermesEventsURL works with Docker Desktop (Windows/macOS) and with Linux when the container maps
// host.docker.internal to the host gateway (see deploy/hermes-bridge/README.md).
const DefaultHermesEventsURL = "http://host.docker.internal:8080/api/v1/internal/hermes/events"

func Load() Config {
	_ = godotenv.Load()
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	root := os.Getenv("DATA_ROOT")
	if root == "" {
		root = "./data"
	}
	hermesURL := os.Getenv("HERMES_BASE_URL")
	container := os.Getenv("HERMES_CONTAINER")
	if container == "" {
		container = "deploy-hermes-1"
	}
	// Production must not default to a Docker-capable mode: the application user has no Docker access there.
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("HERMES_EXEC_MODE")))
	if mode == "" {
		mode = "docker"
		if env == "production" {
			mode = "bridge"
		}
	}
	socket := strings.TrimSpace(os.Getenv("HERMES_BRIDGE_SOCKET"))
	if socket == "" {
		socket = "/run/ahlan-hermes/bridge.sock"
	}
	eventsURL := strings.TrimSpace(os.Getenv("HERMES_EVENTS_URL"))
	if eventsURL == "" {
		eventsURL = DefaultHermesEventsURL
	}
	return Config{HTTPAddr: addr, DatabaseURL: os.Getenv("DATABASE_URL"), DataRoot: root, Environment: env, SessionHours: 24, HermesBaseURL: hermesURL, HermesAPIKey: os.Getenv("HERMES_API_KEY"), HermesContainer: container,
		HermesExecMode: mode, HermesBridgeSocket: socket, HermesEventsURL: eventsURL}
}
