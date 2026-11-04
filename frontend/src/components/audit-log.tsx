"use client";

import { useEffect, useMemo, useState } from "react";
import { Activity, Filter, Search } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { api, ApiError } from "@/lib/api";
import { auditActionLabel, auditActorLabel, auditTargetLabel } from "@/lib/labels";
import { chooseTenant } from "@/lib/auth";
import { Pagination, usePagination } from "@/components/pagination";

type AuditRecord = { id?: string; action?: string; tenant_id?: string; occurred_at?: string; actor_type?: string; actor_name?: string; target_type?: string; target_name?: string };
const displayAction = auditActionLabel;
const auditErrorMessage = (status: number) => status === 400 ? "Audit sistem hanya untuk superadmin. Buka menu Aktivitas di dashboard pesantren Anda." : status === 403 ? "Anda tidak punya akses ke riwayat aktivitas ini." : "Riwayat aktivitas belum dapat dimuat.";
const displayTime = (value?: string) => { if (!value) return "Waktu tidak tersedia"; const date = new Date(value); return Number.isNaN(date.valueOf()) ? value : new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(date); };

export function AuditLog({ mode = "platform" }: { mode?: "platform" | "ponpes" }) {
  const [logs, setLogs] = useState<AuditRecord[]>([]); const [query, setQuery] = useState(""); const [actor, setActor] = useState("all"); const [target, setTarget] = useState("all"); const [state, setState] = useState<"loading" | "ready" | "error">("loading"); const [errorMessage, setErrorMessage] = useState("");
  const load = async () => { setState("loading"); try { let scope = "";
      if (mode === "ponpes") {
        // Pesantren admins only ever see their own pesantren's activity.
        const tenants = await api<{ data?: { id: string }[] }>("/api/v1/tenants");
        const id = chooseTenant(tenants?.data ?? []);
        if (!id) { setLogs([]); setState("ready"); return; }
        scope = `?tenant_id=${id}`;
      }
      const body = await api<{ data?: AuditRecord[] }>(`/api/v1/audit-logs${scope}`); setLogs(body?.data ?? []); setState("ready"); } catch (error) { setErrorMessage(error instanceof ApiError ? auditErrorMessage(error.status) : "Riwayat aktivitas belum dapat dimuat."); setState("error"); } };
  useEffect(() => { queueMicrotask(() => void load()); }, []);
  const actors = useMemo(() => [...new Set(logs.map(log => log.actor_type).filter(Boolean))] as string[], [logs]);
  const targets = useMemo(() => [...new Set(logs.map(log => log.target_type).filter(Boolean))] as string[], [logs]);
  const filtered = useMemo(() => logs.filter(log => { const text = `${log.action} ${log.tenant_id} ${log.actor_type} ${log.actor_name} ${log.target_type} ${log.target_name}`.toLowerCase(); return text.includes(query.toLowerCase()) && (actor === "all" || log.actor_type === actor) && (target === "all" || log.target_type === target); }), [logs, query, actor, target]);
  const paged = usePagination(filtered, 10, `${query}|${actor}|${target}`);
  return <AppShell mode={mode} title={mode === "platform" ? "Audit sistem" : "Aktivitas sistem"}><PageHeader eyebrow={mode === "platform" ? "Keamanan & kepatuhan" : "Riwayat ruang kerja"} title={mode === "platform" ? "Audit log sistem" : "Riwayat aktivitas"} description={mode === "platform" ? "Catatan perubahan administrator, worker, dan sistem. Percakapan Telegram tersedia di menu Percakapan." : "Catatan perubahan yang dilakukan pengguna dan sistem di pesantren Anda. Percakapan Telegram ada di menu Percakapan."} action={<button className="button" onClick={() => void load()}>Muat ulang</button>} /><section className="stats"><div className="stat"><span className="stat-label">Total aktivitas</span><strong className="stat-value">{state === "ready" ? logs.length : "—"}</strong><span className="stat-note">Dalam hasil saat ini</span></div><div className="stat"><span className="stat-label">Hasil pencarian</span><strong className="stat-value">{state === "ready" ? filtered.length : "—"}</strong><span className="stat-note">Aktivitas yang tampil</span></div><div className="stat"><span className="stat-label">Cakupan akses</span><strong className="stat-value">{mode === "platform" ? "Platform" : "Pesantren"}</strong><span className="stat-note">Mengikuti peran Anda</span></div><div className="stat"><span className="stat-label">Status</span><strong className="stat-value">{state === "ready" ? "Tersedia" : state === "loading" ? "Memuat" : "Gagal dimuat"}</strong><span className="stat-note">Hanya untuk dibaca</span></div></section><section className="panel panel-pad"><div className="audit-toolbar"><div className="input-with-icon audit-search"><Search size={15}/><input className="input" value={query} onChange={event => setQuery(event.target.value)} placeholder="Cari aktivitas, pengguna, atau objek…" aria-label="Cari riwayat aktivitas"/></div><select className="input audit-filter" value={actor} onChange={event => setActor(event.target.value)} aria-label="Filter pelaku"><option value="all">Semua pelaku</option>{actors.map(item => <option key={item} value={item}>{auditActorLabel(item)}</option>)}</select><select className="input audit-filter" value={target} onChange={event => setTarget(event.target.value)} aria-label="Filter objek"><option value="all">Semua objek</option>{targets.map(item => <option key={item} value={item}>{auditTargetLabel(item)}</option>)}</select><Filter size={17} color="var(--faint)"/></div>{state === "error" ? <div className="empty">{errorMessage}</div> : filtered.length ? <><div className="log-list">{paged.rows.map(log => <div className="log-item" key={log.id ?? `${log.action}-${log.occurred_at}`}><div className="log-dot"><Activity size={13}/></div><div><strong>{displayAction(log.action)}</strong><span>{log.actor_name || auditActorLabel(log.actor_type)} · {log.target_name || auditTargetLabel(log.target_type)}</span></div><time>{displayTime(log.occurred_at)}</time></div>)}</div><Pagination paged={paged} label="aktivitas"/></> : <div className="empty">Tidak ada aktivitas yang cocok dengan filter.</div>}</section></AppShell>;
}
