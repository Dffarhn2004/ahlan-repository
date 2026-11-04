# 21 — Risk register & Architecture Decision Records

## Risk register

| Risk ID | Risiko | Dampak | Kemungkinan | Mitigasi | Gate |
|---|---|---|---|---|---|
| R-01 | Hermes profile bukan filesystem sandbox | Critical | High jika native local tools | Multiplex + scoped Go tool; alternatif container per tenant | P0 |
| R-02 | Public Telegram allow-all memicu command tak-aman | Critical | Unverified | Real Telegram security test; block public go-live | P0 |
| R-03 | Hermes hooks drop saat overload/down | High untuk akurasi log | Medium | Jangan jadikan source of truth bisnis; outbox, observability gap | P0 |
| R-04 | Bot token bocor | High | Medium | Mask, encrypt, restricted secret manager, rotation | P0 |
| R-05 | Provisioned profile auto-discover sebelum parked | High | Medium | Staging/atomic parked + integration test | P0 |
| R-06 | File path traversal/symlink race | Critical | Medium | Safe-root FS primitives, no relative traversal | P0 |
| R-07 | PDF scan gagal dibaca | Medium | High | Processing status + OCR/backlog; abstain | P0/P1 |
| R-08 | Data anak/PII tersimpan di chat | High | High | Minimal collection, privacy, retention & access controls | P0 |
| R-09 | Single host/storage crash | High | Medium | Backup & recovery drill, sizing, monitoring | P0 |
| R-10 | Kinerja multiplex tidak skala ratusan profiles | High | Unverified | Capacity benchmark and horizontal partition | P1 |
| R-11 | Hermes breaking changes saat upgrade | High | High | Pin version, staged upgrade, suite | P0 |
| R-12 | Information stale antar tahun akademik | Medium | High | File version, validity dates, abstain on conflicting docs | P0 |
| R-13 | Crawler dipakai untuk SSRF ke jaringan internal/metadata | Critical | Medium | IP validation, pinned connect, egress terpisah, suite AT-023..026 | P1 gate |
| R-14 | Konten web basi/salah/berisi prompt injection masuk jawaban bot | High | Medium | Untrusted text, sitasi URL+tanggal, unggahan > crawl, exclude page | P1 |
| R-15 | Crawl membebani situs ponpes / melanggar robots / kuota storage jebol | Medium | Medium | robots.txt, delay per host, batas halaman, interval >= 6 jam, kuota bersama | P1 |

## ADR-001 — Go sebagai control plane

**Context:** Backend tidak perlu mengimplementasikan bot AI. **Decision:** modular monolith Go (Gin+pgx+sqlc). **Consequence:** type-safe explicit code, reliable worker; lebih banyak implementasi manual auth/API dibanding framework besar.

## ADR-002 — Telegram langsung ke Hermes

**Context:** Minimalkan reimplementasi dan latency dari proxy Go. **Decision:** direct bot ↔ profile, **conditional public security**. **Fallback:** restricted ingress jika Hermes direct public command gating tidak aman. **Consequence:** telemetry messages tidak bisa dijamin lengkap melalui webhook best-effort.

## ADR-003 — File asli di Hermes DataPonpes

**Context:** Admin ingin mengelola PDF/JPG di lingkungan Hermes. **Decision:** persistent folder `/DataPonpes/<uuid>` pada volume Hermes, metadata di PostgreSQL. **Consequence:** butuh filesystem-safe CRUD, backup combined DB/volume; object storage bisa ditambah nanti tanpa mengganti UX.

## ADR-004 — Secure retrieval pada multiplex

**Context:** Profile bukan sandbox. **Decision:** bot publik multiplex tidak memiliki unrestricted filesystem/terminal tools; gunakan Go `search_tenant_documents` yang dibatasi tenant. **Alternative:** per-tenant container dengan read-only mount untuk native `read_file`.

## ADR-005 — Provisioning asynchronous dan parked

**Context:** CLI/gateway/token validation bisa gagal. **Decision:** queued steps, job lease/idempotency, create profile offline, activate setelah security tests. **Consequence:** perlu progress UI, reconciler, manual recovery.

## ADR-006 — Logging layered

**Decision:** business audit transactionally durable di Go; Hermes outbound hooks sebagai telemetry best-effort. **Consequence:** analytic counts mungkin undercount, harus dinyatakan di UI/report.

## ADR-007 — PostgreSQL jobs dulu

**Decision:** `FOR UPDATE SKIP LOCKED` + jobs/outbox untuk pilot; Redis atau queue dedicated setelah butuh scale. **Consequence:** lebih sedikit komponen awal, observability queue tetap dibutuhkan.

## ADR-008 — Crawl dijalankan Go worker, bukan tool browser Hermes

**Context:** Admin ingin Hermes "meng-crawl" situs ponpes secara berkala dan manual. **Decision:** crawler adalah modul Go (`internal/crawler`) yang dipicu job queue; hasilnya menjadi berkas di `DataPonpes/<tenant>/_crawl` dan masuk pipeline ekstraksi/retrieval biasa. Bot publik tetap tanpa browser/web tool (ADR-004, SEC-02). Hermes dipakai hanya untuk membaca hasil, dan opsional (P1+) untuk peringkasan di profile internal non-publik. **Alternative:** web tool/cron Hermes per profile — ditolak untuk MVP karena kontrol SSRF, kuota, audit, dan manual-trigger tidak terpusat; dapat ditinjau ulang setelah SETUP-06. **Consequence:** satu jalur penjadwalan (PostgreSQL), halaman SPA murni butuh render (backlog), harus ada egress terpisah.

## Change control

Setiap perubahan ADR wajib mencantumkan alasan, risiko, alternatif, hasil pengujian, migrasi, dan versi keputusan. Jangan mengubah ADR-002 atau ADR-004 diam-diam hanya agar demo terlihat jalan.
