"use client";

import { useState } from "react";
import { Check, Copy, ExternalLink } from "lucide-react";

/** Shows a pesantren's public link page (/p/<slug>) with copy and preview buttons. */
export function ShareLinkCard({ slug, title = "Link bagikan untuk calon santri" }: { slug?: string; title?: string }) {
  const [copied, setCopied] = useState(false);
  if (!slug) return null;
  const url = `${typeof window === "undefined" ? "" : window.location.origin}/p/${slug}`;
  async function copy() {
    try { await navigator.clipboard.writeText(url); setCopied(true); setTimeout(() => setCopied(false), 2000); } catch { window.prompt("Salin tautan ini:", url); }
  }
  return <div className="panel panel-pad" style={{ marginBottom: 20 }}>
    <h2 className="section-title">{title}</h2>
    <p className="section-hint">Satu halaman berisi deskripsi, Telegram, WhatsApp, dan Maps. Tombol yang belum siap otomatis nonaktif.</p>
    <div style={{ display: "flex", gap: 10, flexWrap: "wrap", alignItems: "center", marginTop: 12 }}>
      <input className="input" readOnly value={url} aria-label="Link bagikan" style={{ flex: "1 1 260px" }} onFocus={event => event.currentTarget.select()} />
      <button className="button button-primary" onClick={() => void copy()}>{copied ? <Check size={15} /> : <Copy size={15} />}{copied ? "Tersalin" : "Salin link"}</button>
      <a className="button" href={url} target="_blank" rel="noopener noreferrer"><ExternalLink size={15} />Pratinjau</a>
    </div>
  </div>;
}
