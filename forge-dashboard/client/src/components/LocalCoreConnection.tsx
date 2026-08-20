/**
 * ForgeLocal — Connexion locale mémoire seule.
 * Le code de bootstrap et le Bearer court restent uniquement dans la mémoire
 * du navigateur : aucune URL, journal, analytique ou persistance navigateur.
 */
import { FormEvent, useEffect, useRef, useState } from "react";
import { KeyRound, Link2, LoaderCircle, LockKeyhole, Unplug } from "lucide-react";
import {
  CoreReadOnlyGroup,
  CoreReadOnlyProfile,
  CoreReadOnlyRuntime,
  CoreReadOnlySummary,
  createCoreReadOnlyClient,
} from "@/lib/coreReadOnly";

export type LocalCoreSnapshot = {
  summary: CoreReadOnlySummary;
  profiles: CoreReadOnlyProfile[];
  groups: CoreReadOnlyGroup[];
  runtimes: CoreReadOnlyRuntime[];
  expiresAt: string;
};

type LocalCoreConnectionProps = {
  onConnected: (snapshot: LocalCoreSnapshot) => void;
  onDisconnected: () => void;
  onWriteConnected?: (adminToken: string) => void;
  onWriteDisconnected?: () => void;
};

function isLoopbackHostname(hostname: string) {
	return hostname === "localhost" || /^127(?:\.\d{1,3}){3}$/.test(hostname) || hostname === "::1" || hostname === "[::1]";
}

export function resolveLocalCoreBaseURL() {
	const candidate = String(import.meta.env.VITE_CORE_BASE_URL ?? "").trim();
	if (candidate) {
		try {
			const configured = new URL(candidate);
			if (configured.protocol === "http:" && isLoopbackHostname(configured.hostname)) {
				return configured.origin;
			}
		} catch {
			// A malformed or non-loopback build value must never redirect a local code.
		}
	}
	return `http://${window.location.hostname === "[::1]" ? "[::1]" : window.location.hostname}:19280`;
}

export function LocalCoreConnection({ onConnected, onDisconnected, onWriteConnected, onWriteDisconnected }: LocalCoreConnectionProps) {
	const clientRef = useRef(createCoreReadOnlyClient(resolveLocalCoreBaseURL()));
  const isLoopback = isLoopbackHostname(window.location.hostname);
  const [code, setCode] = useState("");
  const [state, setState] = useState<"idle" | "connecting" | "connected" | "expired" | "error">("idle");
  const [expiresAt, setExpiresAt] = useState("");
  const [message, setMessage] = useState("Le Core local n’est pas encore relié.");
  const [adminToken, setAdminToken] = useState("");
  const [adminState, setAdminState] = useState<"idle" | "linking" | "linked" | "error">("idle");
  const [adminMessage, setAdminMessage] = useState("Les écritures exigent le jeton d’administration du Core (mémoire seule, jamais enregistré).");

  const linkAdmin = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isLoopback) return;
    const trimmed = adminToken.trim();
    if (trimmed.length < 12) {
      setAdminState("error");
      setAdminMessage("Jeton trop court : consultez la sortie du Core (.api-token) sur cette machine.");
      return;
    }
    setAdminState("linking");
    try {
      const probe = await fetch(`${resolveLocalCoreBaseURL()}/api/health`, {
        headers: { Authorization: `Bearer ${trimmed}`, "X-Request-ID": `ui-${crypto.randomUUID()}` },
        credentials: "omit",
        cache: "no-store",
      });
      if (probe.status === 401 || probe.status === 403) {
        setAdminState("error");
        setAdminMessage("Jeton refusé par le Core. Vérifiez la valeur exacte de .api-token.");
        return;
      }
      if (!probe.ok) {
        setAdminState("error");
        setAdminMessage("Le Core local ne répond pas à ce jeton.");
        return;
      }
      setAdminToken("");
      setAdminState("linked");
      setAdminMessage("Contrôle local actif : les écritures sont déléguées au Core.");
      onWriteConnected?.(trimmed);
    } catch {
      setAdminState("error");
      setAdminMessage("Connexion impossible. Vérifiez que ce dashboard est servi par le Core local.");
    }
  };

  const unlinkAdmin = (event?: FormEvent) => {
    event?.preventDefault();
    setAdminState("idle");
    setAdminToken("");
    setAdminMessage("Les écritures exigent le jeton d’administration du Core (mémoire seule, jamais enregistré).");
    onWriteDisconnected?.();
  };

  useEffect(() => {
    if (state !== "connected" || !expiresAt) return;
    const remaining = new Date(expiresAt).getTime() - Date.now();
    const timeout = window.setTimeout(() => {
      clientRef.current.disconnect();
      setState("expired");
      setMessage("La session lecture seule a expiré. Générez un nouveau code local.");
      onDisconnected();
    }, Math.max(remaining, 0));
    return () => window.clearTimeout(timeout);
  }, [expiresAt, onDisconnected, state]);

  const connect = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!isLoopback) return;
    setState("connecting");
    setMessage("Échange du code local en cours…");
    try {
      const session = await clientRef.current.bootstrap(code);
      const [summary, profiles, groups, runtimes] = await Promise.all([
        clientRef.current.getSummary(),
        clientRef.current.listProfiles({ limit: 100 }),
        clientRef.current.listGroups({ limit: 100 }),
        clientRef.current.listRuntimes({ limit: 100 }),
      ]);
      setCode("");
      setExpiresAt(session.expiresAt);
      setState("connected");
      setMessage("Lecture Core active — session limitée et non persistée.");
      onConnected({ summary, profiles: profiles.data, groups: groups.data, runtimes: runtimes.data, expiresAt: session.expiresAt });
    } catch (error) {
      clientRef.current.disconnect();
      setState("error");
      setMessage(error instanceof Error && error.message === "INVALID_BOOTSTRAP_CODE"
        ? "Code refusé, expiré ou déjà utilisé. Générez-en un nouveau localement."
        : "Connexion impossible. Vérifiez que ce dashboard est servi par le Core local.");
      onDisconnected();
    }
  };

  const disconnect = () => {
    clientRef.current.disconnect();
    setState("idle");
    setExpiresAt("");
    setCode("");
    setMessage("La session a été supprimée de la mémoire du navigateur.");
    onDisconnected();
  };

  if (state === "connected") {
    return (
      <div className="local-core-connection local-core-connected">
        <div className="local-core-title"><span className="local-core-dot" /> Lecture Core sécurisée</div>
        <p>{message}</p>
        <small>Expire à {new Date(expiresAt).toLocaleTimeString("fr-FR", { hour: "2-digit", minute: "2-digit" })}</small>
        <LocalCoreAdminPanel
          isLoopback={isLoopback}
          linked={adminState === "linked"}
          linking={adminState === "linking"}
          message={adminMessage}
          token={adminToken}
          onTokenChange={setAdminToken}
          onLink={linkAdmin}
          onUnlink={unlinkAdmin}
        />
        <button data-testid="local-core-disconnect" type="button" onClick={disconnect}><Unplug size={13} /> Déconnecter</button>
      </div>
    );
  }

  if (!isLoopback) {
    return (
      <div className="local-core-connection local-core-unavailable">
          <div className="local-core-title" data-testid="local-core-hosted-refusal"><KeyRound size={13} /> Connexion Core locale</div>
        <p>Cette prévisualisation hébergée ne peut pas recevoir un code local.</p>
        <small>Ouvrez le dashboard depuis l’adresse loopback du Core pour échanger un code à usage unique.</small>
      </div>
    );
  }

  return (
    <form className="local-core-connection" onSubmit={connect}>
      <div className="local-core-title"><KeyRound size={13} /> Session locale lecture seule</div>
      <p>{message}</p>
      <label>
        <span className="sr-only">Code local à usage unique</span>
          <input
            autoComplete="off"
            data-testid="local-core-code"
          inputMode="text"
          maxLength={64}
          onChange={(event) => setCode(event.target.value.trim())}
          placeholder="Code local à usage unique"
          spellCheck={false}
          value={code}
        />
      </label>
      <small>Code valable 10 min, échangé une seule fois. Il n’est jamais enregistré.</small>
      <button data-testid="local-core-connect" disabled={state === "connecting" || code.length !== 64} type="submit">
        {state === "connecting" ? <LoaderCircle className="spin" size={13} /> : <Link2 size={13} />}
        Relier au Core local
      </button>
    </form>
  );
}

