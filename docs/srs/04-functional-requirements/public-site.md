# FR — Situs publik

Sumber: `routes.go` (route sebelum `authenticate`), `frontend/src/app/p/`, `frontend/src/app/page.tsx`, `frontend/src/lib/branding.ts`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-PUB-001 | `GET /api/v1/public/p/:slug`, `.../logo`, `.../links/:kind`, dan `GET /api/v1/public/p` harus dapat dipanggil tanpa cookie sesi. | Must | [CONFIRMED] | urutan route di `App()` | Permintaan tanpa cookie tidak menjawab 401 dari middleware `authenticate`. |
| FR-PUB-002 | Parameter `kind` pada tautan publik harus cocok dengan `^[A-Z]{3,20}$`. | Must | [CONFIRMED] | `kindParam` di `logo.go` | `kind=ab` atau huruf kecil ditolak oleh regex itu pada handler yang memakainya. |
| FR-PUB-003 | Frontend harus menyediakan halaman publik profil di `/p/[slug]` dan `/p/[slug]/[kind]`. | Must | [CONFIRMED] | `frontend/src/app/p/[slug]/page.tsx` dan `[kind]/page.tsx` | Slug yang API-nya mengembalikan data menampilkan profil. Slug tidak dikenal menampilkan keadaan gagal dari respons API, bukan data tenant lain. |
| FR-PUB-004 | Beranda `/` harus menjelaskan kanal WhatsApp dan Telegram, jawaban dari dokumen unggahan, dan pencatatan pertanyaan yang tidak ada di dokumen. | Must | [CONFIRMED] | FAQ di `page.tsx` | Teks FAQ yang terbit memuat ketiga pernyataan itu. |

### Halaman pemasaran lain yang ada di pohon route

`/chatbot-pesantren`, `/chatbot-whatsapp-pesantren`, `/penerimaan-santri-baru`, `/panduan-psb-pesantren`, `/kebijakan-privasi`, `/sitemap` lewat `sitemap.ts`. Isi hukum lengkap tidak disalin ke SRS. Keberadaan route adalah [CONFIRMED]. Apakah teks kebijakan sudah cukup untuk kepatuhan UU PDP adalah [TBD] (OQ-10).

Nama tampilan "AHLAN" dan logo `/logo.png` adalah default. Pengguna dapat menimpa nama/logo di `localStorage` kunci `ahlan.branding` tanpa API (`branding.ts`). Perubahan itu hanya di peramban tersebut [CONFIRMED].
