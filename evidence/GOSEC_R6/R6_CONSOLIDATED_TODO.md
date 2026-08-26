# GOSEC-R6 — todo consolidé

## Décision avant code

La matrice consolidée couvre exactement les 46 findings du JSON R6-C post-commit. Aucun P0 ni P1 n’est démontré dans les tests d’atteignabilité autorisés. Ne pas démarrer un dernier lot de correction sans décision explicite du propriétaire et sans périmètre fermé.

## Revue manuelle restante

Les G703, G704, G302, G305 et G115 restent `MITIGATED_CONTROL_SCANNER_OPEN`. Les G404 de simulation/fingerprint et le G101 du nom de fichier de métadonnées restent `SCANNER_OPEN_MANUAL_REVIEW`. Les cinq G204 restent `BLOCKED_ENVIRONMENT_REQUIRED` pour les vérifications natives `open`, `xdg-open`, `rundll32` et `xattr`.

## Environnements

Camoufox, SystemVault natif, Docker/Buildx, Windows, macOS, GUI natifs, proxy/cookies réels et release restent non exécutés. Ils doivent conserver `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`; aucune simulation ne doit devenir PASS.

## Invariants

Ne pas recréer les packages R6-A/R6-B/R6-C. Ne réécrire aucun commit. Ne pas ajouter sur R5. Ne pas rouvrir T28. Ne pas démarrer T29. Ne pas modifier T31–T38. Ne jamais inclure `evidence/SMOKE_INTEGRATED_PROXY/`.

## Verdict

```text
GOSEC_R6_CONSOLIDATED_RISK_REVIEW_COMPLETE_PENDING_OWNER_DECISION
GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE
FORGELOCAL_PRODUCTION_READY=false
```
