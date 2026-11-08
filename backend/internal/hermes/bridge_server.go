package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Executor is what the bridge is allowed to do. DockerClient satisfies it; tests substitute a fake.
type Executor interface {
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
	PutKnowledgeFile(ctx context.Context, profileKey, name string, body io.Reader) error
	PruneKnowledge(ctx context.Context, profileKey string, names []string) error
	ExtractOne(ctx context.Context, profileKey, name string) (bool, error)
	ReadyDocuments(ctx context.Context, profileKey string) (map[string]bool, error)
	RestartGateway(ctx context.Context) error
}

var _ Executor = DockerClient{}

// DefaultGatewayRestartInterval is the shortest time between two shared-gateway restarts. A restart drops every
// tenant's bot for about a minute, so a runaway or repeated caller must not be able to chain them.
const DefaultGatewayRestartInterval = 2 * time.Minute

// errRateLimited marks an operation refused because it was repeated too soon; it maps to HTTP 429 / BUSY.
var errRateLimited = errors.New("rate limited")

// BridgeServerConfig configures the bridge side.
type BridgeServerConfig struct {
	// Container is the only container the bridge will ever touch. It is fixed at start-up, never taken from a request.
	Container string
	// AllowedEventURL, when set, is the only URL a log hook may call back. Strongly recommended in production.
	AllowedEventURL string
	// AllowedUIDs, when non-empty, restricts callers by the Unix peer credentials of the socket connection.
	AllowedUIDs []uint32
	// MaxConcurrent bounds simultaneous operations so a runaway caller cannot exhaust the host.
	MaxConcurrent int
	// GatewayRestartInterval is the minimum gap between shared-gateway restarts (default DefaultGatewayRestartInterval).
	GatewayRestartInterval time.Duration
	Logger                 *slog.Logger
}

// BridgeServer exposes a fixed allow-list of Hermes operations over HTTP (meant to be served on a Unix socket).
type BridgeServer struct {
	cfg  BridgeServerConfig
	exec Executor
	slot chan struct{}
	log  *slog.Logger
	mux  *http.ServeMux

	restartMu   sync.Mutex
	lastRestart time.Time
}

const maxJSONBody = 2 << 20 // largest JSON request: a persona of up to MaxSoulBytes plus tenant info

// NewBridgeServer validates the configuration and wires the allow-listed routes to exec.
func NewBridgeServer(cfg BridgeServerConfig, exec Executor) (*BridgeServer, error) {
	if err := ValidateContainerName(cfg.Container); err != nil {
		return nil, err
	}
	if cfg.AllowedEventURL != "" {
		if err := ValidateEventURL(cfg.AllowedEventURL); err != nil {
			return nil, err
		}
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 16
	}
	if cfg.GatewayRestartInterval <= 0 {
		cfg.GatewayRestartInterval = DefaultGatewayRestartInterval
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	s := &BridgeServer{cfg: cfg, exec: exec, slot: make(chan struct{}, cfg.MaxConcurrent), log: cfg.Logger, mux: http.NewServeMux()}
	s.routes()
	return s, nil
}

// handler is a route body. profile is the validated profile key, kept for the audit log.
type handler func(w http.ResponseWriter, r *http.Request) (profile string, err error)

func (s *BridgeServer) route(method, path string, h handler) {
	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeError(w, http.StatusMethodNotAllowed, codeForbidden)
			return
		}
		if uids := s.cfg.AllowedUIDs; len(uids) > 0 {
			uid, ok := r.Context().Value(peerUIDKey{}).(uint32)
			if !ok || !uidAllowed(uids, uid) {
				s.log.Warn("bridge caller rejected", "op", path)
				writeError(w, http.StatusForbidden, codeForbidden)
				return
			}
		}
		select {
		case s.slot <- struct{}{}:
			defer func() { <-s.slot }()
		default:
			writeError(w, http.StatusTooManyRequests, codeBusy)
			return
		}
		started := time.Now()
		profile, err := h(w, r)
		status := "ok"
		if err != nil {
			status = "error"
		}
		// Only the operation, the profile and the outcome are logged: never bodies, tokens or command output.
		s.log.Info("bridge op", "op", path, "profile", profile, "result", status, "ms", time.Since(started).Milliseconds())
		if err != nil {
			s.log.Warn("bridge op failed", "op", path, "profile", profile, "error", sanitizeError(err))
			if errors.Is(err, ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, codeInvalidInput)
			} else if errors.Is(err, errRateLimited) {
				writeError(w, http.StatusTooManyRequests, codeBusy)
			} else {
				writeError(w, http.StatusBadGateway, codeExecFailed)
			}
		}
	})
}

