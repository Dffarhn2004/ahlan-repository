"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bot, Building2, FileText, Globe, LayoutDashboard, LogOut, Menu, MessageCircleQuestion, MessageSquare, PanelLeftClose, PanelLeftOpen, ScrollText, Settings, Sparkles, Users, X } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { loadAuth, type AuthContext } from "@/lib/auth";
import { loadBranding } from "@/lib/branding";

type ShellProps = { children: React.ReactNode; mode?: "platform" | "ponpes"; title?: string; active?: string };
const platform = [{ label: "Overview", href: "/admin/dashboard", icon: LayoutDashboard }, { label: "Pesantren", href: "/admin", icon: Building2 }, { label: "Profil Asisten", href: "/admin/hermes", icon: Bot }, { label: "Audit log", href: "/admin/audit", icon: ScrollText }, { label: "Pengaturan", href: "/settings", icon: Settings }];
const ponpes = [{ label: "Dashboard", href: "/ponpes/dashboard", icon: LayoutDashboard }, { label: "Asisten AI", href: "/ponpes/profile", icon: Sparkles }, { label: "Dokumen", href: "/ponpes/files", icon: FileText }, { label: "Sumber Web", href: "/ponpes/sources", icon: Globe }, { label: "Percakapan", href: "/ponpes/conversations", icon: MessageSquare }, { label: "Pertanyaan", href: "/ponpes/questions", icon: MessageCircleQuestion }, { label: "Pengguna", href: "/ponpes/users", icon: Users }, { label: "Aktivitas sistem", href: "/audit", icon: ScrollText }];

