/**
 * T11 — Coffre des sauvegardes (Backups/Restore).
 * Client mémoire seule du contrat BACK-01 du Core : archive .flbackup,
 * AES-256-GCM/AAD, publication atomique et restauration isolée vers un nouvel
 * identifiant (jamais d'écrasement implicite). Le dashboard n'expose aucune
 * clé, aucune référence de vault et aucun chemin d'artefact — le Core retourne
 * uniquement des projections redacted.
 */
import { useState } from "react";
import { ArchiveRestore, AlertTriangle, LoaderCircle, Plus, ShieldCheck, Trash2 } from "lucide-react";
import { toast } from "sonner";
import {
  CoreBackupClient,
  CoreBackupDetail,
  CoreBackupSummary,
  CoreRestoreOperation,
  CoreWriteClient,
} from "@/lib/coreWrite";

function humanizeError(error: unknown): string {
  if (!(error instanceof Error)) return "Erreur inattendue côté Core.";
  const message = error.message;
  if (message.startsWith("CORE_ERROR_")) {
    const code = message.slice("CORE_ERROR_".length);
    if (code === "BACKUP_NOT_FOUND") return "Cette sauvegarde n'existe plus dans le registre du Core.";
    if (code === "BACKUP_IN_USE") return "Purge refusée : la sauvegarde est la source d'une restauration validée. Le Core protège la provenance des profils restaurés.";
    if (code === "BACKUP_NOT_PURGEABLE") return "Cette sauvegarde ne peut pas encore être purgée (état intermédiaire). Le Core refuse l'élagage par principe fail-closed.";
    if (code === "TARGET_EXISTS") return "Le profil cible existe déjà : une restauration crée toujours un nouvel identifiant, jamais un écrasement.";
    if (code === "TARGET_BUSY") return "Le profil cible est verrouillé par une autre opération du Core.";
    if (code === "LOOPBACK_REQUIRED") return "Mutation hors loopback refusée : ouvrez le dashboard depuis la machine du Core.";
    if (code === "CORE_ADMIN_NOT_CONNECTED" || code === "CORE_ADMIN_UNAUTHORIZED") return "Le contrôle local a été retiré ; reliez le jeton d'administration.";
    if (code === "BACKUP_STAGING_REUSED" || code === "BACKUP_SNAPSHOT_LOCKED") return "Une sauvegarde est en cours pour ce profil. Le Core sérialise les opérations.";
    if (code === "PROFILE_NOT_FOUND") return "Ce profil n'existe plus dans le registre du Core.";
    return `Le Core a refusé l'opération (${code}).`;
  }
  if (message === "MISSING_TARGET_PROFILE_ID") return "L'identifiant du profil cible est requis.";
  if (message === "MISSING_PROFILE_ID") return "Aucun profil sélectionné.";
  if (message === "CORE_NOT_LOOPBACK") return "Les écritures exigent une URL loopback du Core.";
  return "Connexion impossible : le Core local ne répond pas.";
}

const STATE_LABEL: Record<string, string> = {
  staging: "En cours",
  published_unregistered: "Publiée",
  committed: "Validée",
  quarantined: "Quarantaine",
};

function shortDigest(sha256: string): string {
  if (!sha256) return "—";
  return `${sha256.slice(0, 8)}•••${sha256.slice(-6)}`;
}

export type BackupVaultProps = {
  client: CoreWriteClient;
  profileIds: string[];
};