function LocalCoreAdminPanel({
  isLoopback,
  linked,
  linking,
  message,
  token,
  onTokenChange,
  onLink,
  onUnlink,
}: {
  isLoopback: boolean;
  linked: boolean;
  linking: boolean;
  message: string;
  token: string;
  onTokenChange: (next: string) => void;
  onLink: (event: FormEvent<HTMLFormElement>) => void;
  onUnlink: () => void;
}) {
  if (linked) {
    return (
      <form className="local-core-admin local-core-admin-linked" onSubmit={onUnlink}>
        <div className="local-core-title"><span className="local-core-dot local-core-admin-dot" /><LockKeyhole size={13} /> Contrôle local (écritures)</div>
        <p data-testid="local-core-admin-message">{message}</p>
        <button data-testid="local-core-admin-unlink" type="submit">Retirer le jeton</button>
      </form>
    );
  }
  if (!isLoopback) {
    return (
      <div className="local-core-admin">
        <div className="local-core-title"><LockKeyhole size={13} /> Contrôle local (écritures)</div>
        <p>Cette prévisualisation hébergée ne peut pas recevoir de jeton d’administration.</p>
        <small>Ouvrez le dashboard depuis l’adresse loopback du Core.</small>
      </div>
    );
  }
  return (
    <form className="local-core-admin" onSubmit={onLink}>
      <div className="local-core-title"><LockKeyhole size={13} /> Contrôle local (écritures)</div>
      <p data-testid="local-core-admin-message">{message}</p>
      <label>
        <span className="sr-only">Jeton d’administration</span>
        <input
          autoComplete="off"
          data-testid="local-core-admin-token"
          inputMode="text"
          maxLength={128}
          onChange={(event) => onTokenChange(event.target.value.trim())}
          placeholder="Jeton d’administration (mémoire seule)"
          spellCheck={false}
          value={token}
        />
      </label>
      <small>Issu du fichier .api-token du Core. Jamais enregistré, valide uniquement en loopback.</small>
      <button data-testid="local-core-admin-link" disabled={linking || token.length < 12} type="submit">
        {linking ? <LoaderCircle className="spin" size={13} /> : <Link2 size={13} />}
        Activer les écritures
      </button>
    </form>
  );
}
