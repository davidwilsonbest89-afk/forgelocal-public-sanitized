# R6 — correction de procédure de baseline

La première tentative, conservée dans `R6_BASELINE_DISCOVERY_RAW.log` et `gosec_r6_baseline.json`, a exécuté Gosec sans exporter `/usr/local/go1.25.13/bin` dans le `PATH`. Gosec a retourné `exit_code=1` et a produit uniquement des erreurs de chargement `go command required, not found`; ce JSON ne constitue pas une baseline de findings et n’est pas utilisé pour le triage.

La relance corrigée est conservée dans `R6_BASELINE_RETRY_RAW.log` et `gosec_r6_baseline_corrected.json`. Elle utilise `PATH=/usr/local/go1.25.13/bin:$PATH`, `GOTOOLCHAIN=local` et le périmètre source-only `./cmd/... ./internal/...`. Elle retourne les 59 findings attendus : G101=1, G115=3, G204=5, G302=5, G304=11, G305=1, G404=17, G703=9 et G704=7.

Cette correction n’a modifié aucun artefact R5 et n’est pas une suppression de finding. La matrice R6 est construite exclusivement depuis le JSON corrigé.
