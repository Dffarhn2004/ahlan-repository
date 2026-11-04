# 09 — File Manager berbasis filesystem Hermes / DataPonpes

## Requirement bisnis

Admin ponpes melihat tampilan folder/file seperti explorer. Folder fisik tetap berada **di volume persisten Hermes**, misalnya `DataPonpes/<tenant_uuid>/PPDB/brosur.pdf`; nama folder tampilan boleh `Ponpes Al-Falah`, tetapi path root memakai tenant UUID untuk keamanan dan stabilitas. Admin tidak dapat melihat/manipulasi profile/config internal Hermes atau root tenant lain.

## Model path dan physical root

```text
<volume_hermes>/DataPonpes/
  <tenant-uuid-A>/
    PPDB/
      brosur-2027.pdf
      syarat.pdf
      asrama.jpg
    Profil/
      sejarah.pdf
    Akademik/
      kurikulum.pdf
  <tenant-uuid-B>/
    PPDB/
      brosur.pdf
```

File manager bekerja dengan `file_id/folder_id`, **bukan** path absolut yang dikirim browser. Untuk UI breadcrumb simpan `parent_id` dan `display_name`; server memetakan ke root tenant. Operasi penamaan/move tidak boleh membuat path ilegal.

## Format P0

PDF, JPEG/JPG, PNG, TXT, dan MD (file MD dianggap konten, bukan instruksi executable). Untuk PDF bertingkat atau scan: file disimpan, tetapi content-read readiness mungkin `NEEDS_OCR`/`FAILED`; jangan klaim dapat dijawab Hermes bila teks belum tersedia. Default max per file 20 MB, max tenant 500 MB (bisa berubah di konfigurasi).

## Penyimpanan aman

1. Derive `tenant_id` dari session/admin atau identity profile terautentikasi, bukan dari body/path user.
2. Operasi disk memakai safe directory handle (Go `os.OpenRoot` pada Go >=1.24 atau alternatif `openat2` dengan pembatasan) serta `O_NOFOLLOW`/symlink checks; pastikan traversal tidak lolos walau file direname paralel.
3. Reject `..`, absolute paths, reserved names, symlinks, hard links tak dikenal, dan file type tidak diizinkan; cek magic bytes dan limits, bukan extension saja.
4. Upload ke temp path private → size limit → content signature + malware scan (bila tersedia) → checksum SHA-256 → `fsync` + atomic move → DB metadata. Bila DB gagal, worker rekonsiliasi orphan temp file.
5. Mode direktori/file restrictive; Go storage service hanya mount subtree `DataPonpes`, tidak diberikan `/etc`, profile secrets, atau Hermes home luas.
6. HTTP preview dengan MIME sesuai, `nosniff`, response limits, CSP sandbox dan attachment-download bila konten aktif. Jangan eksekusi PDF/JPG.
7. Data ponpes sering mencakup informasi anak; sanitasi EXIF bila diperlukan, sanitasi nama dan min content log.

## API file operations

`GET /v1/me/files?parent_id=...`, `POST /v1/me/folders`, `POST /v1/me/files` (multipart), `PATCH /v1/me/files/{id}`, `DELETE /v1/me/files/{id}`, `GET /v1/me/files/{id}/download`, `GET /v1/me/files/{id}/preview`. Idempotency key untuk upload/rename/delete mutasi tertentu.

## Metadata database minimum

`files(id, tenant_id, parent_folder_id, storage_key, filename, ext, mime, size_bytes, sha256, processing_status, version, created_by, deleted_at, created_at, updated_at)`; `folders(id, tenant_id, parent_id, name, created_at)`.

## Status dan konsistensi

- Saat upload sukses disk namun belum tercatat DB, jangan mengembalikan 201; job reconciler pulihkan atau hapus orphan.
- Rename/move harus atomik secara logis: transaksi metadata + operation log; bila filesystem gagal, rollback DB atau jadwalkan reconcile.
- Delete default soft-delete / `.trash` dibatasi dengan job purge terjadwal; log audit tidak otomatis ikut dihapus.
- File siap dilihat setelah checksum valid; file siap dijadikan referensi jawaban setelah content extraction/index `READY`.
- Backup menggabungkan snapshot DB dan filesystem dengan checkpoint manifest agar tidak mismatched.

## Contoh admin acceptance

Admin A dapat upload, preview, rename, delete, dan restore dokumen A. Admin A tidak bisa mengambil file B meski tahu `file_id`, melakukan URL tampering, mengunggah symlink, menyisipkan encoded traversal, atau mengatur `SOUL.md` internal Hermes via file manager.
