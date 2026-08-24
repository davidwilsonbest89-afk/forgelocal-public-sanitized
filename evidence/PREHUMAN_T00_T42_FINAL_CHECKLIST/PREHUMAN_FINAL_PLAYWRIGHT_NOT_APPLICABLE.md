# ForgeLocal — preuve Playwright/T10 non applicable sous les gates

**Décision :** `NOT_APPLICABLE_UNDER_CURRENT_GATES` et `BLOCKED_BY_REQUIRED_PROTECTED_CONFIGURATION`.
**Ce n’est pas un oubli de test :** le harness a été invoqué, mais les tests se terminent à leur bootstrap de configuration avant tout test métier, parce que l’environnement autorisé ne fournit ni Core réel ni credentials protégés.

## Identité et commande

| Élément | Valeur |
|---|---|
| Date de contrôle | 2026-08-24 |
| UTC de début | `2026-08-24T18:58:51Z` |
| UTC de fin | `2026-08-24T18:58:53Z` |
| Clone | `/home/ubuntu/forgelocal-prehuman-fresh-20260824` |
| HEAD | `6ae02e4ceed239b9310fbf3fccb1b5170117251e` |
| CWD de la commande | `/home/ubuntu/forgelocal-prehuman-fresh-20260824/forge-dashboard` |
| Commande exacte | `pnpm exec playwright test --workers=1` |
| Code de sortie | `1` |
| Journal brut | `PREHUMAN_FINAL_EXIT_CHECKLIST_RAW.log`, bloc `check=dashboard_playwright` |

## Localisation des tests et préconditions

Les tests concernés se trouvent notamment dans `forge-dashboard/tests/proxies-t10.spec.ts`, `forge-dashboard/tests/t10-standalone.spec.ts`, `forge-dashboard/tests/automation-t15.spec.ts`, `forge-dashboard/tests/backups-t11.spec.ts`, `forge-dashboard/tests/bootstrap-ro.spec.ts`, `forge-dashboard/tests/environment-t13.spec.ts`, `forge-dashboard/tests/profile-writes-t09.spec.ts` et `forge-dashboard/tests/probes/verify-form-t16.spec.ts`.

Les préconditions exactes manquantes sont `FORGELOCAL_CORE_BASE_URL`, un token fourni uniquement par `FORGELOCAL_TOKEN_PATH` ou `BROWSEFORGE_TOKEN`, et `FORGELOCAL_BINARY` pour la sonde T16. Le harness a produit les erreurs suivantes : `CONFIGURATION_T10_ABSENTE:FORGELOCAL_CORE_BASE_URL`, `CORE_API_TOKEN_ABSENT:aucun token disponible (FORGELOCAL_TOKEN_PATH ou BROWSEFORGE_TOKEN)`, ainsi que les erreurs analogues T05, T09, T11, T13 et T16.

## Extrait de sortie brute

```text
Error: CORE_API_TOKEN_ABSENT:aucun token disponible (FORGELOCAL_TOKEN_PATH ou BROWSEFORGE_TOKEN)
    at automation-t15.spec.ts:86
Error: CONFIGURATION_T11_ABSENTE:FORGELOCAL_CORE_BASE_URL
    at backups-t11.spec.ts:20
Error: CONFIGURATION_T05_ABSENTE:FORGELOCAL_CORE_BASE_URL
    at bootstrap-ro.spec.ts:19
Error: CONFIGURATION_T13_ABSENTE:FORGELOCAL_CORE_BASE_URL
    at environment-t13.spec.ts:20
Error: CONFIGURATION_T16_ABSENTE:FORGELOCAL_BINARY
    at probes/verify-form-t16.spec.ts:24
Error: CONFIGURATION_T09_ABSENTE:FORGELOCAL_CORE_BASE_URL
    at profile-writes-t09.spec.ts:18
Error: CONFIGURATION_T10_ABSENTE:FORGELOCAL_CORE_BASE_URL
    at proxies-t10.spec.ts:20
Error: CONFIGURATION_T10_ABSENTE:FORGELOCAL_CORE_BASE_URL
    at t10-standalone.spec.ts:10
exit_code=1
check=dashboard_playwright
```

La sortie complète, incluant les commandes, les chemins, les timestamps et le code de sortie, est conservée dans le journal brut référencé ci-dessus. Aucun token n’a été créé, importé ou deviné. Aucun Core, navigateur réel, Camoufox, proxy réel ou cookie réel n’a été démarré.

## Gate concernée et condition de levée

Le contrôle reste bloqué par `PUBLIC_RELEASE_BLOCKED`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Une exécution ultérieure nécessiterait une autorisation explicite, une configuration synthétique approuvée et un Core de test autorisé ; cette condition ne peut pas être satisfaite dans la présente passe.

Les contrôles Dashboard indépendants `pnpm install --frozen-lockfile`, TypeScript, build et audit ont réussi. Ils ne transforment pas le contrôle Playwright dépendant du Core en réussite.

## Référence

[1]: ./PREHUMAN_FINAL_EXIT_CHECKLIST_RAW.log "Journal brut de la commande Playwright et des résultats"
