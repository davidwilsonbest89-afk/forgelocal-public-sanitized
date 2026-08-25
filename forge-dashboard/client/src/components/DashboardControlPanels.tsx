import { useMemo, useRef, useState } from "react";
import { Check, CircleHelp, FileArchive, FileCheck2, Filter, KeyRound, LifeBuoy, ListChecks, LockKeyhole, RotateCcw, Settings2, ShieldAlert, Trash2, Upload, Users, X } from "lucide-react";
import { toast } from "sonner";
import type { CoreExtensionSeries, CoreExtensionVersion, CoreWriteClient } from "@/lib/coreWrite";

export type AuditEntry = {
  id: string;
  action: string;
  result: "success" | "error" | "info";
  detail: string;
  at: string;
};

export type DashboardProfileRef = { id: string; name: string };

export type AdvancedFilterState = {
  lifecycle: "all" | "active" | "archived" | "quarantined";
  proxy: "all" | "configured" | "missing";
  candidateOnly: boolean;
  tag: string;
};

function readableError(error: unknown): string {
  const raw = error instanceof Error ? error.message : String(error);
  const code = raw.replace(/^CORE_ERROR_/, "").replace(/^CORE_HTTP_/, "HTTP ");
  const labels: Record<string, string> = {
    PERMISSION_ACK_REQUIRED: "La liste de permissions normalisée doit être reconnue exactement.",
    HIGH_RISK_ACK_REQUIRED: "Une acceptation explicite HIGH_RISK est obligatoire.",
    VERSION_NOT_APPROVED: "La version doit être approuvée avant affectation.",
    PROFILE_NOT_FOUND: "Le profil cible n’existe pas.",
    FORBIDDEN: "Le Core interdit cette opération (403).",
    SERIES_NOT_FOUND: "La série d’extension n’existe pas.",
    VERSION_NOT_FOUND: "La version d’extension n’existe pas.",
    VERSION_REVOKED: "La version est révoquée ou en quarantaine.",
    PURGE_NOT_ALLOWED: "La purge est refusée tant que la version est référencée ou active.",
    INVALID_ARCHIVE: "Le fichier ZIP local est invalide.",
    MANIFEST_INVALID: "Le manifest de l’extension est invalide.",
    CONCURRENT_MUTATION: "Le Core a refusé la mutation concurrente.",
    INTEGRITY_MISMATCH: "L’intégrité du package géré ne correspond plus au digest enregistré.",
    EXTENSION_REPOSITORY_ERROR: "Le Core a refusé l’opération (500).",
    CORE_ADMIN_EXPIRED: "Le jeton d’administration a expiré.",
    CORE_ADMIN_REVOKED: "Le jeton d’administration a été révoqué.",
  };
  return labels[code] ?? code;
}

export function LocalWorkspacePanel({ onAudit }: { onAudit: (action: string, result: AuditEntry["result"], detail: string) => void }) {
  const [selected, setSelected] = useState("atelier-local");
  const [workspaces, setWorkspaces] = useState([
    { id: "atelier-local", name: "Cette machine", detail: "loopback · principal" },
    { id: "qa-local", name: "Validation synthétique", detail: "fixtures · isolé" },
  ]);
  const [newName, setNewName] = useState("");
  const addWorkspace = () => {
    const name = newName.trim();
    if (!name) return;
    const id = `workspace-${Date.now()}`;
    setWorkspaces((items) => [...items, { id, name, detail: "mémoire locale · non synchronisé" }]);
    setSelected(id);
    setNewName("");
    onAudit("workspace.created", "success", `${name} créé en mémoire locale`);
    toast.success("Espace de travail créé", { description: "La sélection reste locale à cette session." });
  };
  return <section className="dashboard-control-panel instrument-plate" data-testid="workspace-panel">
    <span className="plate-code">WRK / LOCAL</span><div className="control-panel-heading"><div><Users size={17} /><h2>Espaces de travail</h2></div><span className="catalog-mode">mémoire locale</span></div>
    <p className="control-panel-copy">Les espaces structurent l’atelier du Dashboard sans synchroniser de comptes, secrets ou données utilisateur.</p>
    <div className="workspace-list">{workspaces.map((item) => <button type="button" key={item.id} className={`workspace-option ${selected === item.id ? "workspace-option-active" : ""}`} onClick={() => { setSelected(item.id); onAudit("workspace.selected", "success", item.name); }} data-testid={`workspace-option-${item.id}`}><span><strong>{item.name}</strong><small>{item.detail}</small></span>{selected === item.id && <Check size={15} />}</button>)}</div>
    <div className="inline-form"><label htmlFor="workspace-name">Nouvel espace<input id="workspace-name" value={newName} onChange={(event) => setNewName(event.target.value)} placeholder="Ex. Audit local" /></label><button type="button" className="action-primary" onClick={addWorkspace} disabled={!newName.trim()} data-testid="workspace-create">Créer</button></div>
  </section>;
}

