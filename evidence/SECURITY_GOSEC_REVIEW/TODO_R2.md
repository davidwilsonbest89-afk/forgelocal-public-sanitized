# TODO — GOSEC-REVIEW-R2

## Lot 1 — filesystem, archives et provenance

- [x] Baseline dédiée et périmètre fermé.
- [x] Tests traversal, chemin absolu, séparateurs Windows, profondeur, symlink, hardlink et archive corrompue/partielle.
- [x] Limites d’entrées et staging transactionnel.
- [x] Tests race, vet, build et tests ciblés.
- [x] Scan Gosec source-only, Gitleaks, Govulncheck, OSV et Trivy.
- [x] Matrice individuelle et conservation ZIP/TAR/bundle/manifest/clone/fsck.
- [x] Publication du commit source `cd0d2e6`.

## Restant avant toute clôture sécurité

- [ ] Revoir les findings G703/G304/G305/G122 encore ouverts; notamment `WalkDir` et les chemins filesystem hors extraction.
- [ ] Lot 2 : G204 et G704 réseau/subprocess avec allowlists, absence de shell implicite, timeouts et arrêt des processus.
- [ ] Lot 3 : G301/G302/G306/G104 permissions et erreurs I/O.
- [ ] Lot 4 : G115/G404 bornes et aléatoire non cryptographique.
- [ ] Lot 5 : G101 contexte exact et redaction.
- [ ] Analyser les 46 avis OSV du `go.mod` et documenter toute exception temporaire.
- [ ] Rendre disponibles Semgrep, Grype, Shellcheck et Yamllint, ou conserver `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` par outil.
- [ ] Ne pas déclarer `FORGELOCAL_PRODUCTION_READY`.
- [ ] Ne pas redémarrer T28 ni démarrer T29; ne pas toucher T31–T38.

## Lot 2 — subprocess et réseau

- [x] Baseline dédiée avec HEAD, queue G204/G704, ports, processus et outils disponibles.
- [x] Validation CLI loopback IPv4/IPv6/localhost et refus des URL externes.
- [x] Refus userinfo/query/fragment/ports invalides et redirections externes.
- [x] Timeout HTTP local et garde-fou `open --base-url`.
- [x] Dial WebSocket borné, deadline de handshake et tests du validateur.
- [x] Subprocess xattr Darwin converti en `CommandContext` timeouté; exécution native non disponible.
- [x] Tests race, vet, build, Gosec, Govulncheck, Gitleaks, OSV et Trivy.
- [ ] Bridge Playwright complet avec session réelle et cycle de vie du subprocess GUI.
- [ ] Revue manuelle des trois findings G204 `openBrowser`.
- [ ] Les 12 findings G204/G704 statiques signalés restent ouverts/classifiés.
- [ ] Analyse des 46 avis OSV `go.mod`.
- [ ] Outils indisponibles : Semgrep, Grype, Shellcheck et Yamllint.

Le Lot 2 reste classifié avec findings ouverts; aucun statut `COMPLETE_NO_OPEN_FINDINGS` n’est autorisé à ce stade.
