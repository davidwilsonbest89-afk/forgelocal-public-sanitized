# Addendum v2.3 — Provenance des composants et exclusion GoLogin

**Identifiant :** `FL-ADD-2.3-20260814`

**Statut :** addendum de gouvernance du cahier des charges ForgeLocal.

**Objet :** encadrer les composants dont les détenteurs ont autorisé l’examen, le portage ou l’intégration, et exclure explicitement GoLogin de toute réutilisation.
**Invariants :** Core Go unique, dashboard React client, RC BACK-01 gelé, pilote suspendu, `SCAN_BLOCKED_UNKNOWN` et `PUBLIC_RELEASE_BLOCKED` inchangés.

Cet addendum complète le cahier des charges v1.0 et l’addendum v2.1. Il ne constitue pas une autorisation de release, d’activation de runtime, d’écriture UI, de backup ou de restauration.

## 1. Décision explicite sur GoLogin

GoLogin est une **référence de marché publique uniquement**. Il est enregistré avec la décision `écarter` et le statut de droits `denied`. Aucun code, asset, dépendance, runtime, documentation interne, capture non publique, schéma d’API, fichier, flux spécifique ou élément propriétaire GoLogin ne peut rejoindre ForgeLocal.

La comparaison publique de positionnement reste limitée au cloud, à la collaboration, au support et à l’expérience SaaS. Elle ne justifie jamais la reproduction d’une interface, d’un contrat ou d’une logique non fournie par une source autorisée.

> **Règle clean-room :** toute fonctionnalité ForgeLocal doit être justifiée par le cahier des charges, une exigence utilisateur légitime ou une source explicitement autorisée. Elle ne doit pas dériver d’un matériau propriétaire GoLogin.

## 2. Registre canonique de provenance

`docs/component-rights-register.json` est la **source de contrôle CI**. `docs/COMPONENT_RIGHTS_REGISTER.md` en est la vue humaine synchronisée. Les emails et autres preuves originales restent dans un espace privé ; Git ne contient que des identifiants redacted, les décisions et les révisions consommées.

Une entrée intégrée doit déclarer l’URL ou l’origine, une révision exacte, la portée d’autorisation, les exclusions, un responsable et une référence redacted vers la preuve. Une entrée `not_required` n’est acceptable que pour une source explicitement déclarée first-party/interne, avec propriétaire et révision de référence. Une source tierce ne peut jamais contourner une autorisation en utilisant `not_required`.

## 3. Séparation des responsabilités

BrowseForge/Core Go reste l’unique control plane et le seul écrivain de profils, sessions, locks, ports et références de secrets. Camoflox peut seulement alimenter un portage sélectif de fiabilité vers Go : queue, limite globale, locks par profil, timeout, annulation, cleanup, contrôle d’endpoint et tests de concurrence. Aucun serveur Node/Electron propriétaire de profils ou sessions n’est admis.

La **CSP** appartient au dashboard ou à la bordure HTTP qui le distribue. Si le Core sert l’interface, il peut émettre les en-têtes CSP, mais cette responsabilité ne relève pas du `LaunchManager`. Le Core conserve ses protections propres : loopback, authentification, rate limiting, validation des entrées et journalisation redacted.

Persona Studio, DonutBrowser, ShardBrowser/ShardX et CloakBrowser restent des sources autorisées uniquement lorsque leur révision exacte est sélectionnée dans le registre et qu’elles passent les gates applicables. Camoufox demeure un runtime candidat non lançable tant que sa qualification indépendante n’est pas terminée. Les fonctionnalités de camouflage, de CAPTCHA, d’humanisation ou de contournement abusif sont exclues : aucun import, asset, runtime ou dépendance de ce type ne peut être ajouté sans nouvelle revue de périmètre.

## 4. Gates de provenance avant intégration

Un composant rejoint le build uniquement après `PROV-01` à `PROV-07` : droits sur la révision exacte, inventaire des dépendances et assets, conformité à l’architecture Core unique, sécurité, fiabilité, SBOM/notices et valeur produit testable. La réussite de `PROV-01` ne dispense jamais des gates sécurité, fiabilité ou release.

## 5. Contrôle CI et release

Le contrôle CI exécute `node scripts/check-component-rights.mjs` puis `node scripts/test-component-rights.mjs`. Il valide le registre JSON, les statuts, les révisions, les responsables, les portées, les inventaires de dépendances et les preuves redacted des composants intégrés, ainsi que les empreintes des manifestes d’entrées build. Les scénarios `authorized`, `denied`, `unknown`, révision absente, `not_required` non first-party et propriétaire absent restent testés par fixture isolée. Les statuts suivent strictement la règle suivante :

| Statut | Règle CI |
|---|---|
| `authorized` | Autorisé seulement si révision, portée, responsable, preuve redacted et dépendances applicables sont déclarés. |
| `not_required` | Autorisé uniquement si la source est first-party/interne, propriétaire assigné et révision de référence déclarée. |
| `denied` | Bloquant. |
| `unknown` | Bloquant. |
| Absente du registre | Bloquant pour toute nouvelle source, dépendance, asset ou runtime contrôlé. |

Le contrôle examine les inputs de build et les chemins d’intégration contrôlés, non la simple présence du mot « GoLogin » dans la documentation de benchmark approuvée. Toute occurrence de marqueur GoLogin dans les manifestes ou le code d’intégration déclenche une revue bloquante. Une future release devra joindre le registre redacted, SBOM, notices, hashes des composants, scans du workspace et de l’archive extraite, manifestes runtime et revue indépendante.

## 6. Décision opérationnelle

Le développement peut poursuivre les composants autorisés **après** passage des gates applicables. Le bootstrap local de session et CAMO-CORE-01 peuvent progresser en parallèle, sans activer de mutation UI ou de runtime. Le RC gelé, le pilote suspendu, `SCAN_BLOCKED_UNKNOWN`, les gates SystemVault et `PUBLIC_RELEASE_BLOCKED` restent inchangés.
