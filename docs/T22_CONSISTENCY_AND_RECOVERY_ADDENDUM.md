# Addendum T22 — cohérence Profile JSON / History SQLite

**Statut :** autorisé pour le sous-lot `T22-CONSISTENCY-AND-TESTS`.  
**Baseline :** `5cfe7df3b5bb24c3d84ba455d3c32569555c4bdc`.

## BASELINE_DISCOVERY

**Date UTC :** `2026-08-19T18:44:47Z` à `2026-08-19T18:44:48Z`.  
**Répertoire qualifié :** `/home/ubuntu/forgelocal-t22-consistency-20260819T1835Z/clone`.  
**Bundle canonique consulté :** `/home/ubuntu/forgelocal-t22-delivery/forgelocal-core-t22-profile-history-5cfe7df.bundle` (`sha256` : `e90e21204f6741aa6d24c926be98385e23adf4873b7e8cbf07645013f460d3bd`).

| Commande | Code de sortie | Sortie déterminante |
|---|---:|---|
| `git rev-parse t22-profile-history-2026-08-19` | 0 | Objet tag annoté `48988552720b1461cf503dc3b4e4822b831a19ff` |
| `git rev-parse 't22-profile-history-2026-08-19^{}'` | 0 | Commit qualifié `5cfe7df3b5bb24c3d84ba455d3c32569555c4bdc` |
| `git fsck --full` | 0 | Aucune sortie, aucun objet corrompu |
| `sha256sum …/forgelocal-core-t22-profile-history-5cfe7df.bundle` | 0 | Hash ci-dessus |
| `git status --short` | 0 | Diff T22-CONSISTENCY explicite, avant commit |

Les commandes complètes, chemins, sorties brutes et horodatages sont conservés dans `evidence/t22-consistency/BASELINE_DISCOVERY.log`.

## Contrat de cohérence

T22 ne revendique pas une transaction ACID unique entre `profile.json` et SQLite. Chaque mutation persistante de Profile écrit d’abord un **marqueur pending durable**, owner-only, adjacent au profil. Ce marqueur porte un identifiant d’opération non secret, une action et une date. Il n’est jamais projeté par l’API ni copié dans un snapshot History.

Après l’écriture Profile, History tente une capture dans sa transaction SQLite. La capture réussie est suivie de la suppression conditionnelle du marqueur portant le même identifiant. Si History échoue, le profil demeure marqué pending et la réponse est un refus explicite ; l’état est alors récupérable, jamais silencieusement déclaré historisé.

Pour une restauration, Profile écrit le marqueur avant `profile.json`; History crée ensuite la version `restore` et son audit redacted. Si le commit SQLite ou l’effacement du marqueur échoue, le marqueur reste durablement présent.

## Reprise au démarrage

Le routeur réel, après l’ouverture du repository History, exécute la récupération sur les profils marqués pending :

1. si la dernière version History est égale au profil courant redacted, la reprise consigne un audit `history_recovery_confirmed` et efface le marqueur ;
2. sinon, elle crée une version immuable `recovery` et un audit `history_recovered`, puis efface le marqueur ;
3. si History est indisponible, le marqueur est conservé et le démarrage échoue de façon explicite plutôt que de supprimer l’évidence de divergence.

Ainsi, aucun profil marqué ne peut être considéré comme synchronisé avant qu’une version History et un audit de confirmation/récupération soient durables.

## Tests obligatoires

La correction doit démontrer : échec injecté de capture après écriture Profile, échec History après restauration, reprise après redémarrage, absence de fuite proxy, pagination réelle, lecture sans écriture Profile/audit Profile, guards complets et concurrence restauration/mutation. Aucun test ne doit qualifier une opération multi-store d’atomique ; il doit prouver la récupération déterministe.
