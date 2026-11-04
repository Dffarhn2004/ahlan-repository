"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { Eye, Globe, LoaderCircle, Pause, Play, RefreshCw, Trash2 } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { api, ApiError } from "@/lib/api";
import { can, chooseTenant, loadAuth, type AuthContext } from "@/lib/auth";
import { parseSourceUrl, sourceKindLabel, type SourceKind } from "@/lib/links";

type Tenant = { id: string; name: string };
type ScheduleMode = "MANUAL" | "INTERVAL";
type Source = {
  id: string; url: string; kind: SourceKind; schedule_mode: ScheduleMode; interval_minutes?: number; enabled: boolean;
  last_status?: "QUEUED" | "RUNNING" | "SUCCEEDED" | "PARTIAL" | "FAILED" | "NEEDS_PUBLIC_SHARING";
  last_run_at?: string; next_run_at?: string; pages_count?: number; error_public_message?: string;
};

const intervals = [{ label: "Setiap 6 jam", value: 360 }, { label: "Setiap 12 jam", value: 720 }, { label: "Setiap hari", value: 1440 }, { label: "Setiap minggu", value: 10080 }];
const errorText: Record<string, string> = {
  CRAWL_URL_FORBIDDEN: "Tautan ditolak karena tidak aman atau tidak dapat dijangkau.",
  CRAWL_SOURCE_LIMIT: "Batas jumlah sumber tercapai. Hapus sumber lama terlebih dahulu.",
  CRAWL_SCHEDULE_TOO_FREQUENT: "Jadwal terlalu sering. Minimal setiap 6 jam.",
  CRAWL_ALREADY_RUNNING: "Pengambilan sedang berjalan untuk sumber ini.",
  CRAWL_RUN_TOO_SOON: "Tunggu 10 menit sebelum mengambil ulang sumber yang sama.",
  CRAWL_SOURCE_EXISTS: "Tautan ini sudah terdaftar.",
  CRAWL_URL_UNREACHABLE: "Alamat situs tidak ditemukan. Periksa penulisan tautan.",
  CRAWL_SOURCE_DISABLED: "Sumber sedang dijeda. Aktifkan terlebih dahulu.",
  CRAWL_FORCED_DISABLED: "Sumber ini dinonaktifkan oleh pengelola platform.",
  MAPS_URL_INVALID: "Tautan tidak valid.",
};

function formatTime(value?: string) {
  return value ? new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short", timeZone: "Asia/Jakarta" }).format(new Date(value)) : "—";
}

function StatusBadge({ source }: { source: Source }) {
  if (!source.enabled) return <span className="status status-muted">Dijeda</span>;
  switch (source.last_status) {
    case "RUNNING": case "QUEUED": return <span className="status status-muted"><LoaderCircle size={11} className="spin" />Sedang diambil</span>;
    case "SUCCEEDED": return <span className="status">Berhasil</span>;
    case "PARTIAL": return <span className="status status-muted">Sebagian</span>;
    case "NEEDS_PUBLIC_SHARING": return <span className="status status-danger">Dokumen belum publik</span>;
    case "FAILED": return <span className="status status-danger">Gagal</span>;
    default: return <span className="status status-muted">Belum pernah diambil</span>;
  }
}

