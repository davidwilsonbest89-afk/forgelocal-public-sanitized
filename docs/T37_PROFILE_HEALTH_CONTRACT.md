# T37 — Profile Health agrégé

**Statut :** `T37_PROFILE_HEALTH_PROJECTION_APPROVED_VERIFIABLE_LOCAL`
**Prérequis :** T36 distant `099ff7bbf384ff900ec44fc3e9aafd1fc273f0dd`.
**Mode :** agrégation read-only, redacted et déterministe.

## Périmètre fermé

T37 agrège des états de contrôles déjà redacted en un état `HEALTHY`, `AT_RISK`, `BROKEN` ou `UNKNOWN`. Les priorités sont fail-closed : un échec produit `BROKEN`, un avertissement produit `AT_RISK` sauf si un échec existe, et un contrôle `UNSUPPORTED` sans autre signal produit `UNKNOWN`.

Les explications utilisent uniquement les codes fixes `CHECK_FAILED`, `CHECK_WARNING` et `OBSERVATION_UNSUPPORTED`. Aucun détail de profil, valeur d’environnement, identifiant runtime, User-Agent, chemin ou donnée d’hôte n’est renvoyé.

## Critères d’acceptation

| Critère | Attendu |
|---|---|
| Agrégation | Priorité déterministe FAIL > WARNING > UNSUPPORTED |
| Déterminisme | Contrôles ordonnés et explications uniques |
| Redaction | Codes fixes uniquement, sans valeurs brutes |
| Lecture seule | Aucune écriture, réseau, runtime ou mutation |
| Qualification | Tests race, suite globale, vet, build, format et Gitleaks |

T37 reste une projection locale et ne constitue pas une conclusion de santé réelle d’un profil. T38 peut commencer après publication et vérification du commit T37.
