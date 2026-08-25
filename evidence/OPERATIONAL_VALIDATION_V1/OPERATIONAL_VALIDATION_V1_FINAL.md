# FORGELOCAL_OPERATIONAL_VALIDATION_V1 — rapport final

**Branche :** `validation/operational-v1`  
**HEAD de publication documentaire :** `b1559ca53852c493ba15e4a06ad89b0c171c7938`  
**Base T28 découverte :** `053fd8f9612eb028c04983e0a23d311a6a099e29`  
**Périmètre :** Core local, Dashboard local, bases et données synthétiques temporaires uniquement.

## Conclusion

La campagne opérationnelle post-T28 a été exécutée séquentiellement sur un clone dédié. Aucun code ni artefact historique de T28 n’a été rouvert. Un défaut reproductible de sécurité des répertoires runtime a été découvert pendant l’installation propre : `profiles`, `data`, `logs` et `browsers` étaient créés en `0755`. Le correctif minimal impose et répare `0700` dans les chemins `init`, `serve` et `mcp-stdio`, avec le test `TestEnsureRuntimeDirsOwnerOnly`. Ce correctif est publié dans le commit `b1559ca53852c493ba15e4a06ad89b0c171c7938`.

Le verdict autorisé n’est pas une préparation production. Il est :

> **`FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE`**

Ce verdict tient compte de deux échecs non critiques et de plusieurs environnements indisponibles. Aucun `FAIL_CRITICAL` n’a été observé : aucune fuite de secret, écriture partielle SQLite, corruption récupérée comme valide, isolation critique rompue ou exécution non autorisée n’a été démontrée.

## Matrice des résultats

| Phase | Statut | Résultat vérifié |
|---|---|---|
| `V-CORE` | **FAIL** | 23 assertions PASS ; CRUD, guards, erreurs, idempotence, redaction, arrêt brutal et reprise PASS. La capacité d’expiration/révocation du token admin est absente et reste FAIL non critique. |
| `V-SQLITE-CRASH-RECOVERY` | **PASS** | WAL, concurrence, verrou, transaction interrompue, reprise, DB absente, copie corrompue, permissions et écriture sous disque presque plein PASS ; aucun secret en clair dans les preuves. |
| `V-T28-RUNTIME-CONTRACT` | **PASS** | Tests T28 ciblés sous race, tests API ciblés, vet, build et diff check exit 0 sur `b1559ca`. Le code métier T28 reste inchangé. |
| `V-RUNTIME-SYNTHETIQUE` | **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` pour la qualification Camoufox/extension complète | Chromium système PASS pour page locale, isolation/persistance contrôlée, cookie synthétique, arrêt brutal et absence de processus résiduel. Camoufox absent ; la sémantique Firefox de l’extension ne doit pas être remplacée silencieusement par Chromium. |
| `V-PROXY-COOKIES-SYNTHETIQUES` | **PASS** | Proxy local authentifié `407/200`, échec fermé si proxy indisponible, registry proxy, affectation profil, import/export de fixtures redacted et refus des fixtures invalides PASS. |
| `V-SYSTEMVAULT` | **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` pour le natif | MemoryVault local : récupération, erreur, suppression, rotation, secret synthétique et concurrence PASS. Secret Service/SystemVault natif absent et non écrit. |
| `V-DOCKER` | **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE** | Docker Engine et Buildx absents ; aucune image, couche ou service n’a été lancé. |
| `V-DASHBOARD-API` | **FAIL** | Core/Dashboard, build/check, 5 specs Playwright T06/T10 sous Chromium système et redaction PASS. Le diagnostic Axe reproduit 1 violation `color-contrast` sérieuse sur 11 nœuds et la console signale des 404/500 d’assets/analytics non disponibles en environnement local. Défaut non critique, non masqué. |
| `V-INSTALLATION-PROPRE` | **PASS avec réserve explicite** | Installation vierge, `init`, `init --force`, permissions top-level 0700 après correctif, token 0600 après startup, health, arrêt, nettoyage et redaction PASS. La réinitialisation d’une configuration existante exige l’option explicite `--force`. |
| `V-SECURITY` | **FAIL / réserves** | Gitleaks : 0 finding, exit 0. Gosec : exit 0 avec 128 findings historiques sur `internal`, non masqués. OSV : 24 packages scannés, exit 1 ; analyse d’appel indisponible dans le contexte du scanner. pnpm audit production : 0 vulnérabilité, exit 0. Govulncheck, Semgrep, Trivy, Syft et Grype sont indisponibles selon la baseline. |
| `V-EVIDENCE` | **PASS après publication du package** | Logs UTC/CWD/exit codes, rapports scanners, manifeste, ZIP, bundle, sidecars, extraction et clone seedé publiés séparément. |

## Correction publiée

La correction opérationnelle est strictement limitée aux permissions des répertoires runtime. `ensureRuntimeDirs` utilise `MkdirAll(..., 0700)` puis `Chmod(..., 0700)` afin de réparer également les répertoires préexistants. `serve` et `mcp-stdio` réutilisent ce helper. Le test `TestEnsureRuntimeDirsOwnerOnly` prouve la réparation de chemins préexistants en `0755`.

Les requalifications post-correctif ont retourné exit 0 pour le test ciblé, la suite CLI, le build Core, le contrat T28 et les specs Dashboard/T10-T06. Aucun ZIP ou bundle T28 historique n’a été recréé.

## Limites d’environnement

Chromium `151.0.7922.71` était disponible et a été exécuté uniquement avec page HTTP locale, profils jetables et données synthétiques. Camoufox n’était pas disponible ; aucune substitution silencieuse n’est déclarée. Docker/Buildx et SystemVault natif n’étaient pas disponibles. Aucun compte réel, cookie réel, proxy commercial, SystemVault natif, release ou service de production n’a été utilisé.

Le Dashboard a été testé avec Chromium système via une configuration Playwright temporaire non publiée dans l’application. Les erreurs console `404` d’analytics placeholder et `500` d’assets `manus-storage` sont conservées dans `V-DASHBOARD-API_A11Y_CONSOLE_RAW.log`. La violation Axe de contraste est également conservée ; elle n’est ni ignorée ni transformée en PASS.

## Gates et lots suivants

Les gates historiques restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`.

T29, T39, T40, T41 et T42 ne sont pas démarrés dans cette campagne. Aucun runtime navigateur d’extension réel, proxy/cookies réels, migration, SystemVault natif ou release n’a été engagé.

## Preuves

Les journaux et rapports sont dans `evidence/OPERATIONAL_VALIDATION_V1/`. Le package ZIP et le bundle delta sont accompagnés de sidecars portables et de la vérification fraîche publiée séparément. Le package source n’est pas présenté comme contenant les journaux produits après sa génération.
