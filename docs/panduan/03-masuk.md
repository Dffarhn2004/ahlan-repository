# 03 — Masuk dan keluar

## Masuk

1. Buka halaman masuk (`/login`).
2. Isi **Alamat email** dan **Kata sandi**.
3. Tekan **Masuk ke dashboard**.

Aplikasi memeriksa apakah Anda sudah punya sesi. Jika sudah, Anda tidak mengisi formulir lagi dan langsung diarahkan.

| Akun | Halaman tujuan |
|---|---|
| Superadmin | `/admin/dashboard` — judul layar **Dashboard superadmin** |
| Anggota pesantren | `/ponpes/dashboard` — judul memakai nama pesantren |
| Bukan superadmin dan tidak terdaftar di pesantren mana pun | Dashboard platform. Daftar pesantren untuk akun ini kosong |

Jika email atau kata sandi tidak cocok, pesan yang muncul adalah: *Email atau kata sandi tidak dapat diverifikasi.*

Butuh bantuan akses: tulisan di halaman masuk mengarah ke ahlan.platformai@gmail.com.

## Sesi

Sesi berlaku 24 jam. Cookie sesi tidak dapat dibaca skrip halaman. Menutup tab tidak selalu mengakhiri sesi. Untuk mengakhiri sesi, gunakan **Keluar**.

## Keluar

1. Buka menu akun di kanan atas.
2. Tekan **Keluar**.

Setelah keluar, membuka halaman dashboard akan mengembalikan Anda ke halaman masuk.

## Jika sesi gagal diperiksa

Layar menampilkan *Sesi belum dapat diperiksa* bila koneksi ke layanan terputus, plus tombol **Coba lagi**. Ini berbeda dari kata sandi yang salah.

## Pengaturan tampilan

**Pengaturan** di menu akun mengubah nama aplikasi dan logo yang terlihat di peramban Anda. Nilai bawaan adalah AHLAN dan `/logo.png`. Perubahan ini tidak mengubah nama pesantren dan tidak berlaku untuk pengurus lain.
