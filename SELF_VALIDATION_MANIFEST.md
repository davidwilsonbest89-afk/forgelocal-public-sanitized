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
