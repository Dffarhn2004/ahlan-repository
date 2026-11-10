# 05 — Panduan admin pesantren

Ruang ini berlabel **Pesantren**. Setelah masuk, judul dashboard adalah nama pesantren Anda. Panduan ini untuk peran **Admin**. Staf dan peninjau memakai bab berikutnya. Mereka melihat menu yang sama, tetapi sebagian tombol ditolak.

Jika Anda mengurus lebih dari satu pesantren, aplikasi mengingat pilihan terakhir di peramban ini.

## Dashboard

Layar ini merangkum dokumen, kanal, percakapan, dan aktivitas pesantren yang sedang dipilih. **Muat ulang** mengambil data lagi. Kartu tautan membagikan halaman publik `/p/{slug}` bila slug sudah ada.

Widget yang tidak berhak Anda lihat dikosongkan, bukan menggagalkan seluruh halaman.

## Asisten AI

Judul: **Asisten & informasi pesantren**.

Di halaman ini Anda:

- menyiapkan dan mengaktifkan asisten,
- menghubungkan Telegram dan WhatsApp,
- mengisi identitas pesantren, alamat, telepon, email, situs, Google Maps, dan tautan media sosial,
- mengunggah logo.

### Menyiapkan kanal

Ikuti nomor pada panel:

1. **Siapkan asisten** jika belum ada.
2. **Hubungkan Telegram** atau **Ganti bot** jika sudah ada.
3. **Hubungkan WhatsApp** atau **Tautkan ulang**.
4. **Aktifkan**. Tombol ini baru bisa dipakai setelah asisten ada dan minimal satu kanal terhubung.

**Nonaktifkan** menghentikan asisten yang sedang berjalan. **Mulai ulang asisten** dan **Perbarui asisten** ada di kelompok aksi yang sama. Anda tidak melihat **Hapus asisten**. Itu hanya untuk superadmin.

### Telegram

Pada dialog **Hubungkan bot Telegram**:

- **Uji coba** — isi **ID Telegram penguji**. Pisahkan dengan koma. Hanya ID ini yang dapat mengobrol. ID bisa diketahui lewat @userinfobot di Telegram, seperti yang tertulis di bantuan isian.
- **Publik** — siapa pun yang menemukan bot dapat mengobrol. Gunakan setelah uji coba selesai.
- **Simpan & aktifkan**.

**Lepas** melepas bot. Token tidak bisa dibaca kembali dari halaman ini.

### WhatsApp

Dialog **Hubungkan WhatsApp** menampilkan QR. Pindai dengan ponsel pemilik nomor asisten. Jika QR kedaluwarsa, tekan **Buat ulang**. **Simpan & aktifkan** menunggu status terhubung.

Pilih juga:

- **Khusus PSB** atau **Umum**,
- **Uji coba** atau **Publik**.

Pada uji coba, isi nomor penguji dengan kode negara (62) atau awalan 0.

### Logo dan profil

Logo: PNG, JPEG, atau WEBP, paling besar 2 MB. Berkas SVG ditolak.

Simpan profil dari isian di halaman yang sama. Tautan Instagram, Facebook, TikTok, dan YouTube harus mengarah ke domain jenis itu. Google Maps disimpan sebagai tautan lokasi. Aplikasi tidak mengambil isi peta untuk dijadikan jawaban.

## Admin WhatsApp

Judul: **Admin WhatsApp**.

Halaman ini mengatur nomor yang sudah tertaut:

- **Khusus PSB** — asisten langsung menjawab.
- **Umum** — asisten bertanya keperluan. Selain PSB, percakapan diarahkan ke admin yang cocok.

**Tambah admin** meminta nama (paling panjang 80 karakter), nomor WhatsApp, dan deskripsi kapan admin itu dihubungi (paling panjang 200 karakter). Paling banyak 10 admin. Nomor yang sama tidak bisa dimasukkan dua kali.

Daftar ini bukan kontak publik. Pengunjung tidak melihat nomor admin routing di halaman `/p/{slug}`.

Tombol mode tidak bisa diubah sebelum nomor tertaut.

## Dokumen

Judul: **Dokumen pesantren**.

