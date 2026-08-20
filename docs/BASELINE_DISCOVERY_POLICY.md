# Politique permanente — BASELINE_DISCOVERY

> **Règle obligatoire.** Aucun code, migration, configuration produit, test modifiant des données ni nouvelle fonctionnalité ne peut commencer avant la clôture documentée de `BASELINE_DISCOVERY` pour le lot concerné.

## Objet

Cette procédure impose la découverte et la qualification de la lignée réellement disponible avant toute écriture. Elle s’applique à **chaque futur lot**, y compris les correctifs documentaires, les remédiations de dépendances et les reprises par un nouveau développeur.

## Recherche obligatoire

Avant de déclarer un commit, bundle, ZIP de preuve, sidecar ou baseline absent, consulter le registre de continuité, le README de handover, les rapports de lots et les décisions canoniques afin d’identifier le commit, le tag, le hash, le nom exact de l’artefact et son dernier emplacement connu.

Rechercher ensuite l’artefact dans le workspace courant, `/home/ubuntu/upload`, les kits extraits, les répertoires de travail historiques, les copies canoniques et les dépôts Git locaux accessibles. Une absence dans le seul workspace actif ne constitue jamais une preuve d’absence.

## Qualification obligatoire

| Artefact | Vérifications minimales |
|---|---|
| ZIP ou archive | SHA-256 direct, sidecar disponible, `unzip -t`, extraction neuve, manifeste non auto-référentiel et re-scan Gitleaks. |
| Bundle Git | `git bundle verify`, clone neuf, checkout du tag attendu, comparaison de `HEAD`, `git status --short` et `git fsck --full`. |
| Commit ou tag | `git rev-parse`, `git show --no-patch --format=fuller` et comparaison avec le registre. |
| Baseline de travail | Clone ou worktree neuf créé seulement après les contrôles précédents. |

## Journal brut obligatoire

Chaque lot doit contenir une section ou un fichier `BASELINE_DISCOVERY_RAW.log` avec les sorties non réécrites. Chaque commande doit indiquer les champs suivants :

| Champ | Exigence |
|---|---|
| `started_utc` | Début en UTC |
| `cwd` | Répertoire exact d’exécution |
| `command` | Commande complète et arguments |
| `path` | Chemin de l’artefact ou du clone contrôlé |
| `exit_code` | Code réel de sortie |
| `completed_utc` et sortie brute | Fin en UTC et résultat complet |

`BLOCKED_MISSING_BASELINE` n’est autorisé qu’après ces recherches documentées. La décision doit nommer exactement le fichier, le hash, le tag ou le commit manquant.

## Conservation

Après chaque lot, les preuves doivent être copiées dans un emplacement canonique, hachées, inscrites au registre et accompagnées d’un bundle, d’un kit, de sidecars et d’un manifeste. Le nettoyage d’espace ne peut supprimer que les répertoires temporaires déjà encapsulés dans une copie canonique vérifiable.

Les gates de sécurité et de release existants restent applicables ; BASELINE_DISCOVERY les précède, mais ne les lève jamais.
