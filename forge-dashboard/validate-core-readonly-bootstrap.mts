/**
 * BOOTSTRAP-RO-01 — preuve client assainie.
 * Le script ne journalise jamais de code, token, en-tête Authorization ni corps HTTP.
 */
import { createCoreReadOnlyClient } from "./client/src/lib/coreReadOnly";

const baseURL = process.env.FORGELOCAL_CORE_BASE_URL;
const code = process.env.BOOTSTRAP_CODE;

if (!baseURL || !code || !/^[a-f0-9]{64}$/i.test(code)) {
  throw new Error("CONFIGURATION_BOOTSTRAP_RO_INVALIDE");
}

function assert(condition: unknown, message: string): asserts condition {
  if (!condition) throw new Error(message);
}

type Observation = {
  url: string;
  method: string;
  hasAuthorization: boolean;
  bodyContainsCode: boolean;
};

const observed: Observation[] = [];
const storageCalls: string[] = [];
const originalFetch = globalThis.fetch.bind(globalThis);

const memoryStorage = {
  getItem(key: string) { storageCalls.push(`get:${key}`); return null; },
  setItem(key: string, value: string) { storageCalls.push(`set:${key}:${value.length}`); },
  removeItem(key: string) { storageCalls.push(`remove:${key}`); },
  clear() { storageCalls.push("clear"); },
  key() { storageCalls.push("key"); return null; },
  get length() { storageCalls.push("length"); return 0; },
};

Object.assign(globalThis, {
  localStorage: memoryStorage,
  sessionStorage: memoryStorage,
  indexedDB: { open() { storageCalls.push("indexeddb:open"); } },
});

globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
  const url = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
  const inheritedHeaders = input instanceof Request ? input.headers : undefined;
  const headers = new Headers(init?.headers ?? inheritedHeaders);
  const body = typeof init?.body === "string" ? init.body : "";
  observed.push({
    url,
    method: init?.method ?? (input instanceof Request ? input.method : "GET"),
    hasAuthorization: headers.has("Authorization"),
    bodyContainsCode: body.includes(code),
  });
  return originalFetch(input, init);
};

const client = createCoreReadOnlyClient(baseURL);
const session = await client.bootstrap(code);
assert(session.scope === "readonly", "PORTEE_SESSION_INCORRECTE");
assert(client.isConnected(), "CLIENT_NON_CONNECTE_APRES_BOOTSTRAP");
await client.getSummary();
await client.listProfiles({ limit: 50 });

assert(observed.length === 3, "NOMBRE_REQUETES_CORE_INATTENDU");
assert(observed.every((entry) => new URL(entry.url).origin === new URL(baseURL).origin), "DESTINATION_HORS_CORE");
assert(observed.every((entry) => !entry.url.includes(code) && !new URL(entry.url).searchParams.has("token")), "SECRET_PRESENT_URL");
assert(observed[0].method === "POST" && !observed[0].hasAuthorization && observed[0].bodyContainsCode, "BOOTSTRAP_HTTP_INCORRECT");
assert(observed.slice(1).every((entry) => entry.method === "GET" && entry.hasAuthorization && !entry.bodyContainsCode), "LECTURES_HTTP_INCORRECTES");
assert(storageCalls.length === 0, "PERSISTANCE_NAVIGATEUR_DETECTEE");
assert(JSON.stringify(client) === "{}", "TOKEN_EXPOSE_PAR_OBJET_CLIENT");

globalThis.fetch = async () => new Response(JSON.stringify({ error: { code: "UNAUTHORIZED" } }), { status: 401 });
try {
  await client.getSummary();
  throw new Error("LE_401_N_A_PAS_ETE_PROPAGE");
} catch (error) {
  assert(error instanceof Error && error.message === "CORE_HTTP_401", "ERREUR_401_INATTENDUE");
}
assert(!client.isConnected(), "TOKEN_NON_INVALIDE_APRES_401");

console.log(JSON.stringify({
  decision: "BOOTSTRAP_RO_CLIENT_E2E_PASS",
  client_memory_only: "PASS",
  token_absent_url: "PASS",
  storage_absente: "PASS",
  analytics_absentes_du_client: "PASS",
  invalidation_apres_401: "PASS",
  requests_core_observees: observed.length,
}));
