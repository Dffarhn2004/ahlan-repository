import type { Metadata } from "next";
import {
  ArrowUpRight,
  BarChart3,
  Bot,
  Check,
  CheckCircle2,
  ChevronDown,
  Clock3,
  FileCheck2,
  FileText,
  LayoutDashboard,
  MessageCircle,
  MessagesSquare,
  Send,
  ShieldCheck,
  Sparkles,
  Users,
  type LucideIcon,
} from "lucide-react";
import { DemoButton, SiteFooter, SiteHeader } from "@/components/landing-ui";

export const metadata: Metadata = {
  title: { absolute: "Chatbot AI untuk Penerimaan Santri | AHLAN Pesantren" },
  alternates: { canonical: "/" },
};

const benefits: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: Clock3, title: "Aktif 24 jam", text: "Pertanyaan calon wali tetap terjawab, bahkan setelah jam kerja pengurus selesai." },
  { icon: FileCheck2, title: "Jawaban terverifikasi", text: "Setiap jawaban merujuk pada brosur, jadwal, biaya, dan dokumen resmi Anda." },
  { icon: Users, title: "Admin lebih fokus", text: "Tim bisa memprioritaskan calon wali yang benar-benar membutuhkan pendampingan." },
];

const features: { icon: LucideIcon; title: string; text: string; label: string }[] = [
  { icon: MessagesSquare, title: "Satu pusat percakapan", text: "WhatsApp dan Telegram terpantau dalam satu dashboard yang rapi.", label: "Inbox terpadu" },
  { icon: FileText, title: "Pengetahuan dari dokumen", text: "Unggah berkas resmi dan jadikan sebagai sumber jawaban asisten.", label: "Knowledge base" },
  { icon: BarChart3, title: "Insight pertanyaan", text: "Temukan topik yang paling sering ditanyakan untuk memperbaiki informasi PSB.", label: "Analitik" },
  { icon: ShieldCheck, title: "Akses yang aman", text: "Ruang kerja terpisah, akses berbasis peran, dan aktivitas yang dapat ditelusuri.", label: "Keamanan" },
];

const steps = [
  { number: "01", title: "Unggah dokumen", text: "Masukkan brosur PSB, rincian biaya, jadwal, dan tata tertib." },
  { number: "02", title: "Hubungkan kanal", text: "Sambungkan WhatsApp atau Telegram yang digunakan pesantren." },
  { number: "03", title: "Mulai melayani", text: "Asisten menjawab, sementara pengurus memantau dari dashboard." },
];

const faqs = [
  { q: "Apakah jawaban AHLAN sesuai brosur pesantren?", a: "Ya. AHLAN menggunakan dokumen resmi yang Anda unggah sebagai sumber jawaban. Sumber tersebut juga dapat diperiksa oleh pengurus." },
  { q: "Apa yang terjadi jika jawaban tidak ada di dokumen?", a: "Pertanyaan tetap tercatat di dashboard agar pengurus dapat menindaklanjuti dan melengkapi informasi yang belum tersedia." },
  { q: "Kanal apa saja yang bisa digunakan?", a: "AHLAN dirancang untuk melayani percakapan melalui WhatsApp dan Telegram, dengan seluruh percakapan terpantau dalam satu dashboard." },
  { q: "Apakah admin perlu mengganti cara kerja?", a: "Tidak. Pengurus cukup menyiapkan dokumen yang sudah digunakan dan menghubungkan kanal pesan. Tim kami membantu proses penyiapannya." },
  { q: "Bagaimana cara memulai?", a: "Jadwalkan demo singkat selama 20 menit. Kami akan menunjukkan alur produk dan membantu menilai kesiapan dokumen pesantren Anda." },
];

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: faqs.map(({ q, a }) => ({ "@type": "Question", name: q, acceptedAnswer: { "@type": "Answer", text: a } })),
};

const eyebrow = "text-xs font-bold uppercase tracking-[.18em] text-emerald-600";
const sectionTitle = "text-balance text-[clamp(2rem,4.5vw,3.5rem)] font-bold leading-[1.05] tracking-[-.055em] text-slate-950";

