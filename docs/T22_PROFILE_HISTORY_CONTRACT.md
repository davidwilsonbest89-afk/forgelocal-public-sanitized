# T22 — Profile History : contrat et matrice de tests

**Statut :** autorisé pour implémentation locale.
**Baseline :** `ee08ffd6d1f997c2f6e0fa8ffbbe783c83af80b8` — tag `t21-test-evidence-correction-2026-08-19`.

## 1. Périmètre fermé

T22 fournit un historique local de la représentation persistante d’un profil, avec snapshots immuables, lecture paginée, diff déterministe et restauration logique protégée par concurrence. Les snapshots couvrent les métadonnées de profil qui sont déjà acceptées par le Core : identité, runtime, groupe, tags, cycle de vie, fingerprint, proxy **sans identifiants**, note et champs personnalisés.

Le lot n’inclut ni données navigateur, cookies, cache, IndexedDB, extensions, mots de passe, valeurs de coffre, runtime, session, proxy réseau réel, import/export global, Dashboard ni release.

## 2. Modèle et invariants

| Élément | Règle |
|---|---|
| Référentiel | SQLite dédié `profile_history.sqlite` sous `cfg.DataDir` ; aucune migration du Profile Store JSON ou de GAP-002 |
| Version | Une suite strictement croissante par `profile_id`, avec `version=1` à la création |
| Snapshot | JSON canonique immuable ; aucun mot de passe, identifiant proxy ou valeur de coffre |
| Captures | Création, modification générale, métadonnées, archive, réouverture et ajout/retrait de tag via les mutations API T22 couvertes |
| Lecture | Liste paginée et lecture de version accessibles seulement au client authentifié loopback |
| Diff | Compare deux versions du même profil et retourne les chemins modifiés ; les valeurs ne sont disponibles que dans les snapshots authentifiés, jamais dans audit/log/erreur/preuve |
| Restauration | Exige `expected_current_version`, refuse conflit ou profil archivé/quarantined, applique un snapshot validé sous lock par profil et crée une nouvelle version `restore` |
| Audit | Événements redacted `history_created`, `history_restored`, `history_restore_conflict`, `history_restore_refused` ; identifiant, version, action et correlation uniquement |
| Écriture interdite | `GET` list/version/diff n’écrit ni Profile JSON ni audit ; une restauration valide écrit uniquement le profil et l’historique associé |

## 3. API T22

| Route | Résultat |
|---|---|
| `GET /api/profiles/{id}/history?limit=&offset=` | Catalogue paginé des versions, sans valeurs métier |
| `GET /api/profiles/{id}/history/{version}` | Snapshot authentifié, valeurs non secrètes seulement |
| `GET /api/profiles/{id}/history/diff?from=&to=` | Diff de mêmes séries, chemins modifiés et versions |
| `POST /api/profiles/{id}/history/{version}/restore` | Restauration avec `expected_current_version`, `correlation_id` et conflit fail-closed |

Toutes les routes restent sous Bearer admin, loopback et garde Origin/Referer existants. Les routes de lecture ne génèrent pas d’audit.

## 4. Matrice minimale de tests

| Domaine | Cas exigés |
|---|---|
| Immutabilité | Création puis mutation : les versions antérieures restent inchangées, l’ordre est strict et concurrent sûr |
| Redaction | Snapshot et diff n’exposent jamais `username`, `password`, valeur de coffre ou token ; audit/log ne contiennent jamais note/champ personnalisé |
| Lecture | Pagination, version inconnue, profil inconnu, lecture sans écriture Profile/audit |
| Diff | Diff identique vide, diff de versions de même série, version inexistante, paramètres invalides |
| Restauration | Succès, conflit optimistic-lock, version inconnue, profil archivé, persistence après recharge, nouvelle version `restore` |
| API | Auth absente, origine distante refusée, loopback autorisé, correlation header sur mutation, erreurs machine-readable |
| Qualité | `go test -count=1 -race ./...`, vet, build, diff check, Gitleaks du delta et Gosec base/head classifié |

## 5. Exclusions absolues

T22 ne change pas les gates de release, ne lance aucun navigateur, Camoufox ou proxy réel et ne qualifie pas SystemVault natif. Les statuts `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.
