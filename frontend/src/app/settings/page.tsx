"use client";

import { useEffect, useState } from "react";
import { Image, KeyRound, Mail, Save, ShieldCheck, Upload, UserRound } from "lucide-react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { loadAuth, type AuthContext } from "@/lib/auth";
import { defaultBranding, loadBranding, saveBranding } from "@/lib/branding";

const profileKey = "pesantren-ai.admin-profile";

export default function SettingsPage() {
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [name, setName] = useState("Administrator");
  const [email, setEmail] = useState("");
  const [notice, setNotice] = useState("");
  const [branding, setBranding] = useState(loadBranding);

  useEffect(() => {
    void (async () => {
      await Promise.resolve();
      const saved = localStorage.getItem(profileKey);
      setBranding(loadBranding());
      if (saved) { try { const profile = JSON.parse(saved) as { name?: string; email?: string }; setName(profile.name ?? "Administrator"); setEmail(profile.email ?? ""); } catch { /* ignore invalid local profile */ } }
      const data = await loadAuth().catch(() => null);
      if (data) { setAuth(data); setEmail(current => current || data.user.email); }
    })();
  }, []);

  function saveProfile(event: React.FormEvent) { event.preventDefault(); localStorage.setItem(profileKey, JSON.stringify({ name: name.trim() || "Administrator", email: email.trim() })); setNotice("Profil berhasil disimpan di perangkat ini."); }
  function uploadLogo(key: "logoOnly", file?: File) { if (!file) return; const reader = new FileReader(); reader.onload = () => { const next = { ...branding, [key]: String(reader.result) }; setBranding(next); saveBranding(next); setNotice("Logo aplikasi berhasil diperbarui."); }; reader.readAsDataURL(file); }
  function updateAppName(value: string) { const next = { ...branding, appName: value }; setBranding(next); saveBranding(next); }
  function resetBranding() { setBranding(defaultBranding); saveBranding(defaultBranding); setNotice("Logo aplikasi dikembalikan ke default."); }

  return <AppShell title="Pengaturan"><PageHeader eyebrow="Akun administrator" title="Pengaturan profil" description="Kelola identitas administrator dan branding aplikasi web."/><div className="settings-layout"><aside className="panel settings-nav"><div className="settings-nav-item active"><UserRound size={17}/><div><strong>Profil</strong><span>Informasi akun</span></div></div><div className="settings-nav-item"><ShieldCheck size={17}/><div><strong>Keamanan</strong><span>Akses dan sesi</span></div></div></aside><section className="panel panel-pad settings-profile"><div className="profile-heading"><div className="profile-avatar"><UserRound size={25}/></div><div><h2 className="section-title">Profil admin</h2><p className="section-hint">Informasi ini digunakan untuk identitas di control plane.</p></div></div><form onSubmit={saveProfile} className="settings-form"><label><span className="field-label">Nama tampilan</span><div className="input-with-icon"><UserRound size={15}/><input className="input" value={name} onChange={event => setName(event.target.value)} required /></div></label><label><span className="field-label">Alamat email</span><div className="input-with-icon"><Mail size={15}/><input className="input" type="email" value={email} onChange={event => setEmail(event.target.value)} required /><span className="field-help">Email login: {auth?.user.email ?? "memuat..."}</span></div></label><div className="settings-divider"/><div className="branding-section"><div><h2 className="section-title">Branding aplikasi</h2><p className="section-hint">Logo ini berlaku untuk seluruh aplikasi web, bukan per tenant.</p></div><div className="branding-grid"><BrandingUpload label="Logo aplikasi (ikon)" src={branding.logoOnly} onChange={file => uploadLogo("logoOnly", file)} /></div><label><span className="field-label">Nama aplikasi</span><input className="input" value={branding.appName} onChange={event => updateAppName(event.target.value)} placeholder="AHLAN" /></label><button className="button" type="button" onClick={resetBranding}>Kembalikan default</button></div><div className="security-note"><KeyRound size={17}/><div><strong>Keamanan akun</strong><span>Password dikelola oleh layanan autentikasi. Gunakan menu keluar setelah selesai menggunakan perangkat bersama.</span></div></div><button className="button button-primary" type="submit"><Save size={15}/>Simpan profil</button>{notice && <p className="notice" role="status">{notice}</p>}</form></section></div></AppShell>;
}

function BrandingUpload({ label, src, onChange }: { label: string; src: string; onChange: (file?: File) => void }) { return <div className="branding-upload"><div className="branding-preview">{src ? <img src={src} alt={label}/> : <Image size={24}/>}</div><div><strong>{label}</strong><label className="button button-ghost"><Upload size={14}/>Unggah<input className="hidden" type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml" onChange={event => { onChange(event.target.files?.[0]); event.currentTarget.value = ""; }}/></label></div></div>; }
