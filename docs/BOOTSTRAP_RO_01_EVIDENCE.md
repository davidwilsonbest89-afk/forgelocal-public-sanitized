# BOOTSTRAP-RO-01 — État de preuve

> **État : `BOOTSTRAP_RO_APPROVED_VERIFIABLE`.** La décision repose sur le run local assaini `T00` à `T05` r10, exécuté sur le commit Core `5123dc4115c45eaaa17a3a24b25fd32611c73642` et le checkpoint dashboard `347197e4`.

Le dossier de sortie local est volontairement exclu de Git afin de ne jamais versionner un code de bootstrap ou un token potentiel. L’archive r10 contient un manifeste `SHA256SUMS` **à chemins relatifs**, sans auto-référence, vérifié directement après extraction avec **19/19** hashes valides. Le protocole, les commandes et l’index de preuve sont versionnés dans [`BOOTSTRAP_RO_01_STRICT_EXECUTION.md`](BOOTSTRAP_RO_01_STRICT_EXECUTION.md).

| Garantie vérifiée | Source du run final | Résultat |
|---|---|---|
| Code unique, échange loopback, rejeu et expiration réelle | `T05_core_contract.log`, `T05_playwright.log` | **PASS** |
| Refus hors loopback | `T05_core_contract.log` et `T05_nonloopback.log` | **PASS** (`403 LOOPBACK_REQUIRED`) |
| URL et stockages navigateur propres | `T05_playwright.log` | **PASS** |
| Invalidation client après `401` | `T05_playwright.log` | **PASS** |
| Lectures redacted et mutation refusée | `T01_api_tests.json`, `T05_core_contract.log` | **PASS** |
| Core/dashboard liés à loopback, PID masqués | `T05_loopback_sockets.log` | **PASS** |
| Worktree cible propre et diff RC | `T00_git_status_short.log`, `T00_git_diff_check.log`, `T00_rc_paths.log`, `T00_environment.log` | **PASS** |
| Intégrité portable et scan du dossier | `SHA256SUMS`, vérification après extraction, Gitleaks | **PASS** (19/19, zéro fuite) |

Cette décision n’autorise ni lancement de Camoufox, ni mutation UI, ni SystemVault, ni modification du RC BACK-01. Le statut de release reste **`PUBLIC_RELEASE_BLOCKED`**, avec `SCAN_BLOCKED_UNKNOWN`, pilote suspendu et cinq gates publics en attente.
