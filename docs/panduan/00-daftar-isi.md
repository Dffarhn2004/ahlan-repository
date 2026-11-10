# Panduan AHLAN

Panduan ini untuk orang yang memakai aplikasi: superadmin platform, pengurus pesantren, staf, dan peninjau. Isinya mengikuti layar dan aturan yang ada di aplikasi pada 10 November 2026.

Spesifikasi requirement (ID, status, kriteria uji) tetap di [docs/srs](../srs/00-index.md). Panduan ini tidak mengulang tabel itu. Kalau layar dan panduan berbeda, ikuti layar.

## Cara membaca

| Anda adalah | Mulai dari |
|---|---|
| Baru mengenal produk | [01 — Mengenal AHLAN](01-mengenal-ahlan.md) lalu [02 — Peran dan use case](02-peran-dan-use-case.md) |
| Superadmin perusahaan | [03 — Masuk](03-masuk.md) lalu [04 — Superadmin](04-superadmin.md) |
| Admin pesantren | [03 — Masuk](03-masuk.md) lalu [05 — Admin pesantren](05-admin-pesantren.md) |
| Staf atau peninjau | [06 — Staf dan peninjau](06-staf-dan-peninjau.md) |
| Calon santri atau wali | [07 — Calon santri dan halaman publik](07-calon-santri-dan-halaman-publik.md) |

## Daftar bab

1. [Mengenal AHLAN](01-mengenal-ahlan.md) — untuk apa produk ini, dua ruang kerja, apa yang tidak dilakukan aplikasi.
2. [Peran dan use case](02-peran-dan-use-case.md) — diagram, matriks hak, alur tiap peran.
3. [Masuk dan keluar](03-masuk.md) — login, arah setelah masuk, sesi.
4. [Panduan superadmin](04-superadmin.md) — dashboard platform, pesantren, asisten global, audit.
5. [Panduan admin pesantren](05-admin-pesantren.md) — profil, kanal, dokumen, sumber web, percakapan, pertanyaan, pengguna.
6. [Staf dan peninjau](06-staf-dan-peninjau.md) — apa yang bisa dikerjakan, apa yang ditolak aplikasi.
7. [Calon santri dan halaman publik](07-calon-santri-dan-halaman-publik.md) — bot, halaman `/p/{slug}`, kalimat saat jawaban tidak ada.
8. [Batas angka dan istilah](08-batas-dan-istilah.md) — ukuran berkas, kuota, kamus singkat.

## Peta menu

Menu kiri bergantung pada ruang kerja, bukan pada peran di dalam pesantren. Staf dan peninjau melihat menu yang sama dengan admin. Jika suatu aksi ditolak, halaman atau tombol menampilkan keadaan gagal, dan server menjawab larangan.

**Platform** (setelah masuk sebagai superadmin): Overview, Pesantren, Profil Asisten, Audit log, Pengaturan.

**Pesantren:** Dashboard, Asisten AI, Admin WhatsApp, Dokumen, Sumber Web, Percakapan, Pertanyaan, Pengguna, Aktivitas sistem.

Nama aplikasi bawaan di layar adalah **AHLAN**. Nama dan logo di peramban Anda bisa berubah lewat Pengaturan; perubahan itu tersimpan di peramban itu saja.
