"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowLeft, Eye, EyeOff, LoaderCircle, Pencil, RefreshCw, X } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { api, ApiError } from "@/lib/api";
import { can, chooseTenant, loadAuth, type AuthContext } from "@/lib/auth";
import { diffLines } from "@/lib/diff";

type Tenant = { id: string; name: string };
type Source = { id: string; url: string; enabled: boolean; last_status?: string; last_run_at?: string; error_public_message?: string };
type EventKind = "NEW" | "CHANGED" | "SOURCE_CHANGED" | "GONE";
type CrawlEvent = { id: string; page_id?: string; url: string; title: string; event: EventKind; lines_added: number; lines_removed: number; before_text?: string; after_text?: string; created_at: string };
type Run = { id: string; trigger: "SCHEDULED" | "MANUAL"; status: string; pages_new: number; pages_changed: number; pages_unchanged: number; pages_failed: number; error_public_message?: string; created_at: string; finished_at?: string };
type Page = { id: string; url: string; title: string; status: "ACTIVE" | "GONE" | "EXCLUDED"; editable: boolean; user_edited: boolean; update_available: boolean; last_changed_at?: string; last_seen_at?: string; body?: string; source_text?: string };
type Tab = "events" | "pages" | "runs";

const eventLabel: Record<EventKind, { text: string; className: string }> = {
  NEW: { text: "Halaman baru", className: "status" },
  CHANGED: { text: "Isi berubah", className: "status status-muted" },
  SOURCE_CHANGED: { text: "Sumber berubah — perlu keputusan", className: "status status-danger" },
  GONE: { text: "Dihapus dari asisten", className: "status status-danger" },
};
const runStatus: Record<string, string> = { QUEUED: "Menunggu", RUNNING: "Berjalan", SUCCEEDED: "Berhasil", PARTIAL: "Sebagian", FAILED: "Gagal", NEEDS_PUBLIC_SHARING: "Dokumen belum publik" };
const errorText: Record<string, string> = {
  CRAWL_RUN_TOO_SOON: "Tunggu 10 menit sebelum mengambil ulang sumber yang sama.",
  CRAWL_ALREADY_RUNNING: "Pengambilan sedang berjalan untuk sumber ini.",
  CRAWL_PAGE_NOT_EDITABLE: "Halaman ini tidak dapat diedit.",
  CRAWL_PAGE_FILE_MISSING: "Berkas halaman ini sudah dihapus dari Dokumen. Ambil ulang sumber untuk membuatnya kembali.",
  INVALID_REQUEST: "Teks tidak boleh kosong atau lebih dari 512 KB.",
};

const when = (value?: string) => value ? new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short", timeZone: "Asia/Jakarta" }).format(new Date(value)) : "—";
const pageName = (item: { title: string; url: string }) => item.title || item.url;

