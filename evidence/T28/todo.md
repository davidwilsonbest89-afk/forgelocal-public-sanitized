# TODO — T28 Extensions locales contrôlées

| Priorité | Action | Condition de clôture | Statut |
|---|---|---|---|
| Haute | Revue indépendante physique T28 | Clone neuf, vérification des sidecars/bundle/ZIP, extraction/checksums/fsck et inspection des assertions | PENDING |
| Haute | Rejouer la suite globale après résolution du finding runtime V6 | `go test -count=1 -race ./...` vert sans modifier artificiellement la configuration runtime | BLOCKED_BY_PREEXISTING_RUNTIME_FINDING |
| Moyenne | Installer/fournir OSV Scanner compatible | Commande OSV réelle exécutée et journalisée ; aucun `127` présenté comme PASS | ENVIRONMENT_PENDING |
| Moyenne | Revue humaine des 182 findings Gosec historiques | Classification individuelle et décision de risque hors T28 | INHERITED |
| Haute | Validation navigateur/compatibilité | Autorisation séparée et environnement dédié ; jamais dans ce lot local-only | NOT_AUTHORIZED |
| Haute | Extension runtime/store | Décision produit et gate runtime distincts | NOT_AUTHORIZED |
| Haute | T29 Password Manager | Décisions explicites SystemVault et secrets ; ne pas commencer ici | BLOCKED |
| Haute | T39–T42 | Dépendances produit et autorisation écrite distinctes | BLOCKED |

Aucun item runtime, navigateur, production ou release ne doit être basculé à `DONE` par cette branche.
