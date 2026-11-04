"use client";

import { useEffect, useState } from "react";
import { Bot, Check, LoaderCircle, RefreshCw, Save, Settings2 } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Pagination, usePagination } from "@/components/pagination";

type SyncResult = { tenant_id: string; tenant_name: string; status: "OK" | "FAILED"; failed?: string[]; restarted: boolean };
type SyncJob = { job_id: string; status: "RUNNING" | "DONE" | "FAILED"; phase: string; total: number; done: number; data: SyncResult[]; error?: string };
type Health = { tenant_id?: string; tenant_name?: string; runtime_status?: string; file_count?: number };
type HermesSettings = { description: string; soul_md: string; enabled_skills: string[]; config_revision?: number };
type HermesSkill = { name: string; description?: string; category?: string };

const emptySettings: HermesSettings = { description: "", soul_md: "", enabled_skills: [] };

export default function HermesPage() {
  const [items, setItems] = useState<Health[]>([]);
  const [settings, setSettings] = useState<HermesSettings>(emptySettings);
  const [skills, setSkills] = useState<HermesSkill[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState("");
  const [error, setError] = useState(false);
  const [tab, setTab] = useState<"profiles" | "settings">("profiles");
  const [syncing, setSyncing] = useState(false);
  const [restartActive, setRestartActive] = useState(true);
  const [syncResults, setSyncResults] = useState<SyncResult[] | null>(null);
  const [syncError, setSyncError] = useState("");
  const [syncProgress, setSyncProgress] = useState("");
  const router = useRouter();
  const profilesPaged = usePagination(items, 10);
  // Syncing takes minutes, so the server runs it as a background job: start it, then poll until it finishes.
  // A few failed polls in a row are tolerated (mobile networks drop), because the job keeps running on the server.
  async function syncAll() {
    if (syncing || !window.confirm(restartActive ? "Sinkronkan semua asisten dan mulai ulang yang sedang aktif? Asisten akan terputus sesaat." : "Sinkronkan pengaturan semua asisten?")) return;
    setSyncing(true); setSyncError(""); setSyncResults(null); setSyncProgress("Memulai sinkronisasi...");
    try {
      const started = await api<{ job_id: string }>("/api/v1/hermes/sync-all", { method: "POST", body: JSON.stringify({ restart: restartActive }) });
      if (!started?.job_id) throw new Error("no job");
      const deadline = Date.now() + 25 * 60 * 1000;
      let failures = 0;
      while (Date.now() < deadline) {
        await new Promise(resolve => setTimeout(resolve, 2000));
        let job: SyncJob | null = null;
        try { job = await api<SyncJob>(`/api/v1/hermes/sync-all/${started.job_id}`); failures = 0; }
        catch { if (++failures >= 5) throw new Error("polling"); continue; }
        if (!job) continue;
        setSyncResults(job.data ?? []);
        setSyncProgress(job.total ? `${job.phase} (${job.done} dari ${job.total} asisten)` : job.phase);
        if (job.status === "FAILED") { setSyncError("Sinkronisasi berhenti karena kesalahan di server. Coba lagi beberapa saat lagi."); return; }
        if (job.status === "DONE") return;
      }
      setSyncError("Sinkronisasi belum selesai setelah 25 menit. Periksa status asisten, lalu coba lagi.");
    } catch { setSyncError("Sinkronisasi gagal dijalankan atau statusnya tidak dapat dipantau. Coba lagi beberapa saat lagi."); }
    finally { setSyncing(false); setSyncProgress(""); }
  }

  useEffect(() => {
    void Promise.allSettled([
      api<{ data: Health[] }>("/api/v1/hermes/overview"),
      api<HermesSettings>("/api/v1/hermes/settings"),
      api<{ data: HermesSkill[] }>("/api/v1/hermes/skills"),
    ]).then(([overviewResult, settingsResult, skillsResult]) => {
      const overview = overviewResult.status === "fulfilled" ? overviewResult.value : null;
      const globalSettings = settingsResult.status === "fulfilled" ? settingsResult.value : null;
      const availableSkills = skillsResult.status === "fulfilled" ? skillsResult.value : null;
      setItems(overview?.data ?? []);
      setSettings({ description: globalSettings?.description ?? "", soul_md: globalSettings?.soul_md ?? "", enabled_skills: globalSettings?.enabled_skills ?? [], config_revision: globalSettings?.config_revision });
      setSkills(availableSkills?.data ?? []);
      setError(!overview && !globalSettings);
    }).finally(() => setLoading(false));
  }, []);

  const save = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setFeedback("");
    try {
      const result = await api<{ status: string; profiles_applied: number; data: HermesSettings }>("/api/v1/hermes/settings", { method: "PATCH", body: JSON.stringify(settings) });
      if (result?.data) setSettings(result.data);
      setFeedback(`Pengaturan global berhasil diterapkan ke ${result?.profiles_applied ?? 0} profile Hermes.`);
    } catch { setFeedback("Pengaturan global gagal diterapkan ke asisten."); }
    finally { setSaving(false); }
  };

  const toggleSkill = (name: string, enabled: boolean) => setSettings(current => ({ ...current, enabled_skills: enabled ? [...current.enabled_skills, name] : current.enabled_skills.filter(item => item !== name) }));

  return <AppShell title="Profil Asisten">
    <PageHeader eyebrow="Pusat kendali platform" title="Profil Asisten" description="Pantau profil asisten dan kelola pengaturan global platform." />
    <nav className="detail-tabs hermes-admin-tabs" aria-label="Navigasi asisten">
      <button type="button" className={tab === "profiles" ? "active" : ""} onClick={() => setTab("profiles")}><Bot size={16} /> Daftar Asisten</button>
      <button type="button" className={tab === "settings" ? "active" : ""} onClick={() => setTab("settings")}><Settings2 size={16} /> Pengaturan global</button>
    </nav>
    {tab === "settings" && <section className="panel panel-pad hermes-global-settings">
      <div className="card-heading"><div><h2 className="section-title"><Settings2 size={18} /> Pengaturan global asisten</h2><p className="section-hint">Perubahan di sini menjadi konfigurasi bersama dan diterapkan ke semua asisten.</p></div><Bot size={20} color="var(--brand)" /></div>
      {loading ? <div className="settings-loading"><span /><span /><span /></div> : error ? <div className="notice">Pengaturan asisten belum dapat dimuat. Muat ulang halaman untuk mencoba lagi.</div> : <form className="global-settings-form" onSubmit={save}>
        <label><span className="field-label">Deskripsi asisten global</span><input className="input" value={settings.description} onChange={event => setSettings({ ...settings, description: event.target.value })} placeholder="Asisten Humanis Layanan Admisi Santri" /></label>
        <label><span className="field-label">SOUL.md global</span><textarea className="input soul-editor" rows={10} value={settings.soul_md} onChange={event => setSettings({ ...settings, soul_md: event.target.value })} placeholder="# Identitas asisten\n\nTuliskan karakter, aturan, dan batasan asisten." /></label>
        <div><span className="field-label">Kemampuan yang diizinkan secara global</span><p className="field-help">Kemampuan yang dicentang akan tersedia pada seluruh asisten.</p><div className="skill-picker">{skills.length ? skills.map(skill => <label className="skill-option" key={skill.name}><input type="checkbox" checked={settings.enabled_skills.includes(skill.name)} onChange={event => toggleSkill(skill.name, event.target.checked)} /><span><strong>{skill.name}</strong><small>{skill.description ?? skill.category ?? "Kemampuan asisten"}</small></span></label>) : <span className="field-help">Daftar kemampuan asisten belum dapat dimuat.</span>}</div></div>
        <div className="global-settings-actions"><span className="field-help">Revision {settings.config_revision ?? 1}</span><button className="button button-primary" type="submit" disabled={saving}>{saving ? <LoaderCircle size={15} className="spin" /> : <Save size={15} />}{saving ? "Menerapkan..." : "Simpan & terapkan ke semua profile"}</button></div>
        {feedback && <div className="action-feedback" role="status"><Check size={15} />{feedback}</div>}
      </form>}
    </section>}
    {tab === "profiles" && <section className="panel panel-pad sync-all-panel" style={{ marginBottom: 20 }}>
      <div className="card-heading"><div><h2 className="section-title">Sinkronkan semua asisten</h2><p className="section-hint">Terapkan ulang persona, aturan layanan, koneksi Telegram, dan pencatatan percakapan ke setiap asisten. Pakai ini setelah ada perubahan sistem atau bila ada asisten yang berperilaku tidak sesuai.</p></div><RefreshCw size={19} color="var(--brand)" /></div>
      <div className="sync-all-actions"><label className="skill-option" style={{ border: 0, padding: 0 }}><input type="checkbox" checked={restartActive} onChange={event => setRestartActive(event.target.checked)} /><span><strong>Mulai ulang asisten yang sedang aktif</strong><small>Diperlukan agar perubahan langsung berlaku. Asisten terputus beberapa detik.</small></span></label><button className="button button-primary" type="button" onClick={() => void syncAll()} disabled={syncing || loading}>{syncing ? <LoaderCircle size={15} className="spin" /> : <RefreshCw size={15} />}{syncing ? "Menyinkronkan..." : "Sinkronkan semua"}</button></div>
      {syncing && syncProgress && <div className="action-feedback" role="status" style={{ marginTop: 12 }}><LoaderCircle size={15} className="spin" />{syncProgress}</div>}
      {syncError && <div className="error" role="alert" style={{ marginTop: 12 }}>{syncError}</div>}
      {syncResults && (syncResults.length ? <div className="sync-all-results" role="status">{syncResults.map(result => <div key={result.tenant_id} className="sync-all-row"><strong>{result.tenant_name}</strong><span className={result.status === "OK" ? "status" : "status status-danger"}>{result.status === "OK" ? (result.restarted ? "Berhasil · dimulai ulang" : "Berhasil") : "Sebagian gagal"}</span>{result.failed?.length ? <small>Gagal: {result.failed.join(", ")}</small> : null}</div>)}</div> : syncing ? null : <div className="empty">Belum ada asisten yang perlu disinkronkan.</div>)}
    </section>}
    {tab === "profiles" && <section className="panel"><div className="table-wrap"><table className="data-table"><thead><tr><th>Pesantren</th><th>Status asisten</th><th>Dokumen</th></tr></thead><tbody>{loading ? <tr><td colSpan={3} className="empty">Memuat status asisten...</td></tr> : error ? <tr><td colSpan={3} className="empty">Status asisten belum tersedia.</td></tr> : items.length ? profilesPaged.rows.map((item, index) => <tr key={`${item.tenant_name ?? "tenant"}-${index}`} className={item.tenant_id ? "row-link" : undefined} onClick={item.tenant_id ? () => router.push(`/admin/tenants/${item.tenant_id}?tab=hermes`) : undefined}><td>{item.tenant_id ? <Link className="table-link" href={`/admin/tenants/${item.tenant_id}?tab=hermes`}><strong>{item.tenant_name ?? "Pesantren"}</strong></Link> : <strong>{item.tenant_name ?? "Pesantren"}</strong>}</td><td><span className="status">{item.runtime_status ?? "UNKNOWN"}</span></td><td>{item.file_count ?? 0} file</td></tr>) : <tr><td colSpan={3} className="empty">Belum ada asisten yang terdaftar.</td></tr>}</tbody></table></div><Pagination paged={profilesPaged} label="pesantren"/></section>}
  </AppShell>;
}
