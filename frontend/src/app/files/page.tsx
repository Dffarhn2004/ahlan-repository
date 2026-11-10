"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { ArrowLeft, ChevronRight, FileText, Folder, FolderPlus, Grid3X3, List, LoaderCircle, Plus, RefreshCw, Search, Trash2, UploadCloud, X, Eye } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { api } from "@/lib/api";
import { Pagination, usePagination } from "@/components/pagination";
import { activeTenantKey } from "@/lib/auth";
import { KnowledgeUploadModal } from "@/components/knowledge-upload-modal";
import { KnowledgeBanner, useKnowledgeStatus, type ReadState } from "@/components/knowledge-status";

type Tenant = { id: string; name: string };
type FolderItem = { id: string; name: string; parent_id?: string };
type FileItem = { id: string; filename: string; size_bytes: number; folder_id?: string; status?: string };
const fileStatus = (status?: string, read?: ReadState) => status === "PENDING_SYNC" ? { label: "Belum dipakai asisten", tone: "status status-muted" } : status === "FAILED" ? { label: "Gagal", tone: "status status-danger" } : read === "READING" ? { label: "Sedang dipelajari", tone: "status status-muted" } : read === "FAILED" ? { label: "Gagal dibaca", tone: "status status-danger" } : read === "UNREAD" ? { label: "Belum dipelajari", tone: "status status-muted" } : { label: "Dipakai asisten", tone: "status" };
const formatSize = (bytes: number) => bytes < 1024 ? `${bytes} B` : bytes < 1024 * 1024 ? `${Math.round(bytes / 1024)} KB` : `${(bytes / 1024 / 1024).toFixed(1)} MB`;

