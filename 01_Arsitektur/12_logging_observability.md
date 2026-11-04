# 12 — Monitoring, telemetry, log dan audit

## Empat jenis catatan, jangan dicampur

1. **Business audit (durable):** siapa mengubah tenant, file, bot, profile, role, approval; di-commit atomik dengan mutasi API.
2. **Provisioning log (durable):** job, step, retry, error, duration, config version; redacted.
3. **AI telemetry (best-effort):** event/session/turn/model/tool/latency dari Hermes signed hooks.
4. **Application/infrastructure logs:** trace HTTP, DB errors, gateway health, system metrics; retensi berbeda.

## Batas Hermes hooks

Hermes mempunyai gateway hooks, plugin hooks dan outbound signed webhooks. Pada outbound webhook, delivery async dan best-effort: queue bounded; error retry terbatas; event bisa drop jika receiver tidak tersedia. Oleh karena itu dashboard **tidak boleh** menyatakan 'seluruh chat 100% tercatat' hanya dari hook. `post_llm_call` juga hanya untuk completion sukses; event error/interrupted memerlukan signal tambahan. Jangan menunggu webhook selesai di jalur balas Telegram.

## Event envelope internal (usulan, bukan payload Hermes asli)

```json
{
  "schema_version": 1,
  "event_id": "018d...uuid",
  "tenant_id": "resolved-on-server",
  "profile_key": "t_uuid",
  "event_type": "ai.turn.completed",
  "session_external_id": "opaque",
  "turn_id": "opaque",
  "occurred_at": "2026-10-24T08:15:00Z",
  "received_at": "2026-10-24T08:15:01Z",
  "model": "provider/model",
  "duration_ms": 1450,
  "usage": {"input_tokens": 300, "output_tokens": 75},
  "status": "success"
}
```

Receiver: verify HMAC with per-profile secret, timestamp freshness/replay as supported by our wrapper, resolve `profile_key` to `tenant_id` from credential mapping **never from unsigned JSON**, dedup by source event id/hash, validate schema, enqueue/store, return 2xx only after durable ingest. Rate limit + body size. Restrict network ingress, hide PII by default.

## Usage & cost

Gunakan provider usage aktual jika tersedia pada events `post_api_request`/auxiliary calls, bukan mengestimasi token dari karakter dan menampilkan sebagai pasti. `model_pricing` versioned per effective date; `ai_usage` stores quantity+cost estimate, currency, source, confidence. Track per-call, per-session, per-tenant; tidak double count tools/retry atau duplicate events.

## Dashboard superadmin

- Tenant count, online bot count, profiles parked, failed provisioning.
- Volume conversation/turn events yang *terobservasi* per tenant/range; explain mungkin tidak lengkap.
- Error types, latency percentile, estimated token/cost usage, queue backlog, index status.
- Per-tenant drill down, secure support access and audit.

## Dashboard admin ponpes

- Status bot, dokumen `READY/FAILED`, pertanyaan terobservasi, konten FAQ paling banyak diminta (jika ada consent/retention), escalations dan link human admin.
- Tidak menampilkan tenant lain atau secrets.

## Retention/PII

Usulan: audit 1 tahun, provisioning 90 hari, detailed AI transcripts OFF by default atau maximum 30 hari dengan kebijakan dan pemberitahuan; agregat tanpa identitas lebih lama. Nilai final harus disesuaikan kewajiban legal dan kontrak, dan dibuat configurable. Hanya simpan plaintext chat bila ada kebutuhan dan dasar yang jelas.

## Alert P0

Bot offline > 5 menit, tenant stuck provisioning > 10 menit, webhook errors > threshold, repeated invalid token, cross-tenant access attempt, disk > 80%, document extraction failures, scheduled backups failed.
