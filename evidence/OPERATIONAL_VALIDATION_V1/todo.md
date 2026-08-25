# Todo — OPERATIONAL_VALIDATION_V1

| Action | Statut |
|---|---|
| Baseline clone/branche/HEAD/inventaire | DONE — `BASELINE_DISCOVERY_RAW.log` |
| V-CORE | DONE — FAIL non critique : token admin sans expiration/révocation |
| V-SQLITE-CRASH-RECOVERY | DONE — PASS sur temporaires |
| V-T28-RUNTIME-CONTRACT | DONE — PASS, T28 non rouvert |
| V-RUNTIME-SYNTHETIQUE | DONE — Chromium local PASS ; Camoufox/extension complète indisponible |
| V-PROXY-COOKIES-SYNTHETIQUES | DONE — PASS local redacted |
| V-SYSTEMVAULT | DONE — MemoryVault PASS ; natif indisponible |
| V-DOCKER | DONE — NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| V-DASHBOARD-API | DONE — FAIL non critique : Axe contraste et erreurs assets/analytics conservées |
| V-INSTALLATION-PROPRE | DONE — PASS après correctif permissions 0700 |
| V-SECURITY | DONE — réserves Gosec/OSV ; Gitleaks et pnpm production PASS |
| Publier le code/corrections et preuves brutes | DONE — `b1559ca53852c493ba15e4a06ad89b0c171c7938` |
| Générer ZIP/bundle/sidecars/manifeste | TODO |
| Vérifier extraction, checksums, bundle, fsck et clone neuf | TODO |
| Publier le rapport final avec hash des artefacts | TODO |
| T29/T39–T42, release et environnements natifs | INTERDITS / NON DÉMARRÉS |


## Clôture de preuve — 2026-08-25

| Action | Résultat | Statut |
|---|---|---|
| Package source cohérent | `b594f236a28483cc975dd054575d6cfb171d4f86` | DONE |
| ZIP OPV1 publié | `664fba54efe36a20269e60bba007082458a5d90c45ca10e9728cdf15fc7fedf5` | DONE |
| Bundle OPV1 publié | `67cce18e08f51540fceea71ba8e2dcb3c71344c24d10c4990e5e5e6460efb46f` | DONE |
| Commit de publication artefacts | `9cc6f5e45fc1aff9ba7a6ca06740cb6ac17538a2` | DONE |
| Vérification publique fraîche | hashes, sidecars, ZIP, manifeste, Gitleaks, bundle verify, seed, checkout source et `git fsck --full` exit 0 | DONE |
| Audit de vraies valeurs secrètes | PASS ; marqueurs synthétiques non détectés dans l’extraction | DONE |
| Verdict OPV1 | `FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE` | FINAL — non production-ready |
| T28 | `T28_APPROVED_VERIFIABLE_LOCAL`, non rouvert | INCHANGÉ |
| Défauts non critiques | V-CORE token expiration/revocation absente ; V-DASHBOARD contraste/assets ; V-SECURITY Gosec/OSV | DOCUMENTÉS |
| Camoufox/SystemVault natif/Docker et scanners absents | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | INCHANGÉ |
| T29/T39–T42 et release | aucun démarrage | INTERDITS / NON DÉMARRÉS |
