# ForgeLocal — Clôture d’environnement V2

**Auteur :** Manus AI  
**Branche :** `validation/final-environment-qualification`  
**Date de clôture :** 2026-08-26  
**Verdict :** qualification partielle, production non prête

> `FORGELOCAL_PRODUCTION_READY=false`

Ce rapport conserve les contrôles nouvellement exécutés pendant cette clôture, les corrections limitées effectivement appliquées, les éléments historiques R7 relus sans les présenter comme de nouvelles exécutions, ainsi que les limitations propres à l’environnement de qualification. Aucune release publique n’a été préparée et aucune gate de production n’a été levée.

## 1. Périmètre et intégrité de la branche

La qualification a été reprise depuis un clone neuf du dépôt `davidwilsonbest89-afk/forgelocal-public-sanitized`. Le checkout a été réalisé explicitement sur `refs/remotes/origin/validation/final-environment-qualification`. Le commit de preuves avant package est `da813e5adabe983df0ff499820417a2c2fb24b9b`, et le commit package est `f53849fc1ca6f03b3300607893903b3ead0ee24a`. Le HEAD final de publication est celui consigné dans le journal d’audit du clone neuf final.

Le préflight est conservé dans [`FINAL_ENVIRONMENT_BASELINE_DISCOVERY_RAW.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/FINAL_ENVIRONMENT_BASELINE_DISCOVERY_RAW.log). Il contient l’UTC, le répertoire de travail, l’architecture, l’espace disque, la mémoire, les processus actifs et les versions disponibles avant installation. Les outils supplémentaires ont ensuite été installés uniquement dans l’espace de qualification ou depuis les dépôts officiels Ubuntu et les releases officielles des projets.

## 2. Corrections effectivement appliquées

Les modifications de code sont limitées à quatre corrections de robustesse et de politique d’installation. La configuration workspace pnpm contient désormais `blockExoticSubdeps: true`, `minimumReleaseAge: 10080` et `trustPolicy: no-downgrade`. Les messages shell contenant une apostrophe sont passés en chaînes double-quotées afin de préserver la syntaxe. Enfin, la génération du token de `start-core-e2e.sh` est séparée de son export pour éviter de masquer un code de retour.

Aucun finding de sécurité conservé comme potentiellement réel n’a été neutralisé par `nosec`, allowlist, exclusion de scanner ou modification de règle. Les warnings Semgrep sur les transports non chiffrés, l’écriture HTTP, l’aléa non cryptographique et la commande runtime restent ouverts et doivent être traités par une revue produit dédiée.

## 3. Références R7 historiques — non nouvelles

Les artefacts suivants sont **hérités de R7**. Ils ont été vérifiés depuis le clone neuf, mais leur origine et leur exécution publique restent historiques.

| Élément | Référence ou hash | Classification |
|---|---|---|
| Commit source R7 | `3656dbad4bfef0381e1f9d837271d293ecffe292` | `INHERITED_FROM_R7` |
| Commit d’évidence R7 | `1dca108197f90e379184474ad37bbb9f386fe309` | `INHERITED_FROM_R7` |
| Commit package R7 | `5cd16b11f62b8c16582973001ce81bff9ef03dcf` | `INHERITED_FROM_R7` |
| Branche R7 distante | `b907dfcd68c290144e2b922e352d5a937e9b3259` | `INHERITED_FROM_R7` |
| `forgelocal-gosec-r7-final-v4.zip` | `80b0546eca714826023bfc4cc3e33381b1b9d3dfe52900771dd00a2cd2ba5ed8` | `PASS`, `INHERITED_FROM_R7` |
| `forgelocal-gosec-r7-final-v4.tar.gz` | `faf43da4e69e20aa4dc59863173b0f64b8a411f102d7dcd4c46b22fc8089fa7e` | `PASS`, `INHERITED_FROM_R7` |
| `forgelocal-gosec-r7-delta-3656dba-1dca108.bundle` | `6f732c627ee58529898753c09f292c5c0e79b9ee4a32cf6bd58b755f7ae4edb0` | `PASS`, `INHERITED_FROM_R7` |
| Manifeste R7 v4 | `evidence/GOSEC_R7/R7_FINAL_MANIFEST_V4.txt` | `INHERITED_FROM_R7` |
| Log public R7 v4 | `evidence/GOSEC_R7/R7_PUBLIC_VERIFICATION_V4_RAW.log` | `INHERITED_FROM_R7` |

Le replay neuf R7 a validé les trois sidecars après normalisation temporaire de leurs chemins absolus historiques, `unzip -t`, lecture TAR, extraction ZIP/TAR, `git bundle verify` et `git fsck --full`. Le détail est dans [`R7_PACKAGE_VERIFICATION_RAW.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/R7_PACKAGE_VERIFICATION_RAW.log).

