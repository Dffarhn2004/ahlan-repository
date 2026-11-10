# FR — Asisten Hermes

Sumber: `routes.go`, `hermes_admin.go`, `00001_init.sql` (`hermes_profiles`), `README.md`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-AST-001 | `POST /api/v1/tenants/:id/hermes/provision` harus hanya dapat dipanggil pemegang `assistant:manage` atau superadmin. | Must | [CONFIRMED] | route `allow("assistant:manage")` | STAFF: 403. ADMIN: bukan 403 dari izin. |
| FR-AST-002 | `POST /api/v1/tenants/:id/hermes/action` harus memakai izin yang sama dengan provision. | Must | [CONFIRMED] | route yang sama | VIEWER: 403. |
| FR-AST-003 | `POST /api/v1/tenants/:id/hermes/sync` harus memerlukan `file:update` (bukan `assistant:manage`). | Must | [CONFIRMED] | route `allow("file:update")` | STAFF dapat memanggil. VIEWER tidak. |
| FR-AST-004 | `GET /api/v1/tenants/:id/knowledge/status` harus memerlukan `tenant:read`. | Must | [CONFIRMED] | route `read` | Bukan anggota: 403. |
| FR-AST-005 | `GET/PATCH /api/v1/hermes/settings`, `POST /api/v1/hermes/sync-all`, dan `GET /api/v1/hermes/sync-all/:job_id` harus menolak pengguna yang bukan superadmin. | Must | [CONFIRMED] | `requireSuperadmin` | Anggota ADMIN: 403. Superadmin PATCH dengan `description` > 2000, `soul` > 50000, atau skills > 100: 400 `INVALID_HERMES_SETTINGS` (`saveHermesSettings`). |
| FR-AST-006 | Sistem harus menolak profile Hermes kedua untuk tenant yang sama. | Must | [CONFIRMED] | `hermes_profiles.tenant_id` UNIQUE dan `profile_key` UNIQUE | Insert kedua untuk `tenant_id` sama gagal di database. |

### Alur pengaturan global

**Normal.** Superadmin mengubah pengaturan. Server menerapkan ke setiap profile key, menyimpan baris global, menulis audit lewat repository save, lalu menjadwalkan restart asisten yang aktif di latar (timeout konteks 10 menit di goroutine `saveHermesSettings`).

**Kesalahan.** Hermes menolak apply: 502 `HERMES_SETTINGS_APPLY_FAILED` dan simpan database tidak dilanjutkan (return terjadi sebelum `SaveGlobalHermesSettings`).

`GET /hermes/overview`, `/skills`, dan `/connection` hanya memakai `authenticate`. Uji yang berlaku: sesi valid mendapat 200, bukan 403 untuk non-superadmin. OQ-03 ditangguhkan; lihat [12](../12-open-questions-and-risks.md).

Satu tenant satu profile adalah invariant database, bukan janji bahwa provision selalu idempoten di semua jalur error Hermes. Idempotensi retry worker disebut PRD FR-002; bukti lengkap idempotensi tidak diringkas sebagai lulus di SRS ini [TBD] OQ-08.