export function AuditPanel({ entries, onClear }: { entries: AuditEntry[]; onClear: () => void }) {
  return <section className="dashboard-control-panel instrument-plate" data-testid="audit-panel"><span className="plate-code">LOG / LOCAL</span><div className="control-panel-heading"><div><ListChecks size={17} /><h2>Journal d’audit</h2></div><button type="button" className="icon-button" aria-label="Effacer la vue du journal" onClick={onClear}><Trash2 size={15} /></button></div><p className="control-panel-copy">Journal de feedback de cette session Dashboard. Les écritures Core restent auditées par le Core et ne sont pas remplacées par cette vue.</p><div className="audit-list">{entries.length === 0 && <div className="catalog-empty"><ListChecks size={18} /><strong>Aucun événement local</strong><span>Les prochaines actions apparaîtront ici.</span></div>}{entries.map((entry) => <article key={entry.id} className="audit-row"><span className={`audit-state audit-state-${entry.result}`} /><div><strong>{entry.action}</strong><span>{entry.detail}</span><small>{entry.at}</small></div></article>)}</div></section>;
}

export function SettingsPanel({ onAudit }: { onAudit: (action: string, result: AuditEntry["result"], detail: string) => void }) {
  const [confirmRisk, setConfirmRisk] = useState(true);
  const [reducedMotion, setReducedMotion] = useState(false);
  const [compact, setCompact] = useState(false);
  const update = (label: string, setter: (value: boolean) => void, value: boolean) => { setter(value); onAudit("settings.updated", "success", `${label}: ${value ? "activé" : "désactivé"}`); };
  return <section className="dashboard-control-panel instrument-plate" data-testid="settings-panel"><span className="plate-code">CFG / SESSION</span><div className="control-panel-heading"><div><Settings2 size={17} /><h2>Réglages</h2></div><span className="catalog-mode">mémoire seule</span></div><p className="control-panel-copy">Ces réglages gouvernent l’interface locale et ne modifient pas le Core, le coffre ou le système hôte.</p><div className="settings-list"><label className="setting-row"><span><strong>Confirmer HIGH_RISK</strong><small>Exiger une action volontaire avant approbation à risque.</small></span><input type="checkbox" checked={confirmRisk} onChange={(event) => update("confirmation high-risk", setConfirmRisk, event.target.checked)} data-testid="setting-confirm-risk" /></label><label className="setting-row"><span><strong>Mouvement réduit</strong><small>Respecter une présentation sans animation non essentielle.</small></span><input type="checkbox" checked={reducedMotion} onChange={(event) => update("mouvement réduit", setReducedMotion, event.target.checked)} data-testid="setting-reduced-motion" /></label><label className="setting-row"><span><strong>Vue compacte</strong><small>Réduire les espacements de la table de profils.</small></span><input type="checkbox" checked={compact} onChange={(event) => update("vue compacte", setCompact, event.target.checked)} data-testid="setting-compact" /></label></div></section>;
}

export function HelpPanel() {
  return <section className="dashboard-control-panel instrument-plate" data-testid="help-panel"><span className="plate-code">HELP / LOCAL</span><div className="control-panel-heading"><div><CircleHelp size={17} /><h2>Aide</h2></div><span className="catalog-mode">guide embarqué</span></div><p className="control-panel-copy">ForgeLocal reste local-first : le Dashboard affiche les projections redacted et délègue les mutations au Core loopback.</p><div className="help-grid"><article><LifeBuoy size={16} /><strong>Connexion</strong><span>Le token admin reste en mémoire et expire côté Core.</span></article><article><KeyRound size={16} /><strong>Clavier</strong><span>Tab atteint chaque contrôle ; Échap ferme les panneaux et dialogues.</span></article><article><LockKeyhole size={16} /><strong>Fail-closed</strong><span>Une URL externe, un runtime candidat ou une permission non reconnue est refusé.</span></article></div><blockquote className="help-quote">Aucune donnée de coffre ne transite par cette interface.</blockquote></section>;
}

