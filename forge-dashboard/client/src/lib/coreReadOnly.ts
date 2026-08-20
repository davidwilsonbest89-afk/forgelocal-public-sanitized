/**
 * Atelier de contrôle local — client lecture seule. This module deliberately
 * retains the Bearer token in a closure only: no localStorage, URL, logging
 * or analytics persistence is permitted. The caller must provide a verified
 * loopback Core base URL; this module never discovers or redirects remotely.
 */
export type CoreReadOnlyProfile = {
  id: string;
  name: string;
  runtime_id: string;
  group?: string;
  tags: string[];
  created_at: string;
  last_used: string;
  proxy_configured: boolean;
};

export type CoreReadOnlyGroup = {
  id: string;
  name: string;
  proxy_mode: string;
  proxy_configured: boolean;
  profile_count: number;
  created_at: string;
  updated_at: string;
};

export type CoreReadOnlyRuntime = {
  id: string;
  display_name: string;
  version?: string;
  architecture?: string;
  status?: string;
  enabled: boolean;
  platform_supported: boolean;
  candidate: boolean;
  launchable: boolean;
};

export type CoreReadOnlyPage<T> = {
  api_version: "v1";
  data: T[];
  page: { limit: number; next_cursor?: string };
};

export type CoreReadOnlySummary = {
  api_version: "v1";
  data: { profiles: number; groups: number; runtimes: number };
};

export type CoreReadOnlySession = {
  expiresAt: string;
  scope: "readonly";
};

export type CoreReadOnlyClient = {
  bootstrap(code: string, signal?: AbortSignal): Promise<CoreReadOnlySession>;
  disconnect(): void;
  isConnected(): boolean;
  getSummary(signal?: AbortSignal): Promise<CoreReadOnlySummary>;
  listProfiles(options?: { limit?: number; cursor?: string; signal?: AbortSignal }): Promise<CoreReadOnlyPage<CoreReadOnlyProfile>>;
  listGroups(options?: { limit?: number; cursor?: string; signal?: AbortSignal }): Promise<CoreReadOnlyPage<CoreReadOnlyGroup>>;
  listRuntimes(options?: { limit?: number; cursor?: string; signal?: AbortSignal }): Promise<CoreReadOnlyPage<CoreReadOnlyRuntime>>;
};

export function createCoreReadOnlyClient(baseURL: string): CoreReadOnlyClient {
  let token: string | undefined;

  const endpoint = baseURL.replace(/\/$/, "");

  async function bootstrap(code: string, signal?: AbortSignal): Promise<CoreReadOnlySession> {
    const normalizedCode = code.trim();
    if (!/^[a-f0-9]{64}$/i.test(normalizedCode)) throw new Error("INVALID_BOOTSTRAP_CODE");
    const response = await fetch(`${endpoint}/api/v1/readonly/session/bootstrap`, {
      method: "POST",
      signal,
      headers: { "Content-Type": "application/json", "X-Request-ID": `ui-${crypto.randomUUID()}` },
      body: JSON.stringify({ code: normalizedCode }),
      credentials: "omit",
      cache: "no-store",
    });
    if (!response.ok) throw new Error("INVALID_BOOTSTRAP_CODE");
    const payload = await response.json() as { token?: string; expires_at?: string; scope?: string };
    if (!payload.token || !payload.expires_at || payload.scope !== "readonly") throw new Error("INVALID_BOOTSTRAP_CODE");
    token = payload.token;
    return { expiresAt: payload.expires_at, scope: "readonly" };
  }

  async function request<T>(path: string, signal?: AbortSignal): Promise<T> {
    if (!token) throw new Error("CORE_NOT_CONNECTED");
    const requestId = `ui-${crypto.randomUUID()}`;
    const response = await fetch(`${endpoint}${path}`, {
      method: "GET",
      signal,
      headers: { Authorization: `Bearer ${token}`, "X-Request-ID": requestId },
      credentials: "omit",
      cache: "no-store",
    });
    if (response.status === 401 || response.status === 403) token = undefined;
    if (!response.ok) throw new Error(`CORE_HTTP_${response.status}`);
    return response.json() as Promise<T>;
  }

  function pagedPath(resource: "profiles" | "groups" | "runtimes", options: { limit?: number; cursor?: string }) {
    const query = new URLSearchParams();
    if (options.limit) query.set("limit", String(options.limit));
    if (options.cursor) query.set("cursor", options.cursor);
    return `/api/v1/readonly/${resource}${query.size ? `?${query.toString()}` : ""}`;
  }

  return {
    bootstrap,
    disconnect() { token = undefined; },
    isConnected() { return Boolean(token); },
    getSummary(signal) { return request<CoreReadOnlySummary>("/api/v1/readonly/summary", signal); },
    listProfiles(options = {}) {
      return request<CoreReadOnlyPage<CoreReadOnlyProfile>>(pagedPath("profiles", options), options.signal);
    },
    listGroups(options = {}) { return request<CoreReadOnlyPage<CoreReadOnlyGroup>>(pagedPath("groups", options), options.signal); },
    listRuntimes(options = {}) { return request<CoreReadOnlyPage<CoreReadOnlyRuntime>>(pagedPath("runtimes", options), options.signal); },
  };
}
