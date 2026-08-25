# TODO — DOCKER-IMAGE-VALIDATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Fournir un daemon Docker et Buildx | `docker version` et `docker buildx version` réussissent | Plateforme |
| Haute | Construire les Dockerfiles hardenés | build reproductible sans navigateur ni production | Conteneurisation |
| Haute | Inspecter l’image | utilisateur, permissions, HEALTHCHECK, ports, capacités et history vérifiés | Sécurité conteneur |
| Haute | Scanner l’image | Trivy image et Gitleaks image archivés, sans secret | Sécurité |
| Moyenne | Tester le cleanup container | arrêt et suppression contrôlés, logs redacted conservés | Plateforme |

Aucun contrôle ne doit être marqué DONE sans exécution dans un moteur Docker/Buildx réel.