export function NotificationsPanel({ notifications, onReadAll }: { notifications: Array<{ id: string; title: string; detail: string; read: boolean }>; onReadAll: () => void }) {
  return <section className="dashboard-control-panel instrument-plate" data-testid="notifications-panel"><span className="plate-code">NTF / LOCAL</span><div className="control-panel-heading"><div><ShieldAlert size={17} /><h2>Notifications</h2></div><button type="button" className="secondary-action" onClick={onReadAll} data-testid="notifications-read-all">Tout marquer lu</button></div><p className="control-panel-copy">Les notifications affichent uniquement les événements de cette session locale.</p><div className="notification-list">{notifications.map((item) => <article key={item.id} className={`notification-row ${item.read ? "notification-row-read" : ""}`}><span className="notification-dot" /><div><strong>{item.title}</strong><span>{item.detail}</span></div>{!item.read && <em>nouveau</em>}</article>)}</div></section>;
}

export function AdvancedFiltersPanel({ filters, onChange }: { filters: AdvancedFilterState; onChange: (next: AdvancedFilterState) => void }) {
  return <section className="advanced-filter-panel instrument-plate" data-testid="advanced-filters-panel"><span className="plate-code">FLT / ADVANCED</span><div className="control-panel-heading"><div><Filter size={17} /><h3>Filtres avancés</h3></div><span className="catalog-mode">sans mutation</span></div><div className="advanced-filter-grid"><label>Lifecycle<select value={filters.lifecycle} onChange={(event) => onChange({ ...filters, lifecycle: event.target.value as AdvancedFilterState["lifecycle"] })}><option value="all">Tous</option><option value="active">Actifs</option><option value="archived">Archivés</option><option value="quarantined">Quarantaine</option></select></label><label>Proxy<select value={filters.proxy} onChange={(event) => onChange({ ...filters, proxy: event.target.value as AdvancedFilterState["proxy"] })}><option value="all">Tous</option><option value="configured">Configuré</option><option value="missing">Manquant</option></select></label><label>Tag<input value={filters.tag} onChange={(event) => onChange({ ...filters, tag: event.target.value })} placeholder="ex. France" /></label><label className="setting-row-inline"><input type="checkbox" checked={filters.candidateOnly} onChange={(event) => onChange({ ...filters, candidateOnly: event.target.checked })} /> Runtime candidat uniquement</label></div></section>;
}

function versionPermissions(version: CoreExtensionVersion): string[] {
  return Array.from(new Set([
    ...(version.manifest.permissions ?? []), ...(version.manifest.optional_permissions ?? []), ...(version.manifest.host_permissions ?? []), ...(version.manifest.optional_host_permissions ?? []), ...(version.manifest.content_script_matches ?? []),
  ])).sort();
}

