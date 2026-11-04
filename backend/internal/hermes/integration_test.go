//go:build integration && linux

package hermes

// Real-Docker integration test: BridgeClient -> Unix socket -> BridgeServer -> DockerClient -> `docker exec` in a
// disposable container. The container is a plain Alpine image with a stub `hermes` CLI that records its arguments,
// so everything that touches the filesystem (.env via stdin, knowledge, hooks, persona) runs for real.
//
//	go test -tags integration -run Integration -v ./internal/hermes
//
// Needs Linux and a reachable Docker daemon. HERMES_IT_IMAGE overrides the image (default alpine:3.20).

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const stubHermes = `#!/bin/sh
echo "$*" >> /tmp/hermes-calls
if [ "$1" = "profile" ] && [ "$2" = "create" ]; then mkdir -p "/opt/data/profiles/$3"; fi
case "$*" in
  *"gateway status"*) echo "Gateway is running";;
  *"tools list"*) printf '  ✓ enabled  terminal  Run commands\n  ✓ enabled  clarify  Ask questions\n  ✗ disabled  web  Web\n';;
  *"chat"*) printf 'transcribed text\nsession_id: abc\n';;
esac
exit 0
`

type itEnv struct {
	container string
	client    *BridgeClient
	argvLog   string
}

func dockerOut(t *testing.T, stdin string, args ...string) string {
	t.Helper()
	cmd := exec.Command("docker", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker %v: %v: %s", args, err, out)
	}
	return string(out)
}

func setupIntegration(t *testing.T) *itEnv {
	t.Helper()
	realDocker, err := exec.LookPath("docker")
	if err != nil {
		t.Skip("docker CLI not available")
	}
	if err := exec.Command(realDocker, "info").Run(); err != nil {
		t.Skip("docker daemon not reachable")
	}
	image := os.Getenv("HERMES_IT_IMAGE")
	if image == "" {
		image = "alpine:3.20"
	}
	name := "ahlan-it-" + uuid.NewString()[:8]
	dockerOut(t, "", "run", "-d", "--name", name, image, "sleep", "900")
	t.Cleanup(func() { _ = exec.Command(realDocker, "rm", "-f", name).Run() })
	dockerOut(t, stubHermes, "exec", "-i", name, "sh", "-c", "cat > /usr/local/bin/hermes && chmod +x /usr/local/bin/hermes")

	// A wrapper in front of docker records every argument vector, to prove secrets never appear on a command line.
	dir := t.TempDir()
	argvLog := filepath.Join(dir, "argv.log")
	wrapper := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> %q\nexec %q \"$@\"\n", argvLog, realDocker)
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	server, err := NewBridgeServer(BridgeServerConfig{Container: name, AllowedEventURL: "http://10.99.0.1:8081/api/v1/internal/hermes/events"}, NewDockerExecutor(name))
	if err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(dir, "bridge.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	httpServer := &http.Server{Handler: server.Handler(), ConnContext: server.ConnContext}
	go func() { _ = httpServer.Serve(listener) }()
	t.Cleanup(func() { _ = httpServer.Close() })
	client, err := NewBridgeClient(socket)
	if err != nil {
		t.Fatal(err)
	}
	return &itEnv{container: name, client: client, argvLog: argvLog}
}

func (e *itEnv) read(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("docker", "exec", e.container, "cat", path).CombinedOutput()
	if err != nil {
		t.Fatalf("cat %s: %v: %s", path, err, out)
	}
	return string(out)
}

func (e *itEnv) exists(path string) bool {
	return exec.Command("docker", "exec", e.container, "test", "-e", path).Run() == nil
}

