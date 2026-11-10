"use client";

import { useEffect, useMemo, useState } from "react";
import { Download, Search, XCircle } from "lucide-react";

export type ExportConversation = { id: string; username?: string; last_message_at: string };

type Range = "all" | "today" | "7" | "30" | "custom";

const ranges: { value: Range; label: string }[] = [
  { value: "all", label: "Semua tanggal" },
  { value: "today", label: "Hari ini" },
  { value: "7", label: "7 hari terakhir" },
  { value: "30", label: "30 hari terakhir" },
  { value: "custom", label: "Pilih tanggal" },
];

function jakartaDate(offsetDays = 0) {
  const today = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Jakarta", year: "numeric", month: "2-digit", day: "2-digit" }).format(new Date());
  const [year, month, day] = today.split("-").map(Number);
  return new Date(Date.UTC(year, month - 1, day + offsetDays)).toISOString().slice(0, 10);
}

const when = (value: string) => new Date(value).toLocaleString("id-ID", { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });
const person = (item: ExportConversation) => item.username ? `@${item.username}` : "Pengguna Telegram";

/** Asks which dates and conversations to include, then downloads the Telegram Excel log. */
export function TelegramExportButton({ tenantId, conversations, disabled, label = "Export Excel", ghost = false }: { tenantId: string; conversations: ExportConversation[]; disabled?: boolean; label?: string; ghost?: boolean }) {
  const [open, setOpen] = useState(false);
  const [range, setRange] = useState<Range>("all");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [scope, setScope] = useState<"all" | "picked">("all");
  const [picked, setPicked] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [notice, setNotice] = useState("");

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => { if (event.key === "Escape") setOpen(false); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open]);

  const visible = useMemo(() => {
    const term = query.trim().toLowerCase();
    return term ? conversations.filter(item => person(item).toLowerCase().includes(term)) : conversations;
  }, [conversations, query]);

  function openModal() {
    setRange("all"); setFrom(""); setTo(""); setScope("all"); setPicked([]); setQuery(""); setNotice(""); setOpen(true);
  }

  function dates(): { from: string; to: string } | null {
    if (range === "all") return { from: "", to: "" };
    if (range === "today") return { from: jakartaDate(), to: jakartaDate() };
    if (range === "7") return { from: jakartaDate(-6), to: jakartaDate() };
    if (range === "30") return { from: jakartaDate(-29), to: jakartaDate() };
    if (!from && !to) { setNotice("Isi tanggal mulai, tanggal akhir, atau keduanya."); return null; }
    if (from && to && from > to) { setNotice("Tanggal mulai tidak boleh setelah tanggal akhir."); return null; }
    return { from, to };
  }

  function download() {
    const windowDates = dates();
    if (!windowDates) return;
    if (scope === "picked" && picked.length === 0) { setNotice("Pilih minimal satu percakapan."); return; }
    const params = new URLSearchParams();
    if (windowDates.from) params.set("from", windowDates.from);
    if (windowDates.to) params.set("to", windowDates.to);
    if (scope === "picked") picked.forEach(id => params.append("conversation", id));
    const queryText = params.toString();
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/tenants/${tenantId}/telegram/export${queryText ? `?${queryText}` : ""}`;
    setOpen(false);
  }

  function toggle(id: string) {
    setPicked(current => current.includes(id) ? current.filter(item => item !== id) : [...current, id]);
  }

  const allVisible = visible.length > 0 && visible.every(item => picked.includes(item.id));

  return <>
    <button className={ghost ? "button button-ghost" : "button"} type="button" onClick={openModal} disabled={disabled || !tenantId}><Download size={ghost ? 14 : 15} />{label}</button>
    {open && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) setOpen(false); }}>
      <div className="modal panel panel-pad telegram-export-modal" role="dialog" aria-modal="true" aria-labelledby="telegram-export-title">
        <div className="card-heading">
          <div><h2 className="section-title" id="telegram-export-title">Ekspor log Telegram</h2><p className="section-hint">Pilih tanggal dan percakapan yang ingin diunduh. Tanggal mengikuti waktu WIB.</p></div>
          <button className="account-action" type="button" onClick={() => setOpen(false)} aria-label="Tutup"><XCircle size={18} /></button>
        </div>
        <label><span className="field-label">Tanggal</span>
          <select className="input" value={range} onChange={event => { setNotice(""); setRange(event.target.value as Range); }}>
            {ranges.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select>
        </label>
        {range === "custom" && <div className="telegram-export-dates">
          <label><span className="field-label">Dari</span><input className="input" type="date" value={from} onChange={event => { setNotice(""); setFrom(event.target.value); }} /></label>
          <label><span className="field-label">Sampai</span><input className="input" type="date" value={to} onChange={event => { setNotice(""); setTo(event.target.value); }} /></label>
        </div>}
        <label><span className="field-label">Percakapan</span>
          <select className="input" value={scope} onChange={event => { setNotice(""); setScope(event.target.value as "all" | "picked"); }}>
            <option value="all">Semua percakapan</option>
            <option value="picked">Pilih percakapan</option>
          </select>
        </label>
        {scope === "picked" && <div className="telegram-export-chats">
          <div className="input-with-icon"><Search size={15} /><input className="input" value={query} onChange={event => setQuery(event.target.value)} placeholder="Cari username…" aria-label="Cari percakapan" /></div>
          <div className="telegram-export-list">
            {visible.length === 0 && <p className="empty">Tidak ada percakapan yang cocok.</p>}
            {visible.map(item => <label key={item.id}><input type="checkbox" checked={picked.includes(item.id)} onChange={() => { setNotice(""); toggle(item.id); }} /><span><strong>{person(item)}</strong><small>Terakhir {when(item.last_message_at)}</small></span></label>)}
          </div>
          <div className="telegram-export-tools">
            <button className="button button-ghost" type="button" disabled={!visible.length} onClick={() => setPicked(current => allVisible ? current.filter(id => !visible.some(item => item.id === id)) : [...new Set([...current, ...visible.map(item => item.id)])])}>{allVisible ? "Hapus pilihan yang tampil" : "Pilih yang tampil"}</button>
            <span>{picked.length} dipilih</span>
          </div>
        </div>}
        {notice && <p className="error" role="status">{notice}</p>}
        <div className="modal-actions">
          <button className="button" type="button" onClick={() => setOpen(false)}>Batal</button>
          <button className="button button-primary" type="button" onClick={download}><Download size={15} />Unduh Excel</button>
        </div>
      </div>
    </div>}
  </>;
}
