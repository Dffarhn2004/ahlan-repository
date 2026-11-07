"use client";

import { useEffect, useState } from "react";
import { LoaderCircle, Pencil, Plus, Trash2, XCircle } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { ProcessingOverlay } from "@/components/processing-overlay";

type Mode = "PSB_ONLY" | "GENERAL";
type Admin = { name: string; phone: string; notify_when: string };

const MAX_ADMINS = 10;
const errorText: Record<string, string> = {
  WHATSAPP_ADMIN_LIMIT: "Maksimal 10 admin.",
  WHATSAPP_ADMIN_PHONE_INVALID: "Nomor tidak valid. Gunakan nomor WhatsApp, misalnya 0812 3456 7890.",
  WHATSAPP_ADMIN_NAME_INVALID: "Nama admin wajib diisi (maksimal 80 karakter).",
  WHATSAPP_ADMIN_WHEN_INVALID: "Deskripsi wajib diisi (maksimal 200 karakter).",
  WHATSAPP_ADMIN_DUPLICATE: "Nomor itu sudah ada di daftar.",
};

/** Early phone check mirroring the server (8-15 digits). The server stays the authority. */
function phoneError(value: string): string {
  const raw = value.trim();
  if (!raw) return "";
  if (!/^[+\d\s().-]+$/.test(raw)) return "Nomor hanya boleh berisi angka, spasi, atau tanda hubung.";
  const digits = raw.replace(/\D/g, "");
  return digits.length < 8 || digits.length > 15 ? "Nomor harus 8–15 digit." : "";
}

