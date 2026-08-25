# TODO — OSV-TOOLCHAIN-RECONCILIATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Obtenir un scanner OSV compatible Go 1.25.13 | les 46 lignes sont recalculées avec la version effective | Sécurité dépendances |
| Haute | Réexaminer toute CVE réellement applicable | correction minimale, tests ciblés et SBOM régénérées dans un lot dédié | Mainteneurs Go |
| Moyenne | Préserver les exceptions individuelles | advisory, package, version et condition de levée restent traçables | Revue indépendante |
| Moyenne | Comparer les binaires effectivement produits | preuve reproductible de la stdlib embarquée | Build/release evidence |

Aucune action ne doit désactiver globalement OSV, lancer un runtime ou lever une gate V6.
