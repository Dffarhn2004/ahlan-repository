package hermes

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestSearchKnowledgeNeverCrossesTenant(t *testing.T) {
	client := DevClient{}
	a, b := uuid.New(), uuid.New()
	result, err := client.SearchKnowledge(context.Background(), a, "biaya", []KnowledgeDocument{
		{TenantID: a, FileID: uuid.New(), Filename: "ppdb.txt", Text: "Informasi biaya PPDB"},
		{TenantID: b, FileID: uuid.New(), Filename: "biaya.txt", Text: "Data tenant lain"},
	})
	if err != nil || len(result) != 1 || result[0].Filename != "ppdb.txt" {
		t.Fatalf("unexpected tenant-scoped result: %+v %v", result, err)
	}
}

func TestSearchKnowledgeRequiresTenant(t *testing.T) {
	if _, err := (DevClient{}).SearchKnowledge(context.Background(), uuid.Nil, "test", nil); err != ErrTenantScopeRequired {
		t.Fatalf("expected tenant scope error, got %v", err)
	}
}

func TestWelcomeMessageGreetsWithPesantrenName(t *testing.T) {
	got := welcomeMessage("Pondok Pesantren Contoh")
	for _, want := range []string{"Assalamu'alaikum", "Pondok Pesantren Contoh", "Mulai percakapan baru"} {
		if !strings.Contains(got, want) {
			t.Errorf("welcome message missing %q: %s", want, got)
		}
	}
	if !strings.Contains(welcomeMessage(" "), "pesantren kami") {
		t.Error("blank name should fall back to a generic greeting")
	}
}

// section returns the indented block that follows a "  name:" line, up to the next line of the same indent.
func localeSection(t *testing.T, parent, name string) string {
	t.Helper()
	lines := strings.Split(publicLocaleID, "\n")
	inParent, start := false, -1
	for i, line := range lines {
		switch {
		case line == parent+":":
			inParent = true
		case inParent && line == "  "+name+":":
			start = i + 1
		case start >= 0 && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    "):
			return strings.Join(lines[start:i], "\n")
		case start >= 0 && line != "" && !strings.HasPrefix(line, " "):
			return strings.Join(lines[start:i], "\n")
		}
	}
	if start < 0 {
		t.Fatalf("locale section %s.%s not found", parent, name)
	}
	return strings.Join(lines[start:], "\n")
}

// /new and /reset must never show the model, provider, context window or the technical tip to a visitor.
func TestPublicLocaleHidesTechnicalSessionInfo(t *testing.T) {
	session := localeSection(t, "gateway", "session")
	for _, key := range []string{"info_model", "info_provider", "info_context", "info_acting_model", "info_endpoint"} {
		if !strings.Contains(session, "    "+key+": \"\"") {
			t.Errorf("gateway.session.%s must be blanked", key)
		}
	}
	reset := localeSection(t, "gateway", "reset")
	if !strings.Contains(reset, "    tip: \"\"") {
		t.Error("gateway.reset.tip must be blanked")
	}
	for _, key := range []string{"header_default", "header_new", "header_titled"} {
		if !strings.Contains(reset, "    "+key+": \"Percakapan") {
			t.Errorf("gateway.reset.%s must be an Indonesian visitor message", key)
		}
	}
	if strings.Contains(reset, "info_") {
		t.Error("info_* keys belong under gateway.session, not gateway.reset")
	}
}

// The log hook runs inside the gateway's reply path. It must hand events to a background thread; a blocking HTTP
// call there delayed every visitor reply by the request timeout whenever the API was unreachable.
func TestLogHookNeverBlocksTheReply(t *testing.T) {
	source := logHookHandler("http://172.18.0.1:8081/api/v1/internal/hermes/events", "secret-key-123", "ponpes_al-hidayah")
	handle := source[strings.Index(source, "def handle("):]
	if strings.Contains(handle, "urlopen") || strings.Contains(handle, "urllib") {
		t.Fatal("handle() must not perform the HTTP call itself")
	}
	for _, want := range []string{"threading.Thread(", "daemon=True", "put_nowait", "queue.Full", `EVENT_URL = "http://172.18.0.1:8081/api/v1/internal/hermes/events"`} {
		if !strings.Contains(source, want) {
			t.Errorf("handler is missing %q", want)
		}
	}
	if python, err := exec.LookPath("python3"); err == nil {
		check := exec.Command(python, "-c", "import ast,sys; ast.parse(sys.stdin.read())")
		check.Stdin = strings.NewReader(source)
		if output, err := check.CombinedOutput(); err != nil {
			t.Fatalf("generated handler is not valid Python: %v: %s", err, output)
		}
	}
}
