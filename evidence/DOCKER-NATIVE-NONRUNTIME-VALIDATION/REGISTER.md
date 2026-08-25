# Registre — DOCKER-NATIVE-NONRUNTIME-VALIDATION

| Champ | Valeur |
|---|---|
| Lot | `DOCKER-NATIVE-NONRUNTIME-VALIDATION` |
| Branche | `audit/t00-t42-docker-native-nonruntime-validation` |
| Base | `1d585abea8f3081d9b7890400551e4318e1a4753` |
| Source hardening | `audit/t00-t42-docker-hardening-v6` |
| V6 gelé | `999374d99b7996504ba91e421850a2fe84afb78d` |
| Périmètre | Dockerfiles, entrypoint et validations natives non-runtime uniquement |
| Docker/Buildx | indisponible, exit 127 |
| Tests réalisés | inspection statique, syntaxe shell, Trivy config si disponible |
| Tests bloqués | build, history d’image, Trivy image, permissions runtime, cleanup container |
| Statut | `DOCKER_NATIVE_NONRUNTIME_VALIDATION_BLOCKED_PENDING_DOCKER_BUILDX` |
| Gate | inchangée ; aucune release et aucune intégration V6 |
| Owner | mainteneurs conteneurisation / sécurité plateforme |

La branche et le paquet devront être vérifiés par clone neuf, checksum, extraction, manifeste, bundle et Gitleaks. La limitation Docker doit rester visible dans chaque rapport public.
