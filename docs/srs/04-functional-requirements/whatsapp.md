# FR — WhatsApp

Sumber: `httpapi/whatsapp.go`, `httpapi/whatsapp_admins.go` (pemanggilan batas), `00015_whatsapp.sql`, `00020_whatsapp_service_mode.sql`, `frontend/src/components/whatsapp-routing.tsx`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-WA-001 | Penyimpanan kanal harus menerima hanya `access_mode` `TESTING` atau `PUBLIC` (default `TESTING`) dan `service_mode` `PSB_ONLY` atau `GENERAL` (default `PSB_ONLY`). | Must | [CONFIRMED] | `saveWhatsApp` | Nilai lain: 400. PUBLIC mengosongkan daftar nomor uji (cabang yang sama polanya dengan Telegram di fungsi itu). |
| FR-WA-002 | `PUT /api/v1/tenants/:id/whatsapp/service` harus mengubah hanya ke `PSB_ONLY` atau `GENERAL` dan memerlukan `bot:update`. | Must | [CONFIRMED] | `setWhatsAppService` | STAFF: 403. ADMIN mengirim `GENERAL`: mode tersimpan. |
| FR-WA-003 | Satu percobaan pairing harus berbatas waktu 3 menit. | Must | [CONFIRMED] | `pairingTimeout = 3 * time.Minute` | Status pairing tidak tetap berjalan tanpa batas waktu yang lebih panjang dari konstanta ini. |
| FR-WA-004 | Admin routing WhatsApp harus memenuhi: nama 1–80 karakter, telepon `^\+[0-9]{8,15}$`, alasan notifikasi 1–200 karakter, dan UI harus menolak lebih dari 10 admin. | Must | [CONFIRMED] | CHECK di `00020`, `MAX_ADMINS = 10` di `whatsapp-routing.tsx` | Nomor `0812...` tanpa normalisasi ke `+` gagal CHECK jika disimpan mentah. Entri ke-11 tidak dapat ditambah dari tombol UI. |
| FR-WA-005 | Nomor admin routing tidak boleh menjadi kontak publik. | Must | [CONFIRMED] | Komentar migrasi `00020`: tabel `whatsapp_admins` bukan `tenant_contacts` dan "never shown to visitors" | `GET /api/v1/public/p/:slug` tidak mengembalikan baris `whatsapp_admins`. Kontak publik hanya dari `tenant_contacts` jika endpoint publik memang memuatnya. |

### Arti mode layanan (teks migrasi, [CONFIRMED] sebagai spesifikasi yang disimpan bersama skema)

- **PSB_ONLY:** nomor khusus penerimaan; asisten menjawab seperti alur PSB.
- **GENERAL:** nomor bersama; asisten menanyakan keperluan, menjawab PSB sendiri, dan menyerahkan keperluan lain ke admin yang cocok plus notifikasi WhatsApp.

Satu kanal per tenant: `whatsapp_channels.tenant_id UNIQUE`. Status kanal: `PENDING`, `ACTIVE`, `SUSPENDED`.

Handoff mesin: `POST /api/v1/internal/hermes/handoff` tidak memakai sesi login (`routes.go`). Autentikasinya satu keluarga dengan event Hermes; lihat IF-013 di [07-external-interfaces.md](../07-external-interfaces.md).
