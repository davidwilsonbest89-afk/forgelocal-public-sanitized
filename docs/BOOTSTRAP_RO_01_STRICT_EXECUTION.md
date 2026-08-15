# BOOTSTRAP-RO-01 — Protocole d’exécution stricte

## Objet et périmètre

Ce protocole remplace une simple déclaration narrative par des sorties d’exécution locales, rejouables, assainies et hachées. Il ne couvre que le bootstrap lecture seule ; il ne qualifie aucun runtime, n’active aucune mutation UI et ne modifie aucun artefact RC.

> Les valeurs de code, token court, Bearer principal et `Authorization` restent hors des sorties. Un run qui les imprimerait doit échouer au contrôle de redaction.

## Préconditions

Le Core est exécuté depuis un worktree détaché du commit cible. Le dashboard reste un projet séparé, fourni explicitement par `DASHBOARD_DIR`. Les deux services principaux sont liés à loopback ; une seconde instance non loopback est créée uniquement comme témoin de refus, dans un répertoire temporaire distinct et avec `--no-runtime`.

| Élément | Valeur du run contrôlé |
|---|---|
| Commit Core cible | `5123dc4115c45eaaa17a3a24b25fd32611c73642` |
| Checkpoint dashboard | `347197e4` |
| Go | `go1.25.13`, `GOTOOLCHAIN=local` |
| Node | `v22.13.0` |
| Playwright | `1.55.0` |
| Répertoire de preuve | `/tmp/forgelocal-bootstrap-ro-strict-20260815-r10` |

## Commande rejouable

Depuis un clone ForgeLocal disposant du dashboard à côté, la commande suivante recrée un dossier de preuve temporaire. Les ports doivent être libres et le résultat ne doit pas être committé.

```bash
cd /chemin/vers/ForgeLocal
DASHBOARD_DIR=/chemin/vers/forgelocal-dashboard \
BOOTSTRAP_RO_EVIDENCE_DIR=/tmp/forgelocal-bootstrap-ro-strict-YYYYMMDD \
FORGELOCAL_TARGET_COMMIT=5123dc4 \
FORGELOCAL_E2E_PORT=20280 \
FORGELOCAL_NONLOOP_PORT=20281 \
FORGELOCAL_DASHBOARD_PORT=3501 \
./scripts/collect-bootstrap-ro-evidence.sh
```

Le collecteur produit un worktree détaché, exécute les tests Go ciblés et le race detector, démarre les deux Core et le dashboard local, puis lance Playwright. Le manifeste `SHA256SUMS` exclut expressément son propre fichier, référence uniquement des chemins relatifs et est vérifiable avec `sha256sum -c` immédiatement après extraction de l’archive.

## Index des preuves du run final

| Fichier | Démonstration | Résultat constaté |
|---|---|---|
| `T01_test_list.log` | Noms sélectionnés : `TestReadOnlyBootstrapIsLoopbackOnlySingleUseAndScopeLimited`, `TestReadOnlySessionBrokerExpiresCodesAndTokens`, `TestReadOnlyRoutesRequireCoreBearerAndReturnRequestID` | **PASS** |
| `T00_environment.log` | Commit cible et parent, versions des outils, ainsi que résultats explicites des contrôles Git | **PASS** |
| `T00_git_status_short.log`, `T00_git_diff_check.log`, `T00_rc_paths.log` | Sorties brutes vides attendues : worktree propre, diff formaté et zéro chemin RC modifié entre le parent et le commit cible | **PASS** |
| `T01_api_tests.json` | Tests API réellement exécutés avec sortie JSON | **PASS** |
| `T02_api_race.json` | Même contrat sous `go test -race` | **PASS** |
| `T03_cli_tests.json` | Code CLI sans Bearer et URL dashboard sans token | **PASS** |
| `T05_loopback_sockets.log` | Core et dashboard liés à `127.0.0.1`, PID masqués ; témoin nonloop séparé | **PASS** |
| `T05_core.log` et `T05_dashboard.log` | Journaux des deux services locaux, sans `Authorization`, Bearer ou valeur de token | **PASS** |
| `T05_playwright.log` | Échange, rejeu `401`, expiration réelle `401`, stockage vide, `401` forcé et déconnexion UI | **PASS** |
| `T05_nonloopback.log` | Échange témoin hors loopback refusé par `403 LOOPBACK_REQUIRED` | **PASS** |
| `T05_core_contract.log` | TTL 600 s/900 s, lectures redacted et profil `0 → 0` après écriture refusée | **PASS** |
| `T05_log_redaction.log` | Absence des motifs interdits dans les logs T05 | **PASS** |
| `SHA256SUMS` | Manifeste relatif sans auto-référence, 19 artefacts vérifiés avant archivage et après extraction | **PASS** |

## Contrôles complémentaires de l’examinateur

Le dossier final a été contrôlé par `sha256sum -c SHA256SUMS` avec 19 vérifications positives, y compris après extraction de l’archive. Un scan Gitleaks en mode `--no-git` sur le dossier de preuve a retourné un rapport JSON vide. `T00` archive les sorties brutes de `git status --short`, `git diff --check` et du diff des deux racines RC entre le parent `b5866b1` et le commit cible ; leurs résultats explicites sont également consignés dans `T00_environment.log`.

La première génération du manifeste incluait par erreur `SHA256SUMS` lui-même et employait des chemins absolus. Ces deux défauts rendaient la vérification après extraction ambiguë. Le collecteur les corrige désormais avant le run r10 : il arrête les services avant hachage, référence des chemins relatifs, exclut son propre manifeste et associe le contrôle hors loopback au résultat T05 global ainsi qu’à sa preuve dédiée. Cette correction est explicitement conservée ici plutôt que masquée.

## Décision

> **`BOOTSTRAP_RO_APPROVED_VERIFIABLE`** — Les dix contrôles de BOOTSTRAP-RO-01 sont étayés par des sorties exécutables, des logs assainis, une E2E navigateur locale, un manifeste de hash valide et une vérification des chemins RC.

La suite autorisée est limitée au raccordement lecture seule de **Groupes** et **Runtimes**, selon le même client mémoire seule. **Proxys**, **Backups** et **Audit** peuvent être préparés sans mutation. Toute release publique, activation de Camoufox, SystemVault ou levée de `PUBLIC_RELEASE_BLOCKED` demeure interdite.