1. **Folder baru** bila Anda ingin mengelompokkan berkas.
2. **Unggah dokumen**. Satu berkas paling besar 20 MB.
3. **Perbarui pengetahuan asisten** setelah ada berkas. Tanpa langkah ini, asisten belum memakai unggahan terbaru.

Saat pembaruan berjalan, spanduk pengetahuan menampilkan kemajuan. Jika gagal, tombol coba lagi ada di spanduk itu.

Anda dapat mengganti nama, mengganti isi, dan menghapus berkas sendiri. Berkas pesantren lain tidak pernah muncul di daftar ini.

## Sumber web

Judul: **Sumber web**.

Asisten dapat mempelajari situs resmi dan berkas Google (Docs, Sheets, Slides, file Drive) milik pesantren. Pengambilan berjalan sesuai jadwal atau saat Anda memintanya.

Buka satu sumber untuk melihat **Hasil pengambilan**. Dari sana Anda dapat membaca halaman, memperbaiki teks, dan menandai halaman selesai ditinjau.

Yang tidak dipelajari sebagai teks, walaupun tautannya boleh disimpan di profil:

- WhatsApp (`wa.me` dan semacamnya),
- Instagram, Facebook, TikTok, YouTube,
- Google Form (`forms.gle`) dan Linktree.

Pengambilan juga menolak alamat jaringan privat dan mematuhi `robots.txt` situs tujuan. Kedalaman bawaan 2 tingkat dan paling banyak 50 halaman. Batas keras 3 tingkat dan 200 halaman. Satu halaman web dipotong pada 2 MB. Berkas Drive dipotong pada 20 MB.

Jika superadmin mematikan paksa sebuah sumber, menjalankan ulang dari akun pesantren tidak menghidupkannya.

## Percakapan

Judul: **Percakapan Telegram**.

Satu baris memuat pertanyaan pengunjung dan balasan asisten. **Ekspor Excel** tersedia bila ada data. **Muat ulang** mengambil catatan terbaru.

Gunakan halaman ini untuk melihat apa yang ditanyakan calon santri. Ubah status percakapan bila tim Anda sudah menindaklanjuti.

Lampiran yang dikirim pengunjung ke asisten dibersihkan dari cache asisten setelah 60 menit. Teks percakapan di dashboard tidak ikut terhapus pada jadwal itu.

## Pertanyaan

Judul: **Pertanyaan belum terjawab**.

Daftar ini terisi ketika balasan asisten memuat *belum tersedia di data kami*. Buka satu pertanyaan, tulis jawaban yang boleh dipakai asisten, atau tandai selesai maupun diabaikan.

Jawaban yang Anda simpan masuk ke bahan asisten pada pembaruan berikutnya. Jangan menaruh data yang tidak boleh dibaca calon santri di jawaban ini, karena jawaban itu menjadi bagian persona asisten.

## Pengguna

Judul: **Pengguna pesantren**.

**Tambah pengguna** meminta email, kata sandi sementara minimal 10 karakter, dan peran:

| Peran di formulir | Dipakai untuk |
|---|---|
| Admin | Pengurus yang mengatur isi, kanal, dan pengguna |
| Staf | Unggah dokumen, perbarui pengetahuan, urus percakapan |
| Peninjau | Membaca, tanpa mengubah |

Anda dapat mengubah peran pengguna yang sudah ada ke salah satu dari tiga peran itu. Peran OWNER tidak ada di pilihan.

## Aktivitas sistem

Menu **Aktivitas sistem** menampilkan jejak pesantren Anda: unggahan, perubahan profil, penyambungan bot, dan perubahan peran. Anda tidak melihat jejak pesantren lain.

## Halaman publik

Slug pesantren menjadi alamat `/p/{slug}`. Bagikan alamat itu dari kartu di dashboard. Pengunjung tidak masuk ke dashboard lewat halaman ini.

## Urutan kerja yang disarankan

1. Lengkapi profil, logo, tautan, dan kontak.
2. Unggah brosur, biaya, jadwal, dan syarat. Perbarui pengetahuan.
3. Siapkan asisten dan hubungkan kanal dalam mode **Uji coba**.
4. Coba pertanyaan dari akun penguji. Periksa **Percakapan** dan **Pertanyaan**.
5. Jawab celah yang muncul, perbarui pengetahuan lagi.
6. Ubah kanal ke **Publik** ketika uji coba selesai, lalu **Aktifkan**.
