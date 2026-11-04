# Ahlan — Gambaran Proyek

Panduan singkat untuk memahami repo ini. Detail keputusan dan kebutuhan ada di PRD ([README.md](README.md) dan folder `00_Produk` … `04_Referensi`).

## Apa ini?

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

## Bacaan lanjutan

1. [05 — Arsitektur & isolasi](01_Arsitektur/05_arsitektur_sistem.md)
2. [07 — Provisioning Hermes](01_Arsitektur/07_provisioning_hermes.md)
3. [14 — Kontrak REST API](02_Teknis/14_api_contract.md)
4. [15 — Keamanan & privasi](02_Teknis/15_security_privacy.md)
5. [24 — Glosarium](04_Referensi/24_glosarium.md)

> Catatan: PRD bertanggal 24 Okt 2026 menyebut Gin + sqlc; implementasi saat ini memakai Fiber (`cmd/api/main_fiber.go`) dan pgx. Jika berbeda, kode yang berlaku.
