import type { Metadata } from "next";
import { cacheLife } from "next/cache";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Suspense, type ReactNode } from "react";
import { ChevronRight, FolderOpen, Globe, Lock, MapPin, MessageCircle, Send, type LucideIcon } from "lucide-react";
import { siteName } from "@/lib/seo";
import { linkKinds } from "@/lib/links";

type PublicLink = {
  slug: string; name: string; description: string; address?: string;
  maps_url: string; telegram_url: string; whatsapp_url: string; socials?: Record<string, string>;
  logo_url?: string; folders?: { kind: string; count: number }[];
};

const backend = process.env.BACKEND_URL ?? "http://localhost:8080";

/** Public data only. A few minutes of caching keeps a viral link from hitting the API on every visit. */
async function getLink(slug: string): Promise<PublicLink | null> {
  "use cache";
  cacheLife("minutes");
  const response = await fetch(`${backend}/api/v1/public/p/${encodeURIComponent(slug)}`);
  if (response.status === 404) return null;
  if (!response.ok) throw new Error(`public link request failed: ${response.status}`);
  return response.json() as Promise<PublicLink>;
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const link = await getLink(slug).catch(() => null);
  if (!link) return { title: `Tidak ditemukan | ${siteName}`, robots: { index: false } };
  const description = link.description.slice(0, 160) || `Tanya jawab penerimaan santri baru ${link.name} lewat Telegram atau WhatsApp.`;
  return {
    title: { absolute: `${link.name} | Penerimaan Santri Baru` },
    description,
    alternates: { canonical: `/p/${link.slug}` },
    openGraph: { title: link.name, description, type: "website", locale: "id_ID" },
  };
}

// lucide-react no longer ships brand icons, so these are drawn inline.
const InstagramIcon = () => <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden><rect x="3" y="3" width="18" height="18" rx="5" /><circle cx="12" cy="12" r="4" /><circle cx="17.2" cy="6.8" r="1" fill="currentColor" stroke="none" /></svg>;
const FacebookIcon = () => <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor" aria-hidden><path d="M24 12.07C24 5.4 18.63 0 12 0S0 5.4 0 12.07C0 18.1 4.39 23.1 10.13 24v-8.44H7.08v-3.49h3.05V9.41c0-3.02 1.8-4.7 4.54-4.7 1.31 0 2.69.24 2.69.24v2.97h-1.52c-1.49 0-1.96.93-1.96 1.89v2.26h3.33l-.53 3.49h-2.8V24C19.61 23.1 24 18.1 24 12.07z" /></svg>;
const YoutubeIcon = () => <svg viewBox="0 0 24 24" width="24" height="24" aria-hidden><rect x="1.5" y="4.5" width="21" height="15" rx="4.5" fill="currentColor" /><path d="M10 9l5 3-5 3z" fill="#fff" /></svg>;
const TiktokIcon = () => <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" aria-hidden><path d="M12.525.02c1.31-.02 2.61-.01 3.91-.02.08 1.53.63 3.09 1.75 4.17 1.12 1.11 2.7 1.62 4.24 1.79v4.03c-1.44-.05-2.89-.35-4.2-.97-.57-.26-1.1-.59-1.62-.93-.01 2.92.01 5.84-.02 8.75-.08 1.4-.54 2.79-1.35 3.94-1.31 1.92-3.58 3.17-5.91 3.21-1.43.08-2.86-.31-4.08-1.03-2.02-1.19-3.44-3.37-3.65-5.71-.02-.5-.03-1-.01-1.49.18-1.9 1.12-3.72 2.58-4.96 1.66-1.44 3.98-2.13 6.15-1.72.02 1.48-.04 2.96-.04 4.44-.99-.32-2.15-.23-3.02.37-.63.41-1.11 1.04-1.36 1.75-.21.51-.15 1.07-.14 1.61.24 1.64 1.82 3.02 3.5 2.87 1.12-.01 2.19-.66 2.77-1.61.19-.33.4-.67.41-1.06.1-1.79.06-3.57.07-5.36.01-4.03-.01-8.05.02-12.07z" /></svg>;

const socialItems = [
  { key: "website", label: "Website", icon: <Globe size={22} />, color: "text-sky-700" },
  { key: "instagram", label: "Instagram", icon: <InstagramIcon />, color: "text-pink-600" },
  { key: "facebook", label: "Facebook", icon: <FacebookIcon />, color: "text-blue-700" },
  { key: "youtube", label: "YouTube", icon: <YoutubeIcon />, color: "text-red-600" },
  { key: "tiktok", label: "TikTok", icon: <TiktokIcon />, color: "text-slate-900" },
];

function Social({ href, label, color, children }: { href?: string; label: string; color: string; children: ReactNode }) {
  const base = "grid size-12 place-items-center rounded-2xl border border-white/80 bg-white/70 shadow-sm";
  if (!href) return <span aria-hidden className={`${base} text-slate-300 opacity-60`}>{children}</span>;
  return <a href={href} target="_blank" rel="noopener noreferrer" aria-label={label} className={`${base} ${color} transition hover:-translate-y-0.5 hover:bg-white hover:shadow-md focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-700`}>{children}</a>;
}

