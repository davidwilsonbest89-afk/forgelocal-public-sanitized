# GOSEC-R6 — registre consolidé

## Références

| Élément | Référence |
|---|---|
| Branche | `validation/gosec-r6` |
| HEAD vérifié | `f84738250fff6a91697629f899689e98da95da52` |
| JSON Gosec autoritatif | `R6_C/gosec_r6c_after_postcommit.json` |
| SHA-256 JSON | `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a` |
| Package R6-C | source `6bdda53`, evidence `b7a86b8`, package `d85ed7e` |
| Vérification package R6-C | clone public, checkout `d85ed7e`, `fsck` exit 0 |

## État consolidé

Le JSON autoritatif contient exactement 46 findings : G101=1, G115=3, G204=5, G302=5, G304=0, G305=1, G404=17, G703=7 et G704=7. Les 11 G304 et les deux G703 réduits au Lot A ne sont pas recréés. La matrice complète est `R6_OPEN_FINDINGS_CONSOLIDATED.tsv`.

| Domaine | Statut |
|---|---|
| Code production | Aucun changement dans la revue consolidée |
| Findings P0 démontrés | 0 |
| Findings P1 démontrés | 0 |
| Findings ouverts | 46, tous scanner-visibles |
| Environnements natifs | Partiels et non exécutés lorsqu’indisponibles |
| Release | Bloquée |

Les classifications détaillées et les priorités sont dans `R6_OPEN_FINDINGS_RISK_DECISION.md`. Cette décision est une revue de risque, pas une clôture Gosec ni une autorisation de production.

## Preuves

`R6_CONSOLIDATED_BASELINE_DISCOVERY_RAW.log` contient la vérification physique R6-C. `R6_CONSOLIDATED_GITLEAKS_EXTRACTION_CORRECTED_RAW.log` contient la commande corrigée avec `--no-git`; la tentative antérieure avec un répertoire non-Git est conservée dans le raw de baseline comme réserve de procédure. `R6_CONSOLIDATED_COMMANDS_RAW.log` et `R6_CONSOLIDATED_TESTS_RAW.log` contiennent les commandes, sorties et exits.

## Invariants

R5 reste à `28f66a1`, `validation/operational-v1` reste à `8004818`, `evidence/SMOKE_INTEGRATED_PROXY/` reste local et non suivi, T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 ne sont pas modifiés.

## Verdict

```text
GOSEC_R6_CONSOLIDATED_RISK_REVIEW_COMPLETE_PENDING_OWNER_DECISION
GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE
FORGELOCAL_PRODUCTION_READY=false
```
