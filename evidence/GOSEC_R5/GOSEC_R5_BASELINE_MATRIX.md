# GOSEC-R5 baseline matrix

Cette matrice est générée depuis le JSON Gosec source-only du commit `079c452`. Elle contient une ligne par finding; aucune ligne n’est supprimée ou clôturée automatiquement.

| Rule | Count | Lot |
|---|---:|---|
| G101 | 1 | R5-C |
| G115 | 3 | R5-C |
| G204 | 5 | R5-B |
| G302 | 5 | R5-C |
| G304 | 15 | R5-A |
| G305 | 1 | R5-A |
| G404 | 17 | R5-C |
| G703 | 9 | R5-A |
| G704 | 7 | R5-B |
| **Total** | **63** | **R5-A/R5-B/R5-C** |

Les décisions autorisées après triage sont `CORRECTED_AND_VERIFIED`, `MITIGATED_CONTROL_SCANNER_OPEN`, `SCANNER_OPEN_MANUAL_REVIEW`, `HISTORICAL_NOT_REACHABLE` et `BLOCKED_ENVIRONMENT_REQUIRED`. La valeur initiale `OPEN_REVIEW` est provisoire et doit être remplacée avant le rapport final.
