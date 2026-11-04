# 16 — Worker, job queue, idempotency, event flows

## Pilihan P0

Pakai **PostgreSQL-backed job queue** dengan `FOR UPDATE SKIP LOCKED`, bukan wajib Redis. Go API dan Go worker berasal dari satu repo, dua binary (`cmd/api`, `cmd/worker`). Worker mengurus provisioning, indexing, health checks dan purge. Ketika load naik, dapat dimigrasikan ke queue khusus.

## Job states

`QUEUED`, `RUNNING`, `WAITING_INPUT`, `SUCCEEDED`, `FAILED`, `DEAD_LETTER`, `CANCELLED`; `attempt_count`, `max_attempts`, `run_after`, `lock_owner`, `lock_expires_at`, `last_error_code`, payload dengan no raw secret, dedup key.

## Claiming jobs secara aman (contoh SQL)

```sql
WITH picked AS (
  SELECT id FROM jobs
  WHERE status = 'QUEUED' AND run_after <= now()
  ORDER BY priority DESC, created_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
UPDATE jobs AS j
SET status = 'RUNNING', lock_owner = $1,
    lock_expires_at = now() + interval '60 seconds',
    attempt_count = attempt_count + 1
FROM picked WHERE j.id = picked.id
RETURNING j.*;
```

Lease heartbeat, shutdown cancellation, bounded concurrency, log metric queue lag, backoff + jitter, max attempt, manual retry berizin. Jangan menjalankan dua job provisioning tenant sama bersamaan.

## At least once; idempotency

Pekerjaan dapat dieksekusi lebih dari sekali karena worker crash/retry. Semua aksi memakai deterministic `profile_key`, unique constraints, step checkpoint, hash config, idempotency key. Filesystem mutations memakai transaction intent dan reconciliation; jangan mengasumsikan operasi filesystem dan DB satu transaksi ACID.

## Outbox

Untuk write bisnis, gunakan satu transaksi DB untuk mutation+`audit_logs`+`outbox_events`. Worker mem-publish update ke internal channels/background tasks. External Hermes webhook events masuk sebagai `ai_events` best-effort dan tidak dapat menggantikan audit bisnis.

## Job contracts

| Job | Trigger | Idempotency | Outcome |
|---|---|---|---|
| `tenant.provision` | create tenant | `(tenant_id,config_revision)` | profile ready/WAITING_INPUT |
| `tenant.activate` | verification + admin action | `(tenant_id,version)` | profile online |
| `tenant.suspend` | superadmin | `(tenant_id,reason_event)` | parked & deny writes |
| `bot.rotate` | admin | `(tenant_id,rotation_id)` | verified bot |
| `document.extract` | file uploaded/revised | `(file_id,version,sha256)` | READY/FAILED |
| `document.purge` | retention ended | `(file_id,version)` | physically removed |
| `tenant.reconcile` | interval | `(tenant_id,time_bucket)` | observed state updated |
| `storage.orphan_cleanup` | interval | `(tenant_id,time_bucket)` | stale orphan handled |
| `crawl.tick` | interval 1 menit | `time_bucket` | enqueue `crawl.run` untuk sumber `next_run_at <= now()` |
| `crawl.run` | tick atau manual | `(source_id,time_bucket)` / run id manual | halaman baru/berubah → `document.extract`; `crawl_runs` final |

## Telemetry HMAC ingest

Webhook receiver 1) rate limits, 2) validates body size, 3) verifies signature + replay freshness, 4) resolves profile identity to trusted tenant, 5) deduplicates event, 6) writes DB, 7) returns 2xx. Hermes outbound webhook signature format must be validated against **pinned** version; custom gateway hook forwarder may be needed to add stable event_id or timestamp.

## Sample retry policy

Transient network issues: 3–5 attempts with exponential backoff (e.g. 10s, 30s, 2m, 10m, 30m). Bad token/permission/security test failures fail closed into WAITING_INPUT/FAILED_SECURITY and require correction; never infinite retry.
