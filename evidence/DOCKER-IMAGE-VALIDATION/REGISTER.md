# Registre — DOCKER-IMAGE-VALIDATION

| Champ | Valeur |
|---|---|
| Lot | `DOCKER-IMAGE-VALIDATION` |
| Branche | `audit/t00-t42-docker-image-validation-v6` |
| Base | tag `t00-t42-v6-local-qualified-2026-08-25` / `999374d99b7996504ba91e421850a2fe84afb78d` |
| Docker CLI/daemon | absents |
| Buildx | absent |
| Inspection statique | Dockerfiles hardenés et entrypoint inspectés |
| Build image | non exécuté |
| Trivy image | non exécuté |
| Container/HEALTHCHECK | non exécuté |
| Décision | `DOCKER_IMAGE_VALIDATION_BLOCKED_PENDING_REAL_DOCKER_BUILDX` |
| Owner | plateforme / sécurité conteneurisation |

Aucune validation container ne peut être déduite d’un contrôle Trivy configuration ou d’une syntaxe shell.
