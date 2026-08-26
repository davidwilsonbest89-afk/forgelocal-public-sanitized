# GOSEC-REVIEW-R2 — Lot 2 matrix

The matrix covers all 13 baseline findings in G204/G704 after Lot 1. Every finding has a disposition; no scanner result is suppressed. `MITIGATED_CONTROL_SCANNER_OPEN` means the execution path has a specific control and regression evidence, while Gosec still reports the static sink. GUI launcher findings remain `NEEDS_MANUAL_REVIEW` because this Linux run cannot prove platform-specific lifecycle/orphan behavior.

| Disposition | Count |
|---|---:|
| MITIGATED_CONTROL_SCANNER_OPEN | 10 |
| NEEDS_MANUAL_REVIEW | 3 |
| **Total** | **13** |

| Rule | Baseline | Head |
|---|---:|---:|
| G204 | 6 | 5 |
| G704 | 7 | 7 |

The detailed one-row-per-finding record is `GOSEC_REVIEW_R2_LOT2_MATRIX.tsv`. IPv6 loopback validation is covered by a unit test; no external host, commercial proxy, real cookie or real credential is used. Darwin-only xattr execution is not run in this Linux environment and is not treated as a PASS.
