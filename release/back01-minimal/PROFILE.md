# Profil de distribution minimal — ForgeLocal BACK-01

## Objet

Ce profil produit un artefact local minimal destiné uniquement à l’administration de sauvegardes chiffrées de profils ForgeLocal. Il ne constitue pas une distribution du serveur BrowseForge historique, un runtime navigateur, ni un package d’automatisation.

> L’artefact ne peut être publié que si son manifeste, sa nomenclature de dépendances, ses licences, ses checksums et ses contrôles de sécurité correspondent exactement au contenu réellement livré.

## Binaire autorisé

| Élément | Valeur |
|---|---|
| Commande source | `./cmd/back01-core` |
| Binaire | `forgelocal-back01-core` |
| Écoute par défaut | `127.0.0.1:45100` |
| Authentification | Bearer token obligatoire via `FORGELOCAL_API_TOKEN` |
| Données locales | `FORGELOCAL_DATA_DIR` ou `~/.forgelocal-back01` |
| Dépendances internes autorisées | `internal/backup`, `internal/profile`, `internal/secrets` |
| Données publiées | binaire, migrations BACK-01, notices, manifestes, SBOM, checksums et signature si disponible |

## Exclusions de build obligatoires

Le graphe de dépendances du binaire ne doit pas importer les packages suivants :

```text
forgelocal/internal/browser
forgelocal/internal/fingerprint
forgelocal/internal/humanize
forgelocal/internal/mcp
forgelocal/internal/runtime
forgelocal/internal/workflow
```

Les répertoires correspondants peuvent demeurer dans le dépôt source, mais ne font pas partie de l’artefact minimal et n’autorisent aucune revendication de test, de sécurité ou de licence sur le binaire minimal.

## Fonctionnalités incluses

| Fonction | Incluse |
|---|---:|
| API locale authentifiée | Oui |
| Snapshot de `browser-data` | Oui |
| Format chiffré `FLBK … FLEND` | Oui |
| AES-256-GCM, AAD, checksum | Oui |
| SQLite local et recovery `published_unregistered` | Oui |
| Coffre système de clés et secrets proxy | Oui |
| Restauration sous nouvel ID | Oui |
| Rejet symlink, archive hostile et TOCTOU | Oui |
| Runtime Chromium/Camoufox | Non |
| Navigation réseau | Non |
| Playwright, MCP, extensions | Non |
| Fingerprinting/humanization | Non |

## Portes de publication

1. Le binaire doit compiler avec le toolchain approuvé, actuellement **Go 1.25.13**.
2. La fermeture des dépendances internes doit correspondre exactement à la liste autorisée.
3. `go test -race ./internal/backup ./internal/profile ./internal/secrets ./cmd/back01-core` doit réussir.
4. `gosec ./internal/backup/... ./internal/profile/... ./internal/secrets/... ./cmd/back01-core/...` ne doit signaler aucune alerte nouvelle non justifiée.
5. Un scan de secrets de l’arborescence staged et de l’archive finale doit être vert.
6. L’artefact final doit contenir un manifeste de provenance, une nomenclature des modules, un inventaire des licences et des checksums SHA-256.
7. L’acceptation AC-BACK-01 complète demeure conditionnelle à la relance d’un profil restauré avec un runtime approuvé ; ce runtime ne fait pas partie de cet artefact.

## Résultat attendu

Le script `scripts/build-back01-minimal.sh` génère sous `dist/back01-minimal/` une archive tar.gz avec un manifeste exact, les hash SHA-256, la nomenclature de dépendances et les notices de licences collectées. Toute dépendance interdite interrompt le build.
