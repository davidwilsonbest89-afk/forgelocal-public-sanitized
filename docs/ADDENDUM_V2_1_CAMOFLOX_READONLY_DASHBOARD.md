# Addendum v2.1 — Camoflox audité et dashboard connecté en lecture seule

**Identifiant :** `FL-ADD-2.1-20260814`
**Date :** 14 août 2026
**Auteur :** Manus AI
**Statut :** addendum interne de référence ; aucune autorisation de publication publique
**Document de base :** [`docs/CAHIER_DES_CHARGES_FORGELOCAL.md`](CAHIER_DES_CHARGES_FORGELOCAL.md), identifiant `FL-CDC-1.0-20260814`
**Attestation :** l’empreinte SHA-256 et le commit de contenu sont consignés dans `docs/ADDENDUM_V2_1_CAMOFLOX_READONLY_DASHBOARD.manifest.json`.

## 1. Objet, portée et invariants de sécurité

Cet addendum précise deux chantiers produit indépendants qui peuvent progresser en parallèle : l’audit suivi d’un portage sélectif de modules Camoflox vers le Core Go, et le raccordement progressif du dashboard React à des lectures locales redacted du Core. Il prévaut uniquement sur les passages contradictoires du document de base.

> **Invariants non négociables :** BrowseForge/Core Go reste l’unique control plane et l’unique écrivain de l’état métier. React est un client d’API sans base locale métier parallèle. Aucun runtime, secret, token persistant, chemin sensible ou valeur proxy ne doit être exposé au navigateur.

Cet addendum ne modifie ni l’archive RC BACK-01 gelée, ni le pilote suspendu, ni `SCAN_BLOCKED_UNKNOWN`, ni les cinq gates de `PUBLIC_RELEASE_BLOCKED`. Il n’autorise aucune nouvelle release BACK-01, aucun lancement Camoufox, ni aucune écriture UI métier.

## 2. Réutilisation sélective et audit de Camoflox

Le propriétaire de ForgeLocal a déclaré détenir les droits nécessaires sur `camoflox-FINAL.zip`. Cette déclaration autorise l’examen du code propriétaire Camoflox, mais ne couvre pas automatiquement ses dépendances tierces, binaires, packages npm, assets, Playwright ou Camoufox. Chaque élément reste soumis à un inventaire de provenance, de dépendances et de licences avant toute distribution.

Avant tout changement de code, le dépôt doit contenir `docs/CAMO_CORE_01_MODULE_REGISTER.md`. Le registre est versionné et chaque ligne contient le chemin source audité, son hash SHA-256, le commit Camoflox de référence lorsqu’il existe, la responsabilité, les dépendances directes et transitives, la décision, la justification, le propriétaire de revue, la date de décision, le risque, les tests exigés et le statut.

| Décision | Sens | Condition minimale |
|---|---|---|
| `porter` | Porter le comportement validé vers Go | Contrat, tests de concurrence, timeout, annulation, audit et reprise après crash. |
| `réimplémenter` | Reproduire l’exigence avec une conception ForgeLocal | Contrat écrit, tests de non-régression et absence de seconde source de vérité. |
| `écarter` | Ne pas intégrer l’élément | Justification versionnée : dépendance, licence, périmètre, sécurité ou risque produit. |

| Module ou pattern Camoflox | Décision initiale | Destination ForgeLocal |
|---|---|---|
| Sémaphore global et queue de lancement | `réimplémenter` | `LaunchManager` Go avec `context.Context`, timeout, annulation, métriques de file et journal d’audit. |
| Sérialisation par profil | `réimplémenter` | Lock par `profile_id`, rejet explicite et journalisé des doublons. |
| Locks de session et nettoyage des zombies | `porter` comme contrat, `réimplémenter` en Go | `SessionManager` avec état Core, audit et réconciliation après crash. |
| Allocation et libération de ports | `réimplémenter` | Stratégie de réservation liée au lancement, définie et testée par runtime. |
| Déconnexion et libération de ressources | `réimplémenter` | Transition de session atomique, nettoyage idempotent et réconciliation au redémarrage. |
| CSP, token API, CORS local et rate limiting | `porter` comme exigences | Serveur API Go unique, loopback par défaut, sans serveur Node parallèle. |
| AES-GCM, rotation de clés, références de secret | `réimplémenter` | Contrats BACK-01 et SystemVault ForgeLocal, sans crypto maison ni stockage clair. |
| Tests de timeout, crash et concurrence | `porter` comme scénarios | Tests Go unitaires, intégration, `-race` et récupération. |
| `buildCamoufoxFingerprint` et réglages Camoufox | `écarter` du Core | Runtime candidat séparé ; aucune promesse de non-détection. |
| `camoufox-js`, Playwright et binaire Camoufox | `écarter` de la release Core | Inventaire, qualification runtime et décision de distribution séparés. |

