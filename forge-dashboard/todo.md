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
- [x] Exécuter Gitleaks et Gosec sur le delta réellement modifié, sans masquer les résultats historiques : Gitleaks 0 fuite, Gosec 0 finding sur lignes ajoutées et 72 findings historiques maintenus sous `SCAN_BLOCKED_UNKNOWN`.
- [x] Archiver les logs bruts, générer un manifeste SHA-256 non auto-référentiel et le rapport consolidé : `forgelocal-reimplementation-evidence-r3.zip`, hash conservé exclusivement dans son sidecar externe `.zip.sha256`, manifeste 17/17.
- [ ] Committer, taguer et pousser exclusivement la réimplémentation vérifiée, en conservant `PUBLIC_RELEASE_BLOCKED`.
