# T23 — Archivage et restauration logique des profils

## BASELINE_DISCOVERY

| Champ | Valeur |
|---|---|
| Début UTC | `2026-08-19T19:48:05Z` |
| Fin UTC | `2026-08-19T19:48:08Z` |
| Copie canonique | `/home/ubuntu/forgelocal-t22-canonical/` |
| Bundle qualifié | `forgelocal-core-t22-pending-operation-d428690.bundle` |
| SHA-256 bundle | `8496861f9c43cb31a29ab41c205305b4add94923af53ab83e2842121193b2524` |
| Tag résolu | `t22-pending-operation-correction-2026-08-19` → `d428690aabe1032ad0d3073e3f809d1df8dbfb29` |
| Clone de travail | `/home/ubuntu/forgelocal-t23-archive-restore-20260819T1938Z/clone` |
| Contrôles | Sidecars, `git bundle verify`, clone neuf, `git fsck --full` et worktree propre : PASS |
| Sortie brute | `/home/ubuntu/forgelocal-t23-archive-restore-20260819T1938Z/evidence/BASELINE_DISCOVERY.log` |

## Objet et limites

T23 rend la transition logique `active → archived → active` durable et explicite. Il ne restaure ni backup `.flbackup`, ni cookies, ni session navigateur, ni données de coffre, ni secret proxy. Il ne lance pas de runtime et ne modifie aucun gate de production.

## Invariants

| Invariant | Règle T23 |
|---|---|
| Persistance avant publication | Une transition est écrite dans `profile.json` avant le remplacement de l’objet mémoire du Store. Une erreur disque ne modifie donc pas l’état mémoire visible. |
| Archive explicite | Une transition réussie fixe `lifecycle_state=archived` et un horodatage `archived_at` UTC. |
| Restauration logique | Une réouverture valide remet `lifecycle_state=active` et efface `archived_at`; elle ne copie aucune donnée de session ou secret. |
| Idempotence archive | Archiver un profil déjà archivé est un no-op : aucune nouvelle version History ni audit de mutation ne sont produits. |
| Refus | Un profil archivé refuse les mutations ordinaires; un profil quarantined refuse archive/reopen selon le contrat de lifecycle existant. |
| Durabilité History | Toute transition effective utilise le marker T22 `operation_id`, la séquence verrouillée par profil et le clear conditionnel après commit SQLite. |
| Reprise | Une interruption après l’écriture Profile laisse le marker durable; le démarrage T22 réconcilie History puis efface uniquement le marker correspondant. |

## Critères de qualification

La matrice T23 doit couvrir : persistance et redémarrage, archive idempotent sans double History, échec d’écriture Profile sans état mémoire partiel, échec History après archive, réconciliation au redémarrage, redaction audit/API, Origin/Referer, et concurrence archive/reopen/mutation sous `-race`.

## Décision de périmètre

Le lot est strictement Core local. Les limites inchangées sont `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`.
