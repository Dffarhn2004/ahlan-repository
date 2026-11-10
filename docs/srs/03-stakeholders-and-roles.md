# 03 — Pemangku kepentingan, aktor, dan hak akses

## 3.1 Pemangku kepentingan

| Pemangku | Kepentingan yang tertulis | Sumber |
|---|---|---|
| Perusahaan platform (AHLAN) | Meng-onboard banyak pesantren dan memantau asisten | `README.md`, persona superadmin di `00_Produk/03_persona_role_userjourney.md` |
| Pengurus pesantren | Mengisi dokumen, kanal, dan menindaklanjuti pertanyaan | Persona admin di dokumen yang sama; menu `/ponpes/*` |
| Calon santri / wali | Bertanya tanpa akun web | Journey C di PRD; kanal Telegram/WhatsApp di kode |
| Operator infrastruktur | Health, backup, upgrade Hermes | Persona di PRD. Tidak ada peran aplikasi `OPERATOR` di `authz/permissions.go` |

## 3.2 Aktor sistem

| Aktor | Cara dikenali di kode |
|---|---|
| Superadmin | `users.is_platform_superadmin`. `/api/v1/me` mengembalikan `role: SUPERADMIN` dan izin peta `Superadmin`. |
| OWNER | Nilai `tenant_users.role` yang diizinkan constraint SQL dan peta izin. API buat/ubah anggota **tidak** menerima nilai ini (`validMemberRole`). |
| ADMIN | Peran tenant. Izin hampir sama dengan OWNER, tanpa `tenant:lifecycle` (suspend/resume). |
| STAFF | Baca tenant, berkas, bot; buat/ubah berkas; kelola percakapan. Tidak menghapus berkas, tidak mengubah bot, anggota, atau profil. |
| VIEWER | Hanya `tenant:read`, `file:read`, `bot:read`. |
| Pengunjung publik | Tanpa cookie sesi. Route di atas `authenticate` di `routes.go`. |
| Calon santri | Pengirim pesan Telegram/WhatsApp. Bukan baris `users`. |
| Hermes | Pemanggil `POST /api/v1/internal/hermes/events` dan `/handoff` dengan header kunci, bukan sesi. |
| Worker | Proses `backend/cmd/worker` untuk job termasuk crawl. |

Login mengarahkan superadmin ke `/admin/dashboard`. Pengguna lain yang punya membership diarahkan ke `/ponpes/dashboard`. Jika bukan superadmin dan tidak punya membership, UI tetap mengarah ke `/admin/dashboard` (`frontend/src/app/login/page.tsx`, fungsi `destination`). Perilaku terakhir adalah [CONFIRMED] di UI; apakah itu disengaja adalah [TBD] (pertanyaan OQ-04).

## 3.3 Peta izin (sumber perilaku API)

Salinan dari `backend/internal/authz/permissions.go`. Superadmin pada route `/tenants/:id` **tidak** dicek terhadap peta ini: `checkTenant` mengembalikan sukses bila `user.Superadmin` (`server.go`). Peta superadmin tetap dipakai frontend lewat `permissions` dari `/me`.

| Izin | SUPERADMIN (peta) | OWNER | ADMIN | STAFF | VIEWER |
|---|---|---|---|---|---|
| tenant:create | Ya | Tidak | Tidak | Tidak | Tidak |
| tenant:read | Ya | Ya | Ya | Ya | Ya |
| tenant:update | Ya | Ya | Ya | Tidak | Tidak |
| tenant:lifecycle | Ya | Ya | Tidak | Tidak | Tidak |
| member:manage | Ya | Ya | Ya | Tidak | Tidak |
| file:read | Ya | Ya | Ya | Ya | Ya |
| file:create | Ya | Ya | Ya | Ya | Tidak |
| file:update | Ya | Ya | Ya | Ya | Tidak |
| file:delete | Ya | Ya | Ya | Tidak | Tidak |
| bot:read | Ya | Ya | Ya | Ya | Ya |
| bot:update | Ya | Ya | Ya | Tidak | Tidak |
| assistant:manage | Ya | Ya | Ya | Tidak | Tidak |
| conversation:manage | Ya | Ya | Ya | Ya | Tidak |
| audit:read | Ya | Ya | Ya | Tidak | Tidak |
| platform:manage | Ya | Tidak | Tidak | Tidak | Tidak |

`platform:manage` tidak muncul sebagai string izin di `routes.go`. Route platform yang ketat memakai `requireSuperadmin`, bukan izin ini. [CONFIRMED]

## 3.4 Matriks kemampuan (apa yang benar-benar bisa dilakukan)

