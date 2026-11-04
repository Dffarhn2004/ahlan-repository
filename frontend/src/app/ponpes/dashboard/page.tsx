"use client";

import { useEffect, useState } from "react";
import { Activity, CheckCircle2, Circle, CircleAlert, FileText, MessageSquare, RefreshCw, Send } from "lucide-react";
import Link from "next/link";
import { AppShell, PageHeader } from "@/components/app-shell";
import { ShareLinkCard } from "@/components/share-link-card";
import { api } from "@/lib/api";
import { chooseTenant } from "@/lib/auth";
import { auditActionLabel } from "@/lib/labels";
import { ActivityChart, HourlyChart, StatusDonut, trendNote, type ConversationStats } from "@/components/dashboard-charts";

type Tenant = { id: string; name: string; slug?: string; lifecycle_status: string; provisioning_status: string };
type FileItem = { id: string; filename: string; status?: string };
type Bot = { username: string; status: string; access_mode?: "TESTING" | "PUBLIC"; allowed_user_ids?: number[] };
type Profile = { description?: string; address?: string; phone?: string; email?: string };
type LogItem = { id?: string; action: string; actor_name?: string; target_name?: string; occurred_at: string };
type Overview = { tenant: Tenant | null; files: FileItem[]; bot: Bot | null; profile: Profile | null; conversations: number | null; logs: LogItem[]; stats: ConversationStats | null };

const empty: Overview = { tenant: null, files: [], bot: null, profile: null, conversations: null, logs: [], stats: null };
const dateTime = (value: string) => new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));

