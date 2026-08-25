# Validation Docker V6 — image hardening non-runtime

**Décision exacte :** `DOCKER_V6_IMAGE_VALIDATION_BLOCKED_ENGINE_BUILDX_UNAVAILABLE`

Le lot a été créé depuis le hardening Docker publié au commit `1d585abea8f3081d9b7890400551e4318e1a4753`, sans modifier la branche V6 gelée. Docker Engine et Buildx sont absents de l’environnement ; aucune construction, inspection d’image ou exécution de conteneur n’est donc déclarée.

## Baseline

| Champ | Valeur |
|---|---|
| Branche | `audit/t00-t42-docker-v6-independent-image-validation` |
| HEAD baseline | `1d585abea8f3081d9b7890400551e4318e1a4753` |
| Base V6 | `999374d99b7996504ba91e421850a2fe84afb78d` |
| Docker / dockerd | absents, codes `127` / `1` selon commande |
| Buildx / Podman / Buildah / nerdctl | absents |
| Trivy | `0.74.0` |
| Espace observé | environ `1.4G` disponible sur `/home/ubuntu` |

## Contrôles exécutés

`bash -n docker/entrypoint.sh` retourne 0. La forme supportée `trivy config --exit-code 1` détecte les deux Dockerfiles et retourne **0 misconfiguration**. Les directives statiquement visibles sont `--no-install-recommends`, `browseforge`, ownership explicite de `/app` et du home, `USER browseforge` et `HEALTHCHECK`. Le diff Dockerfiles depuis la base V6 retourne `diff --check=0`.

Deux commandes Trivy non supportées (`--no-progress`, puis `--scanners`) ont été conservées dans le journal comme diagnostics ; le contrôle statique final utilise la syntaxe réellement supportée et est vert.

## Contrôles non exécutés

Les contrôles suivants restent explicitement `NOT_RUN_ENGINE_UNAVAILABLE` : Buildx de chaque Dockerfile, `docker image inspect`, vérification de l’utilisateur dans l’image, permissions/ownership dans un conteneur, présence et exécution de HEALTHCHECK, labels/environnement sans secret, digest d’image, `trivy image --scanners vuln,secret,misconfig`, `docker history --no-trunc` et cleanup container/volume/processus.

Aucun navigateur, Camoufox, proxy, cookie, donnée utilisateur, SystemVault natif, migration, runtime de production ou release n’a été lancé. Aucun workaround d’installation Docker n’a été appliqué dans le sandbox.

## Condition de reprise

Un environnement dédié doit fournir un Docker Engine fonctionnel et Buildx. La reprise doit commencer par une nouvelle `DOCKER_V6_BASELINE_DISCOVERY_RAW.log`, puis produire le build, l’inspection, le run avec `--entrypoint` inerte, les scans image, l’historique et le cleanup. Tant que ces preuves ne sont pas disponibles, le statut reste bloqué.

**Owner :** plateforme / sécurité conteneurisation. **Gate :** aucune gate V6 n’est levée.
