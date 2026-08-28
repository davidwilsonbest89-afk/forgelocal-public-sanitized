# Passation développeur — Audit exhaustif du Dashboard

**Date :** 2026-08-28 UTC
**Branche :** `validation/final-secret-remediation`
**Baseline de départ :** `4fd54022f3970d2177976a78970d2b81bd8810bf`
**Statut de publication :** commit et push autorisés par l’owner ; références finales à compléter après publication.

## Résumé

Un audit Playwright réel a été exécuté contre le Dashboard et un Core réel sur `127.0.0.1`, avec Chromium local et des fixtures synthétiques uniquement. La campagne finale couvre **55 assertions UI et 1 gate HTTP**, soit **56 assertions PASS**. Elle couvre la navigation, les espaces de travail, les réglages, les filtres, les notifications, les connexions read-only/admin, les actions de profils, tags, archivage, duplication, export, proxy, backup, environnement, runtime, automation, extensions, déconnexion, accessibilité et responsive.

Le résultat signifie qu’aucune anomalie inattendue n’a été observée dans l’environnement qualifié. Il ne signifie pas que les capacités natives absentes de la sandbox sont certifiées en production.

| Mesure | Résultat | Preuve |
|---|---:|---|
| Assertions UI | `55/55 PASS` | `evidence/runtime-qualification/2026-08-28/dashboard-full/DASHBOARD_BUTTON_COVERAGE_MATRIX.tsv` |
| Gate HTTP inattendu | `0` | `UNEXPECTED_HTTP_ERROR_RESPONSES=0` |
| Erreurs JavaScript de page | `0` | `PAGE_FAILURE_COUNT=0` |
| Accessibilité des noms de boutons | `PASS` | `A11Y_BUTTON_NAMES=PASS` |
| Responsive 412×915 | `PASS` | `RESPONSIVE_NO_HORIZONTAL_OVERFLOW=PASS` |
| Check TypeScript | Exit `0` | `DASHBOARD_CHECK_AFTER_RESPONSIVE_FIX.log` |
| Build Dashboard | Exit `0` | `DASHBOARD_BUILD_FINAL.log` |
| Nettoyage | `PASS` | `DASHBOARD_FINAL_CLEANUP.log` |

## Corrections incluses

Dans `ProxyRegistry.tsx` et `Home.tsx`, les callbacks d’affectation et de retrait d’affectation mettent maintenant à jour l’état parent. Le bouton opposé devient donc immédiatement cohérent avec l’état confirmé par le Core. La campagne confirme `PROXY_ASSIGN_UI=PASS`, `PROXY_UNASSIGN_UI=PASS` et `PROXY_DELETE_UI=PASS`.

Dans `index.css`, la largeur minimale fixe de 1120 px a été supprimée pour les petites largeurs et remplacée par une mise en page mobile contrôlée. La campagne 412×915 confirme l’absence d’overflow horizontal.

Dans `internal/api/backup_v1.go`, `internal/api/backup_v1_test.go` et `internal/api/router.go`, le contrat backup attendu par le Dashboard a été câblé avec des projections redacted et un test contractuel. Les credentials et chemins absolus ne sont jamais exposés au Dashboard.

## Résultats à interpréter avec prudence

Le bouton de sauvegarde est correctement fail-closed, mais la création d’archive réelle reste bloquée par l’absence du `SystemVault` natif dans la sandbox ; le raw conserve `BACKUP_CREATE_BLOCKED_SYSTEMVAULT=PASS` comme état environnemental attendu. La consultation d’identité environnementale rend un état vide lorsque aucun navigateur qualifié n’a produit de diagnostic ; le HTTP 404 correspondant est attendu et non classé comme anomalie inattendue.

Les éléments suivants restent hors qualification : Docker/Buildx, Firefox/Camoufox, SystemVault natif, cookies réels, proxys commerciaux, sites externes et environnements Windows/macOS/GUI native. Aucun secret réel, compte réel, cookie réel ou site externe n’a été utilisé.

## Comment vérifier localement

Depuis `repo/forge-dashboard`, exécuter `pnpm run check` puis `pnpm run build`. Pour le parcours UI, lancer un Core loopback avec une base SQLite temporaire, désactiver tout téléchargement runtime externe, démarrer le Dashboard sur loopback, créer uniquement les deux fixtures synthétiques, fournir un code read-only à usage unique, puis exécuter `dashboard_full_button_audit.mjs`. Le harness doit se terminer par `STATUS=PASS`, `BUTTON_AUDIT_CASES=56`, `PAGE_FAILURE_COUNT=0` et `UNEXPECTED_HTTP_ERROR_RESPONSES=0`.

Ne jamais réutiliser un code read-only consommé, ne jamais utiliser un token réel et ne jamais pointer le navigateur vers une URL externe. Supprimer la base temporaire, le code à usage unique, les PID et les ports après chaque campagne.

## Fichiers produit modifiés

```text
forge-dashboard/client/src/components/ProxyRegistry.tsx
forge-dashboard/client/src/index.css
forge-dashboard/client/src/pages/Home.tsx
internal/api/backup_v1.go
internal/api/backup_v1_test.go
internal/api/router.go
```

Le diff de ces six fichiers doit être relu avant merge. Les preuves sont sous `evidence/runtime-qualification/2026-08-28/dashboard-full/`. Le package local et son SHA-256 sont fournis dans ce dossier ; le manifeste et le log de vérification doivent rester cohérents après toute modification.

## Procédure de revue et gates

Le reviewer doit vérifier le diff, le contrat des routes backup, les tests contractuels, la matrice 55/55, le raw complet, les screenshots, le build, le check, le nettoyage et l’absence de secrets. Aucun changement ne doit transformer une limite environnementale en PASS de capacité native.

Les gates de livraison restent obligatoires :

```text
DASHBOARD_BUTTON_AUDIT=PASS_WITH_EXPLICIT_ENVIRONMENT_LIMITS
PUSH_TO_GITHUB=YES — uniquement pour le commit identifié dans cette passation
MERGE_PERFORMED=false
RELEASE_PERFORMED=false
PUBLIC_RELEASE_BLOCKED=true
FORGELOCAL_PRODUCTION_READY=false
INDEPENDENT_REVIEW_PENDING=true
```

Après revue indépendante, l’équipe peut décider séparément si elle autorise un merge. Le push de cette passation et des corrections ne vaut ni merge, ni release, ni déclaration de readiness production.
