# GOSEC-R7 — changelog

R7 a été créée depuis le HEAD R6 distant découvert dynamiquement `3656dbad4bfef0381e1f9d837271d293ecffe292`. La vérification physique du package R6-C depuis un clone neuf est PASS : hashes, ZIP/TAR, manifestes, extraction, Gitleaks `--no-git`, bundle, checkout et `git fsck --full`.

Le JSON Gosec source-only R7 est identique au JSON R6-C autoritatif, SHA-256 `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a`, avec 46 findings : G101=1, G115=3, G204=5, G302=5, G304=0, G305=1, G404=17, G703=7 et G704=7.

Les tests d’archives, runtime, permissions, token, réseau et workflow, la suite race, vet et build ont tous retourné exit code 0. Gosec reste en exit code 1 avec les 46 findings. Gitleaks source-only/extraction, Govulncheck, OSV Go/pnpm, Trivy et Syft sont PASS dans les périmètres documentés. Semgrep, Grype, Shellcheck, Yamllint, Docker/Buildx, Camoufox, SystemVault natif, Windows, macOS et exécution GUI native restent indisponibles ou non exécutés.

Aucun code n’a été modifié dans R7. Aucun P0 ou P1 n’a été démontré. Les contrôles applicatifs et les findings scanner-visible sont documentés sans clôture artificielle. Les packages R6 ne sont pas reconstruits.