Legenda: **Ya** = route mengizinkan. **Tidak** = 403 dari izin atau bukan anggota. **API saja** = endpoint ada, menu UI tidak memfilter peran. **Khusus flag** = hanya `is_platform_superadmin`.

| Kemampuan | Superadmin | OWNER | ADMIN | STAFF | VIEWER | Publik / wali |
|---|---|---|---|---|---|---|
| Masuk dashboard | Ya | Ya | Ya | Ya | Ya | Tidak |
| Melihat semua tenant | Ya (`ListTenants`) | Hanya tenant keanggotaan | sama | sama | sama | Tidak |
| Membuat tenant lewat API | Ya, tetapi **sesi mana pun** juga bisa (celah, [12](12-open-questions-and-risks.md)) | Saat ini API tidak menolak | sama | sama | sama | Tidak |
| Mengubah nama/slug tenant | Ya (bypass) | Ya | Ya | Tidak | Tidak | Tidak |
| Suspend / resume | Ya | Ya (`tenant:lifecycle`) | Tidak | Tidak | Tidak | Tidak |
| Mengelola anggota | Ya | Ya | Ya | Tidak | Tidak | Tidak |
| Menetapkan peran OWNER lewat API | Tidak (`validMemberRole`) | Tidak | Tidak | Tidak | Tidak | Tidak |
| Profil, logo, tautan, kontak, pertanyaan | Ya | Ya (butuh `tenant:update` untuk tulis) | Ya | Baca jika `tenant:read`; tulis ditolak | Baca saja | Tidak |
| Unggah / ubah berkas | Ya | Ya | Ya | Ya | Tidak | Tidak |
| Hapus berkas | Ya | Ya | Ya | Tidak | Tidak | Tidak |
| Unduh berkas | Ya | Ya | Ya | Ya | Ya | Tidak |
| Bot Telegram / WhatsApp (tulis) | Ya | Ya | Ya | Tidak | Tidak | Tidak |
| Provision / aksi asisten | Ya | Ya | Ya | Tidak | Tidak | Tidak |
| Sinkron pengetahuan | Ya | Ya (`file:update`) | Ya | Ya | Tidak | Tidak |
| Paksa sumber crawl mati | Ya (`requireSuperadmin`) | Tidak | Tidak | Tidak | Tidak | Tidak |
| Pengaturan Hermes global dan sync-all | Ya | Tidak | Tidak | Tidak | Tidak | Tidak |
| Overview/skills/connection Hermes | Ya, dan **sesi mana pun** (celah) | API saat ini tidak menolak | sama | sama | sama | Tidak |
| Audit semua tenant | Ya, `tenant_id` opsional | Hanya jika `audit:read` dan `tenant_id` diisi | sama | Tidak | Tidak | Tidak |
| Percakapan tenant | Ya | Ya jika `conversation:manage` untuk ubah status | sama | Ya untuk kelola percakapan | Baca lewat membership pada sebagian route Telegram (`member`) | Tidak di dashboard |
| Melihat token bot | Tidak. `getBot` mengosongkan `CredentialRef` | sama | sama | sama | sama | Tidak |
| Halaman `/p/{slug}` | Ya, tanpa login | Ya | Ya | Ya | Ya | Ya |
| Berbicara ke bot | Tidak diatur sebagai peran web | Tidak | Tidak | Tidak | Tidak | Ya, jika kanal PUBLIC atau nomor/ID ada di daftar uji |

Menu UI platform (`AppShell` mode `platform`): Overview, Pesantren, Profil Asisten, Audit log, Pengaturan.  
Menu UI pesantren: Dashboard, Asisten AI, Admin WhatsApp, Dokumen, Sumber Web, Percakapan, Pertanyaan, Pengguna, Aktivitas sistem.  
Sumber: `frontend/src/components/app-shell.tsx`. Menu pesantren **tidak** disembunyikan menurut `can(permission)`.

## 3.5 Perbedaan PRD dan kode

| Topik | PRD (`03_persona_role_userjourney.md`) | Kode |
|---|---|---|
| Staff tidak mengubah bot | Ya | Sesuai: STAFF tidak punya `bot:update` |
| Superadmin melihat percakapan dengan approval | Terbatas, purpose/approval | Tidak ada kolom approval di route percakapan. Superadmin lolos `checkTenant` |
| MFA saat login | Journey A langkah 1 | Kolom `mfa_enabled` ada; `login` tidak membacanya |
| Ponpes Staff sebagai peran P1 | Ya | Peran `STAFF` sudah ada di kode, bukan hanya backlog |

Diagram use case per peran ada di [05-use-cases.md](05-use-cases.md).
