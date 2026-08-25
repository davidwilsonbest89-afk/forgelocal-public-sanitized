# TODO — LFS-CONSERVATION-RESTORATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Répliquer le store LFS restauré | copie canonique vérifiée par OID et taille | Administrateur dépôt |
| Haute | Contrôler la provenance des douze objets | revue indépendante des commits porteurs et sources de récupération | Responsable preuves |
| Moyenne | Rejouer fsck sur une machine propre | `git lfs fsck=0` après clone et fetch ciblé documentés | Revue indépendante |
| Moyenne | Préserver l’interdiction de pull global | procédure opérateur filtrée par chemins et OID | Mainteneurs dépôt |

La restauration technique est considérée vérifiée, mais aucune action de release ou de runtime n’est autorisée par ce lot.
