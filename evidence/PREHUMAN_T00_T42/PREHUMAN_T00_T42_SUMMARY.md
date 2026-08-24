# ForgeLocal T00–T42 — résumé de prévalidation humaine

## Identité de la prévalidation

La prévalidation a été effectuée depuis un clone neuf de la branche `audit/t28-t42-evidence-correction`, au commit attendu `6ae02e4ceed239b9310fbf3fccb1b5170117251e`. Le tag baseline `t00-t27-complete-20260820` résout vers `69411e65c880d168832a65fc8475cc97d562a9ad`. La capacité initiale était de 28 Go libres ; après installation contrôlée des outils et la réhydratation LFS ciblée, 19 Go restaient libres. Le seuil d’arrêt de 10 Go n’a pas été franchi.

## Qualification globale

`git fsck --full`, `git lfs fsck`, les contrôles de sidecars portables, les tests `unzip -t` et extractions fraîches, `git diff --check t00-t27-complete-20260820..HEAD`, `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, Syft SBOM, la cohérence documentaire, la présence des gates et l’absence d’artefacts interdits ont été exécutés depuis le clone neuf. Les contrôles réussis ont retourné `exit_code=0`.

Gitleaks sur la plage complète conserve le signal historique `APi=REDACTED` et retourne `exit_code=1`; il reste donc classé `SCAN_BLOCKED_UNKNOWN`. Gosec retourne les constats historiques de baseline et de head avec leurs codes scanner réels ; la comparaison normalisée donne `194` contre `194`, `new_findings=[]` et `resolved_findings=[]`.

## Statuts par lot

| Lots | Statut de prévalidation | Justification |
|---|---|---|
| T00–T23 | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne d’artefacts | Deux morceaux LFS reconstruits temporairement ; SHA-256 et `unzip -t` conformes. Pas de prétention à des ZIP unitaires absents. |
| T24 | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne d’artefacts | ZIP LFS, hash et manifest global conformes. |
| T25 | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne d’artefacts | ZIP LFS, hash et manifest global conformes. |
| T26 | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne d’artefacts | ZIP LFS, hash et manifest global conformes. |
| T27 / CR01–CR05 | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne d’artefacts | Tarball, ZIP, bundles et sidecars vérifiés localement ; aucune autorisation produit déduite. |
| T28 | `BLOCKED` | Contrat documentaire, autorisation produit et implémentation extension absentes. |
| T29 | `BLOCKED` | Contrat documentaire ; import/export SystemVault non autorisé et non exécuté. |
| T30 | `PENDING_REMOTE_EVIDENCE_RECONCILIATION` | Commit et hashes identifiés, mais kit/branche distante canonique non rattaché à la clôture finale. |
| T31–T38 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | Code redacted réellement testé ; bundles et résultats conservés ; baseline logs historiques absents de la branche finale, reconstructions postérieures explicitement étiquetées. |
| T39 | `BLOCKED` | Import/export de secrets volontairement non implémenté faute d’autorisation T28/T29 et de qualification SystemVault native. |
| T40 | `BLOCKED` | Gate d’intégration documentaire ; aucune intégration runtime exécutée. |
| T41 | `BLOCKED` | Readiness de release bloquée ; aucune release exécutée. |
| T42 | `BLOCKED` | Clôture technique produite ; clôture produit et release interdites. |

## Gates et limites

Les gates suivantes restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. La prévalidation n’a exécuté aucun runtime réel, Camoufox, proxy réel, cookie réel, SystemVault natif, migration utilisateur, récupération de géolocalisation réelle ou release.

> La sortie `T00_T42_PREHUMAN_VALIDATION_READY_FOR_INDEPENDENT_REVIEW` signifie que la chaîne de preuve est préparée pour revue humaine. Elle ne constitue ni une approbation produit, ni une levée de gate, ni une autorisation d’exécution ou de publication.

## Artefacts

Les journaux détaillés sont `PREHUMAN_T00_T42_BASELINE_DISCOVERY_RAW.log`, `PREHUMAN_T00_T42_HISTORICAL_AUDIT_RAW.log`, `PREHUMAN_T00_T42_LFS_REHYDRATION_RAW.log`, `PREHUMAN_T00_T42_GLOBAL_QUALIFICATION_RAW.log` et `PREHUMAN_VALIDATION_RAW.log`. Le SBOM CycloneDX est `PREHUMAN_T00_T42_SBOM.cdx.json`; le rapport Gitleaks est `PREHUMAN_T00_T42_GITLEAKS.json`; la comparaison Gosec est `PREHUMAN_T00_T42_GOSEC_BASELINE_HEAD_DIFF.log`.
