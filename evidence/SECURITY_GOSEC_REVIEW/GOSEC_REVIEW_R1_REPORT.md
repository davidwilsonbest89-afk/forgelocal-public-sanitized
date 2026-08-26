# GOSEC-REVIEW-R1 — rapport final

## Verdict

```text
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
FORGELOCAL_PRODUCTION_READY=false
```

La revue R1 est une revue automatisée/agent du code, des tests et des chemins d’exécution. Elle ne constitue pas une revue humaine indépendante et ne clôt pas la sécurité.

## Résumé exécutif

Le gap G112 du serveur HTTP Core a été corrigé avec `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` et `IdleTimeout`. Les régressions pour en-têtes lents, client abandonné et arrêt propre passent. Le scan Gosec source-only post-G112 contient **176 findings** contre 177 en baseline; la matrice individuelle conserve les 177 lignes de référence et classe G112 comme `CORRECTED_AND_VERIFIED`.

Les 176 findings encore signalés par Gosec restent visibles. Un finding G704 du pont WebSocket est classé `MITIGATED_CONTROL_SCANNER_OPEN` : le contrôle loopback/URL/port/chemin est testé, mais le scanner continue de signaler le `net.Dial`. Les 175 autres lignes restent `NEEDS_MANUAL_REVIEW`. Aucun finding n’a été masqué par `nolint`, skip, allowlist ou suppression globale.

## Matrice individuelle

La matrice complète est dans `GOSEC_R1_INDIVIDUAL_REVIEW.md` et `GOSEC_R1_INDIVIDUAL_FINDINGS.tsv`.

| Disposition | Nombre | Signification |
|---|---:|---|
| `CORRECTED_AND_VERIFIED` | 1 | G112 absent du scan post-correctif et couvert par les trois régressions dédiées. |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 1 | G704 durci et testé, mais toujours signalé statiquement. |
| `NEEDS_MANUAL_REVIEW` | 175 | Finding ouvert; précondition, chemin et remédiation doivent encore être examinés avant clôture. |

Les priorités restent les archives/filesystem (`G703`, `G304`, `G305`, `G122`, `G110`), subprocess (`G204`), permissions (`G301`, `G302`, `G306`), erreurs I/O (`G104`), puis `G404`, `G115` et `G101`.

## Scans R1

| Scan | Périmètre | Exit code | Résultat |
|---|---|---:|---|
| Gosec | `./cmd/... ./internal/...` | 1 | 176 findings, gate ouverte; aucun skip/allowlist. |
| Govulncheck | `./cmd/... ./internal/...` | 0 | aucune vulnérabilité trouvée. |
| Gitleaks | dépôt | 0 | aucune fuite détectée. |
| Trivy | filesystem, vuln/secret/misconfig | 0 | 0 vulnérabilité, 0 misconfiguration, 0 secret. |
| OSV | `go.mod` actuel | 1 | 46 vulnérabilités de dépendances; remédiation séparée requise. |
| OSV | `forge-dashboard/pnpm-lock.yaml` actuel | 0 | 0 vulnérabilité. |
| OSV | scan récursif dépôt | 1 | 98 correspondances, dont des SBOM historiques; ce nombre n’est pas utilisé comme décompte actuel unique. |
| Semgrep, Grype, Shellcheck, Yamllint | environnement | non exécuté | outils absents de l’environnement. |

Les versions et sorties brutes sont conservées dans `GOSEC_REVIEW_R1_SCANS_RAW.log`, `OSV_GO_MOD_R1_RAW.log`, `OSV_PNPM_LOCK_R1_RAW.log`, `OSV_R1_RAW.log` et les JSON associés. La présence de 46 vulnérabilités OSV dans `go.mod` maintient un gate dépendances ouvert, même si Govulncheck n’a trouvé aucune vulnérabilité atteignable selon son analyse.

## Tests et non-régressions

| Test | Résultat |
|---|---|
| `go test -race ./cmd/... ./internal/...` | PASS |
| `go vet ./cmd/... ./internal/...` | PASS |
| `go build ./cmd/... ./internal/...` | PASS |
| tests G112 timeout/client abandonné/shutdown | PASS |
| tests contrat proxy et WebSocket | PASS |
| `pnpm run check` Dashboard | PASS |
| test statique runner | PASS |
| syntaxe runner | PASS |
| preflight runner sans variables sensibles | PASS : échec contrôlé avant démarrage lorsque l’environnement obligatoire manque |

Les parcours fonctionnels déjà validés restent inchangés : `SMOKE_DASHBOARD_PROFILE_CREATE_PASS`, `SMOKE_DASHBOARD_PROFILE_RESTART_PASS` et `SMOKE_PROXY_LOCAL_PASS`. Ils n’ont pas été recommencés comme campagne complète; seules les régressions ciblées ont été rejouées.

## Runner Playwright

Le runner est versionné dans `tools/validation/integrated_smoke_runner.mjs`, avec `README.md`, configuration `.env.example` redacted et `test_runner_contract.mjs`. Il n’embarque aucune credential de fixture, n’utilise plus `waitForTimeout`, attend les sessions par `profile_id`, journalise des diagnostics minimaux et force le verdict négatif si `external_forward_observed` devient vrai.

Le runner reste limité au Chromium système et au loopback synthétique. Camoufox, SystemVault natif, Docker/Buildx, proxy commercial, cookies réels et sites externes ne sont pas validés.

## Provenance Git

| Élément | Commit |
|---|---|
| Correctif WebSocket ciblé | `ca00a0ffaf70c667e0583faafc3e477885a4c67f` |
| G112 + matrice R1 | `81958ea510c8ff07e7d4d5f1d918b26e2714ecad` |
| Runner publié | `6aaa1aa2a142a45724f98537fb48c629503d1ea1` |
| HEAD distant au début de la finalisation | `6aaa1aa2a142a45724f98537fb48c629503d1ea1` |

## Limites et décision

Aucun risque critique Gosec n’a été confirmé comme exploitable par cette revue automatisée, mais cette absence de confirmation ne vaut pas approbation de sécurité. Les findings Gosec restants et les 46 vulnérabilités OSV Go demeurent ouverts. Les outils absents ne sont pas traités comme PASS.

La décision R1 est donc **classifiée avec findings ouverts**, et non une clôture :

```text
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
FORGELOCAL_PRODUCTION_READY=false
```

T28 n’est pas redémarré, T29 n’est pas commencé et T31–T38 ne sont pas touchés. Les prochaines actions doivent traiter les findings restants par petits lots avec preuves de frontière, sans déclaration de production readiness.
