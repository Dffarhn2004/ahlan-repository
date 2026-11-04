# 02 — Ruang lingkup, prioritas, kebutuhan fungsional

## Batas produk

**In-scope P0:** login + RBAC, superadmin CRUD tenant, portal admin tenant, penyediaan Hermes Profile otomatis, koneksi satu bot Telegram per tenant, folder `DataPonpes`, upload/list/download/rename/delete berkas, pembacaan dokumen terotorisasi, QA yang aman (fallback jika tidak tahu), audit perubahan dokumen, logs event, monitoring sederhana, suspend/resume, job retries, dua-tenant pilot.

**P1:** crawl sumber web ponpes terjadwal + manual (`25_crawl_sumber_web.md`), lead management dan form pendaftaran terstruktur, human handoff, document version restore, index/semantic search, penggunaan token dan perkiraan cost, kuota tenant, admin tambahan, alert & backup restore drill.

**P2:** subscription billing, self-serve onboarding penuh, WhatsApp/web-chat, OCR lanjutan, knowledge approval workflow, analitik conversion, tenant migration lintas host.

## Requirements catalogue

| ID | Level | Pernyataan kebutuhan | Kriteria lulus ringkas |
|---|---|---|---|
| FR-001 | P0 | Superadmin membuat tenant | Tenant + job tersimpan atomik; tidak langsung aktif |
| FR-002 | P0 | Worker membuat profil Hermes secara idempoten | Retry tidak menduplikasi profil |
| FR-003 | P0 | Setiap tenant punya bot Telegram berbeda | `bot_id` unik, profile tepat |
| FR-004 | P0 | Admin tenant hanya melihat tenant miliknya | Semua endpoint dan file ops enforce scoped access |
| FR-005 | P0 | File disimpan fisik di folder DataPonpes Hermes | Verify checksum + metadata DB |
| FR-006 | P0 | Admin mengelola folder dan file | CRUD + safe path + audit |
| FR-007 | P0 | Hermes dapat membaca dokumen tenant sendiri | Negative test lintas-tenant gagal tertutup |
| FR-008 | P0 | AI abstain bila jawaban tidak didukung dokumen | Eval set knowledge gap lulus |
| FR-009 | P0 | Tenant dapat di-suspend dan di-resume | Bot A mati/nyala tanpa memengaruhi B |
| FR-010 | P0 | Status/riwayat provisioning dapat dilihat superadmin | Job stages + error aman + retry |
| FR-011 | P0 | Audit perubahan admin tersimpan durable | Record before 2xx response API |
| FR-012 | P0 | Telemetry Hermes muncul di dashboard | Ingest signed, dedup, toleransi drop |
| FR-013 | P0 | Token Telegram tidak pernah dibaca kembali plaintext | Storage terenkripsi/secret mounted; UI masked |
| FR-014 | P0 | Telegram publik tidak dapat memicu administrative slash/tool | Security test spesifik versi Hermes lulus |
| FR-015 | P0 | Bot tidak dapat membaca config/secret/profile tenant lain | Defense-in-depth, bukan prompt-only |
| FR-016 | P0 | File upload berbahaya ditolak | Type/size/check/quarantine valid |
| FR-017 | P1 | Calon santri menjadi lead | User consent + create idempotent + status |
| FR-018 | P1 | Admin dapat takeover percakapan | Explicit handoff + status |
| FR-019 | P1 | Statistik usage/AI cost per tenant | Raw usage saat tersedia + cost versioned |
| FR-020 | P1 | Reindex/reprocess dokumen | Retry, status dokumen, rollback versi |
| FR-021 | P2 | Admin mengelola paket langganan | Billing tanpa mengganggu tenant data |
| FR-022 | P0 | Admin ponpes mengisi tautan Google Maps; bot dapat membagikannya | Validasi https + allowlist host; tampil di `admission-info`; tidak pernah di-fetch server |
| FR-023 | P1 | Admin menambah sumber crawl web milik ponpes | Validasi SSRF/skema/kuota; maks 5 sumber |
| FR-024 | P1 | Crawl berjalan terjadwal (interval/cron, `Asia/Jakarta`) | Interval efektif >= 6 jam; satu run aktif per sumber |
| FR-025 | P1 | Admin memicu crawl manual kapan saja | `202`; `409` bila sedang jalan; rate limit 1/10 menit |
| FR-026 | P1 | Hasil crawl masuk knowledge base dengan sitasi URL + tanggal | Pipeline ekstraksi sama dengan upload; halaman hilang tidak muncul di retrieval |
| FR-027 | P1 | Crawler aman dari SSRF dan menghormati robots.txt | Test IP privat/metadata/redirect/DNS-rebinding gagal tertutup |

## Nonfunctional requirements

| ID | Sasaran awal |
|---|---|
| NFR-01 | Kegagalan Go dashboard tidak langsung memutus semua bot (kecuali tool yang membutuhkan Go) |
| NFR-02 | REST API p95 < 400 ms untuk list/CRUD tanpa file besar; provisioning asynchronous |
| NFR-03 | Upload P0 <= 20 MB/file, maksimum 500 MB/tenant (configurable) |
| NFR-04 | File permissions: admin tenant hanya ke root tenant; Go bukan root FS seluruh Hermes |
| NFR-05 | Setiap write bisnis mempunyai audit trail `actor`, `tenant_id`, `action`, `request_id`, `occurred_at` |
| NFR-06 | Seluruh endpoint memiliki authn/authz, kecuali endpoint health dan callback bertanda tangan |
| NFR-07 | Backup terenkripsi, restorasi tenant dapat diuji; tetapkan RPO/RTO sebelum komersial |
| NFR-08 | Retry job eksponensial, idempotency key, dead-letter/manual retry |
| NFR-09 | File operation aman dari symlink race, traversal, MIME spoof, decompression bomb |
| NFR-10 | Security regression test dua tenant wajib lulus setiap rilis |

## Out-of-scope tegas untuk P0

- Tidak mempersilakan admin menulis `.env`, `config.yaml`, `SOUL.md` bebas dari file manager.
- Tidak memberikan terminal/shell/web-browser generik ke calon santri. Crawl web dilakukan oleh worker Go, bukan oleh bot publik.
- Tidak meng-crawl Google Maps, media sosial, atau situs yang butuh login.
- Tidak menjadikan AI penentu penerimaan santri atau penyetuju pembayaran.
- Tidak memproses data sensitif/anak di luar kebutuhan minimum alur informasi.
- Tidak menjanjikan log pesan 100% lengkap bila hanya bersumber dari hook Hermes.

## Requirement traceability

Implementasi `FR-001..010` dibagi pada dokumen 06–10; `FR-011..016` pada 12, 15, 16; API pada 14; validasi seluruh FR pada 20.
