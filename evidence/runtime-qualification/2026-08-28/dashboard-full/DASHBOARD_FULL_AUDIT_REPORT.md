# Audit exhaustif du Dashboard ForgeLocal

**Date d’exécution :** 2026-08-28 04:26–04:32 UTC
**Périmètre :** Dashboard local, Core réel compilé depuis la branche `validation/final-secret-remediation`, Chromium système local, données synthétiques uniquement, adresses `127.0.0.1` uniquement.

## Verdict exécutif

La campagne finale couvre **55 assertions UI** et **1 gate HTTP**, soit **56 lignes de couverture**. Les 56 assertions sont PASS et le harness final se termine avec `STATUS=PASS`, `PAGE_FAILURE_COUNT=0`, `UNEXPECTED_HTTP_ERROR_RESPONSES=0` et `TOKEN_MATCH_COUNT=0`.

Le résultat ne constitue pas une certification de production. Deux réponses HTTP restent visibles dans la trace, mais elles sont attendues et classées explicitement : HTTP 500 lors de la sauvegarde, car le `SystemVault` natif n’est pas disponible dans la sandbox, et HTTP 404 lors de la consultation d’un diagnostic environnemental absent pour un profil qui n’a pas lancé de navigateur qualifié. Le Dashboard rend ces états sous forme d’état guarded/empty state.

| Indicateur | Résultat | Preuve |
|---|---:|---|
| Assertions UI | `55/55 PASS` | `DASHBOARD_BUTTON_COVERAGE_MATRIX.tsv` |
| Gate HTTP inattendu | `0` | `UNEXPECTED_HTTP_ERROR_RESPONSES=0` |
| Cas d’environnement explicitement limités | `6` | Matrice de couverture |
| Erreurs JavaScript page | `0` | `PAGE_FAILURE_COUNT=0` |
| Overflow horizontal à 412 px | `0` | `RESPONSIVE_NO_HORIZONTAL_OVERFLOW=PASS` |
| Build TypeScript | Exit `0` | `DASHBOARD_BUILD_FINAL.log` |
| Check TypeScript | Exit `0` | `DASHBOARD_CHECK_AFTER_RESPONSIVE_FIX.log` |
| Processus/données temporaires résiduels | `0` | `DASHBOARD_FINAL_CLEANUP.log` |

## Surface couverte

Le parcours Playwright réel couvre le chargement initial, la navigation vers les espaces locaux, le journal d’audit, les réglages, l’aide et les notifications. Il vérifie l’ouverture et la modification des filtres avancés, leur remise à zéro, la recherche, le filtre de groupe, les filtres d’état, la copie d’identifiant et les contrôles d’accessibilité des boutons.

La liaison Core couvre la connexion read-only par code à usage unique, le chargement des groupes et runtimes, le rejet guarded d’un jeton administrateur trop court, la connexion admin valide, la déconnexion admin et la déconnexion read-only. Le parcours de profils couvre la création via Core, la visibilité du profil créé, l’ajout et la suppression d’un tag, l’archivage, la réouverture, la duplication et l’export.

La surface proxy couvre la création, l’affectation, le retrait d’affectation et la suppression. La surface backup couvre le bouton de sauvegarde, l’ouverture du panneau, l’actualisation, ainsi que les états détail/restauration/purge lorsque le catalogue est vide à cause de l’indisponibilité du SystemVault. Les panneaux Identité navigateur, Runtime qualifié, Automation locale et Extensions locales sont ouverts et leurs contrôles de consultation/actualisation sont exercés. Les scénarios Automation restent fail-closed pour une URL externe synthétique non routable.

## Corrections appliquées

Le premier défaut confirmé concernait `ProxyRegistry` : après une affectation ou un retrait d’affectation, l’état parent ne reflétait pas immédiatement la mutation Core, ce qui laissait le bouton opposé non cliquable. Des callbacks explicites ont été ajoutés et reliés à l’état `assignedProxyIds` de `Home.tsx`. Le rerun confirme `PROXY_ASSIGN_UI=PASS`, `PROXY_UNASSIGN_UI=PASS` et `PROXY_DELETE_UI=PASS`.

