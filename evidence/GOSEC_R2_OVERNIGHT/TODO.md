# TODO — GOSEC R2 overnight

## Conservation et lignée

- [x] Baseline UTC/CWD/OS/architecture/outils/branche/remote/HEAD/tags/worktree.
- [x] Vérification clone neuf, ancêtres, merge-base, diff et branche.
- [x] Conservation Lot 1 v2 non auto-référentielle.
- [x] Conservation Lot 2 : hashes, sidecars, ZIP/TAR, bundle, extraction, clone et fsck.

## Hardening exécuté

- [x] Filesystem complémentaire : séparateurs Windows, artifact root-scoped, symlink externe et modes backup.
- [x] Subprocess/réseau : URL loopback, redirections, timeouts et erreurs workflow.
- [x] Permissions/I/O : logs, config, migrations, backups, user-data, captures et préférences.
- [x] Directive Go patchée à 1.25.13 et OSV Go rejoué.
- [x] Tests ciblés, race, vet, build, Gosec, Govulncheck, Gitleaks, OSV, Trivy et Syft.

## Ouvert ou indisponible

- [ ] Gosec source-only : 132 findings restent signalés, dont G104/G204/G301/G302/G304/G306/G703/G704 et autres règles.
- [ ] Revue manuelle des trois lancements GUI `xdg-open/open/rundll32`.
- [ ] Bridge Playwright complet avec session réelle.
- [ ] Exécution native Darwin `xattr`.
- [ ] Semgrep, Grype, Shellcheck et Yamllint absents.
- [ ] Camoufox, SystemVault natif, Docker/Buildx et release non exécutés.
- [ ] Lot 3 non démarré; nouvelle autorisation et nouvelle baseline requises.

Le statut ne doit pas devenir `GOSEC_REVIEW_R2_COMPLETE_NO_OPEN_FINDINGS`.
