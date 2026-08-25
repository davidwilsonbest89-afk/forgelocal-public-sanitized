# SELF_VALIDATION_MANIFEST — v4

Ce manifeste est **non auto-référentiel** : ni ce fichier ni `SHA256SUMS` ne figurent dans `SHA256SUMS`. Les fichiers listés couvrent les livrables principaux de l’auto-vérification.

| Livrable | Vérification |
|---|---|
| `SELF_VALIDATION_SUMMARY.md` | Résumé officiel et statut exact |
| `SELF_VALIDATION_FINDINGS_CLASSIFICATION.md` | Classification individuelle Staticcheck/Gosec |
| `SELF_VALIDATION_SYNTHETIC_E2E_RAW.log` | Commandes, résultats et code Playwright |
| `SELF_VALIDATION_CLEANUP_VERIFICATION.log` | Processus, token, SQLite et répertoires temporaires après run |
| `SELF_VALIDATION_BUNDLE_DELTA_VERIFY.log` | `git bundle verify` et baseline requise |
| `evidence/forgelocal-t00-t42-self-validation-v4-synthetic-e2e.zip` | Wrapper V4 append-only |
| `evidence/forgelocal-t00-t42-self-validation-v4.delta.bundle` | Bundle delta baseline → commit v4 |
| `evidence/forgelocal-t00-t42-self-validation-v4-synthetic-e2e.zip.portable.sha256` | Sidecar portable du wrapper V4 |
| `evidence/forgelocal-t00-t42-self-validation-v4.delta.bundle.portable.sha256` | Sidecar portable du bundle delta |
| `evidence/SELF_VALIDATION_V4/` | Preuves brutes, scans, SBOM CycloneDX/SPDX, licences et manifest interne |
| `evidence/forgelocal-t00-t42-prehuman-final-review-wrapper-v3-code-fixed.zip` | Wrapper V3 historique conservé inchangé |

Les hashes des livrables principaux sont dans `SHA256SUMS`. Les quatre artefacts LFS critiques ont été récupérés individuellement ; aucun token temporaire, profil, cookie, SQLite ou résultat Playwright n’est livré.

## Statut exact

`T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW`

## 2026-08-25 — Manifest V5 enhanced append-only

**Branche :** `audit/t00-t42-self-validation-enhanced-v5`
**HEAD de contenu :** `56126208fed288fff216de8945579ee93c385cc3`
**Prérequis bundle :** `b34fa5c02ff20144abfb5d240db1c67ad1f038f9`
**Statut :** `T00_T42_SELF_VALIDATION_ENHANCED_PENDING_INDEPENDENT_REVIEW`

| Livrable | Vérification |
|---|---|
| `SELF_VALIDATION_V5_REPORT.md` | Rapport renforcé et décision sans levée de gate |
| `SELF_VALIDATION_V5_MANDATORY_ANALYSIS.md` | Analyse Gitleaks explicite, GolangCI-Lint et Go shuffle |
| `SELF_VALIDATION_V5_COMPLEMENTS_ANALYSIS.md` | Classification Semgrep, Grype CycloneDX/SPDX et Axe |
| `evidence/SELF_VALIDATION_V5/` | Logs, JSON, inventaires, scripts et wrapper V3 historique |
| `evidence/forgelocal-t00-t42-self-validation-v5-enhanced.zip` | `unzip -t`, extraction fraîche 175 fichiers et re-scan Gitleaks sans leak |
| `evidence/forgelocal-t00-t42-self-validation-v5.delta.bundle` | `git bundle verify`, delta depuis b34fa5c |
| `evidence/forgelocal-t00-t42-self-validation-v5-enhanced.zip.portable.sha256` | Sidecar portable wrapper |
| `evidence/forgelocal-t00-t42-self-validation-v5.delta.bundle.portable.sha256` | Sidecar portable bundle |
| `SHA256SUMS` | Sommes principales, hors manifeste et fichier de sommes eux-mêmes |

Le Gitleaks `--log-opts` est conservé comme incomplet car il annonce zéro commit ; 58 arbres ont ensuite été scannés explicitement et produisent 348 détections redacted. GolangCI-Lint 2.13.1 produit 2 nouveaux findings sur 90 ; Go shuffle/race termine avec code 0 ; Semgrep produit 18 findings ; Grype retourne 2 correspondances High par SBOM ; Axe bloque volontairement sur 1 violation sérieuse de contraste. Les gates restent fermées.
