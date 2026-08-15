# Addendum v2.4 — Bootstrap local de session lecture seule

**Statut :** contrat produit local ; aucun effet sur les gates publics de release.
**Date :** 15 août 2026.
**Portée :** API Core loopback, CLI locale et dashboard ForgeLocal lecture seule.

## 1. Objectif et frontières

Le dashboard ForgeLocal ne reçoit jamais le Bearer principal du Core. Il consomme uniquement une session courte, non persistée et strictement limitée aux endpoints `/api/v1/readonly/*`.

> Le code de bootstrap et le token de session sont des secrets temporaires. Ils ne doivent apparaître ni dans une URL, ni dans `localStorage`, `sessionStorage`, des analytics, des logs applicatifs ou une preuve CI.

Le Core reste lié au loopback. Les deux étapes du bootstrap vérifient que la requête provient de `127.0.0.0/8` ou `::1`; aucune origine distante ne peut donc émettre ni échanger un code de bootstrap.

Pour une validation de ce contrat sans qualifier, télécharger, activer ni lancer de navigateur, le Core peut être démarré avec `serve --no-runtime`. Ce mode désactive tous les runtimes dans l’instance locale concernée et autorise uniquement l’observation des surfaces API ; il ne constitue pas une qualification de Camoufox.

Le Core n’accepte les prévols CORS que depuis une origine HTTP(S) de loopback (`localhost`, `127.0.0.0/8` ou `::1`). L’origine exacte est renvoyée dans `Access-Control-Allow-Origin`; une origine distante reçoit `403 CORS_ORIGIN_NOT_ALLOWED`. Cette règle permet au dashboard local de faire le bootstrap sans étendre la surface d’accès à une prévisualisation ou à un hôte distant.

Le dashboard local cible le Core sur `http://<hôte-loopback>:19280` par défaut, même s’il est servi par son propre port de développement. Une valeur de compilation `VITE_CORE_BASE_URL` n’est honorée que si elle est une URL HTTP loopback valide ; toute valeur distante ou malformée est ignorée.

## 2. Parcours contractuel

| Étape | Acteur | Route ou commande | Garanties |
|---|---|---|---|
| Émission | CLI locale authentifiée | `forgelocal readonly-session code` | Le Bearer principal reste dans la CLI/Core ; le code aléatoire est utilisable une fois pendant 10 minutes. |
| Échange | Dashboard local | `POST /api/v1/readonly/session/bootstrap` avec `{ "code": "…" }` | Loopback obligatoire, corps limité à 1 KiB, code supprimé avant l’émission du token. |
| Session | Dashboard | `Authorization: Bearer <token court>` vers `/api/v1/readonly/*` | Token aléatoire en mémoire Core, validité 15 minutes, portée lecture seule uniquement. |
| Expiration | Core/Dashboard | Aucun renouvellement implicite | Une nouvelle session impose un nouveau code local. |

L’émission utilise la route authentifiée par le Bearer principal : `POST /api/v1/readonly/session/codes`. Elle n’est accessible qu’en loopback. Elle renvoie seulement le code, son expiration et `scope: "readonly"`.

## 3. Séparation d’autorisation

Le middleware lecture seule accepte soit le Bearer principal, soit un token de session court valide, mais il protège exclusivement les routes suivantes :

- `GET /api/v1/readonly/health`
- `GET /api/v1/readonly/summary`
- `GET /api/v1/readonly/profiles`
- `GET /api/v1/readonly/groups`
- `GET /api/v1/readonly/runtimes`

Toutes les routes métier — création, modification, suppression, lancement de runtime, backups, restauration, sessions navigateur et API Playwright — restent derrière le middleware du Bearer principal uniquement. Un token court n’accorde jamais ces droits.

## 4. Interface historique retirée

Le dashboard HTML historique BrowseForge est désactivé. Il ne reçoit plus de token, ne contient plus de formulaire de connexion, ne lit ni n’écrit `localStorage`/`sessionStorage`, et ne peut plus exécuter de mutation métier. Cette décision remplace le précédent parcours qui pouvait exposer le Bearer principal dans un fragment d’URL ou une persistance navigateur.

## 5. Intégration du dashboard React

Le client React ForgeLocal doit garder le token de session uniquement dans une closure mémoire, l’effacer à l’expiration ou à l’erreur `401`, et revenir à un état **Core indisponible** ou **Session expirée**. Il ne doit pas tenter de contourner les restrictions CORS, loopback ou de scope.

Le raccordement fonctionnel est autorisé seulement depuis une distribution explicitement approuvée par la configuration CORS locale du Core. Le dashboard publié de démonstration reste une maquette et n’est pas une autorité de session.

## 6. Tests obligatoires

La livraison exige notamment la preuve automatisée de :

1. rejet de l’émission et de l’échange hors loopback ;
2. code invalide, expiré ou déjà consommé rejeté ;
3. token court expiré rejeté ;
4. token court accepté sur chaque endpoint lecture seule ;
5. token court refusé sur au moins une route métier ;
6. CLI sans affichage du Bearer principal et ouverture de dashboard sans Bearer dans l’URL ;
7. dashboard historique sans persistance de token.

## 7. État release

Cet addendum ne qualifie pas Camoufox, ne modifie pas le RC BACK-01, ne clôture pas l’alerte `generic-api-key` et ne lève aucun des cinq gates publics. `PUBLIC_RELEASE_BLOCKED` reste obligatoire.
