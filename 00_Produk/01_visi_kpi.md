# 01 — Visi produk, problem, KPI

## Masalah yang diselesaikan

Calon santri/wali sering bertanya berulang tentang persyaratan, biaya, pilihan pendidikan, asrama, jadwal, dan prosedur PPDB. Setiap ponpes memiliki informasi berbeda dan dapat berubah antartahun. Admin menghabiskan waktu untuk respons manual; tidak ada sumber jawaban terkontrol maupun statistik pusat. Perusahaan ingin menawarkan satu layanan SaaS yang dapat dioperasikan secara mandiri oleh banyak ponpes.

## Visi

**Satu ponpes, satu asisten Telegram resmi, satu basis pengetahuan yang dapat dikendalikan admin; satu perusahaan dapat memantau dan mengoperasikan banyak ponpes secara aman.**

## Value proposition

- **Bagi calon santri/wali:** informasi cepat dengan rujukan dokumen, jalur jelas menuju pendaftaran atau admin manusia.
- **Bagi ponpes:** file manager sederhana, bot terpisah dengan identitas ponpes, pengurangan pertanyaan berulang, histori dan analitik.
- **Bagi perusahaan:** provisioning tenant otomatis, kontrol status tenant, biaya AI, penyelesaian masalah, dan fondasi SaaS berlangganan.

## North star dan KPI (target awal untuk pilot, bukan janji SLA)

| Metrik | Definisi | Sasaran pilot (diusulkan) |
|---|---|---|
| Bot availability | Uptime jalur balas Telegram per tenant pada jam uji | >= 99% saat jam uji |
| Answer grounding | Persentase jawaban faktual yang dapat ditelusuri ke dokumen resmi | >= 90% pada set evaluasi terkurasi |
| Cross-tenant leakage | Jawaban/files tenant B terakses melalui tenant A | **0 insiden**, semua negative tests lulus |
| Unsupported claim | Jawaban mengarang biaya/jadwal/syarat yang tidak ada dalam dokumen | <= 2% pada sampel evaluasi awal |
| Provisioning success | Tenant baru aktif tanpa langkah shell manual | >= 95% pada uji ulang yang terkontrol |
| Document freshness | Dokumen yang diubah sudah tercermin pada hasil pencarian | <= 5 menit setelah processing sukses untuk dokumen MVP |
| Event ingest success | Event yang diterima Go tervalidasi dan tersimpan | >= 99.9% untuk request API yang sudah menerima HTTP 2xx |
| Task completion | Admin menyelesaikan unggah dokumen dan aktivasi bot sendiri | >= 80% pada usability test |

> **Penting:** logging webhook Hermes adalah best-effort; metrik event completeness dari Hermes tidak boleh dijanjikan 100% tanpa menambah sumber observasi persisten/reconciliation.

## Sukses pilot

Dua ponpes berbeda dapat mengelola dokumen dan bot masing-masing, salah satu bisa suspend/resume tanpa mengganggu yang lain, pengguna Telegram mendapat jawaban yang tepat atau abstain, Go memiliki audit bisnis dan monitoring memadai, dan test akses silang lulus.

## Bukan tujuan fase awal

Tidak membangun sistem akademik/SIAKAD, virtual classroom, payment gateway, e-KTP verification, auto-bot creation melalui BotFather, model training fine-tune, multi-channel WhatsApp, atau marketing otomatis.
