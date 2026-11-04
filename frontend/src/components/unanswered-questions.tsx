"use client";

import { useEffect, useState } from "react";
import { LoaderCircle, MessageCircleQuestion, Pencil, Trash2 } from "lucide-react";
import { api, ApiError } from "@/lib/api";

type Status = "OPEN" | "ANSWERED" | "RESOLVED" | "IGNORED";
type Question = { id: string; question: string; ask_count: number; status: Status; faq_question: string; answer: string; first_asked_at: string; last_asked_at: string };

const tabs: { status: Status; label: string; empty: string }[] = [
  { status: "OPEN", label: "Belum terjawab", empty: "Tidak ada pertanyaan yang menunggu. Pertanyaan yang tidak bisa dijawab asisten akan muncul di sini." },
  { status: "ANSWERED", label: "Sudah dijawab", empty: "Belum ada pertanyaan yang Anda jawab." },
  { status: "RESOLVED", label: "Selesai", empty: "Belum ada pertanyaan yang ditandai selesai." },
  { status: "IGNORED", label: "Diabaikan", empty: "Tidak ada pertanyaan yang diabaikan." },
];
const errorText: Record<string, string> = {
  QUESTION_LIMIT: "Batas 100 jawaban tercapai. Hapus jawaban lama terlebih dahulu.",
  QUESTION_NOT_FOUND: "Pertanyaan tidak ditemukan. Muat ulang halaman.",
};
const when = (value: string) => new Date(value).toLocaleString("id-ID", { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" });

/** Questions the assistant could not answer. An admin answers them (the answer goes to the assistant), or marks them resolved
 *  when the fix was something else, such as uploading a document. */
export function UnansweredQuestions({ tenantId, canEdit, onOpenCount }: { tenantId: string; canEdit: boolean; onOpenCount?: (count: number) => void }) {
  const [status, setStatus] = useState<Status>("OPEN");
  const [items, setItems] = useState<Question[] | null>(null);
  const [counts, setCounts] = useState<Partial<Record<Status, number>>>({});
  const [answering, setAnswering] = useState("");
  const [faq, setFaq] = useState("");
  const [answer, setAnswer] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);

  const [reload, setReload] = useState(0);
  const load = () => setReload(value => value + 1);
  useEffect(() => {
    if (!tenantId) return;
    let cancelled = false;
    void (async () => {
      try {
        const result = await api<{ data: Question[]; counts: Partial<Record<Status, number>> }>(`/api/v1/tenants/${tenantId}/questions?status=${status}`);
        if (!cancelled) { setItems(result?.data ?? []); setCounts(result?.counts ?? {}); onOpenCount?.(result?.counts?.OPEN ?? 0); }
      } catch { if (!cancelled) { setItems([]); setNotice("Daftar pertanyaan belum dapat dimuat."); } }
    })();
    return () => { cancelled = true; };
  }, [tenantId, status, reload, onOpenCount]);

  const fail = (error: unknown, fallback: string) => setNotice(error instanceof ApiError ? (errorText[error.message] ?? fallback) : fallback);
  const base = `/api/v1/tenants/${tenantId}/questions`;

  async function send(item: Question, body: object, done: string) {
    setBusy(true); setNotice("");
    try { await api(`${base}/${item.id}`, { method: "PATCH", body: JSON.stringify(body) }); setAnswering(""); setNotice(done); load(); }
    catch (error) { fail(error, "Perubahan gagal disimpan."); }
    finally { setBusy(false); }
  }

  async function remove(item: Question) {
    if (!window.confirm(item.status === "ANSWERED" ? "Hapus pertanyaan ini? Asisten tidak akan memakai jawabannya lagi." : "Hapus pertanyaan ini dari daftar?")) return;
    setBusy(true); setNotice("");
    try { await api(`${base}/${item.id}`, { method: "DELETE" }); setNotice("Pertanyaan dihapus."); load(); }
    catch (error) { fail(error, "Pertanyaan gagal dihapus."); }
    finally { setBusy(false); }
  }

  const startAnswer = (item: Question) => { setNotice(""); setAnswering(item.id); setFaq(item.faq_question || item.question); setAnswer(item.answer); };
  const tab = tabs.find(item => item.status === status)!;

  return <section className="panel panel-pad" style={{ display: "grid", gap: 16 }}>
    <div><h2 className="section-title">Pertanyaan belum terjawab</h2><p className="section-hint">Pertanyaan pengunjung yang tidak bisa dijawab asisten dikumpulkan di sini. Jawab langsung agar asisten bisa menjawabnya lain kali, atau tandai selesai jika jawabannya sudah ditambahkan lewat dokumen atau sumber web.</p></div>
    <div style={{ display: "grid", gap: 16 }}>
      <div role="tablist" style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
        {tabs.map(item => <button key={item.status} role="tab" aria-selected={status === item.status} className={status === item.status ? "button button-primary" : "button"} type="button" onClick={() => { setItems(null); setAnswering(""); setStatus(item.status); }}>
          {item.label}{typeof counts[item.status] === "number" && ` (${counts[item.status]})`}</button>)}
      </div>

      {notice && <p className={/disimpan|dihapus|ditandai|dijawab|dibuka/.test(notice) ? "status" : "error"} role="status">{notice}</p>}
      <div>
        {items === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
        {items?.length === 0 && <p className="empty">{tab.empty}</p>}
        {items?.map(item => <div key={item.id} style={{ display: "grid", gap: 10, padding: "14px 0", borderTop: "1px solid var(--line)" }}>
          <div style={{ display: "flex", gap: 12, justifyContent: "space-between", alignItems: "flex-start", flexWrap: "wrap" }}>
            <div style={{ minWidth: 0, flex: 1 }}>
              <strong style={{ overflowWrap: "anywhere" }}><MessageCircleQuestion size={15} style={{ verticalAlign: "-2px", marginRight: 6 }} aria-hidden />{item.question}</strong>
              <span className="field-help" style={{ display: "block" }}>Ditanyakan {item.ask_count}× · terakhir {when(item.last_asked_at)}</span>
              {item.status === "ANSWERED" && answering !== item.id && <span className="field-help" style={{ display: "block", marginTop: 6 }}><strong>Jawaban asisten:</strong> {item.answer}</span>}
            </div>
            {canEdit && answering !== item.id && <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              {item.status === "OPEN" && <>
                <button className="button button-primary" type="button" disabled={busy} onClick={() => startAnswer(item)}><Pencil size={14} />Jawab</button>
                <button className="button" type="button" disabled={busy} title="Pilih ini jika Anda sudah menambahkan jawabannya lewat dokumen atau sumber web" onClick={() => void send(item, { status: "RESOLVED" }, "Pertanyaan ditandai selesai.")}>Tandai selesai</button>
                <button className="button" type="button" disabled={busy} onClick={() => void send(item, { status: "IGNORED" }, "Pertanyaan diabaikan.")}>Abaikan</button>
              </>}
              {item.status === "ANSWERED" && <button className="button" type="button" disabled={busy} onClick={() => startAnswer(item)}><Pencil size={14} />Ubah jawaban</button>}
              {(item.status === "RESOLVED" || item.status === "IGNORED") && <button className="button" type="button" disabled={busy} onClick={() => void send(item, { status: "OPEN" }, "Pertanyaan dibuka kembali.")}>Buka lagi</button>}
              <button className="button" type="button" disabled={busy} onClick={() => void remove(item)}><Trash2 size={14} />Hapus</button>
            </div>}
          </div>
          {answering === item.id && canEdit && <form style={{ display: "grid", gap: 12 }} onSubmit={event => { event.preventDefault(); void send(item, { faq_question: faq, answer }, "Jawaban disimpan dan diteruskan ke asisten."); }}>
            <label><span className="field-label">Pertanyaan (versi umum)</span>
              <input className="input" required maxLength={200} value={faq} onChange={event => setFaq(event.target.value)} />
              <span className="field-help">Tulis ulang secara umum dan hapus nama atau nomor pengunjung. Hanya teks ini dan jawabannya yang diberikan ke asisten.</span></label>
            <label><span className="field-label">Jawaban</span>
              <textarea className="input" required rows={4} maxLength={1500} value={answer} onChange={event => setAnswer(event.target.value)} placeholder="Tulis jawaban yang benar dan lengkap" /></label>
            <div style={{ display: "flex", gap: 8 }}>
              <button className="button button-primary" disabled={busy || !answer.trim() || !faq.trim()}>{busy ? "Menyimpan…" : "Simpan jawaban"}</button>
              <button className="button" type="button" disabled={busy} onClick={() => setAnswering("")}>Batal</button>
            </div>
          </form>}
        </div>)}
      </div>
    </div>
  </section>;
}
