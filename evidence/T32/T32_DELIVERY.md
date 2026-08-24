# T32 — livraison et qualification indépendante

**Commit :** `17b6daf72995e6c36d0d14cd51f36b84138bacb9`
**Baseline :** `t00-t27-complete-20260820` / `69411e65c880d168832a65fc8475cc97d562a9ad`
**Statut :** `T32_CLIENTRECTS_PROJECTED_UNSUPPORTED_APPROVED_VERIFIABLE_LOCAL`

T32 ajoute le contrôle projeté `client-rects`, avec l’état `UNSUPPORTED` et une note fixe. Aucun rectangle DOM, `DOMRect`, coordonnée ou dimension n’est observé, calculé ou persisté.

La qualification exécutée dans la sandbox a obtenu `exit_code=0` pour `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check` et Gitleaks sur le diff binaire exact depuis la baseline. Les tests dédiés vérifient la présence, l’état, la note, le déterminisme et l’absence de géométrie dans le JSON.

Aucun navigateur, runtime, Camoufox, proxy, cookie, port, secret, UI, SystemVault natif ou release n’a été lancé. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.

**Étape suivante autorisée :** T33, QA géolocalisation synthétique, sans position réelle, après publication et vérification du commit distant T32.