## 4. Analyse statique nouvelle

Semgrep a analysé 649 fichiers source-only avec 597 règles. Le résultat final est de **15 findings ouverts** et **4 erreurs de parsing de règles**. Les trois findings de politique pnpm présents avant correction ont disparu après l’ajout des paramètres de sécurité ; ils sont classés `PASS_REMEDIATED`, et non comme findings supprimés artificiellement.

| Finding Semgrep ouvert | Localisation | Règle | Décision |
|---|---|---|---|
| Insecure WebSocket | `API.md:444`, `API.zh-TW.md:468`, `README.md:363`, `README.zh-TW.md:362` | `detect-insecure-websocket` | `FAIL_OPEN_REVIEW` — documentation ou exemples à confirmer |
| Requêtes HTTP non chiffrées | `extension/sidebar/app.js:17,70` | `react-insecure-request` | `FAIL_OPEN_REVIEW` |
| Websocket HTTP interne | `internal/api/sessions.go:432,441` | `detect-insecure-websocket` | `FAIL_OPEN_REVIEW` — transport loopback à contextualiser |
| Ressource externe sans SRI | `forge-dashboard/client/index.html:8` | `missing-integrity` | `FAIL_OPEN_REVIEW` |
| Écriture directe `http.ResponseWriter` | `internal/api/dashboard.go:14` | `no-direct-write-to-responsewriter` | `FAIL_OPEN_REVIEW` |
| Aléa `math/rand` | `internal/fingerprint/pool.go:7` | `math-random-used` | `FAIL_OPEN_REVIEW` — usage non cryptographique à confirmer |
| Aléa `math/rand/v2` | `internal/humanize/keyboard.go:5` | `math-random-used` | `FAIL_OPEN_REVIEW` — usage non cryptographique à confirmer |
| Aléa `math/rand/v2` | `internal/humanize/math.go:5` | `math-random-used` | `FAIL_OPEN_REVIEW` |
| Aléa `math/rand/v2` | `internal/humanize/mouse.go:4` | `math-random-used` | `FAIL_OPEN_REVIEW` |
| Commande runtime dynamique | `internal/runtime/qualification.go:85` | `dangerous-exec-command` | `FAIL_OPEN_REVIEW` — chemin contrôlé mais à borner explicitement |

Les quatre erreurs restantes sont des erreurs de parsing auto-config sur `validation_systemvault_native.json`, `scripts/check-browseforge-chromium-assets.sh`, `scripts/collect-bootstrap-ro-evidence.sh` et `scripts/collect-t06-evidence.sh`. Elles ne sont pas converties en findings de sécurité et ne sont pas masquées par exclusion.

ShellCheck a produit **26 diagnostics** après correction. Le seul niveau `error` restant est `SC2066` sur `scripts/collect-t06-evidence.sh:47`, où la boucle contient volontairement un seul port déjà quoté ; l’expansion non quotée serait moins sûre. Les dix diagnostics `SC2094` concernent la génération de `SHA256SUMS` avec exclusion explicite du fichier de sortie. Les diagnostics `SC2129`, `SC2015`, `SC2034`, `SC2086` et `SC1091` restent documentés comme style, prudence ou environnement.

