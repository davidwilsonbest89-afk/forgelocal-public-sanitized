# DOCKER-HARDENING — lot séparé

**Base :** tag `t00-t42-v6-local-qualified-2026-08-25`, commit `999374d99b7996504ba91e421850a2fe84afb78d`  
**Branche :** `audit/t00-t42-docker-hardening-v6`  
**Périmètre :** `Dockerfile` et `docker/Dockerfile.run` uniquement ; aucun Core, Dashboard métier, wrapper historique ou gate modifié.

Le diagnostic Trivy V6 initial comptait six misconfigurations, trois par Dockerfile : DS-0002 utilisateur root, DS-0026 HEALTHCHECK absent et DS-0029 absence de `--no-install-recommends`. Le patch ajoute `browseforge` comme utilisateur non-root, prépare les répertoires écrits par l’entrypoint, ajoute un HEALTHCHECK sur `http://127.0.0.1:19280/api/health` et corrige les deux installations apt concernées avec `--no-install-recommends`.

Le patch passe `git diff --check`, `bash -n docker/entrypoint.sh` et Trivy config. Le rescan de configuration retourne **0 misconfiguration**. Le build Docker et `docker compose config` n’ont pas pu être exécutés car le binaire Docker n’est pas installé dans l’environnement de qualification ; ces codes 127 sont conservés dans `V6_DOCKER_HARDENING_RAW.log`.

Le passage non-root peut modifier le comportement du runtime protégé : Xvnc/KasmVNC, `vncpasswd`, les chemins `/app`, le cache navigateur et l’écriture du token doivent être validés dans une image réellement construite. En conséquence, ce lot est prêt pour revue mais n’est pas intégré à la branche V6 gelée et aucun runtime réel, candidat Camoufox, proxy réel, cookie réel ou donnée utilisateur n’a été lancé.

**Décision :** `DOCKER_HARDENING_PATCH_READY_PENDING_CONTAINER_BUILD_AND_INDEPENDENT_REVIEW`.
