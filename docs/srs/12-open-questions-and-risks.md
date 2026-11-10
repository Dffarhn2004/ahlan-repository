# 12 — Pertanyaan terbuka, risiko, dan celah

## 12.1 Pertanyaan yang masih terbuka

| ID | Prioritas | Pertanyaan | Bukti yang bertentangan atau kosong |
|---|---|---|---|
| OQ-04 | Sedang | Ke mana pengguna tanpa membership dan bukan superadmin harus diarahkan? | `destination()` di login mengarah ke `/admin/dashboard`. |
| OQ-05 | Tinggi | Apakah superadmin boleh membaca isi percakapan tanpa approval? | PRD: terbatas + approval. Kode: `checkTenant` lolos untuk superadmin. |
| OQ-06 | Sedang | Apakah BR-010 (AI bukan penentu penerimaan) tetap berlaku untuk rilis ini? | Tidak ada modul admission di route. PRD FR-017 lead masih P1. |
| OQ-08 | Sedang | Apakah provision Hermes idempoten pada retry, sesuai PRD FR-002? | UNIQUE profile mencegah dua baris. Jalur gagal di tengah Hermes tidak disimpulkan lulus di SRS. |
| OQ-09 | Sedang | Interval crawl mana yang berlaku: 24 jam pada learn-link, atau minimal 6 jam di PRD FR-024? | `learnIntervalMinutes = 1440`. PRD menulis ≥ 6 jam. |
| OQ-10 | Tinggi | Dasar hukum, retensi, dan teks kebijakan privasi yang mengikat operator? | Halaman `/kebijakan-privasi` ada. Tidak ada tinjauan hukum di repo. |
| OQ-11 | Sedang | Apakah setiap write bisnis menulis audit sebelum 2xx? | `TENANT_CREATED` ya, satu transaksi. Label UI memuat banyak aksi lain. Audit menyeluruh belum dicek handler per handler. |
| OQ-12 | Rendah | Apakah nama berkas `.env` / `SOUL.md` harus ditolak di file manager? | `safeName` menolak path, bukan daftar nama terlarang. PRD melarang sunting config bebas dari file manager. |
| OQ-13 | Sedang | Berapa lama percakapan, audit, dan berkas disimpan? | Tidak ada TTL selain sesi 24 jam, pertanyaan OPEN 200, lampiran cache 60 menit. |
| OQ-14 | Sedang | Apakah unggah ditolak ketika jumlah byte tenant melebihi `storage_quota_bytes`? | Default kuota ada. SRS tidak menemukan penjumlahan kuota di `uploadFile` pada pembacaan ini. |
## 12.2 Ditangguhkan (bukan pertanyaan aktif)

Pemilik produk pada 10 November 2026 meminta butir ini diabaikan. Tidak ada perubahan requirement. Perilaku kode di kolom "Yang berlaku" tetap [CONFIRMED].

| ID | Yang berlaku sampai ada keputusan lain |
|---|---|
| OQ-01 | `POST /api/v1/tenants` cukup dengan sesi. Bukan 403 khusus non-superadmin. |
| OQ-02 | Login tidak memeriksa `mfa_enabled`. |
| OQ-03 | `GET /hermes/overview`, `/skills`, dan `/connection` cukup dengan sesi. |
| OQ-07 | API anggota hanya menerima `ADMIN`, `STAFF`, `VIEWER`. `OWNER` tetap ada di skema dan peta izin. |
| OQ-15 | Menu `/ponpes/*` tampil utuh. Izin ditegakkan di API. |

## 12.3 Risiko

