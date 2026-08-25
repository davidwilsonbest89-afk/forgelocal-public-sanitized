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


## T00–T42 V6 — remédiation des findings réels

Cette section est append-only et ne modifie aucun artefact V3/V4/V5. Le dossier `evidence/SELF_VALIDATION_V6/` contient la découverte brute, les matrices individuelles Gitleaks/Semgrep/Grype, les diffs, les tests ciblés, les logs de requalification, les scans OSV/Trivy/Govulncheck/GolangCI/Semgrep, les SBOM CycloneDX/SPDX propres, l’inventaire de licences, les JSON Axe avant/après et les preuves de cleanup.

| Livrable V6 | Vérification |
|---|---|
| `SELF_VALIDATION_V6_REPORT.md` | Rapport final et statut exact V6 |
| `V6_GITLEAKS_TRIAGE.csv` | 348 findings individuels redacted |
| `V6_SEMGREP_TRIAGE.md` | 18 usages `crypto/rand` analysés individuellement |
| `V6_GRYPE_TRIAGE.md` | Deux advisories x/mod avec version corrigée et chaîne transitive |
| `evidence/SELF_VALIDATION_V6/` | Preuves brutes V6, JSON, SBOM, scans, tests et cleanup |
| `evidence/V6_HISTORY/` | Copies byte-identiques des wrappers V3/V4/V5 et preuves V5 |
| `evidence/forgelocal-t00-t42-self-validation-v6-findings-remediation.zip` | Wrapper V6 append-only |
| `evidence/forgelocal-t00-t42-self-validation-v6-findings-remediation.zip.portable.sha256` | Sidecar portable du wrapper V6 |

Le wrapper V6 contient 297 entrées et son extraction passe `unzip -t`. Les hashes des copies historiques incluses correspondent respectivement à V3 `5b31ae7afcdc032bec46785a2573ceab9ec797261c3e0d31490f3c6fb9dfbe2b`, V4 `f6544091783c2a4d4694d4b5f02c5dd5f0c70d22dab5efbb90abfd81418019bc` et V5 `40858a40752f60e567be4dba965ac2afd987cb8054878335a1ccb8afac5b9bd3`.

**Statut V6 :** `T00_T42_V6_FINDINGS_REMEDIATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`.


## Métadonnées finales de livraison V6

- **Wrapper V6 :** `evidence/forgelocal-t00-t42-self-validation-v6-findings-remediation.zip`
- **SHA-256 wrapper V6 :** `ce722915d70e0aa528927b753c6f18efa5706fc9fa8703ef6f449b6728a5fab6`
- **Sidecar wrapper :** `evidence/forgelocal-t00-t42-self-validation-v6-findings-remediation.zip.portable.sha256`
- **Bundle delta V6 :** `evidence/forgelocal-t00-t42-self-validation-v6.delta.bundle`
- **SHA-256 bundle delta V6 :** `ad4484e795b80eb5b7655228012e695dc4b260d43057477a97ae145d164614c2`
- **Prérequis bundle :** `b34fa5c02ff20144abfb5d240db1c67ad1f038f9`
- **Cible bundle contenu :** `fc080456711dd7f2266911aaec55041fdb1b424c`
- **Vérification wrapper finale :** `V6_WRAPPER_VERIFY.log`, externe au wrapper pour éviter l’auto-référence ; `unzip -t`, sidecar, extraction, checksums internes, bundle, Gitleaks d’extraction et hashes V3/V4/V5 retournent tous code 0.

Le HEAD de branche qui contiendra le paquet final sera postérieur à la cible du bundle. Cette distinction est volontaire : le bundle couvre le contenu source corrigé et ne s’auto-référence pas dans les commits de packaging.
