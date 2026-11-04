# 03 — Persona, peran, dan user journey

## Persona

1. **Company Superadmin** — staf perusahaan yang membuat/menonaktifkan tenant, memantau profil, menyelesaikan provisioning, melihat agregat biaya dan error.
2. **Ponpes Admin** — staf ponpes yang memperbarui dokumen, identitas informasi publik, menjawab pertanyaan yang dieskalasikan, serta memeriksa log miliknya.
3. **Ponpes Staff (P1)** — staf penerimaan yang aksesnya terbatas pada conversation/lead, tidak bisa mengubah bot/token.
4. **Calon santri/wali** — pengguna Telegram yang bertanya sebelum mendaftar; tidak perlu membuat akun web untuk FAQ.
5. **System Operator** — pengelola infrastruktur terbatas yang meninjau health, backup, upgrade Hermes, bukan otomatis mendapat akses isi pesan pribadi.

## Hak akses (ringkas)

| Objek/aksi | Superadmin | Admin Ponpes | Staff Ponpes | Publik Telegram |
|---|---|---|---|---|
| Create/suspend tenant | Ya | Tidak | Tidak | Tidak |
| Lihat metadata semua tenant | Ya | Tidak | Tidak | Tidak |
| Lihat folder/file tenant | By privileged audited access | Tenant sendiri | Jika diberi permission | Tidak; hanya AI retrieval yang diizinkan |
| Upload/edit file | By audited support mode | Tenant sendiri | Opsional | Tidak |
| Hubungkan/ganti bot | Ya / assisted | Tenant sendiri jika diizinkan | Tidak | Tidak |
| Lihat percakapan | Terbatas, purpose/approval + audit | Tenant sendiri | Tenant sendiri sesuai tugas | Miliknya melalui Telegram |
| Melihat token plaintext | Tidak | Tidak | Tidak | Tidak |
| Memakai tool pendaftaran | Tidak via bot | Ya via dashboard | Ya sesuai permission | Hanya melalui API AI yang tervalidasi |

## Journey A — Superadmin onboarding

1. Login + MFA; buka daftar ponpes.
2. Input nama, slug display, kontak admin, kuota, model policy.
3. Submit, lihat status `PENDING/PROVISIONING` dan progress job.
4. Jika worker selesai, `AWAITING_TELEGRAM`.
5. Koneksikan bot + lakukan smoke test/verify dokumen/security gate.
6. Tekan aktifkan ketika seluruh gate terpenuhi. Tenant `ACTIVE`.
7. Monitoring seluruh tenant; bisa suspend dengan alasan dan audit.

## Journey B — Admin mengelola dokumen

1. Login, masuk menu Dokumen.
2. Melihat folder `PPDB`, `Profil`, `Akademik` di root tenant (tidak melihat nama root fisik tenant lain).
3. Buat folder atau unggah PDF/JPG/PNG/TXT dengan progress.
4. Sistem menandai `PROCESSING` kemudian `READY` atau `FAILED`.
5. Admin menguji pertanyaan di bot; respon mengutip dokumen terbaru.
6. Revisi file memicu versi baru/refresh knowledge, dengan audit who/when.

## Journey C — Wali/calon santri

1. Buka `t.me/<botname>` dari brosur atau website ponpes.
2. `/start`, melihat disclaimer bot AI dan tautan kebijakan privasi.
3. Menanyakan jadwal, syarat, fasilitas; mendapat jawaban grounded.
4. Jika dokumen tidak mencakup pertanyaan, bot tidak mengarang dan menawarkan admin resmi.
5. Opsional fase P1: persetujuan, pengisian lead, cek status pendaftaran.

## Journey D — Gangguan

- Bot credential invalid → status `BOT_ERROR`, superadmin melihat alasan aman, admin dapat mengganti token.
- Ekstraksi PDF gagal → file tetap tersimpan; status `FAILED`, QA tidak menganggap dokumen sudah siap.
- Webhook telemetry Go down → Hermes masih menjawab pertanyaan berbasis data cache/berkas yang tersedia; log best-effort bisa hilang dan dashboard memberi gap indikasi, bukan klaim lengkap.
- Tenant di-suspend → tool mutasi ditolak, profil di-park, akses dashboard data lama tetap sesuai policy.
