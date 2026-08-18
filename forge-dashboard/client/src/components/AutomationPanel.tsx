/**
 * T15 — Automation locale contrôlée.
 * Surfaces d'écriture bornées sur les sessions d'automation du Core local :
 * ouverture de session sur un profil réel, navigation strictement loopback-only
 * (fail-closed client ET serveur), empreintes SHA-256 du contenu et des captures
 * (les octets bruts sont hashés côté client puis libérés — jamais stockés,
 * affichés ou logués), et fermeture de session.
 * Contrat mémoire seule : aucun HTML brut, aucune image, aucune coordonnée.
 */
import { useEffect, useState } from "react";
import { LoaderCircle, LockKeyhole, PlayCircle, RefreshCcw, ShieldCheck, TerminalSquare, Trash2, X } from "lucide-react";
import { toast } from "sonner";
import { CoreSessionSummary, CoreWriteClient } from "@/lib/coreWrite";

function humanizeError(error: unknown): string {
  if (!(error instanceof Error)) return "Erreur inattendue côté Core.";
  const message = error.message;
  if (message.startsWith("CORE_ERROR_")) {
    const code = message.slice("CORE_ERROR_".length);
    if (code === "INVALID_URL") return "L'URL refusée n'est pas une adresse locale : l'automation n'atteint jamais un réseau externe.";
    if (code === "SESSION_NOT_FOUND") return "Cette session n'existe plus dans le Core.";
    if (code === "PROFILE_NOT_FOUND") return "Le profil de cette session n'existe plus dans le registre du Core.";
    if (code === "SESSIONS_DISABLED") return "Les sessions d'automation ne sont pas activées sur ce Core.";
    if (code === "TOO_MANY_SESSIONS") return "La limite globale de sessions simultanées du Core est atteinte : fermez une session existante.";
    return `Le Core a refusé la commande (${code}).`;
  }
  if (message === "INVALID_URL" || message === "URL_NOT_LOCAL") return "L'URL saisie n'est pas une adresse locale : l'automation est bouclée sur la machine uniquement.";
  if (message === "MISSING_SESSION_ID") return "Aucune session active : ouvrez une session avant d'exécuter une commande.";
  if (message === "CORE_ADMIN_NOT_CONNECTED" || message === "CORE_ADMIN_UNAUTHORIZED") return "Le contrôle local a été retiré ; reliez le jeton d'administration.";
  if (message === "CORE_NOT_LOOPBACK") return "L'automation exige une URL loopback du Core.";
  return "Connexion impossible : le Core local ne répond pas.";
}

const LOCAL_HINTS = [
  { value: "about:blank", label: "about:blank (page vierge locale)" },
  { value: "http://127.0.0.1:19280/", label: "Core local (port 19280)" },
  { value: "http://localhost:3000/", label: "dashboard local (port 3000)" },
];

export type AutomationPanelProps = {
  client: CoreWriteClient;
  profileIds: string[];
  selectedProfileId?: string;
};

