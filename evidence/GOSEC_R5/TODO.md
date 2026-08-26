# GOSEC-R5 — TODO de reprise développeur

## Conservation immédiate

Construire une seule version canonique du package R5 à partir du HEAD d’évidence finalisé, avec manifest non auto-référentiel, ZIP/TAR, sidecars de hashes, bundle delta, extraction fraîche, clone public neuf, checkout et `git fsck --full`. Conserver séparément toute vérification publique ultérieure et ne pas reconstruire les archives R4.

## Findings de sécurité encore ouverts

Les 59 findings Gosec finaux sont individualisés dans les matrices R5-A, R5-B et R5-C et restent ouverts. Une prochaine campagne peut traiter des contrôles concrets, mais ne doit utiliser ni `nosec`, ni `nolint`, ni allowlist globale, ni réduction de scope. Les G204/G704 doivent rester soumis à une revue des chemins réellement exécutés; les G115/G302/G404/G101 exigent une décision ou une preuve individuelle avant toute clôture.

## Gates d’environnement

Conserver `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` pour Semgrep, Grype, Shellcheck, Yamllint, Docker/Buildx, Camoufox/runtime ciblé, SystemVault natif et proxy/cookies réels. Ne pas lancer ces environnements par contournement. Une autorisation écrite séparée reste requise pour les protocoles natifs et la release.

## Invariants de périmètre

Ne pas démarrer T29. Ne pas rouvrir T28. Ne pas modifier T31–T38. Ne pas écrire sur `validation/operational-v1`, qui doit rester à son HEAD historique vérifié. Ne pas inclure `evidence/SMOKE_INTEGRATED_PROXY/` dans Git, les packages ou les bundles.

## Verdict à maintenir

```text
GOSEC_R5_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R5_PARTIAL_ENVIRONMENT_UNAVAILABLE
PUBLIC_RELEASE_BLOCKED
SCAN_BLOCKED_UNKNOWN
NATIVE_SYSTEMVAULT_NOT_TESTED
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```
