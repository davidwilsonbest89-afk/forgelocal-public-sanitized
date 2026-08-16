# T07-R — Retour de revue indépendante redacted

**Destinataire :** `@hajarbenmlih91-cloud`  
**Objet :** confirmation indépendante de complétude d’une attestation de provenance Camoflox  
**Portée :** contrôle redacted de cohérence et de références consultables sous contrôle d’accès ; ce document n’autorise ni portage, ni intégration, ni exécution, ni T08.

## Mandat de revue

Merci de vérifier les six domaines ci-dessous à partir des références que le détenteur des droits vous donne de manière contrôlée. Ne retournez ni archive, ni code, ni contenu de licence privée, ni valeur de l’alerte. Si une référence manque, n’est pas consultable ou ne prouve pas le domaine indiqué, laissez le booléen correspondant à `false` et retournez `incomplete`.

| Domaine | Ce que vous vérifiez | Booléen JSON à mettre à `true` seulement si vérifié |
|---|---|---|
| Révision ou snapshot | L’identifiant pointe vers un snapshot privé précis et contrôlé | `revision_or_snapshot` |
| Hash de l’archive candidate | Le hash attesté correspond à `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` | `candidate_archive_sha256` |
| Portée des droits | `internal_use=yes`, `modification=yes` et la redistribution est explicite | `rights_scope` |
| Licence ou accord | Une référence de licence ou accord est réellement accessible à une relectrice autorisée | `license_or_agreement` |
| Notices | Les obligations et notices tierces applicables sont référencées et accessibles | `notices` |
| Triage sécurité | La localisation redacted et les deux décisions sont revues sans exposer la valeur | `security_triage` |

## Réponse à reporter dans le JSON

Renseignez une référence de revue non secrète mais traçable, puis les six booléens JSON réels. Les chaînes telles que `"true"` ne sont pas valides : il faut employer `true` sans guillemets.

```json
"independent_review": {
  "reference": "<votre-reference-de-revue-redacted>",
  "coverage": {
    "revision_or_snapshot": false,
    "candidate_archive_sha256": false,
    "rights_scope": false,
    "license_or_agreement": false,
    "notices": false,
    "security_triage": false
  }
}
```

Pour le triage, inscrivez votre décision dans `independent_reviewer_decision` avec l’une des trois valeurs exactes : `REAL_SECRET`, `FALSE_POSITIVE` ou `UNKNOWN`. Elle doit être identique à la décision du détenteur des droits pour être concordante. Une divergence, une absence ou une valeur hors liste équivaut à `UNKNOWN` opérationnel et conserve le blocage.

| Retour possible | Signification | Effet |
|---|---|---|
| Toutes les couvertures `true` et triage concordant | La forme redacted est complète | Peut être soumise au contrôle de complétude, sans `PASS` automatique |
| Au moins une couverture `false` | Preuve insuffisante ou non consultable | T07 reste bloqué |
| Triage divergent ou absent | Conclusion non indépendante ou insuffisante | `UNKNOWN` opérationnel ; T07 reste bloqué |
| Redistribution `not_granted` | Droit de distribution non accordé | Audit passif possible ; distribution future bloquée |

> Votre retour doit rester une confirmation de portée de revue. Il ne doit pas être formulé comme une validation de code, de machine, de runtime, de release ou de mise en production.
