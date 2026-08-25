# V6 — correspondance findings GolangCI → corrections

| Finding initial | Correction | Preuve ciblée |
|---|---|---|
| `ineffassign`, `cmd/server/cli_runtime.go:240` | Remplacement de l’initialisation immédiatement écrasée `cfg := map[string]any{}` par `var cfg map[string]any`; les branches `stdio` et `http` continuent d’affecter la configuration, et le chemin invalide conserve son code 2. | `TestCLIStatusAndMCPConfigJSON`; `V6_TARGETED_TESTS.log`; `V6_GOLANGCI_AFTER_GO.json` |
| `staticcheck SA1019`, `internal/api/sessions.go:310` | Remplacement de `Page.WaitForSelector` par `page.Locator(req.Selector).WaitFor` avec `LocatorWaitForOptions`; le timeout et la réponse 408 restent inchangés. | `go test ./internal/api`; `go test -race ./internal/api`; `V6_GOLANGCI_AFTER_GO.json` |

Le linter compatible Go 1.25 retourne 89 findings historiques dans ce checkout. Aucun finding `ineffassign` ne subsiste et l’ancien `sessions.go:310` n’apparaît plus ; les occurrences SA1019 restantes concernent d’autres fichiers non inclus dans le mandat V6.
