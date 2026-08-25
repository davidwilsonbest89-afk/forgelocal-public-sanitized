# TODO — T30-REMOTE-EVIDENCE-RECONCILIATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Faire accepter la chaîne distante | revue indépendante du commit, tag, ZIP, bundle, sidecars, manifeste et clone | Responsable preuves |
| Haute | Conserver le statut en attente | `T30_REMOTE_EVIDENCE_RECONCILED_PENDING_INDEPENDENT_REVIEW` jusqu’à acceptation formelle | Gouvernance qualité |
| Moyenne | Vérifier la reproductibilité | clone neuf et checksums identiques dans un environnement indépendant | Revue indépendante |
| Permanente | Interdire l’exécution opérationnelle | aucun runtime T30, navigateur, proxy réel, cookie, donnée utilisateur, migration ou release | Mainteneurs produit |

La décision T30 reste documentaire et ne lève aucune gate V6.
