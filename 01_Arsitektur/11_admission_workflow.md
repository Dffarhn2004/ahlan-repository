# 11 — Lead, pendaftaran dan human handoff (P1)

## Batas fase

Pada **P0**, bot memberikan informasi dan link formulir resmi; tidak menganggap chatting sebagai pendaftaran sah. **P1** menambah lead, workflow status dan handoff. Menyimpan data anak membutuhkan pengaturan persetujuan, hak akses, dan retensi sebelum fitur aktif.

## Entities

`prospective_students` (minimal), `leads`, `admission_applications`, `admission_status_history`, `handoff_requests`. Jangan mengumpulkan NIK, dokumen identitas, data kesehatan atau detail sensitif kecuali diperlukan, ada dasar yang sah, dan perlindungannya memadai.

## Suggested flow P1

1. Wali menyatakan ingin mendaftar.
2. Hermes menginformasikan pemakaian data, meminta persetujuan sesuai alur legal dan menganjurkan wali mengisi web form resmi.
3. Go memberi `lead_form_url` scoped tenant dan token pendek-umur; tidak mempercayai data bebas dari chat sebagai fakta terverifikasi.
4. Web form menyimpan lead idempoten. Hermes boleh menanyakan status via tool `get_admission_status` dengan proof-of-ownership.
5. Bila pertanyaan berisiko/di luar dokumen, kirim handoff request; admin mendapat notifikasi dan dapat melanjutkan komunikasi melalui kanal yang disepakati.

## Tool whitelist P1

| Tool | Jenis | Syarat |
|---|---|---|
| `search_tenant_documents` | Read | Scoped tenant/profile |
| `get_admission_information` | Read | Dokumen terbit/approved |
| `create_lead_request` | Write | Consent, rate limit, idempotency, input validation |
| `get_application_status` | Read | OTP/secure token/proof of ownership |
| `request_human_handoff` | Write | Tenant scoped, anti-spam |

AI **tidak** diberi direct SQL, transfer uang, update status diterima, atau akses identitas pemohon lain.

## Transition status

Lead: `NEW -> CONTACTED -> QUALIFIED -> CONVERTED` atau `CLOSED`. Application: `DRAFT -> SUBMITTED -> UNDER_REVIEW -> ACCEPTED/REJECTED/WITHDRAWN`; hanya staf berwenang melalui dashboard bisa mengubah status resmi, dengan audit dan aturan transisi.

## Acceptance

Duplikat pesan/retry tidak menciptakan lead ganda; calon A tidak bisa mendapat status aplikasi B; admin tidak bisa menyunting tenant lain; informasi pembayaran/rekening hanya dari dokumen resmi aktif atau backend tenant, bukan model improvisation.
