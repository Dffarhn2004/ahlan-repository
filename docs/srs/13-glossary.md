# 13 — Glosarium

Istilah di bawah dipakai SRS dengan arti yang terikat artefak repo. Entri yang hanya hidup di PRD ditandai.

| Istilah | Arti |
|---|---|
| AHLAN | Nama aplikasi default di peramban (`branding.ts`). |
| Tenant / pesantren | Baris `tenants`. Satu pelanggan B2B. |
| Superadmin | `users.is_platform_superadmin = true`. Peran string `SUPERADMIN` pada `/me`, bukan nilai `tenant_users.role`. |
| OWNER, ADMIN, STAFF, VIEWER | Nilai `tenant_users.role`. |
| Control plane | API Go: tenant, otorisasi, berkas, audit, job. |
| Data plane | Runtime Hermes yang membalas Telegram/WhatsApp. |
| Hermes profile | Baris `hermes_profiles` plus state di host Hermes. Satu per tenant. |
| Bridge | Proses `cmd/hermes-bridge` dengan route tetap. Produksi tidak memanggil Docker dari API. |
| DataPonpes | Direktori `{DATA_ROOT}/DataPonpes/{tenant_uuid}`. |
| Knowledge / pengetahuan | Berkas dan hasil crawl yang disinkronkan agar asisten mengutipnya. |
| Marker abstain | Frasa `belum tersedia di data kami`. |
| PSB | Penerimaan santri baru. `service_mode=PSB_ONLY` berarti nomor khusus alur itu. |
| GENERAL | Mode WhatsApp nomor bersama: tanya keperluan, lalu arahkan non-PSB ke admin. |
| TESTING / PUBLIC | Mode akses kanal. TESTING membatasi user id Telegram atau nomor WhatsApp. |
| Pairing | Penautan perangkat WhatsApp, batas 3 menit di API. |
| Crawl source | URL yang diambil worker. `forced_disabled` hanya diubah superadmin. |
| Linktree (sebutan PRD) | Halaman publik `/p/{slug}`, bukan produk Linktree. |
| Audit log | Baris `audit_logs`. |
| Event Hermes | POST internal dengan `X-Hermes-Event-Key`. |
| Handoff | Penyerahan percakapan ke manusia. Route internal `/internal/hermes/handoff` dan aksi audit `CONVERSATION_HANDOFF`. |
| Lead | Istilah PRD untuk prospek pendaftar. Tidak ada tabel lead di migrasi saat SRS ditulis. |
| Security gate | Istilah PRD untuk uji sebelum bot publik. Tidak ada objek gate di skema. |
| Kuota | `storage_quota_bytes`, default 524288000. |
| Sesi | Baris `sessions` dan cookie `session`. |
| WIB | Zona yang dipakai crawler, UTC+7. |