func (s *BridgeServer) routes() {
	// Each handler validates every field before it reaches the executor; DockerClient validates again (defence in depth).
	s.route(http.MethodGet, bridgeHealth, func(w http.ResponseWriter, r *http.Request) (string, error) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		// The container name is fixed in configuration; this proves it exists without exposing anything.
		if err := DockerClientInspect(ctx, s.cfg.Container); err != nil {
			return "", err
		}
		return "", writeJSON(w, map[string]any{"status": "ok"})
	})
	s.route(http.MethodPost, bridgeProfileCreate, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqCreate
		if err := decode(r, &in); err != nil {
			return "", err
		}
		tenant, err := uuid.Parse(in.TenantID)
		if err != nil || tenant == uuid.Nil {
			return in.ProfileKey, fmt.Errorf("%w: tenant id", ErrInvalidInput)
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		_, err = s.exec.CreateParkedProfile(r.Context(), tenant, in.ProfileKey)
		return in.ProfileKey, finish(w, err)
	})
	s.route(http.MethodPost, bridgeProfileAction, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqAction
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if !profileActions[in.Action] {
			return in.ProfileKey, fmt.Errorf("%w: action", ErrInvalidInput)
		}
		return in.ProfileKey, finish(w, s.exec.ProfileAction(r.Context(), in.ProfileKey, in.Action))
	})
	s.route(http.MethodPost, bridgeProfileSettings, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqSettings
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if len(in.SoulMD) > MaxSoulBytes || len(in.Description) > MaxDescriptionBytes || len(in.TenantInfo) > MaxSoulBytes || len(in.EnabledSkills) > 200 || len(in.WhatsAppAdmins) > 20 || len(in.HandoffURL) > 512 || len(in.HandoffKey) > 512 {
			return in.ProfileKey, fmt.Errorf("%w: profile settings too large", ErrInvalidInput)
		}
		settings := ProfileSettings{Description: in.Description, SoulMD: in.SoulMD, EnabledSkills: in.EnabledSkills, TenantInfo: in.TenantInfo, WhatsAppGeneral: in.WhatsAppGeneral, WhatsAppAdmins: in.WhatsAppAdmins, HandoffURL: in.HandoffURL, HandoffKey: in.HandoffKey}
		return in.ProfileKey, finish(w, s.exec.ApplyProfileSettings(r.Context(), in.ProfileKey, settings))
	})
	s.route(http.MethodPost, bridgeClientExperience, func(w http.ResponseWriter, r *http.Request) (string, error) {
		key, err := decodeProfile(r)
		if err != nil {
			return key, err
		}
		return key, finish(w, s.exec.ApplyClientExperience(r.Context(), key))
	})
	s.route(http.MethodPost, bridgeWelcomeHook, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqWelcome
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		return in.ProfileKey, finish(w, s.exec.InstallWelcomeHook(r.Context(), in.ProfileKey, in.Pesantren))
	})
	s.route(http.MethodPost, bridgeLogHook, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqLogHook
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if err := ValidateEventURL(in.EventURL); err != nil {
			return in.ProfileKey, err
		}
		if err := ValidateEventKey(in.EventKey); err != nil {
			return in.ProfileKey, err
		}
		if s.cfg.AllowedEventURL != "" && in.EventURL != s.cfg.AllowedEventURL {
			return in.ProfileKey, fmt.Errorf("%w: event URL not allowed", ErrInvalidInput)
		}
		return in.ProfileKey, finish(w, s.exec.InstallTelegramLogHook(r.Context(), in.ProfileKey, in.EventURL, in.EventKey))
	})
	s.route(http.MethodPost, bridgeTelegram, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqTelegram
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if len(in.Mode) > 32 || len(in.AllowedUserIDs) > 500 {
			return in.ProfileKey, fmt.Errorf("%w: allowed users", ErrInvalidInput)
		}
		if err := ValidateTelegramToken(in.Token); err != nil {
			return in.ProfileKey, err
		}
		return in.ProfileKey, finish(w, s.exec.ConfigureTelegram(r.Context(), in.ProfileKey, in.Token, in.Mode, in.AllowedUserIDs))
	})
	s.route(http.MethodPost, bridgeWhatsApp, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqWhatsApp
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if len(in.Mode) > 32 || len(in.AllowedNumbers) > 500 {
			return in.ProfileKey, fmt.Errorf("%w: allowed numbers", ErrInvalidInput)
		}
		if err := ValidateWhatsAppNumber(in.Phone); err != nil {
			return in.ProfileKey, err
		}
		for _, number := range in.AllowedNumbers {
			if err := ValidateWhatsAppNumber(number); err != nil {
				return in.ProfileKey, err
			}
		}
		return in.ProfileKey, finish(w, s.exec.ConfigureWhatsApp(r.Context(), in.ProfileKey, in.Phone, in.Mode, in.AllowedNumbers))
	})
	s.route(http.MethodPost, bridgeChannelRemove, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqChannelRemove
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if err := ValidateChannelPrefix(in.Prefix); err != nil {
			return in.ProfileKey, err
		}
		return in.ProfileKey, finish(w, s.exec.RemoveChannelEnv(r.Context(), in.ProfileKey, in.Prefix))
	})
	s.route(http.MethodPost, bridgePair, s.pair)
	s.route(http.MethodPost, bridgePairStop, func(w http.ResponseWriter, r *http.Request) (string, error) {
		key, err := decodeProfile(r)
		if err != nil {
			return key, err
		}
		s.exec.StopWhatsAppPairing(r.Context(), key)
		return key, finish(w, nil)
	})
	s.route(http.MethodPost, bridgeSessionRemove, func(w http.ResponseWriter, r *http.Request) (string, error) {
		key, err := decodeProfile(r)
		if err != nil {
			return key, err
		}
		return key, finish(w, s.exec.RemoveWhatsAppSession(r.Context(), key))
	})
	s.route(http.MethodPut, bridgeKnowledgePut, func(w http.ResponseWriter, r *http.Request) (string, error) {
		query := r.URL.Query()
		key, name := query.Get("profile_key"), query.Get("name")
		if err := ValidateProfileKey(key); err != nil {
			return "", err
		}
		if err := ValidateKnowledgeName(name); err != nil {
			return key, err
		}
		// A body over the limit is cut off; the executor then stores a truncated file, so reject by Content-Length first.
		if r.ContentLength > MaxKnowledgeBytes {
			return key, fmt.Errorf("%w: file too large", ErrInvalidInput)
		}
		body := http.MaxBytesReader(w, r.Body, MaxKnowledgeBytes)
		return key, finish(w, s.exec.PutKnowledgeFile(r.Context(), key, name, body))
	})
	s.route(http.MethodPost, bridgeKnowledgePrune, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqPrune
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if len(in.Names) > 5000 {
			return in.ProfileKey, fmt.Errorf("%w: too many names", ErrInvalidInput)
		}
		return in.ProfileKey, finish(w, s.exec.PruneKnowledge(r.Context(), in.ProfileKey, in.Names))
	})
	s.route(http.MethodPost, bridgeKnowledgeExtract, func(w http.ResponseWriter, r *http.Request) (string, error) {
		var in reqExtract
		if err := decode(r, &in); err != nil {
			return "", err
		}
		if err := ValidateProfileKey(in.ProfileKey); err != nil {
			return "", err
		}
		if err := ValidateKnowledgeName(in.Name); err != nil {
			return in.ProfileKey, err
		}
		made, err := s.exec.ExtractOne(r.Context(), in.ProfileKey, in.Name)
		if err != nil {
			return in.ProfileKey, err
		}
		return in.ProfileKey, writeJSON(w, respExtract{Made: made})
	})
	s.route(http.MethodPost, bridgeGatewayRestart, func(w http.ResponseWriter, r *http.Request) (string, error) {
		// The restart takes no input at all: there is no profile, container or argument a caller could choose.
		s.restartMu.Lock()
		defer s.restartMu.Unlock()
		if !s.lastRestart.IsZero() && time.Since(s.lastRestart) < s.cfg.GatewayRestartInterval {
			return "", errRateLimited
		}
		// The window starts when the restart is attempted, so a failing restart cannot be hammered either.
		s.lastRestart = time.Now()
		return "", finish(w, s.exec.RestartGateway(r.Context()))
	})
	s.route(http.MethodPost, bridgeKnowledgeReady, func(w http.ResponseWriter, r *http.Request) (string, error) {
		key, err := decodeProfile(r)
		if err != nil {
			return key, err
		}
		ready, err := s.exec.ReadyDocuments(r.Context(), key)
		if err != nil {
			return key, err
		}
		names := make([]string, 0, len(ready))
		for name, ok := range ready {
			if ok {
				names = append(names, name)
			}
		}
		return key, writeJSON(w, respReady{Names: names})
	})
}