Yamllint reste en échec avec **895 diagnostics** : 889 sur le lockfile pnpm, 3 sur `docker/docker-compose.yml`, 1 sur `V6_SEMGREP_RULES.yml`, 1 sur `examples/multi-login.yaml` et 1 sur le workspace pnpm. Les détails bruts sont dans [`yamllint.txt`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/scanners/yamllint.txt). Le lockfile généré n’a pas été reformatté automatiquement, afin de ne pas modifier une sortie de gestionnaire de paquets sans décision dédiée.

## 5. Contrôles frontend et JavaScript

Le contrôle TypeScript `pnpm --dir forge-dashboard run check` est **PASS** avec exit code `0`. Le build Vite/esbuild est **PASS** avec exit code `0`. `check-doc-language` est **PASS**. `test-t07-r-receipt` est **PASS**. `check:component-rights` est **FAIL** car le registre signale des empreintes inattendues pour `go.mod` et `go.sum`. `test-t07-provenance` est **FAIL** car le registre conserve `integration_state=t08-concurrency-in-progress` alors que le test exige `provenance-qualification-blocked`. Ces incohérences de gouvernance n’ont pas été réécrites pour faire passer les tests.

## 6. Régression Go et environnement de dépendances

Les tentatives nouvelles de `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...` et `go build ./cmd/... ./internal/...` ont été lancées avec Go 1.25.13 et une toolchain C vérifiée. Les téléchargements de modules ont rencontré `unexpected EOF` via `proxy.golang.org`; la tentative alternative `GOPROXY=direct` a expiré après 300 secondes. Le replay offline borné a donc produit exit code `1` pour les trois commandes. Ces résultats sont classés `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` pour la validation complète de la toolchain, et non comme preuves de succès.

Gosec a été installé depuis sa release officielle. Le premier appel avec des chemins absolus n’a chargé aucun package ; le second appel correct sur `./...` a été lancé mais a expiré après 300 secondes sans produire une analyse complète exploitable. Le finding historique Gosec R7 ne doit donc pas être compté comme nouvelle exécution. Il reste `INHERITED_FROM_R7`.

Gitleaks source-only et extraction sans Git sont **PASS**, exit code `0`, sans finding. OSV sur le périmètre Go/pnpm et sur le dashboard est **PASS**, exit code `0`. Trivy filesystem est **PASS** au niveau de la commande, avec 6 résultats de scanner conservés dans son JSON. Syft JSON et CycloneDX sont **PASS**, avec 725 artefacts inventoriés. Grype n’a pas terminé dans le délai de 300 secondes, donc il est `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`.

## 7. Matrice proxy locale

Le banc synthétique loopback a validé les scénarios suivants : trafic valide vers une cible locale, séparation des profils A/B, proxy arrêté, port invalide, hôte externe refusé par politique, token expiré, token révoqué et timeout. Tous les scénarios sont **PASS** et aucun processus résiduel n’a été observé. Cette preuve est une simulation contrôlée de contrat ; elle ne vaut pas validation d’un fournisseur proxy externe.

Le détail JSON est conservé dans [`proxy-matrix.json`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/environment/proxy-matrix.json). Les retours importants sont `200` pour le trafic valide, `403` pour l’hôte externe refusé, `407` pour les tokens expiré/révoqué et `28` pour le timeout curl.

## 8. Disponibilité plateforme

| Contrôle | Classification |
|---|---|
| Chromium headless, profil A | `PASS`, exit `0` |
| Chromium headless, profil B | `PASS`, exit `0` |
| Nettoyage des profils/processus | `PASS` |
| Docker/Buildx | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Firefox standard | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Camoufox | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| SystemVault natif | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Windows/macOS natifs | `BLOCKED_ENVIRONMENT_REQUIRED` |
| `xdg-open` | non exécuté ; aucune preuve visuelle réelle distincte du simple `DISPLAY` |

