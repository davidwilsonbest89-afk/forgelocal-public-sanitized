# Cahier des charges complet — ForgeLocal (version 3.9.7)

**Identifiant :** `FL-CDC-3.9.7-20260815` · **Date :** 15 août 2026
**Statut :** document directeur du lot actuel, approuvé par le propriétaire `@boucheriechefimane-cmd`
**Précédence :** en cas de conflit avec la roadmap v3.8 ou un document antérieur, `v3.9.7` prévaut pour le développement actuel ; le prompt source exact est `PROMPT_FORGELOCAL_IMPLEMENTATION_REAL_v3.9.7.md` et l’addendum produit `v3.9.1`.

Ce document reproduit intégralement le cahier des charges fourni par le propriétaire le 15 août 2026 et est conservé tel quel dans le référentiel documentaire ForgeLocal. Les statuts de release (`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, pilote BACK-01 suspendu, cinq gates publics) sont maintenus et aucun lot ne peut présenter une progression produit comme une autorisation de release.

## Points directeurs retenus pour le lot

1. **Architecture fondatrice :** BrowseForge/Core Go unique control plane et unique écrivain de l’état métier ; React client mémoire seule ; SQLite base métier ; SystemVault pour les secrets ; Chromium qualifié runtime de référence ; Camoufox `candidate / non lançable`.
2. **Provenance :** le registre de provenance (`docs/component-rights-register.json`) reste la source machine-readable ; les gates `PROV-01` à `PROV-07` s’appliquent à toute dépendance tierce ; le statut Camoflox après T07 est `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION` (PROV-05 et PROV-07 différés à T08).
3. **T08 autorisé :** réimplémentation Go indépendante d’un module hashé à la fois, premier module `lib/concurrency.js` (décision `reimplementer` enregistrée avant code, exclusions explicites), sans port runtime réel, sans Camoufox, sans lancement navigateur.
4. **Périmètre exclusif :** vol de données, credential/cookie theft, fraude, CAPTCHA solving/bypass, contournement anti-bot/anti-fraude, accès non autorisé et contournement d’un mécanisme de sécurité sont définitivement exclus.
5. **Maturité :** une fonctionnalité n’est livrée que par preuves intégrées — code, tests positifs/négatifs, concurrence (`-race`), critères d’acceptation, preuves redacted et démonstration locale ; un code `0` sans test sélectionné ou un scan sans périmètre vaut échec de validation.

## Cohérence avec les registres

Le cahier reprend fidèlement les statuts enregistrés : T07 clôturé pour réimplémentation Go sélective, T08 autorisé à périmètre restreint (queue bornée, limite globale, sérialisation par profil, timeout, annulation `context`, cleanup idempotent, reprise après crash, audit redacted), T09–T21 à venir, lots R00–R03 distincts des succès produit. Les critères T08 AC-CAMO-01 à AC-CAMO-05 coïncident avec la spécification `docs/T08_CONCURRENCY_SPEC.md`.

---

# Annexe — Texte intégral du cahier des charges FL-CDC-3.9.7-20260815

# Cahier des charges complet — ForgeLocal

| Métadonnée | Valeur |
| --- | --- |
| **Identifiant** | `FL-CDC-3.9.7-20260815` |
| **Version** | 3.9.7 — Application du prompt d’implémentation réelle et audit option par option |
| **Date** | 15 août 2026 |
| **Statut produit** | Développement local autorisé dans le périmètre décrit ; release publique bloquée |
| **Architecture fondatrice** | BrowseForge/Core Go unique, React client, SQLite métier, stockage navigateur isolé, coffre système |
| **Code ForgeLocal visé** | MIT ou Apache-2.0 |
| **Périmètre release actuel** | Ubuntu 24.04.4 LTS amd64, candidat BACK-01 distinct et gelé |

> **But du document.** Ce cahier conserve la vision produit ForgeLocal de v3.8, mais définit le périmètre d’exécution strict du MVP local. En cas de conflit entre une fonction de roadmap v3.8 et ce document, **v3.9.7 prévaut pour le développement actuel**. Les fonctions reportées restent documentées comme roadmap ; elles ne doivent pas être codées dans ce lot.

### 0.1 Règle de périmètre MVP strict

Le développeur doit suivre le prompt source exact `PROMPT_FORGELOCAL_IMPLEMENTATION_REAL_v3.9.7.md` et l’addendum produit v3.9.1, ne doit pas élargir le périmètre, ne doit pas remplacer l’architecture et ne doit pas introduire les fonctions P2 exclues du MVP.

BrowseForge/Core Go reste l’unique control plane et l’unique écrivain de l’état métier. React reste uniquement le client. SQLite reste la base métier. Les secrets restent dans SystemVault. Chromium qualifié reste le runtime de référence et Camoufox reste `candidate / non lançable`.

### 0.2 Fonctions P2 et vision future conservées mais reportées

| Fonction de la vision v3.8 | Statut dans v3.9 |
| --- | --- |
| Cloud, Cloud Launch, Web App distante et Profile Sync Cloud | Reportés ; aucun code MVP |
| Profile Sharing, RBAC, 2FA et multi-user distant | Reportés ; aucun code MVP |
| Android, Enterprise et gouvernance distante | Reportés ; aucun code MVP |
| Marketplace, Proxy Devices et services proxy commerciaux | Reportés ou exclus du MVP |
| Cookie Bot, Action Synchronizer et Behavioral Simulation | Capacités produit acceptées par l’addendum ; implémentation soumise aux contrats, tests et restrictions du §16 |
| Vol de données, credential/cookie theft, fraude, CAPTCHA solving/bypass, contournement anti-bot/anti-fraude, accès non autorisé et contournement d’un mécanisme de sécurité | Exclus et non implémentés |

Le report ne supprime pas la vision v3.8 ; il empêche uniquement son implémentation pendant le lot MVP strict.

### 0.4 Réconciliation de l’addendum v3.9.1

L’addendum v3.9.1 valide comme capacités produit ForgeLocal les neuf domaines suivants : **Fingerprint spoofing configurable, New Fingerprint, Canvas Noise, WebGL Noise/Spoofing, Audio Noise, Hardware Spoofing, Behavioral Simulation, Cookie Bot et Action Synchronizer**.

Ces capacités restent soumises aux principes du produit : Core Go unique, SQLite source de vérité métier, React client, SystemVault pour les secrets, isolation par profil, audit, versionnement du contrat d’environnement, reproductibilité, tests positifs/négatifs et états explicites `CONFIGURED`, `OBSERVED`, `SUPPORTED`, `UNSUPPORTED`, `RUNTIME_DEFINED`, `WARNING` et `FAIL`.

La validation de leur périmètre produit ne constitue pas une promesse de non-détection ni une autorisation de contourner un contrôle tiers. Les restrictions du §16 restent intégralement applicables : pas de vol de données, credential/cookie theft, fraude, CAPTCHA solving/bypass, contournement anti-bot/anti-fraude, accès non autorisé ou contournement d’un mécanisme de sécurité.

Une capacité qui n’est pas supportée par le runtime doit rester `UNSUPPORTED` ou `RUNTIME_DEFINED`. Une fonctionnalité non prouvée par Core Go, son contrat, ses tests, son isolation, sa redaction et sa preuve ne doit pas être déclarée `IMPLEMENTED`.

### 0.3 Ordre d’implémentation obligatoire

L’ordre du prompt MVP est imposé : **Notes + Custom Fields → Profile Templates → Profile History → Archive/Restore → Bulk Operations → Cookies Import/Export → ProxyProvider intégré → Extensions → History/Bookmarks → Password Manager Policy → Environment Diagnostics → Canvas/WebGL/Audio diagnostics → ClientRects → Geolocation QA → Hardware diagnostics → Font Bundle → Drift Detection → Profile Health → Local Session Tracking → Import/Export complet → API → Dashboard → tests complets**.

---

## 1. Vision, positionnement et limites éthiques

ForgeLocal est un **gestionnaire local-first d’environnements de navigation isolés** destiné à la confidentialité, à l’assurance qualité, aux tests multi-environnements, au développement et à l’automatisation expressément autorisée. Il ne doit pas être présenté comme un produit garantissant une absence de détection ou un moyen de contourner une protection.

La proposition de valeur repose sur le contrôle local des données, la reproductibilité des profils de test, l’isolation des sessions, l’auditabilité du cycle de vie et une API locale maîtrisée. Le produit ne dépend pas d’un cloud par défaut et ne conserve pas de secret dans son interface web.

ForgeLocal peut se positionner commercialement dans la catégorie **Browser Identity, Browser Environment Isolation, Privacy and Authorized Automation**. Le terme « anti-detect » peut servir à situer le marché, mais il ne définit ni une promesse produit ni un mécanisme de contournement. La valeur commerciale recherchée est la capacité à construire, conserver et rejouer des **environnements navigateur cohérents, explicites et contrôlables**.

| Usage autorisé | Usage exclu |
| --- | --- |
| QA, développement, tests de compatibilité, isolation légitime de sessions et automatisation d’un système autorisé | Vol de données, credential/cookie theft, fraude, CAPTCHA solving/bypass, contournement anti-bot/anti-fraude, accès non autorisé ou contournement d’un mécanisme de sécurité |
| Diagnostic local de proxy et connectivité | — |
| API locale liée au loopback pour des scripts autorisés | API exposée publiquement, cloud launch, marketplace ou serveur de contrôle distant dans le MVP |

### 1.1 Cas d’usage autorisés et différenciation commerciale

| Cas d’usage | Valeur ForgeLocal | Limite opérationnelle |
| --- | --- | --- |
| QA web multi-environnements | Rejouer des scénarios avec runtime, locale, timezone, viewport et permissions documentés | Les scénarios visent une application ou un environnement autorisé du client |
| Vérification de contenu, campagnes et expériences localisées | Tester l’affichage dans des configurations régionales de test cohérentes | Aucun accès à une ressource protégée par contournement de règle ou de contrôle |
| Recherche web et veille documentaire | Séparer durablement les sessions de recherche, données et préférences de contexte | Respect des conditions applicables, limites raisonnables et règles du site |
| Collecte de données publiques autorisée | Automatiser des consultations/extractions autorisées via API, Playwright ou CDP | Respect de l’autorisation du système cible |
| Confidentialité client/projet | Isoler stockage, cookies, extensions autorisées, proxy et diagnostics entre projets | Aucun secret dans dashboard, logs, SQLite en clair ou backups non chiffrés |
| Automatisation d’un workflow détenu ou autorisé | API locale, sessions explicites et journal d’audit | L’utilisateur doit être autorisé à piloter le système cible |

### 1.2 Profils d’environnement persistants

Un profil ForgeLocal combine l’isolation de stockage et un **contrat d’environnement persistant**. Cette configuration reste stable tant que l’utilisateur ne la modifie pas explicitement ; chaque changement est versionné, audité et vérifiable dans le cadre de QA.

```
Profil ForgeLocal
├── Stockage isolé : cookies, cache, IndexedDB, extensions autorisées
├── Runtime : navigateur, version, provenance et état de qualification
├── Environnement : OS supporté, viewport, locale, timezone, permissions
├── Réseau : politique proxy/DNS redacted et résultat de connectivité
├── Diagnostics : capacités navigateur et confidentialité redacted
└── Audit : version du contrat, changements et résultats de validation
```

Les presets doivent représenter uniquement des environnements de test réellement supportés par le runtime et l’OS déclarés. ForgeLocal refuse les configurations contradictoires, par exemple un preset de navigateur ou de plateforme incompatible avec le runtime qualifié.

Le contrat d’identité ne promet pas une imitation d’un tiers. Il distingue toujours quatre catégories : **configurable légitimement**, **observable seulement**, **dépendant du runtime**, ou **non supporté sans modification du moteur**. Toute propriété non supportée doit apparaître comme `unsupported` ou `runtime_defined`, et ne doit jamais être simulée silencieusement.

---

## 2. Décisions fondatrices

### 2.1 Architecture retenue

ForgeLocal repose sur **BrowseForge/Core Go comme unique control plane et unique écrivain de l’état métier**. React/TypeScript fournit le dashboard ; il ne possède ni base métier parallèle, ni secret persistant, ni logique concurrente de lancement. SQLite stocke les métadonnées, les états, les opérations et l’audit ; les données lourdes de navigateur restent dans un répertoire isolé par profil ; les secrets restent dans le coffre natif ou dans le mécanisme chiffré de secours validé.

```
┌────────────────────────────────────────────────────────────┐
│ Dashboard React local                                        │
│ Profils · Groupes · Runtimes · Proxys · Backups · Audit     │
│ Client API mémoire seule — aucun secret persistant           │
└───────────────────────────┬────────────────────────────────┘
                            │ HTTPS/HTTP loopback authentifié
                            ▼
┌────────────────────────────────────────────────────────────┐
│ ForgeLocal Core Go / BrowseForge                             │
│ API · ProfileManager · SessionManager · RuntimeManager       │
│ ProxyProvider · BackupManager · AuditLogger · Migrations     │
└───────┬─────────────────┬─────────────────┬────────────────┘
        │                 │                 │
        ▼                 ▼                 ▼
   SQLite WAL        browser-data/      SystemVault
   métier/audit      isolé par profil   secrets / key_id
        │
        ▼
 Runtime Chromium qualifié — candidat Camoufox séparé et non lançable
```

### 2.2 Répartition des sources étudiées

| Source | Rôle admis dans ForgeLocal | Interdictions structurelles |
| --- | --- | --- |
| **BrowseForge** | Base du Core Go : profils, runtimes, sessions, API locale, health checks, CLI et structure de contrôle | Aucun runtime alpha/non qualifié activé par défaut ; aucun changement qui crée un second control plane |
| **Persona Studio** | Inspiration UX, groupes, tags, diagnostics et parcours de gestion | Pas de bridge FastAPI/subprocess écrivant l’état métier ; pas de CORS permissif |
| **Camoflox privé** | Source propriétaire de patterns de fiabilité, désormais qualifiée pour réimplémentation Go sélective | Pas d’import direct ; pas de backend Node/Electron ; pas de runtime ou lancement Camoflox |
| **Camofox-browser public** | Référence externe distincte, jamais substitut implicite du Camoflox privé | Révision, licence, SBOM et gates distincts avant toute étude technique |
| **ShardBrowser / ShardX** | Référence de séparation profil/runtime/session et de répertoires isolés | Aucun moteur binaire non qualifié ou fermé intégré implicitement |
| **CloakBrowser** | Référence ergonomique de SDK local | Aucun mécanisme de camouflage, d’évasion ou de CAPTCHA |
| **DonutBrowser** | Référence desktop, import/export et ergonomie | Aucun fork ou intégration AGPL sans décision explicite de licence |
| **GoLogin** | Benchmark marché public uniquement | Code, asset, dépendance, schéma interne, flux propriétaire et comportement spécifique totalement exclus |

> **Règle clean-room GoLogin.** Toute fonctionnalité ForgeLocal doit provenir du cahier des charges, d’un besoin légitime ou d’un composant explicitement autorisé ; elle ne doit jamais être dérivée d’un matériau propriétaire GoLogin.

### 2.3 Principes non négociables

1. Le Core Go est l’unique source de vérité et l’unique propriétaire des mutations SQLite, profils, sessions, locks, ports et secrets.

1. React est un client : aucune logique métier parallèle, aucun token dans `localStorage`, `sessionStorage`, IndexedDB, URL, log ou analytics.

1. Chaque jalon exige code intégré, tests réellement exécutés, preuves redacted et démonstration locale ; une validation textuelle seule ne suffit pas.

1. Une source externe ne devient pas intégrable par son nom, sa licence apparente ou une autorisation générale : elle suit le registre de provenance et les gates `PROV-01` à `PROV-07`.

1. Le RC BACK-01 et les lots produit sont séparés. Une progression produit ne lève jamais un gate de release.

---

## 3. Périmètre fonctionnel cible

### 3.1 Profils et groupes

Dans le MVP strict, les ajouts de profils sont limités aux notes, Custom Fields `text`, `number`, `boolean`, `select`, templates locaux, historique lisible, restauration logique, clone isolé, historique/bookmarks, politique de mots de passe et opérations bulk explicitement listées dans le prompt. Aucun partage, RBAC, workspace distant ou sync cloud n’est ajouté.

Un profil représente un environnement de navigation isolé. Il possède un identifiant stable, un nom, des tags, un groupe éventuel, un runtime qualifié, une référence de stockage, une configuration réseau redacted et des états de session. Un groupe applique des politiques communes sans copier les secrets de proxy dans le dashboard.

| Fonction | Priorité | Exigence principale |
| --- | --- | --- |
| Lister, rechercher, filtrer profils et groupes | P0 | Pagination, données redacted, états chargement/vide/erreur/réessai |
| Créer, modifier, archiver un profil | P0 après T09 | Validation serveur, transaction SQLite, audit et `correlation_id` |
| Cloner un profil | P1 | Explicite, isolé, journalisé ; jamais de mélange de session |
| Gérer les tags et groupes | P0 | Contrat ASCII documenté pour `COLLATE NOCASE`; pas de promesse de normalisation Unicode complète |
| Afficher le statut de session | P0 | `stopped`, `queued`, `starting`, `running`, `stopping`, `error`, avec motif redacted |
| Quick Profile | P1 | Création guidée à partir d’un preset d’environnement qualifié |
| Custom Profile | P1 | Paramètres explicites, validation de cohérence et version du contrat |
| Clone | P1 | Copie isolée des métadonnées autorisées, jamais des secrets ni de la session active |
| Import/Export | P1 | Dry-run, redaction, parité, audit et format versionné |
| Création/modification en lot | P2 | Opérations bornées, auditables et sans duplication de secrets |
| Notes, recherche et filtres | P1 | Métadonnées non sensibles, indexées et redacted |
| Archivage/restauration logique | P1 | État réversible, transactionnel et audité |

### 3.1.1 Profile Lifecycle

Chaque profil suit un cycle d’état explicite et auditable :

```
CREATED → CONFIGURED → VALIDATED → READY
READY → STARTING → RUNNING → STOPPING → STOPPED → READY
```

Les états exceptionnels sont `ERROR`, `RECOVERY`, `QUARANTINED` et `ARCHIVED`. Une transition invalide est refusée par le Core, enregistrée avec un motif redacted et ne doit pas laisser de session, lock ou mutation partielle. Le statut `READY WITH WARNING` est possible uniquement lorsque les contrôles bloquants passent et que les avertissements sont affichés à l’utilisateur.

### 3.2 Runtimes

Le runtime doit être décrit et qualifié comme `stable`, `candidate`, `experimental`, `external` ou `disabled`. Sa version, sa provenance, son hash, sa signature si disponible, sa licence et sa stratégie de rollback sont explicitement enregistrés.

| Runtime | Statut actuel | Règle |
| --- | --- | --- |
| Chromium QA qualifié | Runtime de référence pour les tests locaux | Version pinée, provenance et checksums requis |
| Camoufox | `candidate`, non lançable | Qualification indépendante, provenance, licence, tests runtime et gate séparé obligatoires |
| Runtime alpha/non signé | `disabled` | Aucune activation implicite |

### 3.2.1 Cycle de vie runtime et compatibilité identité

Le cycle runtime suit une chaîne vérifiable : **registre → version exacte → compatibilité profil → préparation → lancement futur → health check → récupération → rollback**. Un profil ne peut être associé qu’à un runtime présent dans le registre et compatible avec son contrat d’environnement. Une mise à jour ne doit jamais modifier silencieusement l’identité persistante du profil ; elle crée une nouvelle version de contrat, soumise à validation et audit.

| Étape | Exigence | Preuve attendue |
| --- | --- | --- |
| Résolution | Sélection d’un runtime qualifié et versionné | `runtime_id`, version, checksum et décision de qualification |
| Compatibilité | Vérification OS, architecture, viewport, locale, timezone et permissions | Résultat du Consistency Engine |
| Préparation | Création ou réouverture du stockage isolé | Répertoire de profil contrôlé et état redacted |
| Health check | Vérification du processus et des capacités attendues | Rapport `pass`, `warning` ou `fail` |
| Recovery | Nettoyage idempotent après erreur ou interruption | Locks, processus et état SQLite réconciliés |
| Rollback | Retour à une version qualifiée précédente | Référence de version et événement d’audit |

### 3.3 Proxys

ForgeLocal implémente d’abord une interface générique `ProxyProvider`. Les secrets, identifiants et URL sensibles sont uniquement référencés par un `proxy_secret_ref` opaque dans le coffre ; ils ne figurent ni dans SQLite en clair, ni dans les réponses API, ni dans les logs, ni dans les backups.

Le test proxy mesure seulement connectivité, protocole, latence approximative et erreur technique. Il ne mesure pas une supposée « réputation » d’IP et ne cherche pas à contourner des politiques de plateformes.

### 3.4 Import, export, backup et restauration

L’import JSON vers SQLite doit être transactionnel, précédé d’un `dry-run`, contrôlé par parité et journalisé par source. Les mutations doivent rester atomiques ; le journal d’opération durable suit les transitions `started → validated/committed/failed` et marque après crash les opérations inachevées `INTERRUPTED_BEFORE_COMPLETION` sans reprise implicite.

BACK-01 définit le backup `.flbackup` : conteneur unique `FLBK … FLEND`, AES-256-GCM, AAD, nonce issu de `crypto/rand`, permissions restrictives, publication atomique, `fsync`, restauration isolée vers un nouvel identifiant et réconciliation au démarrage. Les archives ne doivent contenir aucun secret en clair.

### 3.5 Matrice de couverture fonctionnelle Browser Identity

ForgeLocal vise une **parité fonctionnelle de catégorie** avec les fonctionnalités publiques documentées chez GoLogin, sans reproduire son implémentation, son code, ses assets, ses services ou ses mécanismes propriétaires. Chaque capacité ci-dessous devient un contrat indépendant, versionné et testé.

| Famille | Capacité ForgeLocal cible | Priorité | Condition |
| --- | --- | --- | --- |
| Navigator | User-Agent, version runtime, plateforme, langues, DoNotTrack et Client Hints selon support | P0 | Runtime qualifié et Consistency Engine |
| Screen | Résolution, viewport, orientation, profondeur couleur, pixel ratio et DPI | P0 | Fixture et matrice runtime |
| Hardware | CPU threads, RAM rapportée, GPU vendor/renderer et capacités observées | P0 | Distinction `reported`/`observed`, aucune valeur silencieusement fabriquée |
| Rendering | Canvas, WebGL, AudioContext et ClientRects via modes déclarés et compatibles runtime | P1 | Rendering Profile, reproductibilité et audit |
| Fonts | FontBundle déclaré, versionné, compatible avec le runtime et testable | P1 | Provenance et test de rendu |
| Media devices | Caméras, microphones, sorties audio, permissions et capacités de fixture | P1 | Autorisation de test et absence de secret |
| Réseau | HTTP/HTTPS/SOCKS5, DNS, WebRTC, géolocalisation de scénario et connectivité | P0/P1 | ProxyProvider, tests de fuite et redaction |
| Stockage | Cookies, LocalStorage, SessionStorage, IndexedDB, extensions, historique, bookmarks et password-manager policy | P0/P1 | Isolation par profil et non-contamination |
| Plugins/extensions | Extensions autorisées, stockage d’extension, plugins nécessaires et politique d’import | P1 | Inventaire, licence et risque documentés |
| Profil | Création, clone, import/export, recherche, tags, groupes, notes, archivage et opérations bulk bornées | P0/P1 | SQLite atomique, audit et limites de charge |
| Workflow | CDP/Playwright/Puppeteer/Selenium compatibles, headless/headful et arrêt global | P1 | Cible autorisée, timeout, audit et tests négatifs |
| API/SDK | REST local, SDK documentés, intégration MCP optionnelle et clients versionnés | P1/P2 | Loopback, token mémoire seule, limites et redaction |
| Collaboration | Partage de profils, session lock, RBAC, dossiers, invitations et politiques d’accès | P2 | Workspace gouverné, audit et séparation des secrets |
| Workspace security | 2FA pour espace distant optionnel, suivi des sessions, révocation et appareils autorisés | P2 | Uniquement si un service distant est activé |
| Web app/cloud | Dashboard web, lancement distant et sessions cloud optionnelles | P2 | Architecture, menace, chiffrement, consentement et gates dédiés |
| Plateformes | Ubuntu, Windows, macOS et Android selon matrice native distincte | P1/P2 | Aucune promesse avant preuve par OS/runtime |

Les fonctions de modification des signaux sont traitées comme des capacités dual-use : ForgeLocal doit indiquer la valeur configurée, la valeur observée, la capacité du runtime, l’impact attendu et la base d’autorisation. Toute incompatibilité devient `warning` ou `fail` ; aucune modification cachée n’est admise.

### 3.6 Paramètres de profil avancés

Un profil peut déclarer, selon les capacités réellement supportées : start URL, restauration des onglets, sauvegarde de bookmarks/historique, politique de sauvegarde des mots de passe, LocalStorage, IndexedDB, stockage d’extensions, Google services, plugins nécessaires, extensions système, launch arguments contrôlés et DNS personnalisé. Chaque paramètre possède une valeur par défaut sûre, un statut de compatibilité, un impact privacy et un événement d’audit.

### 3.7 Bulk, workflows et actions synchronisées

Pour le MVP strict, les opérations bulk sont limitées à : sélection multiple, archivage, restauration, changement de groupe, ajout/suppression de tags, export multiple et validation d’environnement multiple. Elles exigent concurrence bornée, idempotence, progression visible, erreurs par profil, absence d’arrêt partiel silencieux et audit avec `correlation_id`.

ForgeLocal prévoit des opérations en lot bornées : création, clone, archivage, import/export, assignation de groupe, validation d’environnement et lancement de scénarios autorisés. Un orchestrateur peut synchroniser des actions entre profils uniquement pour un workflow détenu ou explicitement autorisé ; il applique une limite de concurrence, un arrêt global, un timeout par tâche, l’idempotence et un audit redacted. Les fonctions explicitement listées dans la table des exclusions sont hors périmètre.

### 3.8 Collaboration, workspace et sécurité d’accès

Le périmètre P2 prévoit des workspaces séparés, dossiers, tags, droits `viewer`, `runner`, `editor`, `manager` et `admin`, session lock par profil, invitations sans transmission de mots de passe, révocation, suivi des sessions et journal d’accès. Un service distant éventuel devra être conçu comme une extension séparée du Core local et ne pourra pas devenir une nouvelle source de vérité métier sans révision d’architecture.

### 3.9 Offre commerciale et feature tiers

Les offres Pro, Team et Enterprise appartiennent à la vision v3.8 mais ne sont pas implémentées dans ce MVP strict. Le lot courant reste local et ne comprend ni Cloud, ni partage, ni RBAC, ni 2FA, ni Enterprise.

| Offre | Fonctionnalités cibles |
| --- | --- |
| Free/Community | Core local, profils limités, isolation, Identity Engine de base, diagnostics, Chromium qualifié, API loopback limitée et import/export local |
| Pro | Profils avancés, bulk, ProxyProvider, automation CDP/Playwright, SDK, backups avancés, Health historique et support prioritaire |
| Team | Workspaces, partage, session lock, RBAC, audit centralisé et politiques d’organisation |
| Enterprise | SSO/2FA, gouvernance, déploiement administré, multi-OS qualifié, support contractuel, API gouvernée et intégrations optionnelles |

Le modèle commercial peut commencer par une édition Free/Community et évoluer vers des abonnements Pro, Team et Enterprise. Les limites commerciales portent sur l’échelle, la collaboration, le support et les fonctions administrées ; elles ne doivent pas désactiver les protections de sécurité fondamentales.

---

## 4. Architecture technique détaillée

### 4.1 Composants Core

| Composant | Responsabilité | Règle critique |
| --- | --- | --- |
| `ProfileManager` | Validation et mutation des profils | Seul écrivain métier avec transactions SQLite |
| `GroupManager` | Groupes, tags et politiques | Pas de secret hérité renvoyé au dashboard |
| `SessionManager` | États et cycle de vie de session | États atomiques, `correlation_id`, cleanup idempotent |
| `RuntimeManager` | Résolution, provenance, version et rollback runtime | Aucun lancement d’un runtime non qualifié |
| `LaunchManager` | Queue, locks, timeout, annulation et attachement runtime | Développement T08 limité et testable sous `-race` |
| `ProxyProvider` | Références proxy, test de connectivité, association | Secrets dans le coffre seulement |
| `BackupManager` | Backup/restauration chiffrés et recovery | BACK-01, snapshots isolés, artefacts malformés mis en quarantaine |
| `ImportManager` | Migration JSON → SQLite et rapports | Dry-run, parité, rollback métier, journal durable |
| `AuditLogger` | Événements exploitables sans secrets | Logs redacted, audit append-only et `correlation_id` métier |
| `LocalApi` | API loopback et bootstrap de session | Bearer court, mémoire seule, CORS strict et refus hors loopback |

### 4.2 SQLite et stockage local

SQLite fonctionne en WAL et conserve les métadonnées de profils, groupes, runtimes, fournisseurs proxy, références de secrets, sessions, backups, restaurations, audit et migrations. Les données Chromium volumineuses restent sous `browser-data/<profile_id>/` avec isolation filesystem.

Les migrations versionnées actuellement attendues sont les suivantes :

| Version | Objet |
| --- | --- |
| v1 | BACK-01 : `backup_operations`, `backups`, `restore_operations`, `audit_events`, `schema_migrations` |
| v2 | Schéma produit canonique : profils, groupes, runtimes et références de stockage |
| v3 | Références proxy et index partiels sur `proxy_provider_id` / `proxy_secret_ref` |
| v4 | Journal durable d’imports et d’opérations de migration |

Les requêtes proxy canoniques doivent contenir les prédicats requis par les index partiels. Des tests `EXPLAIN QUERY PLAN` doivent empêcher une régression de plan pour les lectures de profils et groupes associées aux proxies.

### 4.3 Coffre et secrets

Le coffre système prioritaire est Keychain, Secret Service ou Credential Manager selon l’OS. SQLite ne conserve que des références telles que `key_id` ou `proxy_secret_ref`. Le fallback chiffré doit être documenté, ne jamais être un fallback clair et suivre le format crypto approuvé BACK-01.

| Interdit | Obligatoire |
| --- | --- |
| Secret dans `profile.json`, SQLite en clair, log, API, capture de preuve ou archive non chiffrée | Redaction, permissions restrictives, séparation référence/valeur, tests de fuite |
| Token persistent du dashboard | Session locale courte et mémoire seule |
| CORS ouvert ou API réseau | Loopback strict et vérification de l’origine |

### 4.4 API locale et bootstrap

Le Core expose une API versionnée sous `/api/v1`, liée à `127.0.0.1` et `::1` seulement. Le bootstrap lecture seule fournit un code unique, expirant, échangeable uniquement depuis une origine loopback locale. Le token issu de l’échange est en mémoire, jamais persistant. Les erreurs `401` purgent immédiatement la session React.

| Catégorie | Exemples | Règle |
| --- | --- | --- |
| Health | `GET /health` | `X-Request-ID`, aucun `correlation_id` métier artificiel |
| Lecture seule | Profils, groupes, runtimes, résumé, audit redacted | Pagination, limites, champs redacted, Bearer requis |
| Mutation future | Profils, groupes, proxies, backups, import | Contrat serveur séparé, validation, transaction, audit et `correlation_id` |
| Lancement futur | Sessions/runtime | Autorisé seulement après CAMO-CORE-01 et gate runtime |
| Diagnostics | Navigator, Battery, Network, WebRTC, Storage, Plugins/MIME, Input, Audio, WebGL, Fonts, Performance et Permissions | Lecture redacted, loopback, pagination si nécessaire et aucun secret |

---

## 5. Dashboard React

Le dashboard adopte l’identité « Atelier de contrôle local » et doit être honnête sur son état : données démo clairement marquées lorsque le Core est absent ; actions métier non disponibles affichées comme désactivées ; Camoufox présenté comme candidat non lançable.

Les écrans Profils, Groupes, Runtimes, Proxys, Backups et Audit utilisent les mêmes états : `chargement`, `vide`, `erreur`, `réessai`, `Core indisponible`, `données de démonstration`. Une réponse invalide, un timeout ou un `401` ne laisse jamais d’anciennes données démo se faire passer pour des données Core.

La création et l’édition d’un profil doivent exposer uniquement les paramètres supportés par le contrat serveur : nom, runtime qualifié, OS/architecture de test, locale, timezone, viewport, permissions, politique réseau redacted, identité persistante et diagnostics. Les diagnostics ajoutés en v3.9.1 sont affichés avec `PASS`, `WARNING`, `FAIL`, `UNSUPPORTED` ou `RUNTIME DEFINED`. Les actions indisponibles doivent afficher leurs préconditions et ne jamais être présentées comme opérationnelles.

| Jalons UI déjà prouvés | État |
| --- | --- |
| Bootstrap local mémoire seule, code unique/expirant, loopback strict, `401` et absence de persistance | **Validé T05 / BOOTSTRAP-RO-01** |
| Profils lecture seule redacted | **Validé dans le flux bootstrap** |
| Groupes et Runtimes depuis projections SQLite/Core, pagination et Playwright | **Validé T06** |
| Mutations, lancement, proxies, backups et restauration UI | **Interdits jusqu’à leurs contrats Core dédiés** |

---

## 5.1 Diagnostics d’environnement, de compatibilité et de confidentialité

ForgeLocal ajoute un **Environment & Privacy Diagnostics Engine** destiné à la QA, à la reproductibilité et au diagnostic local. Ce module ne constitue pas un moteur de falsification d’empreinte, de bruit Canvas/Audio, de WebGL simulé ou d’évasion de détection. Il décrit et vérifie l’environnement réellement utilisé ou explicitement configuré pour des scénarios de test contrôlés.

Chaque profil peut être associé à un contrat d’environnement versionné. Ce contrat indique les propriétés de compatibilité utiles au test, les politiques de permissions et les résultats de diagnostic redacted ; il ne doit ni prétendre reproduire l’identité d’un tiers, ni générer de valeurs aléatoires pour contourner une plateforme.

| Domaine | Donnée QA autorisée | Diagnostic / critère de cohérence | Falsification interdite |
| --- | --- | --- | --- |
| OS et navigateur | OS supporté, runtime, canal, version, provenance et hash | Runtime compatible avec l’OS et la matrice de qualification | User-Agent ou OS inventé pour imiter un autre appareil |
| Écran | Viewport, échelle, DPI et rendu testés | Régression visuelle et cohérence du scénario de test | Résolution aléatoire destinée à tromper un tiers |
| Canvas, WebGL et Audio | Capacités API, erreurs et rendu de fixture QA | Comparaison stable entre deux versions du même runtime qualifié | Bruit artificiel, spoofing GPU/Canvas/Audio ou masquage d’API |
| Polices | Bundle de test déclaré et disponibilité vérifiée | Test de rendu avec fixtures autorisées | Énumération ou falsification pour imiter une machine tierce |
| Hardware | Capacités réelles utiles à la compatibilité et aux performances | Rapport redacted sans identifiants matériels inutiles | CPU, RAM ou GPU inventés/masqués |
| CPU/RAM exposés | Mesure de capacités et limites pour tests de performance | Comparaison avec le runtime qualifié | Valeurs fabriquées pour imiter une machine tierce |
| Media devices | Permissions, nombre et capacités dans une fixture QA | Observé, attendu et écart redacted | Masquage ou fabrication destiné à tromper un tiers |
| ClientRects/touch/sensors | Tests de compatibilité d’interface et d’APIs | Résultat reproductible par runtime | Émulation cachée à des fins d’évasion |
| Timezone et locale | Paramètres explicites de scénario QA | Langue, format, timezone et test cohérents | Changement automatisé pour contourner une règle externe |
| Géolocalisation | Permission et simulation dans une application QA contrôlée | Consentement, origine de test et journal redacted | Fausse localisation destinée à contourner une politique tierce |
| WebRTC | Permission, connectivité de test et diagnostic de fuite locale | Rapport de confidentialité et test de régression local | Masquage/manipulation visant à éviter la détection d’une plateforme |
| Réseau | Proxy explicite, DNS, connectivité et latence | Transport cohérent avec la politique profil/groupe | État et comportement déterminés par le contrat réseau |

Le contrat d’environnement ne stocke ni cookies, ni empreinte persistante, ni identifiant matériel, ni valeur proxy secrète. Les mesures détaillées restent locales et redacted dans les preuves ; le dashboard présente seulement un résumé d’état, la version du contrat et les anomalies exploitables.

### 5.2 Contrat de données `EnvironmentProfile`

| Champ | Description | Exigence |
| --- | --- | --- |
| `id` | Identifiant versionné du contrat | Stable et non devinable |
| `runtime_id` / `runtime_version` | Runtime et version qualifiés | Doivent exister dans la matrice runtime |
| `os_target` | OS/architecture de test supportés | Déclaration de compatibilité, non usurpation |
| `viewport_policy` | Dimensions et échelle du scénario | Valeurs explicites et justifiées par le test |
| `locale` / `timezone` | Paramètres de scénario QA | Contrôlés et journalisés ; pas de rotation furtive |
| `permission_policy` | Géo, caméra, micro, WebRTC et notifications | Consentement explicite et origine de test documentée |
| `network_policy_ref` | Référence redacted proxy/DNS/connectivité | Secret externe au contrat et au dashboard |
| `diagnostic_version` | Version des tests de capacités | Comparaison reproductible |
| `compatibility_status` | `unknown`, `qualified`, `warning`, `failed` | Résultat de test, jamais score de détection |
| `rendering_profile_ref` | Référence des fixtures Canvas/WebGL/Audio | Observations et reproductibilité, jamais bruit ou spoofing |
| `font_bundle_ref` | Bundle de fonts déclaré et versionné | Présence, rendu et provenance redacted |
| `media_device_policy` | Permissions et capacités de fixtures QA | Observation et compatibilité, jamais présentation falsifiée |
| `hardware_capability_policy` | Capacités CPU/RAM/GPU réellement disponibles | Mesure de compatibilité/performance, jamais valeurs fabriquées |
| `generator_version` | Version du générateur de contrat | Reproductibilité et audit des presets |
| `navigator_diagnostics` | Diagnostic Navigator Extended | Version, statut et valeurs redacted |
| `battery_diagnostics` | Diagnostic batterie | `observed`, `runtime_defined` ou `unsupported` |
| `network_information` | Informations réseau disponibles | Résultat redacted sans secret |
| `webrtc_diagnostics` | Permission, ICE/STUN/TURN et leak status | IP redacted et statut versionné |
| `storage_capabilities` | Capacités et isolation du stockage | SessionStorage/WebSQL explicitement couverts |
| `plugin_inventory_ref` / `mime_type_diagnostics` | Inventaire et types MIME | Observé/supporté/non supporté |
| `input_device_capabilities` | Touch, pointer, hover, sensors, orientation | Compatibilité runtime |
| `audio_diagnostics` / `webgl_diagnostics` | Détails rendering et fixtures | Hash redacted et régression |
| `font_rendering_diagnostics` | Mesures de rendu des fonts | Drift détectable |
| `performance_diagnostics` | Timings de performance runtime | Comparaison et régression |
| `permission_matrix_ref` | États attendus/observés par origine | `PASS`, `WARNING` ou `FAIL` |

### 5.3 Critères d’acceptation diagnostics QA

| ID | Critère |
| --- | --- |
| AC-ENV-01 | Un contrat d’environnement est versionné, redacted et relié à un runtime qualifié. |
| AC-ENV-02 | Les diagnostics Canvas/WebGL/Audio sont reproductibles entre deux exécutions du même runtime qualifié et ne modifient aucune API navigateur. |
| AC-ENV-03 | La permission WebRTC et l’exposition locale éventuelle sont rapportées sans adresse sensible dans l’interface ou les logs. |
| AC-ENV-04 | Locale, timezone, viewport et permissions sont des paramètres QA explicites, jamais une identité imitée. |
| AC-ENV-05 | Toute simulation de géolocalisation est bornée à une application ou origine de test contrôlée et journalisée. |
| AC-ENV-06 | Aucune valeur ne vise à contourner CAPTCHA, anti-bot, anti-fraude ou une règle de plateforme. |

### 5.4 Consistency Engine et observabilité

Le **Consistency Engine** compare l’identité configurée, le runtime réellement utilisé et l’environnement observé. Il produit un résultat `pass`, `warning` ou `fail`, accompagné d’anomalies redacted et d’une version de diagnostic. Il ne produit pas de score de détection et ne cherche pas à optimiser une identité contre un tiers.

```
IDENTITY CONFIGURED
        ↓
QUALIFIED RUNTIME
        ↓
OBSERVED ENVIRONMENT
        ↓
CONSISTENCY ENGINE
        ↓
PASS / WARNING / FAIL
```

Pour chaque catégorie, le cahier exige une déclaration explicite de capacité :

| Catégorie | Configurable pour QA | Observable | Dépend du runtime | Non supporté sans modification du moteur |
| --- | --- | --- | --- | --- |
| User-Agent et version | Preset runtime documenté | Valeur réellement exposée | Oui | Toute imitation arbitraire d’un autre runtime |
| OS et architecture | Cible de compatibilité | Signaux réellement observés | Oui | Usurpation silencieuse du système |
| Locale, timezone, viewport | Oui, dans les limites du runtime | Oui | Partiellement | Rotation furtive ou incohérente |
| WebRTC | Permissions et scénario de connectivité | Exposition/fuite locale redacted | Oui | Masquage destiné à éviter une détection |
| Canvas, WebGL et Audio | Fixtures et tests de régression | Rendu et capacités | Oui | Bruit, spoofing ou modification cachée |
| Polices et hardware | Bundle de test et capacités utiles | Disponibilité et performances redacted | Oui | Inventaire ou fabrication d’une machine tierce |

### 5.4.1 Rendering Profile

Le **Rendering Profile** regroupe les observations et fixtures QA nécessaires à la reproductibilité du rendu, sans modifier les APIs du navigateur :

| Sous-domaine | Données conservées | Validation |
| --- | --- | --- |
| Canvas | Rendu observé, fixture déclarée et hash redacted | Comparaison stable entre exécutions du runtime qualifié |
| WebGL | Vendor/renderer observés et capacités exposées | Compatibilité avec le runtime, sans spoofing |
| Audio | Capacités observées et résultat de fixture audio | Reproductibilité, sans bruit artificiel |
| ClientRects | Résultats de fixtures de mise en page | Régression UI et cohérence du viewport |
| Fonts | Bundle déclaré, disponibilité et rendu de référence | Provenance et compatibilité vérifiables |
| Media devices | Permissions, capacités et nombre observés dans une fixture | Rapport redacted, sans présentation falsifiée |

### 5.4.2 EnvironmentProfile Generator

Le générateur crée un **Environment Preset** composé de :

```
Environment Preset
├── Runtime
├── OS target
├── Architecture
├── Screen
├── Locale
├── Timezone
├── Languages
├── Permissions
├── Fonts bundle
├── Media-device fixture
├── Network policy
└── Hardware capabilities
```

Le **Consistency Engine** vérifie :

```
CONFIGURED
     ↓
SUPPORTED
     ↓
OBSERVED
     ↓
CONSISTENT ?
   ↙       ↘
 PASS     WARNING/FAIL
```

### 5.4.3 New QA Environment

**New QA Environment** crée un nouveau contrat à partir de configurations réellement supportées :

```
New Environment

Runtime: Chromium 151
OS target: Ubuntu
Locale: fr-FR
Timezone: Europe/Paris
Viewport: 1920×1080
Permissions: default
Network: Proxy #12
Fonts: Ubuntu Standard

[Validate Environment]
```

Après validation, le résultat peut être :

```
QUALIFIED ✓
```

Il s’agit d’un profil de test reproductible.

### 5.4.4 Profile Stability / Drift Detection

ForgeLocal conserve un état de stabilité du profil et compare l’environnement courant à sa référence :

```
Profile Health

Identity
   ✓ unchanged

Runtime
   ✓ Chromium 151.0.x

Locale
   ✓ unchanged

Timezone
   ✓ unchanged

Viewport
   ✓ unchanged

WebGL
   ✓ unchanged

Fonts
   ⚠ 2 fonts changed

Permissions
   ✓ unchanged

Storage
   ✓ isolated

Environment consistency
   ✓ PASS
```

Cette fonction devient un **Environment Drift Detector**. Toute différence entre l’état de référence et l’état courant produit un rapport de stabilité et un résultat `PASS`, `WARNING` ou `FAIL`.

### 5.4.5 Browser Environment Coverage — MVP v3.9.2

Cette section étend uniquement la couverture diagnostique du MVP local strict. Chaque module observe ce que le runtime permet réellement d’observer et retourne un statut `observed`, `configured`, `runtime_defined` ou `unsupported`. Aucune valeur observée n’est silencieusement remplacée par une valeur fabriquée.

| Module | Couverture MVP | Résultat attendu |
| --- | --- | --- |
| Navigator Extended Diagnostics | `appName`, `appVersion`, `platform`, `product`, `vendor`, `language`, `languages[]`, `maxTouchPoints`, `userAgent`, `userAgentData/Client Hints`, `DoNotTrack` | Observation redacted et état de capacité |
| Battery Diagnostics | Disponibilité API, charging state/time et niveau exposé lorsque disponible | `observed`, `runtime_defined` ou `unsupported` |
| Network Information Diagnostics | type, `effectiveType`, `downlink`, RTT et `saveData` lorsque disponibles | Observation redacted sans secret proxy |
| WebRTC Diagnostics | permission, ICE candidates, STUN/TURN, media capabilities et leak status | Rapport redacted et tests de non-fuite |
| Storage Capabilities | Cookies, cache, LocalStorage, SessionStorage, IndexedDB, WebSQL capability et Extension Storage | Capacité et isolation par profil |
| Plugin/MIME Diagnostics | Inventaire plugin, version, MIME types, capabilities et support runtime | `observed`, `supported`, `unsupported` ou `runtime_defined` |
| Input & Device Capability Diagnostics | touch, pointer, hover, sensors et orientation lorsque disponibles | Compatibilité et reproductibilité |
| Audio Diagnostics | AudioContext, sampleRate, channelCount, buffer size, output capabilities, fixture et hash redacted | Reproductibilité du runtime qualifié |
| WebGL Diagnostics | vendor, renderer, version, shading language, limites, extensions, fixture et hash redacted | Régression sans spoofing |
| Font Rendering Diagnostics | bundle ID/version/provenance, dimensions, baseline, ascent/descent, smoothing, hinting et kerning lorsque testables | Régression de rendu et drift |
| Runtime Performance Diagnostics | JavaScript, navigation, rendering, startup et resource timing | Comparaison de versions et régression |
| Permission Matrix | Notifications, caméra, microphone, géolocalisation, WebRTC et media | État attendu/observé par origine de test |

Le flux obligatoire reste :

```
CONFIGURED → SUPPORTED → OBSERVED → CONSISTENCY ENGINE → PASS / WARNING / FAIL
```

Le Consistency Engine ne produit jamais `detection score`, `anti-detect score`, `ban probability`, `fraud probability` ou `stealth score`. Les changements sont signalés par Drift Detection avec un diff redacted ; aucune correction automatique silencieuse n’est effectuée.

### 5.4.6 Extension du contrat `EnvironmentProfile`

Lorsque pertinent, le contrat peut référencer les modules suivants, chacun avec une version de diagnostic : `navigator_diagnostics`, `battery_diagnostics`, `network_information`, `webrtc_diagnostics`, `storage_capabilities`, `plugin_inventory_ref`, `mime_type_diagnostics`, `input_device_capabilities`, `audio_diagnostics`, `webgl_diagnostics`, `font_rendering_diagnostics`, `performance_diagnostics` et `permission_matrix_ref`. Les données sensibles restent redacted et les capacités non supportées sont retournées comme `unsupported` ou `runtime_defined`.

### 5.4.7 Extension de Profile Health et Drift Detection

Profile Health intègre Navigator, Screen, Hardware, Battery, Network, WebRTC, Canvas, WebGL, Audio, Fonts, Plugins, MIME types, Storage, Permissions, Touch/device capabilities, Performance et Runtime qualification. Drift Detection compare ces domaines entre l’environnement de référence et l’environnement courant, notamment les changements de batterie, réseau, inventaire plugins, MIME types, rendu fonts, capacités WebGL/audio, permissions et isolation du stockage. Le résultat est `PASS`, `WARNING` ou `FAIL` avec diff redacted.

### 5.4.8 Critères T13 ENV-01 à ENV-15

| ID | Exigence |
| --- | --- |
| ENV-01 | Générer un `EnvironmentProfile` versionné à partir de paramètres supportés |
| ENV-02 | Associer un contrat d’identité à un runtime qualifié |
| ENV-03 | Vérifier la cohérence navigateur exposé et plateforme observée |
| ENV-04 | Vérifier screen, viewport, DPI et rendu de fixture |
| ENV-05 | Valider un `FontBundle` déclaré, versionné et redacted |
| ENV-06 | Produire les diagnostics de capacités WebGL sans spoofing |
| ENV-07 | Produire les diagnostics de reproductibilité Canvas |
| ENV-08 | Produire les diagnostics de reproductibilité Audio |
| ENV-09 | Produire les diagnostics de capacités des media devices |
| ENV-10 | Produire les diagnostics de capacités hardware réellement disponibles |
| ENV-11 | Vérifier la cohérence WebRTC dans un scénario de test contrôlé |
| ENV-12 | Vérifier locale et timezone avec les observations disponibles |
| ENV-13 | Vérifier la cohérence des permissions et de leur origine de test |
| ENV-14 | Exécuter le Consistency Engine sur plusieurs signaux redacted |
| ENV-15 | Alimenter Profile Health avec des contrôles vérifiables, jamais un score de détection |

### 5.4.9 Matrice MVP / Roadmap / Explicitly Excluded

| Catégorie | Statut v3.9.3 |
| --- | --- |
| Navigator Extended, Battery, Network Information, WebRTC, Storage/SessionStorage/WebSQL, Plugins/MIME, Input/Device, Audio, WebGL, Fonts, Performance, Permissions, Profile Health, Drift Detection et les neuf capacités de l’addendum | **MVP / PRODUCT SCOPE** |
| Cloud, Cloud Launch, Web App distante, Profile Sync Cloud, Profile Sharing, RBAC, 2FA distant, Team workspace, Android et Enterprise | **ROADMAP** |
| CAPTCHA solving, CAPTCHA/anti-bot bypass, anti-fraude bypass, contournement de contrôle d’accès, fraude, credential/cookie theft, account farming, quota evasion et detection evasion explicite | **EXPLICITLY EXCLUDED** |

### 5.4.10 Matrice de tests obligatoire v3.9.3

Chaque nouveau module doit fournir un test nominal, un test `unsupported`, un test `runtime_defined`, un test de redaction, un test d’isolation lorsque pertinent, un test de drift, un test API, un test React lorsqu’il existe un affichage et un test négatif de sécurité lorsque pertinent.

| Module | Tests minimum |
| --- | --- |
| Navigator Extended, Battery, Network Information | Nominal, `unsupported`, `runtime_defined`, redaction, API et UI |
| WebRTC/STUN/TURN | Leak status, redaction IP, isolation, API et test négatif de fuite |
| SessionStorage/WebSQL | Isolation Profil A/Profil B, capacité runtime et redaction |
| Plugins/MIME | Inventaire, support runtime, redaction, drift et absence d’injection |
| Touch/Device | Capacité observée, unsupported et reproductibilité |
| Audio/WebGL/Fonts | Fixture, hash redacted, reproductibilité, drift et absence de spoofing |
| Performance | Timings, comparaison runtime, régression et absence de simulation comportementale |
| Permissions | Matrice origine/état attendu/observé, redaction et erreurs invalides |

Le dashboard n’affiche un diagnostic que si le contrat Core/API existe. Chaque valeur est présentée avec `PASS`, `WARNING`, `FAIL`, `UNSUPPORTED` ou `RUNTIME DEFINED`. Les données démo restent explicitement identifiées.

### 5.5 Matrice de qualification runtime

Chaque runtime candidat passe une matrice par OS, version, architecture et contrat d’environnement. La matrice documente le binaire, la provenance, le hash, la licence, le rollback, les résultats de stockage, les extensions autorisées, les permissions, les diagnostics et les limites observées.

| Dimension | Preuve requise |
| --- | --- |
| Provenance | Référence autorisée, version exacte, checksum et signature lorsque disponible |
| Compatibilité | OS, architecture, installation et rollback testés |
| Stabilité | Démarrage/arrêt, crash recovery, réouverture de profil et cleanup locks |
| Environnement QA | Diagnostics Canvas/WebGL/Audio, viewport, locale/timezone, permissions et WebRTC en scénario contrôlé |
| Sécurité | Scan artefact, SBOM, licence, permissions et absence de fuite dans logs |
| Décision | `qualified`, `candidate`, `rejected` ou `disabled`, avec raison et relecteur |

### 5.5.1 Suite de non-régression runtime

Chaque version de Chromium ou du runtime doit être évaluée avec la même suite de tests avant promotion. La suite couvre l’identité configurée, l’environnement observé, le stockage, les permissions, WebRTC, les graphiques, les polices, les media devices, l’automatisation autorisée et la réouverture du profil. Une différence entre deux versions produit un rapport de régression ; elle ne doit pas être masquée par une modification silencieuse du contrat.

| Famille | Contrôle minimal | Résultat |
| --- | --- | --- |
| Identity | Contrat, persistance et valeurs observées | `pass`, `warning` ou `fail` |
| Environment | Locale, timezone, viewport, permissions et géolocalisation de test | Écart attendu/observé |
| Storage | Cookies, cache, IndexedDB, extensions et non-contamination | Aucun transfert entre profils |
| Network/WebRTC | Connectivité, DNS, proxy redacted et exposition locale | Rapport sans adresse sensible |
| Graphics/media | Canvas/WebGL/Audio, fonts et media devices en fixtures QA | Reproductibilité ou régression explicitée |
| Automation | CDP/Playwright, arrêt, timeout et audit | Scénario autorisé terminé proprement |
| Recovery | Crash, réouverture, cleanup et rollback | État préservé ou échec explicite |

### 5.6 Automation Layer autorisée

L’automatisation est un lot P1 et reste subordonnée aux contrats Core, à la qualification du runtime et à l’autorisation sur la cible. Le flux cible est : **API locale → sélectionner un profil → configurer un contrat d’environnement → associer une politique réseau → préparer le runtime → lancer après gate → connecter CDP/Playwright → exécuter le scénario → fermer et auditer**. L’automatisation ne doit pas modifier silencieusement l’identité persistante ni contourner une limite, un CAPTCHA, une authentification ou un contrôle d’accès.

| Contrôle | Exigence |
| --- | --- |
| Autorisation | Cible détenue, administrée ou explicitement autorisée par l’utilisateur |
| Identité | Contrat persistant inchangé pendant le scénario, sauf modification explicite et auditée |
| Arrêt | Timeout, annulation, bouton d’arrêt global et cleanup idempotent |
| Données | Aucun secret dans le script, le dashboard, les logs ou les preuves |
| Traçabilité | `correlation_id`, résultat, durée, erreurs redacted et artefact de test |

### 5.7 Environment Health et Profile Health Center

ForgeLocal expose un indicateur **Environment Health** fondé sur des contrôles vérifiables, et non un score de détection. Le score est décomposable, versionné et accompagné des preuves ou anomalies redacted. Il ne doit jamais prétendre mesurer la probabilité qu’une plateforme détecte un utilisateur.

| Contrôle | Exemple d’état |
| --- | --- |
| Browser, OS et runtime | `PASS`, `WARNING` ou `FAIL` |
| Locale, timezone et réseau | Cohérent, écart à examiner ou échec |
| WebRTC, graphics, fonts et media devices | Observé, attendu et compatibilité |
| Storage isolation | `PASS` uniquement après test de non-contamination |
| Identity persistence | `PASS` après fermeture/réouverture |
| Runtime qualification | `PASS` uniquement pour une version qualifiée |

Le dashboard peut regrouper ces résultats dans un **Profile Health Center**. Toute recommandation doit être formulée comme une correction de configuration ou un approfondissement QA, jamais comme une optimisation d’évasion.

## 6. Sécurité, confidentialité et modèle de menace

ForgeLocal traite la sécurité comme une fonction P0. Les menaces prises en compte incluent l’accès local non autorisé, l’exposition réseau accidentelle de l’API, le vol de cookies, le runtime compromis, l’import malformé, les traversées de chemins, les backups corrompus, les secrets dans logs/preuves et les courses de sessions.

| Contrôle | Exigence |
| --- | --- |
| Loopback | Test positif sur `127.0.0.1`/`::1` et refus réel hors loopback |
| Authentification | Bearer en mémoire, code unique/expirant, invalidation après `401` |
| Entrées | Validation stricte des IDs, chemins, multipart, JSON/ZIP et tailles/durées bornées |
| Filesystem | Rejet symlinks/traversées, opérations sous racine contrôlée, permissions restrictives |
| Logs/preuves | Redaction obligatoire ; sentinelles interdites dans DOM, API, archives et logs |
| Non-contamination | Cookie, stockage, proxy, identité et audit du profil A ne doivent jamais apparaître dans le profil B |
| Crypto | AES-256-GCM, AAD, nonce aléatoire, pas de crypto maison |
| Concurrence | `context`, timeout, annulation, locks, cleanup idempotent et tests `-race` |
| Scan secrets | Gitleaks verrouillé, scan du delta réel et de l’archive extraite ; le log doit indiquer un périmètre non vide |

---

## 7. Provenance, licences et gouvernance

### 7.1 Registre de composants

`docs/component-rights-register.json` est la source machine-readable. `docs/COMPONENT_RIGHTS_REGISTER.md` est sa vue humaine. Toute dépendance, asset, runtime ou composant importé doit déclarer la source, la révision précise, le statut de droits, l’étendue des droits, les exclusions, le responsable et une `evidence_ref` redacted hors dépôt.

Une CI bloquante doit refuser tout élément tiers `denied`, `unknown` ou sans entrée ; une source interne correctement déclarée peut être `not_required`. Les emails, accords bruts, archives privées, tokens et données personnelles restent hors Git.

### 7.2 Gates `PROV-01` à `PROV-07`

| Gate | Exigence | État Camoflox après T07 |
| --- | --- | --- |
| `PROV-01 Rights` | Révision exacte et droits vérifiés | **PASS** pour snapshot privé attesté |
| `PROV-02 Dependency` | Dépendances, assets, runtimes et bases inventoriés | **PASS** pour qualification ; nouvelle dépendance à revoir |
| `PROV-03 Architecture` | Aucun second écrivain ou control plane | **PASS** |
| `PROV-04 Security` | Scans, modèle de menace et logs redacted | **PASS** pour snapshot redacted hashé |
| `PROV-05 Reliability` | Timeout, annulation, concurrence, recovery et non-régression | **Différé à T08** |
| `PROV-06 SBOM` | SBOM, notices et artefact de provenance | **PASS** pour revue privée ; distribution future à revalider |
| `PROV-07 Product` | Valeur légitime démontrée | **Différé à T08** |

### 7.3 Décision T07

Le statut final est :

```
T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION
```

Cette décision autorise **seulement** l’étude d’un module à la fois et sa réimplémentation indépendante en Go. Elle n’autorise pas la copie/import du code Camoflox, l’exécution d’un backend Node/Electron, l’activation d’un runtime/Camoufox, ni une release.

---

## 8. Qualité, tests et preuves obligatoires

### 8.1 Règle générale

> **Règle de maturité.** Une fonctionnalité n’est pas considérée comme livrée sur la base de sa seule spécification ou de son implémentation. Elle est livrée uniquement lorsqu’elle possède un contrat technique, une implémentation, des tests positifs, des tests négatifs, des tests de concurrence lorsque pertinents, des critères d’acceptation, des preuves reproductibles et un statut explicite dans l’archive de preuves.

Pour chaque jalon, le développeur doit fournir : un commit propre, les versions d’outils, les commandes exactes, la liste préalable des tests, les sorties réellement exécutées, les scénarios positifs et négatifs, une archive de preuves redacted à hashes relatifs, une démonstration locale et une décision de passage limitée.

### 8.1.1 Niveaux de validation et absence de certificats intermédiaires

Un jalon produit ne nécessite pas automatiquement une GitHub Release, une signature de release, un certificat, une attestation d’artefact ou une approbation de production. Ces éléments sont réservés à l’artefact destiné à la distribution publique ou au franchissement d’un environnement protégé.

| Niveau | Exigences obligatoires | Certificat/signature/release |
| --- | --- | --- |
| Développement | Tests unitaires et intégration pertinents, lint, scans CI applicables et démonstration locale lorsque le jalon est déclaré | **Aucun** |
| `main` protégée | Revue de code selon les règles du dépôt, CI obligatoire, tests sélectionnés, `-race` lorsque pertinent et contrôles sécurité | **Aucun certificat de release** |
| Jalon T08–T20 | Contrat, implémentation, tests positifs/négatifs, concurrence si pertinente, critères d’acceptation, preuves redacted et statut explicite | **Aucun certificat ni GitHub Release** |
| Version interne/MVP | Régression, artefact interne identifiable et rapport de limites | Signature facultative, non bloquante par défaut |
| Release publique | Audit final, SBOM/notices, licence/runtime, sécurité, build, provenance, signature/attestation si exigées par la politique de distribution, approbation et publication | **Contrôles requis uniquement ici** |

Cette règle ne supprime jamais les tests, les preuves redacted ou les contrôles de sécurité des jalons. Elle supprime seulement la répétition administrative des certificats et approbations de release avant qu’un artefact ne soit destiné au public.

Un code de sortie `0` sans test sélectionné, une sortie `skip`, un scan sans périmètre, un manifeste non portable, un secret, un statut annoncé sans archive ou un environnement non conforme vaut **échec de validation**.

### 8.2 Outillage de référence

| Outil | Usage |
| --- | --- |
| Go `1.25.13`, `GOTOOLCHAIN=local` | Core, migrations, tests, `go vet`, race detector |
| SQLite | Migrations, intégrité, plans de requêtes |
| Node 22 et pnpm | Dashboard, scripts provenance et Playwright |
| Playwright | E2E dashboard local ↔ Core local |
| Gitleaks `8.18.4` | Delta, snapshots et archives extraites |
| Gosec / Govulncheck / GolangCI-Lint | Contrôles Go ciblés et graphe distribuable |
| `sha256sum`, `unzip`, `curl`, `jq`, `ss`, `lsof`, `flock`, `timeout` | Vérifications de preuves, réseau, ports et artefacts |

### 8.3 Environnements de validation

| Environnement | Tests autorisés / requis |
| --- | --- |
| Sandbox | API loopback, SQLite, migrations, Core, React, Playwright, `-race`, scans, backups de fixtures, T08 sans runtime |
| Hôte Ubuntu natif avec session déverrouillée | SystemVault, coffre verrouillé/révoqué, permissions natives et preuve runtime locale |
| GitHub administré | Rulesets et CI pour `main`; approbations et environnement `production-release` uniquement au stade de release publique; absence de bypass |

### 8.4 Jalons de produit

| Jalon | Objet | Statut |
| --- | --- | --- |
| T00 | Intégrité workspace, RC gelé et outils | **Validé comme préflight des runs T05/T06** |
| T01 | Baseline Go et tests sélectionnés | **Validé dans les preuves bootstrap** |
| T02 | Migrations SQLite | **Validées dans les flux produit** |
| T03 | API bootstrap et loopback | **Validé** |
| T04 | Build React et absence de persistance token | **Validé** |
| T05 | `BOOTSTRAP-RO-01` dashboard local → Core local | **`BOOTSTRAP_RO_APPROVED_VERIFIABLE`** |
| T06 | Groupes/Runtimes lecture seule depuis SQLite/Core | **`T06_APPROVED_VERIFIABLE`** |
| T07 | Provenance Camoflox | **Clôturé pour réimplémentation Go sélective** |
| T08 | Fiabilité Core : queue/locks/cleanup/recovery | **`T08_APPROVED_VERIFIABLE_LOCAL`** (16/08/2026 — archive `t08-r2-final.zip` SHA-256 `4918ac9876545904c822ff72fb3dfcc4f8b12f6fb2214452e308a39b4c0719bb`, 13/13 tests sous `-race`, zéro data race, `go vet` propre, Gitleaks JSON `[]`, cleanup borné documenté ; périmètre strict sans runtime/Camoufox/Proxy/UI/release — **ne qualifie pas T09**) |
| T09 | Écritures profils, audit et `correlation_id` | À venir |
| T10 | Import, proxies, backup/restauration intégrés | À venir |
| T11 | Fuzzing, hardening, charge et endurance | À venir |
| T12 | Démonstration locale MVP et checkpoint | À venir |
| T13 | Browser Environment & Consistency Engine : ENV-01 à ENV-15 et diagnostics v3.9.3 et capacités validées par l’addendum | À préparer après T08/T09 |
| T14 | Runtime lifecycle, Runtime Health, recovery et rollback | À préparer après T08 |
| T15 | Automation Layer contrôlée : CDP/Playwright, Behavioral Simulation, Cookie Bot et Action Synchronizer selon les contrats de l’addendum et les restrictions du §16 | À préparer après T09/T14 |
| T16 | UX produit : création de profil, diagnostics et états d’autorisation | À préparer après contrats Core |
| T17 | Qualification multi-OS : Ubuntu puis Windows/macOS selon moyens de test | Roadmap P1, aucune promesse avant preuve |
| T18 | Collaboration/Enterprise : sync, équipes, RBAC, politiques et cloud optionnels | Roadmap P2, hors MVP local |
| T19 | Profile Health Center, Profile Stability, Environment Drift Detection et nouveaux diagnostics | À préparer après T13/T14 |
| T20 | OpenAPI, SDK local et tests contractuels | À préparer après T09/T13 |
| T21 | Parité fonctionnelle future : profils avancés, bulk, automation, collaboration et feature tiers | Reporté après MVP strict ; aucune implémentation dans ce lot |

### 8.5 Contrats API et SDK local

L’API locale doit être versionnée et documentée avec un schéma OpenAPI ou équivalent. Les routes sont séparées entre lecture, mutation future, diagnostic et lancement ; chaque route définit authentification, pagination, limites, erreurs, redaction, `X-Request-ID` et `correlation_id` lorsqu’il s’agit d’une opération métier. Aucun SDK ou exemple ne doit persister un token ou un secret.

| Domaine | Contrat cible | Condition |
| --- | --- | --- |
| Profils | Lister, consulter, créer, modifier, archiver, cloner | T09 et validation serveur |
| Identité | Consulter le contrat et sa version | T13, sans valeur sensible |
| Diagnostics | Navigator, Battery, Network, WebRTC, Storage, Plugins/MIME, Input, Audio, WebGL, Fonts, Performance et Permissions | T13, redacted, sans score de détection |
| Health | Contrôles décomposés, nouveaux diagnostics et état global | T13/T19, preuves attachées |
| Runtime | Qualification, compatibilité, health et rollback | T14 |
| Automation | Scénario autorisé, timeout, arrêt et audit | T15 |

Les routes de lancement restent désactivées tant que le runtime n’est pas qualifié et que les contrôles Core correspondants ne sont pas validés.

### 8.6 Roadmap produit et périmètre commercial

La cible de la première version commerciale locale est un produit **ForgeLocal White** centré sur profils, isolation, Chromium qualifié, proxy par profil redacted, Browser Identity Engine de QA, diagnostics d’environnement, cohérence, identité persistante, API loopback, import/export contrôlé et dashboard. Cette cible ne constitue pas une promesse de couverture complète des alternatives commerciales ; chaque capacité est annoncée uniquement après preuve dans la matrice de qualification.

| Bloc produit | Cible | Priorité | Condition de sortie |
| --- | --- | --- | --- |
| Profile Manager | CRUD, groupes, tags, clone, import/export | P0 | SQLite, API, audit et tests d’isolation |
| Session/Storage Isolation | Cookies, cache, IndexedDB et extensions séparés | P0 | Tests de non-fuite et recovery |
| Runtime Manager | Version, lancement futur, health check, recovery, rollback | P0 | Runtime qualifié et CAMO-CORE-01 validé |
| Proxy par profil | HTTP/HTTPS/SOCKS5, test et référence coffre | P0 | Contrat secret redacted et tests de connectivité |
| Browser Identity Engine | Contrat persistant, Generate Environment Profile, New QA Environment, Rendering Profile, cohérence et observabilité | P0 | T13 ENV-01 à ENV-15 et matrice runtime validés |
| Environment QA | WebRTC, Canvas/WebGL/Audio, fonts, hardware, locale et viewport | P0 | Diagnostic reproductible, sans falsification |
| Automation | CDP/Playwright/Puppeteer/Selenium, headless/headful et API locale | P1 | Autorisation, arrêt, audit et T15 validés |
| UX/Desktop | Dashboard commercial local et états honnêtes | P1 | T16 et tests navigateur |
| Multi-OS | Ubuntu d’abord, puis Windows/macOS si prouvé | P1 | Matrice native par OS |
| Bulk/Workflow | Opérations en lot et actions synchronisées bornées | P1 | Limites, idempotence, arrêt global et audit |
| Team/Cloud | Synchronisation, partage et lancement distant optionnels | P2 | Nouvelle architecture, menace et gates ; hors MVP |
| Enterprise | SSO, RBAC, politiques et API gouvernée | P2 | Revue sécurité et gouvernance dédiée |

La roadmap ne modifie pas les statuts actuels : le RC BACK-01 reste gelé, `PUBLIC_RELEASE_BLOCKED` reste actif, Camoufox reste non lançable tant que sa qualification n’est pas terminée et aucun module de contournement n’est ajouté.

---

## 9. Périmètre T08 autorisé

T08 est une réimplémentation Go indépendante de la fiabilité Core. Le premier module est choisi dans le registre CAMO-CORE-01 avec une décision `réimplémenter`, un hash source de référence et des exclusions explicites.

### 9.1 Objectif T08

Construire un `LaunchManager`/`SessionManager` Go qui applique une queue bornée, une limite globale, une sérialisation par profil, un timeout, une annulation via `context.Context`, un cleanup idempotent, une reprise après crash et un audit redacted — **sans lancer de navigateur ni runtime** à ce stade.

### 9.2 Critères d’acceptation T08

| ID | Critère |
| --- | --- |
| AC-CAMO-01 | Deux demandes simultanées sur le même profil produisent une seule session ou un refus explicite et audité. |
| AC-CAMO-02 | La limite globale est respectée ; une attente expirée libère ses ressources. |
| AC-CAMO-03 | Après crash simulé, locks, états SQLite et ressources sont réconciliés sans session fantôme. |
| AC-CAMO-04 | Une erreur d’attachement/ressource ne laisse aucun lock ou état bloqué. |
| AC-CAMO-05 | Les scénarios passent sous `go test -race`, sans secret dans les logs ou l’audit. |

### 9.3 Interdictions T08

T08 n’autorise pas ports runtime réels, Camoufox, lancement navigateur, proxy, backup, restauration, import, mutation UI ou release. Toute extension de périmètre exige une nouvelle décision et des tests propres.

---

## 10. BACK-01 et séparation release

### 10.1 Release publique : gate unique et final

La release publique est un événement distinct du développement produit. Elle ne peut être approuvée qu’après un candidat explicitement destiné à la distribution et la fermeture des gates listés ci-dessous. Les succès de T08 à T20, les merges sur `main` ou une version interne ne valent pas approbation de release.

Le pipeline cible est :

```
main protégée → candidat interne → régression complète → sécurité → SBOM/licences
→ build distribuable → provenance/signature/attestation selon politique de distribution
→ approbation finale → publication
```

Aucune étape précédente ne doit demander ou produire un certificat de release uniquement pour faire progresser le produit.

BACK-01 est techniquement validé en local pour le moteur de backup/restauration chiffré, mais le candidat de release public reste gelé. Le candidat RC `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz` a le SHA-256 :

```
553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6
```

Les travaux T05 à T08 ne modifient pas le RC, `release/back01-minimal` ou `dist/back01-minimal`.

| Gate release public | Statut | Condition de levée |
| --- | --- | --- |
| `SCAN_BLOCKED_UNKNOWN` | Bloquant | Triage indépendant de l’alerte historique de provenance du RC, distincte du triage Camoflox T07 |
| SystemVault natif | Bloquant | Matrice par OS cible dans une session graphique native déverrouillée |
| Anti-fuite intégré | Bloquant | Flux réel profil → backup → restauration, sans sentinelle exposée |
| Signature mainteneur | Bloquant | Signature de release réelle, clé publique et vérification indépendante |
| Licence/runtime/OS | Bloquant | Revue distribution runtime, SBOM/notices et compatibilité démontrée |
| Gouvernance GitHub | Bloquant | Rulesets réels, Code Owners, approbations distinctes et absence de bypass |

> **Statuts maintenus :** `PUBLIC_RELEASE_BLOCKED`, pilote suspendu, cinq gates publics en attente. T07 ne lève aucun de ces gates.

---

## 11. Roadmap complète

| Ordre | Lot | Sortie exigée |
| --- | --- | --- |
| 1 | T08 — fiabilité Core | Module Go limité, registre à jour, tests `-race`, recovery et archive de preuve |
| 2 | T09 — écritures profils | Contrats mutation, SQLite atomique, audit, `correlation_id`, E2E UI négatif/positif |
| 3 | T10 — cycle de données | Proxies, import JSON→SQLite, `.flbackup`, restauration isolée et journal |
| 4 | T11 — hardening | Fuzzing, limites, scans, charge, endurance, crash recovery et plans SQLite |
| 5 | T12 — MVP local | Démonstration utilisateur, checklists, checkpoint et périmètre de pilotage local |
| 6 | T13 — Identity/Consistency/Environment Health | Generate Environment Profile, New QA Environment, ENV-01 à ENV-15, Rendering Profile et health décomposable, sans falsification |
| 7 | T14 — Runtime Health et qualification | Qualification par OS/version/architecture, lancement autorisé, recovery, rollback et scénarios QA |
| 8 | T21 — Parité fonctionnelle progressive | Profils avancés, bulk, automation, collaboration et cloud optionnel, chacun avec contrat et preuve dédiée |
| 8 | R00 | Triage indépendant de l’alerte RC `generic-api-key` inconnue |
| 9 | R01 | SystemVault réel sur Ubuntu natif puis cibles annoncées |
| 10 | R02 | Artefact, runtime, SBOM, signature, licence et compatibilité |
| 11 | R03 | Rulesets et environnement de release GitHub réellement administrés |

Les lots R00–R03 ne peuvent pas être simulés par les succès produit. Chaque transition doit conserver la séparation d’artefacts, d’évidence et de décision.

---

## 12. Livrables attendus

Le développeur fournit à chaque livraison : code source, migration le cas échéant, tests sélectionnés et exécutés, sortie `-race` quand applicable, scan redacted, manifeste portable, démonstration locale et rapport de limites. À maturité MVP, il fournit également le guide de build, la spécification API, le modèle de données, le guide utilisateur, le SBOM, les notices, la stratégie runtime et le diagnostic exportable sans cookies, mots de passe, tokens ou contenu de pages.

---

## 13. Règle de décision finale

Le prompt source exact `PROMPT_FORGELOCAL_IMPLEMENTATION_REAL_v3.9.7.md` constitue la consigne d’exécution du lot. Le présent v3.9.7 ne considère jamais une interface, un contrat ou un diagnostic comme une fonctionnalité complète sans mécanisme réel, tests et preuve.

> **ForgeLocal avance par preuves intégrées, pas par déclarations.** Un jalon est accepté seulement si son code, ses tests positifs et négatifs, son environnement, ses preuves redacted et son démonstrateur local sont cohérents. Les statuts de release restent indépendants jusqu’à la fermeture explicite de chaque gate.

## 14. Executive Gap Analysis

L’audit distingue désormais la présence documentaire d’une fonctionnalité et sa complétude opérationnelle. Une fonction n’est `🟢 COMPLET` que si son objectif, comportement, Core, API, stockage, UI, configuration, validation, persistance, versioning, migration, import/export, clone, backup/restore, logs, audit, tests et récupération sont couverts. Les statuts sont `🟢 COMPLET`, `🟡 PARTIEL`, `🟠 À RENFORCER`, `🔵 À VÉRIFIER`, `🔴 MANQUANT` et `⚫ EXCLU`.

| Domaine audité | Statut initial | Lacune principale | Action |
| --- | --- | --- | --- |
| Bootstrap local et lecture redacted | 🟢 COMPLET | T05 vérifié ; mutations hors périmètre du jalon | Maintenir la régression |
| Catalogues profils/groupes/runtimes | 🟢 COMPLET | T06 vérifié en lecture seule | Maintenir la régression |
| Écritures Profile Manager | 🟡 PARTIEL | Contrats de mutation et E2E à finaliser | GAP-001 |
| Notes et Custom Fields | 🟡 PARTIEL | Modèle, validation, export et audit à implémenter | GAP-002 |
| Templates locaux | 🟡 PARTIEL | Versioning, clone indépendant et migration | GAP-003 |
| Clone/isolation | 🟠 À RENFORCER | Session, locks, tokens et stockage à prouver | GAP-004 |
| Archive/restore | 🟠 À RENFORCER | Recovery, collisions et corruption à couvrir | GAP-005 |
| Cookies et stockage | 🟠 À RENFORCER | Dry-run, import atomique, SessionStorage et Cache Storage | GAP-006 |
| History/bookmarks/preferences | 🟡 PARTIEL | Persistance, export, restore et non-contamination | GAP-007 |
| Extensions/plugins/MIME | 🟡 PARTIEL | Inventaire distinct, permissions et support runtime | GAP-008 |
| ProxyProvider | 🟠 À RENFORCER | Health, import/export redacted et erreurs | GAP-009 |
| Identity/Fingerprint | 🟡 PARTIEL | Configuration, observation, versioning et drift à relier | GAP-010 |
| Canvas/WebGL/Audio/ClientRects | 🟡 PARTIEL | Matrice runtime et régression complète | GAP-011 |
| Fonts/media devices/hardware | 🟡 PARTIEL | Capacités réelles, isolation et statuts runtime | GAP-012 |
| Permissions/locale/timezone/geolocation | 🟡 PARTIEL | Matrice par origine, persistance et migration | GAP-013 |
| WebRTC/network diagnostics | 🟠 À RENFORCER | STUN/TURN, redaction et leak status | GAP-014 |
| Runtime Manager Chromium | 🟡 PARTIEL | Update/downgrade, process, crash recovery et rollback | GAP-015 |
| Camoufox | 🔵 À VÉRIFIER | Qualification séparée non terminée | GAP-016 |
| Automation locale | 🟡 PARTIEL | Stop global, idempotence, timeout et audit | GAP-017 |
| Behavioral Simulation | 🟠 À RENFORCER | Scénarios autorisés, interruption et limites explicites | GAP-018 |
| Cookie Bot | 🟠 À RENFORCER | Limité à l’initialisation autorisée du propre profil | GAP-019 |
| Action Synchronizer | 🟠 À RENFORCER | Profils explicitement sélectionnés, concurrence et audit | GAP-020 |
| API/SDK | 🟡 PARTIEL | Contrats OpenAPI, erreurs, limites et versions | GAP-021 |
| Profile History/versioning | 🟡 PARTIEL | Historisation complète et migration | GAP-022 |
| Observability/Health/Drift | 🟡 PARTIEL | Causes, correction contrôlée et vérification après correction | GAP-023 |
| Backups | 🟠 À RENFORCER | Flux réel backup/restore et corruption recovery | GAP-024 |
| Security/SystemVault | 🔵 À VÉRIFIER | Preuve native hors sandbox requise | GAP-025 |
| Multi-OS | 🔴 MANQUANT pour les cibles non Ubuntu | Matrices natives séparées | GAP-026 |
| Cloud/collaboration/RBAC/2FA/Android/Enterprise | ⚫ EXCLU du MVP | Roadmap v3.8, aucun code local actuel | Report P2 |
| Vol de données, credential/cookie theft, fraude, CAPTCHA solving/bypass, contournement anti-bot/anti-fraude, accès non autorisé et contournement d’un mécanisme de sécurité | ⚫ EXCLU | Hors périmètre définitif | Ne pas implémenter |

## 15. Matrice GoLogin → ForgeLocal

Cette matrice reprend uniquement les familles GoLogin déjà documentées dans l’inventaire officiel sauvegardé à partir des pages publiques GoLogin [6] [7] [8] [9]. Elle ne suppose pas qu’un nom similaire signifie une implémentation équivalente.

| Fonction GoLogin | Module ForgeLocal | Statut actuel | Manque | Action obligatoire | Priorité |
| --- | --- | --- | --- | --- | --- |
| Profils persistants/isolation | Profile Manager + Browser Storage | 🟡 PARTIEL | CRUD complet, recovery et migration | GAP-001/GAP-004 | P0 |
| Fingerprint settings | Fingerprint/Environment Engine | 🟡 PARTIEL | Sous-signaux, versioning et drift | GAP-010 | P0 |
| Canvas/WebGL/Audio/ClientRects | Rendering Profile | 🟡 PARTIEL | Qualification runtime et régression | GAP-011 | P0 |
| Fonts/media devices | FontBundle + Device Diagnostics | 🟡 PARTIEL | Inventaire, permissions et isolation | GAP-012 | P0 |
| WebRTC/DNS | Network/WebRTC Diagnostics | 🟠 À RENFORCER | Leak test, redaction et health | GAP-014 | P0 |
| Cookies/local storage/extensions | Storage Engine | 🟠 À RENFORCER | Import/export/backup/clone atomiques | GAP-006 | P0 |
| Proxy assignment/testing | ProxyProvider | 🟠 À RENFORCER | Health, Vault et erreurs | GAP-009 | P0 |
| Profile operations/bulk | Profile Manager + CAMO-CORE | 🟡 PARTIEL | Idempotence, progression et rollback | GAP-027 | P1 |
| Templates/clone/history | Profile Manager | 🟡 PARTIEL | Versioning et migration | GAP-003/GAP-022 | P1 |
| Headful/headless/runtime | Runtime Manager | 🟡 PARTIEL | Qualification, launch gate et recovery | GAP-015 | P0 |
| CDP/Playwright/Puppeteer/Selenium | Automation Layer | 🟡 PARTIEL | Contrats, stop, timeout et audit | GAP-017 | P1 |
| Local API/SDK | API Layer | 🟡 PARTIEL | OpenAPI, SDK versioning et rate limits locaux | GAP-021 | P1 |
| Cloud/Web App/Cloud Launch | Extension distante | ⚫ EXCLU du MVP | Architecture distante non livrée | Roadmap P2 | P2 |
| Sharing/session lock/RBAC | Workspace distant | ⚫ EXCLU du MVP | Gouvernance distante non livrée | Roadmap P2 | P2 |
| Marketplace proxy/profils | Service commercial | ⚫ EXCLU | Non prévu | Ne pas ajouter | — |

## 16. Nouvelles exigences issues de l’audit

| ID | Nom | Exigence d’acceptation |
| --- | --- | --- |
| GAP-001 | Profile Manager lifecycle | Create, Read, Update, Archive, Restore, Clone, Import, Export, Search, Filter, Sort, Tags, Groups, History, Versioning, Health, Recovery, Backup et Migration possèdent validation, erreur, audit, rollback lorsque nécessaire et tests. |
| GAP-002 | Notes/Custom Fields | Notes et champs `text/number/boolean/select` persistés, validés serveur, auditables et exportables sans secret. |
| GAP-003 | Templates versionnés | Un template est versionné, validé, cloné en profil indépendant et migrable sans copier de secret. |
| GAP-004 | Clone isolé | Le clone reçoit un nouvel ID, stockage, session et audit ; locks, tokens, secrets et processus ne sont jamais copiés. |
| GAP-005 | Restore transactionnel | Archive/restore préserve l’ID logique et l’audit, refuse les collisions et récupère proprement après crash. |
| GAP-006 | Storage Lifecycle | Cookies, cache, LocalStorage, SessionStorage, IndexedDB, Cache Storage, Service Workers et extensions sont isolés, testés, importables/exportables selon le format et restaurables. |
| GAP-009 | Proxy Health | HTTP/HTTPS/SOCKS5, connectivité, latence, erreurs et association profil sont testés ; les credentials restent dans SystemVault. |
| GAP-010 | Identity Lifecycle | Chaque catégorie possède configuration, observation, support runtime, validation, version, historique, drift et comportement après restart/update/clone/import/restore. |
| GAP-011 | Rendering Regression | Fixtures Canvas/WebGL/Audio/ClientRects et fonts sont rejouées sur le runtime qualifié ; tout écart produit un diagnostic redacted. |
| GAP-012 | Device Capability | CPU/RAM/GPU/media/touch sont distingués entre réel, configuré, observé et non supporté ; aucune capacité impossible n’est déclarée complète. |
| GAP-014 | WebRTC Privacy | Permissions, ICE, STUN/TURN, media capabilities et leak status sont testés ; les adresses sensibles sont redacted partout. |
| GAP-015 | Runtime Lifecycle | Installation, qualification, health, launch, stop, crash recovery, rollback, migration, update et downgrade lorsque supporté sont couverts. |
| GAP-017 | Automation Contract | API locale, CDP/Playwright/Puppeteer/Selenium, timeout, annulation, stop global, idempotence et audit sont contractuels. |
| GAP-021 | API Contract | Toutes les routes `/api/v1` définissent auth, redaction, limites, pagination, erreurs, `X-Request-ID` et `correlation_id` pour les mutations. |
| GAP-023 | Observability | Toute anomalie peut être détectée, enregistrée, expliquée, affichée et revalidée après correction ; aucun Detection Score. |
| GAP-025 | Native Security | SystemVault, permissions natives, suppression sécurisée et flux backup sont vérifiés sur hôte Ubuntu natif, séparément du sandbox. |
| GAP-026 | OS Matrix | Chaque OS annoncé possède une matrice runtime, stockage, permissions, réseau, performance et recovery réellement exécutée avant promotion. |
| GAP-027 | Bulk Operations | Chaque opération en lot est bornée, idempotente, annulable, observable par profil, récupérable après erreur et couverte par tests de concurrence. |
| GAP-028 | Secure Deletion and Recovery | La suppression d’un profil, de ses données isolées et de ses exports suit une politique documentée, vérifiable et compatible avec les mécanismes de recovery autorisés. |

Pour chaque exigence, le contrat précise objectif, comportement, composants, API, stockage, UI, configuration, validation, persistance, versioning, migration, import/export, clone, backup/restore, logs, audit, tests unitaires/intégration/régression/multi-runtime, erreurs, crash recovery, OS, dépendances et critères d’acceptation.

## 17. Architecture mise à jour

```
React Dashboard
      ↓ loopback, Bearer mémoire seule, redaction
Local API / Contract Layer
      ↓
BrowseForge / Core Go — unique control plane
      ├── Profile Manager + Notes/Custom Fields/Templates
      ├── Environment & Consistency + Fingerprint Lifecycle
      ├── Rendering, Hardware, Device & Permission Diagnostics
      ├── Storage / Cookie / Extension Lifecycle
      ├── ProxyProvider + SystemVault references
      ├── Runtime Manager + qualification/recovery
      ├── Automation Controller + queue/locks/timeouts
      ├── Backup/Restore + operation journal
      └── Audit / Health / Drift / Observability
      ↓
SQLite métier + Browser Storage isolé + SystemVault + Qualified Runtime
```

Aucun second control plane, écrivain React ou backend Node parallèle n’est autorisé. Chaque module déclare ses dépendances, son contrat, son stockage, ses migrations, son statut et ses preuves.

## 18. Roadmap mise à niveau

| Priorité | Lots | Conditions de sortie |
| --- | --- | --- |
| P0 | T08 Core reliability, T09 profile writes, T10 data cycle, T11 hardening, T12 local MVP, T13 identity/environment, T14 runtime qualification | Code, tests positifs/négatifs, isolation, redaction, audit, démonstration et preuve réelle |
| P1 | T15 automation contrôlée, T16 UX, T19 health/drift, T20 OpenAPI/SDK, T21 gap closure | Contrats, tests API/E2E, régression, limites runtime et preuve par module |
| P2 | T17 multi-OS non Ubuntu, T18 collaboration, Cloud/Web App, RBAC, 2FA, Android, Enterprise | Étude séparée, architecture, sécurité, coûts et preuves natives avant implémentation |
| Exclu | CAPTCHA solving/bypass, anti-bot/anti-fraude bypass, credential/cookie theft, fraude, contournement de sécurité et comportement destiné à tromper un mécanisme de sécurité | Aucun code, endpoint, écran ou prototype caché |

Une fonctionnalité ne passe au statut `🟢 COMPLET` qu’après satisfaction de la checklist ci-dessous.

## 19. Checklist de validation finale

| Contrôle | Question obligatoire |
| --- | --- |
| Objectif | Le cas d’usage et la limite d’autorisation sont-ils écrits ? |
| Architecture | Le Core Go reste-t-il l’unique control plane et écrivain ? |
| Contrat | Le schéma, la version et les états supportés sont-ils définis ? |
| API/Storage/UI | Les trois couches nécessaires existent-elles et sont-elles cohérentes ? |
| Persistance | Restart, update, clone, import/export et backup/restore sont-ils testés ? |
| Validation | Les valeurs configurées, supportées et observées sont-elles comparées ? |
| Sécurité | Aucun secret, cookie, token, credential ou donnée sensible inutile ne fuit-il ? |
| Isolation | Deux profils A/B démontrent-ils l’absence de contamination ? |
| Runtime | Compatibilité, lancement, arrêt, crash recovery et rollback sont-ils prouvés ? |
| Tests | Nominal, négatif, redaction, API, E2E, régression et `-race` pertinent passent-ils ? |
| Audit | Mutations, validations, drift et erreurs ont-ils un `correlation_id` et un audit redacted ? |
| Preuve | Commandes, versions, sorties, manifeste et limites sont-ils reproductibles ? |
| Statut | Le statut annoncé correspond-il à la preuve, sans déclarer `COMPLET` sur une simple UI ? |

### Classification initiale après audit

Cette classification est un état de travail du CDC, pas une déclaration de livraison :

| Statut | Nombre initial | Interprétation |
| --- | --- | --- |
| 🟢 COMPLET | 2 | T05 bootstrap et T06 catalogues lecture seule avec preuves disponibles |
| 🟡 PARTIEL | 17 | Une partie existe, mais le cycle complet ou les tests manquent |
| 🟠 À RENFORCER | 11 | Architecture présente, détails/tests/recovery insuffisants |
| 🔵 À VÉRIFIER | 4 | Dépendance à une preuve native ou runtime non encore qualifiée |
| 🔴 MANQUANT | 5 | Fonction non livrée dans le code actuel |
| ⚫ EXCLU | 9 | Cloud/abuse/bypass et fonctions hors MVP selon la matrice |

Les 20 lacunes les plus importantes sont sélectionnées parmi GAP-001 à GAP-028 selon leur impact : Profile Manager, templates, clone, restore, storage lifecycle, proxy health, identity lifecycle, rendering regression, device capability, WebRTC, runtime lifecycle, automation contract, API contract, observability, native security et OS matrix. Les exigences prioritaires sont détaillées dans la section 16, y compris GAP-027 et GAP-028.

Les fonctionnalités GoLogin non couvertes dans le MVP restent Cloud/Web App/Cloud Launch, partage distant, RBAC, 2FA distant, marketplace, multi-OS non prouvé et services commerciaux distants. Elles ne sont pas déclarées complètes parce que ForgeLocal demeure local-first et que les preuves natives ou l’architecture distante ne sont pas encore établies.

## 20. Règle absolue de vérité d’implémentation

Le cahier distingue obligatoirement ce qui est réellement implémenté, partiellement implémenté, spécifié, prévu dans la roadmap, testé, démontré ou inexistant. Une interface, un bouton, une route API, une structure Go, une migration, un commentaire, un nom de module ou une spécification ne suffit jamais à déclarer une fonctionnalité complète.

Les statuts autorisés sont exactement `🟢 COMPLET`, `🟡 PARTIEL`, `🟠 À RENFORCER`, `🔵 À VÉRIFIER`, `🔴 MANQUANT` et `⚫ EXCLU`. `🟢 COMPLET` exige le mécanisme réel, le contrat, Core, API si nécessaire, stockage, UI si nécessaire, configuration, validation, persistance, versioning, migration, import/export, clone, backup/restore si pertinent, audit, logs redacted, tests positifs/négatifs, concurrence si pertinente, crash recovery, isolation et preuve reproductible.

La parité GoLogin n’est jamais déclarée sur la base d’un nom similaire. Chaque capacité suit la chaîne : `GoLogin capability → équivalent ForgeLocal → implémentation réelle → API → stockage → runtime → tests → evidence`. Si une étape manque, le CDC écrit `pas de parité déclarée` et conserve un statut inférieur à `🟢 COMPLET`.

## 21. Matrice de complétude obligatoire

| GoLogin capability | ForgeLocal module | Core | API | Storage | UI | Runtime | Tests | Evidence | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Capacité benchmarkée | Module exact | `yes/no` vérifiable | `yes/no` vérifiable | `yes/no` vérifiable | `yes/no` vérifiable | `yes/no` vérifiable | commandes/résultats | manifeste/hash redacted | statut autorisé |

Aucune ligne ne peut être `🟢 COMPLET` si une colonne nécessaire n’est pas réellement validée. Les états `CONFIGURED`, `SUPPORTED`, `OBSERVED`, `UNSUPPORTED`, `RUNTIME_DEFINED`, `WARNING` et `FAIL` restent séparés des statuts de livraison.

## 22. Exigences d’implémentation par module

Le Profile Manager couvre réellement Create, Read, Update, Archive, Restore, Clone, Import, Export, Search, Filter, Sort, Tags, Groups, History, Versioning, Health, Recovery, Backup et Migration. Chaque opération possède validation serveur, erreurs explicites, audit, `correlation_id`, rollback lorsque pertinent et tests positifs/négatifs/concurrence.

Le clone reçoit un nouvel ID, un nouveau stockage isolé, une nouvelle session et un nouvel audit. Tokens, secrets, locks, processus, credentials et identifiants de session sensibles ne sont jamais copiés.

Le Storage Engine isole Cookies, LocalStorage, SessionStorage, IndexedDB, Cache, Cache Storage, Service Workers, Extensions, Browser Preferences, History et Bookmarks. Les scénarios Profile A/Profile B, restart, clone, import, export, backup, restore et crash recovery sont obligatoires ; une séparation de répertoires seule ne suffit pas.

Le ProxyProvider couvre HTTP, HTTPS et SOCKS5, association profil, validation, connectivité, health, latence et erreurs. Les credentials restent dans SystemVault et sont absents de l’API, des logs, de l’audit, des exports, des fixtures et des preuves.

Le Browser Identity/Environment Engine couvre Generate Environment Profile, New QA Environment, Rendering Profile, configuration, observation, validation, version, historique et drift. Chaque attribut distingue `configured`, `supported`, `observed` et `unsupported`. Une propriété enregistrée seulement dans SQLite ne doit jamais être déclarée réellement modifiée par le runtime.

Canvas, WebGL, Audio, ClientRects et Fonts disposent de fixtures reproductibles exécutées sur le runtime qualifié. Tout écart produit un diagnostic redacted, la version du fixture et les informations runtime disponibles ; aucun score de détection n’est produit.

Le Runtime Manager couvre Install, Detect, Qualify, Health, Launch, Stop, Crash Recovery, Rollback, Update, Downgrade lorsque supporté, Migration, Headful, Headless, CDP, Playwright, Puppeteer et Selenium. Un runtime n’est jamais déclaré `qualified`, `supported` ou `launchable` avant un test réel démontrant cette propriété. Camoufox reste `🔵 À VÉRIFIER`, non lancé et non importé.

L’Automation Layer exige timeout, contexte annulable, global stop, idempotence, audit et `correlation_id`. Elle ne fonctionne que lorsqu’un runtime est qualifié. Les Bulk Operations imposent concurrence bornée, idempotence, annulation, global stop, état par profil, recovery et audit.

Backup/Restore vérifie archive valide, métadonnées, intégrité, chiffrement, identité du profil, collision, corruption, recovery transactionnel, crash recovery et audit. AES-256-GCM, AAD, nonce aléatoire et aucune crypto maison restent obligatoires.

## 23. API, secrets, isolation et observabilité

L’API est versionnée sous `/api/v1`. Chaque route définit authentification, autorisation, validation d’entrée, limites, pagination, erreurs, redaction, `X-Request-ID` et `correlation_id`. OpenAPI ou équivalent est obligatoire avant toute déclaration de maturité.

Le bootstrap utilise un Bearer token en mémoire seule, un code unique, une expiration et une invalidation sur `401`. Le Core écoute uniquement sur `127.0.0.1` et `::1`, et le refus hors loopback est testé explicitement.

Aucun secret ne doit apparaître dans le DOM, les réponses API, les logs, l’audit, SQLite non sécurisé, les archives, les preuves, les captures, les fixtures ou les tests. Les sentinelles de test vérifient cette règle.

Chaque anomalie doit pouvoir être détectée, enregistrée, expliquée, affichée, corrigée lorsque possible et revalidée. Aucun `Detection Score`, `Stealth Score`, `Ban Probability` ou prétention à savoir si un environnement trompe un système externe n’est autorisé.

## 24. Procédure obligatoire pour chaque GAP

Chaque GAP suit le cycle : lire le contrat, inspecter le code, identifier l’existant réel, identifier le manque, implémenter uniquement le périmètre, écrire les tests positifs et négatifs, ajouter les tests de concurrence si nécessaires, exécuter les tests, exécuter les scans, produire les preuves redacted, vérifier les preuves et mettre à jour le statut.

La transition obligatoire est : `specification → implementation → test → execution → evidence → verification → status`. Aucun passage direct de spécification à `🟢 COMPLET` n’est autorisé. Si seule l’UI existe, le statut est `🟡 PARTIEL`. Si seule la spécification existe, le statut est `🔵 À VÉRIFIER`. Si rien n’existe, le statut est `🔴 MANQUANT`. Une fonction volontairement hors MVP est `⚫ EXCLU`.

## 25. T08 — Première tâche autorisée

T08 ne traite que le `LaunchManager / SessionManager` en Go : queue bornée, limite globale, sérialisation par profil, timeout, contexte annulable, cleanup idempotent, crash recovery et audit redacted.

T08 ne lance aucun navigateur, ne lance pas Camoufox, ne lance pas de runtime réel, n’utilise pas de proxy, ne modifie pas l’UI, n’implémente pas backup/restore/import et ne produit aucune release.

Les critères AC-CAMO-01 à AC-CAMO-05 sont obligatoires : demandes concurrentes d’un même profil sérialisées ou refusées explicitement et auditées ; limite globale respectée ; expiration libérant les ressources ; crash simulé sans session fantôme ni lock bloqué ; erreur d’attachement sans état bloqué ; tests sous `go test -race` et absence de secrets dans les logs/audits.

## 26. Outils et rapport obligatoire

Les versions de référence restent Go `1.25.13` avec `GOTOOLCHAIN=local`, Node 22, pnpm, Playwright, Gitleaks 8.18.4, Gosec, Govulncheck, GolangCI-Lint et SQLite. Aucune version ne change sans décision explicite.

Après chaque tâche, le rapport contient : `TASK`, `GAP`, `IMPLEMENTED`, `NOT IMPLEMENTED`, `FILES CHANGED`, `API CHANGED`, `DATABASE CHANGED`, `UI CHANGED`, `TESTS WRITTEN`, `TESTS EXECUTED`, `TEST RESULTS`, `RACE RESULT`, `SECURITY SCAN`, `EVIDENCE`, `LIMITATIONS`, `CURRENT STATUS` et `NEXT ALLOWED STEP`.

## Références de cadrage

[1]: https://github.com/nczz/BrowseForge "BrowseForge"

[2]: https://github.com/TechQaiser/persona-studio "Persona Studio"

[3]: https://github.com/ProxyShard/ShardBrowser "ShardBrowser / ShardX"

[4]: https://github.com/CloakHQ/CloakBrowser "CloakBrowser"

[5]: https://github.com/zhom/donutbrowser "DonutBrowser"

[6]: https://gologin.com/features/ "GoLogin — benchmark de marché"

[7]: https://support.gologin.com/en/articles/14810056-profile-fingerprint-settings "GoLogin — Profile fingerprint settings"

[8]: https://support.gologin.com/en/articles/14854407-advanced-profile-settings "GoLogin — Advanced profile settings"

[9]: https://gologin.com/docs/api-reference/introduction/quickstart "GoLogin — API Quickstart"

[10]: https://support.gologin.com/en/articles/14854406-profile-settings "GoLogin — Profile settings"

[11]: https://support.gologin.com/en/articles/14810058-what-s-safe-to-change "GoLogin — What’s safe to change"

[12]: https://support.gologin.com/en/articles/15647978-overview "GoLogin — Cloud Browser overview"

[13]: https://support.gologin.com/en/collections/3069861-profiles "GoLogin — Profiles collection"