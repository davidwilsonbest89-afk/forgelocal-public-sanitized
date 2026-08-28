# Central gate summary — A–D

The four lots are accepted as physically received and integrity-verified evidence. Their technical statuses remain partial and their open gates are not closed.

| Lot | Technical status | Central treatment |
|---|---|---|
| A | `LOT_A_PARTIAL_WITH_OPEN_FINDINGS` | Findings remain open; no Semgrep or Grype closure is inferred. |
| B | `LOT_B_PARTIAL_WITH_OBSERVED_BLOCKED_OR_UNSUPPORTED` | Mixed outcomes remain classified; no global functional PASS is asserted. |
| C | `LOT_C_PARTIAL_WITH_ENVIRONMENT_BLOCKERS` | Docker/Buildx remains environment-blocked. |
| D | `LOT_D_PARTIAL_WITH_ENVIRONMENT_BLOCKERS` | Lineage and environment-dependent checks remain open or blocked. |

```text
LOT_ROWS=4
GATE_ROWS=16
UNASSIGNED_GATES=0
HIDDEN_GATES=0
RELEASE_GATES_LIFTED=0
PRODUCTION_READY_ASSERTIONS=0
PUBLIC_RELEASE_BLOCKED=true
FORGELOCAL_PRODUCTION_READY=false
INDEPENDENT_REVIEW_PENDING=true
```