## 3. Exigences de fiabilité CAMO-CORE-01

Le portage ne doit jamais introduire un processus Node ou Electron gérant profils, sessions, locks ou ports. Chaque session Core doit porter un identifiant, un profil, un runtime, un état, un instant de lancement, un `correlation_id` métier, un handle de processus ou endpoint et une raison de fin redacted.

L’allocation de port est une ressource contrôlée. Pour chaque runtime, le registre doit définir la stratégie de réservation, son propriétaire, sa durée maximale et le protocole de libération. Le Core ne doit pas seulement observer un port libre puis le relâcher avant lancement. Lorsqu’un handle de réservation est possible, il est conservé jusqu’à l’attachement prévu ; sinon, le conflit est traité comme une erreur récupérable. Dans tous les cas, le Core vérifie, dans un timeout borné, que le processus démarré possède réellement son endpoint. Tout échec réconcilie de façon idempotente le lock, la ressource et l’état de session, puis crée un événement d’audit redacted.

| ID | Critère d’acceptation |
|---|---|
| `AC-CAMO-01` | Deux demandes simultanées pour le même profil n’ouvrent qu’une session ou produisent un refus explicite et journalisé. |
| `AC-CAMO-02` | La limite globale est respectée ; une demande en attente expire visiblement et libère toutes ses ressources. |
| `AC-CAMO-03` | Après arrêt brutal du Core ou du runtime, la reprise réconcilie locks, états SQLite et ressources sans session fantôme. |
| `AC-CAMO-04` | Un conflit de port ou endpoint absent ne laisse ni lock, ni ressource de port, ni état de session bloqué. |
| `AC-CAMO-05` | Les scénarios de concurrence passent sous `go test -race` et n’exposent aucun secret dans les logs ni l’audit. |

## 4. Client API local en lecture seule

Le client lecture seule peut progresser en parallèle de CAMO-CORE-01 dès que le contrat API correspondant est stabilisé. Il ne démarre aucun runtime et n’écrit aucune donnée métier. Il remplace progressivement les données de démonstration par des réponses redacted du Core Go.

Chaque endpoint de lecture est documenté par un schéma JSON versionné, une liste blanche de champs rendus, des exemples redacted, une politique d’erreur et des tests de contrat. Les collections utilisent une pagination par curseur avec un paramètre `limit` soumis à un plafond serveur documenté ; le Core renvoie un curseur suivant opaque lorsque des résultats supplémentaires existent. Le serveur ne fournit jamais de secret, cookie, stockage navigateur, chemin absolu, valeur proxy, URL signée ni détail système non nécessaire au parcours.

| Endpoint cible | But UI | Données interdites |
|---|---|---|
| `GET /api/v1/health` | Disponibilité, version API et état du Core | Token, chemins absolus et détails système sensibles. |
| `GET /api/v1/dashboard/summary` | Compteurs de profils et runtimes, état coffre et ressources mesurées | Secrets, valeurs proxy et données de profil non nécessaires. |
| `GET /api/v1/profiles` | Registre filtrable, trié et paginé | Cookies, stockage navigateur, référence secrète et chemin non assaini. |
| `GET /api/v1/profiles/{id}` | Panneau de profil sélectionné | Valeur de coffre, données navigateur et identifiants proxy. |
| `GET /api/v1/groups` | Groupes et héritage de politique | Secret de groupe et configuration confidentielle. |
| `GET /api/v1/runtimes` | Statut et provenance runtime résumée | URL signées, jetons et détails non nécessaires. |

Les identifiants techniques de requête sont générés ou acceptés par le Core et retournés dans des en-têtes de réponse documentés, par exemple `X-Request-ID`. Le `correlation_id` reste réservé aux opérations métier et à leur audit. Le client UI peut associer son diagnostic à l’identifiant de requête sans le présenter comme une preuve d’écriture métier.