export default function PonpesDashboard() {
  const [data, setData] = useState<Overview>(empty);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [days, setDays] = useState(14);

  async function load(range = days) {
    setLoading(true); setError("");
    try {
      const response = await api<{ data: Tenant[] }>("/api/v1/tenants");
      const tenants = response?.data ?? [];
      const id = chooseTenant(tenants);
      const tenant = tenants.find(item => item.id === id) ?? null;
      if (!id) { setData(empty); return; }
      // Secondary widgets may be forbidden for some roles; they degrade to empty instead of failing the page.
      const [files, bot, profile, conversations, logs, stats] = await Promise.all([
        api<{ data: FileItem[] }>(`/api/v1/tenants/${id}/files`).catch(() => null),
        api<Bot>(`/api/v1/tenants/${id}/bot`).catch(() => null),
        api<Profile>(`/api/v1/tenants/${id}/profile`).catch(() => null),
        api<{ data: unknown[] }>(`/api/v1/conversations?tenant_id=${id}`).catch(() => null),
        api<{ data: LogItem[] }>(`/api/v1/audit-logs?tenant_id=${id}`).catch(() => null),
        api<ConversationStats>(`/api/v1/conversations/stats?tenant_id=${id}&days=${range}`).catch(() => null),
      ]);
      setData({ tenant, files: files?.data ?? [], bot: bot ?? null, profile: profile ?? null, conversations: conversations ? (conversations.data ?? []).length : null, logs: (logs?.data ?? []).slice(0, 6), stats: stats ?? null });
    } catch {
      setError("Informasi beranda belum dapat dimuat. Silakan coba lagi.");
    } finally { setLoading(false); }
  }
  useEffect(() => { queueMicrotask(() => void load()); }, []);

  const { tenant, files, bot, profile, conversations, logs, stats } = data;
  function changeRange(next: number) { setDays(next); void load(next); }
  const provisioned = tenant?.provisioning_status === "SUCCEEDED";
  const running = bot?.status === "ACTIVE";
  const pendingFiles = files.filter(file => file.status === "PENDING_SYNC").length;
  const profileDone = Boolean(profile && (profile.description || profile.address || profile.phone || profile.email));
  const steps = [
    { done: profileDone, label: "Lengkapi informasi resmi pesantren", href: "/ponpes/profile" },
    { done: files.length > 0, label: files.length ? `${files.length} dokumen tersimpan${pendingFiles ? `, ${pendingFiles} belum dipakai asisten` : ""}` : "Unggah minimal satu dokumen", href: "/ponpes/files" },
    { done: provisioned, label: "Siapkan asisten", href: "/ponpes/profile" },
    { done: Boolean(bot), label: "Hubungkan bot Telegram", href: "/ponpes/profile" },
    { done: running, label: "Aktifkan asisten", href: "/ponpes/profile" },
  ];
  const next = steps.find(step => !step.done);
  const statusText = !provisioned ? "Belum disiapkan" : !bot ? "Menunggu Telegram" : running ? "Aktif" : "Siap diaktifkan";

  return <AppShell mode="ponpes" title="Dashboard">
    <PageHeader eyebrow="Ruang kerja pesantren" title={tenant?.name ?? "Selamat datang"} description="Kelola informasi yang membantu calon santri dengan lebih cepat dan konsisten." action={<button className="button" onClick={() => void load()} disabled={loading}><RefreshCw size={15} className={loading ? "spin" : ""} />Muat ulang</button>} />
    {error && <div className="notice notice-danger" role="alert"><CircleAlert size={17} /><span>{error}</span><button className="button" onClick={() => void load()}>Coba lagi</button></div>}
    <ShareLinkCard slug={tenant?.slug} />
    <section className="stats">
      <div className="stat"><span className="stat-label">Status asisten</span><strong className="stat-value">{loading ? "—" : statusText}</strong><span className={`stat-note ${running ? "positive" : "warning"}`}>{running ? <><CheckCircle2 size={12} />Siap melayani calon santri</> : "Perlu diatur"}</span></div>
      <div className="stat"><span className="stat-label">Dokumen</span><strong className="stat-value">{loading ? "—" : files.length}</strong><span className="stat-note"><FileText size={12} />{pendingFiles ? `${pendingFiles} belum dipakai asisten` : "Semua sudah dipakai asisten"}</span></div>
      <div className="stat"><span className="stat-label">Percakapan Telegram</span><strong className="stat-value">{loading ? "—" : conversations ?? "—"}</strong><span className="stat-note"><MessageSquare size={12} />{conversations === null ? "Tidak tersedia untuk peran Anda" : "Total percakapan masuk"}</span></div>
      <div className="stat"><span className="stat-label">Akses Telegram</span><strong className="stat-value">{loading ? "—" : bot ? (bot.access_mode === "PUBLIC" ? "Publik" : "Uji coba") : "Belum ada"}</strong><span className="stat-note"><Send size={12} />{bot ? (bot.access_mode === "PUBLIC" ? `@${bot.username} terbuka untuk umum` : `@${bot.username} · ${bot.allowed_user_ids?.length ?? 0} penguji`) : "Hubungkan bot untuk memulai"}</span></div>
    </section>
    {stats && <>
      <section className="stats" style={{ marginTop: 20 }}>
        <div className="stat"><span className="stat-label">Pertanyaan masuk ({days} hari)</span><strong className="stat-value">{stats.user_messages}</strong><span className="stat-note">{trendNote(stats.user_messages, stats.prev_user_messages)}</span></div>
        <div className="stat"><span className="stat-label">Balasan asisten ({days} hari)</span><strong className="stat-value">{stats.bot_replies}</strong><span className="stat-note">{trendNote(stats.bot_replies, stats.prev_bot_replies)}</span></div>
        <div className="stat"><span className="stat-label">Pengunjung unik ({days} hari)</span><strong className="stat-value">{stats.visitors}</strong><span className="stat-note">{trendNote(stats.visitors, stats.prev_visitors)}</span></div>
        <div className="stat"><span className="stat-label">Rata-rata pertanyaan per pengunjung</span><strong className="stat-value">{stats.visitors ? (stats.user_messages / stats.visitors).toFixed(1) : "—"}</strong><span className="stat-note">Semakin tinggi, semakin aktif diskusinya</span></div>
      </section>
      <section className="chart-grid">
        <div className="panel panel-pad">
          <div className="chart-head">
            <div><h2 className="section-title">Aktivitas percakapan</h2><p className="section-hint">Pertanyaan masuk dan balasan asisten per hari</p></div>
            <div className="chart-toggle">{[7, 14, 30].map(range => <button key={range} className={days === range ? "active" : ""} onClick={() => changeRange(range)}>{range} hari</button>)}</div>
          </div>
          <ActivityChart daily={stats.daily} />
        </div>
        <div className="panel panel-pad">
          <h2 className="section-title">Status percakapan</h2>
          <p className="section-hint">Seluruh percakapan Telegram</p>
          <StatusDonut breakdown={stats.status_breakdown} />
        </div>
      </section>
      <section className="chart-grid" style={{ gridTemplateColumns: "minmax(0,1fr)" }}>
        <div className="panel panel-pad">
          <h2 className="section-title">Jam tersibuk</h2>
          <p className="section-hint">Kapan calon santri paling sering bertanya ({days} hari terakhir, WIB)</p>
          <HourlyChart hourly={stats.hourly} />
        </div>
      </section>
    </>}
    <section className="grid-two">
      <div className="panel panel-pad">
        <h2 className="section-title">Siapkan asisten AI Anda</h2>
        <p className="section-hint">{next ? "Selesaikan langkah berikut sebelum layanan dibagikan kepada calon santri." : "Semua langkah selesai. Asisten siap melayani calon santri."}</p>
        <div className="setup-list">{steps.map(step => <div key={step.label}><span className={step.done ? "setup-done" : "setup-pending"}>{step.done ? <CheckCircle2 size={16} /> : <Circle size={16} />}</span><span>{step.label}</span></div>)}</div>
        <div style={{ display: "flex", gap: 10, marginTop: 18, flexWrap: "wrap" }}>{next ? <Link className="button button-primary" href={next.href}>Lanjutkan: {next.label}</Link> : <Link className="button button-primary" href="/ponpes/conversations">Lihat percakapan</Link>}<Link className="button" href="/ponpes/files">Kelola dokumen</Link></div>
      </div>
      <div className="panel panel-pad">
        <h2 className="section-title">Aktivitas terbaru</h2>
        {logs.length ? <div className="log-list">{logs.map((log, index) => <div className="log-item" key={log.id ?? index}><div className="log-dot"><Activity size={13} /></div><div><strong>{auditActionLabel(log.action)}</strong><span>{log.target_name || log.actor_name || "Sistem"}</span></div><time>{dateTime(log.occurred_at)}</time></div>)}</div> : <div className="empty">{loading ? "Memuat aktivitas…" : "Aktivitas akan tampil saat layanan mulai digunakan."}</div>}
        {logs.length > 0 && <Link className="text-link" href="/audit" style={{ display: "inline-block", marginTop: 12 }}>Lihat semua aktivitas</Link>}
      </div>
    </section>
  </AppShell>;
}
