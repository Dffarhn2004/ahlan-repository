"use client";

/* eslint-disable @next/next/no-img-element -- the QR code is a generated data URL */

import { useCallback, useEffect, useState } from "react";
import { Activity, Bot, CheckCircle2, Clock3, LoaderCircle, MessageCircle, RefreshCw, Sparkles, Trash2, XCircle } from "lucide-react";
import { api } from "@/lib/api";
import { provisionLabel } from "@/lib/labels";
import { KnowledgeBanner, useKnowledgeStatus } from "@/components/knowledge-status";

type Tenant = { id: string; provisioning_status: string; slug: string };
type TelegramBot = { telegram_bot_id: number; username: string; status: string; access_mode?: "TESTING" | "PUBLIC"; allowed_user_ids?: number[] };
type WhatsAppChannel = { phone_number: string; status: string; access_mode?: "TESTING" | "PUBLIC"; allowed_numbers?: string[] };
type Pairing = { status: "none" | "starting" | "qr" | "connected" | "error"; qr?: string; phone_number?: string; error?: string };
type RunAction = "start" | "restart" | "stop" | "update" | "delete";

const actionLabels: Record<RunAction, string> = { start: "Aktifkan asisten", restart: "Mulai ulang asisten", stop: "Nonaktifkan asisten", update: "Perbarui asisten", delete: "Hapus asisten" };
const statusClass = (value: string) => { const text = value.toLowerCase(); return text.includes("fail") || text.includes("suspend") || text.includes("error") ? "status status-danger" : text.includes("pending") || text.includes("queue") || text.includes("progress") ? "status status-muted" : "status"; };
const botStatusLabel = (value: string) => ({ PENDING: "Menunggu", VERIFIED: "Terverifikasi", ACTIVE: "Aktif", SUSPENDED: "Dinonaktifkan", INVALID: "Tidak valid" } as Record<string, string>)[value] ?? value;

/** Blocking popup for slow work, so people understand a long wait is normal and do not close or retry. */
function ProcessingOverlay({ title, steps }: { title: string; steps: string[] }) {
  const [index, setIndex] = useState(0);
  useEffect(() => {
    const timer = window.setInterval(() => setIndex(current => Math.min(current + 1, steps.length - 1)), 5000);
    return () => window.clearInterval(timer);
  }, [steps.length]);
  return <div className="processing-backdrop" role="alertdialog" aria-modal="true" aria-live="polite" aria-label={title}>
    <div className="processing-card">
      <div className="processing-orb" aria-hidden="true"><span className="processing-ring" /><span className="processing-ring processing-ring-2" /><Sparkles size={26} /></div>
      <h3>{title}</h3>
      <p className="processing-step" key={index}>{steps[index]}…</p>
      <div className="processing-bar" aria-hidden="true"><span /></div>
      <ol className="processing-dots" aria-hidden="true">{steps.map((step, i) => <li key={step} className={i < index ? "done" : i === index ? "current" : ""} />)}</ol>
      <p className="processing-note">Proses ini membutuhkan waktu sekitar 15–60 detik karena asisten disiapkan khusus untuk pesantren Anda. Mohon jangan menutup atau memuat ulang halaman.</p>
    </div>
  </div>;
}

