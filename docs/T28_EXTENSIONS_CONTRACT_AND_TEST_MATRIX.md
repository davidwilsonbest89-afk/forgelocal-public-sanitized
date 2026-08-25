# T28 — Contrat des extensions locales contrôlées et matrice de tests

**Statut du document :** contrat préalable à l’implémentation.  
**Périmètre :** Core local uniquement ; aucun runtime navigateur, chargement ou exécution d’extension.

## 1. Objectif et frontières

T28 gère l’import explicite, l’inventaire, l’approbation, l’affectation, la mise à jour, le rollback, la révocation/quarantaine et la purge contrôlée de packages d’extension conservés localement. Une affectation `ready` signifie uniquement qu’une relation de configuration est validée ; elle ne déclenche aucun navigateur, proxy, processus enfant, moteur d’extension ou appel réseau.

Les formats acceptés sont exclusivement des archives ZIP locales contenant un fichier `manifest.json` à la racine. Les CRX, répertoires décompressés, URLs, téléchargements de store, scripts d’auto-update et packages fournis par Internet sont refusés hors périmètre.

## 2. Identités canoniques

| Concept | Règle contractuelle |
|---|---|
| Série | `extension_series_id`, identifiant ForgeLocal généré localement par le serveur ; il représente la lignée logique et non une identité navigateur. |
| Version | `extension_version_id`, identifiant ForgeLocal immuable créé à chaque import ; jamais réutilisé. |
| Package | SHA-256 exact du fichier ZIP importé, format `zip`, taille et manifest extrait normalisé. |
| Manifest | Métadonnées informatives : `name`, `version`, `manifest_version`, permissions, host permissions et matches de content scripts. Elles ne prouvent aucune identité runtime. |
| Version active | Au plus une version active par série ; une mise à jour crée une nouvelle ligne et archive l’ancienne dans la même transaction logique. |
| Affectation | Relation explicite entre une version approuvée et un profil existant ; aucune affectation implicite. |

Une clé ou un champ éventuel du manifest ne devient jamais un ID navigateur stable. T28 ne tente pas d’en déduire une compatibilité ou une identité de runtime.

## 3. Archive ZIP et parseur sûr

Les limites contractuelles initiales sont les suivantes : taille ZIP maximale de **100 MiB**, taille décompressée cumulée maximale de **200 MiB**, **500 entrées** au maximum et `manifest.json` limité à **1 MiB**. Une archive excédant une limite est rejetée avec `ARCHIVE_LIMIT_EXCEEDED` avant toute extraction de contenu non nécessaire.

Le parseur doit lire l’archive sans exécuter de code et sans extraire les fichiers du package dans un répertoire de travail. Il inspecte les noms, tailles déclarées, types d’entrées et CRC selon les capacités de la bibliothèque ZIP, puis lit uniquement `manifest.json` pour le contrat T28. Il refuse les chemins absolus, les chemins contenant `..` après normalisation, les séparateurs ou formes traversantes, les doublons de noms, les liens/symlinks détectables via les attributs Unix, les archives chiffrées, les archives corrompues, l’absence de manifest et tout JSON invalide ou non objet.

Le manifest conserve `permissions`, `optional_permissions`, `host_permissions`, `optional_host_permissions` et les `matches` des content scripts lorsqu’ils sont présents. Les valeurs sont normalisées en listes triées et dédupliquées pour comparer les acknowledgements, sans effacer les valeurs inconnues.

## 4. Classification des permissions

Toutes les permissions sont importables et autorisables : aucune denylist implicite ne peut refuser un import sur le seul nom d’une permission. Les permissions explicitement sensibles (`cookies`, `webRequest`, `webRequestBlocking`, `debugger`, `nativeMessaging`, `management`, `proxy`, `downloads`, `clipboardRead`) ainsi que `<all_urls>`, `*://*/*`, `file:///*`, toute permission d’hôte globale et toute valeur non reconnue produisent le signal `HIGH_RISK`. Une valeur non reconnue est conservée et classée `UNCLASSIFIED_HIGH_RISK`.

`HIGH_RISK` est un signal, pas un refus automatique. L’approbation exige simultanément l’acknowledgement exact de la liste normalisée visible et `accept_high_risk=true` lorsqu’un risque est présent. L’affectation se fait dans une mutation distincte.

## 5. Stockage local-first et compensation

Le ZIP est copié de manière atomique dans un store géré sous `--base-dir`, par exemple `extensions/objects/<deux-premiers-caractères>/<sha256>.zip`. Le chemin complet et le package ne sont jamais renvoyés par l’API ni écrit dans l’audit. Les blobs sont immuables et ne sont jamais remplacés.

Les séries, versions, états, permissions normalisées, affectations et événements d’audit résident dans une base SQLite dédiée sous `--base-dir`. Le Profile Store canonique `profile.json` n’est pas migré ; GAP-002 reste ouvert. Le repository obtient l’existence d’un profil par l’abstraction Profile Store existante.

La réservation logique d’une version et l’écriture SQLite sont transactionnelles. Le blob est d’abord écrit atomiquement ; si SQLite échoue ensuite, la ligne n’est pas publiée et le blob non référencé est conservé comme objet orphelin local à traiter par une purge explicite ou un diagnostic, jamais supprimé automatiquement au risque d’effacer un objet référencé par une autre transaction. Si le stockage échoue, aucune ligne d’affectation ou de version partielle n’est validée. Les erreurs doivent être compensées et exposées par un code stable, sans chemin utilisateur complet.

## 6. États et opérations

