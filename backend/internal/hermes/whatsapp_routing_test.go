package hermes

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWhatsAppRoutingRulesListAdminsWithoutNumbers(t *testing.T) {
	rules := whatsappRoutingRules([]WhatsAppAdmin{{Name: "Ustadz Ahmad", Phone: "+6281234567890", NotifyWhen: "surat-menyurat\ndan legalisir"}})
	if !strings.Contains(rules, "- Ustadz Ahmad, dihubungi jika: surat-menyurat dan legalisir") {
		t.Errorf("admin line missing or multi-line: %q", rules)
	}
	if strings.Contains(rules, "81234567890") {
		t.Error("admin phone must never reach the persona")
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
