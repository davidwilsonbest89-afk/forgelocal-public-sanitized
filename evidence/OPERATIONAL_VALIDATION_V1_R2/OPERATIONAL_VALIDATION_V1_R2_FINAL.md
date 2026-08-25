# ForgeLocal — rapport final de validation opérationnelle V1 R2

**Directive appliquée :** `FORGELOCAL-OPERATIONAL-V1-R2`
**Branche :** `validation/operational-v1`
**HEAD source R2 :** `1e128bd3b2f1cb9b668afff25c7c155316fd0267`
**HEAD baseline R2 :** `84f7bd2bd7c956a4766d1bd1ff2808c59f9c6eed`
**Date des derniers contrôles :** 2026-08-25 UTC
**Périmètre :** uniquement le correctif R2 du cycle de vie du token admin et les corrections Dashboard React ; aucun lot T29 ni T31–T38 n’a été démarré ou modifié.

## 1. Conclusion exécutive

La campagne R2 a corrigé les deux défauts explicitement demandés. Côté Core, le token admin possède désormais des métadonnées versionnées, une expiration à TTL fixe de 15 minutes, une révocation persistante, une validation centralisée et des raisons d’erreur distinctes (`missing`, `malformed`, `invalid`, `expired`, `revoked`). La révocation est exposée par `POST /api/auth/revoke` sur la boucle locale, sans exposer la valeur du token dans les métadonnées, les journaux ou les réponses de validation.

Côté Dashboard, la sonde utilise une route réellement protégée, les raisons d’expiration/révocation sont propagées jusqu’à l’interface, les nœuds Axe précédemment signalés ont été corrigés, le focus clavier est explicite, les assets externes non disponibles ont été remplacés par des fallbacks locaux et le chargement analytique placeholder est conditionnel. La requalification finale a obtenu **0 violation Axe, 0 violation serious/critical, 0 erreur console, 0 requête échouée et 0 réponse HTTP mauvaise** sur le périmètre exécuté.

> **Verdict R2 exact : `FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE`**

Ce verdict signifie que les validations synthétiques autorisées et les corrections R2 sont probantes, mais que l’environnement ne permet pas de qualifier Camoufox, le SystemVault natif ni Docker/Buildx. Le dépôt n’est donc **pas prêt pour une release ou la production**. Les réserves sécurité restent publiées, notamment les findings Gosec historiques et les advisories OSV signalés par la version du scanner.

## 2. Matrice des validations