## 5. Session locale, token et états UI

L’API est liée au loopback par défaut. Toute authentification Bearer est fournie uniquement à travers un mécanisme de bootstrap local authentifié défini par le Core. Le token ne vit que dans la mémoire du processus React pendant une session locale courte ; il ne doit jamais être persisté dans `localStorage`, `sessionStorage`, IndexedDB, cache persistant, URL, service worker, logs, traces, export, analyse d’usage ou outil d’analytics. Les en-têtes `Authorization` sont exclus des journaux applicatifs et de diagnostic.

React ne conserve aucun cache métier persistant. Un cache mémoire de courte durée est admis seulement si le contrat définit sa durée, s’il est invalidé à toute reconnexion Core et si l’interface abandonne les données affichées lorsqu’une réponse est invalide, une erreur réseau, un `401` ou un changement de données est détecté. Les écrans Profils, Groupes, Runtimes, Proxys, Sauvegardes et Audit partagent les états `chargement`, `vide`, `erreur`, `réessai`, `Core indisponible` et `données de démonstration`.

> Une action indisponible est désactivée et explique la précondition manquante. Elle ne doit jamais ressembler à une mutation réelle.

## 6. Frontière avant les écritures UI

La lecture seule est autorisée après validation du contrat concerné. Création, modification, archivage, lancement, isolation, arrêt, proxy, backup, restauration et import restent indisponibles jusqu’à validation des contrats Core dédiés.

| Opération future | Précondition Core obligatoire | Preuve UI attendue |
|---|---|---|
| Créer, éditer ou archiver un profil | Validation serveur, transaction SQLite, audit et réponse canonique | `correlation_id`, retour du profil canonique et erreur actionnable. |
| Lancer, isoler ou arrêter | CAMO-CORE-01, runtime autorisé, lock et état de session | États attente/en cours/réussi/échoué, jamais de lancement implicite. |
| Proxy | `ProxyProvider`, secret dans le coffre et test redacted | Aucun secret dans le formulaire et connectivité redacted. |
| Backup ou restauration | BACK-01, coffre validé et contrats de reprise | Confirmation explicite, statut de tâche et audit redacted. |
| Import JSON → SQLite | Migration P1, préimage chiffrée, dry-run et parité | Résumé des conflits, consentement explicite et rapport redacted. |

Camoufox reste affiché comme **runtime candidat non lançable** jusqu’à sa qualification indépendante complète. Il n’est ni approuvé, ni distribué, ni utilisable comme preuve pour un autre runtime.

## 7. Séquence de réalisation et non-régression release

Les séquences A et C peuvent commencer en parallèle. La séquence C ne dépend pas de la livraison complète de CAMO-CORE-01 : elle débute dès la stabilisation et les tests de contrat des endpoints lecture seule correspondants. Les mutations UI ne commencent qu’après leurs préconditions Core respectives.

| Séquence | Travail | Dépendance | Sortie attendue |
|---|---|---|---|
| A | Créer et valider le registre CAMO-CORE-01 | Aucune | Décision par module et dépendance, revue technique et sécurité. |
| B | Livrer CAMO-CORE-01 | A | Queue, locks, cleanup, stratégie de ports Go et tests de concurrence verts. |
| C | Livrer le client lecture seule | Contrats API lecture seule stabilisés | Vue d’ensemble et profils alimentés par réponses redacted. |
| D | Ajouter Groupes et Runtimes | C | États communs de chargement, erreur, réessai et absence de cache persistant. |
| E | Définir les contrats de mutation | B et contrats dédiés | Schémas API, validation, audit et `correlation_id` avant tout bouton actif. |
| F | Continuer les gates de release séparément | Indépendante | Triage de scan, nouvelle chaîne si nécessaire, SystemVault natif et revue indépendante. |

Le candidat RC reste gelé ; le pilote reste suspendu ; `SCAN_BLOCKED_UNKNOWN`, `PUBLIC_RELEASE_BLOCKED` et les cinq gates publics demeurent inchangés. Toute future archive de release doit posséder une chaîne indépendante liant source, runtime, SBOM, manifeste, checksums, scans et preuves.
