# Politique permanente — BASELINE_DISCOVERY et conservation canonique

## Objet

Cette politique s’applique à tout lot ForgeLocal ouvert après T22. Elle empêche de commencer un correctif ou une fonctionnalité depuis un répertoire de travail non qualifié, ou d’annoncer une conservation sans preuve transportable.

## Précondition obligatoire avant écriture de code

Chaque lot doit inclure une section nommée exactement **`BASELINE_DISCOVERY`** dans son addendum ou son rapport. La section doit être rédigée avant la première modification de code et contenir les éléments suivants.

| Élément | Exigence |
|---|---|
| Horodatage | `started_utc` et `completed_utc` en UTC |
| Répertoires | `cwd`, chemins de recherche et chemin de baseline retenue |
| Commandes | Commandes complètes, sans ellipses |
| Résultats | Sorties brutes et `exit_code` de chaque contrôle |
| Références Git | Commit, tag annoté résolu, `git status --short` et `git fsck --full` |
| Conservation | Bundle, sidecar SHA-256, `git bundle verify`, clone neuf et HEAD obtenu |
| Décision | Référence exacte autorisée pour le worktree ; ou `BLOCKED_MISSING_BASELINE` documenté |

La recherche doit couvrir le workspace actif, `/home/ubuntu/upload`, les kits extraits, les dossiers de travail historiques, les dépôts Git locaux et les répertoires de livraison accessibles. L’absence ne peut être déclarée qu’après cette recherche consignée.

## Conservation de clôture

Chaque lot qualifié produit un bundle Git, un sidecar portable, un kit ZIP, un sidecar portable et un manifeste interne non auto-référentiel. Les sidecars externes doivent porter uniquement le nom du fichier, jamais un chemin absolu. La clôture exige la vérification du ZIP, du manifeste après extraction neuve, du bundle dans un clone neuf, ainsi que le re-scan du kit extrait.

> Aucun statut de clôture ne peut être annoncé avant que le kit distribué et ses sidecars soient cohérents avec la copie canonique hashée.

## Invariants

Cette politique n’autorise aucun lot produit, runtime, proxy réel, Camoufox, SystemVault natif ou release. Les gates existants demeurent la source de vérité.
