# GOSEC-R4 todo

## Findings restant ouverts

La matrice finale conserve 63 findings Gosec : G101=1, G115=3, G204=5, G302=5, G304=15, G305=1, G404=17, G703=9 et G704=7. Aucun de ces findings ne doit être transformé en PASS sans nouvelle preuve ou décision de revue explicite.

Les priorités suivantes sont recommandées : revoir individuellement les G703/G304/G305 des chemins utilisateur et archives; décider si les G115 peuvent être supprimés par une forme de preuve statique acceptable; confirmer la politique des permissions d’exécutables runtime; documenter formellement l’usage non cryptographique des G404; et faire relire le G101 admin-token.

## Gates séparées

Docker/Buildx, Camoufox natif, Darwin/xattr/GUI, SystemVault natif, proxy/cookies réels et release restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` ou bloqués par leur gate. Ces validations ne doivent pas être simulées dans la sandbox.

## Invariants produit

T28 ne doit pas être rouvert dans ce lot. T29 ne doit pas démarrer. T31–T38 restent préservés. Toute nouvelle campagne doit partir d’une autorisation étroite, d’un nouveau baseline et d’une branche dédiée.