export function AutomationPanel({ client, profileIds, selectedProfileId }: AutomationPanelProps) {
  const [busy, setBusy] = useState<Record<string, boolean>>({});
  const [sessions, setSessions] = useState<CoreSessionSummary[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [navigateUrl, setNavigateUrl] = useState("");
  const [lastNavigate, setLastNavigate] = useState<{ ok: boolean; message: string; url?: string } | null>(null);
  const [contentDigest, setContentDigest] = useState<{ sha256_hex: string; length_bytes: number } | null>(null);
  const [screenshotDigest, setScreenshotDigest] = useState<{ sha256_hex: string; length_bytes: number } | null>(null);
  const [loadNonce, setLoadNonce] = useState(0);

  // T15 — Recharge automatique : au montage et après toute mutation de session
  // (ouverture, fermeture, rafraîchissement), l'inventaire est repris depuis le
  // Core — jamais depuis un état local préservé.
  useEffect(() => {
    loadSessions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadNonce]);

  // T15 — Le panneau se charge à la demande seulement après liaison du contrôle
  // local, pour ne jamais projeter de sessions quand le Core n'est pas atteint.
  const loadSessions = () => {
    setBusy((b) => ({ ...b, list: true }));
    void client.sessions
      .listSessions()
      .then(({ data }) => {
        setSessions(Array.isArray(data) ? data : []);
        setLoadNonce((n) => n + 1);
      })
      .catch(() => setSessions([]))
      .finally(() => setBusy((b) => ({ ...b, list: false })));
  };

  const run = async (id: string, action: () => Promise<unknown>) => {
    if (busy[id]) return;
    setBusy((b) => ({ ...b, [id]: true }));
    try {
      await action();
    } catch (error) {
      toast.error("Commande d'automation refusée", { description: humanizeError(error) });
    } finally {
      setBusy((b) => ({ ...b, [id]: false }));
    }
  };

  const selectedProfileReal = Boolean(selectedProfileId && profileIds.includes(selectedProfileId));

  const openSession = () => {
    if (!selectedProfileReal || !selectedProfileId) {
      toast.error("Profil requis", { description: "Seul un profil réel du registre Core peut ouvrir une session d'automation locale." });
      return;
    }
    void run("open", () =>
      client.sessions.createSession(selectedProfileId).then(({ data }) => {
        setActiveSessionId(data.session_id);
        setLastNavigate(null);
        setContentDigest(null);
        setScreenshotDigest(null);
        toast.success("Session locale ouverte", { description: "Le navigateur qualifié est piloté sur la boucle locale uniquement." });
        setLoadNonce((n) => n + 1);
      })
    );
  };

  const closeSession = (sessionId: string) => {
    void run(`close-${sessionId}`, () =>
      client.sessions.deleteSession(sessionId).then(() => {
        if (activeSessionId === sessionId) setActiveSessionId(null);
        setLoadNonce((n) => n + 1);
        toast.success("Session fermée", { description: "Le processus local a été libéré." });
      })
    );
  };

  // T15 — La session active affichée doit toujours exister dans l'inventaire Core ;
  // toute session disparue du Core est immédiatement désélectionnée.
  useEffect(() => {
    if (activeSessionId && sessions.length > 0 && !sessions.some((s) => s.session_id === activeSessionId)) {
      setActiveSessionId(null);
      setContentDigest(null);
      setScreenshotDigest(null);
    }
  }, [activeSessionId, sessions]);

  const submitNavigate = (event: React.FormEvent) => {
    event.preventDefault();
    if (!activeSessionId) return;
    const url = navigateUrl.trim();
    if (!url) return;
    void run("navigate", () =>
      client.sessions.navigate(activeSessionId, url).then(({ data }) => {
        setLastNavigate({ ok: true, message: "Navigation locale acceptée", url: data?.url ?? url });
        setContentDigest(null);
        setScreenshotDigest(null);
        toast.success("Navigation locale enregistrée", { description: "Seule l'empreinte et la longueur sont projetées — jamais l'URL atteinte." });
      }).catch((error) => {
        setLastNavigate({ ok: false, message: humanizeError(error), url });
      })
    );
  };

  const fetchDigest = (which: "content" | "screenshot") => {
    if (!activeSessionId) return;
    void run(which, () =>
      (which === "content" ? client.sessions.content(activeSessionId) : client.sessions.screenshot(activeSessionId))
        .then(({ data }) => {
          if (which === "content") setContentDigest(data);
          else setScreenshotDigest(data);
          toast.success("Empreinte calculée", { description: "Les octets bruts ont été libérés ; seule l'empreinte SHA-256 subsiste." });
        })
    );
  };

  return (
    <section className="catalog-panel instrument-plate" data-testid="automation-panel"><span className="plate-code">AUT / T15</span>
      <div className="catalog-panel-heading"><div><TerminalSquare size={17} /><h3>Automation locale</h3></div>
        <button type="button" className="icon-button" aria-label="Actualiser les sessions" onClick={loadSessions} disabled={busy.list}><RefreshCcw size={14} /></button>
      </div>

      {!selectedProfileReal && (
        <div className="automation-empty" data-testid="automation-no-profile"><LockKeyhole size={18} /><strong>Profil réel requis</strong>
          <p>Sélectionnez un profil du registre Core dans le rail pour ouvrir une session d'automation. Aucun runtime n'est lancé avant votre commande explicite, et aucune URL externe n'est atteinte.</p>
          <button type="button" className="action-primary" disabled onClick={() => undefined}><PlayCircle size={14} /> Ouvrir une session locale</button>
        </div>
      )}

      {selectedProfileReal && sessions.length === 0 && (
        <div className="automation-empty" data-testid="automation-empty"><PlayCircle size={18} /><strong>Aucune session active</strong>
          <p>Le panneau n'atteint que la boucle locale : navigation, empreinte de contenu et de capture sont bornées par le Core (fail-closed).</p>
          <button type="button" className="action-primary" onClick={openSession} data-testid="automation-open-session"><PlayCircle size={14} /> Ouvrir une session locale</button>
        </div>
      )}

      {selectedProfileReal && sessions.length > 0 && (
        <div className="automation-list" data-testid="automation-sessions">
          {sessions.map((session) => (
            <article key={session.session_id} className={`automation-session ${activeSessionId === session.session_id ? "automation-session-active" : ""}`} onClick={() => setActiveSessionId(session.session_id)}>
              <div><strong>session {session.session_id.slice(0, 12)}…</strong><span>profil {session.profile_id?.slice(0, 12)}… · runtime local</span></div>
              <button type="button" className="icon-button" aria-label={`Fermer la session ${session.session_id}`} onClick={(event) => { event.stopPropagation(); closeSession(session.session_id); }} disabled={busy[`close-${session.session_id}`]}><Trash2 size={14} /></button>
            </article>
          ))}

          {activeSessionId && (
            <div className="automation-commands" data-testid="automation-commands">
              <form className="automation-navigate" onSubmit={submitNavigate}>
                <label className="sr-only" htmlFor="automation-url">URL locale à atteindre</label>
                <input id="automation-url" value={navigateUrl} onChange={(event) => setNavigateUrl(event.target.value)} placeholder="http://127.0.0.1:… · about:blank · http://localhost:…" data-testid="automation-url-input" />
                <button type="submit" disabled={busy.navigate} data-testid="automation-navigate-button"><LoaderCircle size={14} className={busy.navigate ? "spin" : ""} /> Atteindre</button>
              </form>
              <div className="automation-hints" aria-label="Adresses locales suggérées">{LOCAL_HINTS.map((hint) => (
                <button key={hint.value} type="button" className="hint-chip" onClick={() => setNavigateUrl(hint.value)}>{hint.label}</button>
              ))}</div>

              {lastNavigate && <p className={`automation-feedback ${lastNavigate.ok ? "feedback-good" : "feedback-bad"}`} data-testid="automation-navigate-feedback">
                {lastNavigate.ok ? <ShieldCheck size={13} /> : <X size={13} />}
                <span>{lastNavigate.message}{lastNavigate.url && !lastNavigate.ok ? ` : ${lastNavigate.url}` : ""}</span>
              </p>}

              <div className="automation-digests">
                <button type="button" className="digest-button" onClick={() => fetchDigest("content")} disabled={busy.content} data-testid="automation-content-digest"><LockKeyhole size={13} /><span>{busy.content ? <LoaderCircle size={13} className="spin" /> : "Empreinte du contenu"}</span></button>
                <button type="button" className="digest-button" onClick={() => fetchDigest("screenshot")} disabled={busy.screenshot} data-testid="automation-screenshot-digest"><LockKeyhole size={13} /><span>{busy.screenshot ? <LoaderCircle size={13} className="spin" /> : "Empreinte de la capture"}</span></button>
              </div>

              {contentDigest && (
                <p className="digest-line" data-testid="automation-content-digest-line"><LockKeyhole size={12} /><span>Contenu : <code>{contentDigest.sha256_hex.slice(0, 16)}…</code> · {contentDigest.length_bytes} octets — les octets bruts ont été libérés.</span></p>
              )}
              {screenshotDigest && (
                <p className="digest-line" data-testid="automation-screenshot-digest-line"><LockKeyhole size={12} /><span>Capture : <code>{screenshotDigest.sha256_hex.slice(0, 16)}…</code> · {screenshotDigest.length_bytes} octets — les octets bruts ont été libérés.</span></p>
              )}
            </div>
          )}
        </div>
      )}
    </section>
  );
}
