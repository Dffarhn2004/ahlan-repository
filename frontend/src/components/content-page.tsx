import Link from "next/link";
import { ArrowLeft, ArrowUpRight, Check, ChevronDown, Clock3, FileText } from "lucide-react";
import { DemoButton, SiteFooter, SiteHeader } from "@/components/landing-ui";
import { siteUrl } from "@/lib/seo";

export type ContentSection = { heading: string; paragraphs?: string[]; list?: string[]; ordered?: boolean };
export type ContentFaq = { q: string; a: string };
export type ContentPageData = {
  path: string;
  kicker: string;
  title: string;
  lead: string;
  sections: ContentSection[];
  faqs?: ContentFaq[];
  related?: { href: string; label: string }[];
  /** Tampilkan CTA demo di akhir halaman. */
  cta?: boolean;
  updated?: string;
};

function sectionId(heading: string) {
  return heading.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "");
}

export function ContentPage({ data }: { data: ContentPageData }) {
  const jsonLd = [
    { "@context": "https://schema.org", "@type": "WebPage", name: data.title, description: data.lead, url: `${siteUrl}${data.path}`, inLanguage: "id-ID", isPartOf: { "@id": `${siteUrl}/#website` } },
    ...(data.faqs?.length
      ? [{ "@context": "https://schema.org", "@type": "FAQPage", mainEntity: data.faqs.map(({ q, a }) => ({ "@type": "Question", name: q, acceptedAnswer: { "@type": "Answer", text: a } })) }]
      : []),
  ];

  return (
    <div className="min-h-screen overflow-x-clip bg-white text-slate-900">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c") }} />
      <SiteHeader />

      <main>
        <header className="relative isolate overflow-hidden border-b border-slate-200 bg-[linear-gradient(135deg,#ecfdf5_0%,#fff_58%,#f8fafc_100%)]">
          <div className="absolute -left-28 -top-36 -z-10 size-[430px] rounded-full border-[70px] border-emerald-100/70" />
          <div className="absolute right-[8%] top-10 -z-10 size-72 rounded-full bg-emerald-100/40 blur-3xl" />
          <div className="mx-auto w-full max-w-[1212px] px-4 py-14 sm:py-20 lg:py-24">
            <Link href="/" className="group inline-flex items-center gap-2 text-xs font-semibold text-slate-500 transition-colors hover:text-emerald-700">
              <ArrowLeft size={14} className="transition-transform group-hover:-translate-x-1" /> Kembali ke beranda
            </Link>
            <h1 className="mt-8max-w-4xl text-balance text-[clamp(2.35rem,6vw,4.65rem)] font-bold leading-[1.02] tracking-[-.06em] text-slate-950">{data.title}</h1>
            <p className="mt-6 max-w-3xl text-pretty text-base leading-7 text-slate-600 sm:text-lg sm:leading-8">{data.lead}</p>
            <div className="mt-7 flex flex-wrap items-center gap-x-6 gap-y-3 text-xs font-medium text-slate-400">
              <span className="inline-flex items-center gap-2"><FileText size={14} className="text-emerald-600" /> Panduan AHLAN</span>
              {data.updated && <span className="inline-flex items-center gap-2"><Clock3 size={14} className="text-emerald-600" /> Diperbarui {data.updated}</span>}
            </div>
          </div>
        </header>

        <div className="mx-auto flex w-full max-w-[1212px] gap-12 px-4 py-16 lg:py-20">
          <aside className="hidden w-[220px] shrink-0 lg:block">
            <div className="sticky top-28">
              <p className="text-[11px] font-bold uppercase tracking-[.16em] text-slate-400">Di halaman ini</p>
              <nav aria-label="Daftar isi" className="mt-4 border-l border-slate-200">
                {data.sections.map((section, index) => (
                  <a key={section.heading} href={`#${sectionId(section.heading)}`} className="block border-l-2 border-transparent py-2 pl-4 text-sm leading-5 text-slate-500 transition-colors hover:border-emerald-500 hover:text-emerald-700">
                    <span className="mr-2 text-[10px] font-bold text-slate-300">0{index + 1}</span>{section.heading}
                  </a>
                ))}
                {data.faqs && <a href="#pertanyaan-umum" className="block border-l-2 border-transparent py-2 pl-4 text-sm text-slate-500 hover:border-emerald-500 hover:text-emerald-700">FAQ</a>}
              </nav>
            </div>
          </aside>

          <article className="min-w-0 flex-1">
            <div className="space-y-5">
              {data.sections.map((section, sectionIndex) => (
                <section id={sectionId(section.heading)} key={section.heading} className="scroll-mt-28 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm shadow-slate-900/[.025] sm:p-8">
                  <div className="flex items-start gap-4">
                    <span className="grid size-9 shrink-0 place-items-center rounded-xl bg-emerald-50 text-xs font-extrabold text-emerald-700">0{sectionIndex + 1}</span>
                    <h2 className="pt-1 text-balance text-xl font-bold leading-snug tracking-[-.035em] text-slate-950 sm:text-2xl">{section.heading}</h2>
                  </div>
                  <div className="mt-5 sm:pl-[52px]">
                    {section.paragraphs?.map((paragraph) => <p key={paragraph} className="mt-3 text-[15px] leading-7 text-slate-600 first:mt-0">{paragraph}</p>)}
                    {section.list && (section.ordered ? (
                      <ol className="mt-5 space-y-3">
                        {section.list.map((item, index) => <li key={item} className="flex gap-3 text-[15px] leading-7 text-slate-600"><span className="mt-1 grid size-5 shrink-0 place-items-center rounded-full bg-slate-900 text-[9px] font-bold text-white">{index + 1}</span><span>{item}</span></li>)}
                      </ol>
                    ) : (
                      <ul className="mt-5 space-y-3">
                        {section.list.map((item) => <li key={item} className="flex gap-3 text-[15px] leading-7 text-slate-600"><span className="mt-1 grid size-5 shrink-0 place-items-center rounded-full bg-emerald-50 text-emerald-700"><Check size={12} strokeWidth={3} /></span><span>{item}</span></li>)}
                      </ul>
                    ))}
                  </div>
                </section>
              ))}
            </div>

            {data.faqs && (
              <section id="pertanyaan-umum" className="mt-16 scroll-mt-28">
                <p className="text-xs font-bold uppercase tracking-[.16em] text-emerald-600">FAQ</p>
                <h2 className="mt-3 text-3xl font-bold tracking-[-.05em] text-slate-950">Pertanyaan yang sering diajukan</h2>
                <div className="mt-7 overflow-hidden rounded-2xl border border-slate-200 bg-white">
                  {data.faqs.map(({ q, a }) => (
                    <details key={q} className="group border-b border-slate-200 last:border-b-0">
                      <summary className="flex cursor-pointer list-none items-center justify-between gap-4 px-5 py-5 text-left text-sm font-semibold text-slate-900 [&::-webkit-details-marker]:hidden sm:px-6 sm:text-base">{q}<span className="grid size-8 shrink-0 place-items-center rounded-full bg-slate-100"><ChevronDown size={16} className="transition-transform group-open:rotate-180" /></span></summary>
                      <p className="max-w-2xl px-5 pb-5 pr-12 text-sm leading-7 text-slate-500 sm:px-6">{a}</p>
                    </details>
                  ))}
                </div>
              </section>
            )}

            {data.cta && (
              <aside className="relative isolate mt-16 overflow-hidden rounded-[2rem] bg-slate-950 p-7 text-white sm:p-10">
                <div className="absolute -bottom-20 -right-16 -z-10 size-56 rounded-full border-[38px] border-emerald-500/10" />
                <p className="text-xs font-bold uppercase tracking-[.16em] text-emerald-400">Lihat langsung cara kerjanya</p>
                <h2 className="mt-4 max-w-xl text-balance text-2xl font-bold leading-tight tracking-[-.04em] sm:text-3xl">AHLAN menjawab dari brosur pesantren Anda.</h2>
                <p className="mt-3 max-w-xl text-sm leading-6 text-slate-400">Demo 20 menit. Cukup siapkan brosur PSB dan kanal pesan yang sudah digunakan.</p>
                <div className="mt-7"><DemoButton variant="light">Jadwalkan demo</DemoButton></div>
              </aside>
            )}

            {data.related && (
              <nav aria-label="Halaman terkait" className="mt-14 border-t border-slate-200 pt-7">
                <p className="text-xs font-bold uppercase tracking-[.16em] text-slate-400">Baca selanjutnya</p>
                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                  {data.related.map((related) => <Link key={related.href} href={related.href} className="group flex items-center justify-between gap-4 rounded-xl border border-slate-200 bg-slate-50/60 px-4 py-3.5 text-sm font-semibold text-slate-700 transition-all hover:border-emerald-200 hover:bg-emerald-50 hover:text-emerald-800">{related.label}<ArrowUpRight size={15} className="shrink-0 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" /></Link>)}
                </div>
              </nav>
            )}
          </article>
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
