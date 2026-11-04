import Image from "next/image";
import Link from "next/link";
import { ArrowRight, ChevronDown, LogIn, Menu } from "lucide-react";
import { demoHref, siteName, siteUrl } from "@/lib/seo";

const external = demoHref.startsWith("http");
export const demoLinkProps = external ? { target: "_blank", rel: "noopener noreferrer" } : {};

export const publicPages = [
  { href: "/chatbot-pesantren", label: "Chatbot pesantren" },
  { href: "/penerimaan-santri-baru", label: "Penerimaan santri baru" },
  { href: "/chatbot-whatsapp-pesantren", label: "WhatsApp & Telegram" },
  { href: "/panduan-psb-pesantren", label: "Panduan PSB" },
];

const focus = "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-500";

export function DemoButton({ children = "Jadwalkan demo", variant = "solid", className = "" }: { children?: React.ReactNode; variant?: "solid" | "light"; className?: string }) {
  const style = variant === "solid"
    ? "bg-emerald-600 text-white shadow-[0_8px_24px_rgba(5,150,105,.22)] hover:-translate-y-0.5 hover:bg-emerald-700"
    : "bg-white text-slate-950 shadow-lg shadow-black/10 hover:-translate-y-0.5 hover:bg-emerald-50";
  return (
    <a href={demoHref} {...demoLinkProps} className={`group inline-flex h-12 items-center justify-center gap-2 rounded-xl px-5 text-sm font-semibold transition-all ${style} ${focus} ${className}`}>
      {children} <ArrowRight size={16} className="transition-transform group-hover:translate-x-1" />
    </a>
  );
}

export function SiteHeader({ home = false }: { home?: boolean }) {
  const base = home ? "" : "/";
  const links = [
    { href: `${base}#solusi`, label: "Solusi" },
    { href: `${base}#fitur`, label: "Fitur" },
    { href: `${base}#cara-kerja`, label: "Cara kerja" },
    { href: `${base}#faq`, label: "FAQ" },
  ];
  const pageLinks = [
    { href: "/chatbot-pesantren", label: "Chatbot" },
    { href: "/penerimaan-santri-baru", label: "Penerimaan santri baru" },
    { href: "/chatbot-whatsapp-pesantren", label: "WhatsApp & Telegram" },
    { href: "/panduan-psb-pesantren", label: "Panduan PSB" },
  ];
  const navClass = `whitespace-nowrap rounded-lg px-3 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-950 ${focus}`;
  return (
    <header className="sticky top-0 z-50 border-b border-slate-200/70 bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex h-[72px] w-full max-w-[1212px] items-center gap-5 px-4">
        <Link href="/" className={`flex shrink-0 items-center gap-2.5 rounded-lg ${focus}`} aria-label={`${siteName} — beranda`}>
          <Image src="/logo.png" alt="" width={36} height={36} priority className="rounded-xl" />
          <span className="text-lg font-extrabold tracking-[-.04em] text-slate-950">AHLAN<span className="text-emerald-600">.</span></span>
        </Link>

        <nav aria-label="Navigasi utama" className="hidden min-w-0 flex-1 items-center gap-0.5 lg:flex">
          {links.map((link) => <a key={link.label} href={link.href} className={navClass}>{link.label}</a>)}
          <div className="group relative">
            <button type="button" aria-haspopup="true" className={`${navClass} inline-flex items-center gap-1`}>
              Halaman <ChevronDown size={14} className="transition-transform group-hover:rotate-180 group-focus-within:rotate-180" aria-hidden="true" />
            </button>
            <div className="invisible absolute left-0 top-full w-60 pt-2 opacity-0 transition-all group-focus-within:visible group-focus-within:opacity-100 group-hover:visible group-hover:opacity-100">
              <div className="rounded-2xl border border-slate-200 bg-white p-2 shadow-2xl shadow-slate-900/10">
                {pageLinks.map((link) => <Link key={link.href} href={link.href} className="block rounded-xl px-3 py-2.5 text-sm font-medium text-slate-700 hover:bg-slate-100 hover:text-emerald-700">{link.label}</Link>)}
              </div>
            </div>
          </div>
        </nav>

        <div className="ml-auto flex shrink-0 items-center gap-2 lg:ml-0">
          <a href={demoHref} {...demoLinkProps} className={`hidden h-10 items-center rounded-lg bg-slate-950 px-4 text-sm font-semibold text-white transition-colors hover:bg-emerald-700 sm:inline-flex ${focus}`}>Coba demo</a>
          <Link href="/login" className={`inline-flex h-10 items-center gap-2 rounded-lg border border-slate-200 bg-white px-3.5 text-sm font-semibold text-slate-800 shadow-sm transition-colors hover:border-slate-300 hover:bg-slate-50 sm:px-4 ${focus}`}>
            <LogIn size={16} aria-hidden="true" />
            <span>Masuk</span>
          </Link>
          <details className="group relative lg:hidden">
            <summary aria-label="Buka menu" className={`grid size-10 cursor-pointer list-none place-items-center rounded-lg border border-slate-200 text-slate-700 transition-colors hover:bg-slate-50 [&::-webkit-details-marker]:hidden ${focus}`}><Menu size={19} /></summary>
            <nav aria-label="Navigasi seluler" className="absolute right-0 top-12 w-64 rounded-2xl border border-slate-200 bg-white p-2 shadow-2xl shadow-slate-900/10">
              {links.map((link) => <a key={link.label} href={link.href} className="block rounded-xl px-3 py-2.5 text-sm font-medium text-slate-700 hover:bg-slate-100">{link.label}</a>)}
              <div className="my-1 border-t border-slate-100" />
              {pageLinks.map((link) => <Link key={link.href} href={link.href} className="block rounded-xl px-3 py-2.5 text-sm font-medium text-slate-700 hover:bg-slate-100">{link.label}</Link>)}
              <a href={demoHref} {...demoLinkProps} className="mt-1 block rounded-xl bg-emerald-600 px-3 py-2.5 text-center text-sm font-semibold text-white sm:hidden">Jadwalkan demo</a>
            </nav>
          </details>
        </div>
      </div>
    </header>
  );
}