export default function SourcesPage() {
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [tenantId, setTenantId] = useState("");
  const [sources, setSources] = useState<Source[] | null>(null);
  const [notice, setNotice] = useState("");
  const [url, setUrl] = useState("");
  const [mode, setMode] = useState<ScheduleMode>("INTERVAL");
  const [interval, setIntervalMinutes] = useState(1440);
  const [saving, setSaving] = useState(false);
  const [busyId, setBusyId] = useState("");

  const canEdit = can(auth, "tenant:update");
  const base = `/api/v1/tenants/${tenantId}/crawl/sources`;
  const parsed = url.trim() ? parseSourceUrl(url) : null;
  const urlError = parsed && !parsed.ok ? parsed.error : "";

  const load = useCallback(async (id: string) => {
    try { const result = await api<{ data: Source[] }>(`/api/v1/tenants/${id}/crawl/sources`); setSources(result?.data ?? []); }
    catch { setSources([]); setNotice("Daftar sumber belum dapat dimuat."); }
  }, []);

  useEffect(() => {
    void (async () => {
      try {
        const [authData, tenantData] = await Promise.all([loadAuth(), api<{ data: Tenant[] }>("/api/v1/tenants")]);
        setAuth(authData);
        const id = chooseTenant(tenantData?.data ?? []);
        if (!id) { setSources([]); return; }
        setTenantId(id);
        await load(id);
      } catch { setSources([]); setNotice("Halaman belum dapat dimuat."); }
    })();
  }, [load]);

  // Poll while any source is being fetched so the status updates without a manual refresh.
  const active = sources?.some(source => source.last_status === "RUNNING" || source.last_status === "QUEUED");
  useEffect(() => {
    if (!active || !tenantId) return;
    const timer = window.setTimeout(() => void load(tenantId), 4000);
    return () => window.clearTimeout(timer);
  }, [active, sources, tenantId, load]);

  function fail(error: unknown, fallback: string) {
    setNotice(error instanceof ApiError ? (errorText[error.message] ?? fallback) : fallback);
  }

  async function add(event: React.FormEvent) {
    event.preventDefault();
    const check = parseSourceUrl(url);
    if (!check.ok) { setNotice(check.error); return; }
    setSaving(true); setNotice("");
    const body = { url: url.trim(), kind: check.kind, schedule: mode === "MANUAL" ? { mode } : { mode, interval_minutes: interval } };
    try {
      await api(base, { method: "POST", body: JSON.stringify(body) });
      setUrl(""); setNotice(mode === "INTERVAL" ? "Sumber ditambahkan. Pengambilan pertama dimulai sebentar lagi." : "Sumber ditambahkan. Pilih \"Ambil sekarang\" untuk mulai mengambil informasi.");
      await load(tenantId);
    } catch (error) { fail(error, "Sumber gagal ditambahkan."); }
    finally { setSaving(false); }
  }

  async function act(source: Source, run: () => Promise<unknown>, failure: string) {
    setBusyId(source.id); setNotice("");
    try { await run(); await load(tenantId); } catch (error) { fail(error, failure); }
    finally { setBusyId(""); }
  }

  const runNow = (source: Source) => act(source, () => api(`${base}/${source.id}/run`, { method: "POST" }), "Pengambilan gagal dimulai.");
  const toggle = (source: Source) => act(source, () => api(`${base}/${source.id}`, { method: "PATCH", body: JSON.stringify({ enabled: !source.enabled }) }), "Status sumber gagal diubah.");
  const remove = (source: Source) => { if (window.confirm("Hapus sumber ini? Informasi yang sudah diambil dari sumber ini tidak akan dipakai lagi oleh asisten.")) void act(source, () => api(`${base}/${source.id}`, { method: "DELETE" }), "Sumber gagal dihapus."); };

  return <AppShell mode="ponpes" title="Sumber Web">
    <PageHeader eyebrow="Pengetahuan asisten" title="Sumber web" description="Asisten dapat mempelajari situs resmi dan dokumen Google (Docs, Sheets, Slides, file Drive) milik pesantren, diperbarui otomatis sesuai jadwal atau saat Anda minta." />
    <div style={{ display: "grid", gap: 20 }}>
      {canEdit && <form className="panel panel-pad" onSubmit={add}>
        <div className="card-heading"><div><h2 className="section-title">Tambah sumber</h2><p className="section-hint">Tempel tautan situs resmi atau dokumen Google. Dokumen Google harus dibagikan sebagai &quot;Siapa saja yang memiliki link&quot; (hanya melihat); kami tidak meminta akses ke akun Anda.</p></div></div>
        <label style={{ display: "block" }}><span className="field-label">Tautan</span>
          <input className="input" required value={url} onChange={event => setUrl(event.target.value)} placeholder="https://pesantren.id atau https://docs.google.com/document/d/…" aria-invalid={Boolean(urlError)} />
          {urlError ? <span className="error">{urlError}</span> : parsed?.ok && <span className="field-help">Terdeteksi: {sourceKindLabel[parsed.kind]}</span>}
        </label>
        <div className="form-row" style={{ marginTop: 16 }}>
          <label><span className="field-label">Pembaruan</span>
            <select className="input" value={mode} onChange={event => setMode(event.target.value as ScheduleMode)}><option value="INTERVAL">Otomatis sesuai jadwal</option><option value="MANUAL">Hanya saat saya minta</option></select>
          </label>
          {mode === "INTERVAL" && <label><span className="field-label">Jadwal</span>
            <select className="input" value={interval} onChange={event => setIntervalMinutes(Number(event.target.value))}>{intervals.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
            <span className="field-help">Waktu ditampilkan dalam WIB.</span>
          </label>}
        </div>
        <div style={{ marginTop: 18 }}><button className="button button-primary" disabled={!tenantId || saving || Boolean(urlError)}>{saving ? "Menambahkan…" : "Tambah sumber"}</button></div>
      </form>}
      {notice && <p className={notice.includes("ditambahkan") ? "status" : "error"} role="status">{notice}</p>}
      <section className="panel panel-pad">
        <div className="card-heading"><div><h2 className="section-title">Sumber terdaftar</h2><p className="section-hint">Hasil pengambilan dipakai asisten dengan menyebut sumber dan tanggalnya. Dokumen yang Anda unggah tetap diutamakan jika isinya berbeda.</p></div></div>
        {sources === null && <p className="empty"><LoaderCircle size={14} className="spin" /> Memuat…</p>}
        {sources?.length === 0 && <p className="empty"><Globe size={16} /> Belum ada sumber web.</p>}
        <div style={{ display: "grid", gap: 12 }}>
          {sources?.map(source => <article key={source.id} className="panel panel-pad" style={{ display: "grid", gap: 8 }}>
            <div style={{ display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap", alignItems: "center" }}>
              <div style={{ minWidth: 0 }}><strong>{sourceKindLabel[source.kind] ?? source.kind}</strong><div className="field-help" style={{ overflowWrap: "anywhere" }}>{source.url}</div></div>
              <StatusBadge source={source} />
            </div>
            <div className="field-help">
              {source.schedule_mode === "MANUAL" ? "Manual" : `Otomatis tiap ${intervals.find(item => item.value === source.interval_minutes)?.label.replace("Setiap ", "") ?? `${source.interval_minutes} menit`}`}
              {" · "}Terakhir: {formatTime(source.last_run_at)}
              {source.schedule_mode !== "MANUAL" && source.enabled && <> · Berikutnya: {formatTime(source.next_run_at)}</>}
              {typeof source.pages_count === "number" && <> · {source.pages_count} halaman</>}
            </div>
            {source.last_status === "NEEDS_PUBLIC_SHARING" && <p className="error">Dokumen belum dapat dibaca. Ubah akses menjadi &quot;Siapa saja yang memiliki link&quot; di Google, lalu pilih Ambil sekarang.</p>}
            {source.last_status === "FAILED" && source.error_public_message && <p className="error">{source.error_public_message}</p>}
            {canEdit && <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              <Link className="button" href={`/ponpes/sources/${source.id}`}><Eye size={14} />Lihat hasil</Link>
              <button className="button" type="button" disabled={busyId === source.id || !source.enabled || source.last_status === "RUNNING" || source.last_status === "QUEUED"} onClick={() => void runNow(source)}><RefreshCw size={14} />Ambil sekarang</button>
              <button className="button" type="button" disabled={busyId === source.id} onClick={() => void toggle(source)}>{source.enabled ? <><Pause size={14} />Jeda</> : <><Play size={14} />Aktifkan</>}</button>
              <button className="button" type="button" disabled={busyId === source.id} onClick={() => remove(source)}><Trash2 size={14} />Hapus</button>
            </div>}
          </article>)}
        </div>
      </section>
    </div>
  </AppShell>;
}
