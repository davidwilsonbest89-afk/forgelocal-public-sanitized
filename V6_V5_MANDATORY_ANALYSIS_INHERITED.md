# Analyse comparative obligatoire — v5

Les résultats sont produits avec GolangCI-Lint 2.13.1 (binaire publié, compilé avec Go 1.27.0) contre la baseline et le HEAD, et avec un scan Gitleaks par arbre de commit pour compenser le mode `--log-opts` qui annonçait zéro commit. Les valeurs d’alerte Gitleaks restent redacted.

| Contrôle | Baseline | HEAD | Nouveau | Résolu | Décision |
|---|---:|---:|---:|---:|---|
| GolangCI-Lint 2.13.1 | 90 | 90 | 2 | 2 | Findings conservés et classés ci-dessous |
| Gitleaks arbres de commits | — | 348 détections cumulées | — | — | Findings historiques redacted, non-PASS |

## Findings GolangCI-Lint nouveaux

| # | Linter | Règle | Fichier | Ligne | Message | Risque | Propriétaire | Condition de levée |
|---:|---|---|---|---:|---|---|---|---|
| 1 | ineffassign | `ineffassign` | `cmd/server/cli_runtime.go` | 240 | ineffectual assignment to cfg | Revue sécurité/qualité requise avant correction ; aucune exploitation établie par le scan seul | Mainteneurs ForgeLocal | Correction/test ou justification versionnée, puis rerun avec code attendu |
| 2 | staticcheck | `staticcheck` | `internal/api/sessions.go` | 310 | SA1019: (github.com/mxschmitt/playwright-go.Page).WaitForSelector is deprecated: Use web assertions that assert visibility or a locator-based [Locator.WaitFor] instead. Read more about [locators]. | Revue sécurité/qualité requise avant correction ; aucune exploitation établie par le scan seul | Mainteneurs ForgeLocal | Correction/test ou justification versionnée, puis rerun avec code attendu |

## Findings GolangCI-Lint résolus entre baseline et HEAD

2 diagnostic(s) présents à la baseline et absents au HEAD selon la clé linter/message/position. Cette variation est conservée pour revue et ne constitue pas une approbation globale.

## Gitleaks explicite par arbres

La plage contient 58 commits. Le scan explicite a produit 348 détections cumulées, principalement `generic-api-key` dans des fixtures et preuves historiques. Les JSON individuels sont livrés sans secret en clair ; les alertes ne sont pas reclassées automatiquement comme faux positifs.

| Règle | Détections cumulées |
|---|---:|
| `generic-api-key` | 348 |

## Compléments Go

`go test -shuffle=on -count=3 ./...` et `go test -shuffle=on -count=3 -race ./...` ont terminé avec code 0. Aucun code produit n’a été modifié pendant cette campagne.
