# T32 — Contrat des diagnostics ClientRects

**Statut :** `T32_CLIENTRECTS_PROJECTED_UNSUPPORTED_VERIFIABLE_LOCAL`
**Baseline :** `t00-t27-complete-20260820` / `69411e65c880d168832a65fc8475cc97d562a9ad`
**Prérequis :** T31 publié sur `work/t31-canvas-webgl-audio`.

## Périmètre fermé

T32 ajoute le contrôle projeté `client-rects` à la réponse de diagnostic d’environnement. Il reste à l’état `UNSUPPORTED` et porte uniquement la note fixe `runtime observation not implemented`.

Aucune géométrie DOM n’est lue, calculée ou persistée. Le contrat interdit les valeurs `x`, `y`, `width`, `height`, les objets `DOMRect`, les rectangles de bounding box, les dimensions d’éléments et toute empreinte de mise en page.

## Invariants

La réponse conserve `PROJECTED_METADATA_ONLY`, l’ordre déterministe des contrôles, le refus des profils inconnus et la redaction de T31. Aucun navigateur, runtime, Camoufox, proxy, cookie, port ou UI n’est lancé. L’état `UNSUPPORTED` n’est jamais converti en `PASS` par défaut.

## Critères d’acceptation

| Critère | Résultat attendu |
|---|---|
| Présence | `client-rects` figure exactement dans la projection |
| État | `UNSUPPORTED` |
| Note | `runtime observation not implemented` |
| Déterminisme | Position et valeur stables entre deux diagnostics identiques |
| Redaction | Aucun `DOMRect`, rectangle, coordonnée ou dimension dans le JSON |
| Qualification | `go test -race ./...`, `go vet ./...`, `go build ./...`, format Git et Gitleaks du delta passent |

T32 reste un incrément local de contrat et de projection. Il n’autorise ni observation réelle de layout ni release. T33 sera le prochain lot après publication et vérification du commit T32.
