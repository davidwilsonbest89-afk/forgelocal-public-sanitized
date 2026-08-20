# T24 — Opérations bulk bornées : contrat Core local et matrice de tests

## Statut et autorisation

| Champ | Valeur |
|---|---|
| Jalon | T24 — Opérations bulk bornées |
| Autorisation | Instruction explicite du propriétaire reçue le 2026-08-20 : démarrer T24 avec les tests nécessaires. |
| Baseline qualifiée | `df969533cdd446be41d868bebea2c8a106f543d5` |
| Tag de baseline | `t23-scope-and-evidence-correction-2026-08-19` |
| Bundle de baseline | `forgelocal-core-t23-scope-evidence-df96953.bundle` |
| Journal brut | `/home/ubuntu/forgelocal-t24-baseline-discovery/BASELINE_DISCOVERY_RAW.log` |
| Décision de départ | Baseline T23 qualifiée par bundle, clone neuf, tag, `git fsck --full` et sidecar corrigé dans le journal. |

## BASELINE_DISCOVERY

La découverte a commencé à `2026-08-20T01:41:13Z`. Elle a recherché les bundles, ZIP et sidecars dans le workspace, `/home/ubuntu/upload`, les copies canoniques et les livraisons T00–T23. Le bundle retenu a passé `git bundle verify`; un clone neuf a été créé, checkouté au tag T23, puis vérifié avec `git rev-parse HEAD`, `git status --short` et `git fsck --full`. Le premier contrôle de sidecar avait un mauvais répertoire courant, sans remettre en cause le bundle ; il a été rejoué depuis le répertoire du sidecar à `2026-08-20T01:41:25Z` avec `sha256sum -c` et résultat `OK`.

Les commandes, chemins, dates UTC, codes de sortie et sorties brutes sont conservés dans le journal brut ci-dessus. Aucun code T24 n’a été écrit avant cette qualification.

## Périmètre fermé

T24 expose une unique opération de mutation bulk Core local, authentifiée et limitée au loopback. Les actions admises sont :

| Action | Effet par profil | Sémantique idempotente |
|---|---|---|
| `archive` | Passage `active → archived` via le store T23 | Un profil déjà archivé renvoie `noop`. |
| `reopen` | Passage `archived → active` via le store T23 | Un profil déjà actif est refusé par la règle de lifecycle existante. |
| `add_tag` | Ajout du tag validé par le store | Un tag déjà présent est un `noop` bulk explicite. |
| `remove_tag` | Retrait du tag validé par le store | Un tag absent est un `noop` bulk explicite. |
| `set_group` | Association à un groupe local existant | Une association déjà identique est un `noop`. |
| `clear_group` | Retrait d’association de groupe | Un profil sans groupe est un `noop`. |

Les actions sont appliquées **séquentiellement dans l’ordre exact de `profile_ids`**, avec un maximum strict de `50` cibles distinctes. L’absence de transaction globale est volontaire : chaque profil possède son verrou, sa mutation durable, sa capture History et son audit individuel. Le résultat est donc un `partial success` explicite et non une réussite atomique fictive.

## Contrat API

```text
POST /api/profiles/bulk
Authorization: Bearer <mémoire locale>
Origine : loopback autorisée seulement

{
  "operation": "archive|reopen|add_tag|remove_tag|set_group|clear_group",
  "profile_ids": ["id-1", "id-2"],
  "tag": "requis pour add_tag/remove_tag",
  "group": "requis pour set_group"
}
```

La requête est refusée avant toute mutation si elle dépasse la taille admise, contient des identifiants vides ou dupliqués, une action inconnue, un champ hors contrat, un tag invalide ou un groupe absent. `set_group` ne peut viser qu’un groupe connu du `groupStore`. Les cibles archivées ou quarantained restent soumises aux guards individuels du store.

La réponse réussie porte le `correlation_id` de requête, un résumé et une entrée par cible. Chaque entrée ne contient que `id`, `status` (`changed`, `noop`, `failed`), un code machine-readable redacted et l’état/l’attribut non sensible résultant. Aucun secret, référence de coffre, URL proxy, chemin local, contenu de note ou champ personnalisé n’est renvoyé ou audité.

## Concurrence, annulation, reprise et audit

1. Chaque cible utilise `WithHistorySequence` et les primitives de verrouillage du store existant ; une cible verrouillée retourne `PROFILE_LOCKED` sans bloquer les cibles suivantes.
2. Une même requête T24 ne traite pas deux fois un identifiant, et deux requêtes concurrentes restent sérialisées par profil.
3. La requête vérifie `r.Context().Err()` avant chaque cible. Si elle est annulée, les mutations déjà terminées sont retournées comme telles dans l’exécution encore ouverte ; aucune nouvelle cible n’est engagée.
4. Le timeout de verrou par profil existant est conservé. T24 ne crée pas de runtime, session, proxy ou processus.
5. Toute mutation réussie capture History avec une action bulk spécifique. Toute erreur et tout résultat sont audités en forme redacted : action, id, résultat, code, ordre et `correlation_id`, jamais les valeurs métier sensibles.

## Exclusions absolues

T24 n’implémente pas de Dashboard, export multiple, import/export de cookies, extension, runtime, lancement, proxy réel, test de proxy réseau, Camoufox, SystemVault natif, cloud, RBAC, secret réel, backup, migration globale de profil ou release. Il ne lève aucun invariant de produit.

## Matrice de tests obligatoire

| ID | Contrôle attendu |
|---|---|
| T24-UNIT-01 | Archive/réouverture bulk : transitions, no-op, ordre déterministe et résumé exact. |
| T24-UNIT-02 | Ajout/retrait bulk de tag : validation, no-op et absence de double tag. |
| T24-UNIT-03 | Définition/retrait bulk de groupe : groupe existant requis, no-op et refus lifecycle. |
| T24-UNIT-04 | Requête invalide ou >50 cibles : refus avant toute écriture. |
| T24-UNIT-05 | Une cible inconnue, verrouillée ou quarantined échoue localement sans annuler les cibles valides suivantes. |
| T24-UNIT-06 | Deux requêtes concurrentes sur les mêmes cibles : absence de data race, absence de double History, état final cohérent. |
| T24-UNIT-07 | Annulation de contexte avant cible suivante : aucune nouvelle mutation n’est engagée. |
| T24-API-01 | Route authentifiée loopback : `401` sans Bearer, `403` hors loopback, CORS/Origin refusé, body borné et unknown fields refusés. |
| T24-API-02 | Réponses/audits redacted : aucune référence de secret, URL proxy, chemin, note ou custom field. |
| T24-QUAL-01 | `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check`. |
| T24-QUAL-02 | Gitleaks du delta exact, Gosec base→head, Govulncheck, OSV-Scanner, Trivy filesystem et SBOM Syft. |
| T24-QUAL-03 | Bundle Git, sidecar, clone neuf, `git fsck --full`, ZIP, manifeste non auto-référentiel et re-scan d’extraction. |

## Conditions de sortie

T24 ne pourra être présenté que comme `IMPLEMENTED_PENDING_INDEPENDENT_REVIEW` tant que le bundle, le kit, les sidecars, la copie canonique et les preuves brutes n’auront pas été produits et vérifiés. Aucun T25 ne commence automatiquement.
