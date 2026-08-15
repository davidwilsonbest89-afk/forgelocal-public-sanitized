# T07-R — Demande de preuves au détenteur des droits

**Objet :** preuves de provenance privée et de licence nécessaires à la clôture T07 de ForgeLocal.

Le candidat reste non intégré et non exécutable. Cette demande ne sollicite ni code source en clair, ni archive, ni clé, ni token. Elle vise uniquement les références et attestations permettant une revue indépendante autorisée.

## Éléments demandés

| Priorité | Élément attendu | Format sûr accepté |
|---|---|---|
| P0 | Révision source vérifiable | Commit exact d’un dépôt privé contrôlé **ou** identifiant d’un snapshot immuable attesté, avec SHA-256 et date |
| P0 | Lien avec l’archive étudiée | Attestation que la révision ou le snapshot correspond à l’archive SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` |
| P0 | Propriété et droits | Identifiant du détenteur ou mandataire, portée d’usage interne, modification, redistribution explicitement `granted` ou `not_granted`, et obligations applicables |
| P0 | Licence / accord | Licence racine ou accord attesté ; référence privée autorisée si le document brut est confidentiel |
| P0 | Notices | Notices et obligations tierces des dépendances ou un emplacement privé contrôlé où les relecteurs autorisés peuvent les vérifier |
| P0 | Relecture autorisée | Référence d’une revue indépendante autorisée vérifiant explicitement le commit/snapshot, le hash de l’archive, la portée des droits, la licence ou l’accord, les notices et le triage |

## Alerte sécurité distincte

Une alerte redacted `generic-api-key` est localisée dans `tests/smoke.test.js:24`. Merci de **ne jamais transmettre la valeur** par message, Git, capture, log ou export. Le mainteneur et la relectrice indépendante doivent seulement retourner une décision redacted : `REAL_SECRET`, `FALSE_POSITIVE` ou `UNKNOWN`.

Si la décision est `REAL_SECRET`, une confirmation de révocation ou rotation est requise avant tout nouveau snapshot. Si elle est `FALSE_POSITIVE`, un nouveau snapshot candidat hashé et rescané doit être fourni. Si elle est `UNKNOWN`, le blocage est maintenu automatiquement.

## Livraison sûre

Les documents sensibles doivent rester dans un espace privé contrôlé. La réponse destinée au dépôt ForgeLocal doit se limiter aux identifiants non secrets, hashes, dates, statuts et références de revue. Le modèle [`t07-private-provenance-attestation.template.json`](t07-private-provenance-attestation.template.json) définit les champs attendus.

> Sans ces éléments, ForgeLocal conservera `T07_PROVENANCE_BLOCKED_PENDING_EVIDENCE` et T08 restera interdit.