| Validation | Résultat R2 | Preuve principale | Portée et réserve |
|---|---|---|---|
| V-CORE admin token | **PASS** | `V-CORE-R2-AUTH_RAW.log` | Validité, missing/malformed/invalid, alias `/api/v1`, révocation immédiate et après restart, expiration après restart, redaction, cleanup. |
| V-DASHBOARD-API | **PASS — portée couverte uniquement** | `V-DASHBOARD-API_RAW.log` | 6 specs Playwright passées : T10 x2, R2 expired/revoked UI, T06 x3 ; typecheck/build pass ; redaction et cleanup pass. |
| V-DASHBOARD Axe/console/réseau | **PASS** | `V-DASHBOARD-API_A11Y_CONSOLE_RAW.log` | Axe 0, serious/critical 0, console errors 0, failed requests 0, bad responses 0. |
| V-SQLite crash recovery | **PASS** | `V-SQLITE-CRASH-RECOVERY_RAW.log` | WAL, concurrence, verrou, transaction SIGKILL, corruption, permissions, disque plein et cleanup synthétiques. |
| V-T28 runtime contract | **PASS** | `V-T28-RUNTIME-CONTRACT_RAW.log` | Contrat API et régressions Go ; ceci ne constitue pas une validation de l’UI Dashboard T28. T28 n’a pas été rouvert. |
| V-proxy/cookies synthétiques | **PASS** | `V-PROXY-COOKIES-SYNTHETIQUES_RAW.log` | Proxy local 407/200, indisponibilité fermée, registre/assignment, fixtures cookies import/export, digest, redaction et cleanup. |
| V-install fresh | **PASS** | `V-INSTALL-FRESH_RAW.log` | `init`, `init --force`, permissions 0700/0600, health, cleanup et `unexpected_sensitive_payload_files=0`. |
| Runtime Chromium synthétique | **PASS` pour Chromium synthétique** | `V-RUNTIME-SYNTHETIQUE_RAW.log` | Isolation, persistance et crash cleanup passés. Cette ligne ne qualifie pas Camoufox. |
| Runtime Camoufox | **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE** | `V-RUNTIME-SYNTHETIQUE_RAW.log` | Camoufox absent ; aucun téléchargement, substitution silencieuse ou exécution non autorisée. |
| SystemVault MemoryVault | **PASS comme double de test** | `V-SYSTEMVAULT-R2_RAW.log` | Double mémoire uniquement ; aucune prétention de qualification native. |
| SystemVault natif | **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE** | `V-SYSTEMVAULT-R2_RAW.log` | Secret Service/`secret-tool` absent ; aucune écriture réelle. |
| Docker/Buildx | **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE** | `V-DOCKER-R2_RAW.log` | Client/daemon Docker et Buildx absents ; aucun daemon privilégié, build ou conteneur lancé. |

## 3. Correctif Core et propriétés de sécurité

Le commit `1e128bd…` ajoute `internal/api/admin_token.go` et intègre son état au routeur. Les métadonnées `.api-token.meta` contiennent la version, le hash SHA-256 du token, `issued_at`, `expires_at` et `revoked_at`; le token en clair reste dans `.api-token` avec permission 0600, tandis que le répertoire est créé en 0700 et les écritures sont atomiques.

La validation est centralisée par `handler.validateAdminAuthorization` et appliquée aux routes REST d’administration, à l’alias `/api/v1/*` et au middleware readonly. En production, `NewRouter` instancie l’état `adminToken`; le fallback de fixture `validBearerToken` reste limité aux handlers construits directement par les tests et ne constitue pas un chemin de contournement dans le routeur réel.

Les tests ciblés couvrent la validité, l’absence, le format malformé, la valeur invalide, la frontière d’expiration, la révocation, le redémarrage, l’absence de valeur en clair dans les métadonnées, l’endpoint de révocation et la concurrence/race. La requalification opérationnelle finale a confirmé : `valid=200`, `missing=401 reason=missing`, `malformed=401 reason=malformed`, `invalid=401 reason=invalid`, alias invalide refusé, révocation HTTP `204`, puis `reason=revoked` immédiatement et après redémarrage ; une métadonnée expirée produit `reason=expired` après redémarrage.

La correction **ne fournit pas de rotation automatique ni de nouvelle émission via l’endpoint de révocation**. Après expiration ou révocation, une émission/rotation opérateur séparée reste nécessaire et n’est pas revendiquée comme implémentée dans R2.

## 4. Dashboard React : preuve et limites de couverture

Le parcours R2 réellement exécuté navigue dans le Dashboard avec Chromium système local, route des réponses synthétiques, clique les contrôles de connexion et vérifie les messages d’état `expired` et `revoked`. La valeur du token n’apparaît pas dans le DOM. Le runner complet a produit 6 specs passées, puis le contrôle indépendant Axe/console/réseau a confirmé l’absence des défauts ciblés.

La couverture Dashboard ne doit pas être étendue abusivement à des fonctions non implémentées. Les éléments suivants restent des limites explicites : les entrées « Espaces de travail », « Journal d’audit », « Réglages », « Aide », « Notifications » et « Filtres avancés » appellent un feedback `unavailable`; certaines actions de ligne restent indisponibles hors snapshot Core ; plusieurs panneaux exigent une connexion Core ou une sélection de profil ; les workflows Dashboard T28 d’installation, provenance, signature, allowlist, rollback et lifecycle d’extension ne sont pas implémentés ni testés comme UI. Ils sont donc classés **`NOT_IMPLEMENTED_PLACEHOLDER`**, et non PASS.

Les workflows API T28 couverts par le contrat restent une preuve séparée et ne ferment pas ces limites UI. Les fonctions déjà présentes et exercées sur le périmètre local, comme le registre synthétique, les lectures T06/T10 et le feedback d’authentification R2, ne valent pas qualification d’un produit Dashboard final.

## 5. Sécurité et qualité

| Contrôle | Résultat | Interprétation R2 |
|---|---:|---|
| `go test -count=1 -race ./internal/api ./cmd/server` | 0 | PASS. |
| `go vet ./internal/api ./cmd/server` | 0 | PASS. |
| `go build ./...` | 0 | PASS. |
| `git diff --check` | 0 | PASS. |
| Gitleaks | 0 finding | Aucun secret détecté dans le périmètre scanné ; les sorties sont redacted. |
| Gosec | exit 0, 128 findings historiques | Réserve non bloquante nouvelle non démontrée ; findings conservés, sans allowlist/nolint globale. Voir `GOSEC_OPERATIONAL.json` et `gosec.out`. |
| OSV | exit 1, 24 packages et advisories Go stdlib listés | Réserve maintenue ; voir `osv.go.out`. |
| govulncheck | exit 0, “No vulnerabilities found.” | PASS sur l’analyse atteignable avec Go 1.25.13 ; ne remplace pas OSV/Gosec. |
| pnpm audit production | exit 0 | PASS dans le runner sécurité. |
| Trivy filesystem | exit 0, 0 vulnérabilité dans la synthèse pnpm/Go affichée | Contrôle filesystem conservé dans `trivy_fs.txt`; l’exécution inclut vuln/misconfig/secret hors `.git`, `node_modules`, `dist` et `data`. |
| Syft | exit 0, 747 artefacts SBOM | Inventaire dans `syft.json`; ce n’est pas un avis de licence finale. |
| Semgrep, Grype, Shellcheck, Yamllint | indisponibles | Non exécutés ; non masqués. |

Le gate `SCAN_BLOCKED_UNKNOWN` reste inchangé. Les résultats Gosec/OSV sont publiés comme réserves et non convertis artificiellement en succès global.

## 6. Garde-fous et lots hors périmètre

| Gate ou lot | État R2 |
|---|---|
| `PUBLIC_RELEASE_BLOCKED` | Inchangé, actif. |
| `SCAN_BLOCKED_UNKNOWN` | Inchangé, actif. |
| `NATIVE_SYSTEMVAULT_NOT_TESTED` | Inchangé, actif. |
| `camoflox_execution_authorized=false` | Inchangé. |
| `t08_authorized=false` | Inchangé. |
| `release_authorized=false` | Inchangé. |
| T28 | Contrat rejoué sans rouvrir ni modifier le code métier T28. |
| T29 | Non démarré. |
| T31–T38 | Non modifiés et non rejoués comme lots produits. |

Aucune donnée utilisateur, aucun compte réel, aucun secret réel, aucun cookie réel, aucun proxy réel, aucun site externe, aucune release ou production n’a été utilisé. Les bases, profils, cookies, packages et serveurs HTTP sont synthétiques et locaux.

## 7. Sources d’évidence et références

Les journaux bruts R2, les sorties de scanners, le manifeste et les sidecars sont distribués dans ce répertoire. Les artefacts compressés sont générés après commit de ces documents et exclus de leur propre manifeste afin d’éviter une circularité de hash. Le rapport final sera vérifié dans une extraction fraîche et un clone public sparse avant publication du verdict.

[1]: https://github.com/ `Dépôt GitHub de référence — branche validation/operational-v1`
[2]: https://github.com/forgelocal `Organisation ou dépôt public ForgeLocal, si accessible`
