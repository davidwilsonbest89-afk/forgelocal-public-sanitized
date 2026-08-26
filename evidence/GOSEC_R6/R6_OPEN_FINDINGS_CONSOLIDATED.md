# GOSEC-R6 — matrice consolidée des findings ouverts

## Autorité et périmètre

La matrice est générée directement depuis `evidence/GOSEC_R6/R6_C/gosec_r6c_after_postcommit.json`, SHA-256 `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a`, au HEAD R6 public `f84738250fff6a91697629f899689e98da95da52`. Le JSON contient exactement 46 entrées et aucune entrée G304. La matrice TSV contient une ligne par finding courant; les 11 G304 et les deux G703 supprimés par R6-A ne sont pas recréés. Leur réduction historique est documentée dans `R6_REDUCTION_HISTORY.md`.

| Règle | Findings actuels |
|---|---:|
| G101 | 1 |
| G115 | 3 |
| G204 | 5 |
| G302 | 5 |
| G304 | 0 |
| G305 | 1 |
| G404 | 17 |
| G703 | 7 |
| G704 | 7 |
| **Total** | **46** |

Chaque ligne de `R6_OPEN_FINDINGS_CONSOLIDATED.tsv` contient règle, fichier, fonction, ligne, entrée, chemin réellement exécuté, actif protégé, préconditions, impacts confidentialité/intégrité/disponibilité, plateforme, garde, test négatif, résultat Gosec, classification, action et priorité.

Les contrôles existants sont documentés comme mitigations mais les findings restent scanner-visible. Aucun finding n’est clos sur la seule présence d’un garde.
