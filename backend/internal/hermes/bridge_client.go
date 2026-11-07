package hermes

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BridgeClient implements Control by calling the local Hermes bridge over a Unix domain socket.
// It holds no Docker access and no credentials: authorisation is the socket's file permission.
type BridgeClient struct {
	http *http.Client
}

var _ Control = (*BridgeClient)(nil)

// Timeouts per kind of operation. Streaming and document reading are bounded by the caller's context instead.
const (
	bridgeShortTimeout = 2 * time.Minute
	bridgeLongTimeout  = 10 * time.Minute
)

func NewBridgeClient(socketPath string) (*BridgeClient, error) {
	socketPath = strings.TrimSpace(socketPath)
	if socketPath == "" {
		return nil, errors.New("HERMES_BRIDGE_SOCKET is not configured")
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "unix", socketPath)
		},
		MaxIdleConns:    8,
		IdleConnTimeout: 30 * time.Second,
	}
	// No client-wide timeout: pairing streams for minutes. Each call sets its own deadline.
	return &BridgeClient{http: &http.Client{Transport: transport}}, nil
}

// BridgeError is a failure reported by the bridge. Only a short code crosses the socket, never command output.
type BridgeError struct {
	Status int
	Code   string
}

func (e *BridgeError) Error() string {
	return fmt.Sprintf("Hermes bridge refused the request: %s (HTTP %d)", e.Code, e.Status)
}

func (e *BridgeError) Unwrap() error {
	if e.Code == codeInvalidInput {
		return ErrInvalidInput
	}
	return nil
}

func (c *BridgeClient) do(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType string) (*http.Response, error) {
	target := "http://bridge" + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Hermes bridge unreachable: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var failure respError
		_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&failure)
		return nil, &BridgeError{Status: resp.StatusCode, Code: failure.Error}
	}
	return resp, nil
}

// call posts req as JSON and decodes the JSON answer into out (when not nil).
func (c *BridgeClient) call(ctx context.Context, timeout time.Duration, path string, req, out any) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodPost, path, nil, bytes.NewReader(payload), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Check asks the bridge whether it can reach its Hermes container.
func (c *BridgeClient) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	resp, err := c.do(ctx, http.MethodGet, bridgeHealth, nil, nil, "")
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *BridgeClient) CreateParkedProfile(ctx context.Context, tenantID uuid.UUID, profileKey string) (Profile, error) {
	if tenantID == uuid.Nil || strings.TrimSpace(profileKey) == "" {
		return Profile{}, ErrTenantScopeRequired
	}
	if err := c.call(ctx, bridgeLongTimeout, bridgeProfileCreate, reqCreate{TenantID: tenantID.String(), ProfileKey: profileKey}, nil); err != nil {
		return Profile{}, err
	}
	return Profile{TenantID: tenantID, ProfileKey: profileKey, Status: "HEALTHY"}, nil
}

func (c *BridgeClient) ProfileAction(ctx context.Context, profileKey, action string) error {
	return c.call(ctx, bridgeLongTimeout, bridgeProfileAction, reqAction{ProfileKey: profileKey, Action: action}, nil)
}

func (c *BridgeClient) RestartGateway(ctx context.Context) error {
	return c.call(ctx, bridgeLongTimeout, bridgeGatewayRestart, struct{}{}, nil)
}

func (c *BridgeClient) ApplyProfileSettings(ctx context.Context, profileKey string, s ProfileSettings) error {
	return c.call(ctx, bridgeShortTimeout, bridgeProfileSettings, reqSettings{ProfileKey: profileKey, Description: s.Description, SoulMD: s.SoulMD, EnabledSkills: s.EnabledSkills, TenantInfo: s.TenantInfo, WhatsAppGeneral: s.WhatsAppGeneral, WhatsAppAdmins: s.WhatsAppAdmins}, nil)
}

func (c *BridgeClient) ApplyClientExperience(ctx context.Context, profileKey string) error {
	return c.call(ctx, bridgeLongTimeout, bridgeClientExperience, reqProfile{ProfileKey: profileKey}, nil)
}

func (c *BridgeClient) InstallWelcomeHook(ctx context.Context, profileKey, pesantren string) error {
	return c.call(ctx, bridgeShortTimeout, bridgeWelcomeHook, reqWelcome{ProfileKey: profileKey, Pesantren: pesantren}, nil)
}

func (c *BridgeClient) InstallTelegramLogHook(ctx context.Context, profileKey, eventURL, eventKey string) error {
	return c.call(ctx, bridgeShortTimeout, bridgeLogHook, reqLogHook{ProfileKey: profileKey, EventURL: eventURL, EventKey: eventKey}, nil)
}

