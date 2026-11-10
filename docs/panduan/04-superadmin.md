# 04 — Panduan superadmin

Ruang ini berlabel **Platform** di menu kiri. Gunakan bab ini jika setelah masuk Anda melihat **Dashboard superadmin**.

## Overview

Judul layar: **Dashboard superadmin**.

Yang tampil:

- Jumlah pesantren, layanan aktif, yang masih dalam penyiapan, dan yang perlu perhatian.
- Pertanyaan masuk, balasan asisten, pengunjung unik, dan jumlah percakapan, untuk 7, 14, atau 30 hari.
- Grafik aktivitas harian, status percakapan, pesantren paling aktif, dan jam tersibuk (WIB).
- Distribusi status tenant dan daftar penyiapan yang belum aktif.
- Lima aktivitas terakhir, dengan tautan **Lihat semua** ke audit.
- Aksi cepat: Kelola pesantren, Profil Asisten, Audit log.

Tombol **Tambah pesantren** di kanan atas membuka daftar pesantren, bukan langsung menyimpan data.

## Pesantren

Judul layar: **Pesantren**.

- Cari menurut nama atau slug.
- Saring **Semua**, **Aktif**, atau **Ditangguhkan**.
- Buka satu baris untuk masuk ke detail.

### Tambah pesantren

1. Tekan **Tambah Pesantren**.
2. Isi **Nama pesantren**, misalnya Pondok Darul Ilmi.
3. Periksa **Slug pesantren**. Slug terisi dari nama. Anda boleh mengubahnya. Huruf kecil, angka, dan tanda hubung saja.
4. Tekan **Buat pesantren**.

Pesantren tersimpan sebagai draf. Asisten belum dibuat pada langkah ini. Buka detail pesantren untuk menyiapkan asisten, kanal, dan dokumen.

Jika slug sudah ada, pembuatan gagal dan isian formulir tetap ada.

## Detail pesantren

Empat tab:

| Tab | Isi |
|---|---|
| Profil tenant | Tautan halaman publik, identitas, tombol **Edit profil** |
| Asisten & pengetahuan | Subtab di bawah |
| Aktivitas | Jejak perubahan pesantren ini |
| Pengguna | Akun yang boleh masuk ke pesantren ini |

Subtab **Asisten & pengetahuan**:

| Subtab | Kegunaan |
|---|---|
| Asisten | Siapkan, hubungkan Telegram, hubungkan WhatsApp, aktifkan, mulai ulang, perbarui, hapus |
| Admin WhatsApp | Mode khusus PSB atau umum, plus admin yang dihubungi |
| Pengetahuan | Unggah berkas dan **Perbarui pengetahuan asisten** |
| Tautan resmi | Tautan yang boleh dibagikan asisten |
| Kontak person | Orang yang boleh disebutkan ke pengunjung |
| Pertanyaan | Pertanyaan yang asisten belum bisa jawab |
| Log Telegram | Tabel atau kelompok percakapan, plus **Export Excel** |

Urutan kerja yang sesuai panel asisten:

1. **Siapkan asisten**.
2. **Hubungkan Telegram** atau **Hubungkan WhatsApp**. Pada uji coba, isi daftar penguji. Pada publik, siapa pun bisa mengobrol.
3. Unggah dokumen di **Pengetahuan**, lalu **Perbarui pengetahuan asisten**.
4. **Aktifkan**.

**Hapus asisten** hanya muncul di akun superadmin. Melepas kanal memakai **Lepas** di baris Telegram atau WhatsApp. Token tidak ditampilkan ulang.

## Profil Asisten

Menu **Profil Asisten** mengatur asisten semua pesantren sekaligus, bukan satu pesantren.

- Daftar profil memuat status runtime dan jumlah berkas.
- Tab pengaturan menyimpan deskripsi, teks persona, dan skill yang dipakai bersama.
- Deskripsi paling panjang 2000 karakter. Teks persona paling panjang 50000 karakter. Skill yang dipilih paling banyak 100.
- Sinkronisasi meminta konfirmasi. Jika Anda memilih mulai ulang asisten yang sedang aktif, bot terputus sesaat. Proses berjalan di latar. Layar menampilkan kemajuan sampai selesai atau gagal.

Jika katalog skill tidak terjangkau, daftar skill kosong dan aplikasi tetap membuka halamannya. Itu berarti layanan asisten sedang tidak memberi katalog, bukan berarti akun Anda ditolak.

## Audit log

**Audit log** di menu platform memuat jejak lintas pesantren: siapa, aksi apa, dan kapan. Contoh aksi yang sudah punya label Indonesia: pesantren dibuat, berkas diunggah, profil diperbarui, bot dihubungkan, percakapan ditandai perlu bantuan.

Membuka tab **Aktivitas** pada satu pesantren membatasi jejak ke pesantren itu.

## Sumber web yang disalahgunakan

Superadmin dapat mematikan satu sumber web pesantren walaupun admin pesantren menginginkannya tetap nyala. Sakelar ini tidak ada di menu sehari-hari admin. Admin tetap dapat menambah dan menjalankan sumber miliknya selama sumber itu tidak dipaksa mati.

## Yang perlu diingat

- Anda dapat membuka setiap pesantren dari daftar, termasuk mengubah isinya.
- Anda tidak dapat membaca token Telegram yang sudah disimpan.
- Formulir pengguna tidak menawarkan peran OWNER.
- Masuk tidak menanyakan kode MFA.
- Membuat pesantren dari layar ini adalah pekerjaan superadmin. Akun pesantren tidak punya menu Platform.
