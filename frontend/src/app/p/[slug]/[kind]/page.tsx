import type { Metadata } from "next";
import { cacheLife } from "next/cache";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import { ArrowLeft, ExternalLink } from "lucide-react";
import { siteName } from "@/lib/seo";
import { linkKinds } from "@/lib/links";

type Folder = { slug: string; name: string; logo_url?: string; kind: string; items: { label: string; url: string; hint: string }[] };

const backend = process.env.BACKEND_URL ?? "http://localhost:8080";

async function getFolder(slug: string, kind: string): Promise<Folder | null> {
  "use cache";
  cacheLife("minutes");
  const response = await fetch(`${backend}/api/v1/public/p/${encodeURIComponent(slug)}/links/${encodeURIComponent(kind)}`);
  if (response.status === 404) return null;
  if (!response.ok) throw new Error(`public folder request failed: ${response.status}`);
  return response.json() as Promise<Folder>;
}

const kindLabel = (kind: string) => linkKinds.find(item => item.value === kind.toUpperCase())?.label ?? kind;

export async function generateMetadata({ params }: { params: Promise<{ slug: string; kind: string }> }): Promise<Metadata> {
  const { slug, kind } = await params;
  const folder = await getFolder(slug, kind).catch(() => null);
  if (!folder) return { title: `Tidak ditemukan | ${siteName}`, robots: { index: false } };
  return { title: { absolute: `${kindLabel(folder.kind)} | ${folder.name}` }, alternates: { canonical: `/p/${folder.slug}/${kind.toLowerCase()}` } };
}

async function FolderList({ params }: { params: Promise<{ slug: string; kind: string }> }) {
  const { slug, kind } = await params;
  const folder = await getFolder(slug, kind);
  if (!folder) notFound();
  return (
    <main className="relative mx-auto flex min-h-screen w-full max-w-md flex-col px-5 pb-8 pt-10">
      <Link href={`/p/${folder.slug}`} className="inline-flex items-center gap-1.5 self-start text-sm font-medium text-[#14633f] hover:underline"><ArrowLeft size={16} aria-hidden />Kembali</Link>
      <div className="mt-6 flex flex-col items-center text-center">
        <div className="grid size-20 place-items-center rounded-full bg-white p-3 shadow-[0_4px_20px_rgba(15,61,42,.15)]">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={folder.logo_url || "/logo.png"} alt="" width={64} height={64} className="size-full object-contain" />
        </div>
        <p className="mt-4 text-sm text-slate-600">{folder.name}</p>
        <h1 className="mt-1 text-balance text-2xl font-extrabold tracking-tight text-[#0f3d2a]">{kindLabel(folder.kind)}</h1>
      </div>
      <ul className="mt-7 space-y-3">
        {folder.items.map(item => <li key={item.url}>
          <a href={item.url} target="_blank" rel="noopener noreferrer" className="flex w-full items-center gap-4 rounded-2xl bg-white px-4 py-3.5 text-slate-900 shadow-[0_2px_14px_rgba(15,61,42,.08)] transition hover:-translate-y-0.5 hover:shadow-[0_6px_20px_rgba(15,61,42,.14)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-700">
            <span className="min-w-0 flex-1 text-left"><span className="block text-[15px] font-semibold">{item.label}</span>{item.hint && <span className="block text-xs text-slate-500">{item.hint}</span>}</span>
            <ExternalLink size={16} className="text-slate-400" aria-hidden /><span className="sr-only">(buka di tab baru)</span>
          </a>
        </li>)}
      </ul>
      <p className="mt-auto pt-12 text-center text-xs text-slate-600">Dibuat dengan <strong className="font-bold text-[#0f3d2a]">{siteName.toUpperCase()}</strong></p>
    </main>
  );
}

export default function Page({ params }: { params: Promise<{ slug: string; kind: string }> }) {
  return (
    <div className="relative min-h-screen overflow-hidden bg-[linear-gradient(180deg,#eaf6ee_0%,#f7fbf8_45%,#e6f3ea_100%)]">
      <Suspense fallback={<div className="mx-auto min-h-screen max-w-md px-5 py-10" aria-busy="true" />}>
        <FolderList params={params} />
      </Suspense>
    </div>
  );
}
