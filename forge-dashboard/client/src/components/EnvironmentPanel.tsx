/**
 * T13 — Identité navigateur / cohérence d'environnement.
 * Projection lecture seule du catalogue de diagnostic du Core : les contrôles
 * (user agent, plateforme, locale, fuseau horaire, géométrie, WebRTC, etc.)
 * sont projetés comme statuts + libellés humains redacted uniquement. Aucune
 * valeur brute d'observation (chaîne UA, coordonnées, hash canvas/audio) ne
 * transite par le dashboard.
 */
import { useState } from "react";
import { Fingerprint, LoaderCircle, RefreshCcw, ShieldCheck } from "lucide-react";
import { toast } from "sonner";
import { CoreEnvDiagnostic, CoreWriteClient } from "@/lib/coreWrite";

function humanizeError(error: unknown): string {
  if (!(error instanceof Error)) return "Erreur inattendue côté Core.";
  const message = error.message;
  if (message.startsWith("CORE_ERROR_")) {
    const code = message.slice("CORE_ERROR_".length);
    if (code === "ENVIRONMENT_DIAGNOSTIC_NOT_FOUND") return "Aucun diagnostic enregistré pour ce profil : lancez un diagnostic via le rail de contrôle local.";
    if (code === "ENVIRONMENT_DIAGNOSTIC_UNAVAILABLE") return "Le catalogue de diagnostic du Core n'est pas configuré sur cette machine.";
    if (code === "ENVIRONMENT_DIAGNOSTIC_ERROR") return "Le catalogue de diagnostic du Core ne répond pas.";
    if (code === "PROFILE_NOT_FOUND") return "Ce profil n'existe plus dans le registre du Core.";
    return `Le Core a refusé la consultation (${code}).`;
  }
  if (message === "CORE_ADMIN_NOT_CONNECTED" || message === "CORE_ADMIN_UNAUTHORIZED") return "Le contrôle local a été retiré ; reliez le jeton d'administration.";
  if (message === "CORE_NOT_LOOPBACK") return "La lecture exige une URL loopback du Core.";
  return "Connexion impossible : le Core local ne répond pas.";
}

const STATUS_STYLE: Record<string, { label: string; cls: string }> = {
  PASS: { label: "Cohérent", cls: "env-status-pass" },
  WARNING: { label: "Dérogation", cls: "env-status-warning" },
  FAIL: { label: "Divergence", cls: "env-status-fail" },
  UNSUPPORTED: { label: "Non pris en charge", cls: "env-status-unsupported" },
  RUNTIME_DEFINED: { label: "Runtime requis", cls: "env-status-runtime" },
};

export type EnvironmentPanelProps = {
  client: CoreWriteClient;
  profileIds: string[];
  selectedProfileId?: string;
};