| Opération | Préconditions | Transition et invariants |
|---|---|---|
| Import | ZIP local valide et profil aucunement requis | crée une série ou rattache explicitement une série existante, crée une version `imported`, copie immuable du blob |
| LIST | loopback + authentification | projection catalogue redacted, sans bytes ni chemins complets |
| GET | loopback + authentification | détail redacted, sans bytes ni chemin complet |
| Approve | version `imported`, acknowledgement exact, high-risk accepté si requis | `imported → approved`, audit redacted |
| Assign | version `approved`, profil existant, non révoquée/quarantainée | relation `assigned`, puis état logique `ready`; aucune exécution |
| Update | nouvelle archive valide, même série explicitement indiquée | nouvelle version immuable ; au plus une active après commit logique |
| Rollback | cible antérieure approuvée, série cohérente, non révoquée | repointe la série, archive l’active, audit |
| Revoke/quarantine | version existante | interdit toute nouvelle affectation et désaffecte explicitement |
| Purge | version non affectée et non cible de rollback | suppression explicite du blob et des métadonnées selon politique, audit obligatoire |

Les mutations d’une même série ou d’un même profil sont sérialisées sous verrou de repository et transaction SQLite `BEGIN IMMEDIATE` ou mécanisme équivalent. En cas de conflit, l’opération échoue fermement avec `CONCURRENT_MUTATION` et aucune affectation partielle n’est créée. Les lectures ne produisent pas d’audit d’écriture.

## 7. API Core proposée

Les routes sont loopback-only, authentifiées par le mécanisme existant du Core et protégées par la validation d’origine existante. Les mutations refusent toute requête non loopback, sans bearer ou avec origine non autorisée. Les réponses utilisent les erreurs stables existantes et les codes T28 suivants : `INVALID_ARCHIVE`, `ARCHIVE_LIMIT_EXCEEDED`, `MANIFEST_INVALID`, `PERMISSION_ACK_REQUIRED`, `HIGH_RISK_ACK_REQUIRED`, `VERSION_NOT_APPROVED`, `PROFILE_NOT_FOUND`, `VERSION_REVOKED`, `SERIES_NOT_FOUND`, `CONCURRENT_MUTATION`, `STORAGE_FAILED`, `DATABASE_FAILED`, `PURGE_NOT_ALLOWED`.

| Méthode | Route | Résultat |
|---|---|---|
| `POST` | `/api/v1/extensions/import` | reçoit un upload local multipart ; réponse `201` avec IDs ForgeLocal, digest tronqué, taille, manifest redacted et état `imported` |
| `GET` | `/api/v1/extensions` | projection catalogue paginée et redacted |
| `GET` | `/api/v1/extensions/{seriesID}` | versions, états et métadonnées redacted |
| `POST` | `/api/v1/extensions/{versionID}/approve` | acknowledgement exact et high-risk éventuel |
| `POST` | `/api/v1/extensions/{versionID}/assign` | profil explicite ; crée l’affectation logique |
| `POST` | `/api/v1/extensions/{seriesID}/update` | importe une version suivante sans remplacer le blob |
| `POST` | `/api/v1/extensions/{seriesID}/rollback` | repointe vers une version approuvée |
| `POST` | `/api/v1/extensions/{versionID}/revoke` | révocation/quarantaine et désaffectation |
| `DELETE` | `/api/v1/extensions/{versionID}` | purge explicite si autorisée |

Les OpenAPI et handlers ne documentent ni bytes de package ni chemin utilisateur complet. Aucune route T28 ne doit appeler un lanceur navigateur, Chromium, Camoufox, proxy, store d’extensions ou processus externe.

## 8. Audit et redaction

Chaque mutation consigne UTC, action, `extension_series_id`, `extension_version_id`, digest tronqué, catégories de permissions, profil pseudonymisé, résultat et code d’erreur stable. Sont interdits dans les logs, réponses et exports de preuve : ZIP, contenu du package, manifest complet, chemin utilisateur complet, bearer, token, cookie, header libre, presse-papiers et payload sensible. Les noms et versions manifest sont des métadonnées contrôlées et ne doivent pas être utilisés pour fabriquer une identité runtime.

## 9. Matrice de tests obligatoire

| ID | Scénario | Résultat attendu |
|---|---|---|
| T28-UT-01 | ZIP synthétique valide avec manifest racine | import `201`, série/version ForgeLocal, état `imported` |
| T28-UT-02 | permissions sensibles et patterns larges | conservation exacte, `HIGH_RISK`, aucun refus d’import |
| T28-UT-03 | valeur inconnue | conservation, `UNCLASSIFIED_HIGH_RISK` |
| T28-UT-04 | approbation sans acknowledgement exact | refus stable, aucun état `approved` |
| T28-UT-05 | high-risk sans `accept_high_risk=true` | refus stable |
| T28-UT-06 | assign avant approval / profil absent | refus, aucune affectation |
| T28-UT-07 | update, active unique, rollback | blobs distincts et états cohérents |
| T28-UT-08 | restart repository | données SQLite relues identiquement |
| T28-UT-09 | concurrence import/approve/assign/rollback/revoke | déterministe, aucune double affectation ou état partiel |
| T28-UT-10 | échec stockage et SQLite injectés | compensation documentée, pas de référence orpheline publiée |
| T28-UT-11 | zip-slip, symlink, archive corrompue, manifest absent/invalide | refus avant mutation |
| T28-UT-12 | limites taille/nombre et hash incohérent | refus stable |
| T28-UT-13 | bearer absent, hors loopback, origine refusée | `401/403`, aucune mutation |
| T28-UT-14 | LIST/GET/audit | projections redacted, aucune fuite |
| T28-UT-15 | garde non-runtime | aucune invocation navigateur/processus externe dans les scénarios |

Les commandes minimales sont `git diff --check`, `go test -count=1 -race ./...`, `go vet ./...` et `go build ./...`, complétées par les tests T28 ciblés et les scans exigés par le mandat.
