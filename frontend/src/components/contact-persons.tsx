"use client";

import { useEffect, useState } from "react";
import { LoaderCircle, Pencil, Plus, TriangleAlert, Trash2 } from "lucide-react";
import { api, ApiError } from "@/lib/api";

type Contact = { id: string; name: string; role: string; phone: string; email: string; hours: string; share_hint: string };
type Draft = { name: string; role: string; phone: string; email: string; hours: string; share_hint: string };

const emptyDraft: Draft = { name: "", role: "", phone: "", email: "", hours: "", share_hint: "" };
const errorText: Record<string, string> = {
  CONTACT_LIMIT: "Batas 20 kontak tercapai. Hapus kontak lama terlebih dahulu.",
  CONTACT_PHONE_INVALID: "Nomor tidak valid. Gunakan nomor HP/WhatsApp, misalnya 0812 3456 7890.",
  CONTACT_EMAIL_INVALID: "Email tidak valid.",
};

/** Early phone check mirroring the server (digits, spaces, dashes, brackets, optional leading +; 8-15 digits). The server stays the authority. */
function phoneError(value: string): string {
  const raw = value.trim();
  if (!raw) return "";
  if (!/^[+\d\s().-]+$/.test(raw)) return "Nomor hanya boleh berisi angka, spasi, atau tanda hubung.";
  const digits = raw.replace(/\D/g, "");
  return digits.length < 8 || digits.length > 15 ? "Nomor harus 8–15 digit." : "";
}

