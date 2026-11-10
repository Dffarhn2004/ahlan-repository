# 01 — Pendahuluan

## 1.1 Tujuan

Dokumen ini menspesifikasikan perilaku perangkat lunak yang **sudah ada di repositori** pada 10 November 2026, ditambah celah terhadap PRD yang belum terbukti di kode.

Pembaca yang dituju: pemilik produk, analis, dan engineer yang perlu tahu kewajiban sistem per peran tanpa membaca seluruh kode.

Dokumen ini bukan pengganti PRD di `00_Produk/`–`04_Referensi/`. PRD adalah niat produk (sebagian P0/P1/P2). SRS ini memisahkan niat itu dari perilaku yang dapat diuji di kode.

## 1.2 Ruang lingkup

**Nama di antarmuka:** AHLAN.  
**Masalah yang diselesaikan (dari PRD dan beranda):** pesantren membutuhkan asisten yang menjawab pertanyaan calon santri/wali dari dokumen resmi, lewat Telegram dan WhatsApp, dengan ruang kerja terpisah per pesantren. Sumber: `README.md` ringkasan, `frontend/src/app/page.tsx` (FAQ dan langkah unggah–hubungkan kanal–melayani).

**In scope SRS (terbukti ada artefaknya):**

- Autentikasi sesi cookie, peran platform dan peran tenant.
- Daftar, buat, ubah, tangguhkan, dan lanjutkan onboarding tenant.
- Anggota tenant dan izin.
- Profil, logo, tautan resmi, kontak, pertanyaan tak terjawab.
- Berkas dan folder di `DataPonpes/{tenant_id}`.
- Asisten Hermes per tenant (provision, aksi, sinkron pengetahuan, pengaturan global).
- Kanal Telegram dan WhatsApp, termasuk mode uji/publik dan mode layanan PSB/umum.
- Crawl sumber web, percakapan, audit.
- Halaman publik `/p/{slug}` dan halaman pemasaran.

**Out of scope yang ditulis PRD dan tidak dijadikan requirement berjalan** (lihat [12](12-open-questions-and-risks.md)):

- Langganan berbayar (PRD FR-021, P2).
- AI sebagai penentu penerimaan atau pembayaran (`00_Produk/02_ruang_lingkup_requirements.md`, out-of-scope P0).
- Admin menulis `.env`, `config.yaml`, atau `SOUL.md` bebas lewat file manager (out-of-scope P0 yang sama).
- Self-serve onboarding penuh (P2 di PRD).

## 1.3 Definisi singkat

Istilah lengkap ada di [13-glossary.md](13-glossary.md). Yang dipakai di seluruh SRS:

| Istilah | Arti di dokumen ini |
|---|---|
| Tenant | Satu pesantren. Tabel `tenants`. |
| Superadmin | Pengguna dengan `users.is_platform_superadmin = true`. Bukan peran di `tenant_users`. |
| Anggota tenant | Baris `tenant_users` dengan peran `OWNER`, `ADMIN`, `STAFF`, atau `VIEWER`. |
| Asisten | Satu Hermes profile per tenant (`hermes_profiles.tenant_id` unik). |
| PSB | Penerimaan santri baru. Mode WhatsApp `PSB_ONLY` dijelaskan di migrasi `00020`. |

## 1.4 Referensi

| ID | Dokumen |
|---|---|
| REF-01 | `README.md` |
| REF-02 | `00_Produk/02_ruang_lingkup_requirements.md` |
| REF-03 | `00_Produk/03_persona_role_userjourney.md` |
| REF-04 | `backend/internal/httpapi/routes.go` |
| REF-05 | `backend/internal/authz/permissions.go` |
| REF-06 | `backend/db/migrations/` (`00001` … `00020`) |
| REF-07 | IEEE Std 830 dan ISO/IEC/IEEE 29148 sebagai kerangka bab, bukan sertifikasi |

## 1.5 Gambaran dokumen

Bab 02–03 menjelaskan produk dan peran. Bab 04 berisi kewajiban fungsional. Bab 05 mengikat peran ke use case. Bab 06–10 mencakup mutu, antarmuka, data, aturan, dan batasan. Bab 11 menelusuri ID ke sumber. Bab 12 menampung semua yang belum boleh dianggap selesai.
