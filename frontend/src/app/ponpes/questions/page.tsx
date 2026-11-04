"use client";

import { useEffect, useState } from "react";
import { AppShell, PageHeader } from "@/components/app-shell";
import { UnansweredQuestions } from "@/components/unanswered-questions";
import { api } from "@/lib/api";
import { can, chooseTenant, loadAuth, type AuthContext } from "@/lib/auth";

type Tenant = { id: string; name: string };

export default function QuestionsPage() {
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

  return <AppShell mode="ponpes" title="Pertanyaan">
    <PageHeader eyebrow="Evaluasi asisten" title="Pertanyaan belum terjawab" description="Lihat apa yang ditanyakan pengunjung tetapi belum bisa dijawab asisten, lalu lengkapi jawabannya." />
    {failed && <p className="error" role="status">Halaman belum dapat dimuat.</p>}
    {tenantId && <UnansweredQuestions tenantId={tenantId} canEdit={can(auth, "tenant:update")} />}
  </AppShell>;
}
