# GOSEC-R5 Lot R5-C — permissions, conversions, randomness et secrets

## Périmètre

R5-C couvre les 26 findings restants des règles G101, G115, G302 et G404. L’analyse est volontairement individuelle et n’ajoute aucun changement de code : les contrôles R4/R5 déjà présents sont rejoués, mais un contrôle applicatif existant ne transforme pas automatiquement une alerte statique en clôture.

Le lot est exécuté sur `validation/gosec-r5` à `HEAD=9412b314a3cdaff51d59b562ff2f97504844d9dc`. Aucun fichier `internal/` ou `cmd/` n’est modifié par R5-C; les preuves sont publiées séparément.

## Résultats exécutés

| Contrôle | Résultat | Preuve |
|---|---:|---|
| Tests ciblés humanize/fingerprint/groups/browser/API/server | PASS | `R5_C_FINAL_RAW.log` |
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS | `R5_C_FINAL_RAW.log` |
| `go vet ./cmd/... ./internal/...` | PASS | `R5_C_FINAL_RAW.log` |
| `go build ./cmd/... ./internal/...` | PASS | `R5_C_FINAL_RAW.log` |
| Gosec source-only | exit 1 avec 59 findings ouverts | `gosec_after_r5c.json`, `R5_C_FINAL_RAW.log` |

Le scan reste à **59 findings**, avec la distribution suivante : G101=1, G115=3, G204=5, G302=5, G304=11, G404=17, G703=9 et G704=7. La matrice `R5_C_CLASSIFICATION.tsv` contient une ligne par finding et conserve tous les statuts `OPEN_REVIEW`.

## Décisions de classification

Les **G404** concernent les fonctions de humanisation et la sélection de fingerprints, qui utilisent un aléatoire non cryptographique pour produire une variation comportementale et non pour générer un secret, un token ou une clé. Le remplacement aveugle par `crypto/rand` changerait la sémantique et n’est pas justifié sans décision produit; les findings restent cependant ouverts au regard du scanner.

Les **G302** sont répartis entre répertoires privés en `0700`, fichiers privés en `0600` et binaires installés en `0755`. La politique de permissions est couverte par les tests existants, mais le finding statique est conservé car les attentes de Gosec ne constituent pas à elles seules la politique produit. Les **G115** restent ouverts malgré les bornes d’archives et de runtime déjà présentes; la preuve de bornage ne supprime pas l’obligation de revue de chaque conversion. Le **G101** de `internal/api/admin_token.go` pointe une constante de version/metadata et non la valeur d’un secret, mais reste ouvert en revue manuelle plutôt que supprimé automatiquement.

Aucune directive `nosec`, `nolint`, allowlist ou réduction artificielle du périmètre n’a été utilisée. Les gates restent `GOSEC_R5_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R5_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
