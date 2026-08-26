# GOSEC-R6 Lot C — permissions, conversions, aléatoire et motif secret

## Baseline et périmètre

Le Lot C a été exécuté sur `validation/gosec-r6` depuis le commit source post-commit `6bdda53c917b9617f530729528a0fa6bf80b94f2`. La vérification physique préalable du package R6-B a été effectuée depuis un clone public neuf; les hashes, extractions, manifestes, bundle, checkout et `git fsck --full` sont documentés dans `R6_C_BASELINE_DISCOVERY_RAW.log`.

Le JSON Gosec source-only autoritatif utilise `./cmd/... ./internal/...`. Il contient 26 findings Lot C sur 46 findings globaux : G302=5, G115=3, G404=17 et G101=1. La matrice `R6_C_FINDING_MATRIX.tsv` contient une ligne individuelle pour chaque finding.

| Règle | Baseline | Après Lot C | Évolution | Décision dominante |
|---|---:|---:|---:|---|
| G302 | 5 | 5 | 0 | MITIGATED_CONTROL_SCANNER_OPEN |
| G115 | 3 | 3 | 0 | MITIGATED_CONTROL_SCANNER_OPEN |
| G404 | 17 | 17 | 0 | SCANNER_OPEN_MANUAL_REVIEW |
| G101 | 1 | 1 | 0 | SCANNER_OPEN_MANUAL_REVIEW |
| **Total Lot C** | **26** | **26** | **0** |  |

## G302 — permissions

Le finding groups concerne `os.Chmod(dataDir, 0700)`, ce qui protège le répertoire contenant `groups.json`; la persistance utilise un fichier temporaire en 0600 puis un rename atomique. Une régression vérifie maintenant le répertoire 0700 et `groups.json` 0600. Les trois findings browser concernent les exécutables runtime en 0755. Ce mode est fonctionnellement nécessaire pour un binaire, tandis que les racines runtime/staging sont 0700 et les archives sont bornées et filtrées. Le finding CLI concerne les répertoires runtime réparés en 0700.

Ces contrôles ne clôturent pas automatiquement Gosec. Les cinq lignes restent `MITIGATED_CONTROL_SCANNER_OPEN`; la vérification native du propriétaire et d’un accès par un autre utilisateur doit encore être effectuée dans un environnement multi-utilisateur représentatif.

## G115 — conversions

Les trois conversions sont contrôlées par les flux précédents. `fingerprintInt` ne retourne qu’une valeur strictement positive avant conversion; le chemin `FingerprintSeed` est également conditionné à une valeur positive. Les tailles d’archives refusent les valeurs négatives, appliquent des limites par fichier et globales, puis vérifient que la taille uint64 tient dans int64 avant `io.CopyN`. Les tests couvrent valeur négative, valeur très grande refusée, valeur positive importante et seed positif/négatif. Les trois findings restent scanner-visible sous `MITIGATED_CONTROL_SCANNER_OPEN`, sans suppression du finding.

## G404 — aléatoire

Les 17 occurrences utilisent `math/rand/v2` pour la simulation de mouvements clavier/souris, les délais/jitter et la sélection d’un fingerprint inutilisé. La génération du token d’administration utilise séparément `crypto/rand`. Aucun usage G404 observé ne produit une clé, un nonce de sécurité, un mot de passe, une autorisation ou une décision d’authentification. Remplacer mécaniquement ces appels par `crypto/rand` modifierait la sémantique de la simulation et pourrait casser la cohérence des profils. Les occurrences restent donc `SCANNER_OPEN_MANUAL_REVIEW`, jamais `CLOSED`.

## G101 — motif secret

Le seul G101 vise le littéral de nom de fichier `.api-token.meta`, pas une valeur de credential. Le fichier stocke un digest SHA-256, les dates et l’état de révocation; le token brut est généré séparément par `crypto/rand` et n’est pas persisté dans les métadonnées. Les tests admin-token vérifient la persistance sans valeur brute, les transitions missing/malformed/invalid/expired/revoked et la redaction. Le finding reste `SCANNER_OPEN_MANUAL_REVIEW` parce que Gosec le signale, sans allowlist ni masquage.

## Gates et limites

Les tests ciblés permissions/bornes/token, `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, `go build ./cmd/... ./internal/...` et `git diff --check` sont PASS dans `R6_C_FINAL_POSTCOMMIT_RAW.log`. Gosec retourne exit code 1 avec 46 findings globaux et 26 Lot C. Gitleaks, Govulncheck, OSV Go/pnpm et Trivy sont PASS dans les périmètres documentés; les outils absents restent indisponibles dans le raw s’ils ne sont pas présents.

Aucun secret réel, compte réel, cookie, proxy commercial, site externe, runtime de production, SystemVault natif, Camoufox, Docker/Buildx ou plateforme Windows/macOS n’a été utilisé. Les environnements non disponibles ne sont pas simulés. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts.

Le verdict Lot C est `GOSEC_R6_LOT_C_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
