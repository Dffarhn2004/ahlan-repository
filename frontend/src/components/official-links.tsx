"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { LoaderCircle, Pencil, Plus, Trash2 } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { canLearnLink, linkKinds, validateOfficialLink, type LinkKind } from "@/lib/links";

type OfficialLink = { id: string; kind: LinkKind; label: string; url: string; share_hint: string; share: boolean; show_public: boolean; learn: boolean; crawl_source_id?: string };
type Draft = { kind: LinkKind; label: string; url: string; share_hint: string; share: boolean; show_public: boolean; learn: boolean };

const emptyDraft: Draft = { kind: "PENDAFTARAN", label: "", url: "", share_hint: "", share: true, show_public: false, learn: false };
const socialKinds: LinkKind[] = ["WEBSITE", "INSTAGRAM", "FACEBOOK", "YOUTUBE", "TIKTOK"];
const errorText: Record<string, string> = {
  LINK_EXISTS: "Tautan ini sudah terdaftar.",
  LINK_LIMIT: "Batas 20 tautan tercapai. Hapus tautan lama terlebih dahulu.",
  LINK_DOMAIN_MISMATCH: "Alamat tidak sesuai dengan jenis tautan yang dipilih.",
  LINK_URL_INVALID: "Tautan tidak valid. Gunakan alamat https:// yang lengkap.",
  LINK_NEEDS_PURPOSE: "Pilih minimal satu: dibagikan, ditampilkan ke publik, atau dipelajari.",
  LINK_CANNOT_LEARN: "Isi tautan ini tidak dapat dipelajari asisten (misalnya formulir atau media sosial).",
  CRAWL_SOURCE_LIMIT: "Batas jumlah Sumber Web tercapai. Hapus sumber lama atau jangan pilih \"pelajari\".",
  CRAWL_URL_FORBIDDEN: "Tautan ditolak karena tidak aman atau tidak dapat dijangkau.",
  CRAWL_URL_UNREACHABLE: "Alamat situs tidak ditemukan. Periksa penulisan tautan.",
};