// pair streams the pairing events as NDJSON; the final outcome travels in an HTTP trailer.
func (s *BridgeServer) pair(w http.ResponseWriter, r *http.Request) (string, error) {
	key, err := decodeProfile(r)
	if err != nil {
		return key, err
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		return key, errors.New("streaming unsupported")
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Trailer", "X-Bridge-Result")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	encoder := json.NewEncoder(w)
	runErr := s.exec.PairWhatsApp(r.Context(), key, func(event PairingEvent) {
		_ = encoder.Encode(event)
		flusher.Flush()
	})
	if runErr != nil {
		w.Header().Set("X-Bridge-Result", codeExecFailed)
	} else {
		w.Header().Set("X-Bridge-Result", "ok")
	}
	// The status line is already sent; the failure is reported through the trailer, so the route must not write again.
	if runErr != nil && r.Context().Err() == nil {
		s.log.Warn("bridge pairing ended with an error", "profile", key, "error", sanitizeError(runErr))
	}
	return key, nil
}

var profileActions = map[string]bool{"start": true, "stop": true, "restart": true, "update": true, "delete": true}

// Handler returns the HTTP handler to serve on the Unix socket.
func (s *BridgeServer) Handler() http.Handler { return s.mux }

type peerUIDKey struct{}

// ConnContext records the connecting process's UID (Linux SO_PEERCRED) for the allowed-UID check.
func (s *BridgeServer) ConnContext(ctx context.Context, conn net.Conn) context.Context {
	if uid, ok := peerUID(conn); ok {
		return context.WithValue(ctx, peerUIDKey{}, uid)
	}
	return ctx
}

func uidAllowed(allowed []uint32, uid uint32) bool {
	for _, candidate := range allowed {
		if candidate == uid {
			return true
		}
	}
	return false
}

func decode(r *http.Request, out any) error {
	data, err := readAll(r.Body, maxJSONBody)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("%w: malformed request", ErrInvalidInput)
	}
	return nil
}

