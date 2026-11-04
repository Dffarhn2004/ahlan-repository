# 04 — UX halaman web dan design requirements

## Navigasi superadmin

- **Overview:** total tenants (active/pending/suspended), bot health, traffic sample, error rate, penyelesaian provisioning, usage/cost estimasi.
- **Ponpes:** search/filter/CRUD, detail tenant (termasuk link Google Maps), akses status, paket/kuota, alasan suspend, status sumber crawl dan tombol paksa-nonaktif (P1).
- **Provisioning:** job timeline per tenant, step yang gagal, retry (manual dengan permission), log redacted.
- **AI Profiles:** profile name, versi config, gateway status, health, role tools, uji koneksi.
- **Telegram Bots:** bot username/id, connected_at, status, token last4 masked, rotate/revoke.
- **Monitoring:** sesi dan event per-tenant, latency, error, usage; filter tenant/range/type.
- **Audit:** perubahan admin, actor, perubahan metadata, ekspor sesuai permission.
- **Settings:** model/provider policy, biaya per model, rate limit, runtime baseline, maintenance.

## Navigasi admin ponpes

- **Dashboard:** nama ponpes, bot status, jumlah pertanyaan, notifikasi error, dokumen baru.
- **AI & Bot:** preview identitas asisten (jangan free-edit keseluruhan Hermes config); connect bot, test chat, jam informasi operasional.
- **File Manager:** breadcrumb; folder grid/list; upload, rename, move, delete, preview, status ekstraksi.
- **Informasi Ponpes:** profil, tahun ajaran, pengumuman, link resmi, kontak admin, **link Google Maps** (dengan tombol "Buka di Maps" untuk uji tautan, pesan error jelas bila bukan URL Google Maps), website resmi.
- **Sumber Web (P1):** daftar sumber crawl, jadwal (Manual / Setiap N jam / Harian), tombol **Crawl sekarang**, riwayat run, halaman terkumpul + exclude. Detail di `01_Arsitektur/25_crawl_sumber_web.md`.
- **Percakapan:** list + detail per user dengan data minim dan kontrol akses.
- **Calon Pendaftar (P1):** leads, status, handoff, catatan internal.
- **Tim & Pengaturan (P1):** undangan pengguna, role, kuota, retensi.

## File Manager — interaksi spesifik

| Interaksi | Aturan |
|---|---|
| Buat folder | Nama 1–100 karakter, aman, tidak boleh reserved/path separator |
| Upload file | Pilih folder, drag/drop, progress, maksimum ukuran konfigurabel |
| Preview | Browser preview aman untuk PDF/JPG/PNG; `Content-Disposition: inline` terbatas |
| Rename/move | Tidak bisa keluar dari root tenant; update metadata dan index |
| Delete | Soft-delete ke trash khusus tenant (mis. 7 hari MVP) lalu purge worker |
| Daftar file | Pagination server-side, sort filename/size/date/status |
| Status file | `UPLOADING`, `PROCESSING`, `READY`, `FAILED`, `DELETED` |
| Konten file | Tidak boleh diinterpretasikan sebagai perintah admin atau instruksi Hermes |

## State halaman

Setiap UI punya loading/empty/error/retry state, disabled state selama provisioning, serta success/error yang tidak membocorkan rahasia. Status aktif ditampilkan hanya ketika cek gateway sukses. Konfirmasi wajib untuk rotate bot, suspend, delete, bulk action.

## Aksesibilitas dan kegunaan

Semua controls dapat dinavigasi keyboard, tooltip tidak menjadi satu-satunya sumber informasi, focus state jelas, teks Indonesia, tanggal `Asia/Jakarta` pada UI (tetapi tersimpan UTC), ukuran file dalam MB, copyable request_id untuk support, mobile-friendly minimum 360px.

## Wireframe tekstual

```text
[Logo] [Ponpes Al-Falah ▾]                       [Profil]
Sidebar: Dashboard | Dokumen | AI & Bot | Percakapan

DOKUMEN
DataPonpes / PPDB                            [+ Folder] [Upload]
---------------------------------------------------------------
[folder] Biaya   [folder] Syarat   [folder] Brosur

Name                  Type   Size      Index      Updated
brosur-2027.pdf       PDF    2.4 MB    READY      24 Oct 2026
ketentuan-asrama.jpg  JPG    840 KB    PROCESSING 24 Oct 2026
```

## Acceptance UX

User tidak perlu memahami istilah CLI, profile, `.env`, file path filesystem asli, atau token mentah setelah disimpan. Hanya superadmin yang melihat provisioning detail internal teredaksi.
