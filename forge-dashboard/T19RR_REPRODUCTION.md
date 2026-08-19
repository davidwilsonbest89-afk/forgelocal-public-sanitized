# T19-RR — Reproduction indépendante Dashboard ↔ Core

Ce guide permet de rejouer les tests intégrés du dashboard contre un Core local en utilisant uniquement un clone du bundle canonique et une base temporaire. Il ne requiert ni Camoufox, ni proxy réseau réel, ni token préexistant.

## Préconditions

Le Core est compilé depuis `cmd/server` et Chromium local est qualifié. Les variables suivantes doivent désigner les processus réellement démarrés. Elles ne contiennent aucun secret ; le jeton est lu depuis un fichier temporaire fourni par le Core.

```sh
export FORGELOCAL_BINARY="$PWD/../forge-core"
export FORGELOCAL_BASE_DIR="/tmp/forgelocal-t19rr-data"
export FORGELOCAL_CORE_BASE_URL="http://127.0.0.1:19280"
export FORGELOCAL_DASHBOARD_URL="http://127.0.0.1:3100"
export FORGELOCAL_TOKEN_PATH="/tmp/forgelocal-t19rr-token.txt"
```

Le dashboard doit être lancé explicitement sur la même URL que `FORGELOCAL_DASHBOARD_URL`, par exemple :

```sh
pnpm exec vite --host 127.0.0.1 --port 3100
```

## Fixture T15 autosuffisante

`tests/automation-t15.spec.ts` vérifie d’abord, avec un délai maximal de cinq secondes, le healthcheck Core et la racine du Dashboard configuré. Les échecs sont explicites : `T15_CORE_UNAVAILABLE` ou `T15_DASHBOARD_UNAVAILABLE`. Ensuite seulement, `beforeAll` crée la fixture HTML synthétique dans un répertoire temporaire unique, utilise l’URL `file://` correspondante, puis `afterAll` supprime ce répertoire. Aucun fichier sous `/tmp/t15-fixtures` ne doit être préparé manuellement ou ne doit subsister après un run réussi.

Les scénarios T14 et T15 utilisent `FORGELOCAL_DASHBOARD_URL` pour le navigateur, et construisent les en-têtes `Origin` et `Referer` à partir de cette URL. T15 construit également son URL healthcheck depuis `FORGELOCAL_CORE_BASE_URL`. Ils n’imposent donc ni `localhost`, ni le port `3000`, ni le port `19280`.

## Ordre de validation

Exécuter d’abord les scénarios directement affectés, puis la suite séquentielle complète :

```sh
pnpm run check
pnpm exec playwright test tests/automation-t15.spec.ts --workers=1
pnpm exec playwright test tests/runtime-t14.spec.ts --workers=1
pnpm exec playwright test --workers=1
```

Le résultat attendu du dernier run est **22 passed / 0 failed**. Le run doit échouer rapidement si le dashboard ou le Core ne répondent pas sur les URL exportées ; il ne doit jamais attendre un formulaire sur un port implicite.

Pour chaque run, archiver le log brut contenant la commande exacte, le répertoire de travail, les URLs non secrètes, l’horodatage UTC, `git rev-parse HEAD` et l’exit code. Ne jamais écrire la valeur du token dans un log.

Un script de preuve peut partager un cache **immutable** du binaire Chromium entre plusieurs bases-dir temporaires pour éviter un téléchargement répété. Cela ne change pas l’isolement : chaque run conserve une nouvelle `base-dir`, une base SQLite propre, un profil propre et un token temporaire distinct. Aucun cache de navigateur, profil, cookie, base SQLite ou jeton n’est réutilisé.

## Contraintes permanentes

Les jetons restent mémoire seule côté dashboard et temporaires côté harness. Les sorties de test ne doivent jamais afficher leur valeur. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` ne sont pas modifiés par ce guide.
