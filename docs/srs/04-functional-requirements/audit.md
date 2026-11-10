# FR — Audit

Sumber: `00001_init.sql` tabel `audit_logs`, `tenants.go` `auditLogs`, `labels.ts`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-AUD-001 | Superadmin harus dapat membaca audit tanpa `tenant_id` (semua) atau dengan `tenant_id`. Pengguna lain harus mengirim `tenant_id`, menjadi anggota tenant itu, dan memiliki `audit:read`. | Must | [CONFIRMED] | `auditLogs` | STAFF dengan `tenant_id` miliknya: 403. ADMIN: 200 hanya untuk tenant itu. Tanpa query untuk non-superadmin: 400 `TENANT_REQUIRED`. |
| FR-AUD-002 | Pembuatan tenant harus menulis audit `actor_type=USER`, `action=TENANT_CREATED`, `target_type=TENANT` dalam transaksi yang sama dengan insert tenant. | Must | [CONFIRMED] | `CreateTenant` | Jika insert audit gagal, tenant tidak ter-commit. |

### Aksi yang sudah punya label Indonesia di UI

`TENANT_CREATED`, `TENANT_UPDATED`, `TENANT_SUSPENDED`, `TENANT_ONBOARDING`, `MEMBER_CREATED`, `MEMBER_ROLE_UPDATED`, `FILE_UPLOADED`, `FILE_DELETED`, `FILE_RENAMED`, `FOLDER_CREATED`, `PROFILE_UPDATED`, aksi Hermes `HERMES_*`, `KNOWLEDGE_SYNCED`, `ASSISTANT_SYNCED`, `KNOWLEDGE_IMAGES_READ`, aksi Telegram, dan tiga aksi percakapan. Sumber: `auditLabels` di `frontend/src/lib/labels.ts`.

[INFERRED] Setiap aksi berlabel itu benar-benar ditulis server pada jalur sukses yang sesuai. Label membuktikan kontrak tampilan, bukan bahwa setiap handler memanggil `RecordAudit` sebelum 2xx. PRD NFR-05 menuntut audit pada setiap write bisnis sebelum respons sukses. Pembuktian "semua write" tidak dilakukan baris-per-baris di SRS ini; celahnya OQ-11.

Kolom yang wajib ada pada baris: `actor_type`, `action`, `occurred_at`, `details_redacted`. `tenant_id` boleh null (tipe kolom tanpa `NOT NULL`).
