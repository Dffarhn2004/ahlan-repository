package hermes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// knowledgeVersion identifies one exact text of the pesantren's knowledge.
func knowledgeVersion(knowledge string) string {
	sum := sha256.Sum256([]byte(knowledge))
	return hex.EncodeToString(sum[:8])
}

// syncKnowledgePlugin installs the plugin that gives conversations started before a knowledge change the newest
// knowledge. Hermes builds a conversation's system prompt (SOUL.md) once, when the conversation starts, so a changed
// document otherwise reaches visitors only after /new. The plugin runs before each model call; for a conversation whose
// prompt holds an older version it adds the current knowledge to that turn's message.
//
// The knowledge and its version are written on every settings update, so the plugin always reads the latest. It loads
// when the gateway starts, like the other plugins.
func (c DockerClient) syncKnowledgePlugin(ctx context.Context, profileKey, knowledge string) error {
	dir := "/opt/data/profiles/" + profileKey + "/plugins/ponpes-knowledge"
	files := map[string]string{
		dir + "/plugin.yaml":  "name: ponpes-knowledge\nversion: \"0.1.0\"\ndescription: \"Give conversations that began before a knowledge update the newest documents\"\nprovides_hooks:\n  - pre_llm_call\n",
		dir + "/__init__.py":  knowledgePluginSource,
		dir + "/profile.txt":  profileKey,
		dir + "/knowledge.md": strings.TrimSpace(knowledge),
		dir + "/version.txt":  knowledgeVersion(knowledge),
	}
	for target, content := range files {
		command := exec.CommandContext(ctx, "docker", "exec", "-i", c.Container, "sh", "-c", "mkdir -p "+shellQuote(path.Dir(target))+" && cat > "+shellQuote(target))
		command.Stdin = strings.NewReader(content)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("knowledge plugin failed: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}
	// Enabling a plugin that is already on is harmless, so the result is not an error.
	_ = exec.CommandContext(ctx, "docker", "exec", c.Container, "hermes", "-p", profileKey, "plugins", "enable", "ponpes-knowledge").Run()
	return nil
}

// knowledgePluginSource runs inside the gateway before every model call. It remembers, per conversation, which
// knowledge version that conversation's system prompt holds (seen on its first turn). A conversation it has never seen
// from the start (for example one running since before this plugin, or before a container restart) counts as older.
// For an older conversation the current knowledge is added to the turn, and since that addition lasts one turn only,
// it is repeated every turn until the visitor starts a new conversation. Any failure adds nothing.
const knowledgePluginSource = `"""ponpes-knowledge: give conversations that began before a knowledge update the newest documents."""

import json
import sys
import time
from pathlib import Path

_DIR = Path(__file__).parent
_PROFILE = (_DIR / "profile.txt").read_text(encoding="utf-8").strip()
_STATE = Path("/tmp/ponpes-knowledge-%s.json" % _PROFILE)
_KEEP = 2000
_HEADER = (
    "[Pembaruan data pesantren] Data di bawah ini adalah versi TERBARU dan menggantikan informasi pesantren, "
    "dokumen, biaya, jadwal, atau kontak yang tertulis sebelumnya dalam percakapan ini bila berbeda. "
    "Jangan menyebut bahwa data diperbarui."
)


def _log(line):
    print("ponpes-knowledge: " + time.strftime("%Y-%m-%d %H:%M:%S ") + line, file=sys.stderr, flush=True)


def _read(name):
    try:
        return (_DIR / name).read_text(encoding="utf-8")
    except Exception:
        return ""


def _load():
    try:
        return json.loads(_STATE.read_text(encoding="utf-8"))
    except Exception:
        return {}


def _save(state):
    try:
        if len(state) > _KEEP:
            state = dict(list(state.items())[-_KEEP:])
        temporary = _STATE.with_suffix(".tmp")
        temporary.write_text(json.dumps(state), encoding="utf-8")
        temporary.replace(_STATE)
    except Exception as error:
        _log("could not save state: %s" % error)


def _on_pre_llm_call(**kwargs):
    try:
        version = _read("version.txt").strip()
        knowledge = _read("knowledge.md").strip()
        session = str(kwargs.get("session_id") or "")
        if not version or not knowledge or not session:
            return None
        state = _load()
        known = state.get(session)
        first = kwargs.get("is_first_turn")
        if first is None:
            first = not (kwargs.get("conversation_history") or [])
        if known is None and first:
            # A new conversation: its system prompt was just built from the current knowledge.
            state[session] = version
            _save(state)
            return None
        if known == version:
            return None
        _log("session %s holds %s, current is %s: adding the current knowledge" % (session, known or "unknown", version))
        return {"context": _HEADER + "\n\n" + knowledge}
    except Exception as error:
        _log("failed, adding nothing: %s" % error)
        return None


def register(ctx):
    ctx.register_hook("pre_llm_call", _on_pre_llm_call)
`
