package hermes

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

// Control is everything the API and the worker may ask of the Hermes runtime. Handlers depend on this interface only,
// so they never know whether the work is done by the local Docker CLI (development) or by the privileged bridge
// (production, where the application user has no Docker access).
type Control interface {
	CreateParkedProfile(ctx context.Context, tenantID uuid.UUID, profileKey string) (Profile, error)
	ProfileAction(ctx context.Context, profileKey, action string) error
	ApplyProfileSettings(ctx context.Context, profileKey string, settings ProfileSettings) error
	ApplyClientExperience(ctx context.Context, profileKey string) error
	InstallWelcomeHook(ctx context.Context, profileKey, pesantren string) error
	InstallTelegramLogHook(ctx context.Context, profileKey, eventURL, eventKey string) error

	ConfigureTelegram(ctx context.Context, profileKey, token, mode string, allowedUserIDs []int64) error
	ConfigureWhatsApp(ctx context.Context, profileKey, phone, mode string, allowedNumbers []string) error
	RemoveChannelEnv(ctx context.Context, profileKey, prefix string) error

	PairWhatsApp(ctx context.Context, profileKey string, onEvent func(PairingEvent)) error
	StopWhatsAppPairing(ctx context.Context, profileKey string)
	RemoveWhatsAppSession(ctx context.Context, profileKey string) error

	SyncKnowledge(ctx context.Context, profileKey string, files []KnowledgeFile) error
	ExtractKnowledgeText(ctx context.Context, profileKey string, names []string, progress func(name string, ok bool)) (created int, failed []string)
	ReadyDocuments(ctx context.Context, profileKey string) (map[string]bool, error)

	// RestartGateway restarts the shared gateway process (all tenants' bots go offline briefly).
	RestartGateway(ctx context.Context) error
}

var _ Control = DockerClient{}

// Execution modes for NewControl.
const (
	ExecModeDocker = "docker" // call the Docker CLI directly; needs Docker access, meant for local development
	ExecModeBridge = "bridge" // talk to the local privileged bridge over a Unix socket; used in production
)

// ControlConfig selects and configures the Control implementation.
type ControlConfig struct {
	Mode         string
	Container    string // docker mode only; the bridge owns its own container name
	BridgeSocket string // bridge mode only
}

// NewControl builds the Control for cfg. An unknown mode is an error, never a silent fallback to Docker.
func NewControl(cfg ControlConfig) (Control, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
	case "", ExecModeDocker: // config.Load never leaves the mode empty; this only keeps zero-value configs (tests) working
		return DockerClient{Container: cfg.Container}, nil
	case ExecModeBridge:
		return NewBridgeClient(cfg.BridgeSocket)
	}
	return nil, fmt.Errorf("unsupported HERMES_EXEC_MODE %q (use %q or %q)", cfg.Mode, ExecModeDocker, ExecModeBridge)
}

// Unavailable returns a Control whose every operation fails with err. The API uses it when the execution mode is
// misconfigured, so it still serves everything that does not need Hermes instead of silently choosing another mode.
func Unavailable(err error) Control { return unavailable{err} }

type unavailable struct{ err error }

func (u unavailable) CreateParkedProfile(context.Context, uuid.UUID, string) (Profile, error) {
	return Profile{}, u.err
}
func (u unavailable) ProfileAction(context.Context, string, string) error { return u.err }
func (u unavailable) RestartGateway(context.Context) error                { return u.err }
func (u unavailable) ApplyProfileSettings(context.Context, string, ProfileSettings) error {
	return u.err
}
func (u unavailable) ApplyClientExperience(context.Context, string) error      { return u.err }
func (u unavailable) InstallWelcomeHook(context.Context, string, string) error { return u.err }
func (u unavailable) InstallTelegramLogHook(context.Context, string, string, string) error {
	return u.err
}
func (u unavailable) ConfigureTelegram(context.Context, string, string, string, []int64) error {
	return u.err
}
func (u unavailable) ConfigureWhatsApp(context.Context, string, string, string, []string) error {
	return u.err
}
func (u unavailable) RemoveChannelEnv(context.Context, string, string) error { return u.err }
func (u unavailable) PairWhatsApp(context.Context, string, func(PairingEvent)) error {
	return u.err
}
func (u unavailable) StopWhatsAppPairing(context.Context, string)                  {}
func (u unavailable) RemoveWhatsAppSession(context.Context, string) error          { return u.err }
func (u unavailable) SyncKnowledge(context.Context, string, []KnowledgeFile) error { return u.err }
func (u unavailable) ExtractKnowledgeText(_ context.Context, _ string, names []string, progress func(string, bool)) (int, []string) {
	for _, name := range names {
		if progress != nil {
			progress(name, false)
		}
	}
	return 0, names
}
func (u unavailable) ReadyDocuments(context.Context, string) (map[string]bool, error) {
	return nil, u.err
}

// readAll is a small helper for tests and the bridge to drain a bounded body.
func readAll(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w: body too large", ErrInvalidInput)
	}
	return data, nil
}
