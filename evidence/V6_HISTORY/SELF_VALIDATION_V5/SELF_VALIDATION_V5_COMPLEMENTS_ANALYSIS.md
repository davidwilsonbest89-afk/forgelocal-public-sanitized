# Analyse des contrôles complémentaires — v5

Cette analyse conserve les résultats non nuls et les distingue des erreurs d’outillage. Aucun finding n’est masqué ni transformé en PASS.

| Contrôle | Résultat | Détail |
|---|---|---|
| Semgrep 1.174.0 | FINDINGS | 18 résultat(s), règles : `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read`=18 |
| Grype 0.117.0 CycloneDX | PASS technique / findings à classer | 2 correspondance(s), sévérités : High=2 |
| Grype 0.117.0 SPDX | PASS technique / findings à classer | 2 correspondance(s), sévérités : High=2 |
| Axe Playwright | BLOCKED_BY_FINDINGS | 2 violation(s), dont 1 sérieuse(s)/critique(s) : serious=1, moderate=1 |

## Semgrep

| # | Règle | Fichier | Ligne | Message | Risque | Propriétaire | Condition de levée |
|---:|---|---|---:|---|---|---|---|
| 1 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/api/correlation.go` | 25 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 2 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/api/readonly_session.go` | 88 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 3 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/api/router.go` | 516 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 4 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/api/router.go` | 728 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 5 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/api/router.go` | 814 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 6 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/backup/service.go` | 369 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 7 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/backup/service.go` | 462 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 8 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/launch/id.go` | 25 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 9 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/localvault/localvault.go` | 182 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 10 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/localvault/localvault.go` | 294 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 11 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/mcp/screenshot_artifacts.go` | 267 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 12 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/mcp/web_session_pool.go` | 557 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 13 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/profile/store.go` | 1050 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 14 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/profile/store.go` | 1085 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 15 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/profilemigration/migrator.go` | 1048 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 16 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/proxies/store.go` | 477 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 17 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/secrets/keyring.go` | 39 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |
| 18 | `home.ubuntu.forgelocal_self_validation_v5.go-math-rand-read` | `internal/templates/store.go` | 513 | Review math/rand use for security-sensitive values. | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |

## Grype

| SBOM | # correspondances | Sévérités | Packages principaux | Décision |
|---|---:|---|---|---|
| CycloneDX | 2 | {'High': 2} | golang.org/x/mod | À trier par CVE/package/version ; aucun upgrade automatique |
| SPDX | 2 | {'High': 2} | golang.org/x/mod | À trier par CVE/package/version ; aucun upgrade automatique |

## Axe

1. `color-contrast` — impact `serious`, Ensure the contrast between foreground and background colors meets WCAG 2 AA minimum contrast ratio thresholds. Cibles : `['.metric-primary > .metric-coordinates'], ['article:nth-child(2) > .metric-coordinates'], ['article:nth-child(3) > .metric-coordinates'], ['article:nth-child(4) > .metric-coordinates']`. Condition de levée : correction UI ciblée ou justification accessibilité revue humainement, puis rerun Axe.
2. `meta-viewport` — impact `moderate`, Ensure <meta name="viewport"> does not disable text scaling and zooming. Cibles : `['meta[name="viewport"]']`. Condition de levée : correction UI ciblée ou justification accessibilité revue humainement, puis rerun Axe.

## Décision

Les contrôles complémentaires ont accru la couverture mais ne lèvent pas les gates. Semgrep et Axe produisent des findings ; Grype termine techniquement mais ses correspondances doivent être triées ; la campagne reste en attente de revue indépendante.
