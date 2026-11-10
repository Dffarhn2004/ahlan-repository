# 02 — Deskripsi umum

## 2.1 Perspektif produk

AHLAN adalah control plane web untuk banyak pesantren. Pesan calon santri tidak diproksi penuh oleh API Go. Telegram dan WhatsApp terhubung ke Hermes; Go mengatur tenant, berkas, otorisasi, dan menerima event. Sumber: `README.md` bagian arsitektur.

```
Pengunjung web ──▶ Next.js (frontend/)
Pengurus       ──▶ Next.js ──cookie sesi──▶ API Go Fiber (backend/cmd/api)
                                              │
                                              ├─ PostgreSQL (metadata)
                                              ├─ berkas DataPonpes/{tenant_id}
                                              ├─ worker (backend/cmd/worker)
                                              └─ Hermes bridge (backend/cmd/hermes-bridge)
Calon santri ──Telegram/WhatsApp──▶ Hermes profile tenant
Hermes ──X-Hermes-Event-Key──▶ POST /api/v1/internal/hermes/events
```

PRD 24 Oktober 2026 menyebut Gin + sqlc. Implementasi memakai Fiber dan pgx (`README.md`, `backend/go.mod`, `backend/cmd/api/main_fiber.go`). SRS mengikuti implementasi.

## 2.2 Fungsi utama

| Fungsi | Siapa memakainya di kode | Artefak |
|---|---|---|
| Masuk / keluar | Semua pengguna dashboard | `auth.go`, `/login` |
| Kelola pesantren | Superadmin di UI platform; API buat tenant saat ini terbuka untuk sesi mana pun | `tenants.go`, `/admin` |
| Kelola isi pesantren | Anggota sesuai izin | `/ponpes/*` |
| Jawab calon santri | Hermes, di luar sesi web | event + bridge |
| Halaman publik profil | Tanpa login | `GET /api/v1/public/p/:slug`, `frontend/src/app/p/` |

## 2.3 Karakteristik pengguna

| Pengguna | Ciri yang terbukti | Bukan akun web |
|---|---|---|
| Superadmin platform | `is_platform_superadmin`; setelah login diarahkan ke `/admin/dashboard` | — |
| Admin / staf / viewer pesantren | Baris `tenant_users`; diarahkan ke `/ponpes/dashboard` jika punya membership | — |
| Pengunjung publik | Tidak punya sesi | Halaman `/`, `/p/{slug}`, halaman pemasaran |
| Calon santri / wali | Tidak wajib akun web (PRD journey C; kanal di kode adalah Telegram dan WhatsApp) | [INFERRED] bahwa mereka tidak login ke dashboard — tidak ada alur registrasi publik di `routes.go` |
| Hermes / worker | Pemanggil mesin, kunci event, bukan peran `tenant_users` | — |

Detail hak ada di [03](03-stakeholders-and-roles.md).

## 2.4 Lingkungan operasi

| Komponen | Bukti |
|---|---|
| API | Go 1.25, Fiber v2, `HTTP_ADDR` default `:8080` (`backend/go.mod`, `backend/.env.example`) |
| Database | PostgreSQL via `DATABASE_URL`, migrasi goose (`00001_init.sql` memakai `+goose`) |
| Frontend | Next.js 16.4.3, React 19, dev port 3000 (`frontend/package.json`, `README.md`) |
| Berkas | `DATA_ROOT`, default `./data`; path fisik `DataPonpes/{uuid}` (`storage.go`) |
| Hermes dev | mode `docker`; produksi default `bridge` bila `APP_ENV=production` (`.env.example`) |
| Deploy | `deploy-vps.sh` memakai SSH alias `chatbot-ahlan-server` dan `ahlan-server`, unit `ahlan-api`, `ahlan-worker`, `ahlan-hermes-bridge` (`README.md`) |
| Zona crawl | `Asia/Jakarta` / WIB di `crawler.go` (`time.FixedZone("WIB", 7*3600)`) |

## 2.5 Asumsi dan ketergantungan

| ID | Pernyataan | Status |
|---|---|---|
| ASM-01 | Satu tenant punya paling banyak satu baris `hermes_profiles` (`tenant_id` UNIQUE di `00001_init.sql`). | [CONFIRMED] |
| ASM-02 | Binary PDF/gambar tidak disimpan di PostgreSQL; DB menyimpan metadata (`README.md`, tabel `files`). | [CONFIRMED] |
| ASM-03 | Ketersediaan jawaban chat bergantung pada proses Hermes, bukan hanya pada API dashboard (`README.md` prinsip control plane). | [CONFIRMED] sebagai arsitektur; ambang SLA tidak ada di kode |
| ASM-04 | Nilai `.env.example` hanya untuk lokal dan tidak dipakai di produksi (`README.md`). | [CONFIRMED] sebagai instruksi |
| DEP-01 | Sistem bergantung pada Hermes (profile, kanal, pengetahuan) lewat bridge atau Docker CLI. | [CONFIRMED] |
| DEP-02 | WhatsApp pairing bergantung pada skrip di dalam lingkungan Hermes (`whatsapp.go`, `whatsappBridgeDir`). | [CONFIRMED] |

Anggaran, RPO/RTO komersial, dan target pengguna bersamaan tidak ada di kode. Itu dicatat sebagai [TBD] di [10](10-constraints.md) dan [12](12-open-questions-and-risks.md), bukan sebagai angka.
