# FR — Tautan resmi dan crawl

Sumber: `internal/links/links.go`, `httpapi/links.go`, `internal/crawler/crawler.go`, `internal/crawler/guard.go`, `00012_crawl_sources.sql`, route crawl di `routes.go`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-KB-001 | Sistem harus menerima hanya jenis tautan yang ada di peta `kinds`, paling banyak 20 tautan per tenant, dan host yang wajib untuk jenis bersyarat. | Must | [CONFIRMED] | `MaxLinksPerSite = 20`, `kinds`, `ErrLinkLimit` | Tautan ke-21 ditolak. `INSTAGRAM` ke domain di luar `instagram.com` / `instagr.am`: `LINK_DOMAIN_MISMATCH`. Jenis tidak dikenal: `LINK_KIND_INVALID`. |
| FR-KB-002 | Crawl situs harus memakai kedalaman default 2 dan halaman default 50, dan tidak boleh melebihi kedalaman 3 serta 200 halaman. Halaman HTML dibatasi 2 MiB; unduhan Google Drive dibatasi 20 MiB. | Must | [CONFIRMED] | konstanta di `crawler.go` | Opsi `MaxPages=1000` dipotong ke 200. |
| FR-KB-003 | Crawler harus menolak alamat loopback, privat, link-local, dan prefix yang diblokir di `guard.go`, serta menghormati `robots.txt` untuk user-agent crawler. | Must | [CONFIRMED] | `IPAllowed`, tes `crawler_test.go` | Target host yang mengarah ke IP privat gagal. Path yang di-disallow robots tidak diambil (`ErrForbiddenURL` pada start page yang dilarang). |
| FR-KB-004 | Admin tenant (`tenant:update`) harus dapat menambah, mengubah, menghapus, menjalankan sumber, serta membaca run, event, dan halaman. Sunting isi halaman memakai izin yang sama. | Must | [CONFIRMED] | route crawl di `tenantRoutes` | VIEWER dapat GET. VIEWER tidak dapat POST run (403). |
| FR-KB-005 | `PATCH /api/v1/crawl/sources/:sid/force` harus hanya untuk superadmin. | Must | [CONFIRMED] | `requireSuperadmin`, komentar `forceCrawlSource` | ADMIN pesantren: 403. Superadmin dapat mematikan sumber meskipun admin pesantren menginginkannya nyala (`forced_disabled` di `00012`). |

### Jenis tautan

| Kind | Boleh dipelajari crawler | Host |
|---|---|---|
| PENDAFTARAN, BROSUR, JADWAL, BIAYA, WEBSITE, LAINNYA | Ya, kecuali host di daftar `unlearnable` | https publik (aturan `Validate`) |
| WHATSAPP | Tidak (host ada di `unlearnable`) | `wa.me`, `api.whatsapp.com`, `chat.whatsapp.com`, `wa.link` |
| INSTAGRAM, YOUTUBE, FACEBOOK, TIKTOK | Tidak | allowlist di `kinds` |

Interval belajar tautan yang dijadikan sumber: `learnIntervalMinutes = 1440` di `links.go` (24 jam). [CONFIRMED] untuk jalur learn-from-link. Apakah setiap sumber crawl manual memakai angka yang sama perlu dilihat per handler; jangan menyamakan dengan klaim PRD "minimal 6 jam" tanpa uji (PRD FR-024). Perbedaan itu [TBD] OQ-09.

Host yang tidak dipelajari sebagai teks: `forms.gle`, `wa.me`, `instagram.com`, `facebook.com`, `fb.com`, `fb.me`, `tiktok.com`, `youtube.com`, `youtu.be`, `linktr.ee`.
