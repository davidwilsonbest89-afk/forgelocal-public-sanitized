# T36 — Détection de dérive des diagnostics

**Statut :** `T36_DIAGNOSTIC_DRIFT_APPROVED_VERIFIABLE_LOCAL`
**Prérequis :** T35 distant `891b2d166a30a96d5ca872b473251ed1a3706ba3`.
**Mode :** comparaison de contrôles redacted uniquement.

## Périmètre fermé

T36 compare deux snapshots constitués uniquement de couples nom/état de contrôle. Les références `baseline_id` et `current_id` sont opaques ; aucune valeur brute d’environnement, empreinte, adresse, chemin ou donnée runtime n’est comparée ou retournée.

Les écarts sont normalisés en `ADDED`, `MISSING` ou `CHANGED`, triés par clé. Une limite explicite `max_changes` produit `within_limit=false` dès que le nombre d’écarts dépasse le seuil. Un seuil négatif est ramené à zéro, ce qui maintient une décision fail-closed.

## Critères d’acceptation

| Critère | Attendu |
|---|---|
| Baseline | Identifiant explicite et distinct du snapshot courant |
| Comparaison | États déclarés uniquement, aucune valeur brute |
| Normalisation | `ADDED`, `MISSING`, `CHANGED` |
| Déterminisme | Findings triés lexicographiquement |
| Seuil | `within_limit` est faux au-delà du seuil |
| Qualification | Tests race, suite globale, vet, build, format et Gitleaks |

Le lot ne détecte pas une dérive de machine réelle et ne lance aucun runtime. T37 peut commencer après publication et vérification du commit T36.
