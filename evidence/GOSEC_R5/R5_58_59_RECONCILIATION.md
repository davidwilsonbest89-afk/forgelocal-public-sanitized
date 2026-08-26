# R5 — note de réconciliation 58/59

## Conclusion autoritative

La réconciliation a été exécutée sur `evidence/GOSEC_R5/gosec_r5_final.json`, sans modifier le JSON ni les rapports R5 existants. Le fichier contient **59 entrées `Issues`**, et son SHA-256 est `f24905f47bee87164511232caf0538d551e2199460678775390f789b7a2c3e9f`. Le hash du fichier du worktree correspond au blob Git du HEAD `d01cea88117a0f6694c5f26d606ccc5694fa6c3a`; le JSON suivi par Git est donc exactement celui vérifié.

La distribution complète est :

| Règle | Findings |
|---|---:|
| G101 | 1 |
| G115 | 3 |
| G204 | 5 |
| G302 | 5 |
| G304 | 11 |
| G305 | 1 |
| G404 | 17 |
| G703 | 9 |
| G704 | 7 |
| **Total JSON** | **59** |

L’écart ne provient pas d’une 59e entrée sans `rule_id`. Il provient de l’omission de **G305=1** dans la liste textuelle fournie pour le mandat R6. Les rapports R5 finaux et le JSON contiennent bien G305; la liste complète est donc **59**, et la liste R6 initiale était **58** parce qu’elle oubliait G305.

## Intégrité et décision de blocage

Le JSON R5 original est conservé. La note et le raw sont des preuves complémentaires; aucune réécriture silencieuse du rapport R5, du package R5 ou de ses archives n’est effectuée. Le blocage initial `GOSEC_R6_BLOCKED_R5_FINDING_COUNT_RECONCILIATION_REQUIRED` est levé uniquement pour la cohérence arithmétique : la baseline R6 doit inclure explicitement G305=1.

La future branche R6 devra partir du HEAD distant réellement vérifié de `validation/gosec-r5`, sans modifier les commits R5. Ses statuts restent `GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