export function BackupVault({ client, profileIds }: BackupVaultProps) {
  const [backups, setBackups] = useState<CoreBackupSummary[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [loading, setLoading] = useState(false);
  const [selected, setSelected] = useState<CoreBackupSummary | null>(null);
  const [detail, setDetail] = useState<CoreBackupDetail | null>(null);
  const [restores, setRestores] = useState<CoreRestoreOperation[]>([]);
  const [busy, setBusy] = useState<string | null>(null);

  const backupsClient: CoreBackupClient = client.backups;

  const loadBackups = async () => {
    setLoading(true);
    try {
      const { data } = await backupsClient.listBackups();
      setBackups(data);
      setLoaded(true);
    } catch (error) {
      toast.error(humanizeError(error));
      setLoaded(false);
    } finally {
      setLoading(false);
    }
  };

  const loadDetail = async (backup: CoreBackupSummary) => {
    setSelected(backup);
    try {
      const [{ data: detailData }, { data: restoresData }] = await Promise.all([
        backupsClient.getBackup(backup.id),
        backupsClient.getBackupRestores(backup.id),
      ]);
      setDetail(detailData);
      setRestores(restoresData);
    } catch (error) {
      toast.error(humanizeError(error));
      setDetail(null);
      setRestores([]);
    }
  };

  const runAction = async (label: string, id: string, action: () => Promise<unknown>) => {
    if (!client.isConnected()) {
      toast.error("Contrôle local requis", { description: "Reliez le jeton d'administration dans le panneau État du Core." });
      return;
    }
    setBusy(id);
    try {
      await action();
      await loadBackups();
      if (selected) await loadDetail(selected);
      toast.success("Action validée par le Core", { description: `Opération ${label} exécutée et auditée côté Core.`, descriptionClassName: "text-foreground/70" });
    } catch (error) {
      const message = String(error);
      if (message.includes("CORE_ADMIN_UNAUTHORIZED") || message.includes("CORE_ERROR_LOOPBACK")) {
        toast.error("Le contrôle local a été retiré", { description: "Le jeton d'administration n'est plus accepté par le Core." });
      } else {
        toast.error(humanizeError(error));
      }
    } finally {
      setBusy(null);
    }
  };

  if (!client.isConnected()) {
    return (
      <div className="backup-vault backup-vault-disconnected" role="region" aria-label="Sauvegardes">
        <div className="backup-vault-empty"><ShieldCheck size={24} /><strong>Coffre de sauvegarde fermé</strong><span>Connectez le contrôle local d'administration pour consulter les sauvegardes chiffrées du Core. Les clés ne quittent jamais la machine.</span></div>
      </div>
    );
  }

  return (
    <div className="backup-vault" role="region" aria-label="Sauvegardes">
      <div className="backup-vault-head">
        <div>
          <p className="section-kicker"><span /> Coffre de sauvegardes</p>
          <h2>Archives <em>chiffrées</em> et restaurations isolées</h2>
          <p>Aucune donnée en clair dans les réponses Core : empreinte SHA-256 seulement, aucun chemin ni identifiant de clé.</p>
        </div>
        <div className="backup-vault-actions">
          <button className="backup-refresh" type="button" onClick={() => void loadBackups()} disabled={loading}>
            {loading && <LoaderCircle size={15} className="spin" />} Actualiser
          </button>
        </div>
      </div>

      <div className="backup-layout">
        <section className="backup-list instrument-plate" aria-label="Registre des sauvegardes">
          <span className="plate-code">BKV / REG / 01</span>
          {!loaded && !loading && backups.length === 0 && (
            <div className="backup-list-empty">
              <ArchiveRestore size={22} />
              <strong>Aucune sauvegarde dans le registre</strong>
              <span>Créez une archive chiffrée d'un profil via le bouton « Sauvegarder ». L'archive porte une empreinte SHA-256 et un état de publication atomique.</span>
            </div>
          )}
          {loading && backups.length === 0 && (
            <div className="backup-list-empty"><LoaderCircle size={20} className="spin" /><strong>Registre en lecture</strong><span>Le Core projette le registre de sauvegardes en lecture seule redacted.</span></div>
          )}
          <ul className="backup-rows">
            {backups.map((backup) => (
              <li key={backup.id}>
                <button
                  className={`backup-row ${selected?.id === backup.id ? "backup-row-selected" : ""}`}
                  type="button"
                  data-testid={`backup-row-${backup.id.slice(0, 12)}`}
                  onClick={() => void loadDetail(backup)}
                >
                  <span className="backup-name"><strong>{shortDigest(backup.sha256)}</strong><span>profil {backup.profile_id.slice(0, 12)}{backup.profile_id.length > 12 ? "•••" : ""}</span></span>
                  <span className={`backup-state state-${backup.state}`} title={STATE_LABEL[backup.state] ?? backup.state}>{STATE_LABEL[backup.state] ?? backup.state}</span>
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section className="backup-detail instrument-plate" aria-label="Détail de la sauvegarde">
          <span className="plate-code">BKV / DET / 01</span>
          {!selected && !detail && (
            <div className="backup-detail-empty"><ArchiveRestore size={20} /><strong>Sélectionnez une archive</strong><span>L'état, l'empreinte et l'historique des restaurations sont projetés depuis le registre du Core.</span></div>
          )}
          {detail && (
            <div className="backup-detail-body">
              <div className="detail-field"><label>Identifiant</label><span>{detail.id}</span></div>
              <div className="detail-field"><label>Profil source</label><span>{detail.profile_id}</span></div>
              <div className="detail-field"><label>État</label><span className={`backup-state state-${detail.state}`}>{STATE_LABEL[detail.state] ?? detail.state}</span></div>
              <div className="detail-field"><label>Empreinte SHA-256</label><span className="digest-mono">{shortDigest(detail.sha256)}</span><small>Exposée par le Core uniquement comme résumé de confiance, jamais la clé ni le chemin.</small></div>
              <div className="detail-field"><label>Quarantaine</label><span>{detail.quarantined ? "Contenu non vérifiable — isolé par le Core" : "Aucun signal de quarantaine"}</span></div>
              {detail.last_restored_target_profile_id && (
                <div className="detail-field"><label>Dernière restauration vers</label><span>{detail.last_restored_target_profile_id}</span><small>Une restauration crée toujours un nouvel identifiant de profil ; aucun écrasement n'est possible.</small></div>
              )}
              <div className="detail-actions">
                <button
                  className="backup-purge"
                  type="button"
                  data-testid="backup-purge"
                  disabled={busy === `purge-${detail.id}` || detail.state !== "committed"}
                  title={detail.state !== "committed" ? "Seules les archives validées peuvent être purgées (fail-closed)" : "Purger l'archive du registre Core"}
                  onClick={() => {
                    if (!window.confirm(`Purger cette archive du registre Core ? L'élagage est définitif côté registre (fail-closed) ; le Core refuse si l'archive est source d'une restauration validée.`)) return;
                    void runAction("élagage", `purge-${detail.id}`, () => backupsClient.purgeBackup(detail.id));
                  }}
                >
                  {busy === `purge-${detail.id}` ? <LoaderCircle size={15} className="spin" /> : <Trash2 size={15} />} Purger
                </button>
              </div>
            </div>
          )}

          <div className="restore-history">
            <p className="detail-section-kicker">Historique des restaurations</p>
            {!restores.length && <div className="restore-empty"><AlertTriangle size={16} /><span>Aucune restauration enregistrée pour cette archive.</span></div>}
            <ul className="restore-rows">
              {restores.map((op) => (
                <li key={op.restore_id} className={`restore-row state-${op.state}`}>
                  <span><strong>{op.target_profile_id}</strong><span>cible créée depuis {op.source_profile_id.slice(0, 14)}{op.source_profile_id.length > 14 ? "•••" : ""}</span></span>
                  <span className={`backup-state state-${op.state}`}>{op.state === "committed" ? "Validée" : op.state}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="restore-new">
            <p className="detail-section-kicker">Restaurer vers un nouvel identifiant</p>
            <RestoreControl
              sourceProfileId={selected?.profile_id ?? ""}
              backupId={selected?.id ?? ""}
              client={client}
              busy={busy?.startsWith("restore-")}
              onRestored={() => void loadDetail(selected!)}
              profileIds={profileIds}
            />
          </div>
        </section>
      </div>
    </div>
  );
}

function RestoreControl({
  sourceProfileId,
  backupId,
  client,
  busy,
  onRestored,
  profileIds,
}: {
  sourceProfileId: string;
  backupId: string;
  client: CoreWriteClient;
  busy: boolean | undefined;
  onRestored: () => void;
  profileIds: string[];
}) {
  const [target, setTarget] = useState("");
  const [busyBackup, setBusyBackup] = useState<string | null>(null);
  const [suggest, setSuggest] = useState(false);

  if (!backupId) {
    return <div className="restore-quiet"><span>Sélectionnez une archive validée pour amorcer une restauration.</span></div>;
  }

  const deriveSuggestion = () => {
    const nonce = Date.now().toString(36);
    setTarget(`${sourceProfileId.slice(0, 10)}-restore-${nonce}`);
    setSuggest(true);
  };

  const restore = async () => {
    if (!client.isConnected()) return;
    if (!target.trim()) {
      toast.error("L'identifiant cible est requis", { description: "Utilisez la suggestion ou saisissez un identifiant de profil cible qui n'existe pas encore." });
      return;
    }
    const identifier = target.trim();
    if (profileIds.includes(identifier)) {
      toast.error("Identifiant cible déjà utilisé", { description: "Une restauration crée un profil isolé ; choisissez un identifiant qui n'existe pas dans le registre." });
      return;
    }
    setBusyBackup(backupId);
    try {
      await client.backups.restoreBackup(backupId, identifier);
      onRestored();
      toast.success("Restauration isolée validée", { description: `Profil ${identifier} créé depuis l'archive. L'écrasement est impossible par conception.`, descriptionClassName: "text-foreground/70" });
      setTarget("");
      setSuggest(false);
    } catch (error) {
      const message = String(error);
      if (message.includes("CORE_ADMIN_UNAUTHORIZED") || message.includes("CORE_ERROR_LOOPBACK")) {
        toast.error("Le contrôle local a été retiré", { description: "Le jeton d'administration n'est plus accepté par le Core." });
      } else {
        toast.error(humanizeError(error));
      }
    } finally {
      setBusyBackup(null);
    }
  };

  return (
    <div className="restore-control">
      <label className="restore-field">
        <span className="sr-only">Identifiant du profil cible</span>
        <input
          value={target}
          onChange={(event) => { setTarget(event.target.value); setSuggest(false); }}
          placeholder="identifiant du nouveau profil"
          disabled={busy}
          aria-description={suggest ? "Identifiant suggéré ; modifiable." : "Le profil cible n'existe pas encore : il est créé isolé lors de la restauration."}
        />
        <button type="button" className="restore-suggest" data-testid="backup-restore-suggest" onClick={deriveSuggestion} disabled={busy} aria-label="Suggérer un identifiant cible">
          Suggérer
        </button>
      </label>
      <button className="restore-submit" type="button" data-testid="backup-restore-submit" disabled={busy || !target.trim()} onClick={() => void restore()}>
        {busy && busyBackup === backupId ? <LoaderCircle size={15} className="spin" /> : <ArchiveRestore size={15} />} Restaurer
      </button>
      <small className="restore-note">La restauration crée le profil cible à partir de l'archive chiffrée. L'intégrité AES-256-GCM est vérifiée par le Core ; toute archive corrompue ou altérée est refusée et signalée.</small>
    </div>
  );
}

export type BackupSourceActionProps = {
  profileId: string;
  busy: boolean;
  disabled: boolean;
  onSave: () => void;
};

export function BackupCreateAction({ profileId, busy, disabled, onSave }: BackupSourceActionProps) {
  return (
      <button
      className="profile-backup-action"
      aria-label={`Sauvegarder ce profil`}
      type="button"
      data-testid={`backup-create-${profileId}`}
      disabled={busy || disabled}
      title={busy || disabled ? "Une sauvegarde est en cours" : "Sauvegarder ce profil dans le coffre Core"}
      onClick={(event) => {
        event.stopPropagation();
        if (!window.confirm(`Sauvegarder ce profil dans le coffre Core ? L'archive sera chiffrée (AES-256-GCM) et publiée de manière atomique.`)) return;
        onSave();
      }}
    >
      {busy ? <LoaderCircle size={14} className="spin" /> : <ArchiveRestore size={14} />}
    </button>
  );
}