| ID | Risiko | Dampak | Status bukti |
|---|---|---|---|
| R-01 | Sesi pesantren mana pun dapat membuat tenant DRAFT. | Tenant liar, slug dipakai, audit mencatat aktor itu. | [CONFIRMED] celah kode |
| R-02 | Sesi mana pun dapat membaca overview Hermes dan status koneksi. | Kebocoran metadata platform ke akun pesantren. | [CONFIRMED] route |
| R-03 | UI tidak menyembunyikan aksi. Pengguna melihat tombol lalu mendapat 403. | Bukan bypass jika API konsisten. Tetap membingungkan dan mudah disalahuji. | [CONFIRMED] shell |
| R-04 | PRD menjanjikan security gate sebelum bot publik. Mode PUBLIC bisa disimpan dari `saveBot` tanpa langkah gate terpisah yang terlihat di fungsi itu. | Bot publik sebelum uji isolasi. | [INFERRED] dari tidak adanya panggilan gate di `saveBot` |
| R-05 | Dokumentasi Hermes berubah cepat. | Perilaku jawaban bot menyimpang dari SRS. | [CONFIRMED] catatan `README.md` |
| R-06 | `mfa_enabled` memberi kesan kontrol yang tidak jalan. | Operator mengira MFA aktif. | [CONFIRMED] |

## 12.4 Kode atau dokumen tanpa requirement produk yang ditutup

Hal berikut ada di repo dan **sengaja tidak** dijadikan FR "sistem harus …" karena tidak lengkap sebagai kewajiban yang disepakati, atau hanya pendukung:

| Artefak | Alasan tidak dijadikan FR |
|---|---|
| Route `t.All("/*", read, notImplemented)` | Menjawab 501. Bukan fitur. |
| Halaman pemasaran `/chatbot-pesantren`, `/chatbot-whatsapp-pesantren`, `/penerimaan-santri-baru`, `/panduan-psb-pesantren` | Ada sebagai route. Salinan pemasaran tidak disalin menjadi requirement fungsional selain FR-PUB-004 untuk beranda. |
| `frontend/src/lib/seo.ts`, `sitemap.ts`, `manifest.ts` | Pendukung situs. Tidak ada sasaran indeks yang diukur. |
| Template `templates/SOUL.template.md`, `KNOWLEDGE_POLICY.template.md` | Bahan prompt. Isi tidak diverifikasi sama dengan persona yang tersimpan di `hermes_global_settings`. |
| `deploy-vps.sh`, `migrate-hermes.sh` | Prosedur operasi. Bukan perilaku pengguna. |
| Paket Excel export log Telegram | Endpoint `telegram/export` disebut di use case anggota. Format kolom Excel tidak dispesifikasikan baris-per-baris. |
| Izin `platform:manage` | Ada di peta, tidak dipakai `Allows` di route yang dibaca. |
| Status siklus `ARCHIVED`, `DELETING`, `DELETED` | Diizinkan CHECK. Tidak ada route khusus selain suspend/resume yang menulis `SUSPENDED` dan `ONBOARDING`. |
| PRD FR-017 lead, FR-018 takeover manusia, FR-019 biaya token, FR-021 billing | Tidak ada tabel lead atau invoice di migrasi `00001`–`00020`. Tetap backlog PRD, bukan perilaku as-is. |
| Seed `cmd/seed` | Membuat superadmin dari `SEED_ADMIN_EMAIL`. Kredensial contoh di `.env.example` dilarang untuk produksi oleh README, bukan oleh requirement runtime. |

## 12.5 Item PRD P0 yang tidak dinyatakan lulus

Dari katalog `00_Produk/02_ruang_lingkup_requirements.md`: FR-002 idempoten (OQ-08), FR-008 eval abstain (hanya marker yang [CONFIRMED]), FR-010 riwayat job provisioning di UI (route ada, kelengkapan tampilan tidak diaudit layar-per-layar), FR-014 slash command administratif (disebut di komentar `hermes.go` "limits slash commands"; uji ancaman tidak dilampirkan), FR-016 karantina unggahan berbahaya (ada batas ukuran dan nama aman; MIME spoof dan karantina tidak dijadikan lulus). NFR-02, NFR-07, NFR-10 (uji regresi dua tenant wajib tiap rilis) tidak punya artefak hasil di SRS ini.
