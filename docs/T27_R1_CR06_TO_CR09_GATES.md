# T27-R1 — Gates CR-06 à CR-09

## CR-06 — Machine d’état externe, sans écriture de store

| État | Entrée autorisée | Sortie autorisée | Effet interdit |
|---|---|---|---|
| `DISCOVERED` | baseline vérifiée | `QUALIFIED` ou `REJECTED` | écriture Profile, runtime ou secret |
| `QUALIFIED` | preuves locales complètes | `PENDING_AUTHORIZATION` | lancement de runtime |
| `PENDING_AUTHORIZATION` | décision produit explicite | `AUTHORIZED` ou `REJECTED` | activation implicite |
| `AUTHORIZED` | autorisation écrite et environnement admis | lot dédié | release automatique |
| `REJECTED` | refus de policy, scan ou preuve | terminal | contournement utilisateur |

Cette machine est documentaire. Elle ne devient jamais un état persistant dans Profile, SQLite métier ou configuration utilisateur dans T27-R1.

## CR-07 — Autorisation écrite distincte

Toute exécution d’automation, runtime, proxy réel ou fournisseur réel reste interdite. Une autorisation valable doit identifier le périmètre, l’environnement, le propriétaire, les données admises, la durée, le plan de révocation et les critères de preuve. Aucune autorisation n’est reçue dans ce lot : **`CR07_BLOCKED_MISSING_WRITTEN_AUTHORIZATION`**.

## CR-08 — Validation native

`NATIVE_SYSTEMVAULT_NOT_TESTED` reste actif. Aucun environnement natif qualifié et aucun coffre utilisateur n’est accessible dans cette sandbox : **`CR08_NOT_TESTED_NATIVE_ENVIRONMENT_REQUIRED`**.

## CR-09 — CI, provenance et publication

La CI cible devra exécuter, dans cet ordre, toolchain lock, test `-race`, vet, build, test Dashboard, audit pnpm, Gitleaks, scans de dépendances, SBOM, contrôle des notices et paquet de conservation. Les résultats doivent être annexés à une provenance immuable et signée avant toute décision de publication.

T27-R1 ne configure ni secret CI, ni publication, ni release. `PUBLIC_RELEASE_BLOCKED` et `SCAN_BLOCKED_UNKNOWN` restent actifs.
