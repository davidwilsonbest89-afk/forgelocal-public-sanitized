# GOSEC-R5 — rapport final de hardening et de conservation

## Verdict

La campagne GOSEC-R5 est **classée avec findings ouverts**. Les lots R5-A, R5-B et R5-C ont été exécutés sur la branche dédiée `validation/gosec-r5`. Le hardening source publié concerne le confinement de fichiers de configuration, de marqueurs runtime et du token bootstrap. Les findings Gosec restants ont été individualisés et conservés ouverts lorsque le contrôle était intentionnel, dépendant de la plateforme ou nécessitait une revue manuelle.

Les verdicts finaux sont :

```text
GOSEC_R5_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R5_PARTIAL_ENVIRONMENT_UNAVAILABLE
FORGELOCAL_PRODUCTION_READY=false
```

## Lignée GitHub et rôles des commits

| Référence | Rôle |
|---|---|
| `2d528eae1788c4e05bd51d1fa7235ecc3200c66c` | HEAD R4 découvert et vérifié au démarrage de R5 |
| `079c452444b6bc55d885afafdd845a72bffd7ab4` | commit de baseline R5 |
| `096e3f591f273801cf735669d7a135daf6a7f601` | baseline/matrice/threat model R5 publiés |
| `54ed3a4964806eeb4880c9ebb3949d410c335174` | commit source R5-A : hardening config, marqueur runtime et token bootstrap |
| `ad2dd61bce4acf5e735b08a9afdc45605eaf90ff` | preuves R5-A publiées |
| `9412b314a3cdaff51d59b562ff2f97504844d9dc` | preuves R5-B et correction de rattachement R5-A publiées |
| `c4b4c95151e92a6b17b08d304c7a4d178116796e` | preuves R5-C publiées |
| `ef73ba4bd684bbf9e1f03c8b9411c2d47a463c15` | dernier HEAD d’évidence avant ce rapport |

La branche `validation/operational-v1` reste inchangée à `80048180d4ec5241c08146ade698d53a3f29454d`. Aucun commit R5 n’y a été poussé. T28 n’a pas été rouvert, T29 n’a pas été démarré et T31–T38 n’ont pas été modifiés.

## Résultats Gosec

Le périmètre autoritatif est source-only : `gosec -fmt json ... ./cmd/... ./internal/...`. Les artefacts et snapshots historiques ne sont jamais inclus dans ce scan. La baseline comptait 63 findings; le scan final compte **59 findings**, sans nouveau finding par rapport au scan R5-C.

| Règle | Baseline R5 | Final R5 | Variation |
|---|---:|---:|---:|
| G101 | 1 | 1 | 0 |
| G104 | 0 | 0 | 0 |
| G107 | 0 | 0 | 0 |
| G115 | 3 | 3 | 0 |
| G204 | 5 | 5 | 0 |
| G302 | 5 | 5 | 0 |
| G304 | 15 | 11 | -4 |
| G305 | 1 | 1 | 0 |
| G404 | 17 | 17 | 0 |
| G703 | 9 | 9 | 0 |
| G704 | 7 | 7 | 0 |
| **Total** | **63** | **59** | **-4** |

Les quatre réductions G304 sont rattachées aux ouvertures root-scoped et aux refus de symlink de R5-A/R5-B. Les 59 findings sont conservés dans les JSON finaux et les matrices de lots; aucune directive `nosec`, `nolint`, allowlist ou réduction artificielle du scope n’a été utilisée.

## Tests et gates disponibles

Les suites `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, `go build ./cmd/... ./internal/...` et `pnpm run check` ont retourné PASS dans la campagne finale. Gitleaks, OSV Go, OSV pnpm, Trivy filesystem et Syft ont également retourné PASS selon `R5_FINAL_GATES_RAW.log`. Govulncheck global a rencontré des snapshots `artifacts/` non compilables; la procédure corrigée source-only `govulncheck ./cmd/... ./internal/...` a retourné PASS et est conservée dans `R5_GOVULNCHECK_CORRECTED_RAW.log`.

| Gate ou environnement | Statut final | Limite |
|---|---|---|
| Semgrep, Grype, Shellcheck, Yamllint | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | outils absents; aucune installation privilégiée |
| Docker/Buildx image | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | daemon/environnement dédié non disponible |
| Camoufox/runtime ciblé | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | aucun runtime réel lancé |
| SystemVault natif | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | aucun environnement natif |
| proxy/cookies réels | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | loopback et fixtures synthétiques seulement |
| release | BLOCKED | autorisation et environnement de release absents |

Les invariants restent : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false` et `FORGELOCAL_PRODUCTION_READY=false`.

## Fichiers de preuve

Les raw logs, JSON Gosec, matrices individualisées, rapports par lot, baseline, threat model et procédures corrigées sont présents sous `evidence/GOSEC_R5/`. Le raw historique `evidence/SMOKE_INTEGRATED_PROXY/` est conservé localement mais exclu de Git et du package R5.

Ce rapport ne constitue ni une approbation de production, ni une clôture de sécurité, ni une validation native de navigateur, de vault, de proxy ou de release.

## Package et vérification

Le package R5 v1 doit être construit à partir du HEAD d’évidence finalisé après ce rapport, avec manifest non auto-référentiel, ZIP/TAR, sidecars de hash, bundle delta, extraction fraîche, clone public neuf et `git fsck --full`. Toute vérification qui échoue restera documentée comme FAIL; aucune archive R4 canonique ne doit être reconstruite ou remplacée.
