import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/seo";

const pages: { path: string; priority: number }[] = [
  { path: "/", priority: 1 },
  { path: "/chatbot-pesantren", priority: 0.8 },
  { path: "/penerimaan-santri-baru", priority: 0.8 },
  { path: "/chatbot-whatsapp-pesantren", priority: 0.7 },
  { path: "/panduan-psb-pesantren", priority: 0.7 },
  { path: "/kebijakan-privasi", priority: 0.3 },
];

const backend = process.env.BACKEND_URL ?? "http://localhost:8080";

/** Slugs of pesantren whose link page is live. A backend outage must not break the sitemap, so failures yield none. */
async function pesantrenSlugs(): Promise<string[]> {
  try {
    const response = await fetch(`${backend}/api/v1/public/p`, { next: { revalidate: 3600 } });
    if (!response.ok) return [];
    return ((await response.json()) as { data?: string[] }).data ?? [];
  } catch {
    return [];
  }
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const slugs = await pesantrenSlugs();
  return [
    ...pages.map(({ path, priority }) => ({ url: `${siteUrl}${path}`, changeFrequency: "monthly" as const, priority })),
    ...slugs.map(slug => ({ url: `${siteUrl}/p/${encodeURIComponent(slug)}`, changeFrequency: "weekly" as const, priority: 0.5 })),
  ];
}
