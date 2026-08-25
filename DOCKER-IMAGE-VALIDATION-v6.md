# Docker image validation — lot séparé post-V6

**Décision :** `DOCKER_IMAGE_VALIDATION_BLOCKED_PENDING_REAL_DOCKER_BUILDX`

Le lot a été créé depuis le tag V6 gelé et ne modifie ni les Dockerfiles hardenés, ni l’entrypoint, ni le Core/Dashboard métier. Il constate l’état réel de l’environnement : `docker`, `dockerd`, `buildx`, Podman, Buildah et nerdctl ne sont pas disponibles ; aucun socket Docker n’existe.

## Contrôles réalisés

| Contrôle | Résultat |
|---|---|
| Inventaire Dockerfiles | `Dockerfile` et `docker/Dockerfile.run` inspectés |
| `git diff --check` | à exécuter sur le lot documentaire avant packaging |
| `bash -n docker/entrypoint.sh` | contrôle statique autorisé |
| Trivy configuration | contrôle statique autorisé ; résultat hérité du hardening précédent : 0 misconfiguration |
| Docker Buildx build | non exécuté : moteur absent |
| Image history / inspect | non exécuté : aucune image construite |
| Trivy image | non exécuté : aucune image disponible |
| Container HEALTHCHECK | non exécuté : aucun container démarré |
| utilisateur non-root / permissions en image | non exécuté : nécessite une image réelle |

## Interdictions maintenues

Aucun navigateur, Camoufox, proxy, cookie, donnée utilisateur, workflow de production, SystemVault natif, migration, release ou lot T28/T29/T39/T40/T41/T42 n’est lancé. Un remplacement par Podman ou une simulation n’est pas accepté comme preuve Docker.

Le lot reste **BLOCKED**, et non DONE, APPROVED ou NOT_APPLICABLE. La construction et l’inspection devront être rejouées dans un environnement Docker/Buildx réel, puis accompagnées de l’historique d’image, du scan Trivy image, de la vérification non-root, des permissions, du HEALTHCHECK et du cleanup container.

**Owner suivant :** plateforme / sécurité conteneurisation. **Condition de clôture :** daemon Docker et Buildx disponibles, build réussi, inspect et scans image archivés, tests sans navigateur ni production passés.
