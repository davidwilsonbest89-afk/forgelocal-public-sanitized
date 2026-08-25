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


## Addendum R1

| Action | Résultat R1 | Statut |
|---|---|---|
| Baseline/lignée depuis clone public neuf | Tags, commits, parents, ancêtres et fsck consignés | DONE — preuve R1 |
| Tests globaux baseline/HEAD sous race | Deux exécutions exactes code 0 ; finding historique non reproduit | DONE — divergence historique documentée |
| Tests T28 ciblés sous race | Code 0 | DONE |
| OSV réel explicite | v1.9.2, code 1, 46 avis par `go.mod` baseline/head | DONE — findings à traiter séparément |
| Gitleaks plage non vide | 8 commits réels, code 0 par arbre | DONE |
| Gitleaks extraction ZIP | code 0, aucun leak | DONE |
| Gosec baseline/head | 6/6 findings historiques, new=0, resolved=0 | DONE |
| Documentation `update_url` et gate exact | Corrigée et contrôlée | DONE |
| Revue indépendante des artefacts R1 | Clone neuf, package, bundle et assertions à inspecter | PENDING |

Le statut T28 reste `T28_EVIDENCE_QUALIFICATION_R1_READY_FOR_INDEPENDENT_REVIEW`, jamais `APPROVED`.


## T28-FINAL-CLOSURE-PASS — 2026-08-25

| Action | Résultat | Statut |
|---|---|---|
| Reconciler HEAD GitHub depuis clone neuf | HEAD découvert sans sélection manuelle ; lignée et absence de diff `internal/` post-implémentation vérifiées | DONE |
| Réutiliser et vérifier ZIP/bundle R1 | Hashes, sidecars neutres, extraction, manifeste, checksums, bundle seedé, fsck et Gitleaks cohérents | DONE |
| Couverture ZIP corrompu/manifest/limites | Tests T28 ajoutés et passants | DONE |
| Conservation des permissions et `update_url` | Test ciblé passant ; permissions conservées, high-risk explicite, URL ignorée | DONE |
| Purge et lifecycle | Test ciblé passant ; version affectée/active protégée, révocation puis purge explicite | DONE |
| Intégrité post-import | Bug concret corrigé ; taille/digest vérifiés avant approbation, affectation et rollback | DONE |
| Revue propriétaire | Décision finale `T28_APPROVED_VERIFIABLE_LOCAL` à confirmer par le propriétaire | PENDING_OWNER_ACCEPTANCE |
| Runtime navigateur/SystemVault/proxy/cookies/release | Hors périmètre et non autorisé | NOT_AUTHORIZED |
| T29/T39/T40/T41/T42 | Ne pas démarrer avant acceptation finale T28 | BLOCKED |


## T28-POSTFIX-FINAL-CLOSURE — 2026-08-25

| Action | Résultat | Statut |
|---|---|---|
| Correctif taille/SHA-256 publié | `f0701da849ce0f9073397bb42ded5e2e76b29ef1` | DONE |
| Régressions Approve/Assign/Rollback | `e806d4e915d1b702362389411d8ac823551df044`, tests sous race code 0 | DONE |
| Requalification depuis clone/worktree neuf | `go test`, `go vet`, `go build`, `git diff --check` code 0 | DONE |
| Gitleaks diff non vide et extraction finale | code 0, aucun leak ; diagnostic 0-commit conservé séparément | DONE |
| Nouveau ZIP post-correctif | `4efda01771ed7af135769dfa68caa8bdc6f226ca7cad5bf894dcfce05f5c8923` | DONE |
| Nouveau bundle post-correctif | `e9f65a5b9a734933f20ecf13b05f73136e604dc006b50dfc0d40286d91262097` | DONE |
| Vérification fraîche des artefacts | sidecars neutres, unzip, extraction, manifest/checksums, bundle verify, clone seedé, fsck | DONE |
| Statut T28 | `T28_APPROVED_VERIFIABLE_LOCAL` | DONE — Core local seulement |
| Runtime navigateur, extension, proxy/cookies, SystemVault, migration, production, release | non exécutés et non autorisés | NOT_AUTHORIZED |
| T29/T39/T40/T41/T42 | ne pas démarrer dans cette mission | BLOCKED |
