# 09 — Aturan bisnis

Aturan ini adalah batas domain yang dipakai lebih dari satu requirement. Jika angka berubah di kode, ubah aturan dan FR yang merujuknya bersama-sama.

| ID | Aturan | Prioritas | Status | Sumber | Cara menguji |
|---|---|---|---|---|---|
| BR-001 | Satu pesantren adalah satu tenant dan paling banyak satu Hermes profile serta paling banyak satu bot Telegram dan satu kanal WhatsApp. | Must | [CONFIRMED] | UNIQUE `hermes_profiles.tenant_id`, `telegram_bots.tenant_id`, `whatsapp_channels.tenant_id` | Insert duplikat ditolak database. |
| BR-002 | Berkas unggahan pengetahuan satu objek paling besar 20 MiB. Logo paling besar 2 MiB dan hanya PNG, JPEG, WEBP. | Must | [CONFIRMED] | `maxUploadBytes`, `maxLogoBytes`, `logoTypes` | Lihat FR-FILE-002 dan FR-FILE-006. |
| BR-003 | Kuota penyimpanan awal tenant 500 MiB (524288000 byte). | Must | [CONFIRMED] | default kolom `storage_quota_bytes` | Tenant baru tanpa override punya nilai itu. Penegakan kuota pada setiap unggah (jumlah kumulatif versus default) tidak diverifikasi di SRS ini [TBD] OQ-14. |
| BR-004 | Mode akses kanal hanya `TESTING` (daftar identitas wajib) atau `PUBLIC` (daftar dikosongkan). | Must | [CONFIRMED] | `saveBot`, `saveWhatsApp` | FR-TG-001, FR-WA-001. |
| BR-005 | Nomor WhatsApp bersama memakai `GENERAL` (tanya keperluan, PSB dijawab asisten, selain itu ke admin). Nomor khusus PSB memakai `PSB_ONLY`. | Must | [CONFIRMED] | komentar `00020_whatsapp_service_mode.sql` | Nilai kolom hanya dua itu. |
| BR-006 | Asisten yang tidak punya jawaban harus memakai kalimat yang memuat `belum tersedia di data kami`. Sistem menghitung pertanyaan dari frasa itu. | Must | [CONFIRMED] | `questions.Marker` | FR-QC-003. |
| BR-007 | Tautan resmi paling banyak 20 per tenant. Kontak publik paling banyak 20. Admin routing WhatsApp paling banyak 10 di UI. | Must | [CONFIRMED] | `MaxLinksPerSite`, `MaxPerSite`, `MAX_ADMINS` | FR-KB-001, FR-QC-004, FR-WA-004. |
| BR-008 | Media sosial, `wa.me`, `forms.gle`, YouTube, dan `linktr.ee` tidak dijadikan bahan bacaan crawler. | Must | [CONFIRMED] | `unlearnable` di `links.go` | `CanLearn` mengembalikan false untuk host itu. |
| BR-009 | Crawl satu situs dibatasi kedalaman 3 dan 200 halaman, default 2 dan 50, serta menolak IP internal. | Must | [CONFIRMED] | `crawler.go`, `guard.go` | FR-KB-002, FR-KB-003. |
| BR-010 | AI bukan pihak yang menetapkan santri diterima atau pembayaran sah. | Must | [TBD] | Out-of-scope PRD. Tidak ada tabel keputusan kelulusan di migrasi. | Tidak ada endpoint penerimaan. Konfirmasi produk: OQ-06. |
| BR-011 | Peran yang dapat diundang dari aplikasi adalah ADMIN, STAFF, dan VIEWER. OWNER hanya ada jika ditulis di luar API. | Must | [CONFIRMED] | `validMemberRole` versus CHECK SQL | FR-MEM-003. |
| BR-012 | Superadmin platform bukan baris peran tenant. Flag `is_platform_superadmin` melewati pemeriksaan izin route `/tenants/:id`. | Must | [CONFIRMED] | `checkTenant` | Anggota VIEWER yang juga superadmin (jika data demikian) lolos route tenant. Kombinasi itu tidak dicegah skema. |
