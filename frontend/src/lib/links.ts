/** Client-side checks mirroring the server rules in PRD 25. The server stays the authority; these only give early feedback. */

const mapsHosts = new Set(["maps.google.com", "maps.app.goo.gl"]);

export function validateMapsUrl(value: string): string {
  const raw = value.trim();
  if (!raw) return "";
  if (raw.length > 2048) return "Tautan terlalu panjang.";
  let url: URL;
  try { url = new URL(raw); } catch { return "Tautan tidak valid."; }
  if (url.protocol !== "https:") return "Tautan harus diawali https://.";
  if (url.username || url.password) return "Tautan tidak boleh memuat nama pengguna atau kata sandi.";
  const host = url.hostname.toLowerCase();
  const ok = mapsHosts.has(host)
    || ((host === "www.google.com" || host === "google.com") && url.pathname.startsWith("/maps"))
    || (host === "goo.gl" && url.pathname.startsWith("/maps"));
  return ok ? "" : "Gunakan tautan dari Google Maps (maps.app.goo.gl atau google.com/maps).";
}

export type SourceKind = "SITE" | "SINGLE_PAGE" | "GOOGLE_DOC" | "GOOGLE_SHEET" | "GOOGLE_SLIDES" | "GOOGLE_DRIVE_FILE";

export const sourceKindLabel: Record<SourceKind, string> = {
  SITE: "Situs web",
  SINGLE_PAGE: "Satu halaman",
  GOOGLE_DOC: "Google Docs",
  GOOGLE_SHEET: "Google Sheets",
  GOOGLE_SLIDES: "Google Slides",
  GOOGLE_DRIVE_FILE: "File Google Drive",
};

export type ParsedSource = { ok: true; kind: SourceKind } | { ok: false; error: string };

/** Detects the source kind from a pasted URL. Google links must be publicly shared ("siapa saja yang memiliki link"). */
export function parseSourceUrl(value: string): ParsedSource {
  const raw = value.trim();
  if (!raw) return { ok: false, error: "Tautan wajib diisi." };
  if (raw.length > 2048) return { ok: false, error: "Tautan terlalu panjang." };
  let url: URL;
  try { url = new URL(raw); } catch { return { ok: false, error: "Tautan tidak valid." }; }
  if (url.protocol !== "https:" && url.protocol !== "http:") return { ok: false, error: "Hanya tautan http/https yang didukung." };
  if (url.username || url.password) return { ok: false, error: "Tautan tidak boleh memuat nama pengguna atau kata sandi." };
  if (url.port && url.port !== "80" && url.port !== "443") return { ok: false, error: "Port selain 80/443 tidak didukung." };
  const host = url.hostname.toLowerCase();
  if (host === "docs.google.com") {
    const match = /^\/(document|spreadsheets|presentation)\/d\/([\w-]{20,})/.exec(url.pathname);
    if (!match) return { ok: false, error: "Tautan Google Docs/Sheets/Slides tidak dikenali." };
    return { ok: true, kind: ({ document: "GOOGLE_DOC", spreadsheets: "GOOGLE_SHEET", presentation: "GOOGLE_SLIDES" } as const)[match[1] as "document" | "spreadsheets" | "presentation"] };
  }
  if (host === "drive.google.com") {
    if (!/^\/file\/d\/[\w-]{20,}/.test(url.pathname)) return { ok: false, error: "Hanya tautan file Drive (/file/d/…) yang didukung, bukan folder." };
    return { ok: true, kind: "GOOGLE_DRIVE_FILE" };
  }
  if (host === "maps.google.com" || host === "maps.app.goo.gl" || ((host === "www.google.com" || host === "google.com") && url.pathname.startsWith("/maps"))) {
    return { ok: false, error: "Tautan Google Maps disimpan di Informasi pesantren, bukan di Sumber Web." };
  }
  if (host === "localhost" || /^[\d.]+$/.test(host) || host.includes(":") || !host.includes(".")) return { ok: false, error: "Gunakan nama domain publik." };
  return { ok: true, kind: "SITE" };
}

