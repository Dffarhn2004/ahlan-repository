# 14 — Kontrak REST API (usulan internal platform)

**Prefix:** `/api/v1`; **Auth:** session cookie HTTP-only untuk browser, `Authorization: Bearer` untuk layanan tepercaya; **Response:** JSON kecuali download/multipart. Semua id UUID; schema OpenAPI akan dibuat dari kontrak ini pada implementasi.

## Standard envelope

```json
{
  "data": {"id": "uuid"},
  "meta": {"request_id": "req_01", "pagination": null}
}
```

Error:

```json
{
  "error": {"code": "TENANT_FORBIDDEN", "message": "Akses tidak diizinkan"},
  "meta": {"request_id": "req_01"}
}
```

Kode umum: `400` validation, `401` unauthenticated, `403` forbidden, `404` not found (jangan ungkap tenant lain), `409` conflict/idempotency, `413` too large, `415` unsupported type, `422` business rule, `429` rate limit, `500` redacted, `503` dependency offline.

## Superadmin endpoints (admin scope global)

| Method | Endpoint | Deskripsi | Mutasi/audit |
|---|---|---|---|
| GET | `/admin/tenants` | List/search/filter tenant | - |
| POST | `/admin/tenants` | Create tenant + enqueue provisioning | `tenant.create` |
| GET | `/admin/tenants/{id}` | Detail tenant + profile/bot status | - |
| PATCH | `/admin/tenants/{id}` | Update metadata/kuota | `tenant.update` |
| POST | `/admin/tenants/{id}/activate` | Verify gates + start profile | `tenant.activate` |
| POST | `/admin/tenants/{id}/suspend` | Park/deny tools | `tenant.suspend` |
| POST | `/admin/tenants/{id}/resume` | Verify/resume | `tenant.resume` |
| POST | `/admin/tenants/{id}/archive` | Archive soft-delete | `tenant.archive` |
| GET | `/admin/tenants/{id}/provisioning` | Jobs/steps/errors redacted | - |
| POST | `/admin/tenants/{id}/provisioning/retry` | Retry failed job (idempotent) | `provision.retry` |
| PATCH | `/admin/crawl/sources/{id}` | Paksa nonaktif (`forced_disabled`) / ubah batas | `crawl.force_disable` |
| GET | `/admin/analytics` | Tenant & usage summary | - |
| GET | `/admin/audit-logs` | Filtered audit | - |

## Tenant admin endpoints

| Method | Endpoint | Deskripsi | Mutasi/audit |
|---|---|---|---|
| GET | `/me/tenant` | Current tenant profile | - |
| PATCH | `/me/tenant` | Public display/contact fields | `tenant.settings_update` |
| GET | `/me/bot` | Telegram bot status, masked info | - |
| POST | `/me/bot/connect` | Validate BotFather token + bind | `bot.connect` |
| POST | `/me/bot/rotate` | Replace token (controlled) | `bot.rotate` |
| GET | `/me/files` | List files by parent folder id | - |
| POST | `/me/folders` | Create folder | `folder.create` |
| PATCH | `/me/folders/{id}` | Rename/move folder | `folder.update` |
| DELETE | `/me/folders/{id}` | Soft-delete folder | `folder.delete` |
| POST | `/me/files` | Upload multipart | `file.upload` |
| GET | `/me/files/{id}` | File metadata + index status | - |
| GET | `/me/files/{id}/download` | Stream original | - |
| GET | `/me/files/{id}/preview` | Safe preview | - |
| PATCH | `/me/files/{id}` | Rename/move | `file.update` |
| DELETE | `/me/files/{id}` | Soft delete | `file.delete` |
| GET | `/me/crawl/sources` | List sumber crawl + jadwal + status terakhir (P1) | - |
| POST | `/me/crawl/sources` | Tambah sumber (validasi SSRF/kuota) | `crawl.source_create` |
| PATCH | `/me/crawl/sources/{id}` | Ubah jadwal/batas/enable | `crawl.source_update` |
| DELETE | `/me/crawl/sources/{id}` | Hapus sumber; hasil crawl di-soft-delete | `crawl.source_delete` |
| POST | `/me/crawl/sources/{id}/run` | Crawl manual; `202`, `409 CRAWL_ALREADY_RUNNING`, `429` bila < 10 menit | `crawl.run_manual` |
| GET | `/me/crawl/sources/{id}/runs` | Riwayat run + hitung halaman | - |
| GET | `/me/crawl/sources/{id}/pages` | Halaman terkumpul + status index | - |
| PATCH | `/me/crawl/pages/{id}` | Exclude / include halaman | `crawl.page_update` |
| GET | `/me/ai/events` | Recent observed AI events | - |
| GET | `/me/ai/health` | Runtime health/last event | - |
| POST | `/me/ai/test-question` | Controlled test; not raw unrestricted shell | `ai.test` |