func TestIntegrationBridgeAgainstRealDocker(t *testing.T) {
	e := setupIntegration(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	const key, other = "ponpes_it", "ponpes_other"
	const token = "123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw"
	const token2 = "987654321:BBHdqTcvCH1vGWJxfSeofSAs0K5PALDzzz"
	root := "/opt/data/profiles/" + key
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}

	if err := e.client.Check(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}
	for _, k := range []string{key, other} {
		_, err := e.client.CreateParkedProfile(ctx, uuid.New(), k)
		must("create "+k, err)
	}
	if !e.exists(root) {
		t.Fatal("profile directory was not created")
	}

	// .env written through stdin: unrelated lines survive, the previous token is replaced, nothing is left half-written.
	dockerOut(t, "FOO=bar\nTELEGRAM_BOT_TOKEN=old\nWHATSAPP_ENABLED=false\n", "exec", "-i", e.container, "sh", "-c", "cat > "+root+"/.env")
	must("telegram", e.client.ConfigureTelegram(ctx, key, token, "TESTING", []int64{111, 222}))
	env := e.read(t, root+"/.env")
	for _, want := range []string{"FOO=bar", "TELEGRAM_BOT_TOKEN=" + token, "TELEGRAM_ALLOWED_USERS=111,222", "TELEGRAM_HOME_CHANNEL=123456789", "WHATSAPP_ENABLED=false"} {
		if !strings.Contains(env, want) {
			t.Errorf(".env missing %q:\n%s", want, env)
		}
	}
	if strings.Contains(env, "=old") {
		t.Errorf("old token survived:\n%s", env)
	}
	must("telegram again", e.client.ConfigureTelegram(ctx, key, token2, "PUBLIC", nil))
	env = e.read(t, root+"/.env")
	if strings.Contains(env, token+"\n") || strings.Contains(env, "TELEGRAM_ALLOWED_USERS") || !strings.Contains(env, "TELEGRAM_ALLOW_ALL_USERS=true") || strings.Count(env, "TELEGRAM_BOT_TOKEN=") != 1 {
		t.Errorf("re-configuration not clean:\n%s", env)
	}
	must("whatsapp", e.client.ConfigureWhatsApp(ctx, key, "628123456789", "PUBLIC", nil))
	must("remove telegram", e.client.RemoveChannelEnv(ctx, key, "TELEGRAM_"))
	env = e.read(t, root+"/.env")
	if strings.Contains(env, "TELEGRAM_") || !strings.Contains(env, "WHATSAPP_ENABLED=true") || !strings.Contains(env, "FOO=bar") || e.exists(root+"/.env.tmp") {
		t.Errorf("channel removal damaged .env:\n%s", env)
	}

	// Persona, hooks, lockdown.
	must("settings", e.client.ApplyProfileSettings(ctx, key, ProfileSettings{Description: "desc", SoulMD: "SOUL-TEXT", EnabledSkills: []string{"a"}, TenantInfo: "INFO-TEXT"}))
	if soul := e.read(t, root+"/SOUL.md"); !strings.Contains(soul, "SOUL-TEXT") || !strings.Contains(soul, "INFO-TEXT") {
		t.Errorf("SOUL.md incomplete: %s", soul)
	}
	must("welcome", e.client.InstallWelcomeHook(ctx, key, "Pesantren It"))
	if !strings.Contains(e.read(t, root+"/hooks/ponpes-welcome/welcome.txt"), "Pesantren It") {
		t.Error("welcome text missing")
	}
	hookURL := "http://10.99.0.1:8081/api/v1/internal/hermes/events"
	must("loghook", e.client.InstallTelegramLogHook(ctx, key, hookURL, "event-key-for-it"))
	if h := e.read(t, root+"/hooks/ponpes-telegram-log/handler.py"); !strings.Contains(h, hookURL) || !strings.Contains(h, key) {
		t.Errorf("log hook wrong: %s", h)
	}
	if err := e.client.InstallTelegramLogHook(ctx, key, "http://evil.example/e", "event-key-for-it"); err == nil {
		t.Error("a callback URL outside the allow-list must be refused")
	}
	must("experience", e.client.ApplyClientExperience(ctx, key))
	calls := e.read(t, "/tmp/hermes-calls")
	for _, want := range []string{"profile create " + key, "-p " + key + " config set display.language id", "-p " + key + " tools disable --platform telegram terminal"} {
		if !strings.Contains(calls, want) {
			t.Errorf("hermes CLI never received %q:\n%s", want, calls)
		}
	}
	must("start", e.client.ProfileAction(ctx, key, "start"))
	if !strings.Contains(e.read(t, "/tmp/hermes-calls"), "-p "+key+" gateway start") {
		t.Error("gateway start not issued")
	}

	// Knowledge: sync, read, prune, and strict per-profile isolation.
	must("seed other", e.client.SyncKnowledge(ctx, other, []KnowledgeFile{{Name: "secret.txt", Body: strings.NewReader("OTHER TENANT DATA")}}))
	docs := []KnowledgeFile{
		{Name: "ppdb.txt", Body: strings.NewReader("biaya 1 juta")},
		{Name: "folder_" + uuid.NewString() + "/brosur.txt", Body: strings.NewReader("brosur isi")},
		{Name: "stale.txt", Body: strings.NewReader("to be pruned")},
	}
	must("sync", e.client.SyncKnowledge(ctx, key, docs))
	created, failed := e.client.ExtractKnowledgeText(ctx, key, []string{docs[0].Name, docs[1].Name, docs[2].Name}, nil)
	if created != 3 || len(failed) != 0 {
		t.Fatalf("extract created=%d failed=%v", created, failed)
	}
	ready, err := e.client.ReadyDocuments(ctx, key)
	must("ready", err)
	if !ready["ppdb.txt"] || !ready[docs[1].Name] || len(ready) != 3 {
		t.Fatalf("ready = %v", ready)
	}
	must("resync without stale", e.client.SyncKnowledge(ctx, key, []KnowledgeFile{{Name: "ppdb.txt", Body: strings.NewReader("biaya 1 juta")}, {Name: docs[1].Name, Body: strings.NewReader("brosur isi")}}))
	if e.exists(root+"/knowledge/stale.txt") || e.exists(root+"/knowledge/stale.txt"+TranscriptSuffix) {
		t.Error("deleted document was not pruned")
	}
	if !e.exists("/opt/data/profiles/" + other + "/knowledge/secret.txt") {
		t.Error("pruning one tenant removed another tenant's file")
	}
	if soul := func() string {
		must("settings 2", e.client.ApplyProfileSettings(ctx, key, ProfileSettings{SoulMD: "S"}))
		return e.read(t, root+"/SOUL.md")
	}(); strings.Contains(soul, "OTHER TENANT DATA") || !strings.Contains(soul, "biaya 1 juta") {
		t.Errorf("persona must embed only this tenant's documents: %s", soul)
	}
	if err := e.client.SyncKnowledge(ctx, key, []KnowledgeFile{{Name: "../" + other + "/knowledge/secret.txt", Body: strings.NewReader("x")}}); err == nil {
		t.Error("path traversal into another profile must be refused")
	}
	if got := e.read(t, "/opt/data/profiles/"+other+"/knowledge/secret.txt"); got != "OTHER TENANT DATA" {
		t.Errorf("other tenant's file changed: %q", got)
	}

	// WhatsApp session removal.
	dockerOut(t, "", "exec", e.container, "sh", "-c", "mkdir -p "+root+"/whatsapp/session && touch "+root+"/whatsapp/session/creds.json")
	must("session remove", e.client.RemoveWhatsAppSession(ctx, key))
	if e.exists(root + "/whatsapp") {
		t.Error("WhatsApp session was not removed")
	}

	// Secrets never appeared on a command line of the host.
	argv, _ := os.ReadFile(e.argvLog)
	for _, secret := range []string{token, token2, "event-key-for-it"} {
		if strings.Contains(string(argv), secret) {
			t.Errorf("secret %q leaked into a docker command line", secret[:6]+"…")
		}
	}
	// Only `docker exec`/`docker inspect` on the configured container were ever issued.
	for _, line := range strings.Split(strings.TrimSpace(string(argv)), "\n") {
		if !strings.HasPrefix(line, "exec ") && !strings.HasPrefix(line, "inspect ") {
			if !strings.HasPrefix(line, "run ") && !strings.HasPrefix(line, "rm ") && !strings.HasPrefix(line, "info") {
				t.Errorf("unexpected docker subcommand: %s", line)
			}
		}
	}
}
