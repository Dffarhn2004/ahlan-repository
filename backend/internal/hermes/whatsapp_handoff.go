package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// handoffNotice is sent once to a visitor who writes again while an admin has taken the conversation over.
const handoffNotice = "Pesan Anda sudah kami teruskan ke admin. Mohon menunggu, admin akan menghubungi Anda lewat WhatsApp."

// handoffAPIConfig tells the in-container scripts where to report and check hand-offs.
func handoffAPIConfig(settings ProfileSettings) string {
	body, _ := json.Marshal(map[string]string{"url": settings.HandoffURL, "key": settings.HandoffKey})
	return string(body)
}

// syncHandoffPlugin installs (routing) or removes the plugin that keeps the assistant silent while an admin has taken a
// WhatsApp conversation over. It runs on the gateway's pre_gateway_dispatch hook, so it only loads when the gateway
// starts, exactly like the notify hook.
func (c DockerClient) syncHandoffPlugin(ctx context.Context, profileKey string, settings ProfileSettings, routing bool) error {
	dir := "/opt/data/profiles/" + profileKey + "/plugins/ponpes-handoff"
	if !routing || settings.HandoffURL == "" || settings.HandoffKey == "" {
		if output, err := exec.CommandContext(ctx, "docker", "exec", c.Container, "rm", "-rf", dir).CombinedOutput(); err != nil {
			return fmt.Errorf("handoff plugin removal failed: %w: %s", err, strings.TrimSpace(string(output)))
		}
		return nil
	}
	files := map[string]string{
		dir + "/plugin.yaml": "name: ponpes-handoff\nversion: \"0.1.0\"\ndescription: \"Keep the assistant silent while an admin has taken a WhatsApp conversation over\"\nprovides_hooks:\n  - pre_gateway_dispatch\n",
		dir + "/__init__.py": handoffPluginSource,
		dir + "/api.json":    handoffAPIConfig(settings),
		dir + "/profile.txt": profileKey,
		dir + "/notice.txt":  handoffNotice,
	}
	for target, content := range files {
		command := exec.CommandContext(ctx, "docker", "exec", "-i", c.Container, "sh", "-c", "mkdir -p "+shellQuote(path.Dir(target))+" && cat > "+shellQuote(target))
		command.Stdin = strings.NewReader(content)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("handoff plugin failed: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}
	// Plugins may need switching on; enabling one that is already on is harmless, so the result is not an error.
	_ = exec.CommandContext(ctx, "docker", "exec", c.Container, "hermes", "-p", profileKey, "plugins", "enable", "ponpes-handoff").Run()
	return nil
}

// handoffPluginSource runs inside the gateway before every inbound message. It asks the API whether the chat is handed
// over; if so the message is dropped (logged by the API) and, once per hand-off, the visitor gets a short notice.
// Any failure lets the message through, so a broken API never silences the assistant.
const handoffPluginSource = `"""ponpes-handoff: stay silent while an admin has taken a WhatsApp conversation over."""

import json
import subprocess
import sys
import threading
import time
import urllib.request
from pathlib import Path

_DIR = Path(__file__).parent
_PROFILE = (_DIR / "profile.txt").read_text(encoding="utf-8").strip()
_NOTICE = (_DIR / "notice.txt").read_text(encoding="utf-8").strip()
_noticed = set()


def _log(line):
    stamp = time.strftime("%Y-%m-%d %H:%M:%S ") + line
    print("ponpes-handoff: " + stamp, file=sys.stderr, flush=True)
    try:
        with open("/tmp/ponpes-handoff.log", "a", encoding="utf-8") as handle_file:
            handle_file.write(stamp + "\n")
    except Exception:
        pass


def _is_handed_off(chat_id, user_id, text):
    config = json.loads((_DIR / "api.json").read_text(encoding="utf-8"))
    body = json.dumps({"profile_key": _PROFILE, "action": "check", "chat_id": chat_id, "user_id": user_id, "text": text}).encode("utf-8")
    request = urllib.request.Request(config["url"], data=body, headers={"Content-Type": "application/json", "X-Hermes-Event-Key": config["key"]})
    with urllib.request.urlopen(request, timeout=2) as response:
        return bool(json.loads(response.read().decode("utf-8")).get("handoff"))


def _end_handoff(chat_id, user_id):
    try:
        config = json.loads((_DIR / "api.json").read_text(encoding="utf-8"))
        body = json.dumps({"profile_key": _PROFILE, "action": "end", "chat_id": chat_id, "user_id": user_id}).encode("utf-8")
        request = urllib.request.Request(config["url"], data=body, headers={"Content-Type": "application/json", "X-Hermes-Event-Key": config["key"]})
        urllib.request.urlopen(request, timeout=2).read()
        _log("handoff released for %s" % chat_id)
    except Exception as error:
        _log("handoff release failed: %s" % error)


def _send_notice(chat_id):
    try:
        subprocess.run(["hermes", "-p", _PROFILE, "send", "--to", "whatsapp:" + chat_id, "--quiet", _NOTICE], capture_output=True, text=True, timeout=60)
    except Exception as error:
        _log("notice failed: %s" % error)


def _on_pre_gateway_dispatch(event=None, **kwargs):
    try:
        source = getattr(event, "source", None)
        platform = getattr(source, "platform", None)
        if source is None or str(getattr(platform, "value", platform)) != "whatsapp":
            return None
        chat_id = str(getattr(source, "chat_id", "") or "")
        text = str(getattr(event, "text", "") or "")
        if not chat_id:
            return None
        user_id = str(getattr(source, "user_id", "") or "")
        if text.startswith("/"):
            # /new starts a fresh conversation, so it also releases a hand-off; other commands pass untouched.
            if text.strip().lower().split()[0] in ("/new", "/reset"):
                _noticed.discard(chat_id)
                _end_handoff(chat_id, user_id)
            return None
        if not _is_handed_off(chat_id, user_id, text[:2000]):
            _noticed.discard(chat_id)
            return None
        _log("handoff: dropped a message from %s" % chat_id)
        if chat_id not in _noticed:
            _noticed.add(chat_id)
            threading.Thread(target=_send_notice, args=(chat_id,), daemon=True).start()
        return {"action": "skip", "reason": "handoff"}
    except Exception as error:
        _log("check failed, letting the message through: %s" % error)
        return None


def register(ctx):
    ctx.register_hook("pre_gateway_dispatch", _on_pre_gateway_dispatch)
`
