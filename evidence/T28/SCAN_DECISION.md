# Décision de scans — T28

## Résumé autoritatif

Les contrôles finaux ont été exécutés depuis le module T28 au HEAD d’implémentation `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578`.

| Scan | Résultat autoritatif | Interprétation |
|---|---|---|
| Gitleaks extraction `--no-git` | code 0, rapport vide | Aucun secret détecté dans l’extraction T28 |
| Gitleaks diff | code 0, rapport vide | Aucun secret détecté ; le compteur historique `0 commits scanned` de `--log-opts` est conservé comme limitation de cet outil/clone shallow |
| Gosec `./internal/extensions` | code 0, `found=0` | Aucun finding propre au repository T28 |
| Gosec package API | code 1, 182 findings | Findings hérités dans d’autres fichiers API ; aucun finding dans `extensions.go` ou le test T28 |
| Govulncheck `./...` | code 0 | Aucune vulnérabilité trouvée |
| OSV Scanner | code 127 | Binaire absent ; aucune conclusion OSV simulée |
| Syft SPDX | code 0 | SBOM généré |
| `git diff --check` | code 0 au HEAD final | Aucun whitespace error introduit par T28 |

Les premiers diagnostics de CWD incorrect, d’option Gitleaks non supportée et de rapport Gosec avant correction sont conservés dans `SCANS_RAW.log` pour la traçabilité. Seuls les contrôles explicitement marqués corrigés/final sont utilisés pour le verdict.

## Suite Go globale

La suite globale `go test -count=1 -race ./...` échoue uniquement sur le finding V6 préexistant `TestNewRegistryLoadsBrowseForgeChromiumFromDefaultConfig` dans `internal/runtime/runtime_test.go`, qui attend une configuration Docker/GHCR Chromium absente. Les packages T28 passent sous race detector. `go vet ./...` et `go build ./...` passent.

## Sécurité du périmètre

Aucun navigateur, Camoufox, runtime d’extension, téléchargement de package, proxy/cookie réel, SystemVault natif, migration, production runtime ou release n’a été exécuté. Les packages et rapports ne contiennent pas de donnée utilisateur réelle.


## Addendum R1 — qualification probatoire du 25 août 2026

| Contrôle R1 | Résultat autoritatif | Conclusion |
|---|---|---|
| Test global baseline | `go test -count=1 -race ./...`, code 0 sur `999374d99b7996504ba91e421850a2fe84afb78d` | PASS ; le finding runtime historique n’est pas reproduit dans ce worktree propre |
| Test global HEAD | `go test -count=1 -race ./...`, code 0 sur `8b96981e9ffae245307d2b5cb279d6013c6dc11e` | PASS ; `internal/extensions` est présent et passe sous race |
| Tests T28 ciblés | `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'`, code 0 | PASS |
| OSV Scanner | v1.9.2, scan ciblé `go.mod` baseline/head, code 1 et 46 identifiants uniques sur chaque côté | Scan réel qualifié ; avis dépendances/std-lib à traiter séparément, aucun résultat inventé |
| Gitleaks plage | 8 commits réels inspectés, arbres des chemins modifiés, `gitleaks detect --no-git --redact`, code 0 pour chaque arbre | PASS, aucun leak ; ancien résultat `0 commits scanned` conservé comme historique séparé |
| Gitleaks ZIP | extraction fraîche et scan `--no-git` | PASS, aucun leak |
| Gosec baseline/head | 6 findings de chaque côté ; après normalisation par règle/fichier/détail : `new_findings=0`, `resolved_findings=0` | Findings historiques identiques, aucun finding introduit par T28 |

Les tentatives OSV v2.5.1, v2.4.0 et v2.3.8 sont conservées dans `OSV_R1_RAW.log` : elles exigeaient respectivement un Go 1.26.x absent. La version v1.9.2 a été installée explicitement avec Go 1.25.13 et le scan réel a ensuite été exécuté.

La qualification R1 ne lève aucune gate et ne remplace pas la revue indépendante. La valeur exacte maintenue est `camoflox_execution_authorized=false`. T28 ne lance ni navigateur, ni extension, ni proxy, ni processus externe ; `update_url`/`updateURL` est ignoré, non suivi et non exécuté.
