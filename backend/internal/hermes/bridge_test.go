package hermes

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testKey = "ponpes_al-hidayah"

func TestValidateProfileKey(t *testing.T) {
	for _, ok := range []string{"ponpes_al-hidayah", "ponpes_a", "ponpes_abc_123"} {
		if err := ValidateProfileKey(ok); err != nil {
			t.Errorf("%q should be valid: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "default", "ponpes_", "ponpes_../x", "ponpes_a b", "ponpes_a;rm", "ponpes_A", "../ponpes_a", "ponpes_a\n", "-p", strings.Repeat("ponpes_a", 20)} {
		if err := ValidateProfileKey(bad); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q should be rejected, got %v", bad, err)
		}
	}
}

func TestValidateKnowledgeName(t *testing.T) {
	for _, ok := range []string{"ppdb.pdf", "Biaya & Fasilitas (2026).docx", "folder_" + uuid.NewString() + "/brosur.png", "a.teks.txt"} {
		if err := ValidateKnowledgeName(ok); err != nil {
			t.Errorf("%q should be valid: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "/etc/passwd", "../x", "a/../b", "a/b/c", "a//b", ".", "..", "a\x00b", "a\nb", `a\b`, "folder/..", strings.Repeat("x", 300)} {
		if err := ValidateKnowledgeName(bad); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q should be rejected, got %v", bad, err)
		}
	}
}

func TestValidateChannelInputs(t *testing.T) {
	if ValidateChannelPrefix("TELEGRAM_") != nil || ValidateChannelPrefix("WHATSAPP_") != nil {
		t.Error("channel prefixes should be valid")
	}
	for _, bad := range []string{"", "HOME", "TELEGRAM", ".*", "TELEGRAM_BOT_TOKEN="} {
		if ValidateChannelPrefix(bad) == nil {
			t.Errorf("prefix %q should be rejected", bad)
		}
	}
	if ValidateTelegramToken("123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw") != nil {
		t.Error("real-shaped token should be valid")
	}
	for _, bad := range []string{"", "abc", "123:short", "123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw\nX=1", "123456789:AAH dqTcvCH1vGWJxfSeofSAs0K5"} {
		if ValidateTelegramToken(bad) == nil {
			t.Errorf("token %q should be rejected", bad)
		}
	}
	if ValidateEventURL("http://172.17.0.1:8080/api/v1/internal/hermes/events") != nil {
		t.Error("event URL should be valid")
	}
	for _, bad := range []string{"", "file:///etc/passwd", "http://u:p@host/x", "http://host/x?a=1", "http://host/ x", "javascript:alert(1)", "http://host/x\"+__import__('os')"} {
		if ValidateEventURL(bad) == nil {
			t.Errorf("event URL %q should be rejected", bad)
		}
	}
}

func TestContainerNameIsValidated(t *testing.T) {
	if ValidateContainerName("hermes-ahlan") != nil {
		t.Error("hermes-ahlan should be valid")
	}
	for _, bad := range []string{"", "-v /:/host", "a b", "a;b", "../x"} {
		if ValidateContainerName(bad) == nil {
			t.Errorf("container %q should be rejected", bad)
		}
	}
	if _, err := NewBridgeServer(BridgeServerConfig{Container: "bad name"}, &fakeExec{}); err == nil {
		t.Error("bridge must refuse an invalid container name")
	}
}

func TestSanitizeErrorRedactsSecrets(t *testing.T) {
	out := sanitizeError(errors.New("failed 123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw and key abcdefghijklmnopqrstuvwxyz0123456789"))
	if strings.Contains(out, "AAHdq") || strings.Contains(out, "abcdefghijklmnop") {
		t.Fatalf("secret leaked: %s", out)
	}
}

// fakeExec records what the bridge asked the executor to do.
type fakeExec struct {
	mu      sync.Mutex
	calls   []string
	files   map[string]string
	pruned  []string
	failing bool
}

func (f *fakeExec) record(call string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call)
	if f.failing {
		return errors.New("boom token 123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw")
	}
	return nil
}
func (f *fakeExec) called() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}
func (f *fakeExec) CreateParkedProfile(_ context.Context, _ uuid.UUID, key string) (Profile, error) {
	return Profile{}, f.record("create " + key)
}
func (f *fakeExec) ProfileAction(_ context.Context, key, action string) error {
	return f.record("action " + key + " " + action)
}
func (f *fakeExec) ApplyProfileSettings(_ context.Context, key string, _ ProfileSettings) error {
	return f.record("settings " + key)
}
func (f *fakeExec) ApplyClientExperience(_ context.Context, key string) error {
	return f.record("experience " + key)
}
func (f *fakeExec) InstallWelcomeHook(_ context.Context, key, _ string) error {
	return f.record("welcome " + key)
}
func (f *fakeExec) InstallTelegramLogHook(_ context.Context, key, _, _ string) error {
	return f.record("loghook " + key)
}
func (f *fakeExec) ConfigureTelegram(_ context.Context, key, _, _ string, _ []int64) error {
	return f.record("telegram " + key)
}
func (f *fakeExec) ConfigureWhatsApp(_ context.Context, key, _, _ string, _ []string) error {
	return f.record("whatsapp " + key)
}
func (f *fakeExec) RemoveChannelEnv(_ context.Context, key, prefix string) error {
	return f.record("remove " + key + " " + prefix)
}
func (f *fakeExec) PairWhatsApp(_ context.Context, key string, onEvent func(PairingEvent)) error {
	onEvent(PairingEvent{Event: "started"})
	onEvent(PairingEvent{Event: "qr", QR: "QRDATA"})
	return f.record("pair " + key)
}
func (f *fakeExec) StopWhatsAppPairing(_ context.Context, key string) { _ = f.record("stop " + key) }
func (f *fakeExec) RemoveWhatsAppSession(_ context.Context, key string) error {
	return f.record("session-remove " + key)
}
func (f *fakeExec) PutKnowledgeFile(_ context.Context, key, name string, body io.Reader) error {
	data, _ := io.ReadAll(body)
	f.mu.Lock()
	if f.files == nil {
		f.files = map[string]string{}
	}
	f.files[key+"/"+name] = string(data)
	f.mu.Unlock()
	return f.record("put " + key + " " + name)
}
func (f *fakeExec) PruneKnowledge(_ context.Context, key string, names []string) error {
	f.mu.Lock()
	f.pruned = names
	f.mu.Unlock()
	return f.record("prune " + key)
}
func (f *fakeExec) ExtractOne(_ context.Context, key, name string) (bool, error) {
	return true, f.record("extract " + key + " " + name)
}
func (f *fakeExec) RestartGateway(_ context.Context) error { return f.record("gateway-restart") }
func (f *fakeExec) ReadyDocuments(_ context.Context, key string) (map[string]bool, error) {
	return map[string]bool{"a.pdf": true}, f.record("ready " + key)
}

