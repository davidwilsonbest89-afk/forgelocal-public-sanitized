# Registre — DOCKER-V6-INDEPENDENT-IMAGE-VALIDATION

| Champ | Valeur |
|---|---|
| Lot | `DOCKER-V6-INDEPENDENT-IMAGE-VALIDATION` |
| Source | hardening Docker V6 / `1d585abea8f3081d9b7890400551e4318e1a4753` |
| Branche | `audit/t00-t42-docker-v6-independent-image-validation` |
| Baseline brute | `DOCKER_V6_BASELINE_DISCOVERY_RAW.log` |
| Docker/Buildx | indisponibles ; aucune simulation |
| Contrôles statiques | `bash -n=0`, `trivy config=0` et 0 misconfiguration, diff Dockerfiles propre |
| Build image | non exécuté |
| Image inspect/run | non exécuté |
| Trivy image/history | non exécuté |
| Cleanup container | non exécuté |
| Décision | `DOCKER_V6_IMAGE_VALIDATION_BLOCKED_ENGINE_BUILDX_UNAVAILABLE` |
| Owner | plateforme / sécurité conteneurisation |
