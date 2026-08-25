# TODO — DOCKER-V6-INDEPENDENT-IMAGE-VALIDATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Fournir un Docker Engine et Buildx réels | `docker version` et `docker buildx version` réussissent dans un environnement dédié | Plateforme |
| Haute | Construire chaque Dockerfile | build Buildx terminé, image/tag/digest enregistrés | Conteneurisation |
| Haute | Inspecter l’image | non-root, HEALTHCHECK, labels, environnement sans secret, permissions et ownership vérifiés | Sécurité conteneur |
| Haute | Scanner l’image | `trivy image --scanners vuln,secret,misconfig` et `docker history --no-trunc` archivés | Sécurité |
| Haute | Contrôler le cleanup | aucun conteneur, volume, secret, token ou processus résiduel | Plateforme |
| Permanente | Ne pas simuler | tant que Engine/Buildx sont absents, les contrôles runtime restent `NOT_RUN_ENGINE_UNAVAILABLE` | Gouvernance qualité |

Aucun navigateur, workflow de production, release ou gate V6 ne doit être activé par ce lot.
