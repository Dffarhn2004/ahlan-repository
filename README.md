# PRD — SaaS AI Penerimaan Santri (Go × Hermes × Telegram)

**Versi:** 1.0 (baseline MVP)  
**Tanggal:** 24 Oktober 2026  
**Status:** Draft untuk implementasi; keputusan terbuka dan dependensi dibedakan secara eksplisit.  
**Bahasa:** Indonesia

## Ringkasan satu paragraf

Platform B2B SaaS untuk banyak pondok pesantren. Perusahaan mengelola tenant via superadmin; tiap ponpes punya admin, file manager, bot Telegram, dan satu Hermes Profile tersendiri. Calon santri menghubungi bot Telegram milik ponpes secara langsung; Hermes menjawab menggunakan dokumen resmi yang disimpan **sebagai file fisik di volume Hermes**, dengan pembatasan akses per tenant. Backend **Go (Gin + pgx + sqlc)** menjadi control plane untuk provisioning, otorisasi, berkas, aktivitas, monitoring, dan data pendaftaran. PostgreSQL menyimpan metadata dan data bisnis, **bukan binary PDF/JPG**.

## Keputusan arsitektur yang sudah disepakati

| Keputusan | Status | Implikasi |
|---|---|---|
| Go menggantikan NestJS | Disetujui | Modular monolith, API dan worker terpisah |
| Telegram terhubung langsung ke Hermes Profile | Disetujui, **bersyarat keamanan** | Go tidak mem-proxy seluruh pesan; wajib validasi akses bot publik |
| Satu tenant = satu profile Hermes + satu bot Telegram | Disetujui | Isolasi state; pengamanan filesystem dibuat terpisah |
| Folder dokumen ada pada storage/volume Hermes | Disetujui | `DataPonpes/{tenant_id}` dikelola UI admin melalui API Go |
| Storage eksternal (S3/MinIO) tidak diperlukan untuk MVP | Disetujui | Backup volume dan rekonsiliasi file wajib |
| Profile dibuat otomatis ketika tenant di-onboard | Disetujui | Worker/provisioner idempoten, profile awalnya parked |
| PostgreSQL dipakai untuk data operasional | Disetujui | Metadata file, permissions, logs, leads, jobs |
| Retrieval terotorisasi per-tenant, bukan percaya instruksi prompt | Wajib | Multiplex: native filesystem/terminal tools yang tidak tersandbox **dinonaktifkan** untuk bot publik |
| Production public bot hanya setelah security gate lulus | Wajib | Default public access Hermes tidak boleh diasumsikan aman |

## Peta dokumen

### 00 — Produk
- [01 — Visi dan KPI](00_Produk/01_visi_kpi.md)
- [02 — Ruang lingkup & kebutuhan](00_Produk/02_ruang_lingkup_requirements.md)
- [03 — Pengguna, hak akses & user journey](00_Produk/03_persona_role_userjourney.md)
- [04 — UX superadmin dan admin ponpes](00_Produk/04_ux_dashboard.md)

### 01 — Arsitektur
- [05 — Arsitektur & keputusan isolasi](01_Arsitektur/05_arsitektur_sistem.md)
- [06 — Lifecycle tenant](01_Arsitektur/06_lifecycle_tenant.md)
- [07 — Provisioning Hermes](01_Arsitektur/07_provisioning_hermes.md)
- [08 — Integrasi Telegram](01_Arsitektur/08_telegram.md)
- [09 — File manager DataPonpes](01_Arsitektur/09_file_manager.md)
- [10 — Knowledge & retrieval](01_Arsitektur/10_knowledge_retrieval.md)
- [11 — Leads dan alur pendaftaran](01_Arsitektur/11_admission_workflow.md)
- [12 — Monitoring & observability](01_Arsitektur/12_logging_observability.md)
- [25 — Crawl sumber web & field profil ponpes (Google Maps)](01_Arsitektur/25_crawl_sumber_web.md)

### 02 — Implementasi Go / Kontrak Teknis
- [13 — Model data & DDL awal](02_Teknis/13_model_database.md)
- [14 — Kontrak REST API](02_Teknis/14_api_contract.md)
- [15 — Keamanan & privasi](02_Teknis/15_security_privacy.md)
- [16 — Worker, jobs, event](02_Teknis/16_worker_events.md)
- [17 — Struktur repositori Go](02_Teknis/17_go_repository.md)
- [18 — Deployment, operasi & backup](02_Teknis/18_deployment_ops.md)

### 03 — Eksekusi
- [19 — Roadmap dan backlog](03_Delivery/19_roadmap_backlog.md)
- [20 — Acceptance criteria dan testing](03_Delivery/20_acceptance_testing.md)
- [21 — Risiko & ADR](03_Delivery/21_risks_adr.md)
- [22 — Pertanyaan terbuka](03_Delivery/22_open_questions.md)

### 04 — Referensi dan template
- [23 — Referensi Hermes dan verifikasi](04_Referensi/23_referensi_teknis.md)
- [24 — Glosarium](04_Referensi/24_glosarium.md)
- [Template SOUL](templates/SOUL.template.md)
- [Template kebijakan dokumen](templates/KNOWLEDGE_POLICY.template.md)

## Cara membaca & mengerjakan

1. Mulai dari `01`, `02`, `05`, `06`, `15`, `20` untuk scope MVP dan security gate.
2. Bangun Go API dan PostgreSQL menurut `13`, `14`, `17`; kerjakan provisioning menurut `07` dan `16`.
3. Implementasikan file manager `09` dan pembacaan dokumen `10` sebelum merilis bot ke publik.
4. Lakukan test lintas tenant serta threat test **sebelum** `ACTIVE` dapat dipakai publik.
5. Gunakan `22` untuk keputusan yang harus dikonfirmasi saat development; jangan mengisi asumsi dengan perilaku Hermes yang tidak diverifikasi.

## Definisi prioritas

- **P0**: diperlukan untuk pilot aman dengan 2 tenant dan pengujian lintas-tenant.
- **P1**: diperlukan untuk beta berbayar/operasional lebih matang.
- **P2**: perluasan produk setelah validasi pilot.

## Batas keandalan PRD

Dokumentasi Hermes adalah produk yang berubah cepat. Referensi produk dan contoh CLI dikumpulkan per 24 Oktober 2026. **Konfigurasi persis, API yang dapat diotomasi, slash command Telegram dan perilaku gateway harus dibuktikan lewat smoke test pada versi Hermes yang dipin sebelum implementasi produksi.** Tidak semua contoh payload dalam dokumen merupakan API asli Hermes; yang berupa rancangan platform diberi label *kontrak internal usulan*.
