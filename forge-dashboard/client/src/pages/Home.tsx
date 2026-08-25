/**
 * ForgeLocal — Atelier de contrôle local.
 * Cette vue compose des données d’aperçu et, après une session locale valide, des projections Core redacted.
 * Philosophie : établi industriel local, signaux verts rares, jamais d’écriture ou de lancement runtime.
 */
import { useEffect, useMemo, useRef, useState } from "react";
import {
  Activity,
  Archive,
  Bell,
  Boxes,
  Check,
  ChevronDown,
  CircleHelp,
  Clock3,
  Command,
  Copy,
  Ellipsis,
  FileKey,
  FolderLock,
  Gauge,
  HardDrive,
  LayoutDashboard,
  LockKeyhole,
  MonitorPlay,
  MoreHorizontal,
  Network,
  Plus,
  Search,
  Settings2,
  ShieldCheck,
  SlidersHorizontal,
  Sparkles,
  Tag,
  Fingerprint,
  Cpu,
  TerminalSquare,
  X,
} from "lucide-react";
import { toast } from "sonner";
import { LocalCoreConnection, LocalCoreSnapshot, resolveLocalCoreBaseURL } from "@/components/LocalCoreConnection";
import { CoreProxyType } from "@/lib/coreWrite";
import { ProxyRegistry } from "@/components/ProxyRegistry";
import { BackupVault, BackupCreateAction } from "@/components/BackupVault";
import { AutomationPanel } from "@/components/AutomationPanel";
import { EnvironmentPanel } from "@/components/EnvironmentPanel";
import { RuntimePanel } from "@/components/RuntimePanel";
import { createCoreReadOnlyClient } from "@/lib/coreReadOnly";
import { CoreProxy, createCoreWriteClient, type CoreWriteClient } from "@/lib/coreWrite";

type ProfileStatus = "Prêt" | "Actif" | "À vérifier" | "Lecture seule";

type Profile = {
  id: string;
  name: string;
  group: string;
  runtime: string;
  proxy: string;
  lastSeen: string;
  status: ProfileStatus;
  fingerprint: string;
  tags: string[];
};

const profiles: Profile[] = [
  {
    id: "pfl_78c9f1",
    name: "Studio · Paris",
    group: "Création",
    runtime: "Camoufox · candidat",
    proxy: "Hérité du groupe",
    lastSeen: "il y a 8 min",
    status: "Actif",
    fingerprint: "fpr_9a6e•••c1",
    tags: ["France", "Design"],
  },
  {
    id: "pfl_624aa7",
    name: "Recherche · Lyon",
    group: "Recherche",
    runtime: "Runtime local 151",
    proxy: "Direct · coffre lié",
    lastSeen: "il y a 41 min",
    status: "Prêt",
    fingerprint: "fpr_31bb•••f4",
    tags: ["France", "Veille"],
  },
  {
    id: "pfl_f12a95",
    name: "Catalogue · Milan",
    group: "Commerce",
    runtime: "Runtime local 151",
    proxy: "À définir",
    lastSeen: "hier, 16:20",
    status: "À vérifier",
    fingerprint: "fpr_b74c•••93",
    tags: ["Italie", "Catalogue"],
  },
  {
    id: "pfl_d184e2",
    name: "Support · Bruxelles",
    group: "Opérations",
    runtime: "Camoufox · candidat",
    proxy: "Hérité du groupe",
    lastSeen: "hier, 10:08",
    status: "Prêt",
    fingerprint: "fpr_e159•••6b",
    tags: ["Belgique", "Support"],
  },
  {
    id: "pfl_0c771a",
    name: "Laboratoire · Test",
    group: "Recherche",
    runtime: "Runtime local 151",
    proxy: "Local direct",
    lastSeen: "12 août, 11:15",
    status: "Prêt",
    fingerprint: "fpr_87df•••0a",
    tags: ["QA", "Isolé"],
  },
];

const navSections = [
  {
    label: "Atelier",
    items: [
      { icon: LayoutDashboard, label: "Vue d’ensemble", active: true },
      { icon: Boxes, label: "Profils", count: "24" },
      { icon: Tag, label: "Groupes" },
    ],
  },
  {
    label: "Contrôles",
    items: [
      { icon: MonitorPlay, label: "Runtimes" },
      { icon: Network, label: "Proxys" },
      { icon: Archive, label: "Sauvegardes" },
      { icon: Fingerprint, label: "Identité navigateur" },
      { icon: Cpu, label: "Runtime qualifié" },
      { icon: TerminalSquare, label: "Automation locale" },
    ],
  },
];

