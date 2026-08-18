# T19-R — Qualification intégrée Dashboard ↔ Core (post-T18-R)

**Statut : `T19-R_APPROVED_VERIFIABLE_LOCAL`** (parcours réel Dashboard ↔ Core passé 22/22, contrôles Core et scans passants)

## 1. Périmètre et non-restauration

T19-R est une **qualification intégrée clean-room de la lignée actuelle** (`63e167e8f543a25ad3df81d7003a7eae2fecaf70`, post-T18-R ; la base de qualification est `e6e4ebd96f12089dcc38d77226ddfacb8b6806c7`). Il ne restaure pas le T17 historique perdu ; `T17_SOURCE_SNAPSHOT_UNRECOVERABLE` demeure inchangé, tout comme le blocage historique T18 (`T18_BLOCKED_BASELINE_SOURCE_NOT_FOUND` pour la validation globale originale). Aucun artefact T00–T18 n'a été modifié.

## 2. Gap audit (Phase 1)

Le tableau complet figure dans `audit/T19R_GAP_AUDIT.md`. Les contrôles exigés comparés au contrat Core courant :

| Contrôle | Critère | Verdict |
|---|---|---|
| Bootstrap loopback à usage unique | Code émis, scellé, expirant, non rejouable | **PRESENT_AND_TESTED** (T05 bootstrap-ro.spec.ts : émission, relais, 401 au rejeu) |
| Mémoire seule | Aucun stockage persistant navigateur | **PRESENT_AND_TESTED** (localStorage=0, sessionStorage=0, indexedDB=0, caches=0 vérifiés dans le navigateur réel) |
| Expiration / rejeu / 401 | Session lecture seule expirée refusée | **PRESENT_AND_TESTED** (T05_REPLAY : status=401, UI déconnectée) |
| Redaction profils/groupes/runtimes/proxys/backups/sessions | Aucun secret/chemin brut dans les projections | **PRESENT_AND_TESTED** (T06, T09, T10, T11, T13, T14, T15, G15-A sessions) |
| Pagination conforme au Core | Page/limite passées au Core | **PRESENT_AND_TESTED** (listing profils groupe par le Core, specs T09/T10) |
| Refus hors loopback | OriginGuard fail-closed | **PRESENT_AND_TESTED** (proxies-t10 hors loopback : 403 ; T14 W3 : POST/DELETE 403/405) |
| Absence de mutation involontaire | Aucun write accidentel du dashboard lecture seule | **PRESENT_AND_TESTED** (T13 admin-only 401, T09 runtime_id inconnu 400, T14 W3) |
| Panneau runtime qualifié vs catalogue vide | Écrans T14/T15 | **PRESENT_AND_TESTED** (après activation du runtime Chromium qualifié — voir §4) |

Aucun comportement ABSENT, un point d'environnement test documenté au §4.

## 3. Modifications apportées

**Aucun code produit n'a été modifié.** Le dashboard testé est identique à celui du dépôt `forge-dashboard/` (diff récursif hors `node_modules`/`dist`/`.git`/`test-results` : **0 différence**). Le lot ajoute uniquement les preuves de qualification (logs, rapports, scripts de contexte d'exécution) dans `forge-dashboard/proofs/`.

## 4. Écart d'environnement documenté honnêtement

La consigne demandait le Core en `--no-runtime`. Ce mode désactive **tous** les runtimes et produit un catalogue de qualification vide (`{"runtimes":[]}`), ce qui rend impossibles les écrans T14 (catalogue qualifié) et T15 (profils liés à un runtime qualifié). Le premier run avec `--no-runtime` a donné 14 passed / 8 failed exactement pour cette raison (T14 W1–W3, T15 W1–W5). La qualification réelle T14 valide un Chromium local qualifié par son hash SHA-256 — le pilote local approuvé `PILOT_LOCAL_APPROVED — TEMPORARY` (session `PL-20260814-UBU2404AMD64-SYSTEMVAULT-001`) couvre le Chromium sandbox local ; aucun Camoufox, aucun proxy réseau réel n'a été lancé. Le run de qualification final a donc été exécuté avec le runtime `browseforge-chromium` **qualifié réellement** (Chromium 150.0.7871.101, état QUALIFIED, hash enregistré en SQLite) sur `127.0.0.1:19280`.

## 5. Qualification — commandes, timestamps UTC, répertoires, commits, exit codes

