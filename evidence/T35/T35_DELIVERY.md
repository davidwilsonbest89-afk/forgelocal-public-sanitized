# T35 — livraison et qualification indépendante

**Commit :** `6519e84c13f7ae89cae7711d1ff0e568b6b3767e`
**Baseline :** T34 distant `c7f66da0da6f547813d10826dfd7772ad7e0f4b6`
**Statut :** `T35_FONT_BUNDLE_DOCUMENTARY_INVENTORY_APPROVED_VERIFIABLE_LOCAL`

T35 fournit un inventaire documentaire fermé sans binaire de police. L’entrée par défaut est `font-bundle`, avec provenance `NOT_SUPPLIED`, licence `PENDING_REVIEW`, redistribution `BLOCKED` et état `NOT_INCLUDED`. Les chemins vides, absolus, traversants ou contenant des séparateurs sont refusés. Toute redistribution déclarée exige une licence `APPROVED`.

La qualification dans la sandbox a obtenu `exit_code=0` pour `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check` et Gitleaks sur le diff exact depuis la baseline. Les tests dédiés vérifient l’inventaire déterministe, l’absence de chemins hôte et les refus fail-closed.

Aucun fichier de police, chemin de police installé, notice tierce, fournisseur, réseau, navigateur, runtime, proxy, cookie, secret, SystemVault ou release n’a été utilisé. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.

**Étape suivante autorisée :** T36, détection de dérive avec baseline explicite et seuils documentés, après publication et vérification du commit distant T35.