func newBridge(t *testing.T, exec *fakeExec, cfg BridgeServerConfig) (*BridgeClient, http.Handler) {
	t.Helper()
	cfg.Container = "hermes-ahlan"
	server, err := NewBridgeServer(cfg, exec)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("", "br")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	socket := filepath.Join(dir, "b.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Skipf("unix sockets unavailable: %v", err)
	}
	httpServer := &http.Server{Handler: server.Handler(), ConnContext: server.ConnContext}
	go func() { _ = httpServer.Serve(listener) }()
	t.Cleanup(func() { _ = httpServer.Close() })
	client, err := NewBridgeClient(socket)
	if err != nil {
		t.Fatal(err)
	}
	return client, server.Handler()
}

func TestBridgeRoundTripCoversEveryOperation(t *testing.T) {
	exec := &fakeExec{}
	client, _ := newBridge(t, exec, BridgeServerConfig{})
	ctx := context.Background()
	tenant := uuid.New()
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	_, err := client.CreateParkedProfile(ctx, tenant, testKey)
	must("create", err)
	must("start", client.ProfileAction(ctx, testKey, "start"))
	must("settings", client.ApplyProfileSettings(ctx, testKey, ProfileSettings{SoulMD: "x"}))
	must("experience", client.ApplyClientExperience(ctx, testKey))
	must("welcome", client.InstallWelcomeHook(ctx, testKey, "Pesantren"))
	must("loghook", client.InstallTelegramLogHook(ctx, testKey, "http://172.17.0.1:8080/e", "secret-key-123"))
	must("telegram", client.ConfigureTelegram(ctx, testKey, "123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw", "PUBLIC", nil))
	must("whatsapp", client.ConfigureWhatsApp(ctx, testKey, "628123456789", "PUBLIC", nil))
	must("remove", client.RemoveChannelEnv(ctx, testKey, "TELEGRAM_"))
	must("session", client.RemoveWhatsAppSession(ctx, testKey))
	client.StopWhatsAppPairing(ctx, testKey)

	var events []string
	must("pair", client.PairWhatsApp(ctx, testKey, func(e PairingEvent) { events = append(events, e.Event) }))
	if strings.Join(events, ",") != "started,qr" {
		t.Fatalf("pairing events = %v", events)
	}

	must("sync", client.SyncKnowledge(ctx, testKey, []KnowledgeFile{{Name: "a.pdf", Body: strings.NewReader("AAA")}, {Name: "folder_x/b.txt", Body: strings.NewReader("BBB")}}))
	if exec.files[testKey+"/a.pdf"] != "AAA" || exec.files[testKey+"/folder_x/b.txt"] != "BBB" || strings.Join(exec.pruned, ",") != "a.pdf,folder_x/b.txt" {
		t.Fatalf("knowledge not transferred: %v pruned=%v", exec.files, exec.pruned)
	}
	created, failed := client.ExtractKnowledgeText(ctx, testKey, []string{"a.pdf", "b.png"}, nil)
	if created != 2 || len(failed) != 0 {
		t.Fatalf("extract created=%d failed=%v", created, failed)
	}
	ready, err := client.ReadyDocuments(ctx, testKey)
	if err != nil || !ready["a.pdf"] {
		t.Fatalf("ready = %v, %v", ready, err)
	}
	must("gateway restart", client.RestartGateway(ctx))
}