export type LinkKind = "PENDAFTARAN" | "BROSUR" | "JADWAL" | "BIAYA" | "WHATSAPP" | "INSTAGRAM" | "YOUTUBE" | "FACEBOOK" | "TIKTOK" | "WEBSITE" | "LAINNYA";

/** Kinds an official link can have. `hosts` restricts the address; `learn` says whether the page can also be read as knowledge. */
export const linkKinds: { value: LinkKind; label: string; hosts?: string[]; learn: boolean; example: string }[] = [
  { value: "PENDAFTARAN", label: "Pendaftaran online", learn: true, example: "https://forms.gle/…" },
  { value: "BROSUR", label: "Brosur", learn: true, example: "https://drive.google.com/file/d/…" },
  { value: "JADWAL", label: "Jadwal", learn: true, example: "https://docs.google.com/…" },
  { value: "BIAYA", label: "Biaya", learn: true, example: "https://docs.google.com/…" },
  { value: "WHATSAPP", label: "WhatsApp panitia", hosts: ["wa.me", "api.whatsapp.com", "chat.whatsapp.com", "wa.link"], learn: false, example: "https://wa.me/62812…" },
  { value: "INSTAGRAM", label: "Instagram", hosts: ["instagram.com", "instagr.am"], learn: false, example: "https://www.instagram.com/…" },
  { value: "YOUTUBE", label: "YouTube", hosts: ["youtube.com", "youtu.be"], learn: false, example: "https://www.youtube.com/…" },
  { value: "FACEBOOK", label: "Facebook", hosts: ["facebook.com", "fb.com", "fb.me"], learn: false, example: "https://www.facebook.com/…" },
  { value: "TIKTOK", label: "TikTok", hosts: ["tiktok.com"], learn: false, example: "https://www.tiktok.com/@…" },
  { value: "WEBSITE", label: "Website", learn: true, example: "https://pesantren.id" },
  { value: "LAINNYA", label: "Tautan lain", learn: true, example: "https://…" },
];

const hostIs = (host: string, domain: string) => host === domain || host.endsWith(`.${domain}`);
const unlearnableHosts = ["forms.gle", "wa.me", "instagram.com", "facebook.com", "fb.com", "fb.me", "tiktok.com", "youtube.com", "youtu.be", "linktr.ee"];

/** Early feedback mirroring the server's rules for an official link; the server remains the authority. */
export function validateOfficialLink(kind: LinkKind, value: string): string {
  const raw = value.trim();
  if (!raw) return "";
  let url: URL;
  try { url = new URL(raw); } catch { return "Tautan tidak valid."; }
  if (raw.length > 2048) return "Tautan terlalu panjang.";
  if (url.protocol !== "https:") return "Tautan harus diawali https://.";
  if (url.username || url.password) return "Tautan tidak boleh memuat nama pengguna atau kata sandi.";
  const host = url.hostname.toLowerCase();
  if (host === "localhost" || /^[\d.]+$/.test(host) || host.includes(":") || !host.includes(".")) return "Gunakan nama domain publik.";
  const rule = linkKinds.find(item => item.value === kind);
  if (rule?.hosts && !rule.hosts.some(domain => hostIs(host, domain))) return `Tautan ${rule.label} harus dari ${rule.hosts[0]}.`;
  return "";
}

/** Whether an official link can also be registered as a Sumber Web (read for knowledge). */
export function canLearnLink(kind: LinkKind, value: string): boolean {
  if (!linkKinds.find(item => item.value === kind)?.learn) return false;
  const parsed = parseSourceUrl(value);
  if (!parsed.ok) return false;
  try { const host = new URL(value.trim()).hostname.toLowerCase(); return !unlearnableHosts.some(domain => hostIs(host, domain)); } catch { return false; }
}
