# FR — Tenant (pesantren)

Sumber bersama: `tenants.go`, `routes.go`, `00001_init.sql`, `repository.go` `CreateTenant`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-TENANT-001 | Sistem harus mengembalikan seluruh tenant kepada superadmin dan hanya tenant keanggotaan kepada pengguna lain pada `GET /api/v1/tenants`. | Must | [CONFIRMED] | `listTenants` | Dua anggota di dua tenant: masing-masing hanya melihat tenantnya. Superadmin melihat keduanya. |
| FR-TENANT-002 | `POST /api/v1/tenants` harus menyimpan tenant `lifecycle_status=DRAFT`, menulis audit `TENANT_CREATED` dalam transaksi yang sama, dan menjawab 202 bersama body tenant. | Must | [CONFIRMED] | `createTenant`, `CreateTenant` | Slug unik baru menghasilkan baris DRAFT dan satu audit. Slug bentrok gagal (constraint UNIQUE) dan tidak meninggalkan audit tanpa tenant. |
| FR-TENANT-003 | Pengubah dengan `tenant:update` (atau superadmin) harus dapat mengubah nama dan slug non-kosong lewat `PATCH /api/v1/tenants/:id`. | Must | [CONFIRMED] | `updateTenant`, route `allow("tenant:update")` | Nama kosong atau slug kosong: 400 `INVALID_REQUEST`. STAFF: 403. |
| FR-TENANT-004 | Pemegang `tenant:lifecycle` harus dapat mengatur status `SUSPENDED` lewat `POST /tenants/:id/suspend` dan `ONBOARDING` lewat `POST /tenants/:id/resume`, jawaban 204. | Must | [CONFIRMED] | `setTenantStatus` | ADMIN (tanpa izin lifecycle) mendapat 403. OWNER mendapat 204 dan status berubah. |
| FR-TENANT-005 | Status siklus hidup yang disimpan harus salah satu dari `DRAFT`, `ONBOARDING`, `ACTIVE`, `SUSPENDED`, `ARCHIVED`, `DELETING`, `DELETED`. | Must | [CONFIRMED] | `00001_init.sql` CHECK | Insert status lain ditolak database. |
| FR-TENANT-006 | Slug tenant harus unik. | Must | [CONFIRMED] | `tenants.slug UNIQUE` | Dua POST dengan slug sama: yang kedua gagal. |
| FR-TENANT-007 | Kuota penyimpanan default tenant harus 524288000 byte (500 MiB) dan tidak boleh negatif. | Must | [CONFIRMED] | `storage_quota_bytes bigint NOT NULL DEFAULT 524288000 CHECK (>= 0)` | Baris baru tanpa kuota eksplisit bernilai 524288000. |
| FR-TENANT-008 | `GET /api/v1/tenants/:id` harus menjawab 400 untuk id bukan UUID dan 404 `TENANT_NOT_FOUND` jika baris tidak ada, setelah lolos `tenant:read`. | Must | [CONFIRMED] | `getTenant`, `tenantIDParam` | UUID acak yang valid formatnya tetapi tidak ada: 404. Anggota tenant lain: 403 sebelum baca. |

### Alur buat tenant

**Normal (as-is).** Sesi valid mengirim `{name, slug}`. Transaksi insert + audit. Respons 202. Provision Hermes **tidak** ikut dalam transaksi ini (`CreateTenant` tidak memanggil `QueueHermesProvisioning`).

**Kesalahan.** Body rusak: 400 `INVALID_REQUEST`.

**Catatan penelusuran.** Peta izin punya `tenant:create` hanya untuk superadmin, tetapi `POST /tenants` tidak memanggil `requireSuperadmin` atau `Allows(..., "tenant:create")`. Perilaku yang berlaku adalah sesi mana pun dapat membuat tenant. OQ-01 ditangguhkan; lihat [12](../12-open-questions-and-risks.md).

PRD FR-001 mensyaratkan tenant tidak langsung aktif. Status awal `DRAFT` memenuhi bagian itu [CONFIRMED]. Bagian "job tersimpan atomik" **tidak** terpenuhi oleh `CreateTenant` [CONFIRMED] sebagai perbedaan; dicatat di [12](../12-open-questions-and-risks.md).
