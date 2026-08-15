# T07-R — Demande de triage sécurité redacted

**Ticket :** `T07-CAMOFLOX-GEN-API-001`
**Emplacement :** `tests/smoke.test.js:24`
**État initial obligatoire :** `UNKNOWN / BLOCKED`

## Objet

Une alerte Gitleaks `generic-api-key` a été détectée lors de l’examen passif d’un candidat privé. La valeur n’est pas incluse dans ce document, dans Git, dans les logs, dans les exports ni dans les preuves T07.

Le mainteneur désigné et une relectrice indépendante doivent examiner la preuve dans l’espace privé contrôlé, puis consigner chacun une décision redacted : `REAL_SECRET`, `FALSE_POSITIVE` ou `UNKNOWN`.

| Décision conjointe attendue | Action avant toute nouvelle preuve T07 | Effet |
|---|---|---|
| `REAL_SECRET` | Révocation ou rotation confirmée ; nouveau snapshot ; nouveau hash ; rescan redacted | Le blocage reste en place jusqu’à preuve complète |
| `FALSE_POSITIVE` | Justification redacted ; nouveau snapshot candidat redacted, hashé et rescané | Une nouvelle revue T07 reste obligatoire |
| `UNKNOWN` | Aucune action de portage ; conserver l’état fail-closed | T07 et T08 restent bloqués |

Une divergence entre mainteneur et relectrice indépendante conserve automatiquement le résultat `UNKNOWN`. Aucune capture, valeur brute, token, clé ou contenu source n’est accepté comme élément Git versionné.
