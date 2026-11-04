"use client";

import { useState } from "react";
import { FileText, LoaderCircle, UploadCloud, XCircle } from "lucide-react";

const maxUploadSize = 20 * 1024 * 1024;
const formatFileSize = (bytes: number) => bytes < 1024 * 1024 ? `${Math.max(1, Math.round(bytes / 1024))} KB` : `${(bytes / 1024 / 1024).toFixed(1)} MB`;

/** Drag-and-drop upload dialog shared by every knowledge-base upload screen. */
export function KnowledgeUploadModal({ title = "Unggah bahan pengetahuan", uploading, onClose, onSubmit }: { title?: string; uploading: boolean; onClose: () => void; onSubmit: (files: File[]) => void | Promise<void> }) {
  const [selected, setSelected] = useState<File[]>([]);
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState("");

  const add = (incoming: File[]) => {
    const tooLarge = incoming.filter(file => file.size > maxUploadSize);
    setError(tooLarge.length ? `${tooLarge.map(file => file.name).join(", ")} melebihi batas 20 MB dan tidak ditambahkan.` : "");
    setSelected(current => [...current, ...incoming.filter(file => file.size <= maxUploadSize && !current.some(item => item.name === file.name && item.size === file.size && item.lastModified === file.lastModified))]);
  };

  return <div className="modal-backdrop" role="presentation" onMouseDown={event => { if (event.target === event.currentTarget && !uploading) onClose(); }}>
    <form className="modal panel panel-pad" onSubmit={event => { event.preventDefault(); if (selected.length && !uploading) void onSubmit(selected); }}>
      <div className="card-heading"><div><h2 className="section-title">{title}</h2><p className="section-hint">File akan disimpan untuk pesantren ini dan dipakai asisten setelah pengetahuan diperbarui.</p></div><button type="button" className="account-action" onClick={onClose} disabled={uploading} aria-label="Tutup"><XCircle size={18} /></button></div>
      <div>
        <span className="field-label">Pilih file</span>
        <label className={`knowledge-dropzone${dragActive ? " is-active" : ""}`} onDragOver={event => { event.preventDefault(); setDragActive(true); }} onDragLeave={() => setDragActive(false)} onDrop={event => { event.preventDefault(); setDragActive(false); add(Array.from(event.dataTransfer.files)); }}>
          <input className="knowledge-file-input" type="file" multiple onChange={event => { add(Array.from(event.target.files ?? [])); event.target.value = ""; }} />
          <span className="knowledge-dropzone-icon"><UploadCloud size={22} /></span>
          <strong>Tarik file ke sini atau <u>pilih dari komputer</u></strong>
          <span>PDF, Word, Excel, PowerPoint, gambar, teks, CSV, dan lainnya · maks. 20 MB per file</span><span>Tahan Ctrl atau Shift untuk memilih beberapa file sekaligus</span>
        </label>
        {error && <div className="error" role="alert" style={{ marginTop: 10 }}>{error}</div>}
      </div>
      {selected.length > 0 && <div>
        <div className="knowledge-files-head"><span>{selected.length} file dipilih · {formatFileSize(selected.reduce((sum, file) => sum + file.size, 0))}</span><span className="knowledge-head-actions"><label className="knowledge-link">+ Tambah file lagi<input type="file" multiple style={{ display: "none" }} disabled={uploading} onChange={event => { add(Array.from(event.target.files ?? [])); event.target.value = ""; }} /></label><button type="button" className="knowledge-link" onClick={() => setSelected([])} disabled={uploading}>Hapus semua</button></span></div>
        <ul className="knowledge-files">{selected.map(file => <li key={`${file.name}-${file.size}-${file.lastModified}`}><span className="knowledge-file-icon"><FileText size={15} /></span><span className="knowledge-file-name">{file.name}</span><small>{formatFileSize(file.size)}</small><button type="button" className="knowledge-file-remove" aria-label={`Hapus ${file.name}`} disabled={uploading} onClick={() => setSelected(current => current.filter(item => item !== file))}><XCircle size={15} /></button></li>)}</ul>
      </div>}
      <div className="modal-actions"><button className="button" type="button" onClick={onClose} disabled={uploading}>Batal</button><button className="button button-primary" type="submit" disabled={!selected.length || uploading}>{uploading ? <LoaderCircle size={15} className="spin" /> : <UploadCloud size={15} />}{uploading ? "Mengunggah..." : selected.length > 1 ? `Unggah ${selected.length} file` : "Unggah file"}</button></div>
    </form>
  </div>;
}
