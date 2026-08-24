# T28–T42 — requalification depuis clone neuf

**Branche audit de correction :** `audit/t28-t42-evidence-correction`
**Clone neuf contrôlé :** `/home/ubuntu/forgelocal-t42-fresh-20260824`
**HEAD du clone neuf :** `6489af39a4ac4f91f9f7dc1435f10b2bd10dfdc0`
**Espace libre constaté au clonage :** `29452252 KiB`

Les contrôles `git fsck --full`, `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...` et `git diff --check t00-t27-complete-20260820..HEAD` ont retourné `exit_code=0` dans le clone neuf. Les journaux complets associés contiennent UTC, CWD, commande, HEAD, sortie brute, code de sortie et fin UTC.

Le scan Gitleaks cumulatif retourne `exit_code=1` avec le signal générique historique `APi=REDACTED`; le JSON et la sortie brute sont conservés sans transformation en succès. Gosec baseline et head retournent chacun `exit_code=1` à cause des constats historiques ; la comparaison normalisée par chemins relatifs produit `baseline_count=194`, `head_count=194`, `new_findings=[]` et `resolved_findings=[]`. Cette différence de code de sortie est documentée et non masquée.

Cette requalification ne lance aucun runtime réel, Camoufox, proxy, cookie, SystemVault natif, migration utilisateur ou release. Elle ne lève aucune gate et ne change pas les verdicts T28, T29, T39, T40, T41 et T42, qui restent `BLOCKED`. T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION` faute de branche GitHub canonique portant son head et son kit complet.