| Contrôle | Commande | UTC | Répertoire | Commit | Exit |
|---|---|---|---|---|---|
| Core tests -race | `go test -count=1 -race ./...` | 2026-08-18T17:38:32Z | `/home/ubuntu/forgebaseline-reimpl` | `e6e4ebd9` | **0** |
| Vet | `go vet ./...` | 2026-08-18T17:38:XXZ | idem | idem | **0** |
| Build | `go build ./...` | 2026-08-18T17:38:XXZ | idem | idem | **0** |
| Build dashboard | `pnpm run build` | 2026-08-18T16:4XZ | `/home/ubuntu/t19r-work/dashboard` | (dashboard de la baseline) | **0** |
| TypeScript | `npx tsc --noEmit` | 2026-08-18T16:4XZ | idem | idem | **0** |
| Playwright séquentiel | `npx playwright test --workers=1` | 2026-08-18T17:20:11Z | idem | idem | **0 — 22 passed (10.8m), 0 failed, 0 did not run** |
| Gitleaks delta | `gitleaks stdin` sur patch immuable | 2026-08-18T17:39:XXZ | idem | plage `55776a4a..e6e4ebd9` | **0 finding, exit 0** (5 277 octets scannés) |
| Gosec brut | `gosec -out gosec_t19r_raw.json -fmt=json ./...` | 2026-08-18T17:39:27Z | idem | `e6e4ebd9` | exit 1 (194 findings historiques conservés) |
| Gosec filtré | `r5_filter_gosec.py` plage immuable | idem | idem | idem | **0 finding sur 0 ligne Go ajoutée, exit 0** |
| Whitespace | `git diff --check` | idem | idem | idem | **0** |

Le log brut Playwright est `proofs/t19r-playwright-final.log` ; il contient les 22 titres de tests, les diagnostics T05_BROWSER_STORAGE et T05_REPLAY, et la ligne finale `22 passed (10.8m)`.

## 6. Détail des 22 tests E2E passés (résumé)

Les 22 tests couvrent T05 (bootstrap loopback à usage unique, mémoire seule, rejeu 401), T06 (groupes/runtimes), T09 (écritures profils), T10 (registre proxy + refus hors loopback), T11 (Backups), T13 (Identité navigateur redacted), T14 (runtime qualifié — W1 catalogue, W2 absence de chemins/ports/tokens, W3 read-only), T15 (automation CDP locale), ainsi que les 2 sondes standalone T10. Le token admin `FORGELOCAL_API_TOKEN` est lu en mémoire depuis le fichier, sans localStorage.

## 7. Conservation

Commits de preuves poussés sur `forgelocal-baseline-2026-08-17` : `bda77236…` (métadonnées JSON, audit, rapports), `96d8f71664c27cbb99b299623f1cd07f90212d9e` (logs bruts), `7d71ba6151eb0e91de694ef3d1a64131a850a60d` (rapport final v1) puis `63e167e8f543a25ad3df81d7003a7eae2fecaf70` (rapport final définitif — HEAD actuel). Tag annoté : `t19r-dashboard-core-integration-verified-2026-08-18` (pointe sur `63e167e8f543a25ad3df81d7003a7eae2fecaf70`, objet tag vérifiable via `git cat-file -p`), poussé. Bundle : `forgelocal-t19r-dashboard-core-7d71ba6.bundle`, SHA-256 `def2bdc031174cad7f06defeeadf234b96bd93b643e7c0a5841d09ae8ab0c24f`, `git bundle verify` OK. Clone neuf exclusivement depuis le bundle : `HEAD=7d71ba6151eb0e91de694ef3d1a64131a850a60d`, tag vérifié, `git fsck --full` exit 0, qualification Core rejouée (exit 0, 0, 0), Gosec filtré 0 finding / 0 ligne Go ajoutée, Gitleaks patch 0 finding exit 0. Archive ZIP : `forgelocal-t19r-evidence-7d71ba6.zip`, SHA-256 `1fd737ff44c2dc4ed2fc009229077c647dfcdb8bb6d8a9e2a45306b422030d4e`, manifeste non auto-référentiel 23/23.

## 8. Statuts bloquants maintenus (inchangés)

`PUBLIC_RELEASE_BLOCKED` · `SCAN_BLOCKED_UNKNOWN` · `NATIVE_SYSTEMVAULT_NOT_TESTED` · `camoflox_execution_authorized=false` · `t08_authorized=false` · `release_authorized=false`. Aucune release créée.
