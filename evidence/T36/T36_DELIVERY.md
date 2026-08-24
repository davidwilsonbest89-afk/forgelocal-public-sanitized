# T36 — livraison et qualification indépendante

**Commit :** `5085e6ac95d11649a74fb7137c6fca0b5ce7253a`
**Baseline :** T35 distant `891b2d166a30a96d5ca872b473251ed1a3706ba3`
**Statut :** `T36_DIAGNOSTIC_DRIFT_APPROVED_VERIFIABLE_LOCAL`

T36 compare uniquement des états de contrôles redacted entre une baseline et un snapshot courant. Les écarts sont normalisés en `ADDED`, `MISSING` et `CHANGED`, triés par clé, puis comparés à une limite explicite `max_changes`. Les valeurs brutes ne sont ni comparées ni retournées.

La qualification dans la sandbox a obtenu `exit_code=0` pour `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check` et Gitleaks sur le diff exact depuis la baseline. Les tests dédiés couvrent la stabilité, l’ordre déterministe, les trois types d’écart et le refus fail-closed d’un seuil négatif.

Aucun environnement réel, navigateur, runtime, réseau, proxy, cookie, secret, SystemVault ou release n’a été utilisé. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.

**Étape suivante autorisée :** T37, Profile Health agrégé read-only avec explications redacted, après publication et vérification du commit distant T36.