export function SiteFooter() {
  return (
    <footer className="border-t border-slate-200 bg-white">
      <div className="mx-auto grid w-full max-w-[1212px] gap-10 px-4 py-14 md:grid-cols-[1fr_1.25fr]">
        <div>
          <Link href="/" className="flex w-fit items-center gap-2.5"><Image src="/logo.png" alt="" width={34} height={34} className="rounded-xl" /><span className="text-lg font-extrabold tracking-[-.04em] text-slate-950">AHLAN<span className="text-emerald-600">.</span></span></Link>
          <p className="mt-4 max-w-sm text-sm leading-6 text-slate-500">Asisten AI untuk penerimaan santri yang menjawab dari dokumen resmi pesantren Anda.</p>
        </div>
        <nav aria-label="Navigasi footer" className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm sm:grid-cols-3">
          {publicPages.map((page) => <Link key={page.href} href={page.href} className="text-slate-500 transition-colors hover:text-emerald-700">{page.label}</Link>)}
          <Link href="/kebijakan-privasi" className="text-slate-500 hover:text-emerald-700">Kebijakan privasi</Link>
          <Link href="/login" className="font-semibold text-slate-800 hover:text-emerald-700">Masuk pengurus</Link>
        </nav>
      </div>
      <div className="mx-auto flex w-full max-w-[1212px] flex-col justify-between gap-2 border-t border-slate-200 px-4 py-5 text-xs text-slate-400 sm:flex-row">
        <span>© 2026 {siteName}. Seluruh hak dilindungi.</span>
        <a href={siteUrl} className="hover:text-emerald-700">{siteUrl.replace(/^https?:\/\//, "")}</a>
      </div>
    </footer>
  );
}
