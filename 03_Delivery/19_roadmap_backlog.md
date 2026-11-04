# 19 — Roadmap engineering & backlog eksekusi

## Tahap 0 — Environment & architecture spike (P0, gate)

**Goal:** bukti nyata versi Hermes yang dipin mendukung operasi profile & gateway yang direncanakan, serta menemukan cara aman melayani Telegram publik.

- SETUP-01: Buat test host/container Hermes, pin commit/tag dan image digest.
- SETUP-02: Test `profile create`, parked, multiplex serve/unserve, bot token duplicate detection, start/stop isolation.
- SETUP-03: Security spike public bot: access control DM, slash commands, native tools/skills, profile isolation.
- SETUP-04: Test delivery Hermes outbound events (success/error/retry/drop), scopes, payload shape.
- SETUP-05: Test file read PDF/JPG, OCR cases, one profile vs two profiles.
- SETUP-06 (opsional, crawl): Verifikasi di versi Hermes pinned apakah cron/scheduler per-profile dan profile internal non-publik untuk peringkasan teks hasil crawl tersedia dan aman. Keputusan default tetap: scheduler di Go.

**Exit:** Demo + test evidence. Jika public access langsung gagal security gate, revisi pilihan ingress/runtime; jangan lanjut mengaktifkan bot publik.

## Sprint 1 — Backend core

- CORE-01: Go scaffold, migrations, sqlc, auth session, superadmin MFA & RBAC.
- CORE-02: Tenant registry/create/update/read dengan UUID profile key.
- CORE-02b (FR-022): `tenant_profiles`, validasi + allowlist `google_maps_url`, `PATCH /me/tenant`, field di UI Informasi Ponpes, ikut `admission-info`.
- CORE-03: DB job runner/worker + outbox + idempotency.
- CORE-04: Audit mutation transactional.
- CORE-05: OpenAPI baseline + error codes.

**Exit:** Create tenant returns `202`, job queued, permissions enforced.

## Sprint 2 — Hermes lifecycle

- HERMES-01: Trusted provisioner adapter + mock.
- HERMES-02: Staged/parked profile provisioning; race prevention.
- HERMES-03: Render versioned SOUL/policy/config, secrets injection.
- HERMES-04: Telegram BotFather token connect, `getMe`, uniqueness, rotate.
- HERMES-05: Activate/suspend/resume/reconcile/health.

**Exit:** 2 tenants, 2 profiles, 2 Telegram bots, distinct states and no duplicated token.

## Sprint 3 — DataPonpes File Manager

- FILE-01: Safe tenant root abstraction, folder CRUD, file metadata.
- FILE-02: Upload/preview/download/rename/move/delete/trash, quotas, type limits.
- FILE-03: Audit file operations, orphan reconciler, integrity manifest.
- FILE-04: React tenant file manager.

**Exit:** Admin A controls only own files; negative tests pass.

## Sprint 4 — AI knowledge & event monitoring

- KNOW-01: Text extractor PDF and TXT, image/OCR handling policy.
- KNOW-02: Scoped `search_tenant_documents` tool API / allowlist.
- KNOW-03: SOUL safety response and 2-tenant Q&A eval set.
- OBS-01: Signed Hermes event receiver + dedup, redaction, basic dashboard.
- OBS-02: Provisioning timeline, bot status and tenant analytics.

**Exit:** Grounded Q&A and no leakage, event dashboard explicitly marked observed/best effort.

## Sprint 5 — Production pilot hardening

- SEC-01: Public Telegram command/tool lock-down (must pass).
- SEC-02: Cross-tenant security suite (API, FS, Hermes).
- OPS-01: Backup/restore smoke, secret rotation, deploy runbook.
- PERF-01: 2 tenant reliability/latency/load test.
- UX-01: Admin usability test and error handling.

**Exit:** Formal go/no-go based on `20_acceptance_testing.md`.

## Crawl sumber web (P1, bisa dipercepat setelah ADR-008)

- CRAWL-01: Migrasi `crawl_sources/runs/pages`, `files.source_type`, kuota sumber.
- CRAWL-02: Fetcher aman (SSRF guard, redirect re-validate, robots.txt, rate limit, batas ukuran) + egress terpisah.
- CRAWL-03: Ekstraksi teks utama HTML, hash, versioning, deteksi halaman hilang, integrasi `document.extract`.
- CRAWL-04: `crawl.tick` + `crawl.run` + trigger manual + one-active-run + audit.
- CRAWL-05: API `/me/crawl/*` dan UI Sumber Web (jadwal Manual/Interval/Harian, riwayat, exclude).
- CRAWL-06: Sitasi URL+tanggal pada retrieval; aturan prioritas unggahan > crawl; eval set konflik.
- CRAWL-07: Security suite crawler (AT-023..AT-028).

**Exit:** 2 tenant, jadwal harian + crawl manual berjalan, SSRF suite lulus, jawaban bersitasi URL.

## P1 after pilot

Lead/contact request, human handoff, semantic index/pgvector, token cost dashboard, multi-admin roles, document approvals/version history, backup automation, retention management.

## P2 after commercial validation

Subscription billing & quotas, self-service tenant provisioning, tenant migration, multichannel WhatsApp/web, advanced reports/analytics, enterprise per-tenant containers.

## Dependency chain

```text
Security spike
  → Auth/RBAC + tenant registry + DB jobs
  → Trusted Hermes provisioning
  → DataPonpes file manager
  → Secure scoped retrieval
  → Signed monitoring
  → Cross-tenant & Telegram public security tests
  → Pilot go-live
```

## Definition of Done per backlog ticket

Specification linked to FR/SEC ID, schema migration if needed, handler/service tests, integration test for mutations, permission negative test, audit event, OpenAPI update, CI passed, rollback plan, no secret in logs/screenshots, documented operational runbook.
