# GOSEC-R6 — décision consolidée de risque

## Décision

La revue consolidée couvre exactement 46 findings actuels au JSON SHA-256 `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a`. Aucun **P0 démontré** et aucun **P1 démontré** à ce stade : les contrôles applicatifs, tests négatifs et chemins réellement exécutés ne démontrent pas une exploitation critique immédiate dans l’environnement synthétique autorisé. Cette absence de P0/P1 n’est pas une clôture de sécurité.

| Décision | Findings | Règles principales | Sens |
|---|---:|---|---|
| CORRECTION_REQUISE_P0 | 0 | — | Aucune atteignabilité critique démontrée |
| CORRECTION_REQUISE_P1 | 0 | — | Aucun défaut urgent démontré après contrôles disponibles |
| MITIGATED_CONTROL_SCANNER_OPEN | 23 | G115=3, G302=5, G305=1, G703=7, G704=7 | Contrôle présent mais finding encore signalé |
| SCANNER_OPEN_MANUAL_REVIEW | 18 | G101=1, G404=17 | Finding contextuel non cryptographique/synthétique, non clos |
| HISTORICAL_NOT_REACHABLE | 0 | — | Aucun finding courant classé historique |
| BLOCKED_ENVIRONMENT_REQUIRED | 5 | G204=5 | `open`/`xdg-open`/`rundll32`/`xattr` exigeant validation native |

Ce décompte est non chevauchant et totalise 46; la matrice TSV reste l’autorité ligne par ligne.

## P0/P1

Aucun P0/P1 ne doit être corrigé automatiquement dans cette revue. Les G703/G704/G302/G115 restent P2 avec contrôles documentés; les G404/G101 restent P3 en revue manuelle; les G204 natifs restent P2 mais bloqués par environnement. Un dernier lot ne pourrait être proposé qu’après décision du propriétaire et périmètre fermé.

## Environnements bloqués

Les actions natives Windows/macOS, GUI, Camoufox, SystemVault natif et Docker/Buildx n’ont pas été exécutées. Elles ne sont pas simulées et restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`; la matrice utilise `BLOCKED_ENVIRONMENT_REQUIRED` seulement pour les cinq G204 dont la vérification native est nécessaire.

## Verdict

```text
GOSEC_R6_CONSOLIDATED_RISK_REVIEW_COMPLETE_PENDING_OWNER_DECISION
GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE
FORGELOCAL_PRODUCTION_READY=false
```
