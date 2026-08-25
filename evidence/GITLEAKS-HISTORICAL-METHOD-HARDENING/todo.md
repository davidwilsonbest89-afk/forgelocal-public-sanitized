# TODO — GITLEAKS-HISTORICAL-METHOD-HARDENING

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Appliquer la procédure à toute nouvelle plage | rev-list non vide, arbre et JSON pour chaque commit | Sécurité dépôt |
| Haute | Vérifier les configurations et hashes | fichier de configuration exact archivé et checksum validé | Responsable preuves |
| Haute | Interdire les faux PASS | toute plage vide ou incomplète marquée INVALID/BLOCKED | Gouvernance qualité |
| Moyenne | Rejouer sur clone neuf | extraction arbre-par-arbre et synthèse identique | Revue indépendante |

Aucun scan zéro-commit ne doit être présenté comme une absence de secret ou comme une qualification historique.
