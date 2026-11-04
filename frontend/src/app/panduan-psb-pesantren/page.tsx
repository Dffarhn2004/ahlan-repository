import type { Metadata } from "next";
import { ContentPage, type ContentPageData } from "@/components/content-page";

export const metadata: Metadata = {
  title: { absolute: "Panduan PSB Pesantren: Menyiapkan Informasi untuk Calon Wali Santri | AHLAN" },
  description: "Panduan praktis menyiapkan penerimaan santri baru pesantren: dokumen yang perlu ada, pertanyaan yang sering muncul, dan cara menjaga informasi tetap konsisten.",
  alternates: { canonical: "/panduan-psb-pesantren" },
};

const data: ContentPageData = {
  path: "/panduan-psb-pesantren",
  kicker: "Panduan",
  title: "Panduan menyiapkan informasi PSB pesantren",
  lead: "Layanan PSB yang baik dimulai dari informasi yang jelas dan seragam. Panduan ini merangkum dokumen yang sebaiknya disiapkan dan pertanyaan yang hampir selalu muncul dari calon wali santri.",
  sections: [
    {
      heading: "Dokumen yang sebaiknya ada",
      list: [
        "Brosur PSB: gambaran umum pesantren, jenjang, dan keunggulan.",
        "Rincian biaya: biaya pendaftaran, uang pangkal, dan biaya bulanan, termasuk apa saja yang dicakup.",
        "Syarat dan berkas pendaftaran.",
        "Jadwal: gelombang pendaftaran, tes seleksi, pengumuman, dan masuk asrama.",
        "Tata tertib dan kegiatan harian santri.",
        "Kontak panitia dan lokasi.",
      ],
    },
    {
      heading: "Pertanyaan yang hampir selalu muncul",
      list: [
        "Berapa total biaya di tahun pertama?",
        "Apa saja berkas yang harus dibawa?",
        "Kapan pendaftaran ditutup?",
        "Apakah ada tes masuk, dan seperti apa?",
        "Bagaimana aturan kunjungan dan penggunaan gawai?",
        "Apakah ada keringanan atau beasiswa?",
      ],
    },
    {
      heading: "Menjaga informasi tetap konsisten",
      paragraphs: ["Tetapkan satu dokumen sebagai sumber kebenaran untuk setiap topik, lalu pastikan brosur, postingan media sosial, dan jawaban admin mengacu ke sana. Setiap kali ada perubahan jadwal atau biaya, perbarui dokumen utama terlebih dahulu."],
      list: ["Beri tanggal atau versi pada setiap dokumen.", "Hindari menyebut angka berbeda di tempat berbeda.", "Catat pertanyaan baru dari calon wali dan masukkan ke dokumen."],
    },
    {
      heading: "Dokumen yang baik juga membuat chatbot lebih akurat",
      paragraphs: ["Jika Anda memakai AHLAN, kualitas jawaban bergantung pada kejelasan dokumen. Tulis informasi dengan lengkap dan eksplisit, misalnya sebutkan rincian biaya per komponen, bukan hanya total."],
    },
  ],
  related: [
    { href: "/penerimaan-santri-baru", label: "Chatbot untuk penerimaan santri baru" },
    { href: "/chatbot-pesantren", label: "Apa itu chatbot pesantren" },
  ],
  cta: true,
};

export default function Page() { return <ContentPage data={data} />; }
