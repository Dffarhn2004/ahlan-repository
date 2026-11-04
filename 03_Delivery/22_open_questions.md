# 22 — Asumsi, hal yang perlu diputuskan, dan keputusan tertunda

PRD sudah bisa digunakan sebagai baseline pengembangan. Pertanyaan berikut **tidak menghentikan pengerjaan fondasi**; setiap poin memiliki default aman.

| Q | Pertanyaan | Default kerja saat ini | Kapan diputuskan |
|---|---|---|---|
| Q-01 | Nama merek dan domain produk? | Placeholder `Ponpes AI Platform` | Sebelum public web |
| Q-02 | Bot dimiliki ponpes atau perusahaan? | BotFather admin ponpes, platform simpan token terenkripsi | Onboarding policy |
| Q-03 | Apakah calon santri benar-benar boleh chat publik tanpa daftar? | Ya, **hanya setelah security gate Hermes lulus** | Tahap 0 |
| Q-04 | Multiplex atau container tenant? | Multiplex dengan scoped Go tool, fallback isolated container | Tahap 0/Pilot |
| Q-05 | Apakah ingin Hermes membaca native `read_file` dari folder langsung? | Tidak di multiplex publik; tool Go membaca file volume Hermes | Tahap 0 |
| Q-06 | Jenis dokumen wajib? | PDF/JPG/PNG/TXT/MD, OCR scan bukan jaminan P0 | Sprint 3 |
| Q-07 | Batas storage tenant? | 500 MB; file max 20 MB | Pricing design |
| Q-08 | Apakah ada workflow persetujuan dokumen? | Admin tenant langsung publish setelah processing (P0) | P1 |
| Q-09 | Apakah chat penuh harus disimpan? | Off/minimal telemetry by default; optional opt-in audited | Privacy review |
| Q-10 | Akses superadmin ke transkrip? | Restricted support mode + audit + purpose, bukan akses default | Privacy review |
| Q-11 | Retensi dokumen/trash? | Soft-delete 7 hari; audit usulan 1 tahun | Legal/ops |
| Q-12 | Model/provider AI & anggaran per ponpes? | Perusahaan memilih model policy; meter usage | Tahap 0 |
| Q-13 | Pendaftaran ditangani bot atau tautan web? | P0 link; P1 lead/handoff/form | After pilot |
| Q-14 | Legalitas penanganan data anak/wali dan PDP? | Minimize data, formal review sebelum feature P1 | Before public pilot |
| Q-15 | Hosting tunggal atau multi-host? | Single-host Docker Compose pilot | Setelah uji kapasitas |
| Q-16 | SLA berbayar dan harga? | Belum ditetapkan; telemetry dan benchmark dahulu | P1 |
| Q-17 | Backup RPO/RTO yang disepakati? | RPO 24h, RTO 8h hanya target kerja untuk drill | Before commercial SLA |
| Q-18 | Admin ponpes bebas ubah persona? | Form field aman dan renderer template, bukan raw SOUL config | Sprint 2 |
| Q-19 | Sumber crawl apa yang diprioritaskan? | Situs web resmi satu domain; media sosial tidak didukung | Sebelum CRAWL-01 |
| Q-20 | Hasil crawl langsung dipakai bot atau perlu persetujuan admin? | Langsung READY dengan sitasi URL; unggahan menang bila bertentangan; admin dapat exclude. Opsi: mode "review dulu" per sumber | Setelah pilot |
| Q-21 | Batas frekuensi/kuota crawl per paket? | Interval >= 6 jam, 5 sumber, 50 halaman/run, manual 1/10 menit | Pricing design |
| Q-22 | Perlu render JavaScript untuk situs SPA? | Tidak di MVP; status `NEEDS_RENDER` | P1 lanjut |
| Q-24 | Dokumen Google privat (OAuth / service account)? | Tidak; hanya dokumen "siapa saja dengan link" | P1 lanjut |
| Q-23 | Field lokasi tambahan (alamat teks, koordinat manual) selain link Maps? | Hanya `google_maps_url` (+ lat/lng opsional dari URL) | Sprint 1 |

## MVP principles saat ambiguity

- **Default deny** untuk permissions dan network.
- **Default no raw PII** dalam observability dan dashboard agregat.
- **Default document authoritative only when READY**.
- **Default no auto-public bot** hingga security test selesai.
- **Default no mandatory RAG vector DB** hingga kebutuhan retrieval dibuktikan.
- **Default no AI write access** kecuali tool domain spesifik dengan auth/validation.

## Product decisions yang bisa dibuat setelah pilot

Billing paket tenant; self-onboarding; human handoff exact channel; OCR/vision pipeline; advanced semantic search; multi-language answers; SLA & per-tenant tier resource policy; per-tenant runtime isolation standard.
