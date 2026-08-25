# CHANGELOG — DOCKER-NATIVE-NONRUNTIME-VALIDATION

## 2026-08-25

- Création du lot séparé depuis le hardening Docker publié au commit `1d585abea8f3081d9b7890400551e4318e1a4753`.
- Enregistrement de la découverte brute avant modification de preuve.
- Inspection limitée aux Dockerfiles, à l’entrypoint et aux contrôles natifs non-runtime.
- Confirmation de l’absence de Docker/Buildx dans l’environnement ; build, history d’image, Trivy image, permissions runtime et cleanup container restent bloqués.
- Maintien des gates V6, de l’interdiction de runtime de production et de l’absence de modification Core/Dashboard/wrappers historiques.
