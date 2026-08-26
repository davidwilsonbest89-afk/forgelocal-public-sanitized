# GOSEC R2 overnight — rapport final

## Verdict

```text
GOSEC_R2_OVERNIGHT_HARDENING_COMPLETE_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R2_LOT2_PRESERVED_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
FORGELOCAL_PRODUCTION_READY=false
```

Cette mission est une revue automatisée/agent avec tests locaux synthétiques. Elle ne constitue pas une revue humaine indépendante et ne valide ni une release, ni Camoufox, ni SystemVault natif, ni Docker/Buildx.

## Réconciliation de lignée

La vérification a été réalisée depuis un clone GitHub neuf, sans téléchargement LFS implicite, sur `validation/operational-v1`. La lignée est cohérente :

| Référence | Parent direct | Rôle | Relation vérifiée |
|---|---|---|---|
| `20e5181554c52a92d6e1acad15feb426e8804621` | `cba6beb496e0c56f98f03bdf1dc3efbd7adb3809` | HEAD final Lot 1 annoncé | commit présent; aucun tag associé |
| `31385c26bc6ca8944d68b50be71fec8c0783d590` | `20e5181554c52a92d6e1acad15feb426e8804621` | conservation Lot 1 v2 | descendant direct de `20e5181`; merge-base=`20e5181` |
| `701c5949261de261d2044cbff3e125b88c56f1a2` | `31385c26bc6ca8944d68b50be71fec8c0783d590` | source Lot 2 subprocess/réseau | descendant direct de `31385c2` |
| `209148020e8254d7af903bc09485dbaed4014948` | `701c5949261de261d2044cbff3e125b88c56f1a2` | test timeout final Lot 2 | descendant direct de `701c594` |
| `aab0cca666e470ed312aabcdb5e369157ba4f204` | `9e2361eb2428a50ece22f806c8a3c79d3a85ffe1` | HEAD final Lot 2 pré-overnight | branche distante vérifiée |
| `18367a6d651657e8e48afc006ad75bfa95aa46ea` | `aab0cca666e470ed312aabcdb5e369157ba4f204` | filesystem complémentaire | source publié |
| `7f3f5bef4e8dba800c7548532e12fea09a876b46` | `18367a6d651657e8e48afc006ad75bfa95aa46ea` | permissions/I/O | source publié |
| `f796299ef12c7c701937d53afd6e020088d95c3c` | `7f3f5bef4e8dba800c7548532e12fea09a876b46` | directive Go patchée pour OSV | HEAD final overnight |

Les tests d’ancêtre retournent `0` pour `20e5181 → 31385c2`, `31385c2 → 701c594`, `701c594 → 2091480` et `2091480 → aab0cca`. La branche locale et `origin/validation/operational-v1` pointent maintenant vers `f796299...`; les commits n’ont pas de tag Git associé (`tag_absent` consigné, aucun tag inventé).

Le diff `20e5181..31385c2` est limité à la conservation Lot 1 v2 : manifeste corrigé, logs de préservation et ZIP/TAR v2 avec sidecars. Le diff Lot 2 est composé des commits source, tests, preuves et packages listés ci-dessus; les fichiers métier modifiés sont déclarés dans les rapports de chaque lot.

## Corrections publiées pendant l’overnight

Le confinement backup a été corrigé pour normaliser chaque séparateur Windows, avec régression single-backslash et double-backslash. L’ouverture HTTP d’artifacts utilise maintenant `os.OpenRoot`, ce qui empêche un symlink externe de sortir de la racine entre validation et ouverture. Le test HTTP artifact reçoit un `404` et ne sert pas le contenu externe.

Les modes des logs, configs, migrations, backups, user-data Chromium/Firefox, préférences, persona native, captures MCP et sorties workflow ont été resserrés. Les fermetures de tar/gzip/fichier de backup sont propagées explicitement. Le workflow applique une URL HTTP(S) loopback, un timeout, le refus des redirections externes et la propagation des erreurs de marshal/request/JSON.

La directive Go est maintenant `go 1.25.13`, cohérente avec `toolchain go1.25.13`. OSV Go est passé de 46 avis (`go 1.25.0`) à 0 avec cette exigence de toolchain patché; ce changement est publié dans `f796299...` et n’est pas une suppression de scanner.