/** Lists the lines that were added or removed, so a change is understandable without comparing two walls of text. */
function DiffView({ before, after }: { before: string; after: string }) {
  const { added, removed } = diffLines(before, after);
  if (!added.length && !removed.length) return <p className="section-hint">Tidak ada perbedaan isi yang dapat ditampilkan.</p>;
  const block = (lines: string[], title: string, background: string, color: string) => lines.length > 0 && <div style={{ marginTop: 12 }}>
    <span className="field-label">{title} ({lines.length})</span>
    <div style={{ maxHeight: 220, overflow: "auto", padding: 10, borderRadius: 8, background, color, fontSize: 12, lineHeight: 1.6, whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
      {lines.slice(0, 200).map((line, index) => <div key={index}>{line}</div>)}
      {lines.length > 200 && <div>… dan {lines.length - 200} baris lain</div>}
    </div>
  </div>;
  return <>{block(added, "Ditambahkan", "#f1faf4", "#256b49")}{block(removed, "Dihapus", "#fff6f5", "#a23b32")}</>;
}

export default function SourceResultsPage() {
  // useParams suspends under cacheComponents for params not known at build time.
  return <Suspense fallback={null}><SourceResults /></Suspense>;
}

function SourceResults() {
  const { sid } = useParams<{ sid: string }>();
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [tenantId, setTenantId] = useState("");
  const [source, setSource] = useState<Source | null>(null);
  const [events, setEvents] = useState<CrawlEvent[] | null>(null);
  const [runs, setRuns] = useState<Run[] | null>(null);
  const [pages, setPages] = useState<Page[] | null>(null);
  const [tab, setTab] = useState<Tab>("events");
  const [notice, setNotice] = useState("");
  const [diff, setDiff] = useState<CrawlEvent | null>(null);
  const [editing, setEditing] = useState<Page | null>(null);
  const [draft, setDraft] = useState("");
  const [busy, setBusy] = useState(false);

  const canEdit = can(auth, "tenant:update");
  const base = `/api/v1/tenants/${tenantId}/crawl`;

  const load = useCallback(async (id: string) => {
    const root = `/api/v1/tenants/${id}/crawl`;
    try {
      const [list, ev, rn, pg] = await Promise.all([
        api<{ data: Source[] }>(`/api/v1/tenants/${id}/crawl/sources`),
        api<{ data: CrawlEvent[] }>(`${root}/sources/${sid}/events`),
        api<{ data: Run[] }>(`${root}/sources/${sid}/runs`),
        api<{ data: Page[] }>(`${root}/sources/${sid}/pages`),
      ]);
      setSource((list?.data ?? []).find(item => item.id === sid) ?? null);
      setEvents(ev?.data ?? []); setRuns(rn?.data ?? []); setPages(pg?.data ?? []);
    } catch { setEvents([]); setRuns([]); setPages([]); setNotice("Hasil pengambilan belum dapat dimuat."); }
  }, [sid]);

  useEffect(() => {
    void (async () => {
      try {
        const [authData, tenants] = await Promise.all([loadAuth(), api<{ data: Tenant[] }>("/api/v1/tenants")]);
        setAuth(authData);
        const id = chooseTenant(tenants?.data ?? []);
        if (!id) { setEvents([]); setRuns([]); setPages([]); return; }
        setTenantId(id);
        await load(id);
      } catch { setEvents([]); setRuns([]); setPages([]); setNotice("Halaman belum dapat dimuat."); }
    })();
  }, [load]);

  // Keep the page fresh while a crawl is running.
  const running = source?.last_status === "RUNNING" || source?.last_status === "QUEUED";
  useEffect(() => {
    if (!running || !tenantId) return;
    const timer = window.setTimeout(() => void load(tenantId), 4000);
    return () => window.clearTimeout(timer);
  }, [running, source, tenantId, load]);

  const fail = (error: unknown, fallback: string) => setNotice(error instanceof ApiError ? (errorText[error.message] ?? fallback) : fallback);

  async function runNow() {
    setNotice("");
    try { await api(`${base}/sources/${sid}/run`, { method: "POST" }); await load(tenantId); } catch (error) { fail(error, "Pengambilan gagal dimulai."); }
  }

  async function openDiff(event: CrawlEvent) {
    try { setDiff(await api<CrawlEvent>(`${base}/events/${event.id}`)); } catch { setNotice("Perbandingan belum dapat dimuat."); }
  }

  async function openEditor(page: Page) {
    try {
      const full = await api<Page>(`${base}/pages/${page.id}`);
      if (!full) return;
      setEditing(full); setDraft(full.body ?? "");
    } catch { setNotice("Halaman belum dapat dibuka."); }
  }

  async function mutate(action: () => Promise<Page | null>, success: string, failure: string) {
    setBusy(true); setNotice("");
    try {
      const updated = await action();
      if (updated && editing) { setEditing(updated.editable ? updated : null); setDraft(updated.body ?? ""); }
      setNotice(success);
      await load(tenantId);
    } catch (error) { fail(error, failure); }
    finally { setBusy(false); }
  }

  const save = () => editing && mutate(() => api<Page>(`${base}/pages/${editing.id}/content`, { method: "PUT", body: JSON.stringify({ text: draft }) }), "Perubahan disimpan. Asisten memakai teks ini setelah disinkronkan.", "Perubahan gagal disimpan.");
  const resolve = (action: "use_source" | "keep_mine") => editing && mutate(() => api<Page>(`${base}/pages/${editing.id}/resolve`, { method: "POST", body: JSON.stringify({ action }) }), action === "use_source" ? "Halaman dikembalikan ke teks terbaru dari sumber." : "Versi Anda dipertahankan.", "Pilihan gagal diterapkan.");
  const setExcluded = (page: Page, excluded: boolean) => mutate(() => api<Page>(`${base}/pages/${page.id}`, { method: "PATCH", body: JSON.stringify({ excluded }) }), excluded ? "Halaman dikeluarkan dari pengetahuan asisten." : "Halaman dimasukkan kembali.", "Perubahan gagal diterapkan.");

  const waiting = pages?.filter(page => page.update_available).length ?? 0;
  const tabs: [Tab, string][] = [["events", "Pembaruan"], ["pages", `Halaman${pages ? ` (${pages.length})` : ""}`], ["runs", "Riwayat"]];

  return <AppShell mode="ponpes" title="Sumber Web">
    <PageHeader eyebrow="Sumber web" title="Hasil pengambilan" description={source?.url ?? "Lihat apa yang dipelajari asisten dari sumber ini, dan koreksi bila perlu."}
      action={<Link className="button" href="/ponpes/sources"><ArrowLeft size={14} />Kembali</Link>} />
    {notice && <p className={/disimpan|dikembalikan|dipertahankan|dikeluarkan|dimasukkan/.test(notice) ? "status" : "error"} role="status" style={{ marginBottom: 16 }}>{notice}</p>}
    {waiting > 0 && <div className="notice notice-danger" role="alert"><span>{waiting} halaman yang Anda edit berubah di sumbernya. Buka tab Halaman untuk memilih versi yang dipakai.</span></div>}
    <div className="panel panel-pad">
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap", alignItems: "center" }}>
        <div className="detail-tabs" style={{ margin: 0, borderBottom: 0 }}>{tabs.map(([key, label]) => <button key={key} type="button" className={tab === key ? "active" : ""} onClick={() => setTab(key)}>{label}</button>)}</div>
        {canEdit && <button className="button" type="button" disabled={!source?.enabled || running} onClick={() => void runNow()}>{running ? <LoaderCircle size={14} className="spin" /> : <RefreshCw size={14} />}Ambil sekarang</button>}
      </div>

      {tab === "events" && <div style={{ marginTop: 16, display: "grid", gap: 10 }}>
        {events === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
        {events?.length === 0 && <p className="empty">Belum ada pembaruan. Perubahan akan muncul di sini setelah sumber diambil.</p>}
        {events?.map(event => <div key={event.id} style={{ display: "flex", gap: 12, justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", padding: "10px 0", borderTop: "1px solid var(--line)" }}>
          <div style={{ minWidth: 0, flex: 1 }}>
            <span className={eventLabel[event.event].className}>{eventLabel[event.event].text}</span>
            <strong style={{ display: "block", marginTop: 6, overflowWrap: "anywhere" }}>{pageName(event)}</strong>
            <span className="field-help">{when(event.created_at)} · +{event.lines_added} / −{event.lines_removed} baris</span>
          </div>
          {event.event !== "NEW" || event.lines_added > 0 ? <button className="button" type="button" onClick={() => void openDiff(event)}><Eye size={14} />Lihat perubahan</button> : null}
        </div>)}
      </div>}

      {tab === "pages" && <div style={{ marginTop: 16, display: "grid", gap: 10 }}>
        {pages === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
        {pages?.length === 0 && <p className="empty">Belum ada halaman terkumpul.</p>}
        {pages?.map(page => <div key={page.id} style={{ display: "flex", gap: 12, justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", padding: "10px 0", borderTop: "1px solid var(--line)", opacity: page.status === "GONE" ? 0.6 : 1 }}>
          <div style={{ minWidth: 0, flex: 1 }}>
            <strong style={{ display: "block", overflowWrap: "anywhere" }}>{pageName(page)}</strong>
            <span className="field-help" style={{ overflowWrap: "anywhere" }}>{page.url}</span>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginTop: 6 }}>
              {page.status === "EXCLUDED" && <span className="status status-muted">Dikecualikan</span>}
              {page.status === "GONE" && <span className="status status-muted">Sudah tidak ada di sumber</span>}
              {page.user_edited && <span className="status">Sudah Anda koreksi</span>}
              {page.update_available && <span className="status status-danger">Sumber berubah</span>}
              {page.status === "ACTIVE" && !page.editable && <span className="status status-muted">Hanya dapat dilihat sebagai berkas</span>}
            </div>
          </div>
          {canEdit && <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            {page.status === "ACTIVE" && page.editable && <button className="button" type="button" onClick={() => void openEditor(page)}><Pencil size={14} />Lihat / edit</button>}
            {page.status === "ACTIVE" && <button className="button" type="button" disabled={busy} onClick={() => void setExcluded(page, true)}><EyeOff size={14} />Kecualikan</button>}
            {page.status === "EXCLUDED" && <button className="button" type="button" disabled={busy} onClick={() => void setExcluded(page, false)}><Eye size={14} />Masukkan kembali</button>}
          </div>}
        </div>)}
      </div>}

      {tab === "runs" && <div style={{ marginTop: 16, display: "grid", gap: 10 }}>
        {runs === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
        {runs?.length === 0 && <p className="empty">Belum ada pengambilan.</p>}
        {runs?.map(run => <div key={run.id} style={{ padding: "10px 0", borderTop: "1px solid var(--line)" }}>
          <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
            <strong>{when(run.created_at)}</strong>
            <span className={run.status === "SUCCEEDED" ? "status" : run.status === "FAILED" || run.status === "NEEDS_PUBLIC_SHARING" ? "status status-danger" : "status status-muted"}>{runStatus[run.status] ?? run.status}</span>
            <span className="field-help">{run.trigger === "MANUAL" ? "Manual" : "Terjadwal"}</span>
          </div>
          <span className="field-help">Baru {run.pages_new} · Berubah {run.pages_changed} · Tidak berubah {run.pages_unchanged} · Gagal {run.pages_failed}</span>
          {run.error_public_message && <p className="field-help">{run.error_public_message}</p>}
        </div>)}
      </div>}
    </div>

    {diff && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) setDiff(null); }}>
      <div className="modal panel panel-pad" role="dialog" aria-modal="true" aria-label="Perubahan halaman" style={{ width: "min(100%, 720px)", maxHeight: "86vh", overflow: "auto" }}>
        <div className="card-heading"><div><h2 className="section-title">{pageName(diff)}</h2><p className="section-hint">{eventLabel[diff.event].text} · {when(diff.created_at)}</p></div>
          <button className="button" type="button" onClick={() => setDiff(null)} aria-label="Tutup"><X size={14} /></button></div>
        <DiffView before={diff.before_text ?? ""} after={diff.after_text ?? ""} />
      </div>
    </div>}

    {editing && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget && !busy) setEditing(null); }}>
      <div className="modal panel panel-pad" role="dialog" aria-modal="true" aria-label="Edit halaman" style={{ width: "min(100%, 820px)", maxHeight: "90vh", overflow: "auto" }}>
        <div className="card-heading"><div><h2 className="section-title">{pageName(editing)}</h2><p className="section-hint" style={{ overflowWrap: "anywhere" }}>{editing.url}</p></div>
          <button className="button" type="button" onClick={() => setEditing(null)} aria-label="Tutup"><X size={14} /></button></div>
        {editing.update_available && <div className="notice notice-danger" role="alert" style={{ display: "block" }}>
          <strong>Sumber berubah sejak Anda mengedit halaman ini.</strong>
          <DiffView before={editing.body ?? ""} after={editing.source_text ?? ""} />
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 12 }}>
            <button className="button" type="button" disabled={busy} onClick={() => void resolve("use_source")}>Pakai versi terbaru dari sumber</button>
            <button className="button" type="button" disabled={busy} onClick={() => void resolve("keep_mine")}>Tetap pakai versi saya</button>
          </div>
        </div>}
        <label style={{ display: "block", marginTop: 12 }}><span className="field-label">Teks yang dibaca asisten</span>
          <textarea className="input" rows={16} value={draft} onChange={event => setDraft(event.target.value)} disabled={!canEdit || busy} style={{ fontFamily: "ui-monospace, Consolas, monospace", lineHeight: 1.5 }} />
          <span className="field-help">Perbaiki angka, nama, atau kalimat yang keliru. Halaman yang Anda edit tidak akan ditimpa pengambilan berikutnya; Anda hanya diberi tahu bila sumbernya berubah.</span>
        </label>
        <div className="modal-actions">
          {editing.user_edited && canEdit && <button className="button" type="button" disabled={busy} onClick={() => void resolve("use_source")}>Kembalikan ke versi sumber</button>}
          <button className="button" type="button" disabled={busy} onClick={() => setEditing(null)}>Tutup</button>
          {canEdit && <button className="button button-primary" type="button" disabled={busy || draft === (editing.body ?? "") || !draft.trim()} onClick={() => void save()}>{busy ? "Menyimpan…" : "Simpan perubahan"}</button>}
        </div>
      </div>
    </div>}
  </AppShell>;
}
