package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"ponpes-saas/internal/links"
)

// ErrLinkLimit means the tenant already has the maximum number of official links.
var ErrLinkLimit = errors.New("link limit reached")

// TenantLink is one official link. Learn is derived: a link is learned exactly when it points at a crawl source.
type TenantLink struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	Kind          string     `json:"kind"`
	Label         string     `json:"label"`
	URL           string     `json:"url"`
	ShareHint     string     `json:"share_hint"`
	Share         bool       `json:"share"`
	ShowPublic    bool       `json:"show_public"`
	Learn         bool       `json:"learn"`
	CrawlSourceID *uuid.UUID `json:"crawl_source_id,omitempty"`
	SortOrder     int        `json:"sort_order"`
}

const tenantLinkColumns = `id, tenant_id, kind, label, url, share_hint, share, show_public, crawl_source_id, crawl_source_id IS NOT NULL, sort_order`

func scanTenantLink(row pgx.Row) (TenantLink, error) {
	var l TenantLink
	err := row.Scan(&l.ID, &l.TenantID, &l.Kind, &l.Label, &l.URL, &l.ShareHint, &l.Share, &l.ShowPublic, &l.CrawlSourceID, &l.Learn, &l.SortOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return l, ErrNotFound
	}
	return l, err
}

func (r Repository) ListTenantLinks(ctx context.Context, tenantID uuid.UUID) ([]TenantLink, error) {
	rows, err := r.DB.Query(ctx, `SELECT `+tenantLinkColumns+` FROM tenant_links WHERE tenant_id=$1 ORDER BY sort_order, created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TenantLink{}
	for rows.Next() {
		l, err := scanTenantLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r Repository) GetTenantLink(ctx context.Context, tenantID, id uuid.UUID) (TenantLink, error) {
	return scanTenantLink(r.DB.QueryRow(ctx, `SELECT `+tenantLinkColumns+` FROM tenant_links WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

// CreateTenantLink inserts a link unless the tenant is at its limit (ErrLinkLimit) or already has the address (ErrConflict).
func (r Repository) CreateTenantLink(ctx context.Context, l TenantLink) (TenantLink, error) {
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, `INSERT INTO tenant_links(tenant_id,kind,label,url,share_hint,share,crawl_source_id,sort_order,show_public)
		SELECT $1,$2,$3,$4,$5,$6,$7,(SELECT coalesce(max(sort_order),0)+1 FROM tenant_links WHERE tenant_id=$1),$9
		WHERE (SELECT count(*) FROM tenant_links WHERE tenant_id=$1) < $8
		RETURNING id`, l.TenantID, l.Kind, l.Label, l.URL, l.ShareHint, l.Share, l.CrawlSourceID, links.MaxLinksPerSite, l.ShowPublic).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return l, ErrLinkLimit
	}
	if isUniqueViolation(err) {
		return l, ErrConflict
	}
	if err != nil {
		return l, err
	}
	return r.GetTenantLink(ctx, l.TenantID, id)
}

func (r Repository) UpdateTenantLink(ctx context.Context, l TenantLink) (TenantLink, error) {
	tag, err := r.DB.Exec(ctx, `UPDATE tenant_links SET kind=$3, label=$4, url=$5, share_hint=$6, share=$7, crawl_source_id=$8, show_public=$9, updated_at=now() WHERE id=$1 AND tenant_id=$2`,
		l.ID, l.TenantID, l.Kind, l.Label, l.URL, l.ShareHint, l.Share, l.CrawlSourceID, l.ShowPublic)
	if isUniqueViolation(err) {
		return l, ErrConflict
	}
	if err != nil {
		return l, err
	}
	if tag.RowsAffected() == 0 {
		return l, ErrNotFound
	}
	return r.GetTenantLink(ctx, l.TenantID, l.ID)
}

func (r Repository) DeleteTenantLink(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `DELETE FROM tenant_links WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// FindCrawlSourceByURL returns the tenant's existing source for an address, so a link can adopt it instead of duplicating it.
func (r Repository) FindCrawlSourceByURL(ctx context.Context, tenantID uuid.UUID, sourceURL string) (CrawlSource, error) {
	return scanCrawlSource(r.DB.QueryRow(ctx, `SELECT `+crawlSourceColumns+` FROM crawl_sources s WHERE s.tenant_id=$1 AND s.url=$2`, tenantID, sourceURL))
}

// SharedLinksByProfileKey lists the links the assistant may hand out, in display order.
func (r Repository) SharedLinksByProfileKey(ctx context.Context, profileKey string) []links.Shared {
	rows, err := r.DB.Query(ctx, `SELECT l.kind, l.label, l.url, l.share_hint FROM tenant_links l JOIN hermes_profiles h ON h.tenant_id=l.tenant_id
		WHERE h.profile_key=$1 AND l.share ORDER BY l.sort_order, l.created_at`, profileKey)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []links.Shared
	for rows.Next() {
		var s links.Shared
		if err := rows.Scan(&s.Kind, &s.Label, &s.URL, &s.Hint); err != nil {
			return nil
		}
		out = append(out, s)
	}
	return out
}