/** Self-service control for a pesantren's assistant: prepare, connect Telegram, then activate. */
export function AssistantPanel({ tenantId, canManage, canDelete = false, onChange, showKnowledge = true }: { tenantId: string; canManage: boolean; canDelete?: boolean; onChange?: () => void; showKnowledge?: boolean }) {
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [bot, setBot] = useState<TelegramBot | null>(null);
  const [wa, setWa] = useState<WhatsAppChannel | null>(null);
  const [waModal, setWaModal] = useState(false);
  const [pairing, setPairing] = useState<Pairing>({ status: "none" });
  const [waMode, setWaMode] = useState<"TESTING" | "PUBLIC">("TESTING");
  const [waAllowed, setWaAllowed] = useState("");
  const [waError, setWaError] = useState("");
  const [busyAction, setBusyAction] = useState<string | null>(null);
  const [message, setMessage] = useState("");
  const [modal, setModal] = useState(false);
  const [token, setToken] = useState("");
  const [username, setUsername] = useState("");
  const [mode, setMode] = useState<"TESTING" | "PUBLIC">("TESTING");
  const [allowed, setAllowed] = useState("");
  const [formError, setFormError] = useState("");
  const { status: knowledge } = useKnowledgeStatus(tenantId);

  const refresh = useCallback(async () => {
    const [latest, latestBot, latestWa] = await Promise.all([api<Tenant>(`/api/v1/tenants/${tenantId}`).catch(() => null), api<TelegramBot>(`/api/v1/tenants/${tenantId}/bot`).catch(() => null), api<WhatsAppChannel>(`/api/v1/tenants/${tenantId}/whatsapp`).catch(() => null)]);
    if (latest) setTenant(latest);
    setBot(latestBot ?? null);
    setWa(latestWa ?? null);
    onChange?.();
    return { bot: latestBot ?? null, wa: latestWa ?? null };
  }, [tenantId, onChange]);

  // Starting or stopping is applied in the background, so keep checking until the channel status really changes.
  async function refreshUntil(wantRunning: boolean) {
    for (let attempt = 0; attempt < 30; attempt += 1) {
      const latest = await refresh();
      if ((latest.bot?.status === "ACTIVE" || latest.wa?.status === "ACTIVE") === wantRunning) return true;
      await new Promise(resolve => window.setTimeout(resolve, 2000));
    }
    return false;
  }
  useEffect(() => { if (tenantId) queueMicrotask(() => void refresh()); }, [tenantId, refresh]);

  const provisioned = tenant?.provisioning_status === "SUCCEEDED";
  const running = bot?.status === "ACTIVE" || wa?.status === "ACTIVE";
  const hasChannel = Boolean(bot || wa);
  const busy = Boolean(busyAction);
  const overall = !provisioned ? { label: "Belum disiapkan", tone: "status status-muted" } : !hasChannel ? { label: "Menunggu saluran", tone: "status status-muted" } : running ? { label: "Aktif", tone: "status" } : { label: "Siap diaktifkan", tone: "status status-muted" };

  async function provision() {
    if (busy) return;
    setBusyAction("provision");
    setMessage("Menyiapkan asisten...");
    try {
      await api(`/api/v1/tenants/${tenantId}/hermes/provision`, { method: "POST" });
      for (let attempt = 0; attempt < 15; attempt += 1) {
        await new Promise(resolve => window.setTimeout(resolve, 1500));
        const latest = await api<Tenant>(`/api/v1/tenants/${tenantId}`);
        if (latest) setTenant(latest);
        const status = latest?.provisioning_status?.toLowerCase() ?? "";
        if (["succeeded", "failed", "cancelled"].includes(status)) {
          setMessage(status === "succeeded" ? "Asisten berhasil disiapkan. Lanjutkan dengan menghubungkan Telegram atau WhatsApp." : `Penyiapan asisten berhenti dengan status ${provisionLabel(status)}.`);
          return;
        }
      }
      setMessage("Penyiapan asisten masih berjalan. Status akan diperbarui setelah selesai.");
    } catch {
      setMessage("Asisten belum dapat disiapkan. Coba lagi beberapa saat lagi.");
    } finally { setBusyAction(null); }
  }

  async function run(action: RunAction) {
    if (busy) return;
    if (action === "delete" && !window.confirm("Hapus asisten pesantren ini beserta koneksi Telegram dan WhatsApp-nya?")) return;
    setBusyAction(action);
    setMessage(`${actionLabels[action]} sedang diproses...`);
    try {
      await api(`/api/v1/tenants/${tenantId}/hermes/action`, { method: "POST", body: JSON.stringify({ action }) });
      if (action === "delete") { await api(`/api/v1/tenants/${tenantId}/bot`, { method: "DELETE" }).catch(() => undefined); await api(`/api/v1/tenants/${tenantId}/whatsapp`, { method: "DELETE" }).catch(() => undefined); }
      if (action === "start" || action === "stop") await refreshUntil(action === "start"); else await refresh();
      setMessage(`${actionLabels[action]} berhasil.`);
    } catch {
      setMessage(`${actionLabels[action]} gagal.`);
    } finally { setBusyAction(null); }
  }

  function openModal() {
    setMode(bot?.access_mode ?? "TESTING");
    setAllowed((bot?.allowed_user_ids ?? []).join(", "));
    setUsername(bot?.username ?? username);
    setFormError("");
    setModal(true);
  }

  async function connect(event: React.FormEvent) {
    event.preventDefault();
    const ids = allowed.split(/[\s,]+/).filter(Boolean);
    if (mode === "TESTING" && (!ids.length || ids.some(item => !/^\d+$/.test(item)))) { setFormError("Masukkan ID Telegram berupa angka. Pisahkan dengan koma untuk lebih dari satu."); return; }
    // The bot's numeric ID is the part of the code before the colon, so there is no need to ask for it separately.
    const tokenMatch = /^(\d+):[\w-]{20,}$/.exec(token.trim());
    if (!tokenMatch) { setFormError("Kode bot tidak sesuai. Salin kembali kode lengkap dari BotFather."); return; }
    setFormError("");
    setBusyAction("gateway");
    try {
      await api(`/api/v1/tenants/${tenantId}/bot`, { method: "PUT", body: JSON.stringify({ telegram_bot_id: Number(tokenMatch[1]), username, credential_ref: token.trim(), access_mode: mode, allowed_user_ids: mode === "TESTING" ? ids.map(Number) : [] }) });
      await api(`/api/v1/tenants/${tenantId}/hermes/action`, { method: "POST", body: JSON.stringify({ action: "start" }) });
      await refreshUntil(true);
      setModal(false); setToken("");
      setMessage("Telegram berhasil dihubungkan dan asisten diaktifkan.");
    } catch { setMessage("Telegram gagal dihubungkan. Periksa kode bot dan pastikan asisten sudah disiapkan."); setModal(false); }
    finally { setBusyAction(null); }
  }

  async function disconnect() {
    if (!window.confirm("Lepas koneksi Telegram dari asisten pesantren ini?")) return;
    setBusyAction("gateway-delete");
    try {
      await api(`/api/v1/tenants/${tenantId}/bot`, { method: "DELETE" });
      // Show the removal right away; stopping or restarting the assistant below can take a while or fail.
      setBot(null);
    } catch { setMessage("Koneksi Telegram gagal dilepas."); setBusyAction(null); return; }
    try {
      // The other channel keeps serving visitors; only a restart drops the removed bot.
      await api(`/api/v1/tenants/${tenantId}/hermes/action`, { method: "POST", body: JSON.stringify({ action: wa ? (wa.status === "ACTIVE" ? "restart" : "stop") : "stop" }) });
      setMessage("Koneksi Telegram berhasil dilepas.");
    } catch { setMessage("Koneksi Telegram dilepas, tetapi asisten belum berhasil dihentikan atau dimulai ulang. Coba nonaktifkan secara manual."); }
    finally { await refresh().catch(() => undefined); setBusyAction(null); }
  }

  // The QR code appears as soon as the dialog opens; poll until the phone has scanned it.
  useEffect(() => {
    if (!waModal || pairing.status === "connected" || pairing.status === "error" || pairing.status === "none") return;
    const timer = window.setInterval(() => { void api<Pairing>(`/api/v1/tenants/${tenantId}/whatsapp/pairing`).then(next => { if (next) setPairing(next); }).catch(() => undefined); }, 1500);
    return () => window.clearInterval(timer);
  }, [waModal, pairing.status, tenantId]);

  async function startPairing() {
    setPairing({ status: "starting" });
    try { setPairing((await api<Pairing>(`/api/v1/tenants/${tenantId}/whatsapp/pairing`, { method: "POST" })) ?? { status: "starting" }); }
    catch { setPairing({ status: "error", error: "failed" }); }
  }

  function openWaModal() {
    setWaMode(wa?.access_mode ?? "TESTING");
    setWaAllowed((wa?.allowed_numbers ?? []).join(", "));
    setWaError("");
    setWaModal(true);
    void startPairing();
  }

  function closeWaModal() {
    setWaModal(false);
    setPairing({ status: "none" });
    void api(`/api/v1/tenants/${tenantId}/whatsapp/pairing`, { method: "DELETE" }).catch(() => undefined);
  }

  async function connectWhatsApp(event: React.FormEvent) {
    event.preventDefault();
    const numbers = waAllowed.split(/[\s,]+/).filter(Boolean);
    if (waMode === "TESTING" && (!numbers.length || numbers.some(item => !/^\+?[\d\s-]{8,18}$/.test(item)))) { setWaError("Masukkan nomor WhatsApp berupa angka, boleh diawali 62 atau 0. Pisahkan dengan koma untuk lebih dari satu."); return; }
    setWaError("");
    setBusyAction("whatsapp");
    try {
      await api(`/api/v1/tenants/${tenantId}/whatsapp`, { method: "PUT", body: JSON.stringify({ access_mode: waMode, allowed_numbers: waMode === "TESTING" ? numbers : [] }) });
      await api(`/api/v1/tenants/${tenantId}/hermes/action`, { method: "POST", body: JSON.stringify({ action: "start" }) });
      await refreshUntil(true);
      setWaModal(false); setPairing({ status: "none" });
      setMessage("WhatsApp berhasil dihubungkan dan asisten diaktifkan.");
    } catch { setMessage("WhatsApp gagal dihubungkan. Pindai ulang kode QR dan pastikan asisten sudah disiapkan."); setWaModal(false); setPairing({ status: "none" }); }
    finally { setBusyAction(null); }
  }

  async function disconnectWhatsApp() {
    if (!window.confirm("Lepas koneksi WhatsApp dari asisten pesantren ini?")) return;
    setBusyAction("whatsapp-delete");
    try {
      await api(`/api/v1/tenants/${tenantId}/whatsapp`, { method: "DELETE" });
      setWa(null);
    } catch { setMessage("Koneksi WhatsApp gagal dilepas."); setBusyAction(null); return; }
    try {
      await api(`/api/v1/tenants/${tenantId}/hermes/action`, { method: "POST", body: JSON.stringify({ action: bot ? (bot.status === "ACTIVE" ? "restart" : "stop") : "stop" }) });
      setMessage("Koneksi WhatsApp berhasil dilepas. Hapus juga perangkat tertaut di aplikasi WhatsApp (Pengaturan → Perangkat tertaut).");
    } catch { setMessage("Koneksi WhatsApp dilepas, tetapi asisten belum berhasil dihentikan atau dimulai ulang. Coba nonaktifkan secara manual."); }
    finally { await refresh().catch(() => undefined); setBusyAction(null); }
  }

  if (!tenant) return <div className="panel panel-pad agent-card"><div className="small-empty"><LoaderCircle size={20} className="spin" /><span>Memuat status asisten…</span></div></div>;

  return <div className="panel panel-pad agent-card">
    <div className="card-heading"><div><h2 className="section-title">Asisten AI</h2><p className="section-hint">{canManage ? "Siapkan asisten, hubungkan Telegram dan/atau WhatsApp, lalu aktifkan." : "Status asisten pesantren Anda."}</p></div><span className={overall.tone}>{overall.label}</span></div>
    {showKnowledge && <KnowledgeBanner status={knowledge} />}
    <div className="agent-step"><span className="agent-step-no">1</span><div className="service-icon"><Sparkles size={18} /></div><div className="agent-step-body"><strong>Pengaturan asisten</strong><span>Identitas dan pengetahuan asisten</span></div><span className={statusClass(tenant.provisioning_status)}>{provisionLabel(tenant.provisioning_status)}</span>{canManage && !provisioned && <button className="button button-primary" type="button" disabled={busy} onClick={() => void provision()}>{busyAction === "provision" ? <LoaderCircle size={14} className="spin" /> : <Sparkles size={14} />}Siapkan asisten</button>}</div>
    <div className="agent-step"><span className="agent-step-no">2</span><div className="service-icon telegram-icon"><Bot size={18} /></div><div className="agent-step-body"><strong>Saluran Telegram</strong><span>{bot ? `@${bot.username} · ${bot.access_mode === "PUBLIC" ? "Publik" : `Uji coba (${bot.allowed_user_ids?.length ?? 0} penguji)`}` : provisioned ? "Belum ada bot yang terhubung" : "Siapkan asisten terlebih dahulu"}</span></div><span className={bot ? statusClass(bot.status) : "status status-muted"}>{bot ? botStatusLabel(bot.status) : "Belum terhubung"}</span>{canManage && <button className="button" type="button" disabled={!provisioned || busy} onClick={openModal}><Bot size={14} />{bot ? "Ganti bot" : "Hubungkan Telegram"}</button>}{canManage && bot && <button className="button button-ghost button-danger-text" type="button" disabled={busy} onClick={() => void disconnect()}>{busyAction === "gateway-delete" ? <LoaderCircle size={14} className="spin" /> : <Trash2 size={14} />}Lepas</button>}</div>
    <div className="agent-step"><span className="agent-step-no">3</span><div className="service-icon whatsapp-icon"><MessageCircle size={18} /></div><div className="agent-step-body"><strong>Saluran WhatsApp</strong><span>{wa ? `+${wa.phone_number} · ${wa.access_mode === "PUBLIC" ? "Publik" : `Uji coba (${wa.allowed_numbers?.length ?? 0} penguji)`}` : provisioned ? "Belum ada nomor yang terhubung" : "Siapkan asisten terlebih dahulu"}</span></div><span className={wa ? statusClass(wa.status) : "status status-muted"}>{wa ? botStatusLabel(wa.status) : "Belum terhubung"}</span>{canManage && <button className="button" type="button" disabled={!provisioned || busy} onClick={openWaModal}><MessageCircle size={14} />{wa ? "Tautkan ulang" : "Hubungkan WhatsApp"}</button>}{canManage && wa && <button className="button button-ghost button-danger-text" type="button" disabled={busy} onClick={() => void disconnectWhatsApp()}>{busyAction === "whatsapp-delete" ? <LoaderCircle size={14} className="spin" /> : <Trash2 size={14} />}Lepas</button>}</div>
    {canManage && <div className="agent-step"><span className="agent-step-no">4</span><div className="service-icon"><Activity size={18} /></div><div className="agent-step-body"><strong>Aktifkan asisten</strong><span>{running ? `Asisten aktif dan melayani ${[bot?.status === "ACTIVE" && "Telegram", wa?.status === "ACTIVE" && "WhatsApp"].filter(Boolean).join(" dan ")}.` : hasChannel && provisioned ? "Siap diaktifkan." : "Selesaikan langkah 1 dan hubungkan Telegram atau WhatsApp."}</span></div><div className="action-grid agent-run-actions"><button className={running ? "button" : "button button-primary"} type="button" disabled={busy || (!running && (!provisioned || !hasChannel))} onClick={() => void run(running ? "stop" : "start")}>{busyAction === "start" || busyAction === "stop" ? <LoaderCircle size={14} className="spin" /> : running ? <Clock3 size={14} /> : <Sparkles size={14} />}{running ? "Nonaktifkan" : "Aktifkan"}</button><button className="button" type="button" disabled={busy || !running} onClick={() => void run("restart")}>{busyAction === "restart" ? <LoaderCircle size={14} className="spin" /> : <Activity size={14} />}Mulai ulang</button></div></div>}
    {canManage && provisioned && <div className="agent-footer"><button className="button button-ghost" type="button" disabled={busy} onClick={() => void run("update")}>{busyAction === "update" ? <LoaderCircle size={14} className="spin" /> : <Clock3 size={14} />}Perbarui asisten</button>{canDelete && <button className="button button-ghost button-danger-text" type="button" disabled={busy} onClick={() => void run("delete")}>{busyAction === "delete" ? <LoaderCircle size={14} className="spin" /> : <Trash2 size={14} />}Hapus asisten</button>}</div>}
    {message && <p className="action-feedback" role="status">{message}</p>}
    {(["start", "stop", "restart", "update", "delete", "provision"] as const).includes(busyAction as never) && <ProcessingOverlay title={busyAction === "provision" ? "Menyiapkan asisten" : actionLabels[busyAction as RunAction]} steps={busyAction === "stop" ? ["Menghentikan asisten", "Memperbarui status saluran", "Hampir selesai"] : busyAction === "delete" ? ["Menghapus koneksi saluran", "Membersihkan data asisten", "Hampir selesai"] : ["Menghubungi server asisten", "Menyiapkan lingkungan asisten", "Menyambungkan ke saluran", "Hampir selesai, memastikan semuanya berjalan"]} />}
    {busyAction === "gateway" && <ProcessingOverlay title="Menghubungkan Telegram" steps={["Menyimpan data bot Anda", "Menyiapkan asisten untuk bot ini", "Mengaktifkan asisten dan menyambung ke Telegram", "Hampir selesai, memastikan semuanya berjalan"]} />}
    {busyAction === "whatsapp" && <ProcessingOverlay title="Menghubungkan WhatsApp" steps={["Menyimpan pengaturan akses", "Menyiapkan asisten untuk nomor ini", "Mengaktifkan asisten dan menyambung ke WhatsApp", "Hampir selesai, memastikan semuanya berjalan"]} />}
    {modal && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget && !busy) setModal(false); }}><form className="modal panel panel-pad" onSubmit={connect}>
      <div className="card-heading"><div><h2 className="section-title">Hubungkan bot Telegram</h2><p className="section-hint">Masukkan data bot Telegram Anda. Asisten akan diaktifkan setelah bot tersimpan.</p></div><button type="button" className="account-action" onClick={() => setModal(false)}><XCircle size={18} /></button></div>
      <label><span className="field-label">Kode bot</span><input className="input" type="password" required value={token} onChange={event => setToken(event.target.value)} placeholder="Tempel kode dari BotFather" /><span className="field-help">Kode ini diberikan @BotFather di Telegram saat Anda membuat bot.</span></label>
      <label><span className="field-label">Nama pengguna bot</span><input className="input" required value={username} onChange={event => setUsername(event.target.value.replace(/^@/, ""))} placeholder="contoh: pesantren_bot" /></label>
      <div><span className="field-label">Siapa yang boleh memakai bot?</span>
        <div className="access-toggle" role="radiogroup" aria-label="Mode akses bot"><button type="button" role="radio" aria-checked={mode === "TESTING"} className={mode === "TESTING" ? "active" : ""} onClick={() => setMode("TESTING")}><strong>Uji coba</strong><span>Hanya akun tertentu</span></button><button type="button" role="radio" aria-checked={mode === "PUBLIC"} className={mode === "PUBLIC" ? "active" : ""} onClick={() => setMode("PUBLIC")}><strong>Publik</strong><span>Semua orang</span></button></div>
        {mode === "TESTING" ? <label className="access-detail"><span className="field-label">ID Telegram penguji</span><input className="input" required inputMode="numeric" value={allowed} onChange={event => setAllowed(event.target.value)} placeholder="123456789, 987654321" /><span className="field-help">Pisahkan dengan koma untuk lebih dari satu orang. Hanya akun ini yang dapat mengobrol dengan bot. Untuk mengetahui ID, kirim pesan ke @userinfobot di Telegram.</span></label> : <p className="access-warning" role="note">Mode publik: siapa pun yang menemukan bot ini dapat mengobrol dengan asisten. Gunakan setelah pengujian selesai.</p>}
        {formError && <div className="error" role="alert" style={{ marginTop: 10 }}>{formError}</div>}
      </div>
      <div className="modal-actions"><button className="button" type="button" onClick={() => setModal(false)}>Batal</button><button className="button button-primary" type="submit" disabled={busyAction === "gateway"}>{busyAction === "gateway" ? <LoaderCircle size={15} className="spin" /> : <Bot size={15} />}Simpan &amp; aktifkan</button></div>
    </form></div>}
    {waModal && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget && !busy) closeWaModal(); }}><form className="modal panel panel-pad" onSubmit={connectWhatsApp}>
      <div className="card-heading"><div><h2 className="section-title">Hubungkan WhatsApp</h2><p className="section-hint">Pindai kode QR dengan WhatsApp di ponsel yang nomornya akan dipakai sebagai nomor asisten. Asisten akan diaktifkan setelah tersimpan.</p></div><button type="button" className="account-action" onClick={closeWaModal}><XCircle size={18} /></button></div>
      <div style={{ display: "grid", placeItems: "center", gap: 10, padding: "6px 0 14px" }}>
        <div style={{ width: 232, height: 232, display: "grid", placeItems: "center", background: "#fff", borderRadius: 12, border: "1px solid var(--border, #e5e7eb)", overflow: "hidden" }}>
          {pairing.status === "qr" && pairing.qr ? <img src={pairing.qr} alt="Kode QR WhatsApp" width={216} height={216} /> : pairing.status === "connected" ? <div style={{ display: "grid", placeItems: "center", gap: 6, color: "#15803d" }}><CheckCircle2 size={44} /><strong>Terhubung</strong><span style={{ fontSize: 13 }}>+{pairing.phone_number}</span></div> : pairing.status === "error" ? <div style={{ display: "grid", placeItems: "center", gap: 8, padding: 12, textAlign: "center" }}><XCircle size={32} /><span style={{ fontSize: 13 }}>{pairing.error === "expired" ? "Kode QR kedaluwarsa." : "Kode QR gagal dibuat."}</span><button className="button" type="button" onClick={() => void startPairing()}><RefreshCw size={14} />Buat ulang</button></div> : <div style={{ display: "grid", placeItems: "center", gap: 8 }}><LoaderCircle size={26} className="spin" /><span style={{ fontSize: 13 }}>Menyiapkan kode QR…</span></div>}
        </div>
        {pairing.status === "qr" && <span className="field-help" style={{ textAlign: "center" }}>Buka WhatsApp → Pengaturan → Perangkat tertaut → Tautkan perangkat, lalu pindai kode ini.</span>}
      </div>
      <div><span className="field-label">Siapa yang boleh memakai asisten?</span>
        <div className="access-toggle" role="radiogroup" aria-label="Mode akses WhatsApp"><button type="button" role="radio" aria-checked={waMode === "TESTING"} className={waMode === "TESTING" ? "active" : ""} onClick={() => setWaMode("TESTING")}><strong>Uji coba</strong><span>Hanya nomor tertentu</span></button><button type="button" role="radio" aria-checked={waMode === "PUBLIC"} className={waMode === "PUBLIC" ? "active" : ""} onClick={() => setWaMode("PUBLIC")}><strong>Publik</strong><span>Semua orang</span></button></div>
        {waMode === "TESTING" ? <label className="access-detail"><span className="field-label">Nomor WhatsApp penguji</span><input className="input" required inputMode="tel" value={waAllowed} onChange={event => setWaAllowed(event.target.value)} placeholder="6281234567890, 6285987654321" /><span className="field-help">Pisahkan dengan koma untuk lebih dari satu orang. Hanya nomor ini yang dapat mengobrol dengan asisten. Gunakan kode negara (62) atau awali dengan 0.</span></label> : <p className="access-warning" role="note">Mode publik: siapa pun yang mengirim pesan ke nomor ini akan dilayani asisten. Gunakan setelah pengujian selesai.</p>}
        {waError && <div className="error" role="alert" style={{ marginTop: 10 }}>{waError}</div>}
      </div>
      <div className="modal-actions"><button className="button" type="button" onClick={closeWaModal}>Batal</button><button className="button button-primary" type="submit" disabled={busyAction === "whatsapp" || pairing.status !== "connected"}>{busyAction === "whatsapp" ? <LoaderCircle size={15} className="spin" /> : <MessageCircle size={15} />}Simpan &amp; aktifkan</button></div>
    </form></div>}
  </div>;
}