## Tests et scans

| Contrôle | Résultat |
|---|---:|
| Tests archive traversal/symlink/hardlink/type spécial/limites/rollback | PASS |
| Test artifact symlink externe avec `os.OpenRoot` | PASS |
| Tests modes logger et backup privé | PASS |
| Tests proxy/profile/session/redaction/navigation | PASS |
| Tests workflow | PASS |
| Tests CLI loopback, redirection, timeout, API GET/POST et metadata | PASS |
| `go test -race ./cmd/... ./internal/...` | PASS |
| `go vet ./cmd/... ./internal/...` | PASS |
| `go build ./cmd/... ./internal/...` | PASS |
| Dashboard `pnpm run check` | PASS |
| Gosec source-only | 132 findings, `exit_code=1` |
| Govulncheck source-only | PASS |
| Gitleaks redacted | PASS |
| OSV `go.mod` après Go 1.25.13 | PASS, 0 avis |
| OSV Dashboard lockfile | PASS, 0 avis |
| Trivy vuln/secret/misconfig | PASS, 0/0/0 |
| Syft | PASS |
| Semgrep, Grype, Shellcheck, Yamllint | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE |

Le Gosec final conserve notamment G104=36, G204=5, G301=17, G302=4, G304=20, G306=1, G703=12 et G704=7, ainsi que G101/G107/G115/G118/G122/G305/G404. Aucun finding n’est masqué ou déclaré clos automatiquement. Les matrices individuelles permissions et G204/G704 sont publiées sous `evidence/GOSEC_R2_OVERNIGHT/`.

OSV Go a été analysé individuellement dans `OSV_GO_MOD_46_FINDINGS.tsv` avant la mise à niveau de la directive; après `go 1.25.13`, le scan actuel porte sur 24 paquets et retourne zéro avis. Les outils absents restent explicitement indisponibles.

## Non-régression fonctionnelle

Les contrats Go impactés par le hardening passent et `pnpm run check` passe. La campagne Dashboard/Core/proxy complète n’a pas été refaite inutilement : les preuves canoniques déjà publiées restent `SMOKE_DASHBOARD_PROFILE_CREATE_PASS`, `SMOKE_DASHBOARD_PROFILE_RESTART_PASS` et `SMOKE_PROXY_LOCAL_PASS`. Aucun compte réel, cookie réel, proxy commercial, site externe ou runtime de production n’a été utilisé.

Les tests GUI `xdg-open/open/rundll32`, le subprocess Darwin `xattr` sur la plateforme native et le bridge Playwright avec session réelle ne sont pas exécutés dans cet environnement Linux; ils restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` ou `NEEDS_MANUAL_REVIEW`, jamais PASS implicite.

## Conservation et limites

Le Lot 2 original conserve son ZIP/TAR, ses sidecars, son bundle et son manifeste. Le manifeste Lot 1 v1 auto-référentiel n’a pas été supprimé; la v2 non auto-référentielle et sa preuve de vérification sont publiées. Pour l’overnight, le package v3 `forgelocal-gosec-r2-overnight-final-v3.zip/.tar.gz` contient la baseline, les queues, matrices, sorties brutes, JSON, rapports, changelogs, todo, manifestes, hashes, bundle, extraction fraîche, clone neuf, checkout explicite et fsck. Le package v1 est conservé comme première archive historique, et le package v2 est conservé comme archive intermédiaire. Le package v3 est la version canonique recalculée après la mise à jour finale du rapport et l’ajout de `FINAL_PRESERVATION_RAW.log`.

La première tentative de préservation a échoué sur un chemin relatif incorrect lors du contrôle des sidecars. La seconde a déclenché LFS parce que `GIT_LFS_SKIP_SMUDGE` n’était pas propagé au checkout explicite. Ces deux événements sont documentés dans `FINAL_PRESERVATION_RAW.log`; la troisième vérification, avec chemins neutres et checkout explicitement sans LFS, passe.

Le prochain triage autorisé peut couvrir les findings G703/G304/G301/G104/G404 restants et les dépendances selon une nouvelle baseline. Le Lot 3 n’est pas démarré automatiquement.

T28 n’a pas été rouvert, T29 n’a pas commencé et T31–T38 n’ont pas été modifiés. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent conservées telles quelles.
