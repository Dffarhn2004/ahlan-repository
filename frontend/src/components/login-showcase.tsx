"use client";

import { useEffect, useState } from "react";
import { Bot, FileText, GraduationCap, LockKeyhole, type LucideIcon } from "lucide-react";

const words = ["tertata rapi", "lebih terkendali", "selalu siaga", "mudah dipantau", "tenang dijalankan"];

const features: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: Bot, title: "Asisten AI per pesantren", text: "Jawab pertanyaan wali santri dan calon santri 24 jam lewat Telegram, dengan gaya bahasa yang Anda atur sendiri." },
  { icon: FileText, title: "Dokumen & pengetahuan", text: "Unggah brosur, tata tertib, dan jadwal. Asisten hanya menjawab dari dokumen milik pesantren Anda." },
  { icon: GraduationCap, title: "Pendaftaran calon santri", text: "Pantau percakapan dan minat calon santri dalam satu kotak masuk, tanpa harus membalas satu per satu." },
  { icon: LockKeyhole, title: "Akses & jejak audit", text: "Atur peran pengurus dan lihat siapa mengubah apa. Data tiap pesantren terpisah dan aman." },
];

const WORD_MS = 2600;

export function RotatingWord() {
  const [index, setIndex] = useState(0);
  const [previous, setPrevious] = useState<number | null>(null);
  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    const timer = setInterval(() => setIndex(current => { setPrevious(current); return (current + 1) % words.length; }), WORD_MS);
    return () => clearInterval(timer);
  }, []);
  return (
    <span className="rotator" aria-live="off">
      <span className="rotator-sizer" aria-hidden="true">{words.reduce((a, b) => (a.length >= b.length ? a : b))}</span>
      {previous !== null && <span key={`out-${previous}-${index}`} className="rotator-word rotator-out" aria-hidden="true">{words[previous]}</span>}
      <span key={`in-${index}`} className="rotator-word rotator-in">{words[index]}</span>
    </span>
  );
}

export function FeatureList() {
  const [active, setActive] = useState(0);
  const [paused, setPaused] = useState(false);
  return (
    <ul className="feature-list" onMouseEnter={() => setPaused(true)} onMouseLeave={() => setPaused(false)} onFocus={() => setPaused(true)} onBlur={() => setPaused(false)}>
      {features.map(({ icon: Icon, title, text }, i) => (
        <li key={title} className={i === active ? "feature-item active" : "feature-item"}>
          <button type="button" aria-expanded={i === active} onClick={() => setActive(i)}>
            <span className="feature-icon"><Icon size={17} /></span>
            <span className="feature-body">
              <strong>{title}</strong>
              <span className="feature-text"><span>{text}</span></span>
            </span>
          </button>
          {i === active && <i className="feature-progress" key={active} style={{ animationPlayState: paused ? "paused" : "running" }} onAnimationEnd={() => setActive((active + 1) % features.length)} />}
        </li>
      ))}
    </ul>
  );
}
