# GOSEC-R6 — changelog consolidé

La revue consolidée a vérifié physiquement le package R6-C depuis un clone GitHub neuf et a recalculé le JSON Gosec source-only au HEAD `f84738250fff6a91697629f899689e98da95da52`. Le total actuel est confirmé à 46 : G101=1, G115=3, G204=5, G302=5, G304=0, G305=1, G404=17, G703=7 et G704=7.

Les 11 G304 et les deux G703 supprimés par R6-A sont conservés comme historique de réduction, mais ne sont pas recréés dans la matrice actuelle. Les tests d’archives, de permissions, de token, de réseau et de workflow, la suite race, vet et build ont retourné exit code 0. Gosec retourne exit code 1 avec les 46 findings attendus. Gitleaks source-only et extraction corrigée `--no-git`, Govulncheck, OSV Go/pnpm et Trivy sont documentés avec leurs résultats réels.

Aucun code n’a été modifié pendant la revue consolidée. Aucun P0 ou P1 n’a été démontré sur les chemins synthétiques autorisés. Les G204 restent dépendants d’une validation native pour `open`, `xdg-open`, `rundll32` et `xattr`; les G404 et G101 restent en revue manuelle; les autres findings restent mitigés mais scanner-visibles.

Les archives R6-C n’ont pas été reconstruites. Les nouveaux fichiers de revue sont des preuves complémentaires publiées séparément.
