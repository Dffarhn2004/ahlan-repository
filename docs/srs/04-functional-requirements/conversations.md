# FR — Percakapan

Sumber: `routes.go`, `httpapi/telegram.go` (komentar dan cek peran), `00007_conversations.sql`, `00015_whatsapp.sql`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-CONV-001 | Sistem harus menyimpan percakapan per tenant dengan kanal `TELEGRAM` atau `WHATSAPP`, dan pasangan `(tenant_id, channel, telegram_chat_id)` harus unik. | Must | [CONFIRMED] | `00015_whatsapp.sql` constraint | Dua pesan kanal sama dan chat id sama pada tenant sama memperbarui/menolak duplikat sesuai upsert pemanggil, tidak membuat dua baris unik. |
| FR-CONV-002 | Statistik percakapan tanpa `tenant_id` harus hanya untuk superadmin. | Must | [CONFIRMED] | komentar di `telegram.go`: "only a superadmin may see" | Anggota memanggil stats tanpa tenant: ditolak. Superadmin menerima agregat. |
| FR-CONV-003 | Pengubahan status percakapan harus memerlukan `conversation:manage` pada tenant percakapan itu, kecuali superadmin. | Must | [CONFIRMED] | `setConversationStatus` path di `telegram.go` memakai `authz.Allows(role, "conversation:manage")` | VIEWER: 403. STAFF: bukan 403 dari izin ini. |
| FR-CONV-004 | Event percakapan dari Hermes harus masuk lewat `POST /api/v1/internal/hermes/events` tanpa sesi login, dengan autentikasi header `X-Hermes-Event-Key`. | Must | [CONFIRMED] | `routes.go`, `README.md` | Tanpa header yang cocok dengan rahasia server: ditolak. Dengan kunci yang benar: event diterima sesuai handler `hermesEvent`. |

### Status yang terlihat di label UI

Label audit yang sudah dipetakan, bukan daftar status DB yang tertutup: `CONVERSATION_HANDOFF`, `CONVERSATION_RESOLVED`, `CONVERSATION_ACTIVE` (`frontend/src/lib/labels.ts`). [CONFIRMED] sebagai teks UI untuk aksi tersebut.

PRD mensyaratkan persetujuan sebelum superadmin membaca isi pesan. Kode tidak menunjukkan langkah approval itu. Jangan menulis uji "superadmin diblokir tanpa approval" sebagai perilaku saat ini. Lihat OQ-05.

Lampiran pengunjung di cache profile dihapus setelah 60 menit (`attachmentMaxAgeMinutes = 60` di `hermes.go`). [CONFIRMED] untuk jalur purge itu.