func decodeProfile(r *http.Request) (string, error) {
	var in reqProfile
	if err := decode(r, &in); err != nil {
		return "", err
	}
	if err := ValidateProfileKey(in.ProfileKey); err != nil {
		return "", err
	}
	return in.ProfileKey, nil
}

// finish answers 200 {} for a successful operation that returns no data.
func finish(w http.ResponseWriter, err error) error {
	if err != nil {
		return err
	}
	return writeJSON(w, struct{}{})
}

func writeJSON(w http.ResponseWriter, value any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(respError{Error: code})
}

// sanitizeError keeps log lines short and free of anything shaped like a Telegram token or a long secret.
func sanitizeError(err error) string {
	message := err.Error()
	fields := strings.Fields(message)
	for i, field := range fields {
		trimmed := strings.Trim(field, `"'(),;`)
		if telegramLike(trimmed) || (len(trimmed) >= 32 && !strings.ContainsAny(trimmed, "/\\.")) {
			fields[i] = "[redacted]"
		}
	}
	message = strings.Join(fields, " ")
	if len(message) > 400 {
		message = message[:400] + "…"
	}
	return message
}

func telegramLike(value string) bool {
	idx := strings.Index(value, ":")
	if idx < 3 || idx > 20 {
		return false
	}
	return telegramTokenPattern.MatchString(value)
}

// DockerClientInspect confirms that the configured container exists. It is a variable so tests can replace it.
var DockerClientInspect = func(ctx context.Context, container string) error {
	return DockerClient{Container: container}.inspect(ctx)
}

// NewDockerExecutor returns the Docker-backed executor for the one container the bridge serves.
func NewDockerExecutor(container string) Executor { return DockerClient{Container: container} }
