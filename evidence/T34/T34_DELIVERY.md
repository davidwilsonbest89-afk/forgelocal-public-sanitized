# T34 — livraison et qualification indépendante

**Commit :** `1ab112bce8189c54070c38e218783e6835b93a9e`
**Baseline :** T33 distant `693632791041fde14db14ec8982b8bff1060a8d3`
**Statut :** `T34_HARDWARE_PROJECTION_APPROVED_VERIFIABLE_LOCAL`

T34 expose sept capacités matérielles contractuelles en projection read-only : CPU, mémoire, stockage, GPU, affichage, adaptateurs réseau et thermique. Toutes restent `UNSUPPORTED` avec une note fixe. Aucun sondage de l’hôte n’est effectué.

La qualification dans la sandbox a obtenu `exit_code=0` pour `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check` et Gitleaks sur le diff exact depuis la baseline. Les tests dédiés vérifient le déterminisme, le refus de toute observation hôte et l’absence d’identifiants ou chemins système dans le JSON.

Aucun accès à `/proc`, `/sys`, `cpuinfo`, `meminfo`, hostname, numéro de série, adresse MAC, navigateur, runtime, proxy, cookie, secret, SystemVault ou release n’a été effectué. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.

**Étape suivante autorisée :** T35, Font Bundle avec inventaire, provenance et limites de licence, après publication et vérification du commit distant T34.
