package httpapi

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
	"ponpes-saas/internal/hermes"
	"ponpes-saas/internal/repository"
)

// pairingTimeout is how long a QR code stays scannable before the attempt is dropped.
const pairingTimeout = 3 * time.Minute

// whatsappPairing is one in-progress device linking. The QR code changes every ~20 seconds until it is scanned.
type whatsappPairing struct {
	mu      sync.Mutex
	status  string // starting, qr, connected, error
	qr      string // PNG data URL of the current code
	phone   string // digits of the linked number once connected
	errCode string
	cancel  context.CancelFunc
}

func (p *whatsappPairing) set(update func(*whatsappPairing)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	update(p)
}

func (p *whatsappPairing) snapshot() fiber.Map {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := fiber.Map{"status": p.status}
	if p.status == "qr" {
		out["qr"] = p.qr
	}
	if p.status == "connected" {
		out["phone_number"] = p.phone
	}
	if p.status == "error" {
		out["error"] = p.errCode
	}
	return out
}

func (p *whatsappPairing) linkedPhone() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.phone, p.status == "connected"
}

var nonDigits = regexp.MustCompile(`\D+`)

// jidDigits returns the number part of a WhatsApp ID such as "6281234:12@s.whatsapp.net".
func jidDigits(jid string) string {
	number, _, _ := strings.Cut(jid, "@")
	number, _, _ = strings.Cut(number, ":")
	return nonDigits.ReplaceAllString(number, "")
}

// normalizeNumber turns what an admin types ("0812-…", "+62 812…") into digits with country code.
func normalizeNumber(value string) (string, bool) {
	digits := nonDigits.ReplaceAllString(value, "")
	if strings.HasPrefix(digits, "0") {
		digits = "62" + digits[1:]
	}
	return digits, len(digits) >= 8 && len(digits) <= 15
}

func pairingKey(tenantID uuid.UUID) string { return tenantID.String() }

// startWhatsAppPairing begins linking a number and answers at once; the page then polls for the QR code.
func (s *Server) startWhatsAppPairing(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	profileKey, err := s.repo.HermesProfileKey(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusConflict, "HERMES_PROFILE_REQUIRED")
	}
	// Restarting an attempt replaces the previous one.
	if previous, ok := s.whatsappPairings.Load(pairingKey(id)); ok {
		previous.(*whatsappPairing).cancel()
	}
	client := s.docker()
	// The linked-device session can only be used by one process: stop the running assistant first.
	if _, channelErr := s.repo.GetWhatsAppChannel(c.Context(), id); channelErr == nil {
		_ = client.ProfileAction(c.Context(), profileKey, "stop")
	}
	ctx, cancel := context.WithTimeout(context.Background(), pairingTimeout)
	pairing := &whatsappPairing{status: "starting", cancel: cancel}
	s.whatsappPairings.Store(pairingKey(id), pairing)
	go func() {
		defer cancel()
		err := client.PairWhatsApp(ctx, profileKey, func(event hermes.PairingEvent) {
			switch event.Event {
			case "qr":
				png, encodeErr := qrcode.Encode(event.QR, qrcode.Medium, 320)
				if encodeErr != nil {
					return
				}
				pairing.set(func(p *whatsappPairing) {
					p.status, p.qr = "qr", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(png)
				})
			case "connected":
				phone := ""
				if event.User != nil {
					phone = jidDigits(event.User.ID)
				}
				pairing.set(func(p *whatsappPairing) { p.status, p.qr, p.phone = "connected", "", phone })
			case "error":
				pairing.set(func(p *whatsappPairing) { p.status, p.errCode = "error", event.Error })
			}
		})
		pairing.set(func(p *whatsappPairing) {
			// The bridge exits after a successful link; anything else before that is a failed or expired attempt.
			if p.status != "connected" && p.status != "error" {
				p.status, p.errCode = "error", "expired"
				if err != nil && ctx.Err() == nil {
					p.errCode = "failed"
				}
			}
		})
		if err != nil && ctx.Err() == nil {
			slog.Warn("WhatsApp pairing ended with an error", "tenant", id, "error", err)
		}
	}()
	return c.Status(http.StatusAccepted).JSON(pairing.snapshot())
}

