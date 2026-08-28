# Central acceptance — LOT-A, LOT-B, LOT-C, LOT-D

**Date UTC :** 2026-08-28T01:10:51Z
**Périmètre :** réception physique et vérification locale uniquement.
**Politique :** aucun push GitHub, merge, release, modification de code produit ou clôture de gate.

## Décision

La validation centrale est **acceptée pour la complétude physique et l’intégrité des preuves A–D**. Les quatre ZIP et les quatre sidecars externes sont présents, les hashes concordent, les ZIP sont valides, les extractions fraîches réussissent et les quatre manifestes internes sont vérifiés. Cette acceptation ne transforme pas les lots partiels en validation de production.

```text
LOTS_EXPECTED=4
LOTS_WITH_ZIP=4
LOTS_WITH_SIDECAR=4
DELIVERY_FILES=8
MISSING_FILES=0
CENTRAL_EVIDENCE_ACCEPTANCE=COMPLETE
LOT_A_CENTRAL_ACCEPTED=true
LOT_B_CENTRAL_ACCEPTED=true
LOT_C_CENTRAL_ACCEPTED=true
LOT_D_CENTRAL_ACCEPTED=true
CENTRAL_INTEGRATION_PREPARATION=AUTHORIZED_LOCAL_ONLY
PUBLIC_RELEASE_BLOCKED=true
FORGELOCAL_PRODUCTION_READY=false
```

## Résumé des contrôles exécutés

| Lot | ZIP | SHA-256 observé | ZIP test | Extraction fraîche | Manifeste interne | Acceptation centrale |
|---|---:|---|---:|---:|---:|---|
| LOT-A | présent (186099 octets) | `cd2aa9f285cad7de63494bc84c33fc770011f15b981793a69498fa9c5b557d50` | `0` | `0` | `0` | `CENTRAL_ACCEPTED` |
| LOT-B | présent (33775 octets) | `caf5212ba5000f1c508df865491ea1475050f50e0be99d32d4a07466ecc0def4` | `0` | `0` | `0` | `CENTRAL_ACCEPTED` |
| LOT-C | présent (52464 octets) | `de22feb6e25abeaf550a6c11c183c303628726b981d5e9a3038e873ed8a41b62` | `0` | `0` | `0` | `CENTRAL_ACCEPTED` |
| LOT-D | présent (32643 octets) | `d9fc3f90f4b8b6d98afccff98626b917a9ef0beaf9ba214859c041b0750f8ead` | `0` | `0` | `0` | `CENTRAL_ACCEPTED` |

Les ZIP eux-mêmes sont intrinsèquement valides, les quatre sidecars externes concordent, les extractions fraîches ont réussi et les manifestes internes ont aussi été vérifiés avec exit 0. Cela constitue une acceptation centrale de complétude et d’intégrité des preuves, sans constituer une validation de production.

## Verdicts déclarés par les lots

Les verdicts techniques déclarés dans les fichiers reçus restent séparés de la décision centrale : A est `LOT_A_PARTIAL_WITH_OPEN_FINDINGS`, B est `LOT_B_PARTIAL_WITH_OBSERVED_BLOCKED_OR_UNSUPPORTED`, C est `LOT_C_PARTIAL_WITH_ENVIRONMENT_BLOCKERS` et D est `LOT_D_PARTIAL_WITH_ENVIRONMENT_BLOCKERS`. Aucun statut global de production ou de release n’est accepté ; `PUBLIC_RELEASE_BLOCKED=true` et `FORGELOCAL_PRODUCTION_READY=false` sont maintenus.

## Prochaine action nécessaire

Les huit fichiers physiques requis ont été reçus. Les contrôles centraux sont reproductibles depuis les preuves jointes ; une éventuelle intégration reste locale et n’autorise aucun push GitHub, merge, release ou déploiement.

## Preuves

| Fichier | Rôle |
|---|---|
| `CENTRAL_RECEIPT_RAW.log` | Inventaire physique des 8 fichiers requis |
| `CENTRAL_DELIVERY_INVENTORY.tsv` | Matrice exacte des 4 ZIP et 4 sidecars |
| `CENTRAL_PACKAGE_INTEGRITY_RAW.log` | Codes sidecar, ZIP et extraction par lot |
| `CENTRAL_PACKAGE_INTEGRITY.tsv` | Résumé d’intégrité A–D |
| `CENTRAL_MANIFEST_RAW.log` | Vérification des manifestes internes |
| `CENTRAL_MANIFEST_MATRIX.tsv` | Matrice des manifestes internes |
| `CENTRAL_VERDICT_RECONCILIATION.tsv` | Décision centrale par lot |
