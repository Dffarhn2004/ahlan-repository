import type { Metadata } from "next";

export const siteName = "AHLAN";
export const siteTitle = "Chatbot AI untuk Penerimaan Santri | AHLAN Pesantren";
export const siteDescription = "Jawab pertanyaan PPDB/PSB pesantren otomatis via WhatsApp dan Telegram berdasarkan brosur resmi. Pantau dokumen, percakapan, dan jawaban dari satu dashboard.";
/** Public canonical origin. Development can override this in `.env.local`. */
export const siteUrl = (process.env.NEXT_PUBLIC_SITE_URL ?? "https://ahlanapp.cloud").replace(/\/$/, "");
export const themeColor = "#167554";
export const keywords = ["chatbot pesantren", "penerimaan santri baru", "PSB pesantren", "PPDB pesantren", "chatbot WhatsApp pesantren", "chatbot Telegram pesantren", "AHLAN"];

/** Halaman privat (area login) tidak boleh diindeks mesin pencari. */
export function privateSection(title: string): Metadata {
  return { title, robots: { index: false, follow: false, nocache: true, googleBot: { index: false, follow: false } } };
}

/**
 * Kontak untuk calon pelanggan: tombol demo membuka email ke ahlan.platformai@gmail.com
 * (bisa diganti lewat `NEXT_PUBLIC_CONTACT_EMAIL`).
 */
const contactEmail = process.env.NEXT_PUBLIC_CONTACT_EMAIL || "ahlan.platformai@gmail.com";
const demoMessage = "Assalamu'alaikum, saya ingin menjadwalkan demo AHLAN (20 menit) untuk pesantren kami.";
export const demoHref = `mailto:${contactEmail}?subject=${encodeURIComponent("Jadwalkan demo AHLAN")}&body=${encodeURIComponent(demoMessage)}`;
