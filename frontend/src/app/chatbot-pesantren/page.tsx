import type { Metadata } from "next";
import { ContentPage, type ContentPageData } from "@/components/content-page";

export const metadata: Metadata = {
  title: { absolute: "Chatbot Pesantren: Jawab Pertanyaan Sesuai Dokumen Resmi | AHLAN" },
  description: "Apa itu chatbot pesantren, apa bedanya dengan chatbot umum, dan bagaimana AHLAN menjawab pertanyaan calon wali santri hanya dari dokumen resmi pesantren.",
  alternates: { canonical: "/chatbot-pesantren" },
};

const data: ContentPageData = {
  path: "/chatbot-pesantren",
  kicker: "Chatbot AI untuk pesantren",
  title: "Chatbot pesantren yang menjawab sesuai dokumen resmi",
  lead: "Chatbot pesantren adalah asisten percakapan yang menjawab pertanyaan calon wali santri secara otomatis. Yang membedakan AHLAN: jawabannya bersumber dari brosur, rincian biaya, jadwal, dan tata tertib yang Anda unggah sendiri.",
  sections: [
    {
      heading: "Masalah yang ingin diselesaikan",
      paragraphs: ["Pada musim penerimaan santri baru, pertanyaan yang masuk sebagian besar sama: berapa biayanya, apa syaratnya, kapan jadwalnya, bagaimana aturan di pondok. Admin menjawabnya berulang lewat WhatsApp atau telepon, dan jawabannya bisa berbeda dari satu admin ke admin lain."],
    },
    {
      heading: "Apa bedanya dengan chatbot umum?",
      paragraphs: ["Chatbot umum menjawab dari pengetahuan luas yang tidak spesifik untuk pesantren Anda, sehingga bisa keliru soal biaya atau aturan. AHLAN dirancang agar jawaban merujuk dokumen resmi pesantren, dan dokumen sumbernya dapat diperiksa pengurus."],
      list: [
        "Rujukan jawaban adalah dokumen yang Anda unggah, bukan pengetahuan umum.",
        "Pengurus dapat melihat percakapan dan sumber jawabannya di dashboard.",
        "Informasi yang berubah cukup diperbarui di dokumen, tidak perlu melatih ulang apa pun.",
      ],
    },
    {
      heading: "Apa saja yang bisa dijawab?",
      list: ["Biaya pendaftaran dan biaya bulanan sesuai brosur", "Syarat dan berkas pendaftaran", "Jadwal gelombang pendaftaran dan seleksi", "Tata tertib dan kegiatan harian santri", "Informasi dasar lain yang tertulis di dokumen resmi"],
    },
    {
      heading: "Batasan yang perlu Anda ketahui",
      paragraphs: ["Asisten hanya sebaik dokumen yang diberikan. Jika sebuah informasi tidak ada di dokumen, jawabannya tidak bisa diandalkan, sehingga pengurus perlu melengkapi dokumen dan meninjau percakapan secara berkala. AHLAN tidak menggantikan pengurus; ia mengurangi pekerjaan menjawab pertanyaan dasar agar pengurus bisa fokus pada calon wali yang butuh pendampingan."],
    },
  ],
  faqs: [
    { q: "Apakah chatbot pesantren bisa menjawab sesuai brosur?", a: "Ya, AHLAN menjawab berdasarkan dokumen yang diunggah seperti brosur PSB, rincian biaya, jadwal, dan tata tertib, dengan sumber yang dapat diperiksa pengurus." },
    { q: "Apakah pengurus bisa melihat percakapan?", a: "Ya. Seluruh percakapan dapat dipantau dari dashboard." },
    { q: "Siapa yang menyiapkan dokumennya?", a: "Pengurus pesantren mengunggah dokumen sendiri. Kami membantu penyiapan awal saat demo dan onboarding." },
  ],
  related: [
    { href: "/penerimaan-santri-baru", label: "Chatbot untuk penerimaan santri baru" },
    { href: "/chatbot-whatsapp-pesantren", label: "Chatbot WhatsApp & Telegram" },
    { href: "/panduan-psb-pesantren", label: "Panduan menyiapkan PSB" },
  ],
  cta: true,
};

export default function Page() { return <ContentPage data={data} />; }
