# Status register — Dashboard profile creation lot

**Lot fermé :** création Dashboard → affectation proxy → session Core → navigation locale → arrêt fail-closed → redémarrage/persistance.

**Source publiée :** `0c4b9aebe530dc9cccaed6bf0172fa42b623a4b8` sur `validation/operational-v1`.

| Domaine | Statut | Preuve |
|---|---|---|
| Création réelle de `smoke-alpha` par clic Dashboard | PASS | `SMOKE_DASHBOARD_PROFILE_CREATE_RAW_V8.log`, `creation_via_dashboard=true` |
| Création réelle de `smoke-beta` par clic Dashboard | PASS | `SMOKE_DASHBOARD_PROFILE_CREATE_RAW_V8.log`, `creation_via_dashboard=true` |
| Affectation alpha persistée | PASS | Événement `alpha_assignment_persisted` |
| Session Core alpha avec hop proxy | PASS | HTTP 201/200, `destination_via_proxy=1` |
| Proxy alpha arrêté sans connexion directe | PASS | HTTP 500 Core, destination inchangée, `fail-closed` |
| Mauvais credentials synthétiques | PASS | HTTP 407, aucune destination atteinte |
| Affectation beta distincte et isolation | PASS | `beta_assignment_persisted`, destination beta via proxy beta |
| Cibles externes | PASS | `external_forward_observed=false`, 3 cibles externes bloquées avant forward |
| Persistance après redémarrage Core + Dashboard | PASS | `SMOKE_DASHBOARD_PROFILE_RESTART_RAW_V4.log` |
| Projection UI `Configuration liée` après redémarrage | PASS | `alpha_ui_proxy_configured=true` |
| Révocation du token | PASS | HTTP 204 puis HTTP 401, raison `revoked` |
| Tests Go complets | PASS | `go test ./...` exit 0 |
| Vet/build Go | PASS | `go vet ./...` et `go build ./...` exit 0 |
| Typecheck Dashboard | PASS | `pnpm run check` exit 0 |
| Gitleaks | PASS | 0 leaks found, exit 0 |
| Govulncheck | PASS | No vulnerabilities found, exit 0 |
| Gosec | FAIL — gate inchangé | 177 findings non masqués, exit 1 |
| Camoufox | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Environnement non disponible, aucune installation |
| SystemVault natif | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Secret Service natif non disponible |
| Docker/Buildx | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Aucun daemon ni installation |
| Proxies commerciaux/comptes/cookies/sites externes | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Fixtures loopback uniquement |
| Production-ready | NON DÉCLARÉ | Gosec et gates d’environnement restent ouverts |

Les fichiers T28 historiques, T29 et T31–T38 ne font pas partie du périmètre modifié par ce lot.
