# 10 — Batasan

| ID | Batasan | Status | Sumber |
|---|---|---|---|
| CON-01 | Control plane adalah Go (Fiber, pgx). Bukan NestJS dan bukan Gin, meskipun PRD awal menyebut Gin. | [CONFIRMED] | `README.md`, `go.mod` |
| CON-02 | PostgreSQL wajib. Tidak ada penyimpanan metadata tenant lain di kode. | [CONFIRMED] | `DATABASE_URL` |
| CON-03 | Object storage S3/MinIO tidak dipakai pada MVP yang tertulis. | [CONFIRMED] | keputusan di `README.md`; tidak ada SDK S3 di `go.mod` |
| CON-04 | Produksi memakai bridge socket, bukan Docker CLI dari API. Development boleh `HERMES_EXEC_MODE=docker`. | [CONFIRMED] | `.env.example` |
| CON-05 | Deploy yang terdokumentasi menargetkan VPS dengan alias SSH `chatbot-ahlan-server` dan `ahlan-server` serta unit systemd `ahlan-api`, `ahlan-worker`, `ahlan-hermes-bridge`. | [CONFIRMED] | `README.md` |
| CON-06 | Crawl dijadwalkan dalam zona WIB (UTC+7) di kode crawler. | [CONFIRMED] | `var wib` di `crawler.go` |
| CON-07 | Jawaban chat bergantung pada proses Hermes yang terpasang. Versi Hermes tidak dipin di `go.mod`. | [CONFIRMED] | `README.md` batas keandalan PRD |
| CON-08 | Anggaran, tanggal pilot, dan target RPO/RTO tidak ada di repositori. | [TBD] | PRD NFR-07 menyebut "tetapkan RPO/RTO sebelum komersial" tanpa angka |
| CON-09 | Bahasa antarmuka pengurus yang terlihat di shell dan label audit adalah Indonesia. | [CONFIRMED] | `labels.ts`, salinan `app-shell.tsx` |
| CON-10 | Data anak/calon santri tidak boleh diolah di luar kebutuhan informasi PSB, menurut batas PRD. Tidak ada daftar bidang PII yang disahkan penasihat hukum di repo. | [TBD] | out-of-scope `02_ruang_lingkup_requirements.md` |

Batasan hukum yang disebut PRD tetapi tidak dijabarkan sebagai kontrol teknis (UU PDP, retensi, dasar pemrosesan) tetap pertanyaan terbuka OQ-10, bukan klaim patuh.
