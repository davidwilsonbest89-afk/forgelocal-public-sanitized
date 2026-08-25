# CHANGELOG — DOCKER-V6-INDEPENDENT-IMAGE-VALIDATION

## 2026-08-25

Baseline brute créée avant toute modification documentaire. Docker Engine, dockerd, Buildx, Podman, Buildah et nerdctl sont absents. Aucun build ni conteneur n’a été lancé.

Contrôles statiques exécutés : syntaxe de l’entrypoint = 0 ; `trivy config --exit-code 1` = 0 avec deux Dockerfiles détectés et zéro misconfiguration ; diff Dockerfiles depuis V6 = propre. Deux commandes Trivy initiales avec options incompatibles sont conservées comme diagnostics, sans être utilisées comme résultat final.

Décision maintenue : `DOCKER_V6_IMAGE_VALIDATION_BLOCKED_ENGINE_BUILDX_UNAVAILABLE`.
