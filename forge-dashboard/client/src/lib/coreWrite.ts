/**
 * Atelier de contrôle local — client d'écritures Core (contrat T09 Profile Writes).
 * This module deliberately retains the administrative Bearer token in a closure
 * only: no localStorage, URL, logging or analytics persistence is permitted.
 * Every mutation is rejected unless the Core base URL resolves to loopback, and
 * write endpoints remain unreachable outside the loopback interface (fail-closed
 * per T05/T09). The dashboard is an API client: it never writes to SQLite or the
 * profile filesystem directly.
 */
export type CoreWriteProfile = {
  id: string;
  name: string;
  runtime_id: string;
  group?: string;
  tags: string[];
  lifecycle_state?: "active" | "archived" | "quarantined";
  created_at: string;
  last_used?: string;
  profile_dir?: string;
};

export type CoreWriteError = {
  code: string;
  message: string;
  reason?: string;
};

export type CoreWriteResult<T> = {
  data: T;
  correlationId?: string;
};

export type CoreWriteCreateSpec = {
  name: string;
  runtime_id: string;
  group?: string;
  tags?: string[];
};

// T10 — Proxy registry contract types (Core-owned; the dashboard never writes
// proxy credentials. secret_ref points at the Core vault; has_secret is a
// presence flag only, never the value itself).
export type CoreProxyType = "http" | "socks5";

export type CoreProxy = {
  id: string;
  name: string;
  type: CoreProxyType;
  host: string;
  port: number;
  region?: string;
  secret_ref?: string;
  has_secret: boolean;
  created_at: string;
  updated_at?: string;
};

export type CoreProxyCreateSpec = {
  name: string;
  type: CoreProxyType;
  host: string;
  port: number;
  region?: string;
  secret_ref?: string;
};

export type CoreProxyClient = {
  listProxies(signal?: AbortSignal): Promise<CoreWriteResult<CoreProxy[]>>;
  createProxy(spec: CoreProxyCreateSpec, signal?: AbortSignal): Promise<CoreWriteResult<CoreProxy>>;
  updateProxy(id: string, patch: Partial<CoreProxyCreateSpec>, signal?: AbortSignal): Promise<CoreWriteResult<CoreProxy>>;
  deleteProxy(id: string, signal?: AbortSignal): Promise<CoreWriteResult<{ id: string }>>;
  assignProxy(id: string, profileId: string, signal?: AbortSignal): Promise<CoreWriteResult<{ proxy_id: string; profile_id: string }>>;
  unassignProxy(id: string, profileId: string, signal?: AbortSignal): Promise<CoreWriteResult<{ proxy_id: string; profile_id: string }>>;
};

// T11 — Backup/Restore contract types (Core-owned, redacted projections only:
// no key identifiers, vault references or absolute artifact paths are ever
// exposed to the dashboard).
export type CoreBackupSummary = {
  id: string;
  profile_id: string;
  sha256: string;
  state: "staging" | "published_unregistered" | "committed" | "quarantined";
  error_code?: string;
  created_at: string;
  updated_at: string;
};

export type CoreBackupDetail = {
  id: string;
  profile_id: string;
  sha256: string;
  state: string;
  error_code?: string;
  quarantined?: boolean;
  last_restored_target_profile_id?: string;
  created_at: string;
  updated_at: string;
};

export type CoreRestoreOperation = {
  restore_id: string;
  backup_id: string;
  source_profile_id: string;
  target_profile_id: string;
  state: "staging" | "committed" | "failed";
  error_code?: string;
  created_at: string;
  updated_at: string;
};

