# ForgeLocal — Registre de statut du lot proxy-session

**Source de référence avant conservation :** `9031f8324deeb4d6c171e01a244eeba778cf137c` sur `validation/operational-v1`.

| Catégorie | Élément | Statut exact | Justification |
|---|---|---|---|
| PASS | Intégration Dashboard → Core → session → Chromium → proxy alpha | `SMOKE_INTEGRATED_PROXY_PASS` | `create_status=201`, `navigate_status=200`, `destination_via_proxy=1` |
| PASS | Proxy alpha arrêté | `PASS_FAIL_CLOSED` | `ERR_PROXY_CONNECTION_FAILED`, aucune nouvelle destination, `destination_unchanged=true` |
| PASS | Isolation beta | `PASS_ISOLATION` | destination beta distincte et forwards beta observés, aucun mélange alpha/beta dans le contrôle |
| PASS | Proxy nominal indépendant | `PASS` | HTTP 200 et forwards proxy synthétiques observés |
| PASS | Mauvaise authentification indépendante | `PASS_REJECTED_407` | proxy bad-auth rejette les requêtes et la destination ne reçoit rien |
| PASS | No-proxy | `PASS_EXPLICIT_POLICY` | absence d’affectation conserve le chemin direct/profil/groupe ; test unitaire dédié |
| PASS | Révocation auth | `PASS` | HTTP 204 puis HTTP 401 avec raison `revoked` |
| PASS | Tests/build/vet | `PASS` | `go test ./...`, `go build ./...`, `go vet ./...` |
| PASS | Gitleaks et Govulncheck | `PASS` | aucune fuite détectée et aucune vulnérabilité signalée par l’outil disponible |
| FAIL | Gosec | `FAIL_GATE_UNCHANGED` | 177 findings dans le scan courant ; aucun masquage, allowlist globale ou nolint ajouté |
| NOT_IMPLEMENTED_PLACEHOLDER | Création directe d’un profil dans le Dashboard | `NOT_IMPLEMENTED_PLACEHOLDER` | catalogue readonly runtime vide dans cet environnement ; profils smoke préparés par Core comme précondition documentée |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Camoufox ciblé | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | environnement indisponible et installation interdite |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Docker/Buildx | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | daemon/environnement non disponible et installation interdite |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | SystemVault natif | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | smoke authentifié Core non branché à un backend natif ; les unit tests utilisent uniquement `MemoryVault` |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | Proxy commercial, comptes/cookies réels, sites externes | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | explicitement exclus du périmètre |
| FAIL | Production readiness | `FAIL_NOT_PRODUCTION_READY` | gates environnementales et réserves de sécurité demeurent ouvertes |

Le smoke intégré V5 est la seule source du verdict `SMOKE_INTEGRATED_PROXY_PASS`. Les exécutions V2–V4 interrompues ou invalides sont conservées comme contexte mais ne sont pas utilisées pour déclarer le PASS.
