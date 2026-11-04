const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(readonly status: number, message: string, readonly requestId?: string) {
    super(message);
  }
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T | null> {
  const headers = new Headers(init.headers);
  if (init.body && !(init.body instanceof FormData)) headers.set("Content-Type", "application/json");
  const response = await fetch(`${API}${path}`, { ...init, credentials: "include", headers });
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { code?: string; request_id?: string } | null;
    const requestId = response.headers.get("X-Request-ID") ?? body?.request_id ?? undefined;
    throw new ApiError(response.status, body?.code ?? "REQUEST_FAILED", requestId);
  }
  return response.status === 204 ? null : response.json() as Promise<T>;
}