function VersionCard({ series, version, client, profiles, onAudit, onReload, onAuthLost }: { series: CoreExtensionSeries; version: CoreExtensionVersion; client: CoreWriteClient; profiles: DashboardProfileRef[]; onAudit: (action: string, result: AuditEntry["result"], detail: string) => void; onReload: () => void; onAuthLost: () => void }) {
  const [expanded, setExpanded] = useState(false);
  const [ack, setAck] = useState<string[]>([]);
  const [acceptHighRisk, setAcceptHighRisk] = useState(false);
  const [profileId, setProfileId] = useState(profiles[0]?.id ?? "");
  const [rollbackTarget, setRollbackTarget] = useState("");
  const [busy, setBusy] = useState("");
  const [feedback, setFeedback] = useState("");
  const permissions = versionPermissions(version);
  const highRisk = (version.risk_categories?.length ?? 0) > 0;
  const run = async (action: string, fn: () => Promise<unknown>) => {
    setBusy(action);
    try { await fn(); setFeedback(`${action} · Core confirmé`); onAudit(`extension.${action}`, "success", `${version.id} · Core confirmé`); toast.success("Opération extension appliquée", { description: `${action} · ${version.id}` }); await onReload(); }
    catch (error) { const detail = readableError(error); setFeedback(detail); onAudit(`extension.${action}`, "error", detail); if (error instanceof Error && (error.message.startsWith("CORE_ADMIN_") || error.message === "CORE_ERROR_FORBIDDEN")) onAuthLost(); toast.error("Opération extension refusée", { description: detail }); }
    finally { setBusy(""); }
  };
  const togglePermission = (permission: string, checked: boolean) => setAck((current) => checked ? [...current, permission] : current.filter((item) => item !== permission));
  const canApprove = permissions.every((permission) => ack.includes(permission)) && (!highRisk || acceptHighRisk);
  const rollbackVersions = series.versions.filter((candidate) => candidate.id !== version.id && ["approved", "archived"].includes(candidate.state));
  return <article className="extension-version-card" data-testid={`extension-version-${version.id}`}><div className="extension-version-header"><div><strong>{version.manifest.name || series.id} · v{version.number}</strong><span>{version.state} · digest {version.digest_preview} · {version.size} octets</span></div><div className="extension-version-badges"><span className={`extension-state extension-state-${version.state}`}>{version.state}</span>{highRisk && <span className="risk-badge"><ShieldAlert size={12} /> HIGH_RISK</span>}<button type="button" className="icon-button" aria-label={`Inspecter ${version.id}`} onClick={() => setExpanded((value) => !value)} data-testid={`extension-inspect-${version.id}`}>{expanded ? <X size={15} /> : <FileCheck2 size={15} />}</button></div></div>
    {expanded && <div className="extension-inspection" data-testid={`extension-inspection-${version.id}`}><div className="inspection-grid"><span>Provenance</span><strong>Core · {version.format} · digest {version.digest_preview}</strong><span>Manifest</span><strong>{version.manifest.version || "non déclaré"} · format {version.manifest.manifest_version || "—"}</strong><span>Signature</span><strong>Non exposée par le contrat Core ; intégrité du blob vérifiée par le Core.</strong><span>Allowlist</span><strong>Permissions normalisées et reconnues avant approbation.</strong></div><div className="permission-list">{permissions.length === 0 && <span>Aucune permission déclarée.</span>}{permissions.map((permission) => <label key={permission}><input type="checkbox" checked={ack.includes(permission)} onChange={(event) => togglePermission(permission, event.target.checked)} />{permission}</label>)}</div>{highRisk && <label className="high-risk-confirm"><input type="checkbox" checked={acceptHighRisk} onChange={(event) => setAcceptHighRisk(event.target.checked)} data-testid={`extension-high-risk-${version.id}`} />J’accepte explicitement le périmètre HIGH_RISK : {(version.risk_categories ?? []).join(", ")}</label>}<button type="button" className="action-primary" disabled={busy === "approve" || !canApprove || version.state !== "imported"} onClick={() => void run("approve", () => client.extensions.approve(version.id, ack, acceptHighRisk))} data-testid={`extension-approve-${version.id}`}><Check size={14} /> Approuver avec allowlist</button></div>}
    {feedback && <p className="extension-feedback" data-testid={`extension-feedback-${version.id}`}>{feedback}</p>}<div className="extension-actions">{["approved", "archived"].includes(version.state) && <><label className="compact-select">Profil<select value={profileId} onChange={(event) => setProfileId(event.target.value)} data-testid={`extension-profile-${version.id}`}><option value="">Choisir</option>{profiles.map((profile) => <option key={profile.id} value={profile.id}>{profile.name}</option>)}</select></label><button type="button" className="secondary-action" disabled={!profileId || busy === "assign"} onClick={() => void run("assign", () => client.extensions.assign(version.id, profileId))} data-testid={`extension-assign-${version.id}`}><Users size={14} /> Affecter</button></>}{rollbackVersions.length > 0 && <><label className="compact-select">Rollback<select value={rollbackTarget} onChange={(event) => setRollbackTarget(event.target.value)} data-testid={`extension-rollback-target-${version.id}`}><option value="">Version cible</option>{rollbackVersions.map((candidate) => <option key={candidate.id} value={candidate.id}>v{candidate.number} · {candidate.state}</option>)}</select></label><button type="button" className="secondary-action" disabled={!rollbackTarget || busy === "rollback"} onClick={() => void run("rollback", () => client.extensions.rollback(series.id, rollbackTarget))} data-testid={`extension-rollback-${version.id}`}><RotateCcw size={14} /> Rollback</button></>}{!["quarantined", "revoked"].includes(version.state) && <button type="button" className="secondary-action danger-action" disabled={busy === "revoke"} onClick={() => { if (window.confirm("Mettre cette version en quarantaine ?")) void run("revoke", () => client.extensions.revoke(version.id)); }} data-testid={`extension-revoke-${version.id}`}><ShieldAlert size={14} /> Révoquer / quarantaine</button>}{["quarantined", "revoked", "archived"].includes(version.state) && series.active_version_id !== version.id && <button type="button" className="secondary-action danger-action" disabled={busy === "purge"} onClick={() => { if (window.confirm("Purger définitivement cette version du registre Core ?")) void run("purge", () => client.extensions.purge(version.id)); }} data-testid={`extension-purge-${version.id}`}><Trash2 size={14} /> Purger</button>}</div>
  </article>;
}

