import { api } from "./api";

export type Permission = string;
export type Membership = { tenant_id: string; tenant_name: string; role: string };
export type AuthContext = { user: { id: string; email: string }; role: string; permissions: Permission[]; memberships: Membership[] };
export function can(auth: AuthContext | null, permission: Permission) { return Boolean(auth?.permissions.includes(permission)); }
export function loadAuth() { return api<AuthContext>("/api/v1/me"); }
export const activeTenantKey = "pesantren-ai.active-tenant";
export function chooseTenant<T extends { id: string }>(tenants: T[]) {
  if (typeof window === "undefined") return tenants[0]?.id ?? "";
  const saved = localStorage.getItem(activeTenantKey);
  const id = tenants.some(tenant => tenant.id === saved) ? saved! : tenants[0]?.id ?? "";
  if (id) localStorage.setItem(activeTenantKey, id);
  return id;
}
