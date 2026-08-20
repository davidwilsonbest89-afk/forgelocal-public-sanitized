# Politique permanente — BASELINE_DISCOVERY

> **Règle obligatoire.** Aucun code, migration, configuration produit, test modifiant des données ni nouvelle fonctionnalité ne peut commencer avant la clôture documentée de `BASELINE_DISCOVERY` pour le lot concerné.

## Objet

Cette procédure empêche une conclusion erronée fondée sur le seul workspace actif. Elle impose la découverte puis la qualification de la lignée réellement disponible avant toute écriture. Elle s’applique à **chaque futur lot**, y compris les correctifs documentaires, les remédiations de dépendances et les reprises par un nouveau développeur.

## Recherche obligatoire

Avant de déclarer un commit, bundle, ZIP de preuve, sidecar ou baseline absent, la personne en charge doit consulter le registre de continuité, le README de handover, les rapports de lots et les décisions canoniques afin d’identifier : le commit, le tag, le hash, le nom exact de l’artefact et son dernier emplacement connu.

Elle doit ensuite rechercher l’artefact dans le workspace courant, `/home/ubuntu/upload`, les kits extraits, les répertoires de travail historiques, les copies canoniques et les dépôts Git locaux accessibles. Une absence dans le seul workspace actif ne constitue jamais une preuve d’absence.

## Qualification obligatoire lorsque l’artefact est trouvé

| Artefact | Vérifications minimales |
|---|---|
| ZIP ou archive | SHA-256 direct, sidecar disponible, `unzip -t`, extraction neuve, manifeste non auto-référentiel et re-scan Gitleaks. |
| Bundle Git | `git bundle verify` depuis un dépôt initialisé, clone neuf, checkout du tag attendu, comparaison de `HEAD`, `git status --short` et `git fsck --full`. |
| Commit ou tag | `git rev-parse`, `git show --no-patch --format=fuller` et comparaison avec le registre. |
| Baseline de travail | Clone ou worktree neuf créé seulement après les contrôles précédents ; aucun code ne précède cette étape. |

## Journal brut obligatoire

Chaque lot doit contenir une section ou un fichier `BASELINE_DISCOVERY_RAW.log` incluant les sorties **non réécrites**. Chaque commande doit préciser les six informations suivantes :

| Champ obligatoire | Exemple |
|---|---|
| `started_utc` | `2026-08-20T02:35:36Z` |
| `cwd` | Répertoire exact de lancement |
| `command` | Commande complète, arguments inclus |
| `path` | Chemin de l’artefact ou du clone contrôlé |
| `exit_code` | Code réel de sortie |
| `completed_utc` et sortie brute | Date de fin et contenu produit par la commande |

## Déclaration de blocage

`BLOCKED_MISSING_BASELINE` n’est autorisé qu’après la recherche ci-dessus, avec les commandes, chemins, dates UTC, codes de sortie et sorties brutes distribués avec le rapport. La décision doit nommer exactement le fichier, hash, tag ou commit manquant. Si l’artefact retrouvé correspond à une autre date, archive ou lignée, il doit être préservé et évalué séparément : il ne doit jamais être présenté comme la baseline attendue.

## Conservation après chaque lot

Après un lot, les preuves doivent être copiées dans un emplacement canonique, hachées, inscrites au registre et accompagnées d’un bundle, d’un kit, de sidecars et d’un manifeste. Le nettoyage d’espace ne peut supprimer que les répertoires temporaires déjà encapsulés dans une copie canonique vérifiable.

## Modèle minimal

```text
BASELINE_DISCOVERY
started_utc=<UTC>
lot=<identifiant>
expected_commit=<SHA>
expected_tag=<tag>
expected_artifact=<nom>

COMMAND: <commande complète>
cwd=<chemin>
path=<chemin vérifié>
<sortie brute>
exit_code=<code>
completed_utc=<UTC>

DECISION: BASELINE_QUALIFIED | BLOCKED_MISSING_BASELINE
```

Les gates de sécurité et de release existants restent applicables ; BASELINE_DISCOVERY les précède, mais ne les lève jamais.
