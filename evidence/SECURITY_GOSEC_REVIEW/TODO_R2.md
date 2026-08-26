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
