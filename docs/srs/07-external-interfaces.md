# 07 — Antarmuka eksternal

## 7.1 Antarmuka pengguna

| ID | Antarmuka | Autentikasi | Sumber |
|---|---|---|---|
| IF-001 | `/login` | Publik; sesi aktif diarahkan | `frontend/src/app/login/page.tsx` |
| IF-002 | `/admin/dashboard`, `/admin`, `/admin/tenants/[id]`, `/admin/hermes`, `/admin/audit` | Sesi; UI platform | `app-shell.tsx` daftar `platform` |
| IF-003 | `/ponpes/dashboard`, `profile`, `whatsapp-admin`, `files`, `sources`, `conversations`, `questions`, `users` | Sesi anggota | daftar `ponpes` |
| IF-004 | `/audit`, `/settings`, `/files` | Sesi | halaman yang sama |
| IF-005 | `/`, `/p/[slug]`, `/p/[slug]/[kind]`, halaman pemasaran, `/kebijakan-privasi` | Tanpa sesi | `frontend/src/app/**` |
| IF-006 | Nama aplikasi dan logo shell | `localStorage` `ahlan.branding`, default AHLAN | `branding.ts` |

`NEXT_PUBLIC_API_URL` default `http://localhost:8080` (`login/page.tsx`). Cookie dikirim `credentials: "include"`.

## 7.2 API yang dipakai aplikasi

Basis: `/api/v1`. Rincian kewajiban ada di bab 04. Ringkasan kelompok:

| ID | Kelompok | Auth |
|---|---|---|
| IF-007 | `POST /auth/login`, `POST /auth/logout`, `GET /healthz` | Login dan health publik. Logout tanpa wajib sesi (menghapus cookie yang ada). |
| IF-008 | `GET /me`, `GET/POST /tenants`, `GET /audit-logs`, `GET/PATCH /conversations` | Sesi |
| IF-009 | `/tenants/:id/**` berkas, profil, bot, whatsapp, crawl, hermes, anggota | Sesi + keanggotaan atau superadmin, plus izin pada tabel bab 03 |
| IF-010 | `/hermes/settings`, `/hermes/sync-all`, `PATCH /crawl/sources/:sid/force` | Superadmin |
| IF-011 | `GET /hermes/overview`, `/skills`, `/connection` | Sesi saja. Pembatasan superadmin ditangguhkan (OQ-03). |
| IF-012 | `GET /public/p...` | Tanpa sesi |
| IF-013 | `POST /internal/hermes/events`, `POST /internal/hermes/handoff` | Bukan sesi. Header `X-Hermes-Event-Key` (`README.md`, `EventsApp`) |

Route sisa di bawah `/tenants/:id/*` yang tidak terdaftar menjawab 501 `DOMAIN_ROUTE_NOT_IMPLEMENTED` (`notImplemented`).

## 7.3 Integrasi pihak ketiga dan proses lokal

| ID | Sistem | Arah | Bukti |
|---|---|---|---|
| IF-014 | Hermes profile | API menulis profile, kanal, pengetahuan lewat Docker CLI (dev) atau Unix socket bridge (produksi) | `.env.example`, `server.go` `NewControl` |
| IF-015 | Telegram | Token disimpan sebagai `credential_ref`; tidak dikembalikan GET | `profile_bot.go`, tabel `telegram_bots` |
| IF-016 | WhatsApp | Pairing perangkat lewat bridge di `/opt/hermes/scripts/whatsapp-bridge` | `hermes/whatsapp.go` |
| IF-017 | Situs publik pesantren | HTTP GET oleh worker crawler, User-Agent `PonpesAICrawler/1.0` | `server.go` `FetcherConfig` |
| IF-018 | PostgreSQL | `DATABASE_URL` | `.env.example` |

Tidak ada klien pembayaran, email transaksional, atau object storage S3 di `go.mod`. PRD menyatakan S3/MinIO tidak diperlukan untuk MVP (`README.md` keputusan arsitektur). [CONFIRMED] sebagai ketiadaan dependensi, bukan sebagai larangan selamanya.

## 7.4 Perangkat keras dan protokol

| ID | Protokol | Catatan |
|---|---|---|
| IF-019 | HTTPS untuk tautan resmi | `links.Validate` menolak skema yang tidak diizinkan (lihat paket `links`) |
| IF-020 | HTTP API lokal `:8080` | Pengembangan |
| IF-021 | Unix socket bridge | `/run/ahlan-hermes/bridge.sock` contoh di `.env.example` |
| IF-022 | Listener event terpisah | `HERMES_EVENTS_LISTEN` di `.env.example` bila kontainer tidak menjangkau `127.0.0.1` |

Tidak ada integrasi perangkat absensi, gerbang, atau SMS di kode yang ditemukan.
