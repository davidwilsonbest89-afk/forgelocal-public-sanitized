# DOCKER-HARDENING — découverte

`Dockerfile` est une image de développement basée sur `golang:1.26-bookworm`, installe Node et les dépendances Playwright, puis construit le serveur, les fingerprints et les tests. Il n’a pas d’`ENTRYPOINT` ni de `USER` explicite et ses étapes apt ne demandent pas `--no-install-recommends`.

`docker/Dockerfile.run` est l’image d’exécution basée sur `ubuntu:24.04`. Elle installe les dépendances graphiques, KasmVNC et l’artefact BrowseForge, expose `19280` et `6901`, copie `docker/entrypoint.sh` et démarre ce script en tant que root par défaut. Le script écrit sous `/app/{profiles,data,browsers,logs,backups}`, utilise `vncpasswd`, lance Xvnc avec `-interface 0.0.0.0`, crée le token sous `/app/data/.api-token`, puis lance BrowseForge.

Le passage non-root dans `Dockerfile.run` nécessite donc une préparation des répertoires `/app` et `/root` ainsi qu’une vérification des privilèges Xvnc/KasmVNC et des ports. Il peut modifier le runtime protégé ; le lot Docker est par conséquent séparé du gel V6. Aucun changement n’est intégré au tag V6, aucun runtime réel ou candidat Camoufox n’est lancé, et l’intégration finale reste conditionnée à une autorisation explicite et à un test d’image loopback contrôlé.
