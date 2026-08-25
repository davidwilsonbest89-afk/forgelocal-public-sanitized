# T00–T42 V5 — remédiation des findings réels et qualification V6

**Statut exact :** `T00_T42_V6_FINDINGS_REMEDIATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`

**Branche de travail :** `audit/t00-t42-v6-findings-remediation`  
**Origine immuable :** `audit/t00-t42-self-validation-synthetic-e2e`, HEAD V5 `b34fa5c02ff20144abfb5d240db1c67ad1f038f9`  
**Baseline :** tag `t00-t27-complete-20260820`, résolu à `72d54110c89583beacc556bb103f881b667d8137`  
**Nature :** livraison append-only ; aucune preuve V3, V4 ou V5 n’a été réécrite.

## 1. Décision et périmètre

La remédiation V6 a été menée depuis un clone neuf de la branche V4 demandée. La découverte brute a été produite avant toute modification sous `V6_BASELINE_DISCOVERY_RAW.log`. Les corrections ont été limitées aux deux findings GolangCI-Lint obligatoires, aux deux violations Axe et aux deux advisories `golang.org/x/mod`. Les 18 findings Semgrep ont été inspectés individuellement et se sont révélés être des appels à `crypto/rand`, non à `math/rand`. Les findings Gitleaks historiques ont été individualisés et redacted ; aucune clé, aucun token et aucun secret opérationnel n’a été exposé.

Cette qualification ne vaut pas release. Aucun runtime réel, Camoufox, proxy réel, cookie réel, donnée utilisateur, migration, SystemVault natif, opération de production ou gate n’a été exécuté ou levé.

## 2. Corrections livrées par commits séparés

| Sous-correction | Fichiers principaux | Preuve de test | Commit |
|---|---|---|---|
| GolangCI `ineffassign` | `cmd/server/cli_runtime.go` : déclaration écrasée remplacée par une déclaration nulle adaptée aux branches de configuration | `go test -count=1 ./cmd/server`, race ciblé, linter post-correction ; `TestCLIStatusAndMCPConfigJSON` couvre le chemin CLI | `d3bb894` |
| GolangCI `SA1019` | `internal/api/sessions.go` : `Page.WaitForSelector` remplacé par `page.Locator(...).WaitFor` avec `LocatorWaitForOptions` | `go test -count=1 ./internal/api` et race ciblé ; aucun finding ciblé résiduel dans le JSON linter | `d3bb894` |
| Axe contraste et viewport | `forge-dashboard/client/src/index.css` : couleurs relevées sur les quatre sélecteurs signalés ; `client/index.html` : suppression de `maximum-scale=1` | `pnpm install --frozen-lockfile`, `pnpm exec tsc --noEmit`, `pnpm run build`, E2E Playwright/Axe loopback | `e2fad93` |
| Axe reproductible | `@axe-core/playwright` `4.13.0` ajouté au manifeste/lockfile ; scan ajouté au scénario `bootstrap-ro.spec.ts`, blocage sur `serious`/`critical` | 1 test Playwright passé, JSON Axe après correction à 0 violation | `e2fad93` |
| Advisories Grype | `golang.org/x/mod` mis à jour de `v0.37.0` à `v0.40.0`, avec les dépendances nécessaires induites par son `go.mod` | `go mod tidy`, `go mod verify`, tests race complets, `go vet`, `go build`, Grype SBOM propres à 0 match | `fc08045` |

Aucun `nolint` global, aucune allowlist large et aucune suppression de preuve n’a été utilisée.

## 3. Triage Gitleaks

La preuve V5 conserve les 348 findings redacted sur 58 arbres et les six findings du checkout frais. `V6_GITLEAKS_TRIAGE.csv` contient une ligne individuelle pour chacune des 348 occurrences, avec chemin, commit, type de fichier, règle, ligne, statut, action et hash de contenu non secret. Six chemins uniques de provenance ont été comparés au contenu courant : les six occurrences sont la même empreinte publique de clé de signature PPA, et non une clé privée ou un secret d’accès. Leur statut est `FALSE_POSITIVE_CURRENT_EXACT_EXCEPTION` pour le HEAD V4 et `HISTORICAL_FALSE_POSITIVE_EXACT_EXCEPTION` pour les occurrences historiques.