function ProductPreview() {
  return (
    <div className="relative mx-auto w-full max-w-[630px]">
      <div className="absolute -inset-8 -z-10 rounded-[3rem] bg-emerald-300/20 blur-3xl" />
      <div className="overflow-hidden rounded-[1.4rem] border border-slate-200 bg-white shadow-[0_32px_80px_rgba(15,23,42,.16)]">
        <div className="flex h-11 items-center border-b border-slate-200 bg-slate-50/80 px-4">
          <div className="flex gap-1.5"><span className="size-2.5 rounded-full bg-rose-300" /><span className="size-2.5 rounded-full bg-amber-300" /><span className="size-2.5 rounded-full bg-emerald-300" /></div>
          <div className="mx-auto h-6 w-48 rounded-md border border-slate-200 bg-white" />
        </div>
        <div className="grid min-h-[390px] grid-cols-[62px_1fr] sm:grid-cols-[155px_1fr]">
          <aside className="border-r border-slate-200 bg-slate-950 p-3 text-white">
            <div className="mb-7 flex items-center gap-2 px-1"><span className="grid size-7 place-items-center rounded-lg bg-emerald-500"><Bot size={15} /></span><b className="hidden text-xs tracking-tight sm:block">AHLAN</b></div>
            {[LayoutDashboard, MessagesSquare, FileText, BarChart3].map((Icon, index) => (
              <div key={index} className={`mb-2 flex h-9 items-center gap-2 rounded-lg px-2 text-[10px] ${index === 1 ? "bg-white/10 text-white" : "text-slate-500"}`}><Icon size={14} /><span className="hidden sm:block">{["Ringkasan", "Percakapan", "Dokumen", "Analitik"][index]}</span></div>
            ))}
          </aside>
          <div className="min-w-0 bg-slate-50/60 p-4 sm:p-5">
            <div className="mb-5 flex items-start justify-between gap-3">
              <div><p className="text-[10px] font-semibold uppercase tracking-wider text-emerald-600">Percakapan hari ini</p><h3 className="mt-1 text-base font-bold tracking-tight text-slate-900">Pusat pesan masuk</h3></div>
              <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-[9px] font-bold text-emerald-700"><i className="size-1.5 rounded-full bg-emerald-500" /> Bot aktif</span>
            </div>
            <div className="grid gap-3 lg:grid-cols-[.82fr_1.18fr]">
              <div className="space-y-2">
                {["Ahmad Fauzan", "Ibu Nur Aini", "Bapak Ridwan"].map((name, index) => (
                  <div key={name} className={`rounded-xl border p-3 ${index === 0 ? "border-emerald-200 bg-white shadow-sm" : "border-slate-200 bg-white/60"}`}>
                    <div className="flex items-center gap-2"><span className="grid size-7 shrink-0 place-items-center rounded-full bg-emerald-100 text-[9px] font-bold text-emerald-700">{name.split(" ").map((word) => word[0]).join("").slice(0, 2)}</span><div className="min-w-0"><b className="block truncate text-[10px] text-slate-800">{name}</b><span className="block truncate text-[8px] text-slate-400">{index === 0 ? "Berapa biaya pendaftaran?" : "Informasi PSB terbaru..."}</span></div></div>
                  </div>
                ))}
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3 shadow-sm">
                <div className="border-b border-slate-100 pb-2.5"><b className="text-[10px] text-slate-800">Ahmad Fauzan</b><p className="text-[8px] text-slate-400">WhatsApp · sedang dilayani AI</p></div>
                <div className="space-y-2 py-3 text-[9px] leading-relaxed">
                  <p className="ml-auto max-w-[85%] rounded-lg rounded-br-sm bg-slate-100 px-2.5 py-2 text-slate-700">Assalamu&apos;alaikum, berapa biaya pendaftaran santri baru?</p>
                  <div className="max-w-[92%] rounded-lg rounded-bl-sm bg-emerald-50 px-2.5 py-2 text-emerald-950">Wa&apos;alaikumussalam. Biaya pendaftaran sebesar Rp350.000.<span className="mt-2 flex items-center gap-1 rounded-md bg-white/80 px-1.5 py-1 text-[7px] font-semibold text-emerald-700"><FileCheck2 size={9} /> Brosur PSB 2026.pdf</span></div>
                </div>
                <div className="flex h-8 items-center rounded-lg border border-slate-200 px-2 text-[8px] text-slate-400">Tulis balasan...<Send size={11} className="ml-auto text-emerald-600" /></div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div className="absolute -bottom-5 -left-3 hidden items-center gap-3 rounded-2xl border border-slate-200 bg-white p-3 shadow-xl sm:flex">
        <span className="grid size-9 place-items-center rounded-xl bg-emerald-100 text-emerald-700"><CheckCircle2 size={18} /></span>
        <div><b className="block text-xs text-slate-900">Jawaban terkirim</b><span className="text-[10px] text-slate-400">Berdasarkan dokumen resmi</span></div>
      </div>
    </div>
  );
}