/** Official links the assistant can hand to prospective students and/or learn from. One link may do both. */
export function OfficialLinks({ tenantId, canEdit }: { tenantId: string; canEdit: boolean }) {
  const [items, setItems] = useState<OfficialLink[] | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [editingId, setEditingId] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const base = `/api/v1/tenants/${tenantId}/links`;

  const [reload, setReload] = useState(0);
  const refresh = () => setReload(value => value + 1);
  useEffect(() => {
    if (!tenantId) return;
    let cancelled = false;
    void (async () => {
      try { const result = await api<{ data: OfficialLink[] }>(`/api/v1/tenants/${tenantId}/links`); if (!cancelled) setItems(result?.data ?? []); }
      catch { if (!cancelled) { setItems([]); setNotice("Daftar tautan belum dapat dimuat."); } }
    })();
    return () => { cancelled = true; };
  }, [tenantId, reload]);

  const kind = linkKinds.find(item => item.value === draft?.kind);
  const urlError = draft ? validateOfficialLink(draft.kind, draft.url) : "";
  const learnable = draft ? canLearnLink(draft.kind, draft.url) : false;
  const original = items?.find(item => item.id === editingId);
  const fail = (error: unknown, fallback: string) => setNotice(error instanceof ApiError ? (errorText[error.message] ?? fallback) : fallback);

  const set = <K extends keyof Draft>(key: K, value: Draft[K]) => setDraft(current => current && { ...current, [key]: value });
  const start = (item?: OfficialLink) => {
    setNotice(""); setEditingId(item?.id ?? "");
    setDraft(item ? { kind: item.kind, label: item.label, url: item.url, share_hint: item.share_hint, share: item.share, show_public: item.show_public, learn: item.learn } : emptyDraft);
  };

  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!draft || urlError) return;
    if (original?.learn && (!draft.learn || draft.url.trim() !== original.url) && !window.confirm("Hasil yang sudah dipelajari dari tautan ini (termasuk koreksi Anda) akan dihapus dari pengetahuan asisten. Lanjutkan?")) return;
    setBusy(true); setNotice("");
    try {
      const body = JSON.stringify({ ...draft, url: draft.url.trim(), show_public: draft.show_public && !socialKinds.includes(draft.kind), learn: draft.learn && learnable });
      await api(editingId ? `${base}/${editingId}` : base, { method: editingId ? "PATCH" : "POST", body });
      setDraft(null); setEditingId("");
      setNotice("Tautan disimpan dan diteruskan ke asisten.");
      refresh();
    } catch (error) { fail(error, "Tautan gagal disimpan."); }
    finally { setBusy(false); }
  }

  async function remove(item: OfficialLink) {
    const warning = item.learn ? "Tautan ini juga dipelajari asisten; hasil pelajarannya akan ikut dihapus. Hapus tautan?" : "Hapus tautan ini?";
    if (!window.confirm(warning)) return;
    setBusy(true); setNotice("");
    try { await api(`${base}/${item.id}`, { method: "DELETE" }); refresh(); setNotice("Tautan dihapus."); }
    catch (error) { fail(error, "Tautan gagal dihapus."); }
    finally { setBusy(false); }
  }

  return <section className="panel panel-pad">
    <div className="card-heading"><div><h2 className="section-title">Tautan resmi</h2><p className="section-hint">Daftar tautan yang boleh dibagikan asisten kepada calon santri (pendaftaran, brosur, WhatsApp, media sosial). Asisten hanya membagikan tautan dari daftar ini dan tidak akan membuat tautan sendiri. Tautan berupa situs atau dokumen Google juga bisa dipelajari isinya.</p></div>
      {canEdit && !draft && <button className="button" type="button" onClick={() => start()}><Plus size={14} />Tambah tautan</button>}</div>

    {draft && canEdit && <form onSubmit={save} style={{ display: "grid", gap: 14, marginBottom: 18, paddingBottom: 18, borderBottom: "1px solid var(--line)" }}>
      <div className="form-row">
        <label><span className="field-label">Jenis</span>
          <select className="input" value={draft.kind} onChange={event => { const next = event.target.value as LinkKind; setDraft(current => current && { ...current, kind: next, learn: current.learn && canLearnLink(next, current.url) }); }}>
            {linkKinds.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select></label>
        <label><span className="field-label">Nama tampilan</span>
          <input className="input" required maxLength={80} value={draft.label} onChange={event => set("label", event.target.value)} placeholder="Formulir PPDB 2027" /></label>
      </div>
      <label><span className="field-label">Tautan</span>
        <input className="input" type="url" required value={draft.url} onChange={event => { const value = event.target.value; setDraft(current => current && { ...current, url: value, learn: current.learn && canLearnLink(current.kind, value) }); }} placeholder={kind?.example} aria-invalid={Boolean(urlError)} />
        {urlError && <span className="error">{urlError}</span>}</label>
      <label><span className="field-label">Kapan asisten membagikannya (opsional)</span>
        <input className="input" maxLength={160} value={draft.share_hint} onChange={event => set("share_hint", event.target.value)} placeholder="Saat calon santri bertanya cara mendaftar" />
        <span className="field-help">Petunjuk singkat agar asisten tahu kapan tautan ini cocok dikirim.</span></label>
      <div style={{ display: "grid", gap: 8 }}>
        <label style={{ display: "flex", gap: 8, alignItems: "flex-start" }}><input type="checkbox" checked={draft.share} onChange={event => set("share", event.target.checked)} />
          <span><strong>Bagikan ke calon santri</strong><span className="field-help" style={{ display: "block" }}>Asisten boleh mengirim tautan ini saat relevan.</span></span></label>
        <label style={{ display: "flex", gap: 8, alignItems: "flex-start", opacity: socialKinds.includes(draft.kind) ? 0.6 : 1 }}><input type="checkbox" checked={draft.show_public && !socialKinds.includes(draft.kind)} disabled={socialKinds.includes(draft.kind)} onChange={event => set("show_public", event.target.checked)} />
          <span><strong>Tampilkan di halaman publik</strong><span className="field-help" style={{ display: "block" }}>{socialKinds.includes(draft.kind) ? "Media sosial dan website sudah tampil sebagai ikon di halaman publik (isi di Informasi resmi pesantren)." : "Tautan muncul di halaman publik dalam folder sesuai jenisnya, dan bisa dibuka siapa saja."}</span></span></label>
        <label style={{ display: "flex", gap: 8, alignItems: "flex-start", opacity: learnable ? 1 : 0.6 }}><input type="checkbox" checked={draft.learn && learnable} disabled={!learnable} onChange={event => set("learn", event.target.checked)} />
          <span><strong>Pelajari isi tautan ini juga</strong><span className="field-help" style={{ display: "block" }}>{learnable ? "Isinya diambil otomatis setiap hari dan dipakai asisten untuk menjawab. Kelola di menu Sumber Web." : draft.url.trim() ? "Tidak tersedia untuk tautan ini (formulir, WhatsApp, atau media sosial tidak dapat dibaca isinya)." : "Isi tautan terlebih dahulu."}</span></span></label>
      </div>
      <div style={{ display: "flex", gap: 8 }}>
        <button className="button button-primary" disabled={busy || Boolean(urlError) || (!draft.share && !(draft.show_public && !socialKinds.includes(draft.kind)) && !(draft.learn && learnable))}>{busy ? "Menyimpan…" : editingId ? "Simpan tautan" : "Tambah tautan"}</button>
        <button className="button" type="button" disabled={busy} onClick={() => { setDraft(null); setEditingId(""); }}>Batal</button>
      </div>
    </form>}

    {notice && <p className={/disimpan|dihapus/.test(notice) ? "status" : "error"} role="status" style={{ marginBottom: 12 }}>{notice}</p>}
    {items === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
    {items?.length === 0 && !draft && <p className="empty">Belum ada tautan resmi.</p>}
    <div style={{ display: "grid" }}>
      {items?.map(item => <div key={item.id} style={{ display: "flex", gap: 12, justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", padding: "12px 0", borderTop: "1px solid var(--line)" }}>
        <div style={{ minWidth: 0, flex: 1 }}>
          <strong>{item.label}</strong> <span className="field-help">· {linkKinds.find(kindItem => kindItem.value === item.kind)?.label}</span>
          <a className="field-help" style={{ display: "block", overflowWrap: "anywhere" }} href={item.url} target="_blank" rel="noopener noreferrer">{item.url}</a>
          {item.share_hint && <span className="field-help" style={{ display: "block" }}>Dibagikan saat: {item.share_hint}</span>}
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginTop: 6 }}>
            {item.share && <span className="status">Dibagikan</span>}
            {item.show_public && <span className="status">Tampil di publik</span>}
            {item.learn && item.crawl_source_id && <Link className="status status-muted" href={`/ponpes/sources/${item.crawl_source_id}`}>Dipelajari · lihat hasil</Link>}
          </div>
        </div>
        {canEdit && <div style={{ display: "flex", gap: 8 }}>
          <button className="button" type="button" disabled={busy} onClick={() => start(item)}><Pencil size={14} />Ubah</button>
          <button className="button" type="button" disabled={busy} onClick={() => void remove(item)}><Trash2 size={14} />Hapus</button>
        </div>}
      </div>)}
    </div>
  </section>;
}