Le log brut est [`environment-checks-raw.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/environment/environment-checks-raw.log). Aucun test natif Windows/macOS, aucun secret réel et aucun contournement par simulation n’a été introduit.

## 9. Classification finale et conservation

| Classification | Contenu |
|---|---|
| `PASS` | clone neuf, checkout explicite, R7 package integrity, fsck, pnpm dashboard check/build, doc-language, T07 receipt, Gitleaks, OSV, Trivy command, Syft, Chromium A/B, proxy loopback synthétique |
| `FAIL` | 15 findings Semgrep ouverts, ShellCheck exit 123 avec SC2066, Yamllint exit 123, component-rights, T07 provenance, Go offline replay exit 1 |
| `BLOCKED_ENVIRONMENT_REQUIRED` | contrôles Windows/macOS natifs et leurs stores natifs |
| `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Go complet avec dépendances réseau, Govulncheck complet, Grype complet, Docker/Buildx, Firefox/Camoufox, SystemVault natif |
| `INHERITED_FROM_R7` | commits, manifeste, log public et artefacts ZIP/TAR/bundle R7 |

Les preuves nouvelles sont regroupées dans [`evidence/FINAL_ENVIRONMENT_CLOSURE_V2/`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/). Le package a été construit sans auto-référence depuis le commit de preuves `da813e5…`, puis ajouté par le commit `f53849f…`. Ses sidecars SHA-256 et son contenu ont été vérifiés depuis un clone neuf. La branche dédiée uniquement a été publiée ; aucune branche `main`, release ou production n’a été modifiée.

## Références de preuve

1. [`FINAL_ENVIRONMENT_BASELINE_DISCOVERY_RAW.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/FINAL_ENVIRONMENT_BASELINE_DISCOVERY_RAW.log)
2. [`R7_PACKAGE_VERIFICATION_RAW.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/R7_PACKAGE_VERIFICATION_RAW.log)
3. [`STATIC_ANALYSIS_SUMMARY_FINAL.tsv`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/STATIC_ANALYSIS_SUMMARY_FINAL.tsv)
4. [`FINAL_STATIC_SUMMARY.tsv`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/FINAL_STATIC_SUMMARY.tsv)
5. [`FRONTEND_CHECKS_FINAL_RAW.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/FRONTEND_CHECKS_FINAL_RAW.log)
6. [`FINAL_ENVIRONMENT_TESTS_RAW.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/FINAL_ENVIRONMENT_TESTS_RAW.log)
7. [`proxy-matrix.json`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/environment/proxy-matrix.json)
8. [`environment-checks-raw.log`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2/environment/environment-checks-raw.log)
9. [`R7_FINAL_MANIFEST_V4.txt`](evidence/GOSEC_R7/R7_FINAL_MANIFEST_V4.txt)
10. [`R7_PUBLIC_VERIFICATION_V4_RAW.log`](evidence/GOSEC_R7/R7_PUBLIC_VERIFICATION_V4_RAW.log)
11. [`FINAL_PUBLIC_VERIFICATION_V2_FINAL_RAW.log`](evidence/FINAL_PUBLIC_VERIFICATION_V2_FINAL_RAW.log)
12. [`FINAL_ENVIRONMENT_CLOSURE_V2_PACKAGE_MANIFEST.txt`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2_PACKAGE/FINAL_ENVIRONMENT_CLOSURE_V2_PACKAGE_MANIFEST.txt)
13. [`forgelocal-final-environment-closure-v2.zip`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2_PACKAGE/forgelocal-final-environment-closure-v2.zip) — SHA-256 `dfd008b4d7df4e73be0496a15124062e2152126238126b3135fbe9d62e3fd14c`
14. [`forgelocal-final-environment-closure-v2.tar.gz`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2_PACKAGE/forgelocal-final-environment-closure-v2.tar.gz) — SHA-256 `25530d8a07da93fee1b19d14a635f02e662257e76bdf70f7610631c0590b909e`
15. [`forgelocal-final-environment-closure-v2.bundle`](evidence/FINAL_ENVIRONMENT_CLOSURE_V2_PACKAGE/forgelocal-final-environment-closure-v2.bundle) — SHA-256 `0db4d4e9df7472003eeca6cc4df434790b00bdaa10728065fab1d32e6a4be556`