La configuration V6 `.gitleaks.toml` limite l’exception à l’empreinte publique précise et aux six chemins de provenance exacts. Le scan du checkout avec cette configuration retourne code 0. Le scan historique des quatre commits V6 ajoutés après V5 a été effectué par `git archive` arbre par arbre, sans allowlist historique : 4 arbres, 4 résultats code 0. La plage obligatoire `--log-opts=BASE..HEAD` continue cependant d’annoncer `0 commits scanned`; ce résultat est conservé comme limite d’outil et n’est pas nommé PASS. La couverture historique opposable reste donc le scan explicite par arbres V5, complété par les quatre arbres V6.

## 4. Triage Semgrep

La règle locale `go-math-rand-read` produit toujours 18 résultats individuels. L’inspection de chaque import et de chaque usage est consignée dans `V6_SEMGREP_TRIAGE.md`. Les 18 appels résolvent vers `crypto/rand`; aucun n’utilise `math/rand` pour un secret, un code, une session ou un token. Les usages à impact élevé, tels que nonce AES-GCM, sel de coffre local, clé de backup et token API, sont donc cryptographiquement appropriés. Les usages d’ID ou d’artefact temporaire sont documentés comme faux positifs contextuels, sans remplacement artificiel par un générateur moins sûr.

Le rescan Semgrep après correction retourne exactement 18 résultats, tous de cette règle contextuelle ; aucune règle `eval` ou injection HTML ne matche.

## 5. Triage Grype et vulnérabilités

Les deux correspondances V5 concernaient `golang.org/x/mod v0.37.0`, transitive via `modernc.org/sqlite → modernc.org/libc`. Elles sont identifiées individuellement comme `GO-2026-6180 / CVE-2026-56864` et `GO-2026-6179 / CVE-2026-56865`. Les deux sont corrigées en `v0.40.0`, mise à jour compatible avec le module Go `1.25.0` et le toolchain `go1.25.13`. La chaîne `go mod why -m golang.org/x/mod` et le graphe `go mod graph | grep 'golang.org/x/mod'` sont conservés dans les preuves de triage.

Après correction, les SBOM propres CycloneDX et SPDX ont été régénérés avec Syft `1.51.0`, en excluant uniquement `./forge-dashboard/node_modules`, `./.git` et `./forge-dashboard/dist`. Grype `0.117.0` retourne **0 match** sur chacun de ces deux SBOM. Les JSON Grype bruts sur le répertoire non nettoyé sont également conservés : ils contiennent 46 correspondances provenant du binaire de build `@esbuild/linux-x64` présent dans `node_modules`, identifié par Grype comme stdlib Go `1.23.12`; ce résultat de scanner n’a pas été masqué et n’est pas utilisé comme SBOM de livraison du code/manifeste.

| Contrôle | Résultat V6 |
|---|---:|
| `go mod tidy` | 0 |
| `go mod verify` | 0 |
| `go test -count=1 -race ./...` | 0 |
| `go vet ./...` | 0 |
| `go build ./...` | 0 |
| Grype CycloneDX propre | 0 match |
| Grype SPDX propre | 0 match |
| `govulncheck -json ./...` | 0 |
| OSV Scanner v1.9.2 | 46 résultats stdlib associés à la directive `go 1.25.0`; revue de toolchain requise, non masqués |

OSV Scanner v1.9.2 interprète la version mineure/directive `1.25.0` du module, alors que l’exécution et les binaires de cette qualification utilisent Go `1.25.13`. Les 46 résultats stdlib sont conservés dans `V6_OSV_SCANNER_FINAL.json`; `govulncheck` avec le toolchain effectif retourne 0. Cette différence de modélisation reste une exception de revue indépendante, pas un PASS global.

Trivy `0.74.0` retourne zéro vulnérabilité et zéro secret sur les manifests Go/pnpm analysés, ainsi que six misconfigurations historiques de Docker réparties sur deux Dockerfiles : utilisateur root, absence de HEALTHCHECK et installation apt sans `--no-install-recommends`. Elles sont conservées dans le JSON Trivy et ne sont pas corrigées ici, car modifier l’image d’exécution hors du mandat de la qualification loopback risquerait de changer le runtime protégé sans test d’intégration autorisé.

## 6. Axe et E2E synthétique

Le JSON Axe V5 avant correction est conservé séparément et contenait `color-contrast` de sévérité `serious` et `meta-viewport` de sévérité `moderate`. Le scénario V6 utilise un BrowserContext Playwright explicite, exécute Axe après le bootstrap loopback et échoue sur toute violation `serious` ou `critical`. Le JSON après correction contient 40 règles passées, 2 inapplicables et **0 violation**.