// A shared-gateway restart takes every tenant offline, so a second one inside the window must be refused and must
// never reach the executor.
func TestBridgeRateLimitsGatewayRestart(t *testing.T) {
	exec := &fakeExec{}
	client, _ := newBridge(t, exec, BridgeServerConfig{GatewayRestartInterval: time.Hour})
	ctx := context.Background()
	if err := client.RestartGateway(ctx); err != nil {
		t.Fatalf("first restart: %v", err)
	}
	err := client.RestartGateway(ctx)
	var refused *BridgeError
	if !errors.As(err, &refused) || refused.Status != http.StatusTooManyRequests || refused.Code != codeBusy {
		t.Fatalf("second restart should be refused with 429 BUSY, got %v", err)
	}
	restarts := 0
	for _, call := range exec.called() {
		if call == "gateway-restart" {
			restarts++
		}
	}
	if restarts != 1 {
		t.Fatalf("executor restarted the gateway %d times, want 1", restarts)
	}
}

// The restart route accepts no input, so it cannot be steered at another container or given arguments.
func TestBridgeGatewayRestartTakesNoInput(t *testing.T) {
	exec := &fakeExec{}
	_, handler := newBridge(t, exec, BridgeServerConfig{})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, bridgeGatewayRestart, nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET should be refused, got %d", rec.Code)
	}
	if len(exec.called()) != 0 {
		t.Fatalf("executor reached: %v", exec.called())
	}
}

func TestBridgeRejectsDangerousInput(t *testing.T) {
	exec := &fakeExec{}
	client, _ := newBridge(t, exec, BridgeServerConfig{AllowedEventURL: "http://172.17.0.1:8080/e"})
	ctx := context.Background()
	expectInvalid := func(label string, err error) {
		t.Helper()
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: expected invalid input, got %v", label, err)
		}
	}
	expectInvalid("other tenant's default profile", client.ProfileAction(ctx, "default", "start"))
	expectInvalid("traversal profile", client.ProfileAction(ctx, "ponpes_../../x", "start"))
	expectInvalid("unknown action", client.ProfileAction(ctx, testKey, "exec"))
	expectInvalid("shell action", client.ProfileAction(ctx, testKey, "start; rm -rf /"))
	expectInvalid("bad prefix", client.RemoveChannelEnv(ctx, testKey, "HOME"))
	expectInvalid("bad token", client.ConfigureTelegram(ctx, testKey, "x\nTELEGRAM_ALLOW_ALL_USERS=true", "PUBLIC", nil))
	expectInvalid("event url not allowed", client.InstallTelegramLogHook(ctx, testKey, "http://evil.example/e", "secret-key-123"))
	expectInvalid("path traversal file", client.SyncKnowledge(ctx, testKey, []KnowledgeFile{{Name: "../../etc/passwd", Body: strings.NewReader("x")}}))
	if len(exec.called()) != 0 {
		t.Fatalf("executor must not be reached by rejected requests: %v", exec.called())
	}
}

