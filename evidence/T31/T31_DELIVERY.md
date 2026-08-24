# T31 — livraison et qualification indépendante

**Commit :** `9d40bb4210a76e1fe0574292fc010dd2ae52eb31`
**Baseline :** `t00-t27-complete-20260820` / `69411e65c880d168832a65fc8475cc97d562a9ad`
**Statut :** `T31_CONTRACT_AND_PROJECTED_CONTROLS_APPROVED_VERIFIABLE_LOCAL`

T31 ajoute cinq contrôles de projection explicites — Canvas 2D, Canvas WebGL, WebGL2, AudioContext et OfflineAudioContext — tous maintenus à `UNSUPPORTED` avec une note fixe. Aucune valeur Canvas, renderer WebGL, extension, signal audio, AudioBuffer, hash runtime ou autre observation brute n’est collectée.

La qualification indépendante sur le checkout T31 a exécuté `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check t00-t27-complete-20260820..HEAD`, Gitleaks sur le diff binaire exact et `git fsck --full`. Tous les codes de sortie sont `0`. Les tests dédiés vérifient la présence, l’ordre déterministe, l’état `UNSUPPORTED`, la note fixe et l’absence de valeurs brutes sérialisées.

Le lot reste local et read-only. Aucun navigateur, runtime, Camoufox, proxy réel, cookie, secret, SystemVault natif ou release n’a été lancé. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.

**Étape suivante autorisée :** T32, contrat ClientRects et tests de redaction, après vérification du commit distant T31.
