# TODO — DOCKER-NATIVE-NONRUNTIME-VALIDATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Fournir Docker et Buildx | `docker --version` et `docker buildx version` disponibles dans un environnement isolé | Plateforme |
| Haute | Construire les Dockerfiles hardenés | build reproductible terminé sans démarrer de navigateur | Conteneurisation |
| Haute | Vérifier utilisateur et permissions | utilisateur effectif `browseforge`, écriture `/app`, home et entrypoint contrôlées dans l’image | Sécurité plateforme |
| Moyenne | Inspecter history et Trivy image | historique et scan image archivés, sans secret ni nouvelle capacité/port | Sécurité plateforme |
| Moyenne | Tester cleanup non-runtime | nettoyage contrôlé exécuté sans workflow de production | Mainteneurs Docker |

Aucune action de ce todo ne doit lancer Camoufox, un proxy réel, des cookies ou données utilisateur, SystemVault natif, une migration, un runtime de production, une release ou les gates interdites.