function Action({ href, label, hint, icon: Icon, tone }: { href: string; label: string; hint: string; icon: LucideIcon; tone: string }) {
  const base = "flex w-full items-center gap-4 rounded-2xl px-4 py-3.5";
  const tile = "grid size-12 shrink-0 place-items-center rounded-xl";
  if (!href) return (
    <div aria-disabled="true" className={`${base} cursor-not-allowed bg-slate-100/80 text-slate-400`}>
      <span className={`${tile} bg-slate-200/80 text-slate-400`}><Icon size={22} /></span>
      <span className="min-w-0 flex-1 text-left"><span className="block text-[15px] font-semibold">{label}</span><span className="block text-xs">Belum tersedia</span></span>
      <Lock size={16} aria-hidden />
    </div>
  );
  return (
    <a href={href} target="_blank" rel="noopener noreferrer" className={`${base} bg-white text-slate-900 shadow-[0_2px_14px_rgba(15,61,42,.08)] transition hover:-translate-y-0.5 hover:shadow-[0_6px_20px_rgba(15,61,42,.14)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-700`}>
      <span className={`${tile} ${tone}`}><Icon size={22} /></span>
      <span className="min-w-0 flex-1 text-left"><span className="block text-[15px] font-semibold">{label}</span><span className="block truncate text-xs text-slate-500">{hint}</span></span>
      <ChevronRight size={18} className="text-slate-400" aria-hidden />
    </a>
  );
}

async function LinkCard({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const link = await getLink(slug);
  if (!link) notFound();
  const socials = link.socials ?? {};
  const folders = link.folders ?? [];
  const hasSocial = socialItems.some(item => socials[item.key]);
  return (
    <main className="relative mx-auto flex min-h-screen w-full max-w-md flex-col px-5 pb-8 pt-12">
      <div className="flex flex-col items-center text-center">
        <div className="grid size-24 place-items-center rounded-full bg-white p-3 shadow-[0_4px_20px_rgba(15,61,42,.15)]">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={link.logo_url || "/logo.png"} alt={link.logo_url ? `Logo ${link.name}` : siteName} width={72} height={72} className="size-full object-contain" />
        </div>
        <h1 className="mt-6 text-balance text-[28px] font-extrabold leading-tight tracking-tight text-[#0f3d2a]">{link.name}</h1>
        {link.address && <p className="mt-3 inline-flex max-w-full items-start gap-1.5 text-sm text-slate-700"><MapPin size={15} className="mt-0.5 shrink-0 text-[#14633f]" aria-hidden /><span className="line-clamp-2">{link.address}</span></p>}
        {link.description && <p className="mt-4 line-clamp-4 text-pretty text-[15px] leading-6 text-slate-600">{link.description}</p>}
      </div>
      {hasSocial && <div className="mt-6 flex justify-center gap-3">{socialItems.map(item => <Social key={item.key} href={socials[item.key]} label={item.label} color={item.color}>{item.icon}</Social>)}</div>}
      <nav aria-label="Tautan pesantren" className="mt-7 space-y-3">
        <Action href={link.telegram_url} label="Tanya via Telegram" hint="Chat dengan asisten pesantren" icon={Send} tone="bg-sky-100 text-sky-600" />
        <Action href={link.whatsapp_url} label="Tanya via WhatsApp" hint="Chat dengan asisten pesantren" icon={MessageCircle} tone="bg-emerald-100 text-emerald-600" />
        <Action href={link.maps_url} label="Lokasi di Google Maps" hint="Buka petunjuk arah ke pesantren" icon={MapPin} tone="bg-rose-100 text-rose-500" />
      </nav>
      {folders.length > 0 && <nav aria-label="Folder tautan" className="mt-6 space-y-3">
        <h2 className="px-1 text-xs font-bold uppercase tracking-wide text-slate-500">Informasi &amp; dokumen</h2>
        {folders.map(folder => {
          const label = linkKinds.find(item => item.value === folder.kind)?.label ?? folder.kind;
          return <Link key={folder.kind} href={`/p/${link.slug}/${folder.kind.toLowerCase()}`} className="flex w-full items-center gap-4 rounded-2xl bg-white px-4 py-3.5 text-slate-900 shadow-[0_2px_14px_rgba(15,61,42,.08)] transition hover:-translate-y-0.5 hover:shadow-[0_6px_20px_rgba(15,61,42,.14)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-700">
            <span className="grid size-12 shrink-0 place-items-center rounded-xl bg-amber-100 text-amber-600"><FolderOpen size={22} /></span>
            <span className="min-w-0 flex-1 text-left"><span className="block text-[15px] font-semibold">{label}</span><span className="block text-xs text-slate-500">{folder.count} tautan</span></span>
            <ChevronRight size={18} className="text-slate-400" aria-hidden />
          </Link>;
        })}
      </nav>}
      <p className="mt-auto flex items-center justify-center gap-2 pt-12 text-xs text-slate-600">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src="/logo.png" alt="" width={22} height={22} className="size-[22px]" />
        <span>Dibuat dengan <strong className="font-bold text-[#0f3d2a]">{siteName.toUpperCase()}</strong></span>
      </p>
    </main>
  );
}

export default function Page({ params }: { params: Promise<{ slug: string }> }) {
  return (
    <div className="relative min-h-screen overflow-hidden bg-[linear-gradient(180deg,#eaf6ee_0%,#f7fbf8_45%,#e6f3ea_100%)]">
      <div aria-hidden className="pointer-events-none absolute -left-20 -top-24 size-72 rounded-full bg-emerald-200/50 blur-3xl" />
      <div aria-hidden className="pointer-events-none absolute -right-24 top-0 size-72 rounded-full bg-lime-100/70 blur-3xl" />
      <div aria-hidden className="pointer-events-none absolute -bottom-32 left-1/2 h-72 w-[140%] -translate-x-1/2 rounded-[50%] bg-emerald-200/40 blur-2xl" />
      <Suspense fallback={<div className="mx-auto min-h-screen max-w-md px-5 py-10" aria-busy="true" />}>
        <LinkCard params={params} />
      </Suspense>
    </div>
  );
}
