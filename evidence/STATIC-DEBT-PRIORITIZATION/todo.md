# TODO — STATIC-DEBT-PRIORITIZATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Revoir chaque finding sécurité/fiabilité | décision humaine, correction ou justification et test ciblé | Mainteneurs Go |
| Haute | Planifier les corrections élevées | lot futur séparé, sans modification de V6 gelée | Mainteneurs Go |
| Moyenne | Traiter les findings qualité restants | patch isolé et rerun Staticcheck/GolangCI-Lint | Mainteneurs Go |
| Moyenne | Maintenir le backlog à jour | règle, fichier, ligne, impact, sévérité, test et lot futur conservés | Qualité |

Aucun finding ne doit être marqué PASS uniquement par triage documentaire. Aucune fonctionnalité bloquée, gate ou release ne doit être réactivée par ce lot.
