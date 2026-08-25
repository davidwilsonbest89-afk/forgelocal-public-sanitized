# TODO — POST-V6-ROADMAP-FINAL-MATRIX

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Revue indépendante de tous les lots | acceptation des bundles, ZIP, sidecars, manifestes, clones et statuts | Gouvernance qualité |
| Haute | Validation Docker native | Docker/Buildx disponible, build non-runtime et scans image archivés | Plateforme |
| Haute | Revue licences | décisions officielles pour chaque UNKNOWN avant distribution | Legal / OSS |
| Haute | Revue OSV | scanner compatible Go 1.25.13 ou preuve effective par advisory | Sécurité dépendances |
| Moyenne | Revue LFS durable | réplication canonique et fsck zéro sur environnement indépendant | Administrateur dépôt |
| Moyenne | Lots futurs Staticcheck/GolangCI-Lint | corrections isolées avec tests de non-régression | Mainteneurs Go |

Aucune fonctionnalité T28/T29/T39/T40/T41/T42, runtime de production, release, Camoufox, proxy réel, cookie, donnée utilisateur, SystemVault natif ou migration ne doit commencer sans nouvelle instruction produit explicite.
