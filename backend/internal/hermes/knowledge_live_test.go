package hermes

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestKnowledgePluginAddsKnowledgeOnlyToOlderConversations runs the real plugin with Hermes' hook registration stubbed.
func TestKnowledgePluginAddsKnowledgeOnlyToOlderConversations(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("python not installed")
	}
	dir := t.TempDir()
	for name, content := range map[string]string{
		"__init__.py":  knowledgePluginSource,
		"profile.txt":  "tenant-test-" + filepath.Base(dir),
		"knowledge.md": "## Dokumen\nBiaya Rp1.000",
		"version.txt":  "v1",
		"driver.py": `import importlib.util, os, sys
spec = importlib.util.spec_from_file_location("plugin", "__init__.py")
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)
hooks = {}
class Ctx:
    def register_hook(self, name, fn): hooks[name] = fn
mod.register(Ctx())
call = hooks["pre_llm_call"]
def fail(message):
    print(message); sys.exit(1)
try:
    # A new conversation already holds the current knowledge: nothing is added.
    if call(session_id="a", is_first_turn=True, conversation_history=[]) is not None: fail("new conversation got knowledge")
    if call(session_id="a", is_first_turn=False, conversation_history=["x"]) is not None: fail("current conversation got knowledge")
    # A conversation never seen from its start counts as older: it gets the knowledge, on every turn.
    for turn in range(2):
        out = call(session_id="old", is_first_turn=False, conversation_history=["x"])
        if not out or "Biaya Rp1.000" not in out["context"]: fail("older conversation got nothing on turn %d" % turn)
    # The knowledge changes: the conversation that held v1 now gets v2.
    open("version.txt", "w").write("v2"); open("knowledge.md", "w").write("## Dokumen\nBiaya Rp2.000")
    out = call(session_id="a", is_first_turn=False, conversation_history=["x"])
    if not out or "Rp2.000" not in out["context"]: fail("changed knowledge not given to conversation a")
    # A conversation started after the change is current again.
    if call(session_id="b", is_first_turn=True, conversation_history=[]) is not None: fail("new conversation b got knowledge")
    if call(session_id="b", is_first_turn=False, conversation_history=["x"]) is not None: fail("conversation b got knowledge")
    # Missing knowledge adds nothing.
    open("knowledge.md", "w").write("")
    if call(session_id="a", is_first_turn=False, conversation_history=["x"]) is not None: fail("empty knowledge was added")
finally:
    try: os.remove(str(mod._STATE))
    except Exception: pass
`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command(python, "-I", "driver.py")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("plugin misbehaves: %v: %s", err, output)
	}
}

func TestKnowledgeVersionChangesWithTheText(t *testing.T) {
	if knowledgeVersion("a") == knowledgeVersion("b") || knowledgeVersion("a") != knowledgeVersion("a") {
		t.Error("the version must follow the text")
	}
}
