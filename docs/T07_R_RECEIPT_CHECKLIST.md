# T07-R — Checklist de réception et de complétude

Cette checklist sert exclusivement à déterminer si une attestation **redacted** peut être soumise à une revue indépendante. Elle ne constitue pas une décision `PASS`, n’autorise aucun portage et ne modifie pas le registre canonique.

| Contrôle obligatoire | Critère de refus | Résultat à consigner |
|---|---|---|
| Révision privée | Identifiant, type ou hash de snapshot absent ; valeur de modèle non remplacée | `received` ou `missing` |
| Lien candidat | `attested_candidate_archive_sha256` différent de `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` | `match` ou `mismatch` |
| Revue privée | Référence de revue autorisée absente | `received` ou `missing` |
| Droits | Détenteur, usage interne ou modification absents | `complete` ou `incomplete` |
| Redistribution | Valeur différente de `granted` ou `not_granted` | `explicit` ou `implicit_or_invalid` |
| Licence et notices | Une référence vérifiable par relecteur autorisé est absente | `complete` ou `incomplete` |
| Double triage | Décision absente, invalide ou divergente | `concordant`, `divergent` ou `missing` |
| `REAL_SECRET` | Référence de révocation/rotation absente | `remediated` ou `blocked` |
| `FALSE_POSITIVE` | Nouveau snapshot, hash et re-scan redacted absents | `rescan_ready` ou `blocked` |
| `UNKNOWN` | Toute tentative de déblocage | `blocked` |

Le validateur [`validate-t07-r-attestation.mjs`](../scripts/validate-t07-r-attestation.mjs) ne lit qu’un JSON redacted explicitement fourni. Il ne doit jamais recevoir de code source, archive candidate, clé, token, valeur d’alerte ou licence privée brute.

> Une attestation complète permet seulement une **revue indépendante de complétude**. T08 demeure interdit tant que `PROV-01`, `PROV-04` et `PROV-06` ne sont pas explicitement passés à `PASS` par cette revue.
