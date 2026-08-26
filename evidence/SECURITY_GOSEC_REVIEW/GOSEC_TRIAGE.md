# Triage Gosec — revue ciblée

## Statut

Le scan source-only du commit `e8581550b226dfcb60eb0c6b74b004901148a0e8` a produit **177 findings** : 49 HIGH, 75 MEDIUM et 53 LOW. La classification couvre `./cmd/... ./internal/...` et exclut les copies partielles sous `artifacts/`. Aucun finding n’est fermé, supprimé ou masqué.

## Règle de décision

Un finding est priorisé selon trois critères : présence dans le chemin exécuté par le smoke ou dans un chemin serveur accessible, impact potentiel sur SSRF/path traversal/subprocess/secrets, puis confiance Gosec. Une classification `à confirmer` n’est pas une acceptation du risque ; elle indique seulement qu’une revue de code et un test de frontière sont nécessaires.

## File prioritaire

| Priorité | Findings | Raisonnement et action |
|---|---|---|
| P0 — revue immédiate | `G704` sur `internal/api/sessions.go:477` | Le proxy WebSocket transforme `session.ConnectURL` en cible TCP. Le champ est normalement produit par le manager, mais la frontière doit être défensive : accepter uniquement un endpoint loopback correctement parsé, port explicite et chemin attendu ; refuser toute cible externe ou malformed. Ajouter un test négatif sans supprimer le finding. |
| P0 — revue immédiate | `G703`, `G304`, `G305`, `G110` dans `internal/browser/download.go` | Téléchargement/extraction et chemins issus de données runtime. Les traversals d’archive, les chemins variables et les bombes de décompression peuvent affecter l’intégrité du runtime. Revoir confinement de destination, limites de taille/nombre de fichiers et symlink/TOCTOU avant toute qualification production. |
| P0 — revue immédiate | `G703`, `G122`, `G304`, `G115` dans `cmd/server/cli_runtime.go` | Gestion de runtimes et filesystem. Vérifier que toute source et destination sont root-scoped, que les chemins ne sortent jamais de la base prévue et que les conversions de taille sont bornées. Ajouter des tests de traversal, symlink et overflow. |
| P1 — important | `G704` et `G204` dans `cmd/server/cli.go`/`cmd/server/main.go` | Les helpers de smoke et le lancement de sous-processus construisent des requêtes/commandes avec des valeurs variables. Vérifier si les entrées peuvent être contrôlées par un utilisateur distant ou seulement par un opérateur local ; imposer une allowlist de schémas/hôtes pour le mode local et une construction d’arguments sans shell. |
| P1 — important | `G301`, `G302`, `G306` sur `cmd/server`, `internal/browser` et `internal/config` | Les permissions de répertoires/fichiers doivent être comparées à la sensibilité réelle des données. Corriger en priorité secrets, tokens, profils et artefacts ; conserver une preuve des modes effectifs. |
| P1 — important | `G112` sur `cmd/server/main.go:414` | Absence de `ReadHeaderTimeout` : risque Slowloris sur le serveur HTTP. Ajouter des timeouts explicites compatibles avec les routes locales et un test de configuration. |
| P2 — hardening | `G115` hors chemins critiques et `G404` | Les conversions et sources pseudo-aléatoires doivent être bornées ou documentées selon l’usage. `G404` n’est acceptable que pour les fonctions de fingerprint/humanisation non cryptographiques ; toute décision de sécurité doit utiliser `crypto/rand`. |
| P2 — hygiène | `G104` | Les erreurs ignorées doivent être corrigées lorsqu’elles concernent fermeture, flush, suppression, audit ou sécurité. Les occurrences purement best-effort peuvent être documentées dans la revue, sans suppression globale. |
| À confirmer | `G101` `internal/api/admin_token.go:22` | Confiance LOW. Vérifier qu’il s’agit d’un marqueur/test ou d’un placeholder et qu’aucun secret réel n’est embarqué. Ne pas modifier avant inspection du contexte. |

## Ordre de correction retenu

Le prochain commit de code doit traiter uniquement une tranche P0 vérifiable : validation défensive de l’endpoint WebSocket interne, ou confinement runtime/download si la revue confirme une entrée contrôlable. Chaque correction doit inclure des tests négatifs, un scan Gosec frais et une comparaison du nombre de findings sans allowlist globale. Les autres findings restent visibles et ouverts.

## Verdict intermédiaire

`GOSEC_REVIEW_CLASSIFIED_NOT_CLOSED`. Le gate Gosec demeure ouvert ; la classification ne constitue ni un PASS de sécurité ni une déclaration de production readiness.