Le scénario complet a été exécuté avec un Core compilé localement et lancé uniquement sur `127.0.0.1` avec `--no-runtime`, un token temporaire en fichier `0600`, des répertoires `mktemp`, une base SQLite temporaire et un Dashboard Vite loopback. Un test Playwright est passé en 10,2 minutes. Le rejet du rejeu, l’expiration, le 401 forcé, le stockage navigateur vide et les requêtes redacted sont passés. Le cleanup confirme : token supprimé, base temporaire supprimée, répertoire d’exécution supprimé, ports non écoutés, CSS restauré et aucun processus temporaire résiduel.

## 7. Requalification complète

| Contrôle | Code | Observation |
|---|---:|---|
| `git fsck --full` | 0 | Deux blobs dangling non référencés, aucun objet corrompu signalé |
| `git lfs fsck` | 1 | 14 objets historiques LFS indisponibles dans le dépôt distant ; le fetch ciblé a été tenté et documenté, sans pull global |
| `go test -shuffle=on -count=3 ./...` | 0 | Passé |
| `go test -shuffle=on -count=3 -race ./...` | 0 | Passé |
| `go vet ./...` | 0 | Passé |
| `go build ./...` | 0 | Passé |
| `staticcheck ./...` | 1 | 34 diagnostics historiques, aucun nouveau finding des corrections ciblées |
| GolangCI-Lint 2.13.1 | 1 | 89 findings historiques ; les deux findings ciblés V5 ne subsistent plus |
| Semgrep local | 0 outil / 18 résultats contextuels | Tous `crypto/rand`, matrice individuelle conservée |
| `osv-scanner` | 1 | 46 résultats stdlib liés à la limite de version directive, conservés |
| `trivy fs --scanners vuln,secret,misconfig .` | 0 | 0 vulnérabilité, 0 secret, 6 misconfigurations Docker historiques |
| SBOM CycloneDX/SPDX + Grype | 0 / 0 | 0 match sur les SBOM propres V6 |
| Playwright/Axe synthétique | 0 | 1 test passé, 0 violation Axe, cleanup PASS |

L’inventaire de licences Syft contient 744 composants, dont 741 sans déclaration exploitable dans les métadonnées détectées. Ces éléments restent `UNKNOWN` dans le CSV et nécessitent une revue humaine ; aucune licence n’a été inventée ou automatiquement blanchie.

## 8. Preuves et commits

Le commit initial `638a39f` contient la découverte brute, la matrice Gitleaks, les matrices Semgrep/Grype et la configuration Gitleaks exacte. Le commit `d3bb894` contient la correction Go et les tests/linter ciblés. Le commit `e2fad93` contient le correctif Dashboard, le test Axe, les manifests et les preuves E2E. Le commit `fc08045` contient la mise à jour x/mod et les tests de dépendance. Les preuves finales de requalification, les SBOM, les scans et le wrapper V6 sont ajoutés append-only après ces commits.

Les wrappers V3, V4 et V5 restent byte-identiques à leurs sources et sont inclus comme copies historiques dans le wrapper V6. Les sidecars, le manifeste non auto-référentiel, les checksums, le bundle delta, l’extraction fraîche, `unzip -t`, le rescan Gitleaks d’extraction et les hashes historiques ont été vérifiés dans `V6_WRAPPER_VERIFY.log` : chaque contrôle du wrapper retourne code 0. La valeur de hash du wrapper final est maintenue dans le manifeste racine et son sidecar, hors du contenu hashé du wrapper, afin d’éviter l’auto-référence.

## 9. Gates inchangées

Les gates restent strictement inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. T28, T29, T39, T40, T41 et T42 restent `BLOCKED`. T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`. Cette livraison ne constitue ni une release ni une autorisation d’exécution réelle.

## Références

[1]: https://pkg.go.dev/vuln/GO-2026-6180 "Go Vulnerability Database — GO-2026-6180 / CVE-2026-56864"
[2]: https://pkg.go.dev/vuln/GO-2026-6179 "Go Vulnerability Database — GO-2026-6179 / CVE-2026-56865"
[3]: https://github.com/dequelabs/axe-core-npm/blob/develop/packages/playwright/error-handling.md "axe-core Playwright error handling"
