import type { Metadata } from "next";
import { ContentPage, type ContentPageData } from "@/components/content-page";

export const metadata: Metadata = {
  title: { absolute: "Kebijakan Privasi | AHLAN" },
  description: "Bagaimana AHLAN mengumpulkan, menggunakan, dan melindungi data pengurus pesantren dan percakapan calon wali santri.",
  alternates: { canonical: "/kebijakan-privasi" },
};

// DRAFT: tinjau bersama pemilik produk/penasihat hukum sebelum dipublikasikan, terutama bagian penyedia AI, retensi, dan kontak.
const data: ContentPageData = {
  path: "/kebijakan-privasi",
  kicker: "Privasi",
  title: "Kebijakan privasi",
  lead: "Halaman ini menjelaskan data apa yang diproses AHLAN, untuk apa, dan hak Anda atas data tersebut.",
  updated: "November 2026",
  sections: [
    {
      heading: "Data yang kami proses",
      list: [
        "Data akun pengurus: nama, alamat email atau identitas login, dan peran.",
        "Dokumen yang diunggah pesantren, seperti brosur, rincian biaya, jadwal, dan tata tertib.",
        "Percakapan melalui WhatsApp atau Telegram: pesan yang dikirim calon wali santri, jawaban asisten, serta identitas akun (nomor WhatsApp atau akun Telegram) yang tersedia dari platform tersebut.",
        "Jejak aktivitas: catatan perubahan penting di ruang kerja pesantren.",
      ],
    },
    {
      heading: "Untuk apa data digunakan",
      list: [
        "Menjalankan layanan: menjawab pertanyaan berdasarkan dokumen pesantren dan menampilkan percakapan kepada pengurus.",
        "Menjaga keamanan dan menelusuri perubahan melalui jejak aktivitas.",
        "Memberikan dukungan teknis kepada pesantren.",
      ],
      paragraphs: ["Data pesantren tidak kami jual kepada pihak ketiga."],
    },
    {
      heading: "Pemrosesan oleh pihak ketiga",
      paragraphs: ["Untuk menghasilkan jawaban, isi pertanyaan dan potongan dokumen yang relevan dikirim ke penyedia model AI pihak ketiga. Percakapan juga melewati platform WhatsApp (Meta) atau Telegram yang tunduk pada kebijakan privasinya masing-masing. Kami juga memakai penyedia infrastruktur hosting untuk menjalankan layanan."],
    },
    {
      heading: "Pemisahan dan akses",
      paragraphs: ["Setiap pesantren memiliki ruang kerja sendiri, dan akses pengurus diatur berdasarkan peran. Tidak ada sistem yang sepenuhnya bebas risiko; jika Anda menemukan masalah keamanan, mohon segera hubungi kami."],
    },
    {
      heading: "Peran pesantren sebagai pengelola data",
      paragraphs: ["Pesantren menentukan dokumen apa yang diunggah dan bagaimana layanan digunakan. Kami menyarankan pesantren memberi tahu calon wali santri bahwa percakapan mereka dapat dilihat pengurus, serta tidak meminta data sensitif melalui chatbot."],
    },
    {
      heading: "Penyimpanan dan penghapusan",
      paragraphs: ["Data disimpan selama pesantren menggunakan layanan. Pesantren atau calon wali santri dapat meminta penghapusan data tertentu dengan menghubungi kami."],
    },
    {
      heading: "Kontak",
      paragraphs: ["Pertanyaan tentang privasi dapat disampaikan melalui kontak resmi yang tercantum di halaman utama AHLAN."],
    },
  ],
  related: [{ href: "/", label: "Kembali ke beranda" }],
};

export default function Page() { return <ContentPage data={data} />; }
