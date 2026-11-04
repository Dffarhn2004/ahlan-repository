"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AlertTriangle, CheckCircle2, LoaderCircle, Sparkles } from "lucide-react";
import { api } from "@/lib/api";

export type ReadState = "READY" | "READING" | "UNREAD" | "FAILED";
export type KnowledgeStatus = { prepared: boolean; reading: boolean; total: number; ready: number; failed: number; files: Record<string, ReadState> };

/** Tracks how far the assistant is in reading the tenant's documents; polls while reading is in progress. */
export function useKnowledgeStatus(tenantId: string) {
  const [status, setStatus] = useState<KnowledgeStatus | null>(null);
  const [finished, setFinished] = useState(false);
  const [tick, setTick] = useState(0);
  const wasReading = useRef(false);
  const refresh = useCallback(() => setTick(value => value + 1), []);
  useEffect(() => {
    if (!tenantId) return;
    let cancelled = false;
    let timer: number | undefined;
    const poll = async () => {
      const next = await api<KnowledgeStatus>(`/api/v1/tenants/${tenantId}/knowledge/status`).catch(() => null);
      if (cancelled) return;
      if (next) {
        setStatus(next);
        if (wasReading.current && !next.reading && next.total > 0 && next.ready === next.total) setFinished(true);
        wasReading.current = next.reading;
      }
      if (next?.reading) timer = window.setTimeout(() => void poll(), 3000);
    };
    void poll();
    return () => { cancelled = true; window.clearTimeout(timer); };
  }, [tenantId, tick]);
  // The "all done" confirmation only needs to stay up for a few seconds.
  useEffect(() => {
    if (!finished) return;
    const timer = window.setTimeout(() => setFinished(false), 8000);
    return () => window.clearTimeout(timer);
  }, [finished]);
  return { status, finished, refresh };
}

/** One document's reading state, shown beside the file name. Renders nothing for files not yet handed to the assistant. */
export function ReadBadge({ state }: { state?: ReadState }) {
  if (!state) return null;
  if (state === "READY") return <span className="status">Siap dipakai</span>;
  if (state === "READING") return <span className="status status-muted"><LoaderCircle size={11} className="spin" />Sedang dipelajari</span>;
  if (state === "FAILED") return <span className="status status-danger">Gagal dibaca</span>;
  return <span className="status status-muted">Belum dipelajari</span>;
}

/** Tells the pesantren when the assistant is still learning, so a test question is not mistaken for a fault. */
export function KnowledgeBanner({ status, finished = false, onRetry, busy = false }: { status: KnowledgeStatus | null; finished?: boolean; onRetry?: () => void; busy?: boolean }) {
  if (!status?.prepared || status.total === 0) return null;
  const percent = Math.round((status.ready / status.total) * 100);
  if (status.reading) {
    const left = status.total - status.ready;
    return <div className="kp" role="status" aria-live="polite">
      <div className="kp-head">
        <span className="kp-icon"><Sparkles size={16} /></span>
        <div className="kp-title"><strong>Asisten sedang mempelajari dokumen</strong><span>{left} dokumen lagi · jawaban bisa belum lengkap</span></div>
        <span className="kp-count">{status.ready}/{status.total} · {percent}%</span>
      </div>
      <div className="kp-track" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={percent} aria-label="Kemajuan membaca dokumen"><div className="kp-fill live" style={{ width: `${Math.max(percent, 5)}%` }} /></div>
      <p className="kp-hint">Gambar dan file kantor dibaca bergantian. Anda boleh meninggalkan halaman ini, proses tetap berjalan.</p>
    </div>;
  }
  const unread = status.total - status.ready;
  if (unread > 0) {
    return <div className="kp warn" role="status">
      <div className="kp-head">
        <span className="kp-icon"><AlertTriangle size={16} /></span>
        <div className="kp-title"><strong>{status.failed > 0 ? `${status.failed} dokumen gagal dibaca` : `${unread} dokumen belum dipelajari`}</strong><span>Asisten belum dapat menjawab dari dokumen tersebut.</span></div>
        {onRetry && <button className="button" type="button" onClick={onRetry} disabled={busy}>{busy ? <LoaderCircle size={14} className="spin" /> : null}{busy ? "Memproses..." : "Coba baca lagi"}</button>}
      </div>
    </div>;
  }
  if (finished) {
    return <div className="kp done" role="status">
      <div className="kp-head">
        <span className="kp-icon"><CheckCircle2 size={16} /></span>
        <div className="kp-title"><strong>Selesai dipelajari</strong><span>Semua {status.total} dokumen siap dipakai asisten.</span></div>
      </div>
    </div>;
  }
  return null;
}
