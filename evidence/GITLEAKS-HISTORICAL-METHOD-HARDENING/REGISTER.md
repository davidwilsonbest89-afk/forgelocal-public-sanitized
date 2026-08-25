# Registre — GITLEAKS-HISTORICAL-METHOD-HARDENING

| Champ | Valeur |
|---|---|
| Lot | `GITLEAKS-HISTORICAL-METHOD-HARDENING` |
| Branche | `audit/t00-t42-gitleaks-historical-method-hardening-v6` |
| Base | `999374d99b7996504ba91e421850a2fe84afb78d` |
| Plage de preuve | `b34fa5c02ff20144abfb5d240db1c67ad1f038f9..fc080456711dd7f2266911aaec55041fdb1b424c` |
| Commits inventoriés | 4 |
| JSON par arbre | oui, quatre fichiers archivés |
| Configuration | `.gitleaks.toml` versionné et passé explicitement |
| Règle zéro-commit | `INVALID_ZERO_COMMIT_RANGE`, jamais PASS |
| Décision | `GITLEAKS_HISTORICAL_TREE_BY_TREE_METHOD_HARDENED` |
| Owner | sécurité dépôt / responsable des preuves |

Le lot formalise une méthode et ne modifie pas la baseline V6 ni les résultats historiques. Toute future plage doit produire son propre inventaire et ses propres JSON.