/** How the connected WhatsApp number serves visitors, and which admins the assistant routes other needs to. */
export function WhatsAppRouting({ tenantId, canEdit }: { tenantId: string; canEdit: boolean }) {
  const [mode, setMode] = useState<Mode | null>(null);
  const [admins, setAdmins] = useState<Admin[]>([]);
  const [linked, setLinked] = useState(true);
  const [draft, setDraft] = useState<Admin | null>(null);
  const [editing, setEditing] = useState(-1);
  const [busy, setBusy] = useState(false);
  const [working, setWorking] = useState<"mode" | "save" | "delete" | null>(null);
  const [notice, setNotice] = useState("");
  const [modalError, setModalError] = useState("");
  const base = `/api/v1/tenants/${tenantId}/whatsapp`;

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const channel = await api<{ service_mode?: Mode }>(base);
        const list = await api<{ data: Admin[] }>(`${base}/admins`);
        if (cancelled) return;
        setMode(channel?.service_mode === "GENERAL" ? "GENERAL" : "PSB_ONLY");
        setAdmins(list?.data ?? []);
      } catch (error) {
        if (cancelled) return;
        if (error instanceof ApiError && error.status === 404) setLinked(false);
        else setNotice("Pengaturan WhatsApp belum dapat dimuat.");
        setMode("PSB_ONLY");
      }
    })();
    return () => { cancelled = true; };
  }, [base]);

  const failure = (error: unknown, fallback: string) => error instanceof ApiError ? (errorText[error.message] ?? fallback) : fallback;

  async function changeMode(next: Mode) {
    if (!canEdit || busy || next === mode) return;
    setBusy(true); setWorking("mode"); setNotice("");
    try {
      await api(`${base}/service`, { method: "PUT", body: JSON.stringify({ service_mode: next }) });
      setMode(next);
      setNotice(next === "GENERAL" ? "Nomor ini sekarang melayani banyak keperluan." : "Nomor ini sekarang khusus PSB.");
    } catch (error) { setNotice(failure(error, "Mode layanan gagal diubah.")); }
    finally { setBusy(false); setWorking(null); }
  }

  /** Saves the whole list; the server rewrites the persona and restarts a running assistant. */
  async function persist(next: Admin[]): Promise<boolean> {
    try {
      const result = await api<{ data: Admin[] }>(`${base}/admins`, { method: "PUT", body: JSON.stringify({ data: next }) });
      setAdmins(result?.data ?? next);
      return true;
    } catch (error) { setModalError(failure(error, "Daftar admin gagal disimpan.")); return false; }
  }

  const open = (index: number) => { setModalError(""); setNotice(""); setEditing(index); setDraft(index >= 0 ? { ...admins[index] } : { name: "", phone: "", notify_when: "" }); };
  const close = () => { if (!busy) { setDraft(null); setEditing(-1); } };
  const draftPhoneProblem = draft ? phoneError(draft.phone) : "";

  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!draft || draftPhoneProblem) return;
    setBusy(true); setWorking("save"); setModalError("");
    const entry = { name: draft.name.trim(), phone: draft.phone.trim(), notify_when: draft.notify_when.trim() };
    const ok = await persist(editing >= 0 ? admins.map((item, i) => (i === editing ? entry : item)) : [...admins, entry]);
    setBusy(false); setWorking(null);
    if (ok) { setDraft(null); setEditing(-1); setNotice("Admin disimpan dan diteruskan ke asisten."); }
  }

  async function remove(index: number) {
    if (!window.confirm(`Hapus ${admins[index].name} dari daftar admin?`)) return;
    setBusy(true); setWorking("delete"); setNotice(""); setModalError("");
    const ok = await persist(admins.filter((_, i) => i !== index));
    setBusy(false); setWorking(null);
    setNotice(ok ? "Admin dihapus." : modalError || "Admin gagal dihapus.");
  }

  if (mode === null) return <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>;
  if (!linked) return <section className="panel panel-pad"><p className="empty">WhatsApp belum terhubung. Hubungkan nomor WhatsApp di menu Asisten AI terlebih dahulu.</p></section>;

  return <section className="panel panel-pad" style={{ display: "grid", gap: 20 }}>
    <div>
      <h2 className="section-title">Layanan nomor WhatsApp</h2>
      <div className="access-toggle" role="radiogroup" aria-label="Layanan nomor WhatsApp" style={{ marginTop: 10 }}>
        <button type="button" role="radio" aria-checked={mode === "PSB_ONLY"} className={mode === "PSB_ONLY" ? "active" : ""} disabled={!canEdit || busy} onClick={() => void changeMode("PSB_ONLY")}><strong>Khusus PSB</strong><span>Asisten langsung menjawab</span></button>
        <button type="button" role="radio" aria-checked={mode === "GENERAL"} className={mode === "GENERAL" ? "active" : ""} disabled={!canEdit || busy} onClick={() => void changeMode("GENERAL")}><strong>Umum</strong><span>Tanya keperluan dulu</span></button>
      </div>
      <span className="field-help" style={{ display: "block", marginTop: 8 }}>
        {mode === "GENERAL" ? "Asisten menanyakan keperluan pengunjung. PSB dijawab asisten; keperluan lain diteruskan ke admin di bawah, yang diberi tahu lewat WhatsApp. Tanpa admin, mode ini berjalan seperti Khusus PSB." : "Cocok untuk nomor khusus admin PSB: asisten langsung menjawab pertanyaan pengunjung."}
      </span>
    </div>

    {mode === "GENERAL" && <div>
      <div className="card-heading"><div><h2 className="section-title">Admin penerima notifikasi</h2><p className="section-hint">Saat pengunjung punya keperluan selain PSB, asisten memilih admin yang cocok lalu Hermes mengirim WhatsApp ke admin itu berisi nomor pengunjung dan keperluannya. Nomor admin tidak ditampilkan ke pengunjung.</p></div>
        {canEdit && <button className="button" type="button" disabled={busy || admins.length >= MAX_ADMINS} onClick={() => open(-1)}><Plus size={14} />Tambah admin</button>}</div>
      {admins.length === 0 && <p className="empty">Belum ada admin. Tanpa admin, asisten tidak meneruskan keperluan selain PSB.</p>}
      {admins.map((item, index) => <div key={`${item.phone}-${index}`} style={{ display: "flex", gap: 12, justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", padding: "12px 0", borderTop: "1px solid var(--line)" }}>
        <div style={{ minWidth: 0, flex: 1 }}>
          <strong>{item.name}</strong><span className="field-help"> · {item.phone}</span>
          <span className="field-help" style={{ display: "block" }}>Dihubungi jika: {item.notify_when}</span>
        </div>
        {canEdit && <div style={{ display: "flex", gap: 8 }}>
          <button className="button" type="button" disabled={busy} onClick={() => open(index)}><Pencil size={14} />Ubah</button>
          <button className="button" type="button" disabled={busy} onClick={() => void remove(index)}><Trash2 size={14} />Hapus</button>
        </div>}
      </div>)}
    </div>}
    {notice && <p className={/gagal|belum|tidak/.test(notice) ? "error" : "status"} role="status">{notice}</p>}

    {draft && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) close(); }}>
      <form className="modal panel panel-pad" onSubmit={save}>
        <div className="card-heading"><div><h2 className="section-title">{editing >= 0 ? "Ubah admin" : "Tambah admin"}</h2><p className="section-hint">Admin ini menerima WhatsApp saat asisten mengarahkan pengunjung kepadanya.</p></div><button type="button" className="account-action" onClick={close} aria-label="Tutup"><XCircle size={18} /></button></div>
        <label><span className="field-label">Nama</span>
          <input className="input" required maxLength={80} value={draft.name} onChange={event => setDraft({ ...draft, name: event.target.value })} placeholder="Ustadz Ahmad" />
          <span className="field-help">Pakai nama pendek. Asisten menyebut nama ini ke pengunjung.</span></label>
        <label><span className="field-label">Nomor WhatsApp</span>
          <input className="input" type="tel" required value={draft.phone} onChange={event => setDraft({ ...draft, phone: event.target.value })} placeholder="0812 3456 7890" aria-invalid={Boolean(draftPhoneProblem)} />
          {draftPhoneProblem ? <span className="error">{draftPhoneProblem}</span> : <span className="field-help">Nomor yang diawali 0 dibaca sebagai Indonesia (+62). Tidak ditampilkan ke pengunjung.</span>}</label>
        <label><span className="field-label">Dihubungi jika</span>
          <textarea className="input" required rows={3} maxLength={200} value={draft.notify_when} onChange={event => setDraft({ ...draft, notify_when: event.target.value })} placeholder="Surat-menyurat, legalisir ijazah, dan urusan keuangan" />
          <span className="field-help">Jelaskan keperluan yang ditangani admin ini. Asisten memakai deskripsi ini untuk memilih admin yang cocok.</span></label>
        {modalError && <div className="error" role="alert">{modalError}</div>}
        <div className="modal-actions"><button className="button" type="button" disabled={busy} onClick={close}>Batal</button><button className="button button-primary" disabled={busy || Boolean(draftPhoneProblem)}>{busy ? "Menyimpan…" : editing >= 0 ? "Simpan perubahan" : "Tambah admin"}</button></div>
      </form>
    </div>}

    {working && <ProcessingOverlay
      title={working === "mode" ? "Mengubah layanan WhatsApp" : working === "delete" ? "Menghapus admin" : "Menyimpan admin"}
      steps={working === "mode" ? ["Menyimpan pilihan layanan", "Memperbarui aturan asisten", "Memulai ulang asisten", "Hampir selesai, memastikan asisten berjalan"] : ["Menyimpan daftar admin", "Memperbarui aturan dan notifikasi asisten", "Memulai ulang asisten agar aturan baru berlaku", "Hampir selesai, memastikan asisten berjalan"]}
      note="Asisten dimulai ulang agar perubahan berlaku, jadi proses ini bisa memakan waktu sekitar 15–60 detik. Mohon jangan menutup atau memuat ulang halaman." />}
  </section>;
}
