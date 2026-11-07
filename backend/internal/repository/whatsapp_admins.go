package repository

import (
	"context"

	"github.com/google/uuid"
)

// WhatsAppAdmin is an admin who is notified on WhatsApp when the assistant routes a visitor to them. Phone is +<digits>.
type WhatsAppAdmin struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	NotifyWhen string `json:"notify_when"`
}

// ListWhatsAppAdmins returns the tenant's notification admins in display order.
func (r Repository) ListWhatsAppAdmins(ctx context.Context, tenantID uuid.UUID) ([]WhatsAppAdmin, error) {
	rows, err := r.DB.Query(ctx, `SELECT name, phone, notify_when FROM whatsapp_admins WHERE tenant_id=$1 ORDER BY sort_order, created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WhatsAppAdmin{}
	for rows.Next() {
		var a WhatsAppAdmin
		if err := rows.Scan(&a.Name, &a.Phone, &a.NotifyWhen); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ReplaceWhatsAppAdmins swaps the tenant's whole admin list in one transaction.
func (r Repository) ReplaceWhatsAppAdmins(ctx context.Context, tenantID uuid.UUID, admins []WhatsAppAdmin) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM whatsapp_admins WHERE tenant_id=$1`, tenantID); err != nil {
		return err
	}
	for i, a := range admins {
		if _, err := tx.Exec(ctx, `INSERT INTO whatsapp_admins(tenant_id,name,phone,notify_when,sort_order) VALUES($1,$2,$3,$4,$5)`, tenantID, a.Name, a.Phone, a.NotifyWhen, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
