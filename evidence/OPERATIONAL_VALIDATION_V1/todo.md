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
