# 08 — Batas angka dan istilah

Angka di bawah adalah batas yang dipakai aplikasi sekarang. Melewatinya membuat aksi gagal dengan pesan di layar, bukan dengan pemotongan diam-diam, kecuali disebut lain.

## Batas

| Hal | Batas |
|---|---|
| Sesi masuk | 24 jam |
| Kata sandi pengguna baru | Minimal 10 karakter |
| Berkas pengetahuan | 20 MB per berkas |
| Kuota awal penyimpanan pesantren | 500 MB |
| Logo | 2 MB, PNG, JPEG, atau WEBP |
| Tautan resmi per pesantren | 20 |
| Kontak publik per pesantren | 20 |
| Admin routing WhatsApp | 10 |
| Nama admin WhatsApp | 1–80 karakter |
| Alasan admin dihubungi | 1–200 karakter |
| Nomor admin WhatsApp | Bentuk `+` diikuti 8–15 digit |
| Pertanyaan terbuka yang disimpan | 200, yang terlama terlepas |
| Jawaban yang diberikan ke asisten | 100 |
| Crawl, bawaan | Kedalaman 2, 50 halaman |
| Crawl, paling tinggi | Kedalaman 3, 200 halaman |
| Satu halaman web saat crawl | 2 MB |
| Berkas Google Drive saat crawl | 20 MB |
| QR WhatsApp | Sekitar 3 menit |
| Lampiran pengunjung di cache asisten | Dihapus setelah 60 menit |
| Deskripsi asisten global | 2000 karakter |
| Persona asisten global | 50000 karakter |
| Skill asisten global | 100 |

Kuota 500 MB adalah nilai awal data pesantren. Panduan ini tidak menjanjikan bahwa setiap unggahan dijumlahkan terhadap kuota itu sebelum ditolak. Batas yang pasti terasa di layar adalah 20 MB per berkas.

## Jenis tautan

| Jenis | Dipakai untuk | Dipelajari sebagai teks |
|---|---|---|
| Pendaftaran, brosur, jadwal, biaya, website, tautan lain | Halaman resmi | Ya, bila situsnya mengizinkan dan bukan domain yang dikecualikan |
| WhatsApp | `wa.me` dan domain WhatsApp yang diizinkan | Tidak |
| Instagram, YouTube, Facebook, TikTok | Domain masing-masing | Tidak |

## Istilah

| Istilah | Arti di aplikasi |
|---|---|
| AHLAN | Nama yang tampil di layar, kecuali diubah di pengaturan peramban |
| Pesantren / tenant | Satu pelanggan. Punya asisten, dokumen, dan kanal sendiri |
| Superadmin | Pengelola platform. Melihat semua pesantren |
| Admin | Pengurus penuh satu pesantren |
| Staf | Mengunggah dokumen dan mengurus percakapan |
| Peninjau | Hanya membaca |
| Slug | Alamat pendek pesantren, dipakai di `/p/{slug}` |
| Asisten | Profil Hermes satu pesantren |
| Uji coba | Hanya identitas yang didaftarkan yang boleh mengobrol |
| Publik | Siapa pun yang menemukan bot atau nomor boleh mengobrol |
| Khusus PSB | Nomor WhatsApp untuk penerimaan. Asisten langsung menjawab |
| Umum | Nomor WhatsApp bersama. Asisten menanyakan keperluan dulu |
| Pengetahuan | Dokumen dan hasil sumber web yang dipakai asisten menjawab |
| Perbarui pengetahuan | Menyalin bahan terbaru ke asisten |
| Pertanyaan belum terjawab | Pertanyaan yang balasannya memuat *belum tersedia di data kami* |
| Aktivitas / audit | Jejak siapa mengubah apa, dan kapan |
| Halaman publik | `/p/{slug}`, tanpa login |
| Draf | Status pesantren yang baru dibuat, sebelum layanan dinyatakan aktif |

## Jika sesuatu gagal

| Gejala | Yang perlu diperiksa |
|---|---|
| Tidak bisa masuk | Email dan kata sandi. Pesan kesalahan sengaja tidak membedakan keduanya |
| Menu pesantren terbuka tetapi aksi gagal | Peran Anda staf atau peninjau. Lihat bab 06 |
| Unggahan ditolak | Ukuran di atas 20 MB, atau nama berkas berisi folder |
| Logo ditolak | Bukan PNG, JPEG, atau WEBP, atau lebih dari 2 MB |
| QR WhatsApp habis | **Buat ulang**. Jangan biarkan dialog lebih dari sekitar 3 menit |
| Tombol aktifkan asisten mati | Asisten belum disiapkan, atau belum ada Telegram maupun WhatsApp |
| Jawaban asisten belum memuat berkas baru | **Perbarui pengetahuan asisten** belum dijalankan |
| Sumber web tidak jalan | Sumber dipaksa mati oleh superadmin, dilarang `robots.txt`, atau alamat tidak diizinkan |
| Tidak melihat aktivitas | Peran staf dan peninjau tidak memiliki akses jejak |
| Halaman publik kosong | Slug salah, atau profil dan tautan publik belum diisi |

Bantuan akun dari halaman masuk: ahlan.platformai@gmail.com.
