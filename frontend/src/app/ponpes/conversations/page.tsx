"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { CheckCircle2, Clock3, Download, LifeBuoy, MessageSquare, RefreshCw, RotateCcw, Search, UsersRound, X } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { api } from "@/lib/api";
import { Pagination, usePagination } from "@/components/pagination";
import { can, chooseTenant, loadAuth, type AuthContext } from "@/lib/auth";

type Tenant = { id: string; name: string };
type Conversation = { id: string; username?: string; status: string; last_message_at: string; started_at?: string; snippet?: string };
type Message = { id: string; sender_type: string; message_text: string; created_at: string };
const pretty = (value: string) => value.toLowerCase().replaceAll("_", " ").replace(/\b\w/g, char => char.toUpperCase());
const statusLabel = (value: string) => ({ ACTIVE: "Aktif", HANDOFF: "Perlu bantuan", RESOLVED: "Selesai" } as Record<string, string>)[value] ?? pretty(value);
const dateTime = (value: string) => new Date(value).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });

export default function ConversationsPage() {
  const [items, setItems] = useState<Conversation[]>([]);
  const [tenantId, setTenantId] = useState("");
  const [tenantName, setTenantName] = useState("");
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [updating, setUpdating] = useState(false);
  const loadedOnce = useRef(false);

  async function load(search = query) {
    // Typing in the search box refreshes quietly instead of blanking the inbox.
    if (!loadedOnce.current) setLoading(true);
    setError("");
    try {
      const tenants = await api<{ data: Tenant[] }>("/api/v1/tenants");
      const list = tenants?.data ?? [];
      const id = chooseTenant(list);
      setTenantId(id);
      setTenantName(list.find(item => item.id === id)?.name ?? "");
      const params = new URLSearchParams(); if (id) params.set("tenant_id", id); if (search.trim()) params.set("q", search.trim());
      const result = await api<{ data: Conversation[] }>(`/api/v1/conversations${params.size ? `?${params}` : ""}`);
      setItems(result?.data ?? []);
    } catch { setError("Percakapan belum dapat dimuat."); } finally { loadedOnce.current = true; setLoading(false); }
  }
  useEffect(() => { void loadAuth().then(setAuth).catch(() => setAuth(null)); }, []);
  // Debounced: refetch from the server so the search also covers message text.
  useEffect(() => { const timer = window.setTimeout(() => void load(query), query ? 350 : 0); return () => window.clearTimeout(timer); }, [query]);

  useEffect(() => {
    if (!selected) { queueMicrotask(() => setMessages([])); return; }
    let cancelled = false;
    queueMicrotask(() => setMessagesLoading(true));
    api<{ data: Message[] }>(`/api/v1/conversations/${selected.id}/messages`)
      .then(result => { if (!cancelled) setMessages(result?.data ?? []); })
      .catch(() => { if (!cancelled) setMessages([]); })
      .finally(() => { if (!cancelled) setMessagesLoading(false); });
    return () => { cancelled = true; };
  }, [selected]);

  // On phones the thread opens as a full-screen sheet; Escape closes it.
  useEffect(() => {
    if (!selected) return;
    const onKey = (event: KeyboardEvent) => { if (event.key === "Escape") setSelected(null); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [selected]);
  const filtered = useMemo(() => items.filter(item => status === "all" || item.status === status), [items, status]);
  const paged = usePagination(filtered, 10, status);
  const active = items.filter(item => item.status === "ACTIVE").length;
  const handoff = items.filter(item => item.status === "HANDOFF").length;
  const canManage = can(auth, "conversation:manage");
  async function changeStatus(next: "ACTIVE" | "HANDOFF" | "RESOLVED") {
    if (!selected || updating) return;
    setUpdating(true); setError("");
    try {
      const updated = await api<Conversation>(`/api/v1/conversations/${selected.id}`, { method: "PATCH", body: JSON.stringify({ status: next }) });
      if (updated) { setItems(current => current.map(item => item.id === updated.id ? { ...item, status: updated.status } : item)); setSelected(current => current ? { ...current, status: updated.status } : current); }
    } catch { setError("Status percakapan gagal diperbarui."); } finally { setUpdating(false); }
  }
  const exportLog = () => { window.location.href = `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/tenants/${tenantId}/telegram/export`; };

  return <AppShell mode="ponpes" title="Percakapan">
    <PageHeader eyebrow="Telegram" title="Percakapan Telegram" description={tenantName ? `Pantau percakapan calon santri untuk ${tenantName}.` : "Pantau percakapan calon santri dengan asisten AI."} action={<><button className="button" onClick={exportLog} disabled={!tenantId || !items.length}><Download size={15} />Ekspor Excel</button><button className="button" onClick={() => void load()} disabled={loading}><RefreshCw size={15} className={loading ? "spin" : ""} />Muat ulang</button></>} />
    {error && <div className="notice notice-danger">{error}</div>}
    <section className="conversation-stats">
      <div><span className="conversation-stat-icon"><MessageSquare size={17} /></span><div><strong>{loading ? "—" : items.length}</strong><span>Total percakapan</span></div></div>
      <div><span className="conversation-stat-icon conversation-stat-green"><CheckCircle2 size={17} /></span><div><strong>{loading ? "—" : active}</strong><span>Sedang aktif</span></div></div>
      <div><span className="conversation-stat-icon conversation-stat-gold"><Clock3 size={17} /></span><div><strong>{loading ? "—" : handoff}</strong><span>Perlu bantuan admin</span></div></div>
      <div><span className="conversation-stat-icon conversation-stat-blue"><UsersRound size={17} /></span><div><strong>{loading ? "—" : new Set(items.map(item => item.username || item.id)).size}</strong><span>Pengguna Telegram</span></div></div>
    </section>
    <section className="conversation-workspace">
      <div className="panel conversation-inbox">
        <div className="conversation-heading"><div><h2 className="section-title">Inbox Telegram</h2><p className="section-hint">Pilih percakapan untuk membaca riwayat pesannya.</p></div><MessageSquare size={19} color="var(--brand)" /></div>
        <div className="conversation-toolbar">
          <div className="input-with-icon"><Search size={15} /><input className="input" value={query} onChange={event => setQuery(event.target.value)} placeholder="Cari username atau isi pesan…" aria-label="Cari percakapan" /></div>
          <select className="input" value={status} onChange={event => setStatus(event.target.value)} aria-label="Filter status"><option value="all">Semua status</option><option value="ACTIVE">Aktif</option><option value="HANDOFF">Perlu bantuan</option><option value="RESOLVED">Selesai</option></select>
        </div>
        {loading ? <div className="conversation-empty">Memuat percakapan…</div> : filtered.length ? <><div className="conversation-directory">{paged.rows.map(item => <button className={`conversation-row${selected?.id === item.id ? " selected" : ""}`} type="button" key={item.id} onClick={() => setSelected(item)} aria-pressed={selected?.id === item.id}><span className="conversation-avatar">{(item.username?.[0] ?? "U").toUpperCase()}</span><span className="conversation-person"><strong>{item.username ? `@${item.username}` : "Pengguna Telegram"}</strong><span>Pesan terakhir {dateTime(item.last_message_at)}</span>{item.snippet && <span className="conversation-snippet">“{item.snippet}”</span>}</span><span className={`status ${item.status === "HANDOFF" ? "status-warning" : ""}`}>{statusLabel(item.status)}</span></button>)}</div><Pagination paged={paged} label="percakapan"/></> : <div className="conversation-empty"><MessageSquare size={30} /><strong>Belum ada percakapan</strong><span>Percakapan akan muncul setelah calon santri menghubungi bot Telegram.</span></div>}
      </div>
      <aside className={`panel conversation-detail${selected ? " open" : ""}`} role={selected ? "dialog" : undefined} aria-label={selected ? "Riwayat percakapan" : undefined}>
        {!selected ? <><div className="conversation-detail-icon"><MessageSquare size={25} /></div><h2>Belum ada percakapan dipilih</h2><p>Pilih percakapan di inbox untuk melihat riwayat pesan dan jawaban asisten.</p></> : <>
          <div className="conversation-thread-head"><button className="conversation-close" type="button" onClick={() => setSelected(null)} aria-label="Tutup percakapan"><X size={18} /></button><strong>{selected.username ? `@${selected.username}` : "Pengguna Telegram"}</strong><span>Dimulai {dateTime(selected.started_at ?? selected.last_message_at)} · {statusLabel(selected.status)}</span>{canManage && <div className="conversation-actions">{selected.status === "ACTIVE" && <button className="button" type="button" disabled={updating} onClick={() => void changeStatus("HANDOFF")}><LifeBuoy size={14} />Tandai perlu bantuan</button>}{selected.status !== "RESOLVED" && <button className="button" type="button" disabled={updating} onClick={() => void changeStatus("RESOLVED")}><CheckCircle2 size={14} />Tandai selesai</button>}{selected.status !== "ACTIVE" && <button className="button button-ghost" type="button" disabled={updating} onClick={() => void changeStatus("ACTIVE")}><RotateCcw size={14} />{selected.status === "HANDOFF" ? "Kembalikan ke asisten" : "Buka kembali"}</button>}</div>}</div>
          {messagesLoading ? <p className="section-hint">Memuat pesan…</p> : messages.length ? <div className="conversation-thread">{messages.map(message => <div key={message.id} className={`chat-bubble ${message.sender_type === "USER" ? "from-user" : "from-bot"}`}><span className="chat-sender">{message.sender_type === "USER" ? "Pengguna" : message.sender_type === "BOT" ? "Asisten" : "Admin"}</span><p>{message.message_text}</p><time>{dateTime(message.created_at)}</time></div>)}</div> : <p className="section-hint">Belum ada pesan pada percakapan ini.</p>}
        </>}
      </aside>
    </section>
  </AppShell>;
}