/** Contact persons the assistant hands to people who ask for a contact. Every saved contact is shown to askers, so the card warns before saving. */
export function ContactPersons({ tenantId, canEdit }: { tenantId: string; canEdit: boolean }) {
  const [items, setItems] = useState<Contact[] | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [editingId, setEditingId] = useState("");
  const [understood, setUnderstood] = useState(false);
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const base = `/api/v1/tenants/${tenantId}/contacts`;

  const [reload, setReload] = useState(0);
  const refresh = () => setReload(value => value + 1);
  useEffect(() => {
    if (!tenantId) return;
    let cancelled = false;
    void (async () => {
      try { const result = await api<{ data: Contact[] }>(`/api/v1/tenants/${tenantId}/contacts`); if (!cancelled) setItems(result?.data ?? []); }
      catch { if (!cancelled) { setItems([]); setNotice("Daftar kontak belum dapat dimuat."); } }
    })();
    return () => { cancelled = true; };
  }, [tenantId, reload]);

  const fail = (error: unknown, fallback: string) => setNotice(error instanceof ApiError ? (errorText[error.message] ?? fallback) : fallback);
  const set = <K extends keyof Draft>(key: K, value: Draft[K]) => setDraft(current => current && { ...current, [key]: value });
  const start = (item?: Contact) => {
    setNotice(""); setEditingId(item?.id ?? ""); setUnderstood(false);
    setDraft(item ? { name: item.name, role: item.role, phone: item.phone, email: item.email, hours: item.hours, share_hint: item.share_hint } : emptyDraft);
  };
  const close = () => { setDraft(null); setEditingId(""); setUnderstood(false); };
  const phoneProblem = draft ? phoneError(draft.phone) : "";

  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!draft || phoneProblem || !understood) return;
    setBusy(true); setNotice("");
    try {
      await api(editingId ? `${base}/${editingId}` : base, { method: editingId ? "PATCH" : "POST", body: JSON.stringify({ ...draft, name: draft.name.trim(), phone: draft.phone.trim(), email: draft.email.trim() }) });
      close();
      setNotice("Kontak disimpan dan diteruskan ke asisten.");
      refresh();
    } catch (error) { fail(error, "Kontak gagal disimpan."); }
    finally { setBusy(false); }
  }

  async function remove(item: Contact) {
    if (!window.confirm(`Hapus kontak ${item.name}? Asisten tidak akan memberikan kontak ini lagi.`)) return;
    setBusy(true); setNotice("");
    try { await api(`${base}/${item.id}`, { method: "DELETE" }); refresh(); setNotice("Kontak dihapus."); }
    catch (error) { fail(error, "Kontak gagal dihapus."); }
    finally { setBusy(false); }
  }

  return <section className="panel panel-pad">
    <div className="card-heading"><div><h2 className="section-title">Kontak person</h2><p className="section-hint">Daftar orang yang boleh dihubungi wali santri atau calon santri, misalnya admin pendaftaran atau ustadz tertentu. Asisten hanya memberikan kontak dari daftar ini, dan memilih yang paling sesuai dengan keperluan wali santri.</p></div>
      {canEdit && !draft && <button className="button" type="button" onClick={() => start()}><Plus size={14} />Tambah kontak</button>}</div>

    <p role="note" style={{ display: "flex", gap: 8, alignItems: "flex-start", margin: "0 0 16px", padding: "10px 12px", borderRadius: 8, background: "var(--warning-soft, #fff6e0)", border: "1px solid var(--warning-line, #f0d9a0)" }}>
      <TriangleAlert size={16} style={{ flex: "none", marginTop: 2 }} aria-hidden />
      <span><strong>Kontak ini akan ditampilkan kepada wali santri.</strong> Siapa pun yang meminta kontak lewat asisten akan menerima nama dan nomor yang tersimpan di sini. Isi hanya nomor yang memang boleh diketahui publik, dan hindari nomor pribadi yang tidak ingin disebarkan.</span>
    </p>

    {draft && canEdit && <form onSubmit={save} style={{ display: "grid", gap: 14, marginBottom: 18, paddingBottom: 18, borderBottom: "1px solid var(--line)" }}>
      <div className="form-row">
        <label><span className="field-label">Nama</span>
          <input className="input" required maxLength={80} value={draft.name} onChange={event => set("name", event.target.value)} placeholder="Ustadz Ahmad" /></label>
        <label><span className="field-label">Jabatan atau bidang (opsional)</span>
          <input className="input" maxLength={80} value={draft.role} onChange={event => set("role", event.target.value)} placeholder="Admin pendaftaran TauD" /></label>
      </div>
      <div className="form-row">
        <label><span className="field-label">Nomor HP / WhatsApp</span>
          <input className="input" type="tel" required value={draft.phone} onChange={event => set("phone", event.target.value)} placeholder="0812 3456 7890" aria-invalid={Boolean(phoneProblem)} />
          {phoneProblem ? <span className="error">{phoneProblem}</span> : <span className="field-help">Nomor yang diawali 0 dibaca sebagai Indonesia (+62).</span>}</label>
        <label><span className="field-label">Email (opsional)</span>
          <input className="input" type="email" maxLength={120} value={draft.email} onChange={event => set("email", event.target.value)} placeholder="admin@pesantren.id" /></label>
      </div>
      <div className="form-row">
        <label><span className="field-label">Jam layanan (opsional)</span>
          <input className="input" maxLength={80} value={draft.hours} onChange={event => set("hours", event.target.value)} placeholder="Senin–Jumat 08.00–15.00" />
          <span className="field-help">Asisten akan menyebutkannya agar wali santri tidak menghubungi di luar jam.</span></label>
        <label><span className="field-label">Hubungi untuk (opsional)</span>
          <input className="input" maxLength={160} value={draft.share_hint} onChange={event => set("share_hint", event.target.value)} placeholder="Pendaftaran santri baru dan beasiswa" />
          <span className="field-help">Petunjuk singkat agar asisten tahu kontak mana yang cocok.</span></label>
      </div>
      <label style={{ display: "flex", gap: 8, alignItems: "flex-start" }}>
        <input type="checkbox" checked={understood} onChange={event => setUnderstood(event.target.checked)} />
        <span><strong>Saya mengerti kontak ini akan ditampilkan kepada wali santri.</strong><span className="field-help" style={{ display: "block" }}>Pastikan pemilik nomor setuju nomornya dibagikan.</span></span>
      </label>
      <div style={{ display: "flex", gap: 8 }}>
        <button className="button button-primary" disabled={busy || Boolean(phoneProblem) || !understood}>{busy ? "Menyimpan…" : editingId ? "Simpan kontak" : "Tambah kontak"}</button>
        <button className="button" type="button" disabled={busy} onClick={close}>Batal</button>
      </div>
    </form>}

    {notice && <p className={/disimpan|dihapus/.test(notice) ? "status" : "error"} role="status" style={{ marginBottom: 12 }}>{notice}</p>}
    {items === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
    {items?.length === 0 && !draft && <p className="empty">Belum ada kontak person. Tanpa kontak, asisten akan mengarahkan wali santri ke informasi kontak pada profil pesantren.</p>}
    <div style={{ display: "grid" }}>
      {items?.map(item => <div key={item.id} style={{ display: "flex", gap: 12, justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", padding: "12px 0", borderTop: "1px solid var(--line)" }}>
        <div style={{ minWidth: 0, flex: 1 }}>
          <strong>{item.name}</strong>{item.role && <span className="field-help"> · {item.role}</span>}
          <span className="field-help" style={{ display: "block" }}>{item.phone}{item.email && ` · ${item.email}`}</span>
          {item.hours && <span className="field-help" style={{ display: "block" }}>Jam layanan: {item.hours}</span>}
          {item.share_hint && <span className="field-help" style={{ display: "block" }}>Hubungi untuk: {item.share_hint}</span>}
          <div style={{ marginTop: 6 }}><span className="status">Ditampilkan ke wali santri</span></div>
        </div>
        {canEdit && <div style={{ display: "flex", gap: 8 }}>
          <button className="button" type="button" disabled={busy} onClick={() => start(item)}><Pencil size={14} />Ubah</button>
          <button className="button" type="button" disabled={busy} onClick={() => void remove(item)}><Trash2 size={14} />Hapus</button>
        </div>}
      </div>)}
    </div>
  </section>;
}
