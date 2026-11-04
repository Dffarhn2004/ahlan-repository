package hermes

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// whatsappBridgeDir is where the Hermes image ships its WhatsApp (Baileys) bridge.
const whatsappBridgeDir = "/opt/hermes/scripts/whatsapp-bridge"

// WhatsAppSessionDir is where a profile's gateway looks for its linked-device credentials.
func WhatsAppSessionDir(profileKey string) string {
	return "/opt/data/profiles/" + profileKey + "/whatsapp/session"
}

// PairingEvent is one JSON line printed by the bridge in --pair-json mode.
type PairingEvent struct {
	Event string `json:"event"` // started, qr, connected, disconnected, error
	QR    string `json:"qr,omitempty"`
	Error string `json:"error,omitempty"`
	User  *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user,omitempty"`
}

// PairWhatsApp starts a fresh device-linking session for the profile and reports every bridge event to onEvent.
// It blocks until the bridge exits (after a successful scan, an error, or ctx being cancelled).
func (c DockerClient) PairWhatsApp(ctx context.Context, profileKey string, onEvent func(PairingEvent)) error {
	if strings.TrimSpace(c.Container) == "" {
		return errors.New("Hermes container is not configured")
	}
	if err := ValidateProfileKey(profileKey); err != nil {
		return err
	}
	session := WhatsAppSessionDir(profileKey)
	// The bridge's npm dependencies are installed on first use; a fresh session guarantees a QR code is shown.
	prepare := "cd " + shellQuote(whatsappBridgeDir) + " && { [ -d node_modules ] || npm install --silent --no-audit --no-fund; } && rm -rf " + shellQuote(session) + " && mkdir -p " + shellQuote(session)
	if output, err := exec.CommandContext(ctx, "docker", "exec", c.Container, "sh", "-c", prepare).CombinedOutput(); err != nil {
		return fmt.Errorf("WhatsApp bridge preparation failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	command := exec.CommandContext(ctx, "docker", "exec", "-e", "WHATSAPP_MODE=bot", "-e", "WHATSAPP_DM_POLICY=pairing", "-w", whatsappBridgeDir, c.Container,
		"node", "bridge.js", "--pair-only", "--pair-json", "--session", session)
	// Killing the local docker client does not stop the process inside the container, so cancel it explicitly.
	command.Cancel = func() error {
		c.StopWhatsAppPairing(context.Background(), profileKey)
		return command.Process.Kill()
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var event PairingEvent
		if json.Unmarshal([]byte(line), &event) == nil && event.Event != "" {
			onEvent(event)
		}
	}
	waitErr := command.Wait()
	// docker exec runs as root, but the gateway runs as the profile's owner and must write its bridge log here.
	own := "chown -R --reference=" + shellQuote("/opt/data/profiles/"+profileKey) + " " + shellQuote("/opt/data/profiles/"+profileKey+"/whatsapp")
	if output, err := exec.CommandContext(context.WithoutCancel(ctx), "docker", "exec", c.Container, "sh", "-c", own).CombinedOutput(); err != nil {
		return fmt.Errorf("WhatsApp session ownership failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return waitErr
}

// StopWhatsAppPairing ends a running pairing process of the profile, if any.
func (c DockerClient) StopWhatsAppPairing(ctx context.Context, profileKey string) {
	if ValidateProfileKey(profileKey) != nil {
		return
	}
	_ = exec.CommandContext(ctx, "docker", "exec", c.Container, "pkill", "-f", "pair-only.*"+WhatsAppSessionDir(profileKey)).Run()
}

// RemoveWhatsAppSession forgets the linked device credentials of the profile.
func (c DockerClient) RemoveWhatsAppSession(ctx context.Context, profileKey string) error {
	if err := ValidateProfileKey(profileKey); err != nil {
		return err
	}
	c.StopWhatsAppPairing(ctx, profileKey)
	output, err := exec.CommandContext(ctx, "docker", "exec", c.Container, "rm", "-rf", "/opt/data/profiles/"+profileKey+"/whatsapp").CombinedOutput()
	if err != nil {
		return fmt.Errorf("WhatsApp session removal failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