export default function FilesPage() {
  const [tenant, setTenant] = useState<Tenant | null>(null); const [tenantId, setTenantId] = useState(""); const [folderPath, setFolderPath] = useState<FolderItem[]>([]); const [folders, setFolders] = useState<FolderItem[]>([]); const [files, setFiles] = useState<FileItem[]>([]); const [folderId, setFolderId] = useState<string | null>(null); const [uploadOpen, setUploadOpen] = useState(false); const [folderOpen, setFolderOpen] = useState(false); const [folderName, setFolderName] = useState(""); const [query, setQuery] = useState(""); const [message, setMessage] = useState(""); const [loading, setLoading] = useState(true); const [uploading, setUploading] = useState(false); const [syncing, setSyncing] = useState(false); const { status: knowledge, finished: knowledgeFinished, refresh: refreshKnowledge } = useKnowledgeStatus(tenantId);
  async function load(id = tenantId, currentFolder = folderId) { if (!id) return; setLoading(true); try { const [fileData, folderData] = await Promise.all([api<{ data: FileItem[] }>(`/api/v1/tenants/${id}/files`), api<{ data: FolderItem[] }>(`/api/v1/tenants/${id}/folders${currentFolder ? `?parent_id=${currentFolder}` : ""}`)]); setFiles(fileData?.data ?? []); setFolders(folderData?.data ?? []); } catch { setMessage("Dokumen belum dapat dimuat."); } finally { setLoading(false); } }
  useEffect(() => { void api<{ data: Tenant[] }>("/api/v1/tenants").then(result => { const list = result?.data ?? []; const saved = localStorage.getItem(activeTenantKey); const id = list.some(item => item.id === saved) ? saved! : list[0]?.id ?? ""; setTenant(list.find(item => item.id === id) ?? null); setTenantId(id); if (id) void load(id); }).catch(() => setMessage("Pesantren belum dapat dimuat.")); }, []);
  const visibleFiles = useMemo(() => files.filter(file => (file.folder_id ?? "") === (folderId ?? "") && file.filename.toLowerCase().includes(query.toLowerCase())), [files, folderId, query]);
  const paged = usePagination(visibleFiles, 10, `${folderId ?? ""}|${query}`);
  async function upload(selected: File[]) {
    if (!selected.length || !tenantId) return;
    setUploading(true); setMessage("");
    try {
      for (const file of selected) { const body = new FormData(); body.append("file", file); if (folderId) body.append("folder_id", folderId); await api(`/api/v1/tenants/${tenantId}/files`, { method: "POST", body }); }
      setUploadOpen(false);
      // Make the new files usable right away when the assistant is already prepared.
      const synced = await syncDocuments(true);
      setMessage(synced ? `${selected.length} dokumen diunggah dan diteruskan ke asisten.` : `${selected.length} dokumen diunggah. Dokumen akan dipakai asisten setelah asisten disiapkan dan pengetahuan diperbarui.`);
      await load();
    } catch { setMessage("Sebagian dokumen gagal diunggah."); } finally { setUploading(false); }
  }
  async function syncDocuments(quiet = false) {
    if (!tenantId || syncing) return false;
    setSyncing(true);
    if (!quiet) setMessage("Memperbarui pengetahuan asisten dari semua dokumen dan folder…");
    try {
      const result = await api<{ files: number; images?: number }>(`/api/v1/tenants/${tenantId}/hermes/sync`, { method: "POST" });
      if (!quiet) { setMessage(`${result?.files ?? 0} dokumen berhasil ditambahkan ke pengetahuan asisten.`); await load(); }
      refreshKnowledge();
      return true;
    } catch { if (!quiet) setMessage("Pengetahuan asisten gagal diperbarui. Pastikan asisten sudah disiapkan."); return false; } finally { setSyncing(false); }
  }
  const replaceInput = useRef<HTMLInputElement>(null); const [replaceTarget, setReplaceTarget] = useState<FileItem | null>(null); const [busyFile, setBusyFile] = useState("");
  async function viewFile(file: FileItem) {
    // The document is fetched with the login cookie and opened from memory, so no link carries credentials.
    setBusyFile(file.id);
    try { const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/tenants/${tenantId}/files/${file.id}`, { credentials: "include" }); if (!response.ok) throw new Error("fetch"); const url = URL.createObjectURL(await response.blob()); const opened = window.open(url, "_blank", "noopener"); if (!opened) { const link = document.createElement("a"); link.href = url; link.download = file.filename; link.click(); } setTimeout(() => URL.revokeObjectURL(url), 60000); }
    catch { setMessage(`Dokumen ${file.filename} belum dapat dibuka.`); } finally { setBusyFile(""); }
  }
  async function updateFile(file: File | undefined) {
    const target = replaceTarget; setReplaceTarget(null); if (replaceInput.current) replaceInput.current.value = "";
    if (!file || !target || !tenantId) return;
    setBusyFile(target.id); setMessage("");
    try {
      const body = new FormData(); body.append("file", file);
      await api(`/api/v1/tenants/${tenantId}/files/${target.id}/content`, { method: "PUT", body });
      const synced = await syncDocuments(true);
      setMessage(synced ? `Dokumen ${target.filename} diperbarui dan diteruskan ke asisten.` : `Dokumen ${target.filename} diperbarui. Asisten memakainya setelah pengetahuan diperbarui.`);
      await load();
    } catch { setMessage(`Dokumen ${target.filename} gagal diperbarui.`); } finally { setBusyFile(""); }
  }
  async function deleteFile(file: FileItem) {
    if (!window.confirm(`Hapus dokumen ${file.filename}? Asisten tidak akan memakai isinya lagi.`)) return;
    setBusyFile(file.id); setMessage("");
    try { await api(`/api/v1/tenants/${tenantId}/files/${file.id}`, { method: "DELETE" }); await syncDocuments(true); setMessage(`Dokumen ${file.filename} dihapus.`); await load(); }
    catch { setMessage(`Dokumen ${file.filename} gagal dihapus.`); } finally { setBusyFile(""); }
  }
  async function createFolder(event: React.FormEvent) { event.preventDefault(); if (!folderName.trim()) return; try { await api(`/api/v1/tenants/${tenantId}/folders`, { method: "POST", body: JSON.stringify({ name: folderName, parent_id: folderId ?? "" }) }); setFolderName(""); setFolderOpen(false); await load(); } catch { setMessage("Folder gagal dibuat."); } }
  return <AppShell mode="ponpes" title="Dokumen"><PageHeader eyebrow="Pengetahuan asisten" title="Dokumen pesantren" description={tenant ? `Kelola dokumen pengetahuan untuk ${tenant.name}.` : "Kelola dokumen pengetahuan asisten AI."} action={<><button className="button" onClick={() => void syncDocuments()} disabled={!tenantId || syncing || !files.length}>{syncing ? <LoaderCircle size={15} className="spin" /> : <UploadCloud size={15} />}{syncing ? "Memperbarui..." : "Perbarui pengetahuan asisten"}</button><button className="button" onClick={() => setFolderOpen(true)} disabled={!tenantId}><FolderPlus size={15}/>Folder baru</button><button className="button button-primary" onClick={() => setUploadOpen(true)} disabled={!tenantId}><UploadCloud size={15}/>Unggah dokumen</button></>} />{message && <div className="notice" role="status">{message}</div>}<KnowledgeBanner status={knowledge} finished={knowledgeFinished} onRetry={() => void syncDocuments()} busy={syncing} /><section className="document-toolbar">{folderPath.length > 0 && <button className="button document-back" type="button" onClick={() => { const next = folderPath.slice(0, -1); const parent = next.at(-1)?.id ?? null; setFolderPath(next); setFolderId(parent); void load(tenantId, parent); }}><ArrowLeft size={14}/>Kembali</button>}<div className="document-breadcrumb"><button type="button" className={!folderId ? "active" : ""} onClick={() => { setFolderPath([]); setFolderId(null); void load(tenantId, null); }}>Semua dokumen</button>{folderPath.map((folder, index) => <span className="breadcrumb-segment" key={folder.id}><ChevronRight size={14}/><button type="button" className={index === folderPath.length - 1 ? "current" : ""} onClick={() => { const next = folderPath.slice(0, index + 1); setFolderPath(next); setFolderId(folder.id); void load(tenantId, folder.id); }}>{folder.name}</button></span>)}</div><div className="input-with-icon document-search"><Search size={15}/><input className="input" value={query} onChange={event => setQuery(event.target.value)} placeholder="Cari dokumen..." /></div><button className="view-toggle active" aria-label="Tampilan daftar"><List size={16}/></button><button className="view-toggle" aria-label="Tampilan grid"><Grid3X3 size={16}/></button></section><section className="document-summary"><span><strong>{visibleFiles.length}</strong> dokumen</span><span><strong>{folders.length}</strong> folder</span><span className="document-policy">Dokumen hanya dapat diakses oleh tenant ini.</span></section><section className="document-grid">{folders.map(folder => <button className="folder-card" key={folder.id} onClick={() => { setFolderPath(path => [...path, folder]); setFolderId(folder.id); void load(tenantId, folder.id); }}><Folder size={24}/><span>{folder.name}</span><small>Folder</small></button>)}{loading ? <div className="document-empty">Memuat dokumen…</div> : visibleFiles.length ? paged.rows.map(file => <div className="document-card" key={file.id}><span className="document-icon"><FileText size={21}/></span><div><strong>{file.filename}</strong><span>{formatSize(file.size_bytes)}</span></div><span className={fileStatus(file.status, knowledge?.files[file.id]).tone}>{fileStatus(file.status, knowledge?.files[file.id]).label}</span><div className="document-actions"><button className="button" type="button" disabled={busyFile === file.id} onClick={() => void viewFile(file)} title="Lihat dokumen"><Eye size={14}/>Lihat</button><button className="button" type="button" disabled={busyFile === file.id || syncing} onClick={() => { setReplaceTarget(file); replaceInput.current?.click(); }} title="Ganti isi dokumen dengan berkas baru"><RefreshCw size={14}/>Update</button><button className="button button-danger" type="button" disabled={busyFile === file.id || syncing} onClick={() => void deleteFile(file)} title="Hapus dokumen"><Trash2 size={14}/>Hapus</button></div></div>) : <div className="document-empty"><UploadCloud size={30}/><strong>Belum ada dokumen di sini</strong><span>Unggah satu atau beberapa dokumen untuk memperkaya pengetahuan asisten.</span><button className="button button-primary" onClick={() => setUploadOpen(true)}><Plus size={15}/>Unggah dokumen</button></div>}</section><input ref={replaceInput} type="file" hidden onChange={event => void updateFile(event.target.files?.[0])} /><Pagination paged={paged} label="dokumen"/>{uploadOpen && <KnowledgeUploadModal uploading={uploading} onClose={() => setUploadOpen(false)} onSubmit={upload} />}{folderOpen && <div className="modal-backdrop"><form className="modal panel panel-pad" onSubmit={createFolder}><div className="modal-heading"><div><span className="modal-kicker">Organisasi dokumen</span><h2 className="section-title">Folder baru</h2></div><button className="account-action" type="button" onClick={() => setFolderOpen(false)}><X size={18}/></button></div><label><span className="field-label">Nama folder</span><input className="input" autoFocus required value={folderName} onChange={event => setFolderName(event.target.value)} placeholder="Contoh: PPDB 2026" /></label><div className="modal-actions"><button className="button" type="button" onClick={() => setFolderOpen(false)}>Batal</button><button className="button button-primary" type="submit">Buat folder</button></div></form></div>}</AppShell>;
}
