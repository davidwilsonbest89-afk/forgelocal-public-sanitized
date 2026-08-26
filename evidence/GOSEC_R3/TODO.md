# GOSEC-R3 todo

## Findings ouverts

Les 94 findings Gosec source-only restants sont conservés dans `GOSEC_R3_FINAL_MATRIX.tsv` avec leur règle, fichier, ligne, classification et justification. Les priorités suivantes exigent une nouvelle autorisation étroite : G104, G302, G304, G703, G704, G404, G115, G107 et G101.

## Outils et environnements non exécutés

Semgrep, Grype, Shellcheck et Yamllint restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`. Camoufox natif, SystemVault natif, Docker/Buildx, GUI Darwin, proxy/cookies réels et release restent des gates séparées.

## Limites de mission

T29 ne doit pas démarrer automatiquement. T28 ne doit pas être rouvert dans cette séquence. T31–T38 doivent rester inchangés. Aucune déclaration production-ready ou release ne peut être faite.
