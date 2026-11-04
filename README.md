# PRD — SaaS AI Penerimaan Santri (Go × Hermes × Telegram)

**Versi:** 1.0 (baseline MVP)  
**Tanggal:** 24 Oktober 2026  
**Status:** Draft untuk implementasi; keputusan terbuka dan dependensi dibedakan secara eksplisit.  
**Bahasa:** Indonesia

## Ringkasan satu paragraf

Platform B2B SaaS untuk banyak pondok pesantren. Perusahaan mengelola tenant via superadmin; tiap ponpes punya admin, file manager, bot Telegram, dan satu Hermes Profile tersendiri. Calon santri menghubungi bot Telegram milik ponpes secara langsung; Hermes menjawab menggunakan dokumen resmi yang disimpan **sebagai file fisik di volume Hermes**, dengan pembatasan akses per tenant. Backend **Go (Gin + pgx + sqlc)** menjadi control plane untuk provisioning, otorisasi, berkas, aktivitas, monitoring, dan data pendaftaran. PostgreSQL menyimpan metadata dan data bisnis, **bukan binary PDF/JPG**.

## Gambaran fitur

Platform SaaS B2B untuk banyak pondok pesantren (ponpes). Setiap ponpes (tenant) mendapat:

- sebuah **asisten AI** (profile Hermes) yang menjawab pertanyaan calon santri / wali lewat **Telegram** dan **WhatsApp**,
- **file manager** untuk dokumen resmi (brosur, syarat pendaftaran, biaya, dll.) yang menjadi sumber jawaban AI,
- **dashboard admin** (kelola berkas, percakapan, pertanyaan, kontak, link publik, audit),
- **halaman publik** bergaya Linktree (`/p/{slug}`) berisi profil dan tautan ponpes.

Superadmin perusahaan meng-onboard tenant; admin ponpes mengelola isinya sendiri.

## Arsitektur sekilas

```
Calon santri ──Telegram/WhatsApp──▶ Hermes (1 profile per tenant)
                                        │  baca dokumen DataPonpes/{tenant_id}
                                        │  kirim event percakapan
                                        ▼
Browser ──▶ Frontend (Next.js) ──▶ API Go (Fiber) ──▶ PostgreSQL
                                        │
                                        ├─▶ Worker (job async, crawler)
                                        └─▶ Hermes Bridge (API sempit, lokal)
                                              └─ buat profile, set channel, tulis knowledge
```

Prinsip penting:

- **Go adalah control plane**: provisioning, otorisasi, file, audit, data bisnis. Pesan chat tidak lewat Go; Telegram/WhatsApp terhubung langsung ke Hermes.
- **Satu tenant = satu profile Hermes** (+ satu bot). Dokumen disimpan sebagai **file fisik** di `DataPonpes/{tenant_id}`, bukan di database. PostgreSQL hanya menyimpan metadata dan data operasional.
- **Hermes Bridge** hanya punya route tetap dan bertipe (lihat [bridge_proto.go](backend/internal/hermes/bridge_proto.go)); tidak ada route yang menerima perintah, nama container, atau path bebas. Ini batas keamanan antara API dan host.
- Hermes mengirim event ke `POST /api/v1/internal/hermes/events`, diautentikasi header `X-Hermes-Event-Key`, bukan sesi login.

> Catatan: PRD (24 Okt 2026) menyebut Gin + sqlc; implementasi saat ini memakai Fiber (`backend/cmd/api/main_fiber.go`) dan pgx. Jika berbeda, kode yang berlaku.

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

## Struktur repo

| Path | Isi |
|---|---|
| `00_Produk` … `04_Referensi` | Dokumen PRD (bahasa Indonesia), dari visi sampai ADR dan glosarium |
| `backend/` | Kode Go (modul tunggal) |
| `backend/cmd/api` | Server HTTP (Fiber) |
| `backend/cmd/worker` | Worker job async |
| `backend/cmd/hermes-bridge` | Bridge lokal ke Hermes |
| `backend/cmd/migrate`, `cmd/seed` | Migrasi DB dan seed admin awal |
| `backend/internal/httpapi` | Handler & routing (auth, tenants, files, knowledge, telegram, whatsapp, crawl, links, dll.) |
| `backend/internal/hermes` | Client/server bridge, validasi, integrasi WhatsApp |
| `backend/internal/repository`, `db` | Akses PostgreSQL (pgx) |
| `backend/internal/authz` | Peta izin per peran |
| `backend/internal/storage` | Penyimpanan file per tenant |
| `backend/internal/crawler` | Crawl sumber web (robots.txt, guard, diff, jadwal) |
| `backend/internal/jobs` | Antrean job |
| `backend/db/migrations` | Migrasi SQL (`00001` … `00019`) |
| `backend/deploy` | `docker-compose.yml` lokal, unit systemd, skrip verifikasi bridge |
| `frontend/` | Next.js (App Router): `login`, `admin`, `files`, `audit`, `settings`, `p/[slug]`, halaman publik PSB |
| `templates/` | Template `SOUL` dan `KNOWLEDGE_POLICY` untuk profile Hermes |
| `deploy-vps.sh` | Deploy ke VPS: upload → build → migrasi → restart → health check → rollback otomatis |
| `migrate-hermes.sh` | Migrasi data Hermes antar host |

## Menjalankan secara lokal

Prasyarat: Go, Node.js, Docker (untuk PostgreSQL).

```bash
# 1. Database + backend lewat Docker
cd backend/deploy && docker compose up -d

# atau manual:
cd backend
cp .env.example .env          # sesuaikan DATABASE_URL, DATA_ROOT, SEED_ADMIN_*
go run ./cmd/migrate
go run ./cmd/seed
go run ./cmd/api              # http://localhost:8080
go run ./cmd/worker

# 2. Frontend (port 3000)
cd frontend
npm install
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
```

Nilai di `.env.example` hanya untuk pengembangan lokal. Jangan pakai kredensial itu di produksi.

Tes backend: `cd backend && go test ./...`. Tes integrasi Hermes membutuhkan lingkungan bridge (lihat [backend/deploy/hermes-bridge/README.md](backend/deploy/hermes-bridge/README.md)).

## Deploy

`./deploy-vps.sh [backend|frontend|all]` dari Git Bash. Skrip bergantung pada alias SSH `chatbot-ahlan-server` dan `ahlan-server`, serta service systemd `ahlan-api`, `ahlan-worker`, `ahlan-hermes-bridge`. Env opsional: `SKIP_TESTS=1`, `FORCE_NPM_CI=1`.

## Yang sengaja tidak ada di Git

Lihat [.gitignore](.gitignore) dan [backend/.gitignore](backend/.gitignore): `backend/data/` (berkas unggahan tenant), `.hermes-staging/` (arsip migrasi Hermes), `.env*` (kecuali `.env.example`), `node_modules/`, `.next/`.

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
