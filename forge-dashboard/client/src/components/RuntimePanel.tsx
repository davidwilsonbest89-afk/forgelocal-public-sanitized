/**
 * T14 — Qualification runtime Chromium.
 * Projection lecture seule redacted du registre de qualification du Core :
 * état de la machine à états (discovered → qualified → ready → running →
 * stopped), version, architecture et date de qualification. Les chemins de
 * binaires, ports debug, jetons et user-data dirs ne quittent jamais le Core.
 */
import { useEffect, useState } from "react";
import { Cpu, LoaderCircle, ShieldCheck } from "lucide-react";
import { toast } from "sonner";
import { CoreRuntimeRecord, CoreWriteClient } from "@/lib/coreWrite";

function humanizeRuntimeError(error: unknown): string {
  if (!(error instanceof Error)) return "Erreur inattendue côté Core.";
  const message = error.message;
  if (message.startsWith("CORE_ERROR_")) {
    const code = message.slice("CORE_ERROR_".length);
    if (code === "RUNTIME_UNAVAILABLE") return "Le registre de qualification du Core n'est pas configuré sur cette machine.";
    return `Le Core a refusé la consultation (${code}).`;
  }
  if (message === "CORE_ADMIN_NOT_CONNECTED" || message === "CORE_ADMIN_UNAUTHORIZED") return "Le contrôle local a été retiré ; reliez le jeton d'administration.";
  if (message === "CORE_NOT_LOOPBACK") return "La lecture exige une URL loopback du Core.";
  return "Connexion impossible : le Core local ne répond pas.";
}

const STATE_STYLE: Record<string, { label: string; cls: string }> = {
  discovered: { label: "Découvert", cls: "runtime-state-discovered" },
  installed: { label: "Installé", cls: "runtime-state-installed" },
  qualifying: { label: "En qualification", cls: "runtime-state-qualifying" },
  qualified: { label: "Qualifié", cls: "runtime-state-qualified" },
  ready: { label: "Prêt", cls: "runtime-state-ready" },
  running: { label: "En cours", cls: "runtime-state-running" },
  stopped: { label: "Arrêté", cls: "runtime-state-stopped" },
  error: { label: "Erreur", cls: "runtime-state-error" },
};

export type RuntimePanelProps = {
  client: CoreWriteClient;
};

export function RuntimePanel({ client }: RuntimePanelProps) {
  const [records, setRecords] = useState<CoreRuntimeRecord[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [loading, setLoading] = useState(false);

  const loadQualified = async () => {
    setLoading(true);
    try {
      const { data } = await client.runtime.listQualified();
      setRecords(data);
      setLoaded(true);
    } catch (error) {
      const message = String(error);
      if (message.includes("CORE_ADMIN_UNAUTHORIZED")) {
        toast.error("Le contrôle local a été retiré", { description: "Le jeton d'administration n'est plus accepté par le Core." });
        setRecords([]);
        setLoaded(false);
      } else {
        toast.error(humanizeRuntimeError(error));
        setRecords([]);
        setLoaded(false);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (client.isConnected() && !loaded && !loading) {
      void loadQualified();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client.isConnected()]);

  if (!client.isConnected()) {
    return (
      <div className="runtime-panel runtime-disconnected" role="region" aria-label="Runtime qualifié">
        <div className="runtime-empty"><ShieldCheck size={24} /><strong>Runtime verrouillé</strong><span>Connectez le contrôle local d'administration pour consulter le registre de qualification du Core. Les chemins et ports ne quittent jamais la machine.</span></div>
      </div>
    );
  }

  return (
    <div className="runtime-panel" role="region" aria-label="Runtime qualifié" data-testid="runtime-panel">
      <div className="runtime-head">
        <div>
          <p className="section-kicker"><span /> Runtime qualifié</p>
          <h2>Chromium <em>réel</em>, qualifié localement</h2>
          <p>Le Core découvre, hash et qualifie le binaire Chromium installé sur cette machine. Seul le catalogue redacted — état, version, architecture — est projeté au dashboard.</p>
        </div>
        <div className="runtime-actions">
          <button className="rt-refresh" type="button" disabled={loading} onClick={() => void loadQualified()}>
            {loading && <LoaderCircle size={15} className="spin" />} Actualiser
          </button>
        </div>
      </div>

      {!loaded && !loading && (
        <div className="runtime-empty"><Cpu size={22} /><strong>Registre non consulté</strong><span>Le catalogue de qualification est projeté à la demande depuis le Core, en lecture seule redacted.</span></div>
      )}

      {loading && (
        <div className="runtime-empty"><LoaderCircle size={20} className="spin" /><strong>Lecture du registre</strong><span>Le Core projette le catalogue de qualification en lecture seule.</span></div>
      )}

      {loaded && records.length === 0 && (
        <div className="runtime-empty"><Cpu size={22} /><strong>Aucun runtime qualifié</strong><span>Aucun binaire Chromium n'a encore été découvert et qualifié sur cette machine. Un runtime non qualifié ne peut lancer aucune session.</span></div>
      )}

      {records.length > 0 && (
        <div className="runtime-layout">
          <section className="instrument-plate rt-summary" aria-label="Catalogue de qualification">
            <span className="plate-code">RT / QLF / 01</span>
            <ul className="rt-record-rows">
			  {records.map((record, index) => {
				const normalizedState = record.state.toLowerCase();
				const style = STATE_STYLE[normalizedState] ?? { label: record.state, cls: "runtime-state-unknown" };
                return (
                  <li key={index} className="rt-record-row">
					<span className="rt-record-name"><strong>Chromium qualifié</strong><span>{record.version}</span></span>
                    <span className={`rt-record-state ${style.cls}`} title={style.label}>{style.label}</span>
                    {record.qualified_at && <span className="rt-record-time"><small>{new Date(record.qualified_at).toLocaleString("fr-FR")}</small></span>}
                  </li>
                );
              })}
            </ul>
            <small className="rt-note">Chemin du binaire, ports de debug et dossiers utilisateur restent sur la machine ; le dashboard ne reçoit jamais ces valeurs.</small>
          </section>
        </div>
      )}
    </div>
  );
}
