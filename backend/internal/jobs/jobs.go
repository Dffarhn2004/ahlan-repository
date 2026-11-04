package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoJob is returned by Claim when the queue is empty. It is an expected
// condition, not a failure.
var ErrNoJob = errors.New("no provisioning job available")

type Job struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	ProfileKey string
	Status     string
	Attempt    int
}
type Repository struct {
	DB    *pgxpool.Pool
	Lease time.Duration
}

func (r Repository) Claim(ctx context.Context) (*Job, error) {
	lease := r.Lease
	if lease <= 0 {
		lease = time.Minute
	}
	row := r.DB.QueryRow(ctx, `UPDATE provisioning_runs SET status='RUNNING', attempt=attempt+1, locked_until=now()+$1::interval, updated_at=now() WHERE id=(SELECT id FROM provisioning_runs WHERE status IN ('QUEUED','RUNNING') AND (locked_until IS NULL OR locked_until < now()) ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id, tenant_id, status, attempt`, fmt.Sprintf("%d seconds", int(lease.Seconds())))
	var j Job
	if err := row.Scan(&j.ID, &j.TenantID, &j.Status, &j.Attempt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoJob
		}
		return nil, err
	}
	if err := r.DB.QueryRow(ctx, `SELECT 'ponpes_' || slug FROM tenants WHERE id=$1`, j.TenantID).Scan(&j.ProfileKey); err != nil {
		return nil, err
	}
	slog.Info("provisioning job claimed", "job_id", j.ID, "tenant_id", j.TenantID, "attempt", j.Attempt)
	return &j, nil
}
func (r Repository) Complete(ctx context.Context, id uuid.UUID, profileKey, runtimeStatus string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var tenant uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT tenant_id FROM provisioning_runs WHERE id=$1`, id).Scan(&tenant); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO hermes_profiles(tenant_id,profile_key,hermes_host_id,runtime_status) VALUES($1,$2,'docker-hermes',$3) ON CONFLICT (tenant_id) DO UPDATE SET profile_key=EXCLUDED.profile_key,hermes_host_id='docker-hermes',runtime_status=EXCLUDED.runtime_status`, tenant, profileKey, runtimeStatus); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE provisioning_runs SET status='SUCCEEDED', locked_until=NULL, updated_at=now() WHERE id=$1`, id); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE tenants SET updated_at=now() WHERE id=$1`, tenant)
	if err != nil {
		return err
	}
	slog.Info("provisioning job completed", "job_id", id, "tenant_id", tenant)
	return tx.Commit(ctx)
}
func (r Repository) Fail(ctx context.Context, id uuid.UUID, code string) error {
	_, err := r.DB.Exec(ctx, `UPDATE provisioning_runs SET status='FAILED', error_code=$2, locked_until=NULL, updated_at=now() WHERE id=$1`, id, code)
	if err != nil {
		slog.Error("provisioning job failed to record", "job_id", id, "error_code", code, "error", err)
	} else {
		slog.Warn("provisioning job failed", "job_id", id, "error_code", code)
	}
	return err
}
