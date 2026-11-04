package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PublicLink is the only tenant data exposed without login. Every field is meant to be shown to the public.
// A channel flag is true only when the bot is ACTIVE and open to everyone (PUBLIC), so visitors are never sent to a bot that will not answer.
type PublicLink struct {
	Slug             string
	Name             string
	Description      string
	GoogleMapsURL    string
	TelegramUsername string
	TelegramReady    bool
	WhatsAppNumber   string
	WhatsAppReady    bool
	Address          string
	// Socials maps lower-case kind (website, instagram, facebook, youtube, tiktok) to the pesantren's official URL.
	Socials map[string]string
	// LogoKey is the storage key of the uploaded logo; empty when none.
	LogoKey string
}

// GetPublicLinkBySlug returns the public link-page data for a tenant, or ErrNotFound for unknown or suspended tenants.
func (r Repository) GetPublicLinkBySlug(ctx context.Context, slug string) (PublicLink, error) {
	var p PublicLink
	var tenantID uuid.UUID
	var website, instagram, facebook, tiktok, youtube string
	err := r.DB.QueryRow(ctx, `
SELECT t.id, t.slug,
       coalesce(nullif(pr.display_name,''), t.name),
       coalesce(pr.description,''),
       coalesce(pr.google_maps_url,''),
       coalesce(tb.username,''),
       coalesce(tb.status='ACTIVE' AND tb.access_mode='PUBLIC', false),
       coalesce(wc.phone_number,''),
       coalesce(wc.status='ACTIVE' AND wc.access_mode='PUBLIC', false),
       coalesce(pr.address,''),
       coalesce(pr.website,''),
       coalesce(pr.instagram_url,''),
       coalesce(pr.facebook_url,''),
       coalesce(pr.tiktok_url,''),
       coalesce(pr.youtube_url,''),
       coalesce(pr.logo_key,'')
FROM tenants t
LEFT JOIN tenant_profiles pr ON pr.tenant_id=t.id
LEFT JOIN telegram_bots tb ON tb.tenant_id=t.id
LEFT JOIN whatsapp_channels wc ON wc.tenant_id=t.id
WHERE t.slug=$1 AND t.lifecycle_status <> 'SUSPENDED'`, slug).Scan(
		&tenantID, &p.Slug, &p.Name, &p.Description, &p.GoogleMapsURL, &p.TelegramUsername, &p.TelegramReady, &p.WhatsAppNumber, &p.WhatsAppReady, &p.Address, &website, &instagram, &facebook, &tiktok, &youtube, &p.LogoKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	if err != nil {
		return p, err
	}
	p.Socials = map[string]string{}
	for kind, value := range map[string]string{"website": website, "instagram": instagram, "facebook": facebook, "tiktok": tiktok, "youtube": youtube} {
		if value != "" {
			p.Socials[kind] = value
		}
	}
	// The profile wins; official links only fill gaps (older pesantren may have saved their pages there).
	rows, err := r.DB.Query(ctx, `SELECT lower(kind), url FROM tenant_links WHERE tenant_id=$1 AND kind IN ('WEBSITE','INSTAGRAM','FACEBOOK','YOUTUBE','TIKTOK') ORDER BY sort_order, created_at`, tenantID)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var kind, url string
		if err := rows.Scan(&kind, &url); err != nil {
			return p, err
		}
		if !seen[kind] {
			seen[kind] = true
			if p.Socials[kind] == "" {
				p.Socials[kind] = url
			}
		}
	}
	return p, rows.Err()
}

// ListPublicLinkSlugs returns slugs of tenants whose link page is worth indexing: not suspended and with at least one
// public, active channel. Used by the sitemap.
func (r Repository) ListPublicLinkSlugs(ctx context.Context) ([]string, error) {
	rows, err := r.DB.Query(ctx, `
SELECT t.slug FROM tenants t
WHERE t.lifecycle_status <> 'SUSPENDED' AND (
  EXISTS (SELECT 1 FROM telegram_bots b WHERE b.tenant_id=t.id AND b.status='ACTIVE' AND b.access_mode='PUBLIC') OR
  EXISTS (SELECT 1 FROM whatsapp_channels w WHERE w.tenant_id=t.id AND w.status='ACTIVE' AND w.access_mode='PUBLIC'))
ORDER BY t.slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		out = append(out, slug)
	}
	return out, rows.Err()
}

// SetProfileLogo stores the logo key and returns the previous one so the caller can delete the old file.
// An empty key clears the logo. ErrNotFound means the tenant has no profile row yet.
func (r Repository) SetProfileLogo(ctx context.Context, tenantID uuid.UUID, key string) (string, error) {
	var old string
	err := r.DB.QueryRow(ctx, `UPDATE tenant_profiles n SET logo_key=NULLIF($2,''), updated_at=now()
		FROM (SELECT coalesce(logo_key,'') AS old FROM tenant_profiles WHERE tenant_id=$1 FOR UPDATE) o
		WHERE n.tenant_id=$1 RETURNING o.old`, tenantID, key).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return old, err
}

// PublicLogoBySlug returns the tenant id and logo key for the public logo endpoint.
func (r Repository) PublicLogoBySlug(ctx context.Context, slug string) (uuid.UUID, string, error) {
	var id uuid.UUID
	var key string
	err := r.DB.QueryRow(ctx, `SELECT t.id, coalesce(pr.logo_key,'') FROM tenants t JOIN tenant_profiles pr ON pr.tenant_id=t.id
		WHERE t.slug=$1 AND t.lifecycle_status <> 'SUSPENDED'`, slug).Scan(&id, &key)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && key == "") {
		return id, "", ErrNotFound
	}
	return id, key, err
}

// PublicLinkFolder is one category of links the pesantren chose to show publicly.
type PublicLinkFolder struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

// PublicLinkItem is one link inside a public folder.
type PublicLinkItem struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Hint  string `json:"hint"`
}

// socialKinds stay as icon buttons on the page, so they never become folders.
const socialKindsSQL = `('WEBSITE','INSTAGRAM','FACEBOOK','YOUTUBE','TIKTOK')`

func (r Repository) ListPublicFolders(ctx context.Context, slug string) ([]PublicLinkFolder, error) {
	rows, err := r.DB.Query(ctx, `SELECT l.kind, count(*) FROM tenant_links l JOIN tenants t ON t.id=l.tenant_id
		WHERE t.slug=$1 AND t.lifecycle_status <> 'SUSPENDED' AND l.show_public AND l.kind NOT IN `+socialKindsSQL+`
		GROUP BY l.kind ORDER BY min(l.sort_order), l.kind`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PublicLinkFolder{}
	for rows.Next() {
		var f PublicLinkFolder
		if err := rows.Scan(&f.Kind, &f.Count); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r Repository) ListPublicFolderLinks(ctx context.Context, slug, kind string) ([]PublicLinkItem, error) {
	rows, err := r.DB.Query(ctx, `SELECT l.label, l.url, l.share_hint FROM tenant_links l JOIN tenants t ON t.id=l.tenant_id
		WHERE t.slug=$1 AND t.lifecycle_status <> 'SUSPENDED' AND l.show_public AND l.kind=$2 AND l.kind NOT IN `+socialKindsSQL+`
		ORDER BY l.sort_order, l.created_at`, slug, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PublicLinkItem{}
	for rows.Next() {
		var i PublicLinkItem
		if err := rows.Scan(&i.Label, &i.URL, &i.Hint); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
