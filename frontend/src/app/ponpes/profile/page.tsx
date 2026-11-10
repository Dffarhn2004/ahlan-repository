"use client";

import { useEffect, useState } from "react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { AssistantPanel } from "@/components/assistant-panel";
import { ContactPersons } from "@/components/contact-persons";
import { OfficialLinks } from "@/components/official-links";
import { api, ApiError } from "@/lib/api";
import { can, chooseTenant, loadAuth, type AuthContext } from "@/lib/auth";
import { validateMapsUrl, validateOfficialLink, type LinkKind } from "@/lib/links";

type Tenant = { id: string; name: string; slug?: string };
type Profile = { display_name?: string; description?: string; address?: string; phone?: string; email?: string; website?: string; google_maps_url?: string; instagram_url?: string; facebook_url?: string; tiktok_url?: string; youtube_url?: string; has_logo?: boolean };

type ProfileTab = "assistant" | "info" | "links" | "contacts";
const profileTabs: { value: ProfileTab; label: string }[] = [
  { value: "assistant", label: "Asisten AI" },
  { value: "info", label: "Informasi pesantren" },
  { value: "links", label: "Tautan resmi" },
  { value: "contacts", label: "Kontak person" },
];

const socialFields: { key: "instagram_url" | "facebook_url" | "tiktok_url" | "youtube_url"; kind: LinkKind; label: string; placeholder: string }[] = [
  { key: "instagram_url", kind: "INSTAGRAM", label: "Instagram", placeholder: "https://www.instagram.com/…" },
  { key: "facebook_url", kind: "FACEBOOK", label: "Facebook", placeholder: "https://www.facebook.com/…" },
  { key: "tiktok_url", kind: "TIKTOK", label: "TikTok", placeholder: "https://www.tiktok.com/@…" },
  { key: "youtube_url", kind: "YOUTUBE", label: "YouTube", placeholder: "https://www.youtube.com/@…" },
];

