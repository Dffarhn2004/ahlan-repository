package hermes

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWhatsAppRoutingRulesUseOnlyTheAdminSentence(t *testing.T) {
	rules := whatsappRoutingRules()
	if !strings.Contains(rules, handoverSentence) {
		t.Error("the fixed hand-over sentence is missing")
	}
	if !strings.Contains(strings.ToLower(handoverSentence), handoverMarker) {
		t.Error("the hand-over sentence must contain the marker the hook looks for")
	}
	if !strings.Contains(rules, "ingin kami teruskan ke admin pondok") {
		t.Error("the offer to pass the visitor to an admin is missing")
	}
	if strings.Contains(strings.ToLower("Apakah Bapak/Ibu ingin kami teruskan ke admin pondok agar dijawab langsung?"), handoverMarker) {
		t.Error("the offer must not contain the hand-over marker")
	}
}

func TestNotifyHookHandlerIsValidPython(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("python not installed")
	}
	file := filepath.Join(t.TempDir(), "handler.py")
	if err := os.WriteFile(file, []byte(notifyHookHandler), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(python, "-I", "-m", "py_compile", file).CombinedOutput(); err != nil {
		t.Fatalf("handler does not compile: %v: %s", err, output)
	}
}

// TestNotifyHookStartsOnlyOnTheHandOverSentence runs the real hook with the outside world stubbed out and checks
// which answers start a hand-off: only the fixed hand-over sentence, which then notifies every admin.
func TestNotifyHookStartsOnlyOnTheHandOverSentence(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("python not installed")
	}
	dir := t.TempDir()
	for name, content := range map[string]string{
		"handler.py":  notifyHookHandler,
		"profile.txt": "tenant",
		"admins.json": `[{"name":"Ani","phone":"+6281111111111","notify_when":"PSB"},{"name":"Budi","phone":"+6282222222222","notify_when":"surat"}]`,
		"api.json":    `{"url":"http://127.0.0.1:1/handoff","key":"k"}`,
		"driver.py": `import importlib.util, subprocess, sys, types, urllib.request
sent = []
class R:
    returncode = 0
    stdout = ""
    stderr = ""
subprocess.run = lambda *a, **k: (sent.append(a[0]), R())[1]
class Resp:
    def read(self): return b"{}"
urllib.request.urlopen = lambda *a, **k: Resp()
spec = importlib.util.spec_from_file_location("handler", "handler.py")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)
cases = [
    ("Apakah Bapak/Ibu ingin kami teruskan ke admin pondok?", 0),
    ("Ani akan menghubungi Anda.", 0),
    ("Silakan hubungi admin lewat nomor lain.", 0),
    ("Baik, pesan Anda kami teruskan ke admin pondok. Admin akan menghubungi Bapak/Ibu lewat WhatsApp, mohon menunggu.", 2),
]
for index, (answer, want) in enumerate(cases):
    before = len(sent)
    mod._notified.clear()
    mod.handle("agent:end", {"platform": "whatsapp", "chat_id": "6281234567890@s.whatsapp.net", "user_id": "6281234567890", "response": answer})
    got = len(sent) - before
    if got != want:
        print("case %d (%r): sent %d, want %d" % (index, answer, got, want))
        sys.exit(1)
`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command(python, "-I", "driver.py")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("hook matched the wrong answers: %v: %s", err, output)
	}
}

func TestHandoffPluginIsValidPython(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("python not installed")
	}
	file := filepath.Join(t.TempDir(), "__init__.py")
	if err := os.WriteFile(file, []byte(handoffPluginSource), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(python, "-I", "-m", "py_compile", file).CombinedOutput(); err != nil {
		t.Fatalf("plugin does not compile: %v: %s", err, output)
	}
}
