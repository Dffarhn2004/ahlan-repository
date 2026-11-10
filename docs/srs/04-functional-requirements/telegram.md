# FR — Telegram

Sumber: `httpapi/profile_bot.go`, `00005_mvp.sql`, route di `routes.go`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-TG-001 | Penyimpanan bot harus menolak `bot_id` ≤ 0, username kosong, atau kredensial kosong, dan harus memakai mode `TESTING` bila mode kosong. Mode selain `TESTING` dan `PUBLIC` harus ditolak. | Must | [CONFIRMED] | `saveBot` | Mode `PUBLIC`: `allowed` dikosongkan dan 200 dengan `CredentialRef` kosong di respons. Mode `TESTING` tanpa user id positif: 400 `TELEGRAM_USER_IDS_REQUIRED`. |
| FR-TG-002 | `GET /api/v1/tenants/:id/bot` harus mengosongkan `CredentialRef` sebelum respons. Tidak ada peran, termasuk superadmin, yang menerima token lewat endpoint ini. | Must | [CONFIRMED] | `getBot` baris `bot.CredentialRef = ""` | Respons JSON tidak memuat token yang baru disimpan. |
| FR-TG-003 | Satu tenant harus punya paling banyak satu baris bot, dan `telegram_bot_id` harus unik antar tenant. | Must | [CONFIRMED] | `telegram_bots.tenant_id UNIQUE`, `telegram_bot_id UNIQUE` | Dua tenant dengan bot id sama: insert/upsert kedua gagal. |
| FR-TG-004 | Penghapusan bot harus menjawab 204, menghapus baris, dan mencoba menghapus env kanal `TELEGRAM_` pada profile. Audit `TELEGRAM_BOT_REMOVED` harus tercatat. Kegagalan bersihkan env tidak boleh menggagalkan penghapusan baris (hanya log peringatan). | Must | [CONFIRMED] | `deleteBot` | Setelah DELETE, GET bot mengembalikan 404 `BOT_NOT_FOUND`. Izin: `bot:update`. |

### Mode akses

| Mode | Aturan di `saveBot` |
|---|---|
| TESTING | Wajib minimal satu Telegram user id positif, duplikat dibuang. |
| PUBLIC | Daftar user id dikosongkan. UI memperingatkan bahwa siapa pun yang menemukan bot dapat mengobrol (`assistant-panel.tsx`). |

Ekspor log `GET .../telegram/export` dan daftar grup `GET .../telegram/conversations` memakai middleware anggota (`member`), yaitu keanggotaan tanpa izin khusus. Mengubah bot tetap `bot:update`. Pesan dev `POST .../telegram/dev-message` memakai `bot:update`.

Status bot yang diizinkan skema: `PENDING`, `VERIFIED`, `ACTIVE`, `SUSPENDED`, `INVALID` (`00005_mvp.sql`). `saveBot` menuliskan `PENDING` pada objek respons setelah upsert.