export function AppShell({ children, mode = "platform", title, active }: ShellProps) {
  const router = useRouter();
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [sessionState, setSessionState] = useState<"checking" | "ready" | "error">("checking");
  const [branding, setBranding] = useState(loadBranding);
  const [accountOpen, setAccountOpen] = useState(false);
  const accountMenuRef = useRef<HTMLDivElement>(null);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  async function checkSession() {
    setSessionState("checking");
    try {
      const current = await loadAuth();
      if (!current) throw new Error("SESSION_EMPTY");
      setAuth(current);
      setSessionState("ready");
    } catch (error) {
      setAuth(null);
      if (error instanceof ApiError && error.status === 401) {
        router.replace("/login");
        return;
      }
      setSessionState("error");
    }
  }
  useEffect(() => { setSidebarCollapsed(window.localStorage.getItem("sidebar-collapsed") === "1"); void checkSession(); const update = () => setBranding(loadBranding()); window.addEventListener("branding-change", update); return () => window.removeEventListener("branding-change", update); }, [router]);
  useEffect(() => {
    if (!accountOpen) return;
    const close = (event: MouseEvent) => { if (!accountMenuRef.current?.contains(event.target as Node)) setAccountOpen(false); };
    document.addEventListener("mousedown", close);
    return () => document.removeEventListener("mousedown", close);
  }, [accountOpen]);
  function toggleSidebar() { setSidebarCollapsed(value => { const next = !value; window.localStorage.setItem("sidebar-collapsed", next ? "1" : "0"); return next; }); }
  const platformMode = mode === "platform";
  const items = platformMode ? platform : ponpes;
  const current = active ?? ({ Overview: "/admin/dashboard", Pesantren: "/admin", "Profil Asisten": "/admin/hermes", Aktivitas: mode === "platform" ? "/admin/audit" : "/audit", "Aktivitas sistem": "/audit", Pengaturan: "/settings", Dashboard: "/ponpes/dashboard", "Asisten AI": "/ponpes/profile", Dokumen: "/ponpes/files", "Sumber Web": "/ponpes/sources", Percakapan: "/ponpes/conversations", Pertanyaan: "/ponpes/questions", Pengguna: "/ponpes/users" }[title ?? ""]);
  const accountName = platformMode ? (auth?.role === "SUPERADMIN" ? "Superadmin" : "Administrator") : (auth?.user.email ?? "Admin Pesantren");
  async function logout() { try { await api("/api/v1/auth/logout", { method: "POST" }); } finally { router.push("/login"); } }

  if (sessionState === "checking") return <main className="session-state"><div className="loading-orb"><Sparkles size={22}/></div><strong>Memeriksa sesi</strong><span>Menyiapkan ruang kerja Anda…</span></main>;
  if (sessionState === "error") return <main className="session-state"><div className="loading-orb session-state-danger"><X size={22}/></div><strong>Sesi belum dapat diperiksa</strong><span>Koneksi ke layanan terputus. Periksa koneksi internet Anda, lalu coba lagi.</span><button className="button button-primary" type="button" onClick={() => void checkSession()}>Coba lagi</button></main>;

  return <div className={`app-shell ${sidebarCollapsed ? "sidebar-collapsed" : ""}`}><aside className={`sidebar ${mobileNavOpen ? "mobile-open" : ""}`}><div className="sidebar-head"><Link className="brand" href={platformMode ? "/admin/dashboard" : "/ponpes/dashboard"} onClick={() => setMobileNavOpen(false)}><span className="brand-lockup"><img className="brand-logo" src={branding.logoOnly} alt="" /><strong>{branding.appName}</strong></span></Link><button className="sidebar-collapse" type="button" onClick={toggleSidebar} aria-label={sidebarCollapsed ? "Tampilkan menu" : "Sembunyikan menu"} title={sidebarCollapsed ? "Tampilkan menu" : "Sembunyikan menu"}>{sidebarCollapsed ? <PanelLeftOpen size={15} /> : <PanelLeftClose size={15} />}</button><button className="mobile-close" type="button" onClick={() => setMobileNavOpen(false)} aria-label="Tutup navigasi"><X size={20}/></button></div><div className="workspace-label"><span></span>{platformMode ? "Platform" : "Pesantren"}</div><nav className="nav-list">{items.map(({ label, href, icon: Icon }) => <Link key={href} href={href} title={label} aria-current={current === href ? "page" : undefined} className={`nav-link ${current === href ? "active" : ""}`} onClick={() => setMobileNavOpen(false)}><Icon size={17} strokeWidth={1.8} /><span>{label}</span></Link>)}</nav></aside><div className={`mobile-overlay ${mobileNavOpen ? "visible" : ""}`} onClick={() => setMobileNavOpen(false)} aria-hidden="true"/><div className="workspace"><header className="topbar"><button className="mobile-menu" type="button" onClick={() => setMobileNavOpen(true)} aria-label="Buka navigasi"><Menu size={20}/></button><div className="breadcrumb"><span>{branding.appName}</span><i>/</i><strong>{title ?? (platformMode ? "Platform" : "Pesantren")}</strong></div><div className="topbar-meta"><span className="connection"><i />Sistem normal</span><div ref={accountMenuRef} className={`account-menu ${accountOpen ? "open" : ""}`}><button type="button" className="account-menu-trigger" onClick={() => setAccountOpen(value => !value)} aria-expanded={accountOpen} aria-label="Buka menu akun"><span className="avatar avatar-small avatar-photo"><img src="/avatar-ustadz.svg" alt="" /></span></button><div className="account-menu-panel"><div className="account-menu-heading"><span className="avatar avatar-photo"><img src="/avatar-ustadz.svg" alt="" /></span><div><strong>{accountName}</strong><span>{platformMode ? "Pemilik platform" : "Admin pesantren"}</span></div></div><Link href="/settings" className="account-menu-item" onClick={() => setAccountOpen(false)}><Settings size={15} />Pengaturan</Link><button type="button" className="account-menu-item account-menu-logout" onClick={() => void logout()}><LogOut size={15} />Keluar</button></div></div></div></header><main className="content">{children}</main></div></div>;
}

export function PageHeader({ eyebrow, title, description, action }: { eyebrow?: string; title: string; description?: string; action?: React.ReactNode }) { return <div className="page-header"><div>{eyebrow && <p className="page-eyebrow">{eyebrow}</p>}<h1 className="page-title">{title}</h1>{description && <p className="page-subtitle">{description}</p>}</div>{action && <div className="page-action">{action}</div>}</div>; }
