# T07-R — Checklist de réception et de complétude

Cette checklist détermine seulement si une attestation **redacted** peut être soumise à une revue indépendante de complétude. Elle ne constitue pas une décision `PASS`, n’autorise aucun portage et ne modifie jamais le registre canonique.

| Contrôle obligatoire | Règle de refus | Résultat redacted à consigner |
|---|---|---|
| JSON d’attestation | JSON illisible, valeur de modèle non remplacée ou champ obligatoire absent | `valid_json` ou `invalid_json` |
| Révision privée | Identifiant, type, hash de snapshot ou référence de revue absents | `received` ou `missing` |
| Périmètre de revue indépendante | La revue ne couvre pas explicitement la révision/snapshot, le hash d’archive, les droits, la licence ou l’accord, les notices et le triage | `complete` ou `incomplete` |
| Lien candidat | Hash attesté différent de l’archive Camoflox étudiée | `match` ou `mismatch` |
| Droits de travail | `internal_use` ou `modification` différent de `yes` | `permitted` ou `blocked` |
| Redistribution | Valeur autre que `granted` ou `not_granted` | `explicit` ou `invalid` |
| Redistribution non accordée | `redistribution=not_granted` | Audit passif possible ; `future_distribution=blocked` |
| Licence et notices | Référence contrôlable par relecteur autorisé absente | `complete` ou `incomplete` |
| Double triage | Décision absente, hors ensemble autorisé ou différente | `UNKNOWN` opérationnel ; `blocked` |
| `REAL_SECRET` concordant | Référence rotation/révocation, nouveau snapshot, hash du nouveau snapshot ou re-scan absent | `blocked` |
| `FALSE_POSITIVE` concordant | Nouveau snapshot redacted, hash ou re-scan absent | `blocked` |
| `UNKNOWN` concordant | Toute tentative de déblocage | `blocked` |

Le validateur [`validate-t07-r-attestation.mjs`](../scripts/validate-t07-r-attestation.mjs) ne reçoit qu’un JSON redacted explicitement fourni dans l’emplacement privé. Il ne doit jamais recevoir de code source, archive candidate, clé, token, valeur d’alerte, licence privée brute ou preuve de droits brute.

> Une attestation complète permet seulement une **revue indépendante de complétude**. Toute divergence de triage ou donnée manquante laisse T07 bloqué. T08 demeure interdit tant que `PROV-01`, `PROV-04` et `PROV-06` ne sont pas explicitement passés à `PASS` par cette revue.
