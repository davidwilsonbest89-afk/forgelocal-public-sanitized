# TODO — T10-T15-E2E-VALIDATION

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Rejouer le run sur environnement indépendant | 7 tests passés avec la même configuration loopback et logs redacted | Revue indépendante |
| Haute | Contrôler les preuves de cleanup | token, base, run root, descendants et ports absents après interruption et succès | Sécurité E2E |
| Moyenne | Vérifier la non-régression Dashboard | rerun T10/T15 après tout changement produit futur | Mainteneurs Dashboard |
| Permanente | Préserver les limites | aucune donnée réelle, Camoufox, proxy réel, cookie, migration, production ou release | Gouvernance qualité |

Le passage local reste `PENDING_INDEPENDENT_REVIEW` et ne lève aucune gate V6.
