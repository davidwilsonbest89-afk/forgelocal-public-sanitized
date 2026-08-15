# BOOTSTRAP-RO-01 — État de preuve

> **État : `BOOTSTRAP_RO_APPROVED_VERIFIABLE`.** La décision repose désormais sur le run local assaini `T00` à `T07`, exécuté sur le commit Core `5123dc4115c45eaaa17a3a24b25fd32611c73642` et le checkpoint dashboard `3ac77e71522b65cd1802ba8119fd7b0d78335c06`.

Le dossier de sortie local est volontairement exclu de Git afin de ne jamais versionner un code de bootstrap ou un token potentiel. Il est accompagné du manifeste `SHA256SUMS`, vérifié après exclusion de son propre fichier. Le protocole, les commandes et l’index de preuve sont versionnés dans [`BOOTSTRAP_RO_01_STRICT_EXECUTION.md`](BOOTSTRAP_RO_01_STRICT_EXECUTION.md).

| Garantie vérifiée | Source du run final | Résultat |
|---|---|---|
| Code unique, échange loopback, rejeu et expiration réelle | `T05_core_contract.log`, `T05_playwright.log` | **PASS** |
| Refus hors loopback | `T05_nonloopback.log` | **PASS** (`403 LOOPBACK_REQUIRED`) |
| URL et stockages navigateur propres | `T05_playwright.log` | **PASS** |
| Invalidation client après `401` | `T05_playwright.log` | **PASS** |
| Lectures redacted et mutation refusée | `T01_api_tests.json`, `T05_core_contract.log` | **PASS** |
| Core/dashboard liés à loopback, PID masqués | `T05_loopback_sockets.log` | **PASS** |
| Intégrité et scan du dossier | `SHA256SUMS`, `T07_manifest.log`, Gitleaks | **PASS** |
| Aucun delta RC dans les chemins gelés | `T06_rc_paths.log` et audit Git | **PASS** |

Cette décision n’autorise ni lancement de Camoufox, ni mutation UI, ni SystemVault, ni modification du RC BACK-01. Le statut de release reste **`PUBLIC_RELEASE_BLOCKED`**, avec `SCAN_BLOCKED_UNKNOWN`, pilote suspendu et cinq gates publics en attente.
