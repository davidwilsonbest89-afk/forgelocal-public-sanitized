# GOSEC-R6 Lot A — todo de reprise

Le Lot A a publié le hardening source dans `142477ae0d576eae937b16660899fd973d6f2464`. Les preuves post-correctif sont dans `R6_A_FINAL_RETRY_RAW.log` et `gosec_r6a_after_retry.json`; la classification individuelle est dans `R6_A_FINDING_CLASSIFICATION.tsv`.

Les sept G703 restants et le G305 restent ouverts sous `MITIGATED_CONTROL_SCANNER_OPEN`. Le prochain travail doit démontrer, finding par finding, la source du chemin, le graphe d’appel, la racine attendue, le comportement traversal/symlink/hardlink/type spécial/TOCTOU et l’impact. Ne pas utiliser `nosec`, `nolint`, skip ou allowlist.

Le Lot B G204/G704 ne doit commencer qu’après la conservation vérifiée du Lot A. Les tests Dashboard/proxy/T28 sont des non-régressions existantes uniquement. T28 ne doit pas être rouvert, T29 ne doit pas démarrer et T31–T38 ne doivent pas être modifiés.

Les chemins Windows/macOS, Camoufox, SystemVault natif, Docker/Buildx et proxy/cookies réels restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`. Les statuts de campagne restent `GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