export function ExtensionsPanel({ client, profiles, onAudit, onAuthLost }: { client: CoreWriteClient; profiles: DashboardProfileRef[]; onAudit: (action: string, result: AuditEntry["result"], detail: string) => void; onAuthLost: () => void }) {
  const [result, setResult] = useState<CoreExtensionSeries[]>([]);
  const [loading, setLoading] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [seriesId, setSeriesId] = useState("");
  const [updateSeries, setUpdateSeries] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);
  const load = async () => {
    setLoading(true);
    try { const response = await client.extensions.list(); setResult(response.data.data ?? []); onAudit("extension.inspect", "success", `${response.data.total} série(s) Core`); }
    catch (error) { onAudit("extension.inspect", "error", readableError(error)); toast.error("Inventaire extensions indisponible", { description: readableError(error) }); }
    finally { setLoading(false); }
  };
  const importFile = async () => {
    if (!file) return;
    setLoading(true);
    try { await client.extensions.importPackage(file, seriesId || undefined); setFile(null); if (fileRef.current) fileRef.current.value = ""; toast.success("Package importé dans le Core"); onAudit("extension.import", "success", file.name); await load(); }
    catch (error) { onAudit("extension.import", "error", readableError(error)); toast.error("Import extension refusé", { description: readableError(error) }); }
    finally { setLoading(false); }
  };
  const updateFile = async (nextFile: File | null) => {
    if (!nextFile || !updateSeries) return;
    setLoading(true);
    try { await client.extensions.updatePackage(updateSeries, nextFile); toast.success("Version extension ajoutée"); onAudit("extension.update", "success", nextFile.name); await load(); }
    catch (error) { onAudit("extension.update", "error", readableError(error)); toast.error("Mise à jour refusée", { description: readableError(error) }); }
    finally { setLoading(false); }
  };
  const totalVersions = useMemo(() => result.reduce((total, series) => total + series.versions.length, 0), [result]);
  return <section className="dashboard-control-panel instrument-plate extensions-panel" data-testid="extensions-panel"><span className="plate-code">T28 / EXT / CORE</span><div className="control-panel-heading"><div><FileArchive size={17} /><h2>Extensions locales</h2></div><button type="button" className="icon-button" aria-label="Actualiser les extensions" onClick={() => void load()} disabled={loading} data-testid="extensions-refresh"><RotateCcw size={15} /></button></div><p className="control-panel-copy">Import, inspection redacted, allowlist de permissions, approbation HIGH_RISK, affectation, révocation/quarantaine, rollback et purge passent par le Core loopback. Aucun package ou secret n’est affiché dans le Dashboard.</p><div className="extension-import-zone"><label htmlFor="extension-import-file"><Upload size={15} /> Package ZIP local<input id="extension-import-file" ref={fileRef} type="file" accept=".zip,application/zip" onChange={(event) => setFile(event.target.files?.[0] ?? null)} data-testid="extension-import-file" /></label><label>Série existante<select value={seriesId} onChange={(event) => setSeriesId(event.target.value)} data-testid="extension-import-series"><option value="">Nouvelle série</option>{result.map((series) => <option key={series.id} value={series.id}>{series.id}</option>)}</select></label><button type="button" className="action-primary" disabled={!file || loading} onClick={() => void importFile()} data-testid="extension-import-submit">Importer</button></div><div className="extension-update-zone"><label>Série à mettre à jour<select value={updateSeries} onChange={(event) => setUpdateSeries(event.target.value)}><option value="">Choisir une série</option>{result.map((series) => <option key={series.id} value={series.id}>{series.id}</option>)}</select></label><label htmlFor="extension-update-file">Nouvelle version<input id="extension-update-file" type="file" accept=".zip,application/zip" onChange={(event) => void updateFile(event.target.files?.[0] ?? null)} data-testid="extension-update-file" /></label></div><div className="extension-summary"><span><strong>{result.length}</strong> série(s)</span><span><strong>{totalVersions}</strong> version(s)</span><span><LockKeyhole size={13} /> Core authentifié</span></div><div className="extension-list">{result.length === 0 && <div className="catalog-empty" data-testid="extensions-empty"><FileArchive size={19} /><strong>Aucune extension importée</strong><span>Choisissez un ZIP local synthétique puis importez-le pour créer une série.</span></div>}{result.map((series) => <article className="extension-series-card" key={series.id} data-testid={`extension-series-${series.id}`}><div className="extension-series-heading"><div><strong>{series.id}</strong><span>{series.active_version_id ? `active ${series.active_version_id}` : "aucune version active"}</span></div><span>{series.versions.length} version(s)</span></div>{series.versions.map((version) => <VersionCard key={version.id} series={series} version={version} client={client} profiles={profiles} onAudit={onAudit} onReload={load} onAuthLost={onAuthLost} />)}</article>)}</div></section>;
}