const statusClasses: Record<ProfileStatus, string> = {
  Actif: "status-live",
  Prêt: "status-ready",
  "À vérifier": "status-review",
  "Lecture seule": "status-readonly",
};

const lifecycleTooltip = (state: string) => state === "archived" ? "Réouvrir ce profil dormant côté Core" : "Archiver ce profil côté Core";

function Initials({ name }: { name: string }) {
  const initials = name
    .split("·")
    .map((part) => part.trim().charAt(0))
    .join("")
    .slice(0, 2);
  return <span className="profile-initials" aria-hidden="true">{initials}</span>;
}

export default function Home() {
  const [query, setQuery] = useState("");
  const [group, setGroup] = useState("Tous les groupes");
  const [status, setStatus] = useState<"Tous" | ProfileStatus>("Tous");
  const [activeId, setActiveId] = useState(profiles[0].id);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [navLabel, setNavLabel] = useState("Vue d’ensemble");
  const [coreSnapshot, setCoreSnapshot] = useState<LocalCoreSnapshot | null>(null);
  const [coreWrite, setCoreWrite] = useState<{ token: string; version: number } | null>(null);
  const writeClientRef = useRef<CoreWriteClient | null>(null);
  const [writePending, setWritePending] = useState<Record<string, boolean>>({});
  const [selectedLifecycle, setSelectedLifecycle] = useState<Record<string, "active" | "archived" | "quarantined">>({});
  const [refreshNonce, setRefreshNonce] = useState(0);
  const [registryProxies, setRegistryProxies] = useState<CoreProxy[]>([]);
  const [assignedProxyIds, setAssignedProxyIds] = useState<Record<string, string>>({});
  const [profileIds, setProfileIds] = useState<string[]>([]);
  const [createBackupPending, setCreateBackupPending] = useState<string | null>(null);
  /**
   * Formulaire de création proxy persistant : le contenu saisi survit aux
   * démontages du panneau ProxyRegistry (expiration de la session lecture
   * seule suivie d'un re-link du jeton d'administration). Mémoire seule.
   */
  const proxyFormRef = useRef({ name: "", type: "http" as CoreProxyType, host: "", port: "", region: "", secretRef: "" });

  // T09/T14 — Le client d'écritures est instancié UNE SEULE FOIS : les panneaux
  // (RuntimePanel, EnvironmentPanel…) déclenchent leur chargement à la demande
  // dans un useEffect porté par client.isConnected(). Si le client était recréé
  // à chaque re-render de cet effet, la référence instable ferait échouer ces
  // chargements (connexion perdue en cours de requête) et multiplierait les
  // requêtes fantômes vers le Core local.
  const writeClient = useMemo(() => createCoreWriteClient(resolveLocalCoreBaseURL()), []);
  useEffect(() => {
    if (!coreWrite || !coreSnapshot) return;
    const client = writeClient;
    client.bind(coreWrite.token);
    writeClientRef.current = client;
    // Découvrir le cycle de vie de chaque profil Core affiché (champ redacted uniquement côté écriture).
    void Promise.allSettled(coreSnapshot.profiles.map(async (profile) => {
      try {
        const { data } = await client.getProfile(profile.id);
        setSelectedLifecycle((previous) => ({ ...previous, [profile.id]: data.lifecycle_state || "active" }));
      } catch {
        setSelectedLifecycle((previous) => ({ ...previous, [profile.id]: "active" }));
      }
    }));
    // T10 : charger le référentiel proxy du Core et l'affectation de chaque profil une fois le contrôle local lié.
    client.proxies.listProxies()
      .then(({ data }) => setRegistryProxies(data))
      .catch(() => setRegistryProxies([]));
    // T11 : recenser les identifiants des profils Core (pour l'unicité des cibles de restauration).
    void Promise.allSettled(coreSnapshot.profiles.map(async (profile) => client.getProfile(profile.id)))
      .then((results) => setProfileIds(results.filter((result) => result.status === "fulfilled").map((result) => (result as PromiseFulfilledResult<{ data: { id: string } }>).value.data.id)));
    client.backups.listBackups()
      .then(({ data }) => void data)
      .catch(() => {});
    void Promise.allSettled(coreSnapshot.profiles.map(async (profile) => {
      try {
        const { data } = await client.getProfile(profile.id);
        const proxyId = (data as unknown as { proxy_id?: string }).proxy_id;
        if (proxyId) setAssignedProxyIds((previous) => ({ ...previous, [profile.id]: proxyId }));
      } catch {
        // Le champ proxy_id est facultatif ; l'absence n'empêche pas le reste du contrat.
      }
    }));
    return () => { client.unbind(); writeClientRef.current = null; };
  }, [coreWrite, coreSnapshot, writeClient]);

  const refreshCoreSnapshot = async () => {
    if (!coreSnapshot) return;
    try {
      const client = createCoreReadOnlyClient(resolveLocalCoreBaseURL());
      const [summary, page, groups, runtimes] = await Promise.all([
        client.getSummary(),
        client.listProfiles({ limit: 250 }),
        client.listGroups({ limit: 100 }),
        client.listRuntimes({ limit: 100 }),
      ]);
      setCoreSnapshot({ summary, profiles: page.data, groups: groups.data, runtimes: runtimes.data, expiresAt: coreSnapshot.expiresAt });
      setRefreshNonce((nonce) => nonce + 1);
    } catch (error) {
      toast.error("Actualisation Core impossible", { description: String(error) });
    }
  };

  const runWrite = async (label: string, id: string, action: () => Promise<unknown>) => {
    const client = writeClientRef.current;
    if (!client?.isConnected()) {
      toast.error("Contrôle local requis", { description: "Reliez le jeton d’administration dans le panneau État du Core." });
      return;
    }
    setWritePending((previous) => ({ ...previous, [`${label}-${id}`]: true }));
    try {
      await action();
      await refreshCoreSnapshot();
      toast.success("Action appliquée au Core", { description: `Opération ${label} validée par le Core Go local.` });
    } catch (error) {
      const message = String(error);
      if (message.includes("CORE_ADMIN_EXPIRED")) {
        setCoreWrite(null);
        toast.error("Le contrôle local a expiré", { description: "Le jeton d’administration a dépassé sa durée de vie ; reliez un nouveau jeton." });
      } else if (message.includes("CORE_ADMIN_REVOKED")) {
        setCoreWrite(null);
        toast.error("Le contrôle local a été révoqué", { description: "Le Core a invalidé ce jeton d’administration." });
      } else if (message.includes("CORE_ADMIN_UNAUTHORIZED") || message.includes("CORE_ERROR_LOOPBACK")) {
        setCoreWrite(null);
        toast.error("Le contrôle local a été retiré", { description: "Le jeton d’administration n’est plus accepté par le Core." });
      } else if (message.includes("CORE_ERROR_")) {
        toast.error("Le Core a refusé l’opération", { description: message.replace("CORE_ERROR_", "") });
      } else {
        toast.error("Action refusée par le Core", { description: message });
      }
    } finally {
      setWritePending((previous) => ({ ...previous, [`${label}-${id}`]: false }));
    }
  };

  const displayedProfiles = useMemo<Profile[]>(() => coreSnapshot
    ? coreSnapshot.profiles.map((profile) => ({
      id: profile.id,
      name: profile.name,
      group: profile.group || "Sans groupe",
      runtime: profile.runtime_id || "Runtime non déclaré",
      proxy: profile.proxy_configured ? "Configuration liée" : "Non configuré",
      lastSeen: profile.last_used || "—",
      status: "Lecture seule",
      fingerprint: "non exposée",
      tags: profile.tags,
    }))
    : profiles, [coreSnapshot]);

  const visibleProfiles = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase("fr-FR");
    return displayedProfiles.filter((profile) => {
      const matchesQuery = !normalizedQuery || [profile.name, profile.group, ...profile.tags]
        .join(" ")
        .toLocaleLowerCase("fr-FR")
        .includes(normalizedQuery);
      const matchesGroup = group === "Tous les groupes" || profile.group === group;
      const matchesStatus = status === "Tous" || profile.status === status;
      return matchesQuery && matchesGroup && matchesStatus;
    });
  }, [displayedProfiles, group, query, status]);

  const selectedProfile = displayedProfiles.find((profile) => profile.id === activeId) ?? displayedProfiles[0] ?? profiles[0];
  const coreGroups = coreSnapshot?.groups ?? [];
  const coreRuntimes = coreSnapshot?.runtimes ?? [];
  const liveProfiles = coreSnapshot?.profiles ?? [];

  const unavailable = (label: string) => {
    setNavLabel(label);
    toast.info(`${label} est prêt pour l’intégration au Core`, {
      description: "Cette maquette locale ne déclenche aucune écriture ni lancement de runtime.",
    });
  };

  return (
    <main className="forgelocal-shell">
      <aside className="sidebar" aria-label="Navigation principale">
        <div className="brand-lockup">
          <span className="brand-stamp" aria-hidden="true"><span className="brand-mark brand-mark-fallback">FL</span></span>
          <div>
            <span className="brand-name">ForgeLocal</span>
            <span className="brand-subtitle">Control desk</span>
          </div>
        </div>

        <div className="workspace-select" role="button" tabIndex={0} onClick={() => unavailable("Espaces de travail")} onKeyDown={(event) => event.key === "Enter" && unavailable("Espaces de travail")}>
          <span className="workspace-orb"><HardDrive size={15} /></span>
          <span><strong>Cette machine</strong><small>atelier-local</small></span>
          <ChevronDown size={16} aria-hidden="true" />
        </div>

        <nav className="sidebar-nav">
          {navSections.map((section) => (
            <div className="nav-section" key={section.label}>
              <p className="nav-eyebrow">{section.label}</p>
              {section.items.map((item) => {
                const Icon = item.icon;
                const isCurrent = item.label === navLabel || ("active" in item && item.active && navLabel === "Vue d’ensemble");
                return (
                  <button
                    className={`nav-item ${isCurrent ? "nav-item-active" : ""}`}
                    key={item.label}
                    onClick={() => ("active" in item && item.active) || item.label === "Groupes" || item.label === "Runtimes" || item.label === "Sauvegardes" || item.label === "Proxys" || item.label === "Identité navigateur" || item.label === "Runtime qualifié" || item.label === "Automation locale" ? setNavLabel(item.label) : unavailable(item.label)}
                    type="button"
                  >
                    <Icon size={17} strokeWidth={1.8} />
                    <span>{item.label}</span>
                    {item.count && <em>{item.count}</em>}
                  </button>
                );
              })}
            </div>
          ))}
        </nav>

        <div className="sidebar-bottom">
          <button className="nav-item" type="button" onClick={() => unavailable("Journal d’audit")}><TerminalSquare size={17} /><span>Journal d’audit</span></button>
          <button className="nav-item" type="button" onClick={() => unavailable("Réglages")}><Settings2 size={17} /><span>Réglages</span></button>
          <div className="local-proof">
            <span className="proof-icon"><ShieldCheck size={15} /></span>
            <p><strong>Local-first</strong><span>{coreSnapshot ? "Lecture Core active" : "Core non connecté"}</span></p>
          </div>
        </div>
      </aside>

      <section className="workbench">
        <header className="topbar">
          <div className="topbar-crumb"><Command size={15} /><span>ForgeLocal</span><span className="crumb-slash">/</span><strong>{navLabel}</strong><span className={`demo-badge ${coreSnapshot ? "core-live-badge" : ""}`} aria-live="polite"><i /> {coreSnapshot ? "Lecture Core locale" : "Démonstration locale"} <b>·</b> {coreSnapshot ? (writeClientRef.current?.isConnected() ? "écritures locales" : "aucune écriture") : "Core non connecté"}</span></div>
          <div className="topbar-actions">
            <button className="icon-button" aria-label="Aide" onClick={() => unavailable("Aide")} type="button"><CircleHelp size={18} /></button>
            <button className="icon-button notification" aria-label="Notifications" onClick={() => unavailable("Notifications")} type="button"><Bell size={18} /><i /></button>
            <div className="operator"><span className="operator-avatar">MA</span><span><strong>Mainteneur</strong><small>local</small></span></div>
          </div>
        </header>

        <div className="content-scroll">
          <section className="workspace-intro instrument-plate plate-hero">
            <span className="plate-code">STR / LOCAL / 001</span>
            <div>
              <p className="section-kicker"><span /> Vue d’ensemble locale</p>
              <h1>Vos profils restent<br /><em>sur cette machine.</em></h1>
              <p className="intro-copy">Préparez, isolez et suivez vos espaces navigateur sans ajouter un second plan de contrôle.</p>
            </div>
            <div className="core-status instrument-plate">
              <span className="plate-code">CORE / RO / 01</span>
              <div className="core-status-top"><span className="pulsing-dot" /><span>État du Core</span></div>
              <strong>{coreSnapshot ? "Lecture locale active" : "En attente de connexion"}</strong>
              <small>{coreSnapshot ? "Les données affichées sont redacted et le token est uniquement en mémoire." : "Le panneau ne lit aucune donnée locale avant le raccordement de l’API authentifiée."}</small>
              <LocalCoreConnection
                onConnected={setCoreSnapshot}
                onDisconnected={() => { setCoreSnapshot(null); setCoreWrite(null); }}
                onWriteConnected={(token) => setCoreWrite({ token, version: Date.now() })}
                onWriteDisconnected={() => { setCoreWrite(null); setSelectedLifecycle({}); setRegistryProxies([]); setAssignedProxyIds({}); setProfileIds([]); }}
              />
            </div>
          </section>

          <section className="metric-strip" aria-label="Indicateurs locaux">
            <article className="metric-card metric-primary instrument-plate">{!coreSnapshot && <span className="demo-data-tag" aria-label="Données de démonstration">Démo</span>}<span className="plate-code">PLT / 01</span><div className="metric-label"><Boxes size={16} /> Profils locaux</div><strong>{coreSnapshot ? coreSnapshot.summary.data.profiles : "24"}</strong><span><Check size={13} /> {coreSnapshot ? "lecture redacted" : "18 prêts à lancer"}</span><i className="metric-coordinates">PFL / {String(coreSnapshot?.summary.data.profiles ?? 24).padStart(3, "0")}</i></article>
            <article className="metric-card instrument-plate">{!coreSnapshot && <span className="demo-data-tag" aria-label="Données de démonstration">Démo</span>}<span className="plate-code">PLT / 02</span><div className="metric-label"><MonitorPlay size={16} /> Runtimes</div><strong>{coreSnapshot ? coreSnapshot.summary.data.runtimes : "02"}</strong><span>{coreSnapshot ? "inventaire Core" : "1 validé · 1 candidat"}</span><i className="metric-coordinates">RUN / {String(coreSnapshot?.summary.data.runtimes ?? 2).padStart(3, "0")}</i></article>
            <article className="metric-card instrument-plate"><span className="demo-data-tag" aria-label="Données de démonstration">Démo</span><span className="plate-code">PLT / 03</span><div className="metric-label"><FolderLock size={16} /> Coffre système</div><strong>—</strong><span>Contrat en attente</span><i className="metric-coordinates">VLT / PEND</i></article>
            <article className="metric-card instrument-plate"><span className="demo-data-tag" aria-label="Données de démonstration">Démo</span><span className="plate-code">PLT / 04</span><div className="metric-label"><Gauge size={16} /> Ressources hôte</div><strong>62<span className="unit">%</span></strong><span>estimation visuelle</span><i className="metric-coordinates">HST / LIVE</i></article>
          </section>

          <section className="profile-zone">
            <div className="profile-zone-header">
              <div><p className="section-kicker"><span /> Registre des profils</p><h2>Postes isolés {!coreSnapshot && <span className="demo-title-tag">démo</span>}</h2><p>{visibleProfiles.length} affiché{visibleProfiles.length !== 1 ? "s" : ""} {coreSnapshot ? "depuis le Core, en lecture seule." : "dans cette maquette de travail, non issus du Core."}</p></div>
              <button className="create-profile" type="button" onClick={() => setDialogOpen(true)}><Plus size={17} /> Préparer un profil</button>
            </div>

            {coreWrite && <ProxyRegistry client={writeClientRef.current!} proxies={registryProxies} onChange={setRegistryProxies} selectedProfileId={selectedProfile?.id} assignedProxyId={assignedProxyIds[selectedProfile?.id ?? ""]} formRef={proxyFormRef} />}
            {coreWrite && navLabel === "Sauvegardes" && <BackupVault client={writeClientRef.current!} profileIds={profileIds} />}
            {coreWrite && navLabel === "Identité navigateur" && <EnvironmentPanel client={writeClientRef.current!} profileIds={profileIds} selectedProfileId={selectedProfile?.id} />}
            {coreWrite && navLabel === "Runtime qualifié" && <RuntimePanel client={writeClientRef.current!} />}
            {coreWrite && navLabel === "Automation locale" && <AutomationPanel client={writeClientRef.current!} profileIds={profileIds} selectedProfileId={selectedProfile?.id} />}
            <div className="profile-toolbar">
              <label className="search-field"><Search size={17} /><span className="sr-only">Rechercher un profil</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Rechercher un profil, un groupe, un tag…" /></label>
              <label className="select-field"><span className="sr-only">Filtrer par groupe</span><select value={group} onChange={(event) => setGroup(event.target.value)}><option>Tous les groupes</option><option>Création</option><option>Recherche</option><option>Commerce</option><option>Opérations</option></select><ChevronDown size={15} /></label>
              <div className="status-filters" aria-label="Filtrer par statut">
                {(["Tous", "Prêt", "Actif", "À vérifier"] as const).map((filter) => <button type="button" key={filter} onClick={() => setStatus(filter)} className={status === filter ? "filter-active" : ""}>{filter}</button>)}
              </div>
              <button className="filter-icon" type="button" aria-label="Plus de filtres" onClick={() => unavailable("Filtres avancés")}><SlidersHorizontal size={17} /></button>
            </div>

            <div className="profile-table-wrap instrument-plate"><span className="plate-code">REG / PFL / 01</span>
              <div className="profile-table-head"><span>Profil</span><span>Runtime</span><span>Proxy</span><span>Dernière activité</span><span>État</span><span /></div>
              <div className="profile-rows">
                {visibleProfiles.map((profile) => (
                  <article className={`profile-row ${selectedProfile.id === profile.id ? "profile-row-selected" : ""}`} key={profile.id} onClick={() => setActiveId(profile.id)}>
                    <div className="profile-name-cell"><Initials name={profile.name} /><div><strong>{profile.name} {!coreSnapshot && <em className="inline-demo">démo</em>}</strong><span>{profile.group} <b>·</b> {profile.id}</span></div></div>
                    <div className="runtime-cell"><MonitorPlay size={14} /><span>{profile.runtime}</span>{profile.runtime.includes("candidat") && <i className="candidate-lock" title="Runtime candidat non lançable">candidat</i>}</div>
                    <div className="proxy-cell"><Network size={14} /><span>{profile.proxy}</span></div>
                    <div className="last-seen"><Clock3 size={14} /><span>{profile.lastSeen}</span></div>
                    {coreSnapshot && writeClientRef.current?.isConnected() && <BackupCreateAction profileId={profile.id} busy={createBackupPending === profile.id} disabled={Boolean(createBackupPending)} onSave={() => {
                      if (!writeClientRef.current?.isConnected()) return;
                      setCreateBackupPending(profile.id);
                      void runWrite("sauvegarde", profile.id, () => writeClientRef.current!.backups.createBackup(profile.id)).finally(() => setCreateBackupPending(null));
                    }} />}
                    <div><span className={`status-pill ${statusClasses[profile.status]}`}><i />{profile.status}</span></div>
                    <button className="row-menu row-menu-primary" aria-label={`Actions pour ${profile.name}`} type="button" data-testid={`row-menu-${profile.id}`} onClick={(event) => {
                      event.stopPropagation();
                      if (!coreSnapshot) {
                        unavailable(`Actions de ${profile.name}`);
                        return;
                      }
                      const lifecycle = selectedLifecycle[profile.id] ?? "active";
                      const isArchived = lifecycle === "archived";
                      if (window.confirm(isArchived ? `Réouvrir ${profile.name} côté Core ?` : `Archiver ${profile.name} côté Core ?`)) {
                        void runWrite(isArchived ? "réouverture" : "archivage", profile.id, () => (isArchived ? writeClientRef.current!.reopenProfile(profile.id) : writeClientRef.current!.archiveProfile(profile.id)));
                        setSelectedLifecycle((previous) => ({ ...previous, [profile.id]: isArchived ? "active" : "archived" }));
                      }
                    }} aria-description={lifecycleTooltip(selectedLifecycle[profile.id] ?? "active")}><MoreHorizontal size={18} /></button>
                  </article>
                ))}
                {visibleProfiles.length === 0 && <div className="empty-profiles"><Search size={22} /><strong>Aucun profil ne correspond aux filtres</strong><span>Essayez de retirer un filtre ou de rechercher un autre terme.</span></div>}
              </div>
            </div>
          </section>

          <section className="catalog-zone" aria-label="Catalogues Core lecture seule">
            <div className="catalog-zone-header">
              <div><p className="section-kicker"><span /> Inventaires du Core</p><h2>Groupes et runtimes <span className={coreSnapshot ? "catalog-live-tag" : "demo-title-tag"}>{coreSnapshot ? "lecture réelle" : "démo"}</span></h2><p>{coreSnapshot ? "Projections SQLite et registre runtime redacted ; aucune commande ne peut être envoyée." : "Connectez le Core local pour remplacer ces surfaces par les inventaires redacted."}</p></div>
              <span className="catalog-readonly"><LockKeyhole size={14} /> Lecture seule</span>
            </div>

            <div className="catalog-grid">
              <section className="catalog-panel instrument-plate" data-testid="core-groups-panel"><span className="plate-code">CAT / GRP / RO</span>
                <div className="catalog-panel-heading"><div><Tag size={17} /><h3>Groupes</h3></div><span data-testid="core-groups-count">{coreSnapshot ? coreGroups.length : "—"}</span></div>
                {coreSnapshot && coreGroups.length === 0 && <div className="catalog-empty" data-testid="core-groups-empty"><Tag size={18} /><strong>Aucun groupe redacted</strong><span>Le Core ne retourne aucun groupe pour cette instance.</span></div>}
                {!coreSnapshot && <div className="catalog-empty"><Tag size={18} /><strong>Inventaire en attente</strong><span>Les données de démonstration ne remplacent pas le Core.</span></div>}
                {coreSnapshot && coreGroups.length > 0 && <div className="catalog-list" data-testid="core-groups-list">{coreGroups.map((item) => <article key={item.id || item.name} className="catalog-row"><div><strong>{item.name}</strong><span>{item.profile_count} profil{item.profile_count !== 1 ? "s" : ""} · proxy {item.proxy_configured ? "configuré" : "non configuré"}</span></div><span className="catalog-mode">{item.proxy_mode || "direct"}</span></article>)}</div>}
              </section>

              <section className="catalog-panel instrument-plate" data-testid="core-runtimes-panel"><span className="plate-code">CAT / RUN / RO</span>
                <div className="catalog-panel-heading"><div><MonitorPlay size={17} /><h3>Runtimes</h3></div><span data-testid="core-runtimes-count">{coreSnapshot ? coreRuntimes.length : "—"}</span></div>
                {coreSnapshot && coreRuntimes.length === 0 && <div className="catalog-empty" data-testid="core-runtimes-empty"><MonitorPlay size={18} /><strong>Aucun runtime redacted</strong><span>Le Core ne retourne aucun runtime pour cette instance.</span></div>}
                {!coreSnapshot && <div className="catalog-empty"><MonitorPlay size={18} /><strong>Inventaire en attente</strong><span>Le registre réel est chargé après la connexion locale.</span></div>}
                {coreSnapshot && coreRuntimes.length > 0 && <div className="catalog-list" data-testid="core-runtimes-list">{coreRuntimes.map((item) => <article key={item.id} className="catalog-row"><div><strong>{item.display_name}</strong><span>{[item.version, item.architecture, item.status].filter(Boolean).join(" · ") || "État non exposé"}</span></div><span className={item.candidate ? "candidate-lock" : "catalog-mode"}>{item.candidate ? "candidat non lançable" : item.launchable ? "catalogué" : "désactivé"}</span></article>)}</div>}
              </section>
            </div>
          </section>
        </div>
      </section>

      <aside className="observation-rail" aria-label="Observabilité locale">
        <div className="rail-top"><p className="rail-eyebrow">Profil sélectionné</p><button className="icon-button" type="button" aria-label="Copier l’identifiant" onClick={() => { navigator.clipboard?.writeText(selectedProfile.id); toast.success("Identifiant copié", { description: "Aucune donnée de coffre n’est incluse." }); }}><Copy size={16} /></button></div>
        <div className="selected-profile instrument-plate"><span className="plate-code">OBS / 14</span><div className="selected-header"><Initials name={selectedProfile.name} /><span className={`status-pill ${statusClasses[selectedProfile.status]}`}><i />{selectedProfile.status}</span></div><p className="selected-demo"><Sparkles size={11} /> {coreSnapshot ? "Données Core redacted" : "Données de démonstration"}</p><h3>{selectedProfile.name}</h3><p>{selectedProfile.group} <b>·</b> {selectedProfile.id}</p><div className="tag-list">{selectedProfile.tags.map((tag) => <span key={tag}>{tag}{coreSnapshot && writeClientRef.current?.isConnected() ? <button type="button" aria-label={`Retirer le tag ${tag}`} onClick={async (event) => { event.stopPropagation(); void runWrite("retrait de tag", selectedProfile.id, () => writeClientRef.current!.removeProfileTag(selectedProfile.id, tag)); }} className="tag-remove"><X size={10} /></button> : null}</span>)}</div>
        <div className="rail-tag-add"><label><input id="rail-new-tag" placeholder="Ajouter un tag…" onKeyDown={async (event) => {
          if (event.key === "Enter" && coreSnapshot) {
            const input = event.currentTarget;
            const tag = input.value.trim().toLocaleLowerCase("fr-FR");
            if (!tag) return;
            input.value = "";
            void runWrite("ajout de tag", selectedProfile.id, () => writeClientRef.current!.addProfileTag(selectedProfile.id, tag));
          }
        }} /></label>{!coreSnapshot && <small>Connectez le Core pour gérer les tags.</small>}</div></div>

        <div className="rail-action-grid" aria-describedby="core-action-help"><button type="button" disabled title="Action disponible après connexion au Core local"><MonitorPlay size={16} /><span>Lancer</span></button><button type="button" disabled title="Action disponible après connexion au Core local"><LockKeyhole size={16} /><span>Isoler</span></button></div>
        <p className="core-action-help" id="core-action-help"><LockKeyhole size={12} /> Action disponible après connexion au Core local.</p>

        <div className="detail-list"><div><span>Runtime</span><strong>{selectedProfile.runtime}</strong></div>{selectedProfile.runtime.includes("candidat") && <p className="candidate-note"><LockKeyhole size={12} /> Runtime candidat non lançable avant qualification indépendante.</p>}<div><span>Proxy</span><strong>{selectedProfile.proxy}</strong></div><div><span>Empreinte</span><strong>{selectedProfile.fingerprint}</strong></div><div><span>Chemin</span><strong>non chargé</strong></div></div>

        <section className="vault-card instrument-plate"><span className="plate-code">VLT / SEALED</span><div className="vault-art-fallback" aria-hidden="true" /><div className="vault-overlay" /><div className="vault-copy"><span><LockKeyhole size={14} /> Coffre système</span><strong>Références, jamais secrets.</strong><p>La connexion au coffre sera vérifiée par le Core.</p></div></section>

        <section className="activity-card instrument-plate"><span className="plate-code">LOG / LOCAL</span><div className="activity-heading"><div><p className="rail-eyebrow">Trace locale</p><h3>Derniers repères</h3></div><Activity size={18} /></div><ol><li><span className="timeline-mark good" /><div><strong>Préimage vérifiée</strong><small>import · 14:22</small></div></li><li><span className="timeline-mark neutral" /><div><strong>Runtime candidat catalogué</strong><small>runtime · 12:10</small></div></li><li><span className="timeline-mark neutral" /><div><strong>Maquette API initialisée</strong><small>ui · maintenant</small></div></li></ol></section>

        <div className="rail-footer"><FileKey size={15} /><span>Rien n’est envoyé au cloud.</span></div>
      </aside>

      {dialogOpen && (
        <div className="dialog-backdrop" role="presentation" onMouseDown={() => setDialogOpen(false)}>
          <section className="prepare-dialog" role="dialog" aria-modal="true" aria-labelledby="prepare-title" onMouseDown={(event) => event.stopPropagation()}>
            <button className="dialog-close" type="button" onClick={() => setDialogOpen(false)} aria-label="Fermer"><X size={18} /></button>
            <p className="section-kicker"><span /> Parcours contrôlé</p><h2 id="prepare-title">Préparer un profil</h2>
            <p>La création est validée par l’API authentifiée du Core Go : nom, runtime et tags sont contrôlés côté serveur.</p>
            <label>Nom de travail<input id="create-profile-name" data-testid="create-profile-name" placeholder="Ex. Veille · Amsterdam" /></label>
            <label>Runtime<select id="create-profile-runtime" data-testid="create-profile-runtime" defaultValue=""><option value="" disabled>Choisir un runtime</option>{coreRuntimes.map((runtime) => <option key={runtime.id} value={runtime.id} disabled={!runtime.launchable}>{runtime.display_name}{runtime.candidate ? " · candidat non lançable" : ""}</option>)}</select></label>
            <label>Groupe<select id="create-profile-group" defaultValue=""><option value="" disabled>Choisir un groupe</option>{coreGroups.length > 0 ? coreGroups.map((item) => <option key={item.id || item.name} value={item.name}>{item.name}</option>) : <option value="">Sans groupe</option>}</select></label>
            <label>Tags (séparés par des virgules)<input id="create-profile-tags" data-testid="create-profile-tags" placeholder="france, veille" /></label>
            <div className="dialog-note"><Sparkles size={16} /><span>Les identifiants proxy et secrets ne sont jamais demandés dans ce panneau.</span></div>
            <div className="dialog-actions"><button type="button" className="secondary-action" onClick={() => setDialogOpen(false)}>Annuler</button><button type="button" className="create-profile" disabled={!coreSnapshot} onClick={async () => { const name = (document.getElementById("create-profile-name") as HTMLInputElement)?.value.trim(); const runtimeId = (document.getElementById("create-profile-runtime") as HTMLSelectElement)?.value; const groupName = (document.getElementById("create-profile-group") as HTMLSelectElement)?.value; const tags = (document.getElementById("create-profile-tags") as HTMLInputElement)?.value.split(",").map((value) => value.trim().toLocaleLowerCase("fr-FR")).filter(Boolean); if (!name || !runtimeId) { toast.error("Nom et runtime requis", { description: "Le Core refuse les profils incomplets." }); return; } setDialogOpen(false); await runWrite("création", "new", () => writeClientRef.current!.createProfile({ name, runtime_id: runtimeId, group: groupName || undefined, tags: tags.length ? tags : undefined })); }} title={!coreSnapshot ? "Reliez la lecture Core puis le contrôle local pour créer un profil." : undefined}>Créer via le Core <LockKeyhole size={15} /></button></div>
            <p className="dialog-core-note"><LockKeyhole size={12} /> Aucune donnée ne transite hors du loopback : le Core valide le nom, le runtime et les tags côté serveur.</p>
          </section>
        </div>
      )}
    </main>
  );
}
