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