func TestBridgeRouteAllowListAndStrictDecoding(t *testing.T) {
	exec := &fakeExec{}
	_, handler := newBridge(t, exec, BridgeServerConfig{})
	call := func(method, path, body string) int {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec.Code
	}
	for _, path := range []string{"/v1/exec", "/v1/docker", "/v1/run", "/", "/v1/profile/create/../exec"} {
		// ServeMux may redirect cleaned paths (301 for GET, 307/308 for other
		// methods depending on the Go version); the target must not exist either.
		code := call(http.MethodPost, path, `{}`)
		if code != http.StatusNotFound && code != http.StatusMovedPermanently &&
			code != http.StatusTemporaryRedirect && code != http.StatusPermanentRedirect {
			t.Errorf("%s should not exist, got %d", path, code)
		}
	}
	if code := call(http.MethodGet, bridgeProfileAction, ``); code != http.StatusMethodNotAllowed {
		t.Errorf("wrong method got %d", code)
	}
	// Unknown fields such as a container override or an argument list are refused.
	if code := call(http.MethodPost, bridgeProfileAction, `{"profile_key":"`+testKey+`","action":"start","container":"other"}`); code != http.StatusBadRequest {
		t.Errorf("container override got %d", code)
	}
	if code := call(http.MethodPost, bridgeProfileAction, `{"profile_key":"`+testKey+`","action":"start","args":["rm"]}`); code != http.StatusBadRequest {
		t.Errorf("args injection got %d", code)
	}
	if code := call(http.MethodPost, bridgeProfileAction, `not json`); code != http.StatusBadRequest {
		t.Errorf("malformed got %d", code)
	}
	if len(exec.called()) != 0 {
		t.Fatalf("executor reached: %v", exec.called())
	}
}

func TestBridgeFailsClosedWithoutPeerIdentity(t *testing.T) {
	exec := &fakeExec{}
	server, err := NewBridgeServer(BridgeServerConfig{Container: "hermes-ahlan", AllowedUIDs: []uint32{1001}}, exec)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, bridgeProfileAction, strings.NewReader(`{"profile_key":"`+testKey+`","action":"start"}`)))
	if rec.Code != http.StatusForbidden || len(exec.called()) != 0 {
		t.Fatalf("expected 403 and no execution, got %d %v", rec.Code, exec.called())
	}
}

func TestBridgeErrorsCarryCodeNotOutput(t *testing.T) {
	exec := &fakeExec{failing: true}
	client, _ := newBridge(t, exec, BridgeServerConfig{})
	err := client.ProfileAction(context.Background(), testKey, "start")
	var bridgeErr *BridgeError
	if !errors.As(err, &bridgeErr) || bridgeErr.Code != codeExecFailed {
		t.Fatalf("expected EXEC_FAILED, got %v", err)
	}
	if strings.Contains(err.Error(), "AAHdq") || strings.Contains(err.Error(), "boom") {
		t.Fatalf("command output leaked to the caller: %v", err)
	}
}

func TestBridgeClientUnreachableSocket(t *testing.T) {
	client, _ := NewBridgeClient(filepath.Join(t.TempDir(), "missing.sock"))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.ProfileAction(ctx, testKey, "start"); err == nil {
		t.Fatal("expected an error when the bridge is down")
	}
}

func TestNewControlModes(t *testing.T) {
	if c, err := NewControl(ControlConfig{Mode: "docker", Container: "x"}); err != nil || c == nil {
		t.Fatalf("docker mode: %v", err)
	}
	if _, err := NewControl(ControlConfig{Mode: "bridge", BridgeSocket: ""}); err == nil {
		t.Fatal("bridge mode without a socket must fail")
	}
	if _, err := NewControl(ControlConfig{Mode: "ssh"}); err == nil {
		t.Fatal("unknown mode must fail, not fall back to docker")
	}
	if c, err := NewControl(ControlConfig{Mode: "bridge", BridgeSocket: "/run/x.sock"}); err != nil {
		t.Fatal(err)
	} else if _, ok := c.(*BridgeClient); !ok {
		t.Fatalf("bridge mode returned %T", c)
	}
}

func TestExtractAllRespectsConcurrencyAndReportsFailures(t *testing.T) {
	var mu sync.Mutex
	running, peak := 0, 0
	created, failed := extractAll([]string{"a", "b", "c", "d", "e", "f"}, nil, func(name string) (bool, error) {
		mu.Lock()
		running++
		if running > peak {
			peak = running
		}
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		running--
		mu.Unlock()
		if name == "c" {
			return false, errors.New("unreadable")
		}
		return true, nil
	})
	if created != 5 || len(failed) != 1 || failed[0] != "c" || peak > readConcurrency {
		t.Fatalf("created=%d failed=%v peak=%d", created, failed, peak)
	}
}
