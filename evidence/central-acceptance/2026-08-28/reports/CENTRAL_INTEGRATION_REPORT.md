# Central local integration — LOT-A, LOT-B, LOT-C, LOT-D

**Scope.** This report consolidates the four physically received lots under `evidence/central-acceptance/2026-08-28/`. It does not modify product code, reopen a lot, close a gate, publish to GitHub, merge, release or deploy.

## Decision

The four lots are **centrally accepted as evidence packages**. Each ZIP has its matching external sidecar; all four sidecar checks, all four ZIP integrity checks, all four fresh extractions and all four internal manifest checks returned exit 0. This acceptance is not production approval.

```text
CENTRAL_EVIDENCE_ACCEPTANCE=COMPLETE
CENTRAL_LOCAL_INTEGRATION=AUTHORIZED_LOCAL_ONLY
LOT_A_CENTRAL_ACCEPTED=true
LOT_B_CENTRAL_ACCEPTED=true
LOT_C_CENTRAL_ACCEPTED=true
LOT_D_CENTRAL_ACCEPTED=true
PUSH_TO_GITHUB=NO
MERGE_PERFORMED=false
RELEASE_PERFORMED=false
PRODUCT_CODE_FILES_CHANGED=0
PUBLIC_RELEASE_BLOCKED=true
FORGELOCAL_PRODUCTION_READY=false
INDEPENDENT_REVIEW_PENDING=true
```

## Required measurable controls

| Control | Result |
|---|---|
| Physical receipt | 4 ZIP, 4 external sidecars, 8 files, 0 missing |
| Sidecar verification | LOT-A, B, C and D exit 0 |
| ZIP integrity | LOT-A, B, C and D `unzip -t` exit 0 |
| Fresh extraction | LOT-A, B, C and D exit 0 |
| Internal manifest | LOT-A, B, C and D exit 0; absolute paths 0; self-references 0 |
| Gate matrix | 4 lot rows, 16 gate rows, unassigned gates 0, hidden gates 0 |
| Product code changes | 0 |
| Out-of-scope changes | 0 |
| GitHub publication | No |

## Technical lot statuses preserved

A remains `LOT_A_PARTIAL_WITH_OPEN_FINDINGS`; its current and historical High/Critical findings and Semgrep findings remain open. B remains `LOT_B_PARTIAL_WITH_OBSERVED_BLOCKED_OR_UNSUPPORTED`; mixed outcomes are not converted into a global PASS. C remains `LOT_C_PARTIAL_WITH_ENVIRONMENT_BLOCKERS`; Docker/Buildx is not requalified. D remains `LOT_D_PARTIAL_WITH_ENVIRONMENT_BLOCKERS`; lineage and environment-dependent checks remain open or blocked.

## Gates

No release or production status is lifted. `PUBLIC_RELEASE_BLOCKED=true`, `FORGELOCAL_PRODUCTION_READY=false` and `INDEPENDENT_REVIEW_PENDING=true` remain authoritative. The token status is separate: no unproven rotation is asserted.

## Evidence map

The `proofs/` directory contains the receipt, package integrity matrix, manifest matrix, gate matrix, allowed-path inventory, regression command matrix, secret-scan evidence, token status and final status matrix. The `lot-packages/` directory contains only the four ZIPs and their four external sidecars. The `reports/` directory contains this report and the received central acceptance report.