func (c *BridgeClient) ConfigureTelegram(ctx context.Context, profileKey, token, mode string, allowedUserIDs []int64) error {
	return c.call(ctx, bridgeShortTimeout, bridgeTelegram, reqTelegram{ProfileKey: profileKey, Token: token, Mode: mode, AllowedUserIDs: allowedUserIDs}, nil)
}

func (c *BridgeClient) ConfigureWhatsApp(ctx context.Context, profileKey, phone, mode string, allowedNumbers []string) error {
	return c.call(ctx, bridgeShortTimeout, bridgeWhatsApp, reqWhatsApp{ProfileKey: profileKey, Phone: phone, Mode: mode, AllowedNumbers: allowedNumbers}, nil)
}

func (c *BridgeClient) RemoveChannelEnv(ctx context.Context, profileKey, prefix string) error {
	return c.call(ctx, bridgeShortTimeout, bridgeChannelRemove, reqChannelRemove{ProfileKey: profileKey, Prefix: prefix}, nil)
}

// PairWhatsApp streams the bridge's NDJSON events until the pairing ends. Cancelling ctx closes the stream, and the
// bridge then stops the pairing process inside the container.
func (c *BridgeClient) PairWhatsApp(ctx context.Context, profileKey string, onEvent func(PairingEvent)) error {
	payload, _ := json.Marshal(reqProfile{ProfileKey: profileKey})
	resp, err := c.do(ctx, http.MethodPost, bridgePair, nil, bytes.NewReader(payload), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var event PairingEvent
		if json.Unmarshal(scanner.Bytes(), &event) == nil && event.Event != "" {
			onEvent(event)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if trailer := resp.Trailer.Get("X-Bridge-Result"); trailer != "" && trailer != "ok" {
		return fmt.Errorf("WhatsApp pairing failed: %s", trailer)
	}
	return nil
}

func (c *BridgeClient) StopWhatsAppPairing(ctx context.Context, profileKey string) {
	_ = c.call(ctx, bridgeShortTimeout, bridgePairStop, reqProfile{ProfileKey: profileKey}, nil)
}

func (c *BridgeClient) RemoveWhatsAppSession(ctx context.Context, profileKey string) error {
	return c.call(ctx, bridgeShortTimeout, bridgeSessionRemove, reqProfile{ProfileKey: profileKey}, nil)
}

// SyncKnowledge streams each document to the bridge, then asks it to drop the documents that no longer exist.
func (c *BridgeClient) SyncKnowledge(ctx context.Context, profileKey string, files []KnowledgeFile) error {
	if err := ValidateProfileKey(profileKey); err != nil {
		return err
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		if strings.TrimSpace(file.Name) == "" || file.Body == nil {
			continue
		}
		if err := ValidateKnowledgeName(file.Name); err != nil {
			return err
		}
		names = append(names, file.Name)
	}
	for _, file := range files {
		if strings.TrimSpace(file.Name) == "" || file.Body == nil {
			continue
		}
		if err := c.putKnowledge(ctx, profileKey, file); err != nil {
			return fmt.Errorf("knowledge sync failed for %s: %w", file.Name, err)
		}
	}
	return c.call(ctx, bridgeLongTimeout, bridgeKnowledgePrune, reqPrune{ProfileKey: profileKey, Names: names}, nil)
}

func (c *BridgeClient) putKnowledge(ctx context.Context, profileKey string, file KnowledgeFile) error {
	ctx, cancel := context.WithTimeout(ctx, bridgeLongTimeout)
	defer cancel()
	resp, err := c.do(ctx, http.MethodPut, bridgeKnowledgePut, url.Values{"profile_key": {profileKey}, "name": {file.Name}}, file.Body, "application/octet-stream")
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *BridgeClient) ExtractKnowledgeText(ctx context.Context, profileKey string, names []string, progress func(name string, ok bool)) (int, []string) {
	return extractAll(names, progress, func(name string) (bool, error) {
		var out respExtract
		// Reading a scanned PDF through the model can take many minutes.
		if err := c.call(ctx, 30*time.Minute, bridgeKnowledgeExtract, reqExtract{ProfileKey: profileKey, Name: name}, &out); err != nil {
			return false, err
		}
		return out.Made, nil
	})
}

func (c *BridgeClient) ReadyDocuments(ctx context.Context, profileKey string) (map[string]bool, error) {
	var out respReady
	if err := c.call(ctx, bridgeShortTimeout, bridgeKnowledgeReady, reqProfile{ProfileKey: profileKey}, &out); err != nil {
		return nil, err
	}
	ready := make(map[string]bool, len(out.Names))
	for _, name := range out.Names {
		ready[name] = true
	}
	return ready, nil
}