## System/internal tool APIs

| Method | Endpoint | Security |
|---|---|---|
| POST | `/internal/hermes/events` | Per-profile HMAC, replay guard, schema validate |
| POST | `/internal/hermes/tools/search-knowledge` | mTLS or scoped service credential bound to profile |
| GET | `/internal/hermes/tools/admission-info` | Profile-bound tenant context |
| POST | `/internal/hermes/tools/create-lead` (P1) | Profile-bound + consent/idempotency |
| GET | `/health/live` | No data, unauthenticated |
| GET | `/health/ready` | Network/internal as appropriate |

## Create tenant example

```http
POST /api/v1/admin/tenants
Idempotency-Key: 5d0...client-generated
Content-Type: application/json

{"name":"Ponpes Al-Falah","slug":"al-falah","admin_email":"admin@example.test"}
```

```json
{
  "data": {
    "tenant_id": "generated-uuid",
    "lifecycle_status": "ONBOARDING",
    "provisioning_run_id": "generated-uuid"
  },
  "meta": {"request_id": "req_123"}
}
```

Return `202 Accepted` karena provisioning asynchronous. Slug hanya untuk display/routing web, bukan filesystem root atau profile key.

## Profil ponpes: field Google Maps

```http
PATCH /api/v1/me/tenant
Content-Type: application/json

{"google_maps_url":"https://maps.app.goo.gl/AbCdEf123","website_url":"https://alfalah.example.test"}
```

`GET /me/tenant` mengembalikan `google_maps_url`, `latitude`, `longitude`, `website_url`. Kirim `null` untuk menghapus. URL tidak lolos allowlist/skema → `422 MAPS_URL_INVALID`. Server tidak membuka URL tersebut. Tool `admission-info` menyertakan `google_maps_url` bila terisi.

## Create crawl source example (P1)

```http
POST /api/v1/me/crawl/sources
Idempotency-Key: crawl-src-1
Content-Type: application/json

{"url":"https://alfalah.example.test","kind":"SITE","max_depth":2,"max_pages":50,
 "schedule":{"mode":"INTERVAL","interval_minutes":1440}}
```

Respons `201`; error domain: `CRAWL_URL_FORBIDDEN` (SSRF/skema/port), `CRAWL_SOURCE_LIMIT`, `CRAWL_SCHEDULE_TOO_FREQUENT`, `CRAWL_ALREADY_RUNNING`. Tenant tidak dapat memilih `tenant_id`; sumber tenant lain → 404.

## File upload example

```http
POST /api/v1/me/files
Idempotency-Key: upload-123
Content-Type: multipart/form-data

parent_folder_id: <uuid-or-null>
file: <binary PDF>
```

Return `201 Created` ketika file benar-benar persisten dan metadata committed, status `PROCESSING` bila ekstraksi belum selesai. `GET` on cross-tenant IDs returns 404; tidak ada `tenant_id` yang dipilih bebas oleh browser.

## Internal credential flow

Go maps authenticated bot profile credential → internal tenant ID. API mengabaikan field `tenant_id` di prompt/model arguments. Every tool call logs authorized scope and request ID without secret bodies.

## Pagination/filter

List endpoint menggunakan `limit` max 100, opaque cursor untuk data besar; sort whitelisted; date range UTC; jangan memakai unrestricted filter/sort SQL string. Perubahan mutasi penting butuh `Idempotency-Key` yang dipersist sampai safe retention.