function ConversationPreview() {
  return (
    <div className="relative mx-auto max-w-md rounded-[2rem] border border-slate-200 bg-white p-3 shadow-[0_24px_60px_rgba(15,23,42,.12)]">
      <div className="rounded-[1.35rem] bg-slate-50 p-5">
        <div className="flex items-center gap-3 border-b border-slate-200 pb-4"><span className="grid size-10 place-items-center rounded-xl bg-emerald-600 text-white"><Bot size={20} /></span><div><b className="block text-sm text-slate-900">Asisten PSB</b><span className="flex items-center gap-1.5 text-[11px] text-emerald-600"><i className="size-1.5 rounded-full bg-emerald-500" /> Online</span></div></div>
        <div className="space-y-3 py-5 text-sm leading-relaxed">
          <p className="ml-auto max-w-[84%] rounded-2xl rounded-br-sm bg-slate-900 px-4 py-3 text-white">Apa saja berkas yang perlu disiapkan?</p>
          <div className="max-w-[92%] rounded-2xl rounded-bl-sm border border-emerald-100 bg-emerald-50 px-4 py-3 text-slate-700">Siapkan formulir pendaftaran, fotokopi KK, akta kelahiran, dan pas foto.<span className="mt-3 flex items-center gap-2 rounded-lg border border-emerald-100 bg-white px-2.5 py-2 text-[11px] font-semibold text-emerald-700"><FileCheck2 size={14} /> Sumber: Brosur PSB.pdf</span></div>
        </div>
        <div className="flex h-11 items-center rounded-xl border border-slate-200 bg-white px-3 text-xs text-slate-400">Ketik pertanyaan...<span className="ml-auto grid size-7 place-items-center rounded-lg bg-emerald-600 text-white"><Send size={13} /></span></div>
      </div>
    </div>
  );
}

