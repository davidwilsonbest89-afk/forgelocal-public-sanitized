# ForgeLocal — Smoke Dashboard profil/proxy après base vierge

## Verdict

Le verdict final du parcours demandé est **`SMOKE_DASHBOARD_PROFILE_CREATE_PASS`**, complété par **`SMOKE_DASHBOARD_PROFILE_RESTART_PASS`**.

La correction source est publiée dans le commit `0c4b9aebe530dc9cccaed6bf0172fa42b623a4b8` sur `validation/operational-v1`.

Le test a utilisé une base temporaire vide, un Core loopback, un Dashboard loopback, deux profils synthétiques créés par de vrais clics Dashboard, deux proxies HTTP locaux authentifiés et des destinations HTTP locales. Aucun compte, cookie utilisateur, proxy commercial, site externe, Camoufox, Docker/Buildx ou SystemVault natif n’a été utilisé.

## Corrections du lot

Le catalogue readonly du Core projette désormais les runtimes effectivement qualifiés même lorsque `runtime_candidates` est vide. Le Dashboard conserve la session readonly authentifiée en mémoire afin que `runWrite` puisse recharger les profils après une mutation. Le rafraîchissement utilise la limite Core autorisée de 100 profils. La projection readonly expose seulement le booléen redacted `proxy_configured` lorsqu’une affectation durable existe ; elle ne renvoie ni endpoint, ni référence secrète, ni credential.

## Preuve v8 — création depuis le Dashboard et chaîne proxy

Le journal `SMOKE_DASHBOARD_PROFILE_CREATE_RAW_V8.log` montre :

| Contrôle | Résultat observé |
|---|---|
| Bootstrap Dashboard/Core | `dashboard_admin_linked=linked` |
| Création profil alpha | `profile_created`, `creation_via_dashboard=true`, identifiant présent |
| Création profil beta | `profile_created`, `creation_via_dashboard=true`, identifiant présent |
| Affectation alpha | persistée, identifiant profil/proxy présents |
| Session Core alpha | création HTTP 201, navigation HTTP 200 |
| Hop destination alpha | `destination_via_proxy=1`, total 3, via proxy 3 |
| Arrêt proxy alpha | session Core navigation HTTP 500, destination inchangée, `result=fail-closed` |
| Mauvais credentials synthétiques | HTTP 407, destination inchangée |
| Affectation beta | persistée, distincte d’alpha |
| Isolation beta | destination beta via proxy beta, proxy alpha non utilisé |
| Cibles externes | 0 forward externe ; 3 tentatives Chromium de fond bloquées par le proxy loopback-only avant tout forward |
| Credentials dans les événements | `credential_value_logged=false` |

## Preuve redémarrage — persistance

Le journal `SMOKE_DASHBOARD_PROFILE_RESTART_RAW_V4.log` montre qu’après arrêt puis redémarrage des services avec le même répertoire de données :

| Contrôle | Résultat |
|---|---|
| Reconnexion Dashboard | PASS |
| Profil alpha présent | PASS |
| Profil beta présent | PASS |
| Affectation proxy alpha présente côté Core | PASS |
| Affectation proxy beta présente côté Core | PASS |
| Affectations distinctes | PASS |
| Indicateur UI alpha `Configuration liée` | PASS |
| Révocation du token après preuve | HTTP 204 puis sonde HTTP 401, raison `revoked` |
| Capture Dashboard | `dashboard_restart_persistence_v4.png` |
| Cleanup final | `CLEANUP_FINAL_RAW.log`, ports et processus résiduels absents |

Les premières exécutions de reprise V1–V3 ont été conservées comme diagnostics redacted : elles ont établi que le Core persistait correctement les affectations et que l’échec restant était celui du sélecteur de capture UI. Elles ne sont pas utilisées comme verdict final. Le journal `CLEANUP_FINAL_RAW.log` confirme la libération des ports et processus de la campagne.

## Vérifications techniques

`go test ./...`, `go vet ./...`, `go build ./...`, `pnpm run check`, Gitleaks et Govulncheck ont réussi sur l’état vérifié. Gitleaks a signalé zéro fuite et Govulncheck aucune vulnérabilité. `git fsck --full` a retourné exit_code=0 lors de la vérification de conservation. Gosec reste en échec avec 177 findings historiques/non masqués ; ce résultat ne constitue pas une régression masquée et le gate reste inchangé. Aucun finding n’a été supprimé par allowlist globale, skip ou nolint.

## Limites et statut des gates

| Catégorie | Statut |
|---|---|
| Smoke Dashboard profil/proxy/session local | PASS |
| Smoke persistance après redémarrage | PASS |
| Fail-closed proxy arrêté et credentials invalides | PASS |
| Gosec | FAIL_GATE_UNCHANGED — 177 findings |
| Création Dashboard de profil | PASS pour ce parcours local corrigé |
| Camoufox/runtime ciblé | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| SystemVault natif | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| Docker/Buildx | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| Proxies commerciaux, comptes/cookies réels, sites externes | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| Production-ready | NON DÉCLARÉ |

Ce lot ne modifie pas T28 historique, ne démarre pas T29 et ne touche pas T31–T38.
