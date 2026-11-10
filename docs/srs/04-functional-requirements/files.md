# FR — Berkas, folder, dan logo

Sumber: `httpapi/files.go`, `httpapi/logo.go`, `internal/storage/storage.go`, `00003_files.sql`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-FILE-001 | Pemegang `tenant:read` harus dapat melihat daftar berkas dan folder tenant sendiri. | Must | [CONFIRMED] | `GET .../files`, `GET .../folders`, izin `read` | Anggota tenant A tidak menerima daftar tenant B (403). VIEWER menerima 200. |
| FR-FILE-002 | Unggah dan penggantian berkas harus menolak ukuran di atas 20 MiB (`20 << 20`) dengan kegagalan permintaan, dan menyimpan berkas yang lolos di `DataPonpes/{tenant_uuid}` dengan nama aman plus checksum SHA-256. | Must | [CONFIRMED] | `maxUploadBytes`, `Store.Put` | Berkas 20 MiB + 1 byte ditolak. Berkas kecil tersimpan; metadata `size_bytes` dan `sha256` terisi; nama berisi `/` atau `..` ditolak `safeName`. |
| FR-FILE-003 | Pengganti nama (`PATCH`) dan penghapus (`DELETE`) harus mengikuti izin `file:update` dan `file:delete`. | Must | [CONFIRMED] | `routes.go` | STAFF dapat PATCH. STAFF tidak dapat DELETE (403). ADMIN dapat DELETE. |
| FR-FILE-004 | Pengunduhan `GET .../files/:file_id` dan `.../download` harus hanya mengembalikan berkas yang `tenant_id`-nya sama dan `deleted_at` kosong. | Must | [CONFIRMED] | `downloadFile`, query `GetFile` di `repository.go` | ID berkas tenant lain: tidak mengembalikan byte berkas itu (404 atau 403 sesuai urutan middleware lalu query). |
| FR-FILE-005 | Pembuatan folder harus memerlukan `file:create`. | Must | [CONFIRMED] | `POST .../folders` | VIEWER: 403. STAFF: folder baru tampil di GET folders. |
| FR-FILE-006 | Logo tenant harus berukuran paling besar 2 MiB, bertipe byte PNG, JPEG, atau WEBP, dan harus menolak SVG. | Must | [CONFIRMED] | `maxLogoBytes`, `logoTypes`, komentar `uploadLogo` | SVG atau `Content-Type` klien yang tidak cocok dengan sniff byte: 400 `LOGO_INVALID`. PNG valid ≤ 2 MiB: tersimpan. Izin tulis: `tenant:update`. |

### Alur unggah

**Normal.** Klien mengirim multipart. Server menolak jika `header.Size` melebihi batas, lalu `Put` menulis ke file sementara, menghitung SHA-256, dan menamai objek dengan UUID plus ekstensi.

**Kesalahan.** Nama path traversal ditolak sebelum tulis. Melebihi `max+1` byte pada reader ditolak `file too large`.

**Penyimpanan vs Hermes.** Metadata ada di tabel `files`. Root disk adalah `{DATA_ROOT}/DataPonpes/{tenant_id}` dengan mode direktori `0700`. Sinkron ke volume Hermes adalah aksi terpisah (FR-AST-003), bukan bagian dari `Put`.
