import type { Metadata } from "next";
import { ContentPage, type ContentPageData } from "@/components/content-page";

export const metadata: Metadata = {
  title: { absolute: "Chatbot WhatsApp & Telegram untuk Pesantren | AHLAN" },
  description: "Layani calon wali santri lewat WhatsApp dan Telegram: asisten menjawab dari dokumen resmi pesantren, pengurus memantau semua percakapan dari satu dashboard.",
  alternates: { canonical: "/chatbot-whatsapp-pesantren" },
};

const data: ContentPageData = {
  path: "/chatbot-whatsapp-pesantren",
  kicker: "Kanal WhatsApp & Telegram",
  title: "Chatbot WhatsApp dan Telegram untuk pesantren",
  lead: "Calon wali santri cukup bertanya lewat aplikasi yang sudah mereka pakai. AHLAN menjawab dari dokumen resmi pesantren, dan pengurus memantau semua percakapan dari satu dashboard.",
  sections: [
    {
      heading: "Di mana calon wali santri bertanya?",
      paragraphs: ["Bagi sebagian besar keluarga di Indonesia, WhatsApp adalah kanal komunikasi utama, sedangkan Telegram dipilih oleh sebagian lainnya. AHLAN mendukung keduanya, sehingga pesantren dapat hadir di tempat calon wali sudah berada."],
      list: [
        "WhatsApp: pesantren menghubungkan nomor WhatsApp-nya ke ruang kerja AHLAN dengan memindai kode QR dari dashboard.",
        "Telegram: pesantren menghubungkan bot Telegram ke ruang kerja yang sama.",
        "Keduanya memakai dokumen resmi yang sama, sehingga jawabannya seragam.",
      ],
    },
    {
      heading: "Cara kerjanya",
      ordered: true,
      list: [
        "Pesantren memiliki ruang kerja AHLAN dan mengunggah dokumen resmi.",
        "Pengurus menghubungkan WhatsApp dan/atau Telegram dari dashboard.",
        "Calon wali santri bertanya lewat kanal yang mereka pilih.",
        "Asisten menjawab berdasarkan dokumen; pengurus melihat percakapan di dashboard.",
      ],
    },
    {
      heading: "Mulai dengan aman: mode uji coba",
      paragraphs: ["Pada WhatsApp, pesantren dapat memulai dari mode uji coba yang hanya melayani nomor tertentu, mencoba kualitas jawaban bersama tim, lalu membuka layanan untuk umum ketika sudah yakin."],
    },
    {
      heading: "Hal yang perlu dipertimbangkan",
      list: [
        "Gunakan nomor khusus layanan PSB, bukan nomor pribadi pengurus.",
        "Ketersediaan dan perilaku WhatsApp ditentukan oleh platform WhatsApp itu sendiri; tanyakan detailnya saat demo.",
        "Asisten menjawab informasi umum PSB. Pembicaraan personal dan keputusan penerimaan tetap oleh pengurus.",
      ],
    },
  ],
  faqs: [
    { q: "Apakah chatbot bisa lewat WhatsApp?", a: "Ya. Pesantren dapat menghubungkan nomor WhatsApp dengan memindai kode QR dari dashboard AHLAN, dan memulainya dengan mode uji coba untuk nomor tertentu." },
    { q: "Apakah masih bisa memakai Telegram?", a: "Ya. Telegram tetap didukung, dan WhatsApp serta Telegram dapat dipakai bersamaan." },
    { q: "Apakah pengurus bisa melihat percakapan?", a: "Ya. Percakapan dari kedua kanal dapat dipantau dari dashboard." },
  ],
  related: [
    { href: "/chatbot-pesantren", label: "Apa itu chatbot pesantren" },
    { href: "/penerimaan-santri-baru", label: "Chatbot untuk PSB" },
  ],
  cta: true,
};

export default function Page() { return <ContentPage data={data} />; }
