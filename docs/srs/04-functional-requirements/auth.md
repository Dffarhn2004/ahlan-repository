# FR — Autentikasi

Sumber bersama: `backend/internal/httpapi/auth.go`, `routes.go`, `server.go`, `frontend/src/app/login/page.tsx`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-AUTH-001 | Sistem harus membuat sesi 24 jam dan cookie `session` (HTTPOnly, SameSite Lax, Path `/`, MaxAge 86400) hanya jika email ada dan bcrypt password cocok. | Must | [CONFIRMED] | `login` di `auth.go` | Login benar mengembalikan 200 dan cookie. Password salah mengembalikan 401 `INVALID_CREDENTIALS` tanpa membedakan email tidak ada vs password salah (satu cabang `err != nil \|\| bcrypt gagal`). |
| FR-AUTH-002 | Sistem harus menghapus sesi server dan cookie pada `POST /api/v1/auth/logout`, lalu menjawab 204. | Must | [CONFIRMED] | `logout` | Setelah logout, `GET /api/v1/me` menjawab 401. |
| FR-AUTH-003 | `GET /api/v1/me` harus menolak permintaan tanpa sesi valid (401 `UNAUTHENTICATED`) dan, jika valid, mengembalikan id, email, role, permissions, dan memberships. | Must | [CONFIRMED] | `authenticate`, `me` | Tanpa cookie: 401. Superadmin: `role=SUPERADMIN`. Anggota non-superadmin: role membership pertama. |
| FR-AUTH-004 | Route di bawah grup `/api/v1` yang memakai `authenticate` harus menolak sesi hilang atau tidak dikenal. | Must | [CONFIRMED] | `routes.go`, `authenticate` | Memanggil `GET /api/v1/tenants` tanpa cookie menghasilkan 401. |
| FR-AUTH-005 | `GET /healthz` harus menjawab JSON `status=ok` tanpa sesi. | Must | [CONFIRMED] | `health`, `routes.go` | Permintaan tanpa cookie mendapat 200 dan `{"status":"ok"}`. |
| FR-AUTH-006 | UI login harus mengarahkan superadmin ke `/admin/dashboard` dan pengguna dengan membership ke `/ponpes/dashboard`. | Must | [CONFIRMED] | `destination` di `login/page.tsx` | Akun seed superadmin masuk ke dashboard platform. Akun anggota pesantren masuk ke `/ponpes/dashboard`. |

### Alur

**Normal.** Pengguna mengirim email dan password ke `POST /api/v1/auth/login`. Server menyimpan token acak (`uuid` digabung dua kali) dengan kedaluwarsa `now+24h`, menyetel cookie, UI memanggil `/me`, lalu `router.replace(destination)`.

**Alternatif.** Sesi masih ada: halaman login memanggil `/me` dan langsung mengarahkan (`useEffect` di `login/page.tsx`).

**Kesalahan.** Body tidak ter-parse: 400 `INVALID_REQUEST`. Kredensial gagal: 401. UI menampilkan satu kalimat: "Email atau kata sandi tidak dapat diverifikasi." (`login/page.tsx`).

MFA tidak bagian dari alur ini. OQ-02 ditangguhkan; lihat [12](../12-open-questions-and-risks.md).