export default function Landing() {
  return (
    <div className="min-h-screen overflow-x-clip bg-white text-slate-900">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c") }} />
      <SiteHeader home />

      <main>
        <section className="relative isolate overflow-hidden pb-20 pt-16 sm:pt-20 lg:pb-28 lg:pt-24">
          <div className="absolute inset-0 -z-20 bg-[linear-gradient(to_bottom,#f0fdf8_0%,#fff_72%)]" />
          <div className="absolute left-1/2 top-0 -z-10 h-[600px] w-[900px] -translate-x-1/2 rounded-full bg-emerald-200/30 blur-[120px]" />
          <div className="mx-auto w-full max-w-[1212px] px-4">
            <div className="mx-auto max-w-4xl text-center">
              <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-white/80 px-3 py-1.5 text-xs font-semibold text-emerald-700 shadow-sm backdrop-blur"><Sparkles size={13} /> AI khusus penerimaan santri baru</div>
              <h1 className="text-balance text-[clamp(2.7rem,7vw,5.35rem)] font-bold leading-[.98] tracking-[-.065em] text-slate-950">Layani calon wali santri, <span className="text-emerald-600">tanpa chat berulang.</span></h1>
              <p className="mx-auto mt-7 max-w-2xl text-pretty text-base leading-7 text-slate-600 sm:text-lg">AHLAN menjawab pertanyaan biaya, syarat, jadwal, dan tata tertib melalui WhatsApp dan Telegram—langsung dari dokumen resmi pesantren Anda.</p>
              <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row"><DemoButton>Jadwalkan demo 20 menit</DemoButton><a href="#solusi" className="group inline-flex h-12 items-center gap-2 rounded-xl px-5 text-sm font-semibold text-slate-700 hover:bg-white">Lihat cara kerjanya <ArrowUpRight size={16} className="transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" /></a></div>
              <p className="mt-4 text-xs text-slate-400">Tanpa instalasi rumit · Dibantu sampai siap digunakan</p>
            </div>
            <div className="mt-16 lg:mt-20"><ProductPreview /></div>
          </div>
        </section>

        <section className="border-y border-slate-200 bg-slate-50/70 py-7">
          <div className="mx-auto flex w-full max-w-[1032px] flex-col items-center justify-between gap-5 px-4 text-center sm:flex-row sm:text-left">
            <p className="text-sm font-semibold text-slate-500">Dibuat untuk alur kerja penerimaan santri</p>
            <div className="flex flex-wrap justify-center gap-x-7 gap-y-3 text-xs font-bold text-slate-400"><span className="flex items-center gap-2"><Check size={14} className="text-emerald-600" /> WhatsApp</span><span className="flex items-center gap-2"><Check size={14} className="text-emerald-600" /> Telegram</span><span className="flex items-center gap-2"><Check size={14} className="text-emerald-600" /> Dokumen resmi</span><span className="flex items-center gap-2"><Check size={14} className="text-emerald-600" /> Dashboard terpusat</span></div>
          </div>
        </section>

        <section id="solusi" className="scroll-mt-24 py-20 sm:py-28">
          <div className="mx-auto w-full max-w-[1212px] px-4">
            <div className="grid gap-8 lg:grid-cols-[.85fr_1.15fr] lg:items-end"><div><p className={eyebrow}>Dampak yang terasa</p><h2 className={`${sectionTitle} mt-4`}>Lebih cepat melayani. Lebih tenang mengelola.</h2></div><p className="max-w-xl text-base leading-7 text-slate-500 lg:ml-auto">AHLAN mengambil pekerjaan repetitif dari meja admin, tanpa menghilangkan kendali pengurus atas informasi yang disampaikan.</p></div>
            <div className="mt-12 grid gap-4 md:grid-cols-3">
              {benefits.map(({ icon: Icon, title, text }, index) => <article key={title} className="group rounded-2xl border border-slate-200 bg-white p-6 transition-all hover:-translate-y-1 hover:border-emerald-200 hover:shadow-xl hover:shadow-emerald-900/5"><div className="flex items-center justify-between"><span className="grid size-11 place-items-center rounded-xl bg-emerald-50 text-emerald-700"><Icon size={20} /></span><span className="text-xs font-bold text-slate-300">0{index + 1}</span></div><h3 className="mt-6 text-lg font-bold tracking-tight text-slate-900">{title}</h3><p className="mt-2 text-sm leading-6 text-slate-500">{text}</p></article>)}
            </div>
          </div>
        </section>

        <section id="fitur" className="scroll-mt-24 bg-slate-950 py-20 text-white sm:py-28">
          <div className="mx-auto grid w-full max-w-[1212px] items-center gap-14 px-4 lg:grid-cols-[1fr_.9fr]">
            <div><p className="text-xs font-bold uppercase tracking-[.18em] text-emerald-400">Dirancang agar tetap akurat</p><h2 className="mt-4 text-balance text-[clamp(2rem,4.5vw,3.6rem)] font-bold leading-[1.04] tracking-[-.055em]">Bukan sekadar chatbot. Ini asisten PSB Anda.</h2><p className="mt-5 max-w-xl leading-7 text-slate-400">AHLAN tidak menjawab dari asumsi. Ia menggunakan sumber yang Anda tentukan, menampilkan rujukan, dan mencatat percakapan agar pengurus tetap punya kendali.</p><ul className="mt-8 space-y-4">{["Jawaban bersumber dari dokumen pesantren", "Sumber jawaban dapat diperiksa", "Pertanyaan yang belum terjawab tetap tercatat"].map((item) => <li key={item} className="flex items-center gap-3 text-sm text-slate-300"><span className="grid size-6 place-items-center rounded-full bg-emerald-500/15 text-emerald-400"><Check size={14} /></span>{item}</li>)}</ul></div>
            <ConversationPreview />
          </div>
        </section>

        <section className="py-20 sm:py-28">
          <div className="mx-auto w-full max-w-[1212px] px-4"><div className="mx-auto max-w-2xl text-center"><p className={eyebrow}>Semua dalam satu tempat</p><h2 className={`${sectionTitle} mt-4`}>Yang dibutuhkan pengurus, tanpa kerumitan.</h2></div><div className="mt-12 grid gap-4 sm:grid-cols-2">{features.map(({ icon: Icon, title, text, label }) => <article key={title} className="relative overflow-hidden rounded-2xl border border-slate-200 bg-slate-50/50 p-6 sm:p-8"><span className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-slate-500"><Icon size={12} className="text-emerald-600" />{label}</span><h3 className="mt-7 text-xl font-bold tracking-tight text-slate-950">{title}</h3><p className="mt-2 max-w-md text-sm leading-6 text-slate-500">{text}</p><div className="absolute -bottom-8 -right-8 size-28 rounded-full border-[18px] border-emerald-100/70" /></article>)}</div></div>
        </section>

        <section id="cara-kerja" className="scroll-mt-24 bg-emerald-50 py-20 sm:py-28">
          <div className="mx-auto w-full max-w-[1212px] px-4"><div className="max-w-2xl"><p className={eyebrow}>Mulai dengan mudah</p><h2 className={`${sectionTitle} mt-4`}>Dari brosur ke layanan otomatis dalam tiga langkah.</h2></div><ol className="mt-12 grid gap-4 md:grid-cols-3">{steps.map((step, index) => <li key={step.number} className="relative rounded-2xl border border-emerald-100 bg-white p-6 shadow-sm"><span className="text-4xl font-bold tracking-[-.06em] text-emerald-200">{step.number}</span><h3 className="mt-8 text-lg font-bold text-slate-950">{step.title}</h3><p className="mt-2 text-sm leading-6 text-slate-500">{step.text}</p>{index < 2 && <ArrowUpRight className="absolute right-5 top-5 hidden text-emerald-300 md:block" size={18} />}</li>)}</ol><div className="mt-9"><DemoButton>Mulai dengan demo singkat</DemoButton></div></div>
        </section>

        <section id="faq" className="scroll-mt-24 py-20 sm:py-28">
          <div className="mx-auto grid w-full max-w-[1112px] px-4 gap-10 lg:grid-cols-[.72fr_1.28fr]"><div><p className={eyebrow}>Pertanyaan umum</p><h2 className={`${sectionTitle} mt-4`}>Yang perlu Anda tahu.</h2><p className="mt-5 text-sm leading-6 text-slate-500">Belum menemukan jawabannya? Jadwalkan demo dan bicarakan kebutuhan pesantren Anda.</p></div><div className="border-t border-slate-200">{faqs.map(({ q, a }) => <details key={q} className="group border-b border-slate-200"><summary className="flex cursor-pointer list-none items-center justify-between gap-4 py-5 text-left text-sm font-semibold text-slate-900 [&::-webkit-details-marker]:hidden sm:text-base">{q}<span className="grid size-8 shrink-0 place-items-center rounded-full bg-slate-100"><ChevronDown size={16} className="transition-transform group-open:rotate-180" /></span></summary><p className="max-w-2xl pb-5 pr-10 text-sm leading-6 text-slate-500">{a}</p></details>)}</div></div>
        </section>

        <section className="mx-auto mb-20 w-full max-w-[1212px] px-4">
          <div className="relative isolate overflow-hidden rounded-[2rem] bg-emerald-600 px-6 py-16 text-center text-white sm:px-12 sm:py-20"><div className="absolute -left-20 -top-28 -z-10 size-80 rounded-full border-[54px] border-white/5" /><div className="absolute -bottom-32 -right-16 -z-10 size-96 rounded-full border-[64px] border-white/5" /><div className="mx-auto grid size-12 place-items-center rounded-2xl bg-white/15"><MessageCircle size={22} /></div><h2 className="mx-auto mt-6 max-w-3xl text-balance text-[clamp(2rem,5vw,3.6rem)] font-bold leading-[1.05] tracking-[-.055em]">Siap membuat musim PSB lebih ringan?</h2><p className="mx-auto mt-4 max-w-xl text-sm leading-6 text-emerald-50 sm:text-base">Lihat bagaimana AHLAN menjawab pertanyaan calon wali langsung dari brosur pesantren Anda.</p><div className="mt-8"><DemoButton variant="light">Jadwalkan demo 20 menit</DemoButton></div></div>
        </section>
      </main>

      <SiteFooter />
    </div>
  );
}
