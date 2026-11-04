# 24 — Glosarium

| Istilah | Arti pada PRD ini |
|---|---|
| Tenant | Satu pondok pesantren sebagai pelanggan perusahaan |
| Superadmin | Operator internal perusahaan yang mengelola semua tenant |
| Admin Ponpes | Pengelola data ponpes sendiri di dashboard |
| Hermes Profile | State/konfigurasi/memory/sesi AI terpisah di Hermes |
| Multiplex Gateway | Satu proses gateway Hermes melayani beberapa profile |
| Parked | Profile belum aktif/berhenti melayani di multiplexer |
| Control Plane | Go backend yang mengatur tenancy, konfigurasi, data bisnis |
| Data Plane | Hermes chat runtime dan jalur jawaban calon santri |
| DataPonpes | Folder custom pada volume Hermes untuk file milik tiap tenant |
| Knowledge base | Dokumen resmi dan index/konten siap dibaca AI; sumber asli tetap filesystem |
| Retrieval tool | Tool terotorisasi untuk mencari excerpt dokumen tenant |
| Audit log | Bukti immutable/durable untuk tindakan admin/operasi bisnis |
| Telemetry | Event diagnostik AI/gateway, bisa hilang bila best-effort |
| Outbox | Record event yang disimpan satu transaksi bersama perubahan data |
| HMAC | Message authentication code untuk membuktikan asal dan integritas event |
| RLS | PostgreSQL Row Level Security; defense-in-depth tenant DB |
| Admission lead | Prospek pendaftar, belum berarti terdaftar atau diterima |
| Grounded answer | Jawaban AI yang didukung sumber dokumen resmi tersedia |
| Abstention | AI mengakui informasi tidak cukup alih-alih mengarang |
| Security gate | Acceptance test wajib sebelum tenant/bot dibuka untuk publik |
| Sumber crawl | URL milik ponpes yang diambil teksnya secara berkala/manual oleh worker Go |
| Crawl run | Satu eksekusi crawl (terjadwal atau manual) beserta hasil per halaman |
| SSRF | Serangan yang membuat server mengakses alamat internal; dicegah oleh guard crawler |
| Link Google Maps | Tautan lokasi resmi ponpes; disimpan dan dibagikan bot, tidak di-crawl |
