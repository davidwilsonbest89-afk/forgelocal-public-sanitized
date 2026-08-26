# GOSEC-REVIEW-R2 — Lot 1 matrix

This matrix covers every baseline finding in rules G703, G304, G305, G122 and G110. It does not suppress scanner output. `MITIGATED_CONTROL_SCANNER_OPEN` means the concrete path received controls and tests, while the static finding remains open where the scanner still reports the tainted sink.

| Disposition | Count |
|---|---:|
| `MITIGATED_CONTROL_SCANNER_OPEN` | 20 |
| `NEEDS_MANUAL_REVIEW` | 23 |

| Rule | Baseline count |
|---|---:|
| `G110` | 3 |
| `G122` | 1 |
| `G304` | 22 |
| `G305` | 3 |
| `G703` | 14 |

## Evidence boundary

The mitigated entries are supported by `GOSEC_REVIEW_R2_LOT1_TARGETED_TESTS_RAW.log`: safe ZIP/TAR extraction, traversal rejection, absolute/deep path rejection, symlink/hardlink rejection, entry-count limit, restore staging and preservation of the pre-existing destination. The remaining `NEEDS_MANUAL_REVIEW` entries are not closed by this lot and must remain in the next queue.
