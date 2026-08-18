# Audit de gap T19-R — qualification intégrée Dashboard ↔ Core sur la lignée post-T18-R

**Baseline du lot :** commit `e6e4ebd96f12089dcc38d77226ddfacb8b6806c7` (tag `t18r-queue-recovery-verified-2026-08-18`), branche `forgelocal-baseline-2026-08-17`. Le source dashboard validé T16/T17 est présent dans la baseline en `forge-dashboard/` (109 fichiers, `pnpm-lock.yaml` et `playwright.config.ts` inclus).

## Contrats Core vérifiés côté production

| Contrat | Implémentation Core | Preuve test Core |
|---|---|---|
| Bootstrap loopback à usage unique | `internal/api/readonly_session.go:107` (`bootstrapReadOnlySession`) | `TestReadOnlyBootstrapIsLoopbackOnlySingleUseAndScopeLimited` |
| Expiration / rejeu / 401 | rejet 401 du rejeu et des tokens expirés | `readonly_session_test.go:72-120` |
| CORS originGuard fail-closed | `internal/api/router.go` (options `http(s)://127.0.0.1|localhost`) | tests `originGuard` T17-R |
| Pagination cursor | `internal/api/readonly.go:38` `json:"page"` | `readonly_test.go:47-57` (`NextCursor`) |
| Redaction sessions | `(*handler).listSessions` | `TestListSessionsRedactsTechnicalFields` (T17-R, handler réel) |
| CLI --no-runtime lecture seule | `cmd/server/main.go:181` | flag réservé diagnostic |

## Contrats dashboard vérifiés côté source

| Contrat | État dans `forge-dashboard/` | Preuve |
|---|---|---|
| Bootstrap loopback à usage unique | Présent : client `coreReadOnly` mémoire seule (closure Bearer), spec `bootstrap-ro.spec.ts` (rejeu → 401, expiration → 401, non-persistance) | `tests/bootstrap-ro.spec.ts` + `validate-core-readonly-bootstrap.mts` |
| Mémoire seule | Token en closure uniquement (commentaires contractuels `coreReadOnly.ts:3`, `coreWrite.ts:4`) | audit source |
| Absence localStorage/sessionStorage | `localStorage` présent uniquement dans `ThemeContext.tsx` (préférence de thème UI, hors contrat secret) ; clients `coreReadOnly.ts`/`coreWrite.ts` n'en contiennent aucun | audit source `grep -l` |
| Redaction des projections | Specs existantes : T11 backups (SHA-256 seulement), T15 (digest + longueur), T09/T10/T13/T14 | specs existantes |
| Pagination conforme Core | `listProfiles/listGroups/listRuntimes` avec `limit`/`cursor` (`coreReadOnly.ts:61-63, 104-107`) | audit source |
| Refus hors loopback | Dépend du Core (originGuard fail-closed T17-R) ; tests `automation-t15.spec.ts` (Origin hostil → ORIGIN_REJECTED) | specs existantes |
| Absence de mutation involontaire | client `coreWrite` mémoire seule, X-Request-ID, invalidation 401/403 | specs existantes |

## Verdict et plan d'adaptation

Aucun contrat n'est ABSENT. Le point unique PRESENT_UNTESTED est l'exécution **intégrée** du dashboard contre le Core réel `--no-runtime` (les specs existantes supposent un Core démarré, mais aucune exécution E2E complète contre la lignée post-G6/T17-R/T18-R n'a été produite depuis ce commit). Le lot consiste donc en la requalification intégrée : installation verrouillée (lock existant), `tsc`, build production, lancement du Core réel `--no-runtime` sur 127.0.0.1, suite Playwright séquentielle, puis scans immuables. Aucun code produit Core ou dashboard n'a besoin d'être modifié a priori ; si une spec échoue pour cause de contrat réellement changé, la correction minimale sera consignée.
