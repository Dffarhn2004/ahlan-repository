export const brandingKey = "ahlan.branding";
export const defaultBranding = { appName: "AHLAN", logoOnly: "/logo.png" };
export type Branding = typeof defaultBranding;

export function loadBranding(): Branding {
  if (typeof window === "undefined") return defaultBranding;
  try { return { ...defaultBranding, ...(JSON.parse(localStorage.getItem(brandingKey) ?? "{}") as Partial<Branding>) }; } catch { return defaultBranding; }
}

export function saveBranding(value: Branding) {
  localStorage.setItem(brandingKey, JSON.stringify(value));
  window.dispatchEvent(new Event("branding-change"));
}
