# Correctifs P0 UI — Core non connecté

- [x] Ajouter un indicateur global persistant « Démonstration locale — Core non connecté ».
- [x] Identifier les métriques et lignes de profils comme données de démonstration.
- [x] Désactiver et expliquer les mutations de profil tant que le Core est indisponible.
- [x] Rendre Camoufox candidat explicitement non lançable avant qualification indépendante.
- [x] Vérifier la compilation, le rendu et créer un checkpoint de correction.

## Réimplémentation clean-room — clôture contrôlée T12–T16

- [x] Réconcilier les contrats T13/T14 avec les projections API redacted et le dashboard actif.
- [x] Rejouer la suite Playwright complète en mode séquentiel : 22/22 scénarios passants.
- [x] Exécuter les contrôles globaux Core : `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`.
- [x] Exécuter Gitleaks sur le delta réellement modifié : 0 fuite. Le premier filtre Gosec est remplacé par R5, qui identifie honnêtement 7 findings sur les lignes ajoutées de la plage immuable.
- [x] Archiver les logs bruts, générer un manifeste SHA-256 non auto-référentiel et le rapport consolidé : `forgelocal-reimplementation-evidence-r3.zip`, hash conservé exclusivement dans son sidecar externe `.zip.sha256`, manifeste 17/17.
- [ ] Committer, taguer et pousser exclusivement la réimplémentation vérifiée, en conservant `PUBLIC_RELEASE_BLOCKED`.

## Correction documentaire R5 — audit indépendant

- [x] Joindre le bundle réel, son sidecar, sa vérification et le sidecar ZIP à la preuve R5.
- [x] Rejouer un clone neuf au tag avec commandes complètes, commit, répertoire, codes de sortie et journaux non vides.
- [x] Refaire Gitleaks et Gosec sur la plage immuable `fe042c58fe3b62197524d631e42496042af70532..419f32497e0b41d78ddfdd77d3dbb5b5b99aab19`.
- [x] Archiver le périmètre de fichiers, les JSON bruts et un filtre de findings paramétré par la plage Git.
- [x] Produire R5 avec manifeste non auto-référentiel, sans code produit, sans release et sans modifier les statuts bloquants.

## G6 — correction ciblée des findings Gosec du delta

- [ ] Cartographier les 5 G304 de `internal/localvault/localvault.go`, le G304 de `internal/runtime/qualification.go` et les 2 G104 de LocalVault.
- [ ] Appliquer exclusivement les correctifs minimaux ou les justifications de chemin de confiance explicitement testées.
- [ ] Ajouter les tests de sécurité et de non-régression nécessaires aux accès fichiers et à leur fermeture.
- [ ] Exécuter `go test -count=1 -race ./...`, vet, build, Gitleaks et Gosec sur un clone frais.
- [ ] Créer uniquement après les contrôles un commit G6, un tag local de conservation, bundle, clone neuf et archive finale non auto-référentielle.
