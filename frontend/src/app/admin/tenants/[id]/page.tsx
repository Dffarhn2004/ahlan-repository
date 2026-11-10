"use client";

import { Pagination, usePagination, type Paged } from "@/components/pagination";
import Link from "next/link";
import type { ComponentType } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Activity, ArrowLeft, Bot, ChevronDown, ChevronUp, Clock3, Eye, FileText, KeyRound, LoaderCircle, MapPin, MessageSquare, Pencil, Plus, Save, Search, ShieldCheck, Sparkles, Trash2, UploadCloud, UserRound, Users, XCircle } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { api } from "@/lib/api";
import { can, loadAuth, type AuthContext } from "@/lib/auth";
import { provisionLabel } from "@/lib/labels";
import { KnowledgeUploadModal } from "@/components/knowledge-upload-modal";
import { AssistantPanel } from "@/components/assistant-panel";
import { WhatsAppRouting } from "@/components/whatsapp-routing";
import { KnowledgeBanner, ReadBadge, useKnowledgeStatus } from "@/components/knowledge-status";
import { ShareLinkCard } from "@/components/share-link-card";
import { OfficialLinks } from "@/components/official-links";
import { ContactPersons } from "@/components/contact-persons";
import { UnansweredQuestions } from "@/components/unanswered-questions";
import { TelegramExportButton } from "@/components/telegram-export";

type Tenant = { id: string; name: string; slug: string; lifecycle_status: string; provisioning_status: string };
type Member = { user_id?: string; email?: string; role: string };
type FileItem = { id: string; filename: string; size_bytes: number };
type Bot = { telegram_bot_id: number; username: string; status: string; credential_ref?: string; access_mode?: "TESTING" | "PUBLIC"; allowed_user_ids?: number[] };
type Log = { id: string; action: string; actor_type: string; target_type?: string; occurred_at: string };
 type Conversation = { id: string; telegram_chat_id?: number; telegram_user_id?: number; username?: string; status: string; last_message_at: string };
type TelegramMessage = { id: string; sender_type: "USER" | "BOT" | "ADMIN"; message_text: string; created_at: string; telegram_message_id?: number };
type ConversationGroup = Conversation & { turns: number; unanswered: number; last_question: string };
type TelegramLogRow = TelegramMessage & { conversation: Conversation };
type TelegramTurn = { id: string; conversation: Conversation; created_at: string; question: string; answer: string; answered_at?: string };
type Tab = "profile" | "hermes" | "log" | "user";
type HermesSubtab = "assistant" | "whatsapp" | "knowledge" | "links" | "contacts" | "questions" | "telegram";
const hermesSubtabs: { value: HermesSubtab; label: string }[] = [
  { value: "assistant", label: "Asisten" },
  { value: "whatsapp", label: "Admin WhatsApp" },
  { value: "knowledge", label: "Pengetahuan" },
  { value: "links", label: "Tautan resmi" },
  { value: "contacts", label: "Kontak person" },
  { value: "questions", label: "Pertanyaan" },
  { value: "telegram", label: "Log Telegram" },
];

const tabs: { value: Tab; label: string; icon: ComponentType<{ size?: number; strokeWidth?: number }> }[] = [
  { value: "profile", label: "Profil tenant", icon: BuildingIcon },
  { value: "hermes", label: "Asisten & pengetahuan", icon: Bot },
  { value: "log", label: "Aktivitas", icon: Activity },
  { value: "user", label: "Pengguna", icon: Users },
];

function BuildingIcon(props: { size?: number; strokeWidth?: number }) { return <Sparkles {...props} />; }
function statusTone(value: string) { const text = value.toLowerCase(); return text.includes("fail") || text.includes("suspend") || text.includes("error") ? "status status-danger" : text.includes("pending") || text.includes("queue") || text.includes("progress") ? "status status-muted" : "status"; }
function pretty(value: string) { return value.toLowerCase().replaceAll("_", " ").replace(/\b\w/g, char => char.toUpperCase()); }
function dayLabel(iso: string) {
  const date = new Date(iso), today = new Date(), yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);
  if (date.toDateString() === today.toDateString()) return "Hari ini";
  if (date.toDateString() === yesterday.toDateString()) return "Kemarin";
  return date.toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
}
function formatBytes(bytes: number) { return bytes < 1024 * 1024 ? `${Math.max(1, Math.round(bytes / 1024))} KB` : `${(bytes / (1024 * 1024)).toFixed(1)} MB`; }

