package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"ponpes-saas/internal/contacts"
)

// ErrContactLimit means the tenant already has the maximum number of contact persons.
var ErrContactLimit = errors.New("contact limit reached")

// TenantContact is one contact person the assistant may hand out.
type TenantContact struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Hours     string    `json:"hours"`
	ShareHint string    `json:"share_hint"`
	SortOrder int       `json:"sort_order"`
}

const tenantContactColumns = `id, tenant_id, name, role, phone, email, hours, share_hint, sort_order`

func scanTenantContact(row pgx.Row) (TenantContact, error) {
	var c TenantContact
	err := row.Scan(&c.ID, &c.TenantID, &c.Name, &c.Role, &c.Phone, &c.Email, &c.Hours, &c.ShareHint, &c.SortOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r Repository) ListTenantContacts(ctx context.Context, tenantID uuid.UUID) ([]TenantContact, error) {
	rows, err := r.DB.Query(ctx, `SELECT `+tenantContactColumns+` FROM tenant_contacts WHERE tenant_id=$1 ORDER BY sort_order, created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TenantContact{}
	for rows.Next() {
		c, err := scanTenantContact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r Repository) GetTenantContact(ctx context.Context, tenantID, id uuid.UUID) (TenantContact, error) {
	return scanTenantContact(r.DB.QueryRow(ctx, `SELECT `+tenantContactColumns+` FROM tenant_contacts WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

// CreateTenantContact inserts a contact unless the tenant is at its limit (ErrContactLimit).
func (r Repository) CreateTenantContact(ctx context.Context, c TenantContact) (TenantContact, error) {
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, `INSERT INTO tenant_contacts(tenant_id,name,role,phone,email,hours,share_hint,sort_order)
		SELECT $1,$2,$3,$4,$5,$6,$7,(SELECT coalesce(max(sort_order),0)+1 FROM tenant_contacts WHERE tenant_id=$1)
		WHERE (SELECT count(*) FROM tenant_contacts WHERE tenant_id=$1) < $8
		RETURNING id`, c.TenantID, c.Name, c.Role, c.Phone, c.Email, c.Hours, c.ShareHint, contacts.MaxPerSite).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrContactLimit
	}
	if err != nil {
		return c, err
	}
	return r.GetTenantContact(ctx, c.TenantID, id)
}

func (r Repository) UpdateTenantContact(ctx context.Context, c TenantContact) (TenantContact, error) {
	tag, err := r.DB.Exec(ctx, `UPDATE tenant_contacts SET name=$3, role=$4, phone=$5, email=$6, hours=$7, share_hint=$8, updated_at=now() WHERE id=$1 AND tenant_id=$2`,
		c.ID, c.TenantID, c.Name, c.Role, c.Phone, c.Email, c.Hours, c.ShareHint)
	if err != nil {
		return c, err
	}
	if tag.RowsAffected() == 0 {
		return c, ErrNotFound
	}
	return r.GetTenantContact(ctx, c.TenantID, c.ID)
}

func (r Repository) DeleteTenantContact(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `DELETE FROM tenant_contacts WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// SharedContactsByProfileKey lists the contacts the assistant may hand out, in display order.
func (r Repository) SharedContactsByProfileKey(ctx context.Context, profileKey string) []contacts.Shared {
	rows, err := r.DB.Query(ctx, `SELECT c.name, c.role, c.phone, c.email, c.hours, c.share_hint FROM tenant_contacts c JOIN hermes_profiles h ON h.tenant_id=c.tenant_id
		WHERE h.profile_key=$1 ORDER BY c.sort_order, c.created_at`, profileKey)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []contacts.Shared
	for rows.Next() {
		var s contacts.Shared
		if err := rows.Scan(&s.Name, &s.Role, &s.Phone, &s.Email, &s.Hours, &s.Hint); err != nil {
			return nil
		}
		out = append(out, s)
	}
	return out
}
