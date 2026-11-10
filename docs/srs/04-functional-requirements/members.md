# FR — Anggota tenant

Sumber: `backend/internal/httpapi/members.go`, `00004_authz.sql`, `permissions.go`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-MEM-001 | Pemegang `member:manage` harus dapat melihat anggota lewat `GET /api/v1/tenants/:id/members`. | Must | [CONFIRMED] | route + `listMembers` | STAFF mendapat 403. ADMIN di tenant yang sama mendapat 200 dan daftar. |
| FR-MEM-002 | Pembuatan anggota harus menolak email tanpa `@` atau password lebih pendek dari 10 karakter dengan 400 `INVALID_MEMBER`. | Must | [CONFIRMED] | `createMember` | Password 9 karakter: 400. Password 10 dan email ber-`@`: 201 dengan `user_id` dan `role`. |
| FR-MEM-003 | Peran yang boleh dikirim saat buat atau ubah anggota harus hanya `ADMIN`, `STAFF`, atau `VIEWER`. Jika role kosong saat buat, sistem harus memakai `ADMIN`. | Must | [CONFIRMED] | `validMemberRole`, default `in.Role = authz.Admin` | Body `role=OWNER` atau `SUPERADMIN`: 400 `INVALID_MEMBER`. Body tanpa role: tersimpan `ADMIN`. |
| FR-MEM-004 | `PATCH /api/v1/tenants/:id/members/:user_id` harus mengubah peran anggota yang sudah ada ke salah satu dari tiga peran di FR-MEM-003 dan menjawab `{status: UPDATED, role}`. | Must | [CONFIRMED] | `updateMemberRole` | Ubah STAFF menjadi VIEWER terlihat di GET members. Peran di luar daftar: 400. |

### Alur

**Normal.** ADMIN/OWNER/superadmin membuka dialog pengguna (UI `/admin/tenants/[id]` dan `/ponpes/users`). Email dinormalisasi `lower` di `CreateMember`. Password disimpan sebagai hash bcrypt, bukan plaintext (`members.go` memakai `bcrypt`).

**Kesalahan.** Bukan anggota tenant: 403 `FORBIDDEN`. Bukan UUID: 400.

**Peran OWNER.** Constraint SQL mengizinkan `OWNER` (`00004_authz.sql`), peta izin mendefinisikannya, API tidak bisa menetapkan lewat endpoint ini. Baris OWNER hanya mungkin lewat tulis database di luar API. [CONFIRMED]
