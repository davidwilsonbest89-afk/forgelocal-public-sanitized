# GOSEC-REVIEW-R2 — Lot 2 : subprocess et réseau

## Verdict

```text
GOSEC_REVIEW_R2_LOT2_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
FORGELOCAL_PRODUCTION_READY=false
```

Le lot est une revue automatisée/agent avec tests locaux synthétiques. Il ne constitue pas une revue humaine indépendante et ne valide ni Camoufox, ni SystemVault natif, ni Docker/Buildx, ni une release.

## Périmètre et commits

Le périmètre fermé couvre les 13 findings G204/G704 présents dans la baseline Lot 2, sur `cmd/server/main.go`, `cmd/server/cli.go`, `cmd/server/cli_runtime.go`, `internal/browser/download.go` et `internal/api/sessions.go`. La matrice une ligne par finding est `GOSEC_REVIEW_R2_LOT2_MATRIX.tsv`.

La baseline a été prise au HEAD Lot 1 `31385c26bc6ca8944d68b50be71fec8c0783d590`. Le correctif source du Lot 2 est publié dans `701c5949261de261d2044cbff3e125b88c56f1a2`; le test de timeout local final est publié dans `209148020e8254d7af903bc09485dbaed4014948`, sur `validation/operational-v1`.

| Règle | Baseline | Head | Disposition R2-Lot 2 |
|---|---:|---:|---|
| G204 | 6 | 5 | 3 mitigés mais encore signalés; 3 GUI restent en revue manuelle. |
| G704 | 7 | 7 | 7 mitigés par validation loopback/redirect; findings statiques encore ouverts. |
| **Total du périmètre** | **13** | **12** | **Aucune clôture globale**. |

Le scan Gosec source-only complet est passé de 155 à **152 findings** après le correctif. Gosec reste volontairement en échec (`exit_code=1`); la baisse n’est pas convertie en clôture automatique.

## Correctifs et contrôles

Les appels `apiGET`, `apiPOST`, `doJSON`, metadata backup et metadata restore refusent maintenant toute URL qui n’est pas HTTP(S) loopback, ainsi que userinfo, query, fragment, port invalide et hôte externe. Le client HTTP a un timeout et refuse les redirections externes avant leur dial.

Le parcours `open --base-url` applique la même restriction avant d’invoquer le navigateur système. Le subprocess macOS `xattr` utilise des arguments séparés et `CommandContext` avec timeout de dix secondes; son exécution réelle n’a pas été lancée sur Linux.

Le pont WebSocket conserve la validation `ws`, loopback IPv4/IPv6/localhost, chemin exact, port 1–65535, absence de userinfo/query/fragment et normalisation vers `127.0.0.1`. Son dial utilise désormais `net.DialTimeout` et une deadline de handshake de cinq secondes; toute réponse backend autre que `101` devient `502` et aucune redirection HTTP n’est suivie.

## Tests exécutés

| Test ou contrôle | Résultat |
|---|---:|
| URL CLI IPv4 loopback | PASS |
| URL CLI IPv6 loopback | PASS |
| URL CLI `localhost` | PASS |
| schéma `ws`/`file` rejeté par client HTTP | PASS |
| hôte externe/IP externe rejeté avant dial | PASS |
| userinfo rejeté | PASS |
| query et fragment rejetés | PASS |
| port absent/invalide/hors limite rejeté selon contrat | PASS |
| redirection externe refusée | PASS |
| API GET/POST externe refusée avant dial | PASS |
| metadata backup/restore externe refusé avant usage du token | PASS |
| `open --base-url` externe refusé avant lancement navigateur | PASS |
| HTTP local nominal | PASS |
| HTTP local lent borné par timeout | PASS |
| WebSocket validator loopback/externe/path/userinfo/query/fragment | PASS |
| dial WebSocket avec timeout et deadline de handshake | Implémenté, bridge complet non exécuté sans session Playwright réelle |
| subprocess xattr Darwin | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| lancement GUI `xdg-open/open/rundll32` | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS |
| `go vet ./cmd/... ./internal/...` | PASS |
| `go build ./cmd/... ./internal/...` | PASS |

Les sorties brutes et les codes de sortie sont dans `GOSEC_REVIEW_R2_LOT2_TARGETED_TESTS_RAW.log` et `GOSEC_REVIEW_R2_LOT2_ALL_SCANS_RAW.log`.

## Scans

| Scan | Résultat |
|---|---:|
| Gosec source-only head | 152 findings, `exit_code=1` |
| Govulncheck source-only | PASS |
| Gitleaks redacted | PASS |
| OSV `go.mod` | 46 avis, `exit_code=1` |
| OSV Dashboard lockfile | PASS, 0 avis |
| Trivy vuln/secret/misconfig | PASS, 0/0/0 |
| Semgrep, Grype, Shellcheck, Yamllint | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |

Aucune destination externe, proxy commercial, cookie réel ou credential réel n’a été utilisé. Le test de redirection externe utilise uniquement la détection/refus de la cible et ne contacte pas le domaine externe.

## Dispositions

La matrice classe 10 findings `MITIGATED_CONTROL_SCANNER_OPEN`, car des contrôles et tests couvrent les chemins concernés mais Gosec continue de signaler les sinks. Les trois findings G204 de `openBrowser` restent `NEEDS_MANUAL_REVIEW` : les arguments sont séparés et les commandes sont constantes, mais le cycle de vie GUI spécifique à chaque plateforme et l’absence de processus orphelin ne sont pas démontrables dans l’environnement Linux autorisé.

Le prochain lot reste à définir après revue du présent résultat. Les findings G204/G704 ne sont pas clos et aucun statut `COMPLETE_NO_OPEN_FINDINGS` n’est justifié.

## Contraintes conservées

T28 n’a pas été redémarré, T29 n’a pas commencé et T31–T38 sont intacts. Les parcours Dashboard/Core/proxy déjà validés n’ont pas été refaits en campagne complète. Camoufox, SystemVault natif, Docker/Buildx, proxies réels, cookies réels et release restent hors périmètre.