export default function TenantDetail({ params }: { params: Promise<{ id: string }> }) {
  const [id, setId] = useState("");
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [tab, setTab] = useState<Tab>("profile");
  const [hermesSubtab, setHermesSubtab] = useState<HermesSubtab>("assistant");
  const [openQuestions, setOpenQuestions] = useState(0);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
  const [conversationMessages, setConversationMessages] = useState<TelegramMessage[]>([]);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [fullLog, setFullLog] = useState<TelegramLogRow[]>([]);
  const [fullTurns, setFullTurns] = useState<TelegramTurn[]>([]);
  const [fullLogLoading, setFullLogLoading] = useState(false);
  const [logView, setLogView] = useState<"table" | "grouped">("table");
  const [detailTurn, setDetailTurn] = useState<TelegramTurn | null>(null);
  const [profile, setProfile] = useState<Record<string, string>>({});
  const [members, setMembers] = useState<Member[]>([]);
  const [files, setFiles] = useState<FileItem[]>([]);
  const [bot, setBot] = useState<Bot | null>(null);
  const [logs, setLogs] = useState<Log[]>([]);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [memberRole, setMemberRole] = useState("ADMIN");
  const [memberModal, setMemberModal] = useState(false);
  const [profileModal, setProfileModal] = useState(false);
  const [editingMember, setEditingMember] = useState<Member | null>(null);
  const [state, setState] = useState<"loading" | "missing" | "ready">("loading");
  const [syncing, setSyncing] = useState(false);
  const [syncMessage, setSyncMessage] = useState("");
  const [uploadModal, setUploadModal] = useState(false);
  const [uploadingFiles, setUploadingFiles] = useState(false);
  const { status: knowledge, finished: knowledgeFinished, refresh: refreshKnowledge } = useKnowledgeStatus(id);
  const filesPaged = usePagination(files, 10);
  const turnsPaged = usePagination(fullTurns, 10);
  const [groupQuery, setGroupQuery] = useState("");
  const [groupStatus, setGroupStatus] = useState("ALL");
  const [groupSort, setGroupSort] = useState<"recent" | "most" | "unanswered">("recent");
  const [chatSearch, setChatSearch] = useState({ id: "", text: "" });
  const chatQuery = chatSearch.id === selectedConversation?.id ? chatSearch.text : "";
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [groups, setGroups] = useState<ConversationGroup[]>([]);
  const [groupTotal, setGroupTotal] = useState(0);
  const [statusOptions, setStatusOptions] = useState<string[]>([]);
  const [loadedKey, setLoadedKey] = useState("");
  const [groupPaging, setGroupPaging] = useState({ page: 1, size: 10, key: "" });
  const groupKey = `${debouncedQuery}|${groupStatus}|${groupSort}`;
  const groupPage = groupPaging.key === groupKey ? groupPaging.page : 1;
  const groupPages = Math.max(1, Math.ceil(groupTotal / groupPaging.size));
  const groupsPaged: Paged<ConversationGroup> = {
    rows: groups, page: groupPage, pages: groupPages, total: groupTotal, size: groupPaging.size,
    from: groupTotal ? (groupPage - 1) * groupPaging.size + 1 : 0, to: Math.min(groupPage * groupPaging.size, groupTotal),
    setPage: next => setGroupPaging(current => ({ ...current, page: Math.min(Math.max(1, next), groupPages), key: groupKey })),
    setSize: next => setGroupPaging({ page: 1, size: next, key: groupKey }),
  };
  const requestKey = `${id}|${groupKey}|${groupPage}|${groupPaging.size}`;
  const groupsLoading = loadedKey !== requestKey;
  const selectedIndex = selectedConversation ? groups.findIndex(item => item.id === selectedConversation.id) : -1;

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(groupQuery.trim()), 300);
    return () => window.clearTimeout(timer);
  }, [groupQuery]);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    const search = new URLSearchParams({ page: String(groupPage), size: String(groupPaging.size), sort: groupSort });
    if (debouncedQuery) search.set("q", debouncedQuery);
    if (groupStatus !== "ALL") search.set("status", groupStatus);
    api<{ data: ConversationGroup[]; total: number; statuses: string[] }>(`/api/v1/tenants/${id}/telegram/conversations?${search}`)
      .then(result => { if (cancelled) return; setGroups(result?.data ?? []); setGroupTotal(result?.total ?? 0); setStatusOptions(result?.statuses ?? []); })
      .catch(() => { if (!cancelled) { setGroups([]); setGroupTotal(0); } })
      .finally(() => { if (!cancelled) setLoadedKey(requestKey); });
    return () => { cancelled = true; };
  }, [id, requestKey, debouncedQuery, groupStatus, groupSort, groupPage, groupPaging.size]);
  const visibleMessages = useMemo(() => {
    const query = chatQuery.trim().toLowerCase();
    return query ? conversationMessages.filter(message => message.message_text.toLowerCase().includes(query)) : conversationMessages;
  }, [conversationMessages, chatQuery]);

  useEffect(() => {
    if (!selectedConversation) return;
    const onKey = (event: KeyboardEvent) => { if (event.key === "Escape") setSelectedConversation(null); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [selectedConversation]);
  const logsPaged = usePagination(logs, 10);
  const membersPaged = usePagination(members, 10);

  useEffect(() => {
    void params.then(async ({ id: tenantId }) => {
      setId(tenantId);
      const selected = new URLSearchParams(window.location.search).get("tab") as Tab | null;
      if (selected && tabs.some(item => item.value === selected)) setTab(selected);
      const sub = new URLSearchParams(window.location.search).get("sub") as HermesSubtab | null;
      if (sub && hermesSubtabs.some(item => item.value === sub)) setHermesSubtab(sub);
      try {
        const [item, authData, memberData, profileData, fileData, logData] = await Promise.all([
          api<Tenant>(`/api/v1/tenants/${tenantId}`), loadAuth(),
          api<{ data: Member[] }>(`/api/v1/tenants/${tenantId}/members`),
          api<Record<string, string>>(`/api/v1/tenants/${tenantId}/profile`).catch(() => null),
          api<{ data: FileItem[] }>(`/api/v1/tenants/${tenantId}/files`),
          api<{ data: Log[] }>(`/api/v1/audit-logs?tenant_id=${tenantId}`),
        ]);
         setTenant(item); setAuth(authData); setMembers(memberData?.data ?? []); setProfile(profileData ?? {}); setFiles(fileData?.data ?? []); setLogs(logData?.data ?? []); setState("ready");
        api<Bot>(`/api/v1/tenants/${tenantId}/bot`).then(setBot).catch(() => setBot(null));
        api<{ counts?: { OPEN?: number } }>(`/api/v1/tenants/${tenantId}/questions`).then(result => setOpenQuestions(result?.counts?.OPEN ?? 0)).catch(() => undefined);
        api<{ data: Conversation[] }>(`/api/v1/conversations?tenant_id=${tenantId}`).then(result => { const next = result?.data ?? []; setConversations(next); setSelectedConversation(null); }).catch(() => { setConversations([]); setSelectedConversation(null); });
      } catch { setState("missing"); }
    });
  }, [params]);

  useEffect(() => {
    if (!selectedConversation) { setConversationMessages([]); return; }
    let cancelled = false;
    setMessagesLoading(true);
    api<{ data: TelegramMessage[] }>(`/api/v1/conversations/${selectedConversation.id}/messages`)
      .then(result => { if (!cancelled) setConversationMessages(result?.data ?? []); })
      .catch(() => { if (!cancelled) setConversationMessages([]); })
      .finally(() => { if (!cancelled) setMessagesLoading(false); });
    return () => { cancelled = true; };
  }, [selectedConversation]);

  useEffect(() => {
    if (!conversations.length) { setFullLog([]); return; }
    let cancelled = false;
    setFullLogLoading(true);
    Promise.all(conversations.map(async conversation => {
      const result = await api<{ data: TelegramMessage[] }>(`/api/v1/conversations/${conversation.id}/messages`);
      return (result?.data ?? []).map(message => ({ ...message, conversation }));
    })).then(groups => { if (cancelled) return; const rows = groups.flat(); setFullLog(rows.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())); const turns: TelegramTurn[] = []; for (const conversation of conversations) { const messages = rows.filter(item => item.conversation.id === conversation.id).sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()); for (let index = 0; index < messages.length; index += 1) { const message = messages[index]; if (message.sender_type === "USER") { const reply = messages[index + 1]?.sender_type === "BOT" ? messages[index + 1] : undefined; turns.push({ id: message.id, conversation, created_at: message.created_at, question: message.message_text, answer: reply?.message_text ?? "Belum ada balasan asisten", answered_at: reply?.created_at }); } } } setFullTurns(turns.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())); })
      .catch(() => { if (!cancelled) { setFullLog([]); setFullTurns([]); } })
      .finally(() => { if (!cancelled) setFullLogLoading(false); });
    return () => { cancelled = true; };
  }, [conversations]);

  const refreshAssistant = useCallback(() => {
    void api<Tenant>(`/api/v1/tenants/${id}`).then(latest => { if (latest) setTenant(latest); }).catch(() => undefined);
    void api<Bot>(`/api/v1/tenants/${id}/bot`).then(setBot).catch(() => setBot(null));
  }, [id]);
  const selectTab = (value: Tab) => { setTab(value); window.history.replaceState({}, "", `?tab=${value}`); };
  const selectSubtab = (value: HermesSubtab) => { setHermesSubtab(value); window.history.replaceState({}, "", `?tab=hermes&sub=${value}`); };
  const saveProfile = async (event: React.FormEvent) => { event.preventDefault(); try { await api(`/api/v1/tenants/${id}/profile`, { method: "PUT", body: JSON.stringify({ ...profile, tenant_id: id }) }); setProfileModal(false); } catch { ; } };
  const saveMember = async (event: React.FormEvent) => { event.preventDefault(); try { if (editingMember) await api(`/api/v1/tenants/${id}/members/${editingMember.user_id}`, { method: "PATCH", body: JSON.stringify({ role: memberRole }) }); else await api(`/api/v1/tenants/${id}/members`, { method: "POST", body: JSON.stringify({ email, password, role: memberRole }) }); setEmail(""); setPassword(""); setMemberRole("ADMIN"); setEditingMember(null); setMemberModal(false); const result = await api<{ data: Member[] }>(`/api/v1/tenants/${id}/members`); setMembers(result?.data ?? []); } catch { setMemberModal(false); } };
  const createMember = saveMember;
  const syncDocuments = async () => {
    if (syncing) return;
    setSyncing(true);
    setSyncMessage("Memperbarui pengetahuan asisten dari semua dokumen dan folder...");
    try {
      const result = await api<{ status: string; files: number; images?: number }>(`/api/v1/tenants/${id}/hermes/sync`, { method: "POST" });
      setSyncMessage(`${result?.files ?? 0} dokumen berhasil ditambahkan ke pengetahuan asisten.`);
      refreshKnowledge();
    } catch {
      setSyncMessage("Pengetahuan asisten gagal diperbarui. Pastikan asisten sudah disiapkan.");
    } finally {
      setSyncing(false);
    }
  };
   const deleteFile = async (file: FileItem) => { if (!window.confirm(`Hapus dokumen ${file.filename}?`)) return; try { await api(`/api/v1/tenants/${id}/files/${file.id}`, { method: "DELETE" }); setFiles(current => current.filter(item => item.id !== file.id)); } catch { window.alert("Dokumen gagal dihapus."); } };
   const editFile = async (file: FileItem) => { const filename = window.prompt("Nama dokumen", file.filename)?.trim(); if (!filename || filename === file.filename) return; try { const updated = await api<FileItem>(`/api/v1/tenants/${id}/files/${file.id}`, { method: "PATCH", body: JSON.stringify({ filename }) }); if (updated) setFiles(current => current.map(item => item.id === file.id ? { ...item, filename: updated.filename } : item)); } catch { window.alert("Nama dokumen gagal diperbarui."); } };
  const uploadKnowledgeFiles = async (selectedFiles: File[]) => { if (!selectedFiles.length || uploadingFiles) return; setUploadingFiles(true); try { for (const file of selectedFiles) { const body = new FormData(); body.append("file", file); await api(`/api/v1/tenants/${id}/files`, { method: "POST", body }); } setFiles((await api<{ data: FileItem[] }>(`/api/v1/tenants/${id}/files`))?.data ?? []); setUploadModal(false); setSyncMessage(`${selectedFiles.length} file berhasil diunggah. Perbarui pengetahuan asisten agar file ini menjadi referensi.`); } catch { setSyncMessage("File gagal diunggah. Ukuran maksimum setiap file adalah 20 MB."); } finally { setUploadingFiles(false); } };

  if (state === "loading") return <AppShell title="Detail pesantren"><div className="tenant-loading-shell" aria-busy="true"><div className="skeleton-back"/><div className="skeleton-tenant-hero"><div className="skeleton-identity"><span className="skeleton-logo"/><div><span className="skeleton-line eyebrow"/><span className="skeleton-line title"/><span className="skeleton-line subtitle"/></div></div><div className="skeleton-statuses"><span/><span/></div></div><div className="skeleton-tabs"><span/><span/><span/><span/></div><div className="skeleton-grid"><div className="skeleton-card"><span className="skeleton-line heading"/><span className="skeleton-line body"/><span className="skeleton-line body short"/><span className="skeleton-button"/></div><div className="skeleton-card"><span className="skeleton-line heading"/><span className="skeleton-line body"/><span className="skeleton-line body short"/><span className="skeleton-button"/></div></div></div></AppShell>;
  if (state === "missing" || !tenant) return <AppShell title="Detail pesantren"><div className="detail-loading"><XCircle size={28} color="var(--danger)" /><strong>Pesantren tidak ditemukan</strong><span>Tenant mungkin sudah dihapus atau akses Anda terbatas.</span><Link href="/admin" className="button">Kembali ke pesantren</Link></div></AppShell>;

  return <AppShell title="Detail pesantren">
    <Link href="/admin" className="back-link"><ArrowLeft size={15} /> Kembali ke semua pesantren</Link>
    <div className="tenant-hero">
      <div className="tenant-identity"><div className="tenant-logo"><BuildingIcon size={27} /></div><div><p className="page-eyebrow">Workspace tenant</p><h1>{tenant.name}</h1><p className="tenant-slug">/{tenant.slug} <span>-</span> ID {tenant.id.slice(0, 8)}</p></div></div>
      <div className="tenant-statuses"><span className={statusTone(tenant.lifecycle_status)}>{pretty(tenant.lifecycle_status)}</span><span className={statusTone(tenant.provisioning_status)}>{provisionLabel(tenant.provisioning_status)}</span></div>
    </div>
    <div className="tenant-overview"><div><span className="overview-label">Layanan</span><strong><span className="live-dot" />{pretty(tenant.lifecycle_status)}</strong><small>Lifecycle tenant</small></div><div><span className="overview-label">Asisten AI</span><strong>{bot ? pretty(bot.status) : "Belum terhubung"}</strong><small>{bot ? `@${bot.username}` : "Siap dikonfigurasi"}</small></div><div><span className="overview-label">Bahan pengetahuan</span><strong>{files.length} <small>file</small></strong><small>Dokumen terdaftar</small></div><div><span className="overview-label">Pengguna</span><strong>{members.length} <small>akun</small></strong><small>Member tenant</small></div></div>
    <nav className="detail-tabs" aria-label="Navigasi detail tenant">{tabs.map(({ value, label, icon: Icon }) => <button key={value} className={tab === value ? "active" : ""} onClick={() => selectTab(value)}><Icon size={16} />{label}</button>)}</nav>
    {tab === "profile" && <ShareLinkCard slug={tenant.slug} />}
    {tab === "profile" && <section className="detail-grid"><div className="panel panel-pad"><div className="card-heading"><div><h2 className="section-title">Informasi profil</h2><p className="section-hint">Identitas yang tampil di ruang kerja pesantren.</p></div>{can(auth, "tenant:update") && <button className="button button-ghost" type="button" onClick={() => setProfileModal(true)}><Pencil size={15} />Edit profil</button>}</div><div className="detail-form profile-view"><div className="form-row"><div><span className="field-label">Nama pesantren</span><p className="profile-value">{profile.display_name || tenant.name}</p></div><div><span className="field-label">Email utama</span><p className="profile-value">{profile.email || "Belum diatur"}</p></div></div><div><span className="field-label">Alamat</span><p className="profile-value">{profile.address || "Belum diatur"}</p></div><div className="form-row"><div><span className="field-label">Telepon</span><p className="profile-value">{profile.phone || "Belum diatur"}</p></div><div><span className="field-label">Website</span><p className="profile-value">{profile.website || "Belum diatur"}</p></div></div><div><span className="field-label">Tentang pesantren</span><p className="profile-value">{profile.description || "Belum diatur"}</p></div><div><span className="field-label">Link Google Maps</span><p className="profile-value">{profile.google_maps_url ? <a href={profile.google_maps_url} target="_blank" rel="noopener noreferrer">{profile.google_maps_url}</a> : "Belum diatur"}</p></div></div></div><div className="panel panel-pad accent-card"><div className="accent-icon"><ShieldCheck size={19} /></div><h2 className="section-title">Tenant terverifikasi</h2><p className="section-hint">Workspace ini dikelola dari control plane AHLAN. Semua perubahan penting tercatat di audit log.</p><Link className="text-link" href="?tab=log" onClick={() => selectTab("log")}>Lihat aktivitas <ArrowLeft size={14} className="rotate-180" /></Link></div></section>}
    {tab === "hermes" && <><nav className="hermes-subtabs" aria-label="Navigasi asisten">{hermesSubtabs.map(item => <button key={item.value} className={hermesSubtab === item.value ? "active" : ""} type="button" onClick={() => selectSubtab(item.value)}>{item.label}{item.value === "questions" && openQuestions > 0 && <span className="status status-danger" style={{ marginLeft: 6, padding: "0 7px" }} aria-label={`${openQuestions} pertanyaan belum terjawab`}>{openQuestions}</span>}</button>)}</nav><div className={`hermes-tab-panel ${hermesSubtab}`}><section className="hermes-layout">
{hermesSubtab === "assistant" && <AssistantPanel tenantId={id} canManage={can(auth, "assistant:manage")} canDelete={auth?.role === "SUPERADMIN"} onChange={refreshAssistant} showKnowledge={false} />}
{hermesSubtab === "whatsapp" && <WhatsAppRouting tenantId={id} canEdit={can(auth, "assistant:manage")} />}
{hermesSubtab === "knowledge" && <div className="panel panel-pad"><div className="card-heading"><div><h2 className="section-title">Bahan pengetahuan</h2><p className="section-hint">Dokumen yang menjadi referensi asisten.</p></div><div className="knowledge-actions"><FileText size={18} color="var(--faint)" />{can(auth, "file:create") && <button className="button button-primary" type="button" onClick={() => setUploadModal(true)}><UploadCloud size={14} />Unggah file</button>}{can(auth, "file:update") && <button className="button button-ghost" type="button" disabled={syncing || !files.length} onClick={() => void syncDocuments()}>{syncing ? <LoaderCircle size={14} className="spin" /> : <Sparkles size={14} />}{syncing ? "Memperbarui..." : "Perbarui pengetahuan asisten"}</button>}</div></div>{syncMessage && <p className="action-feedback" role="status">{syncMessage}</p>}<KnowledgeBanner status={knowledge} finished={knowledgeFinished} onRetry={can(auth, "file:update") ? () => void syncDocuments() : undefined} busy={syncing} />{files.length ? <><div className="file-list">{filesPaged.rows.map(file => <div className="file-item" key={file.id}><div className="file-icon"><FileText size={16} /></div><div className="file-name"><strong>{file.filename}</strong><span>{formatBytes(file.size_bytes)}</span></div><ReadBadge state={knowledge?.files[file.id]} /><div className="file-actions"><a className="file-action" href={`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/tenants/${id}/files/${file.id}`} target="_blank" rel="noreferrer" aria-label={`Lihat ${file.filename}`}><Eye size={14}/><span>Lihat</span></a><button className="file-action" type="button" onClick={() => void editFile(file)} aria-label={`Edit ${file.filename}`}><Pencil size={14}/><span>Edit</span></button><button className="file-action file-action-danger" type="button" onClick={() => void deleteFile(file)} aria-label={`Hapus ${file.filename}`}><Trash2 size={14}/><span>Hapus</span></button></div></div>)}</div><Pagination paged={filesPaged} label="file"/></> : <div className="small-empty"><FileText size={24} /><span>Belum ada bahan pengetahuan.</span></div>}</div>}
{hermesSubtab === "links" && <OfficialLinks tenantId={id} canEdit={can(auth, "tenant:update")} />}
{hermesSubtab === "contacts" && <ContactPersons tenantId={id} canEdit={can(auth, "tenant:update")} />}
{hermesSubtab === "questions" && <UnansweredQuestions tenantId={id} canEdit={can(auth, "tenant:update")} onOpenCount={setOpenQuestions} />}</section>
    {hermesSubtab === "telegram" && <section className="panel panel-pad telegram-log-panel"><div className="card-heading"><div><h2 className="section-title">Log Telegram</h2><p className="section-hint">Satu baris berisi pertanyaan pengguna dan balasan asisten.</p></div><div className="telegram-log-actions"><TelegramExportButton tenantId={id} conversations={conversations} disabled={!fullTurns.length || fullLogLoading} ghost /><MessageSquare size={19} color="var(--brand)" /></div></div><div className="telegram-log-mode" role="tablist" aria-label="Mode tampilan log Telegram"><button className={logView === "table" ? "active" : ""} type="button" role="tab" aria-selected={logView === "table"} onClick={() => setLogView("table")}>Full log</button><button className={logView === "grouped" ? "active" : ""} type="button" role="tab" aria-selected={logView === "grouped"} onClick={() => setLogView("grouped")}>Grouping percakapan</button></div>{!conversations.length ? <div className="small-empty"><MessageSquare size={24} /><strong>Belum ada percakapan Telegram</strong><span>Log akan muncul setelah pengguna mengirim pesan ke bot tenant ini.</span></div> : logView === "table" ? <div className="telegram-full-log">{fullLogLoading ? <div className="telegram-message-empty">Memuat seluruh log pesan…</div> : fullTurns.length ? <div className="telegram-full-table" role="table" aria-label="Full log Telegram"><div className="telegram-full-head" role="row"><span>Waktu</span><span>Pengguna</span><span>Pertanyaan pengguna</span><span>Balasan asisten</span><span>Aksi</span></div>{turnsPaged.rows.map(turn => <div className="telegram-full-row telegram-turn-row" role="row" key={turn.id}><time>{new Date(turn.created_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}</time><span className="telegram-log-person">{turn.conversation.username ? `@${turn.conversation.username}` : "Pengguna Telegram"}</span><p className="telegram-turn-question">{turn.question}</p><p className={`telegram-turn-answer ${turn.answer === "Belum ada balasan asisten" ? "pending" : ""}`}>{turn.answer}{turn.answered_at && <small>Dijawab {new Date(turn.answered_at).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" })}</small>}</p><button className="file-action telegram-log-action" type="button" onClick={() => setDetailTurn(turn)}><Eye size={14} /><span>Lihat detail</span></button></div>)}</div> : <div className="telegram-message-empty">Belum ada pertanyaan tercatat.</div>}{!fullLogLoading && <Pagination paged={turnsPaged} label="pertanyaan"/>}</div> : <>
<div className="tg-toolbar"><label className="tg-search"><Search size={14} /><input className="input" type="search" value={groupQuery} onChange={event => setGroupQuery(event.target.value)} placeholder="Cari username, Chat ID, atau isi pertanyaan" aria-label="Cari percakapan" /></label><select className="input tg-select" value={groupStatus} onChange={event => setGroupStatus(event.target.value)} aria-label="Filter status"><option value="ALL">Semua status</option>{statusOptions.map(status => <option key={status} value={status}>{pretty(status)}</option>)}</select><select className="input tg-select" value={groupSort} onChange={event => setGroupSort(event.target.value as typeof groupSort)} aria-label="Urutkan percakapan"><option value="recent">Terbaru</option><option value="most">Paling banyak pertanyaan</option><option value="unanswered">Belum dijawab terbanyak</option></select></div>
{groups.length ? <div className="tg-group-list" aria-label="Grouping percakapan Telegram">{groups.map(item => { const name = item.username ? `@${item.username}` : "Pengguna Telegram"; const active = selectedConversation?.id === item.id; return <button className={`tg-group-card ${active ? "selected" : ""}`} type="button" key={item.id} aria-pressed={active} onClick={() => setSelectedConversation(item)}><span className="tg-group-avatar">{(item.username ?? "T").charAt(0).toUpperCase()}</span><span className="tg-group-main"><span className="tg-group-top"><strong>{name}</strong><span className="tg-group-id">Chat ID {item.telegram_chat_id ?? "—"}</span></span><span className="tg-group-snippet">{item.last_question || "Belum ada pertanyaan tercatat"}</span></span><span className="tg-group-meta"><span className="tg-group-count"><MessageSquare size={12} />{item.turns} pertanyaan</span>{item.unanswered > 0 && <span className="tg-group-unanswered">{item.unanswered} belum dijawab</span>}<time>{new Date(item.last_message_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}</time></span><span className={`status ${item.status === "ACTIVE" ? "" : "status-muted"}`}>{pretty(item.status)}</span></button>; })}</div> : <div className="telegram-message-empty">{groupsLoading ? "Memuat percakapan…" : "Tidak ada percakapan yang cocok dengan pencarian atau filter."}</div>}
<Pagination paged={groupsPaged} label="percakapan"/></> }</section>}</div></>}{tab === "log" && <section className="panel panel-pad"><div className="card-heading"><div><h2 className="section-title">Aktivitas tenant</h2><p className="section-hint">Riwayat perubahan dan aktivitas operasional.</p></div><Clock3 size={18} color="var(--faint)" /></div>{can(auth, "audit:read") ? logs.length ? <><div className="log-list">{logsPaged.rows.map(log => <div className="log-item" key={log.id}><div className="log-dot"><Activity size={13} /></div><div><strong>{pretty(log.action)}</strong><span>{log.actor_type} {log.target_type ? `- ${log.target_type}` : ""}</span></div><time>{new Date(log.occurred_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}</time></div>)}</div><Pagination paged={logsPaged} label="aktivitas"/></> : <div className="small-empty"><Activity size={24} /><span>Belum ada aktivitas untuk tenant ini.</span></div> : <div className="small-empty"><ShieldCheck size={24} /><span>Anda tidak memiliki akses untuk melihat log.</span></div>}</section>}
    {tab === "user" && <section className="detail-grid"><div className="panel panel-pad"><div className="member-toolbar"><div><h2 className="section-title">Pengguna tenant</h2><p className="section-hint">Kelola akses dan peran pengguna workspace ini.</p></div>{can(auth, "member:manage") && <button className="button button-primary" type="button" onClick={() => { setEditingMember(null); setEmail(""); setPassword(""); setMemberRole("ADMIN"); setMemberModal(true); }}><Plus size={15} />Tambah pengguna</button>}</div>{members.length ? <div className="member-table-wrap"><div className="member-table-head"><span>Pengguna</span><span>Role</span><span>Status</span><span>Aksi</span></div><div className="member-list">{membersPaged.rows.map((member, index) => { const userId = member.user_id ?? ""; const memberEmail = member.email ?? "Pengguna"; return <div className="member-item" key={`${userId}-${memberEmail}-${index}`}><div className="table-person"><div className="avatar"><UserRound size={14} /></div><div><strong>{memberEmail}</strong><small>{userId ? `${userId.slice(0, 8)}...` : "ID tidak tersedia"}</small></div></div><span className={`role-badge role-${member.role.toLowerCase()}`}>{member.role}</span><span className="status">Aktif</span>{can(auth, "member:manage") && userId && <button className="button button-ghost edit-member-button" type="button" onClick={() => { setEditingMember(member); setEmail(memberEmail); setMemberRole(member.role); setPassword(""); setMemberModal(true); }} aria-label={`Edit ${memberEmail}`}><Pencil size={14} />Edit</button>}</div>; })}</div><Pagination paged={membersPaged} label="pengguna"/></div> : <div className="small-empty"><Users size={24} /><span>Belum ada pengguna tenant.</span></div>}</div><div className="panel panel-pad accent-card"><div className="accent-icon"><KeyRound size={19} /></div><h2 className="section-title">Akses aman</h2><p className="section-hint">Password sementara sebaiknya diganti setelah admin pertama kali masuk ke workspace.</p></div></section>}
    {memberModal && can(auth, "member:manage") && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) setMemberModal(false); }}><form className="modal panel panel-pad" onSubmit={createMember}><div className="card-heading"><div><h2 className="section-title">{editingMember ? "Edit pengguna" : "Tambah pengguna"}</h2><p className="section-hint">Atur akses pengguna tenant.</p></div><button className="account-action" type="button" onClick={() => setMemberModal(false)}><XCircle size={18} /></button></div><label><span className="field-label">Email</span><input className="input" type="email" required value={email} disabled={Boolean(editingMember)} onChange={e => setEmail(e.target.value)} /></label>{!editingMember && <label><span className="field-label">Password sementara</span><input className="input" type="password" required minLength={10} value={password} onChange={e => setPassword(e.target.value)} /></label>}<label><span className="field-label">Role</span><select className="input" value={memberRole} onChange={e => setMemberRole(e.target.value)}><option value="ADMIN">ADMIN</option><option value="STAFF">STAFF</option><option value="VIEWER">VIEWER</option></select></label><div className="modal-actions"><button className="button" type="button" onClick={() => setMemberModal(false)}>Batal</button><button className="button button-primary" type="submit"><Save size={15} />{editingMember ? "Simpan perubahan" : "Buat pengguna"}</button></div></form></div>}{}
    {profileModal && can(auth, "tenant:update") && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) setProfileModal(false); }}><form className="modal panel panel-pad" onSubmit={saveProfile}><div className="card-heading"><div><h2 className="section-title">Edit profil tenant</h2><p className="section-hint">Perubahan akan tercatat di audit log.</p></div><button className="account-action" type="button" onClick={() => setProfileModal(false)}><XCircle size={18} /></button></div><label><span className="field-label">Nama pesantren</span><input className="input" value={profile.display_name ?? tenant.name} onChange={e => setProfile({ ...profile, display_name: e.target.value })} required /></label><label><span className="field-label">Email utama</span><input className="input" type="email" value={profile.email ?? ""} onChange={e => setProfile({ ...profile, email: e.target.value })} placeholder="admin@pesantren.id" /></label><label><span className="field-label">Alamat</span><input className="input" value={profile.address ?? ""} onChange={e => setProfile({ ...profile, address: e.target.value })} placeholder="Alamat lengkap pesantren" /></label><div className="form-row"><label><span className="field-label">Telepon</span><input className="input" value={profile.phone ?? ""} onChange={e => setProfile({ ...profile, phone: e.target.value })} placeholder="+62" /></label><label><span className="field-label">Website</span><input className="input" value={profile.website ?? ""} onChange={e => setProfile({ ...profile, website: e.target.value })} placeholder="https://" /></label></div><label><span className="field-label">Tentang pesantren</span><textarea className="input" rows={3} value={profile.description ?? ""} onChange={e => setProfile({ ...profile, description: e.target.value })} placeholder="Ringkasan singkat yang dipakai asisten" /></label><label><span className="field-label">Link Google Maps</span><input className="input" type="url" value={profile.google_maps_url ?? ""} onChange={e => setProfile({ ...profile, google_maps_url: e.target.value })} placeholder="https://maps.app.goo.gl/…" /></label><div className="modal-actions"><button className="button" type="button" onClick={() => setProfileModal(false)}>Batal</button><button className="button button-primary" type="submit"><Save size={15} />Simpan perubahan</button></div></form></div>}
  {selectedConversation && <div className="tg-drawer-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) setSelectedConversation(null); }}><aside className="tg-drawer" role="dialog" aria-modal="true" aria-labelledby="tg-drawer-title">
    <header className="tg-drawer-head"><span className="tg-group-avatar">{(selectedConversation.username ?? "T").charAt(0).toUpperCase()}</span><div className="tg-drawer-title"><h3 id="tg-drawer-title">{selectedConversation.username ? `@${selectedConversation.username}` : "Pengguna Telegram"}</h3><span>Chat ID {selectedConversation.telegram_chat_id ?? "—"} · {conversationMessages.length} pesan</span></div><div className="tg-drawer-nav"><button type="button" className="account-action" disabled={selectedIndex <= 0} onClick={() => setSelectedConversation(groups[selectedIndex - 1])} aria-label="Percakapan sebelumnya"><ChevronUp size={18} /></button><button type="button" className="account-action" disabled={selectedIndex < 0 || selectedIndex >= groups.length - 1} onClick={() => setSelectedConversation(groups[selectedIndex + 1])} aria-label="Percakapan berikutnya"><ChevronDown size={18} /></button><button type="button" className="account-action" onClick={() => setSelectedConversation(null)} aria-label="Tutup"><XCircle size={18} /></button></div></header>
    <div className="tg-drawer-search"><label className="tg-search"><Search size={14} /><input className="input" type="search" value={chatQuery} onChange={event => setChatSearch({ id: selectedConversation.id, text: event.target.value })} placeholder="Cari di percakapan ini" aria-label="Cari di percakapan ini" /></label></div>
    <div className="tg-drawer-body">{messagesLoading ? <div className="telegram-message-empty">Memuat isi percakapan…</div> : visibleMessages.length ? visibleMessages.map((message, index) => { const day = dayLabel(message.created_at); const showDay = index === 0 || day !== dayLabel(visibleMessages[index - 1].created_at); const side = message.sender_type === "USER" ? "user" : "bot"; return <div key={message.id}>{showDay && <div className="tg-day"><span>{day}</span></div>}<article className={`tg-bubble ${side}`}><div className="tg-bubble-meta"><strong>{message.sender_type === "USER" ? "Pengguna" : message.sender_type === "BOT" ? "Asisten" : "Admin"}</strong><time>{new Date(message.created_at).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" })}</time></div><p>{message.message_text}</p></article></div>; }) : <div className="telegram-message-empty">{chatQuery ? "Tidak ada pesan yang cocok." : "Belum ada isi pesan untuk percakapan ini."}</div>}</div>
  </aside></div>}
  {detailTurn && <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget) setDetailTurn(null); }}><div className="modal panel panel-pad telegram-detail-modal" role="dialog" aria-modal="true" aria-labelledby="telegram-detail-title"><div className="card-heading"><div><h2 className="section-title" id="telegram-detail-title">Detail percakapan</h2><p className="section-hint">{detailTurn.conversation.username ? `@${detailTurn.conversation.username}` : "Pengguna Telegram"} · {new Date(detailTurn.created_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}</p></div><button className="account-action" type="button" onClick={() => setDetailTurn(null)} aria-label="Tutup detail"><XCircle size={18} /></button></div><div className="telegram-detail-block"><span className="field-label">Pertanyaan pengguna</span><p>{detailTurn.question}</p></div><div className="telegram-detail-block telegram-detail-answer"><span className="field-label">Balasan asisten</span><p className={detailTurn.answer === "Belum ada balasan asisten" ? "pending" : ""}>{detailTurn.answer}</p>{detailTurn.answered_at && <small>Dijawab {new Date(detailTurn.answered_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}</small>}</div><div className="modal-actions"><button className="button button-primary" type="button" onClick={() => setDetailTurn(null)}>Tutup</button></div></div></div>}
  {uploadModal && <KnowledgeUploadModal uploading={uploadingFiles} onClose={() => setUploadModal(false)} onSubmit={uploadKnowledgeFiles} />}
  </AppShell>;
}
