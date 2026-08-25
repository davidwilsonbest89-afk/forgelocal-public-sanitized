# DOCKER-V6-INDEPENDENT-IMAGE-VALIDATION — manifeste interne

Manifeste non auto-référentiel : les fichiers de contrôle du manifeste et des checksums sont exclus de leur propre liste.

| Fichier | Rôle |
|---|---|
| `BUNDLE_VERIFY.log` | vérification du bundle Git |
| `CHANGELOG.md` | historique du lot |
| `DOCKER_V6_STATIC_CHECKS.log` | journal brut des contrôles statiques |
| `REGISTER.md` | registre du lot |
| `docker-v6-independent-image-validation-v6.delta.bundle` | bundle Git portable |
| `docker-v6-independent-image-validation-v6.delta.bundle.portable.sha256` | sidecar portable du bundle |
| `todo.md` | validations restantes conditionnées à Engine/Buildx |
