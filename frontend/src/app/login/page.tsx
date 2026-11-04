"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, ArrowRight, Eye, EyeOff, ShieldCheck, Sparkles } from "lucide-react";
import { loadBranding } from "@/lib/branding";
import { FeatureList, RotatingWord } from "@/components/login-showcase";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
async function request(path: string, init?: RequestInit) { const response = await fetch(`${API}${path}`, { ...init, credentials: "include", headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) } }); if (!response.ok) throw new Error(); return response.status === 204 ? null : response.json(); }
function destination(auth: { role?: string; memberships?: unknown[] }) { return auth.role === "SUPERADMIN" ? "/admin/dashboard" : auth.memberships?.length ? "/ponpes/dashboard" : "/admin/dashboard"; }

export default function Home() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [checking, setChecking] = useState(true);
  const [error, setError] = useState("");
  const [branding, setBranding] = useState(loadBranding);
  useEffect(() => { void request("/api/v1/me").then(auth => router.replace(destination(auth))).catch(() => setChecking(false)); const update = () => setBranding(loadBranding()); window.addEventListener("branding-change", update); return () => window.removeEventListener("branding-change", update); }, [router]);
  async function login(event: FormEvent) { event.preventDefault(); setError(""); try { await request("/api/v1/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }); const auth = await request("/api/v1/me"); router.replace(destination(auth)); } catch { setError("Email atau kata sandi tidak dapat diverifikasi."); } }
  if (checking) return <main className="login-page login-checking"><div className="loading-orb"><Sparkles size={22}/></div><span>Memeriksa sesi Anda…</span></main>;
  return <main className="login-page"><section className="login-intro"><div className="login-brand"><div className="login-brand-lockup"><img className="login-brand-icon" src={branding.logoOnly} alt="" /><strong>{branding.appName}</strong></div></div><div className="login-message"><p className="login-kicker">Platform pesantren</p><h1>Operasional pesantren,<RotatingWord/></h1><p className="login-lead">Satu ruang kerja untuk layanan AI, dokumen, dan calon santri.</p><FeatureList/></div><div className="login-assurance"><ShieldCheck size={17}/><span>Akses aman untuk administrator platform</span></div></section><section className="login-form-pane"><div className="login-form-wrap"><Link href="/" className="login-back"><ArrowLeft size={15}/>Kembali ke beranda</Link><div className="mobile-login-brand"><div className="login-brand-lockup"><img className="login-brand-icon" src={branding.logoOnly} alt="" /><strong>{branding.appName}</strong></div></div><div className="login-heading"><p className="login-kicker">Selamat datang kembali</p><h2>Masuk ke akun Anda</h2><p>Gunakan akun administrator yang telah terdaftar.</p></div><form onSubmit={login} className="login-form"><label><span className="field-label">Alamat email</span><input className="input login-input" value={email} onChange={event => setEmail(event.target.value)} autoComplete="email" inputMode="email"/></label><label><span className="field-label">Kata sandi</span><div className="password-field"><input className="input login-input" type={showPassword ? "text" : "password"} value={password} onChange={event => setPassword(event.target.value)} autoComplete="current-password"/><button type="button" onClick={() => setShowPassword(!showPassword)} aria-label={showPassword ? "Sembunyikan kata sandi" : "Tampilkan kata sandi"}>{showPassword ? <EyeOff size={17}/> : <Eye size={17}/>}</button></div></label>{error && <p className="error" role="alert">{error}</p>}<button className="button button-primary login-submit" type="submit">Masuk ke dashboard <ArrowRight size={16}/></button></form><p className="login-support">Butuh bantuan akses? Hubungi ahlan.platformai@gmail.com.</p></div></section></main>;
}
