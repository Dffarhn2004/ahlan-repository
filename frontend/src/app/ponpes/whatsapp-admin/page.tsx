"use client";

import { useEffect, useState } from "react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { WhatsAppRouting } from "@/components/whatsapp-routing";
import { api } from "@/lib/api";
import { can, chooseTenant, loadAuth, type AuthContext } from "@/lib/auth";

type Tenant = { id: string; name: string };

export default function WhatsAppAdminPage() {
  const [auth, setAuth] = useState<AuthContext | null>(null);
  const [tenantId, setTenantId] = useState("");
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    void (async () => {
      try {
        const [authData, tenantData] = await Promise.all([loadAuth(), api<{ data: Tenant[] }>("/api/v1/tenants")]);
        setAuth(authData);
        setTenantId(chooseTenant(tenantData?.data ?? []) ?? "");
      } catch { setFailed(true); }
    })();
  }, []);

  return <AppShell mode="ponpes" title="Admin WhatsApp">
    <PageHeader eyebrow="Layanan WhatsApp" title="Admin WhatsApp" description="Atur apakah nomor WhatsApp khusus PSB atau melayani banyak keperluan, dan siapa admin yang dihubungi asisten untuk keperluan selain PSB." />
    {failed && <p className="error" role="status">Halaman belum dapat dimuat.</p>}
    {tenantId && <WhatsAppRouting tenantId={tenantId} canEdit={can(auth, "assistant:manage")} />}
  </AppShell>;
}
