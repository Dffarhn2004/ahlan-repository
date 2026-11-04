import type { Metadata } from "next";
import { ContentPage, type ContentPageData } from "@/components/content-page";

export const metadata: Metadata = {
  title: { absolute: "Chatbot untuk Penerimaan Santri Baru (PSB/PPDB) Pesantren | AHLAN" },
  description: "Bantu penerimaan santri baru dengan chatbot yang menjawab pertanyaan biaya, syarat, dan jadwal PSB secara konsisten via WhatsApp dan Telegram, berdasarkan brosur resmi pesantren.",
  alternates: { canonical: "/penerimaan-santri-baru" },
};

const data: ContentPageData = {
  path: "/penerimaan-santri-baru",
  kicker: "Penerimaan santri baru",
  title: "Chatbot untuk penerimaan santri baru (PSB/PPDB) pesantren",
  lead: "Saat PSB dibuka, pertanyaan calon wali santri datang tanpa henti dan sering di luar jam kerja. AHLAN membantu menjawab pertanyaan dasar secara konsisten lewat WhatsApp dan Telegram, sementara pengurus memantau dan menindaklanjuti.",
  sections: [
    {
      heading: "Titik-titik yang membuat PSB melelahkan",
      list: [
        "Pertanyaan yang sama diulang puluhan kali: biaya, syarat, jadwal, tata tertib.",
        "Jawaban admin, brosur, dan grup tidak selalu seragam.",
        "Pertanyaan yang masuk malam hari baru dijawab keesokan harinya.",
        "Sulit mengetahui pertanyaan apa yang paling sering muncul.",
      ],
    },
    {
      heading: "Bagaimana AHLAN membantu",
      list: [
        "Calon wali bertanya lewat WhatsApp atau Telegram dan mendapat jawaban saat itu juga.",
        "Jawaban merujuk dokumen resmi, sehingga seragam dengan brosur.",
        "Semua percakapan tercatat di dashboard untuk dipantau dan ditindaklanjuti.",
        "Dari percakapan, pengurus melihat pertanyaan yang sering muncul untuk memperbaiki brosur dan sosialisasi.",
      ],
    },
    {
      heading: "Alur kerja selama masa PSB",
      ordered: true,
      list: [
        "Unggah brosur PSB, rincian biaya, jadwal, dan tata tertib.",
        "Bagikan nomor WhatsApp atau tautan bot Telegram melalui brosur, media sosial, atau grup.",
        "Calon wali bertanya, asisten menjawab dari dokumen.",
        "Pengurus meninjau percakapan dan menghubungi calon wali yang perlu pendampingan.",
        "Perbarui dokumen jika ada perubahan jadwal atau biaya.",
      ],
    },
    {
      heading: "Yang tidak dilakukan chatbot",
      paragraphs: ["Chatbot menjawab pertanyaan informasi. Keputusan penerimaan, pembayaran, dan komunikasi personal tetap ditangani pengurus."],
    },
  ],
  faqs: [
    { q: "Bagaimana chatbot membantu penerimaan santri baru?", a: "Chatbot menjawab pertanyaan berulang tentang biaya, syarat, jadwal, dan tata tertib kapan saja lewat WhatsApp atau Telegram, sehingga pengurus punya lebih banyak waktu untuk follow-up calon wali santri." },
    { q: "Apakah cocok untuk pesantren kecil?", a: "Cocok jika pesantren sudah memiliki brosur atau dokumen informasi PSB dan calon wali sudah terbiasa memakai WhatsApp atau Telegram." },
  ],
  related: [
    { href: "/panduan-psb-pesantren", label: "Panduan menyiapkan PSB" },
    { href: "/chatbot-pesantren", label: "Apa itu chatbot pesantren" },
  ],
  cta: true,
};

export default function Page() { return <ContentPage data={data} />; }
