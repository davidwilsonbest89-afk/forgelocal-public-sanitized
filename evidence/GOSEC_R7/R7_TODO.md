# GOSEC-R7 — todo de reprise

## Décision actuelle

R7 n’a démontré aucun P0 ni P1. Aucun correctif de production supplémentaire ne doit démarrer automatiquement. Une décision propriétaire est requise pour tout éventuel lot P2/P3.

## Findings ouverts

Les 46 findings restent scanner-visibles avec leurs classifications et priorités dans `R7_FINDING_MATRIX.tsv`. Les mitigations ne sont pas des clôtures. Les G204 restent dépendants d’une validation native; les G404 et G101 restent en revue manuelle; les G115, G302, G305, G703 et G704 restent mitigés mais ouverts.

## Environnements

Semgrep, Grype, Shellcheck, Yamllint, Docker/Buildx, Camoufox, SystemVault natif, Windows, macOS et validation GUI native restent indisponibles ou non exécutés. Ne pas simuler ces environnements et ne pas les transformer en PASS.

## Invariants

Ne pas recréer ni réécrire les packages R6-A/B/C. Ne pas modifier R5. Ne pas rouvrir T28. Ne pas démarrer T29. Ne pas modifier T31–T38. Ne pas utiliser de compte, cookie, secret, proxy commercial, site externe ou donnée utilisateur. Ne jamais inclure `evidence/SMOKE_INTEGRATED_PROXY/`.

## Verdict à maintenir

```text
GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R7_PARTIAL_ENVIRONMENT_UNAVAILABLE
PUBLIC_RELEASE_BLOCKED
SCAN_BLOCKED_UNKNOWN
NATIVE_SYSTEMVAULT_NOT_TESTED
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```
