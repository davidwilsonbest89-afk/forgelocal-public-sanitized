# GOSEC-R6 Lot B — subprocess et réseau

## Périmètre fermé

Le Lot B a été exécuté depuis le commit source `a436a68d1ca4e223b54c45d7b99efa77617e620b` sur `validation/gosec-r6`. La baseline et le scan final utilisent exclusivement `gosec -fmt json ... ./cmd/... ./internal/...`. La matrice `R6_B_FINDING_MATRIX.tsv` contient une ligne par finding G204/G704.

La baseline recalculée depuis le HEAD R6-A contient **12 findings** : G204=5 et G704=7. Le scan post-correctif contient toujours G204=5 et G704=7. Aucun finding n’a été masqué et le code n’a pas été refactoré artificiellement pour faire disparaître les alertes.

| Règle | Baseline | Après Lot B | Évolution | Classification dominante |
|---|---:|---:|---:|---|
| G204 | 5 | 5 | 0 | MITIGATED_CONTROL_SCANNER_OPEN; plateformes natives indisponibles |
| G704 | 7 | 7 | 0 | MITIGATED_CONTROL_SCANNER_OPEN |
| **Total** | **12** | **12** | **0** |  |

## Analyse subprocess

Les cinq G204 correspondent aux commandes système d’ouverture du navigateur (`open`, `xdg-open`, `rundll32`) et au nettoyage macOS via `xattr`. Les exécutables et options sont codés en dur et les arguments sont séparés; aucun shell implicite ni binaire contrôlable par requête n’est utilisé. Les appels `xattr` sont conditionnés à Darwin et disposent d’un timeout dans les chemins observés. La sandbox Linux n’exécute ni GUI natif, ni macOS, ni Windows; ces aspects restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`. Gosec continue de signaler les sinks, donc la classification reste `MITIGATED_CONTROL_SCANNER_OPEN` et non CLOSED.

Les tests CLI négatifs existants refusent les hôtes externes avant dial, les userinfo, queries non prévues, fragments, schémas non autorisés et ports invalides. Les tests de validation workflow ajoutés en R6-A autorisent uniquement `full_page=true|false` et refusent les queries additionnelles. Les chemins GUI natifs ne sont pas simulés comme PASS.

## Analyse réseau et WebSocket

Les sept G704 sont des alertes de taint sur des destinations HTTP/WebSocket. Les appels API CLI passent par `validateCLILoopbackURL`, qui exige HTTP(S), hôte loopback, port 1–65535, sans userinfo, query ni fragment. Les clients ont des timeouts et refusent les redirections externes. Le pont WebSocket vérifie le endpoint manager-owned, accepte uniquement `ws`, loopback, port valide et chemin exact de la session, puis normalise le dial vers `127.0.0.1` avec timeout et deadline.

Ces contrôles réduisent l’atteignabilité SSRF, mais Gosec reste ouvert; chaque ligne demeure `MITIGATED_CONTROL_SCANNER_OPEN`. IPv6 loopback est couvert par les tests de validation lorsque la logique est pure, mais aucun service IPv6 réel n’est déclaré exécuté si l’environnement ne le fournit pas.

## Validation et limites

Les tests ciblés subprocess/réseau, la suite race, vet et build sont PASS dans `R6_B_FINAL_RAW.log`. Gosec retourne exit code 1 avec 46 findings globaux, dont les 12 Lot B. Govulncheck, OSV Go/pnpm, Trivy et Gitleaks corrigé sont PASS dans leurs périmètres documentés. Le premier raw Gitleaks avec plage vide est conservé; le raw corrigé utilise explicitement `a436a68d1ca4e223b54c45d7b99efa77617e620b^..a436a68d1ca4e223b54c45d7b99efa77617e620b` et retourne PASS.

Semgrep, Grype, Shellcheck et Yamllint restent indisponibles si non présents. Aucun compte, cookie, secret, proxy commercial, site externe, Camoufox, SystemVault natif, Docker/Buildx ou runtime de production n’a été utilisé. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts.

Le verdict Lot B est `GOSEC_R6_LOT_B_CLASSIFIED_WITH_OPEN_FINDINGS`, avec `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