Le second défaut confirmé concernait la responsive design. `body { min-width: 1120px; }` imposait un overflow à une largeur de 412 px. Une media query mobile a été ajoutée : suppression de la largeur minimale globale, empilement du shell, navigation horizontale locale, grille métrique en deux colonnes, tableaux avec overflow interne contrôlé, panneaux mono-colonne et masquage du rail d’observation sur petit écran. Le contrôle final confirme `RESPONSIVE_NO_HORIZONTAL_OVERFLOW=PASS`.

Les routes backup du Core précédemment ajoutées restent montées et redacted. L’audit final confirme que l’interface atteint la route de création et reçoit le refus attendu `BACKUP_KEY_FAILED` lorsque le coffre natif n’est pas disponible, sans fuite de credential.

## Preuves et reproductibilité

Le harness final est `dashboard_full_button_audit.mjs`. Le raw complet est `DASHBOARD_FULL_BUTTON_AUDIT_RAW.log`, le résultat agrégé est `DASHBOARD_FULL_BUTTON_AUDIT_RESULT.log`, et la matrice lisible est `DASHBOARD_BUTTON_COVERAGE_MATRIX.tsv`. Les screenshots `AUDIT_01_INITIAL.png`, `AUDIT_02_CORE_CONNECTED.png`, `AUDIT_03_FINAL.png`, `AUDIT_04_RESPONSIVE.png` et `AUDIT_FAILURE.png` sont conservés à titre de traçabilité ; `AUDIT_FAILURE.png` correspond à une itération diagnostique antérieure et ne doit pas être interprété comme le verdict final.

Les validations statiques finales ont été exécutées avec `pnpm run check` et `pnpm run build`, chacun avec exit code `0`. Le dépôt reste sur `4fd54022f3970d2177976a78970d2b81bd8810bf`, le HEAD distant est identique, et le diff runtime reste non commit. Les six fichiers locaux modifiés sont `forge-dashboard/client/src/components/ProxyRegistry.tsx`, `forge-dashboard/client/src/index.css`, `forge-dashboard/client/src/pages/Home.tsx`, `internal/api/backup_v1.go`, `internal/api/backup_v1_test.go` et `internal/api/router.go`.

## Limites obligatoires

La campagne ne qualifie pas un coffre `SystemVault` natif, Firefox/Camoufox, Docker/Buildx, Windows/macOS/GUI native, des cookies réels, un proxy commercial ou un site externe. Elle vérifie uniquement des fixtures locales et synthétiques. Le bouton de sauvegarde est donc PASS comme comportement UI guarded, mais la création réelle d’une archive reste `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` dans cette sandbox.

La consultation d’identité environnementale est PASS comme parcours UI et état vide attendu ; elle ne prouve pas qu’un navigateur réel a produit un diagnostic. De même, les panneaux Runtime, Automation et Extensions confirment le contrat et les protections UI, mais ne certifient pas les capacités natives absentes de l’environnement.

Les gates de livraison ne changent pas :

```text
DASHBOARD_BUTTON_AUDIT=PASS_WITH_EXPLICIT_ENVIRONMENT_LIMITS
DASHBOARD_ASSERTIONS=55/55
DASHBOARD_HTTP_GATE=PASS
DASHBOARD_REAL_CORE=PASS_WITH_EXPLICIT_ENVIRONMENT_LIMITS
PUSH_TO_GITHUB=NO
PUBLIC_RELEASE_BLOCKED=true
FORGELOCAL_PRODUCTION_READY=false
INDEPENDENT_REVIEW_PENDING=true
```

Aucun secret réel, compte, cookie réel, proxy réel ou site externe n’a été utilisé. Les codes et tokens synthétiques éphémères ont été supprimés après exécution ; aucune valeur de credential n’est conservée dans les logs finaux.