export type CoreBackupClient = {
  listBackups(signal?: AbortSignal): Promise<CoreWriteResult<CoreBackupSummary[]>>;
  getBackup(id: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreBackupDetail>>;
  getBackupRestores(id: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreRestoreOperation[]>>;
  createBackup(profileId: string, signal?: AbortSignal): Promise<CoreWriteResult<{ id: string }>>;
  restoreBackup(id: string, targetProfileId: string, signal?: AbortSignal): Promise<CoreWriteResult<{ restore_id: string; target_profile_id: string }>>;
  purgeBackup(id: string, signal?: AbortSignal): Promise<CoreWriteResult<{ id: string }>>;
};

// T14 — Runtime qualification contract (read-only redacted projection of a
// Core-owned qualification registry; binary paths, debug ports, tokens and
// user-data dirs are never exposed to the dashboard).
export type CoreRuntimeRecord = {
  state: string;
  version: string;
  arch: string;
  qualified_at?: string;
};

export type CoreRuntimeClient = {
  listQualified(signal?: AbortSignal): Promise<CoreWriteResult<CoreRuntimeRecord[]>>;
};

// T13 — Environment consistency diagnostic contract (read-only projection of a
// Core-owned diagnostic catalog; no raw observables — UA strings, coordinates,
// raw canvas/audio hashes — are ever exposed to the dashboard).
export type CoreEnvCheckStatus = "PASS" | "WARNING" | "FAIL" | "UNSUPPORTED" | "RUNTIME_DEFINED";

export type CoreEnvCheck = {
  check: string;
  status: CoreEnvCheckStatus;
  detail: string;
};

export type CoreEnvDiagnostic = {
  profile_id: string;
  stage: string;
  status: string;
  checks: CoreEnvCheck[];
  checked_at: string;
  diagnostic_ref: string;
};

export type CoreEnvClient = {
  getDiagnostic(profileId: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreEnvDiagnostic>>;
};

// T15 — Local controlled automation contract (Core-owned sessions; the dashboard
// never receives raw HTML or image bytes — only sha256 projections and lengths,
// and the client refuses non-local fixture URLs before they ever reach the Core).
export type CoreSessionSummary = {
  session_id: string;
  profile_id: string;
  runtime_id: string;
};

export type CoreSessionNavigateResult = {
  status: number;
  url: string;
};

export type CoreSessionProjection = {
  sha256_hex: string;
  length_bytes: number;
};

export type CoreSessionClient = {
  listSessions(signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionSummary[]>>;
  createSession(profileId: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionSummary>>;
  deleteSession(sessionId: string, signal?: AbortSignal): Promise<CoreWriteResult<{ session_id: string }>>;
  navigate(sessionId: string, url: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionNavigateResult>>;
  content(sessionId: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionProjection>>;
  screenshot(sessionId: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionProjection>>;
  rawContentSignalOnly?(sessionId: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionProjection>>;
};

export type CoreWriteClient = {
  bind(token: string): void;
  unbind(): void;
  isConnected(): boolean;
  environment: CoreEnvClient;
  runtime: CoreRuntimeClient;
  sessions: CoreSessionClient;
  createProfile(spec: CoreWriteCreateSpec, signal?: AbortSignal): Promise<CoreWriteResult<CoreWriteProfile>>;
  getProfile(id: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreWriteProfile>>;
  updateProfile(id: string, patch: Record<string, unknown>, signal?: AbortSignal): Promise<CoreWriteResult<CoreWriteProfile>>;
  archiveProfile(id: string, signal?: AbortSignal): Promise<CoreWriteResult<{ id: string }>>;
  reopenProfile(id: string, signal?: AbortSignal): Promise<CoreWriteResult<{ id: string }>>;
  addProfileTag(id: string, tag: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreWriteProfile>>;
  removeProfileTag(id: string, tag: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreWriteProfile>>;
  proxies: CoreProxyClient;
  backups: CoreBackupClient;
};

const CORRELATION_HEADER = "x-correlation-id";

function readCorrelationId(response: Response): string | undefined {
  return response.headers.get(CORRELATION_HEADER) || response.headers.get(CORRELATION_HEADER.replace(/^x-/, "X-")) || undefined;
}

async function readAdminAuthReason(response: Response): Promise<string | undefined> {
  try {
    const payload = await response.clone().json() as { error?: CoreWriteError };
    return payload.error?.reason;
  } catch {
    return undefined;
  }
}

function adminAuthError(reason?: string): Error {
  const normalized = reason?.trim().toLowerCase();
  if (normalized === "expired" || normalized === "revoked" || normalized === "malformed" || normalized === "missing" || normalized === "invalid") {
    return new Error(`CORE_ADMIN_${normalized.toUpperCase()}`);
  }
  return new Error("CORE_ADMIN_UNAUTHORIZED");
}

function isLoopback(hostname: string): boolean {
  const normalized = hostname.replace(/^\[|\]$/g, "");
  return normalized === "127.0.0.1" || normalized === "::1" || normalized === "localhost" || normalized === "0.0.0.0";
}

// T15 — Memory-only blob projection: the body is hashed client-side and the raw
// bytes are discarded immediately. The dashboard never holds raw HTML or image
// pixels — only a sha256 hex digest and the content length.
async function fetchBlobProjection(url: string, authToken: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionProjection>> {
  if (!authToken) throw new Error("CORE_ADMIN_NOT_CONNECTED");
  const response = await fetch(url, {
    signal,
    headers: { Authorization: `Bearer ${authToken}`, "X-Request-ID": `ui-${crypto.randomUUID()}` },
    credentials: "omit",
    cache: "no-store",
  });
  if (response.status === 401) throw adminAuthError(await readAdminAuthReason(response));
  if (response.status === 403) throw new Error("CORE_ADMIN_UNAUTHORIZED");
  if (!response.ok) throw new Error(`CORE_HTTP_${response.status}`);
  const bytes = await response.arrayBuffer();
  const digest = await crypto.subtle.digest("SHA-256", bytes);
  // Release the raw bytes immediately — never stored, logged or displayed.
  const sha256Hex = Array.from(new Uint8Array(digest)).map(b => b.toString(16).padStart(2, "0")).join("");
  return { data: { sha256_hex: sha256Hex, length_bytes: bytes.byteLength } };
}

export function createCoreWriteClient(baseURL: string): CoreWriteClient {
  const endpoint = baseURL.replace(/\/$/, "");
  const configuredURL = (() => {
    try {
      const parsed = new URL(endpoint);
      return parsed.protocol === "http:" && isLoopback(parsed.hostname) ? endpoint : "";
    } catch {
      return "";
    }
  })();

  let token: string | undefined;

  // T15 — helper interne : le HTML brut est haché puis libéré ; jamais stocké.
  async function sessionContentHash(sessionId: string, signal?: AbortSignal): Promise<CoreWriteResult<CoreSessionProjection>> {
    if (!token) throw new Error("CORE_ADMIN_NOT_CONNECTED");
    return fetch(`${configuredURL}/api/sessions/${encodeURIComponent(sessionId)}/content`, {
      signal,
      headers: { Authorization: `Bearer ${token}`, "X-Request-ID": `ui-${crypto.randomUUID()}` },
      credentials: "omit",
      cache: "no-store",
    }).then(async response => {
      if (response.status === 401) throw adminAuthError(await readAdminAuthReason(response));
      if (response.status === 403) throw new Error("CORE_ADMIN_UNAUTHORIZED");
      if (!response.ok) throw new Error(`CORE_HTTP_${response.status}`);
      const payload = (await response.json()) as { data?: string };
      const bytes = new TextEncoder().encode(payload.data ?? "");
      const digest = await crypto.subtle.digest("SHA-256", bytes);
      const sha256Hex = Array.from(new Uint8Array(digest)).map(b => b.toString(16).padStart(2, "0")).join("");
      return { data: { sha256_hex: sha256Hex, length_bytes: bytes.byteLength } };
    });
  }

  async function mutate<T>(method: string, path: string, body?: unknown, signal?: AbortSignal): Promise<CoreWriteResult<T>> {
    if (!token) throw new Error("CORE_ADMIN_NOT_CONNECTED");
    if (!configuredURL) throw new Error("CORE_NOT_LOOPBACK");
    const response = await fetch(`${configuredURL}${path}`, {
      method,
      signal,
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
        "X-Request-ID": `ui-${crypto.randomUUID()}`,
      },
      credentials: "omit",
      cache: "no-store",
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
    const correlationId = readCorrelationId(response);
    if (response.status === 401 || response.status === 403) token = undefined;
    if (response.status === 401) throw adminAuthError(await readAdminAuthReason(response));
    if (!response.ok) {
      let detail: CoreWriteError | undefined;
      try {
        const payload = await response.json() as { error?: CoreWriteError };
        if (payload.error?.code) detail = payload.error;
      } catch {
        // Non-JSON body: fall back to the HTTP status code error.
      }
      throw new Error(detail ? `CORE_ERROR_${detail.code}` : `CORE_HTTP_${response.status}`);
    }
    const data = (await response.json()) as { data?: T };
    return { data: data.data as T, correlationId };
  }

  const clientRef: CoreWriteClient = {
    bind(nextToken) {
      if (!nextToken.trim()) throw new Error("CORE_ADMIN_TOKEN_EMPTY");
      token = nextToken.trim();
    },
    unbind() {
      token = undefined;
    },
    isConnected() {
      return Boolean(token) && configuredURL !== "";
    },
    createProfile(spec, signal) {
      if (!spec.name.trim()) throw new Error("MISSING_NAME");
      return mutate<CoreWriteProfile>("POST", "/api/profiles", spec, signal);
    },
    getProfile(id, signal) {
      return mutate<CoreWriteProfile>("GET", `/api/profiles/${encodeURIComponent(id)}`, undefined, signal);
    },
    updateProfile(id, patch, signal) {
      return mutate<CoreWriteProfile>("PUT", `/api/profiles/${encodeURIComponent(id)}`, patch, signal);
    },
    archiveProfile(id, signal) {
      return mutate<{ id: string }>("POST", `/api/profiles/${encodeURIComponent(id)}/archive`, undefined, signal);
    },
    reopenProfile(id, signal) {
      return mutate<{ id: string }>("POST", `/api/profiles/${encodeURIComponent(id)}/reopen`, undefined, signal);
    },
    addProfileTag(id, tag, signal) {
      const normalized = tag.trim().toLocaleLowerCase("fr-FR");
      if (!normalized) throw new Error("INVALID_TAG");
      return mutate<CoreWriteProfile>("POST", `/api/profiles/${encodeURIComponent(id)}/tags/${encodeURIComponent(normalized)}`, undefined, signal);
    },
    removeProfileTag(id, tag, signal) {
      const normalized = tag.trim().toLocaleLowerCase("fr-FR");
      if (!normalized) throw new Error("INVALID_TAG");
      return mutate<CoreWriteProfile>("DELETE", `/api/profiles/${encodeURIComponent(id)}/tags/${encodeURIComponent(normalized)}`, undefined, signal);
    },
    environment: {
      getDiagnostic(profileId, signal) {
        if (!profileId) throw new Error("MISSING_PROFILE_ID");
        return mutate<CoreEnvDiagnostic>("GET", `/api/v1/environment/profiles/${encodeURIComponent(profileId)}`, undefined, signal);
      },
    },
    runtime: {
      listQualified(signal) {
        // Le contrat Core renvoie {runtimes} en racine (sans enveloppe `data`) :
        // on fetche la réponse brute plutôt que de passer par mutate() qui unwrap `data`.
        if (!token) throw new Error("CORE_ADMIN_NOT_CONNECTED");
        if (!configuredURL) throw new Error("CORE_NOT_LOOPBACK");
        return fetch(`${configuredURL}/api/v1/runtimes/qualified`, {
          method: "GET",
          signal,
          headers: { Authorization: `Bearer ${token}`, "X-Request-ID": `ui-${crypto.randomUUID()}` },
          credentials: "omit",
          cache: "no-store",
        }).then(async response => {
          if (response.status === 401 || response.status === 403) token = undefined;
          if (response.status === 401) throw adminAuthError(await readAdminAuthReason(response));
          if (!response.ok) throw new Error(`CORE_HTTP_${response.status}`);
          const payload = (await response.json()) as { runtimes?: CoreRuntimeRecord[] };
          return { data: payload.runtimes ?? [], correlationId: readCorrelationId(response) };
        });
      },
    },
    sessions: {
      listSessions(signal) {
        // Le Core renvoie {data: [{session_id,profile_id,runtime_id}], total} (null si vide).
        return mutate<{ items: CoreSessionSummary[] }>("GET", "/api/sessions", undefined, signal).then(res => ({ data: res.data.items ?? res.data ?? [], correlationId: res.correlationId }));
      },
      rawContentSignalOnly(sessionId: string, signal?: AbortSignal) {
        // Utilisé uniquement en interne par content() — hash + libération.
        return sessionContentHash(sessionId, signal);
      },
      createSession(profileId, signal) {
        if (!profileId) throw new Error("MISSING_PROFILE_ID");
        // Le Core renvoie {data: {session_id, profile_id, runtime_id}} en une
        // seule enveloppe ; on déroule pour exposer la session directement.
        return mutate<{ data?: CoreSessionSummary }>("POST", "/api/sessions", { profile_id: profileId }, signal).then(res => ({
          data: (res.data?.data ?? res.data) as CoreSessionSummary,
          correlationId: res.correlationId,
        }));
      },
      deleteSession(sessionId, signal) {
        if (!sessionId) throw new Error("MISSING_SESSION_ID");
        return mutate<{ data?: { session_id: string } }>("DELETE", `/api/sessions/${encodeURIComponent(sessionId)}`, undefined, signal).then(res => ({
          data: { session_id: res.data?.data?.session_id ?? sessionId },
          correlationId: res.correlationId,
        }));
      },
      navigate(sessionId, url, signal) {
        if (!sessionId) throw new Error("MISSING_SESSION_ID");
        // Fail-closed local-only policy, mirroring the Core policy: only file://
        // and http(s) over 127.0.0.1/localhost/::1 are allowed. Anything else is
        // rejected before it ever reaches the network, and the Core refuses it too.
        let normalized: string;
        try {
          const parsed = new URL(url);
          normalized = parsed.href;
          if (parsed.protocol === "file:") {
            // file:// stays local by definition.
          } else if (parsed.protocol === "http:" || parsed.protocol === "https:") {
            if (!isLoopback(parsed.hostname)) throw new Error("URL_NOT_LOCAL");
          } else {
            throw new Error("URL_NOT_LOCAL");
          }
        } catch (err) {
          if (err instanceof Error && err.message === "URL_NOT_LOCAL") throw err;
          throw new Error("INVALID_URL");
        }
        return mutate<CoreSessionNavigateResult>("POST", `/api/sessions/${encodeURIComponent(sessionId)}/navigate`, { url: normalized }, signal);
      },
      async content(sessionId: string, signal?: AbortSignal) {
        if (!sessionId) throw new Error("MISSING_SESSION_ID");
        return sessionContentHash(sessionId, signal);
      },
      async screenshot(sessionId, signal) {
        if (!sessionId) throw new Error("MISSING_SESSION_ID");
        // Le Core renvoie le PNG brut ; on le hache et on le libère immédiatement.
        if (!token) throw new Error("CORE_ADMIN_NOT_CONNECTED");
        const response = await fetch(`${configuredURL}/api/sessions/${encodeURIComponent(sessionId)}/screenshot`, {
          signal,
          headers: { Authorization: `Bearer ${token}`, "X-Request-ID": `ui-${crypto.randomUUID()}` },
          credentials: "omit",
          cache: "no-store",
        });
        if (response.status === 401) throw adminAuthError(await readAdminAuthReason(response));
        if (response.status === 403) throw new Error("CORE_ADMIN_UNAUTHORIZED");
        if (!response.ok) throw new Error(`CORE_HTTP_${response.status}`);
        const bytes = await response.arrayBuffer();
        const digest = await crypto.subtle.digest("SHA-256", bytes);
        const sha256Hex = Array.from(new Uint8Array(digest)).map(b => b.toString(16).padStart(2, "0")).join("");
        return { data: { sha256_hex: sha256Hex, length_bytes: bytes.byteLength } };
      },
    },
    proxies: {
      listProxies(signal) {
        return mutate<{ items: CoreProxy[] }>("GET", "/api/proxies", undefined, signal).then(res => ({ data: res.data.items, correlationId: res.correlationId }));
      },
      createProxy(spec, signal) {
        if (!spec.name.trim()) throw new Error("MISSING_NAME");
        if (spec.port <= 0 || spec.port > 65535 || !Number.isInteger(spec.port)) throw new Error("INVALID_PROXY_PORT");
        if (!spec.host.trim()) throw new Error("INVALID_PROXY_HOST");
        if (spec.type !== "http" && spec.type !== "socks5") throw new Error("INVALID_PROXY_TYPE");
        if (spec.secret_ref && !/^proxy\.ref\.[A-Za-z0-9._-]{1,128}$/.test(spec.secret_ref)) throw new Error("INVALID_PROXY_SECRET_REF");
        return mutate<CoreProxy>("POST", "/api/proxies", spec, signal);
      },
      updateProxy(id, patch, signal) {
        const body: Partial<CoreProxyCreateSpec> = {};
        if (patch.name !== undefined) body.name = patch.name;
        if (patch.type !== undefined) body.type = patch.type;
        if (patch.host !== undefined) body.host = patch.host;
        if (patch.port !== undefined) body.port = patch.port;
        if (patch.region !== undefined) body.region = patch.region;
        return mutate<CoreProxy>("PUT", `/api/proxies/${encodeURIComponent(id)}`, body, signal);
      },
      deleteProxy(id, signal) {
        return mutate<{ id: string }>("DELETE", `/api/proxies/${encodeURIComponent(id)}`, undefined, signal);
      },
      assignProxy(id, profileId, signal) {
        if (!profileId) throw new Error("MISSING_PROFILE_ID");
        return mutate<{ proxy_id: string; profile_id: string }>("POST", `/api/proxies/${encodeURIComponent(id)}/assign?profile_id=${encodeURIComponent(profileId)}`, undefined, signal);
      },
      unassignProxy(id, profileId, signal) {
        if (!profileId) throw new Error("MISSING_PROFILE_ID");
        return mutate<{ proxy_id: string; profile_id: string }>("DELETE", `/api/proxies/${encodeURIComponent(id)}/assign?profile_id=${encodeURIComponent(profileId)}`, undefined, signal);
      },
    },
    backups: {
      listBackups(signal) {
        return mutate<{ items: CoreBackupSummary[] }>("GET", "/api/v1/backups", undefined, signal).then(res => ({ data: res.data.items, correlationId: res.correlationId }));
      },
      getBackup(id, signal) {
        return mutate<CoreBackupDetail>("GET", `/api/v1/backups/${encodeURIComponent(id)}`, undefined, signal);
      },
      getBackupRestores(id, signal) {
        return mutate<{ items: CoreRestoreOperation[] }>("GET", `/api/v1/backups/${encodeURIComponent(id)}/restores`, undefined, signal).then(res => ({ data: res.data.items, correlationId: res.correlationId }));
      },
      createBackup(profileId, signal) {
        if (!profileId) throw new Error("MISSING_PROFILE_ID");
        return mutate<{ id: string }>("POST", `/api/v1/profiles/${encodeURIComponent(profileId)}/backups`, undefined, signal);
      },
      restoreBackup(id, targetProfileId, signal) {
        if (!targetProfileId.trim()) throw new Error("MISSING_TARGET_PROFILE_ID");
        return mutate<{ restore_id: string; target_profile_id: string }>("POST", `/api/v1/backups/${encodeURIComponent(id)}/restore`, { target_profile_id: targetProfileId.trim() }, signal);
      },
      purgeBackup(id, signal) {
        return mutate<{ id: string }>("DELETE", `/api/v1/backups/${encodeURIComponent(id)}`, undefined, signal);
      },
    },
  };
  return clientRef;
}
