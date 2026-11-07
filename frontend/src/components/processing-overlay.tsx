"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";

/** Blocking popup for slow work, so people understand a long wait is normal and do not close or retry. */
export function ProcessingOverlay({ title, steps, note }: { title: string; steps: string[]; note?: string }) {
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
      <p className="processing-note">{note ?? "Proses ini membutuhkan waktu sekitar 15–60 detik karena asisten disiapkan khusus untuk pesantren Anda. Mohon jangan menutup atau memuat ulang halaman."}</p>
    </div>
  </div>;
}
