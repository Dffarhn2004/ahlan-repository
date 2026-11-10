# SRS — AHLAN / SaaS AI Penerimaan Santri

Dokumen yang dibaca pengurus ada di [panduan pengguna](../panduan/00-daftar-isi.md). Folder ini tetap spesifikasi requirement.

| Item | Nilai |
|---|---|
| Identitas dokumen | SRS-AHLAN-001 |
| Produk (nama di UI) | AHLAN (`frontend/src/lib/branding.ts`, default `appName: "AHLAN"`) |
| Nama proyek di PRD | SaaS AI Penerimaan Santri — Go × Hermes × Telegram (`README.md`) |
| Versi SRS | 0.2 |
| Tanggal | 10 November 2026 |
| Mode | Keduanya. Baseline adalah **as-is** (perilaku yang terbukti di kode). Item PRD yang belum terbukti di kode berstatus **[TBD]** dan tidak ditulis seolah sudah berjalan. |
| Standar acuan | Struktur mengikuti ISO/IEC/IEEE 29148 / IEEE 830 (pengantar, deskripsi umum, requirement spesifik, antarmuka, data). ID dan status mengikuti instruksi proyek ini. |
| Bahasa | Indonesia |
| Sumber utama | Kode `backend/`, `frontend/`, migrasi `backend/db/migrations/`, dan dokumen PRD di akar repositori. Jika PRD dan kode berbeda, **kode yang berlaku** (catatan yang sama ada di `README.md`). |

## Riwayat revisi

| Versi | Tanggal | Perubahan |
|---|---|---|
| 0.1 | 2026-11-10 | Draf awal dari reconnaissance repositori. |
| 0.2 | 2026-11-10 | OQ-01, OQ-02, OQ-03, OQ-07, dan OQ-15 ditutup sebagai ditangguhkan. Perilaku kode tetap requirement as-is. |

## Cara membaca

1. Mulai dari [03 — Peran](03-stakeholders-and-roles.md) dan [05 — Use case](05-use-cases.md) untuk melihat apa yang boleh dilakukan tiap peran.
2. Requirement fungsional ada di [04-functional-requirements/](04-functional-requirements/). Satu baris tabel = satu kewajiban yang dapat diuji.
3. Label status:
   - **[CONFIRMED]** terbukti di kode, migrasi, atau kutipan dokumen yang dikutip di kolom Sumber.
   - **[INFERRED]** disimpulkan dari kode, perlu dikonfirmasi pemilik produk.
   - **[TBD]** belum diketahui atau PRD dan kode bertentangan. Daftar lengkap ada di [12](12-open-questions-and-risks.md).
4. Prioritas memakai MoSCoW terhadap perilaku yang ada sekarang: **Must** = jalur yang sudah dipakai produk, **Should** = ada di kode tetapi bukan gerbang rilis, **Could** = disebut PRD dan belum menjadi perilaku kode, **Won't** = dinyatakan di luar lingkup PRD.

## Daftar isi

| File | Isi |
|---|---|
| [01-introduction.md](01-introduction.md) | Tujuan, ruang lingkup, definisi, referensi |
| [02-overall-description.md](02-overall-description.md) | Perspektif produk, fungsi, pengguna, lingkungan, asumsi |
| [03-stakeholders-and-roles.md](03-stakeholders-and-roles.md) | Aktor dan matriks hak akses |
| [04-functional-requirements/](04-functional-requirements/) | Requirement per modul |
| [05-use-cases.md](05-use-cases.md) | Use case dan diagram Mermaid per peran |
| [06-non-functional-requirements.md](06-non-functional-requirements.md) | NFR yang terukur dari kode atau PRD |
| [07-external-interfaces.md](07-external-interfaces.md) | UI, API, integrasi |
| [08-data-requirements.md](08-data-requirements.md) | Entitas dan ERD |
| [09-business-rules.md](09-business-rules.md) | Aturan bisnis dan batas angka |
| [10-constraints.md](10-constraints.md) | Batasan teknis dan hukum yang tertulis |
| [11-traceability-matrix.md](11-traceability-matrix.md) | Requirement ke sumber dan use case |
| [12-open-questions-and-risks.md](12-open-questions-and-risks.md) | TBD, risiko, kode tanpa requirement produk |
| [13-glossary.md](13-glossary.md) | Istilah |

### Modul fungsional

| File | ID |
|---|---|
| [auth.md](04-functional-requirements/auth.md) | FR-AUTH-001 … FR-AUTH-006 |
| [tenancy.md](04-functional-requirements/tenancy.md) | FR-TENANT-001 … FR-TENANT-008 |
| [members.md](04-functional-requirements/members.md) | FR-MEM-001 … FR-MEM-004 |
| [files.md](04-functional-requirements/files.md) | FR-FILE-001 … FR-FILE-006 |
| [assistant.md](04-functional-requirements/assistant.md) | FR-AST-001 … FR-AST-006 |
| [telegram.md](04-functional-requirements/telegram.md) | FR-TG-001 … FR-TG-004 |
| [whatsapp.md](04-functional-requirements/whatsapp.md) | FR-WA-001 … FR-WA-005 |
| [knowledge.md](04-functional-requirements/knowledge.md) | FR-KB-001 … FR-KB-005 |
| [public-site.md](04-functional-requirements/public-site.md) | FR-PUB-001 … FR-PUB-004 |
| [conversations.md](04-functional-requirements/conversations.md) | FR-CONV-001 … FR-CONV-004 |
| [questions-contacts.md](04-functional-requirements/questions-contacts.md) | FR-QC-001 … FR-QC-004 |
| [audit.md](04-functional-requirements/audit.md) | FR-AUD-001 … FR-AUD-002 |

File yang diminta template tetapi tidak dibuat terpisah: tidak ada. Semua 00–13 ada. Tidak ada modul kode besar yang digabung tanpa alasan: Telegram dan WhatsApp dipisah karena mode akses dan tabelnya berbeda.

## Ringkasan status

Dihitung dari baris requirement FR-*, NFR-*, dan BR-* (bukan dari use case atau antarmuka).

| Status | Jumlah |
|---|---|
| [CONFIRMED] | 76 |
| [INFERRED] | 1 |
| [TBD] | 3 |
| **Total FR + NFR + BR** | **80** |

Rincian: 58 FR (semua [CONFIRMED]), 10 NFR (8 [CONFIRMED], 1 [INFERRED] NFR-REL-001, 2 [TBD] NFR-PERF-001 dan NFR-COMP-001), 12 BR (11 [CONFIRMED], 1 [TBD] BR-010). Use case UC-001 … UC-013 dan antarmuka IF-001 … IF-022 tidak masuk angka ini.

Rincian ID ada di [11-traceability-matrix.md](11-traceability-matrix.md). Angka ini harus dihitung ulang jika requirement ditambah.

## Validasi yang ditangguhkan

Pada 10 November 2026 pemilik produk menunda lima butir berikut. SRS tidak menunggu jawaban. Perilaku yang terbukti di kode tetap requirement [CONFIRMED]. Butirnya: OQ-01 (siapa boleh membuat tenant), OQ-02 (MFA), OQ-03 (baca overview Hermes), OQ-07 (peran OWNER), OQ-15 (menu pesantren tidak disaring izin). Rincian ada di [12](12-open-questions-and-risks.md).