export function EnvironmentPanel({ client, profileIds, selectedProfileId }: EnvironmentPanelProps) {
  const [diagnostic, setDiagnostic] = useState<CoreEnvDiagnostic | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [loading, setLoading] = useState(false);

  const loadDiagnostic = async () => {
    if (!selectedProfileId) {
      toast.error("Aucun profil sélectionné", { description: "Sélectionnez un profil dans le registre avant de consulter son identité navigateur." });
      return;
    }
    setLoading(true);
    try {
      const { data } = await client.environment.getDiagnostic(selectedProfileId);
      setDiagnostic(data);
      setLoaded(true);
    } catch (error) {
      const message = String(error);
      if (message.includes("CORE_ADMIN_UNAUTHORIZED")) {
        toast.error("Le contrôle local a été retiré", { description: "Le jeton d'administration n'est plus accepté par le Core." });
        setDiagnostic(null);
        setLoaded(false);
      } else {
        toast.error(humanizeError(error));
        setDiagnostic(null);
        setLoaded(false);
      }
    } finally {
      setLoading(false);
    }
  };

  if (!client.isConnected()) {
    return (
      <div className="environment-panel environment-disconnected" role="region" aria-label="Identité navigateur">
        <div className="environment-empty"><ShieldCheck size={24} /><strong>Identité navigateur verrouillée</strong><span>Connectez le contrôle local d'administration pour consulter le diagnostic de cohérence du Core. Les valeurs d'observation ne quittent jamais la machine.</span></div>
      </div>
    );
  }

  return (
    <div className="environment-panel" role="region" aria-label="Identité navigateur">
      <div className="environment-head">
        <div>
          <p className="section-kicker"><span /> Identité navigateur</p>
          <h2>Cohérence <em>configurée → observée</em></h2>
          <p>Le Core compare la configuration du profil aux observations du runtime. Aucune valeur brute — chaînes UA, coordonnées, hash — n'est projetée : statuts et libellés humains uniquement.</p>
        </div>
        <div className="environment-actions">
          <button className="env-refresh" type="button" disabled={loading || !selectedProfileId} onClick={() => void loadDiagnostic()}>
            {loading && <LoaderCircle size={15} className="spin" />} Consulter
          </button>
        </div>
      </div>

      {!selectedProfileId && (
        <div className="environment-empty"><Fingerprint size={22} /><strong>Aucun profil sélectionné</strong><span>Sélectionnez un profil du registre puis consultez son diagnostic. Le Core sérialise les consultations par profil.</span></div>
      )}

      {selectedProfileId && !loaded && !loading && (
        <div className="environment-empty"><Fingerprint size={22} /><strong>Aucun diagnostic consulté</strong><span>Le catalogue de diagnostic est projeté à la demande depuis le Core. Sélectionnez un profil et consultez son état de cohérence.</span></div>
      )}

      {selectedProfileId && loading && (
        <div className="environment-empty"><LoaderCircle size={20} className="spin" /><strong>Lecture du diagnostic</strong><span>Le Core projette l'état de cohérence en lecture seule redacted.</span></div>
      )}

      {diagnostic && (
        <div className="environment-layout">
          <section className="instrument-plate env-summary" aria-label="Verdict du diagnostic">
            <span className="plate-code">ENV / VER / 01</span>
            <div className="detail-field"><label>Profil</label><span>{diagnostic.profile_id}</span></div>
            <div className="detail-field"><label>Étape</label><span>{diagnostic.stage === "consistent" ? "Cohérent" : "Incohérence détectée"}</span></div>
            <div className="detail-field"><label>Statut agrégé</label><span className={`env-agg ${STATUS_STYLE[diagnostic.status]?.cls ?? "env-status-unknown"}`}>{STATUS_STYLE[diagnostic.status]?.label ?? diagnostic.status}</span></div>
            <div className="detail-field"><label>Dernier diagnostic</label><span>{new Date(diagnostic.checked_at).toLocaleString("fr-FR")}</span></div>
            <div className="detail-field"><label>Référence</label><span className="digest-mono">{diagnostic.diagnostic_ref}</span><small>Référence opaque du catalogue Core, sans valeur d'observation.</small></div>
          </section>

          <section className="instrument-plate env-checks" aria-label="Contrôles de cohérence">
            <span className="plate-code">ENV / CHK / 01</span>
            <p className="detail-section-kicker">{diagnostic.checks.length} contrôles projetés</p>
            <ul className="env-check-rows">
              {diagnostic.checks.map((check) => {
                const style = STATUS_STYLE[check.status];
                return (
                  <li key={check.check} className="env-check-row">
                    <span className="env-check-name"><strong>{check.check}</strong><span>{check.detail}</span></span>
                    <span className={`env-check-status ${style?.cls ?? "env-status-unknown"}`} title={style?.label ?? check.status}>{style?.label ?? check.status}</span>
                  </li>
                );
              })}
            </ul>
            {diagnostic.checks.some((c) => c.status === "FAIL") && (
              <small className="env-note">Divergence détectée : le profil ne doit pas être utilisé pour des sessions exigeant cette propriété tant que la cohérence n'est pas rétablie.</small>
            )}
            {diagnostic.checks.some((c) => c.status === "RUNTIME_DEFINED") && (
              <small className="env-note">Contrôles dépendants du runtime qualifié : seule la qualification runtime (T14) peut certifier ces propriétés.</small>
            )}
          </section>
        </div>
      )}
    </div>
  );
}