func (s *Server) whatsappPairingStatus(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	pairing, ok := s.whatsappPairings.Load(pairingKey(id))
	if !ok {
		return c.JSON(fiber.Map{"status": "none"})
	}
	return c.JSON(pairing.(*whatsappPairing).snapshot())
}

func (s *Server) cancelWhatsAppPairing(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	if pairing, ok := s.whatsappPairings.LoadAndDelete(pairingKey(id)); ok {
		pairing.(*whatsappPairing).cancel()
	}
	return c.SendStatus(http.StatusNoContent)
}

func (s *Server) getWhatsApp(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	channel, err := s.repo.GetWhatsAppChannel(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "WHATSAPP_NOT_FOUND")
	}
	return c.JSON(channel)
}

// saveWhatsApp stores the freshly linked number with its access policy, validated exactly like saveBot.
func (s *Server) saveWhatsApp(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in repository.WhatsAppChannel
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	in.AccessMode = strings.ToUpper(strings.TrimSpace(in.AccessMode))
	if in.AccessMode == "" {
		in.AccessMode = "TESTING"
	}
	if in.AccessMode != "TESTING" && in.AccessMode != "PUBLIC" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_ACCESS_MODE")
	}
	if in.AccessMode == "PUBLIC" {
		in.AllowedNumbers = []string{}
	} else {
		seen := map[string]bool{}
		numbers := []string{}
		for _, raw := range in.AllowedNumbers {
			number, ok := normalizeNumber(raw)
			if !ok {
				return fiber.NewError(http.StatusBadRequest, "INVALID_WHATSAPP_NUMBER")
			}
			if !seen[number] {
				seen[number] = true
				numbers = append(numbers, number)
			}
		}
		if len(numbers) == 0 {
			return fiber.NewError(http.StatusBadRequest, "WHATSAPP_NUMBERS_REQUIRED")
		}
		in.AllowedNumbers = numbers
	}
	// The number comes from the scan the server watched, never from the browser.
	in.PhoneNumber = ""
	if pairing, ok := s.whatsappPairings.Load(pairingKey(id)); ok {
		if phone, linked := pairing.(*whatsappPairing).linkedPhone(); linked {
			in.PhoneNumber = phone
		}
	}
	if in.PhoneNumber == "" {
		existing, existsErr := s.repo.GetWhatsAppChannel(c.Context(), id)
		if existsErr != nil {
			return fiber.NewError(http.StatusConflict, "WHATSAPP_NOT_LINKED")
		}
		in.PhoneNumber = existing.PhoneNumber
	}
	in.TenantID = id
	if err := s.repo.UpsertWhatsAppChannel(c.Context(), in); err != nil {
		return err
	}
	s.whatsappPairings.Delete(pairingKey(id))
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "WHATSAPP_CONNECTED", "WHATSAPP_CHANNEL", id.String(), "+"+in.PhoneNumber+" ("+strings.ToLower(in.AccessMode)+")")
	in.Status = "PENDING"
	return c.JSON(in)
}

func (s *Server) deleteWhatsApp(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	if pairing, ok := s.whatsappPairings.LoadAndDelete(pairingKey(id)); ok {
		pairing.(*whatsappPairing).cancel()
	}
	if profileKey, keyErr := s.repo.HermesProfileKey(c.Context(), id); keyErr == nil {
		client := s.docker()
		if err := client.RemoveChannelEnv(c.Context(), profileKey, "WHATSAPP_"); err != nil {
			slog.Warn("could not clear WhatsApp settings", "profile", profileKey, "error", err)
		}
		if err := client.RemoveWhatsAppSession(c.Context(), profileKey); err != nil {
			slog.Warn("could not remove WhatsApp session", "profile", profileKey, "error", err)
		}
	}
	if err := s.repo.DeleteWhatsAppChannel(c.Context(), id); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "WHATSAPP_REMOVED", "WHATSAPP_CHANNEL", id.String(), "")
	return c.SendStatus(http.StatusNoContent)
}
