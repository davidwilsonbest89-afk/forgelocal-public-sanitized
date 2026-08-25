# DOCKER-NATIVE-NONRUNTIME-VALIDATION — lot séparé

**Décision :** `DOCKER_NATIVE_NONRUNTIME_VALIDATION_BLOCKED_PENDING_DOCKER_BUILDX`

**Base contrôlée :** commit `1d585abea8f3081d9b7890400551e4318e1a4753` de la branche de hardening Docker isolée. Cette branche de validation ne modifie pas la branche V6 gelée et ne modifie pas les Dockerfiles hérités.

## Périmètre fermé

Le lot inspecte uniquement les propriétés natives des Dockerfiles et de l’entrypoint : utilisateur non-root, permissions, `HEALTHCHECK`, historique statique, dépendances apt, ports, capacités, secrets et comportement de cleanup observable par inspection. Aucun conteneur n’est démarré. Aucun navigateur, workflow de production, Camoufox, proxy, cookie, donnée utilisateur, SystemVault natif, migration, runtime ciblé, release, T28, T29, T39, T40, T41 ou T42 n’est lancé.

## Contrôles

Les Dockerfiles contiennent les corrections hardening héritées : `--no-install-recommends` aux emplacements apt signalés, utilisateur `browseforge`, ownership de `/app` et du home, `USER browseforge` et healthcheck loopback sur l’API. L’entrypoint est syntaxiquement inspectable et le contrôle Trivy de configuration du hardening précédent indiquait zéro misconfiguration.

L’environnement courant ne fournit ni binaire `docker` ni Docker Buildx/daemon. Il est donc impossible de produire honnêtement une image, d’inspecter son `history`, d’exécuter Trivy image, de vérifier les permissions dans une couche réellement construite, ou d’exécuter le cleanup dans un conteneur. Les commandes demandant Docker sont considérées comme bloquées, non comme réussies.

Un risque de compatibilité reste ouvert : `docker/entrypoint.sh` écrit sous `/app`, initialise `vncpasswd`, démarre Xvnc/KasmVNC et lance BrowseForge. Le passage non-root doit être vérifié dans un build et un démarrage contrôlés, sans navigateur ni workflow de production, lorsqu’un daemon Docker sera disponible.

## Condition de clôture

La clôture exige un environnement Docker/Buildx disponible, un build reproductible sans modification de fonctionnalité, une inspection de l’utilisateur effectif et des permissions, un contrôle d’historique, Trivy image, puis un test de cleanup non-runtime. Tant que ces preuves n’existent pas, le statut reste bloqué et aucune intégration à V6 n’est autorisée.

## Propriétaire

**Owner :** mainteneurs conteneurisation / sécurité plateforme. **Prochain lot :** validation indépendante Docker native, puis revue du risque non-root de l’entrypoint.