export default function ProfilePage() {
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [tenantId, setTenantId] = useState("");
  const [profile, setProfile] = useState<Profile>({});
  const [notice, setNotice] = useState("");
  const [saving, setSaving] = useState(false);
  const [slug, setSlug] = useState("");
  const [logoBusy, setLogoBusy] = useState(false);
  const [logoNotice, setLogoNotice] = useState("");
  const [logoVersion, setLogoVersion] = useState(0);
  const [tab, setTab] = useState<ProfileTab>("assistant");

  useEffect(() => {
    void (async () => {
      try {
        const [authData, tenantData] = await Promise.all([loadAuth(), api<{ data: Tenant[] }>("/api/v1/tenants")]);
        setAuth(authData);
        const list = tenantData?.data ?? [];
        const id = chooseTenant(list);
        if (!id) return;
        setTenantId(id);
        setSlug(list.find(item => item.id === id)?.slug ?? "");
        const fallback = list.find(item => item.id === id)?.name ?? "";
        try { const stored = await api<Profile>(`/api/v1/tenants/${id}/profile`); setProfile({ ...stored, display_name: stored?.display_name || fallback }); }
        catch { setProfile({ display_name: fallback }); setNotice("Profil belum dibuat. Lengkapi data lalu simpan."); }
      } catch { setNotice("Profil belum dapat dimuat."); }
    })();
  }, []);

  useEffect(() => {
    const selected = new URLSearchParams(window.location.search).get("tab") as ProfileTab | null;
    if (selected && profileTabs.some(item => item.value === selected)) setTab(selected);
  }, []);
  const selectTab = (value: ProfileTab) => { setTab(value); window.history.replaceState({}, "", `?tab=${value}`); };

  const mapsError = validateMapsUrl(profile.google_maps_url ?? "");
  const socialErrors = Object.fromEntries(socialFields.map(field => [field.key, validateOfficialLink(field.kind, profile[field.key] ?? "")])) as Record<string, string>;
  const socialError = Object.values(socialErrors).find(Boolean) ?? "";
  const set = (key: keyof Profile) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => setProfile(current => ({ ...current, [key]: event.target.value }));
  const canEdit = can(auth, "tenant:update");

  async function uploadLogo(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) { setLogoNotice("Ukuran logo maksimal 2 MB."); return; }
    setLogoBusy(true); setLogoNotice("");
    try {
      const form = new FormData(); form.append("file", file);
      await api(`/api/v1/tenants/${tenantId}/logo`, { method: "PUT", body: form });
      setProfile(current => ({ ...current, has_logo: true })); setLogoVersion(Date.now());
      setLogoNotice("Logo disimpan.");
    } catch (error) {
      const code = error instanceof ApiError ? error.message : "";
      setLogoNotice(code === "LOGO_TYPE" ? "Format logo harus PNG, JPG, atau WebP." : code === "PROFILE_REQUIRED" ? "Simpan informasi pesantren terlebih dahulu, lalu unggah logo." : "Logo gagal diunggah.");
    } finally { setLogoBusy(false); }
  }

  async function removeLogo() {
    if (!window.confirm("Hapus logo? Halaman publik akan memakai logo bawaan.")) return;
    setLogoBusy(true); setLogoNotice("");
    try { await api(`/api/v1/tenants/${tenantId}/logo`, { method: "DELETE" }); setProfile(current => ({ ...current, has_logo: false })); setLogoNotice("Logo dihapus."); }
    catch { setLogoNotice("Logo gagal dihapus."); }
    finally { setLogoBusy(false); }
  }

  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (mapsError) { setNotice(mapsError); return; }
    if (socialError) { setNotice(socialError); return; }
    setSaving(true);
    try {
      await api(`/api/v1/tenants/${tenantId}/profile`, { method: "PUT", body: JSON.stringify(profile) });
      setNotice("Profil disimpan dan diteruskan ke asisten. Jika asisten sedang aktif dan jawaban belum berubah, pilih Mulai ulang.");
    } catch { setNotice("Profil gagal disimpan."); }
    finally { setSaving(false); }
  }

  return <AppShell mode="ponpes" title="Asisten AI">
    <PageHeader eyebrow="Asisten AI" title="Asisten & informasi pesantren" description="Aktifkan asisten untuk calon santri dan atur informasi resmi yang menjadi rujukan jawabannya." />
    <nav className="hermes-subtabs" aria-label="Bagian profil">
      {profileTabs.map(item => <button key={item.value} className={tab === item.value ? "active" : ""} type="button" aria-selected={tab === item.value} onClick={() => selectTab(item.value)}>{item.label}</button>)}
    </nav>
    {notice && tab !== "info" && <p className={notice.includes("disimpan") ? "status" : "error"} role="status" style={{ marginBottom: 16 }}>{notice}</p>}
    <div style={{ display: "grid", gap: 20 }}>
      {tab === "assistant" && tenantId && <AssistantPanel tenantId={tenantId} canManage={can(auth, "assistant:manage")} canDelete={auth?.role === "SUPERADMIN"} />}
      {tab === "info" && <form className="panel panel-pad" onSubmit={save}>
        <div className="card-heading"><div><h2 className="section-title">Informasi resmi pesantren</h2><p className="section-hint">Asisten memakai informasi ini saat menjawab. Hindari memasukkan rahasia atau data pribadi.</p></div></div>
        <div style={{ display: "flex", gap: 16, alignItems: "center", flexWrap: "wrap", marginBottom: 20 }}>
          <div style={{ width: 80, height: 80, borderRadius: 16, border: "1px solid var(--line)", display: "grid", placeItems: "center", overflow: "hidden", background: "#fff" }}>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            {profile.has_logo && slug ? <img src={`/api/v1/public/p/${slug}/logo?t=${logoVersion}`} alt="Logo pesantren" style={{ width: "100%", height: "100%", objectFit: "contain" }} /> : <span className="field-help">Belum ada</span>}
          </div>
          <div>
            <span className="field-label">Logo pesantren</span>
            <span className="field-help" style={{ display: "block", marginBottom: 8 }}>Tampil di halaman publik. PNG, JPG, atau WebP, maksimal 2 MB.</span>
            {canEdit && <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              <label className="button" style={{ cursor: logoBusy ? "wait" : "pointer" }}>{logoBusy ? "Memproses…" : profile.has_logo ? "Ganti logo" : "Unggah logo"}<input type="file" accept="image/png,image/jpeg,image/webp" hidden disabled={logoBusy || !tenantId} onChange={event => void uploadLogo(event)} /></label>
              {profile.has_logo && <button className="button" type="button" disabled={logoBusy} onClick={() => void removeLogo()}>Hapus logo</button>}
            </div>}
            {logoNotice && <p className={/disimpan|dihapus/.test(logoNotice) ? "status" : "error"} role="status" style={{ marginTop: 8 }}>{logoNotice}</p>}
          </div>
        </div>
        <div className="form-row">
          <label><span className="field-label">Nama pesantren</span><input required className="input" disabled={!canEdit} value={profile.display_name ?? ""} onChange={set("display_name")} placeholder="Nama resmi pesantren" /></label>
          <label><span className="field-label">Telepon / WhatsApp</span><input className="input" disabled={!canEdit} value={profile.phone ?? ""} onChange={set("phone")} placeholder="+62…" /></label>
        </div>
        <div className="form-row" style={{ marginTop: 16 }}>
          <label><span className="field-label">Email resmi</span><input className="input" type="email" disabled={!canEdit} value={profile.email ?? ""} onChange={set("email")} placeholder="admin@pesantren.id" /></label>
          <label><span className="field-label">Website</span><input className="input" disabled={!canEdit} value={profile.website ?? ""} onChange={set("website")} placeholder="https://" /></label>
        </div>
        <label style={{ display: "block", marginTop: 16 }}><span className="field-label">Tentang pesantren</span><textarea className="input" rows={3} disabled={!canEdit} value={profile.description ?? ""} onChange={set("description")} placeholder="Profil singkat, visi, program unggulan" /></label>
        <label style={{ display: "block", marginTop: 16 }}><span className="field-label">Alamat dan informasi layanan</span><textarea className="input" rows={3} disabled={!canEdit} value={profile.address ?? ""} onChange={set("address")} placeholder="Alamat, jam layanan, atau keterangan penting" /><span className="field-help">Informasi yang mudah dipahami membantu calon santri memperoleh jawaban yang konsisten.</span></label>
        <label style={{ display: "block", marginTop: 16 }}><span className="field-label">Link Google Maps</span><input className="input" type="url" disabled={!canEdit} value={profile.google_maps_url ?? ""} onChange={set("google_maps_url")} placeholder="https://maps.app.goo.gl/…" aria-invalid={Boolean(mapsError)} />
          {mapsError ? <span className="error">{mapsError}</span> : <span className="field-help">Asisten membagikan tautan ini saat calon santri menanyakan lokasi. {profile.google_maps_url && <a className="text-link" href={profile.google_maps_url} target="_blank" rel="noopener noreferrer">Buka di Maps</a>}</span>}
        </label>
        <div style={{ marginTop: 24 }}>
          <h3 className="section-title" style={{ fontSize: 15 }}>Media sosial</h3>
          <p className="section-hint">Tampil sebagai ikon di halaman link bagikan untuk calon santri. Kosongkan jika tidak ada.</p>
          <div className="form-row" style={{ marginTop: 12 }}>
            {socialFields.slice(0, 2).map(field => <label key={field.key}><span className="field-label">{field.label}</span><input className="input" type="url" disabled={!canEdit} value={profile[field.key] ?? ""} onChange={set(field.key)} placeholder={field.placeholder} aria-invalid={Boolean(socialErrors[field.key])} />{socialErrors[field.key] && <span className="error">{socialErrors[field.key]}</span>}</label>)}
          </div>
          <div className="form-row" style={{ marginTop: 16 }}>
            {socialFields.slice(2).map(field => <label key={field.key}><span className="field-label">{field.label}</span><input className="input" type="url" disabled={!canEdit} value={profile[field.key] ?? ""} onChange={set(field.key)} placeholder={field.placeholder} aria-invalid={Boolean(socialErrors[field.key])} />{socialErrors[field.key] && <span className="error">{socialErrors[field.key]}</span>}</label>)}
          </div>
        </div>
        {notice && <p className={notice.includes("disimpan") ? "status" : "error"} style={{ marginTop: 16 }} role="status">{notice}</p>}
        {canEdit && <div style={{ marginTop: 18 }}><button className="button button-primary" disabled={!tenantId || saving}>{saving ? "Menyimpan…" : "Simpan perubahan"}</button></div>}
      </form>}
      {tab === "links" && tenantId && <OfficialLinks tenantId={tenantId} canEdit={canEdit} />}
      {tab === "contacts" && tenantId && <ContactPersons tenantId={tenantId} canEdit={canEdit} />}
    </div>
  </AppShell>;
}
