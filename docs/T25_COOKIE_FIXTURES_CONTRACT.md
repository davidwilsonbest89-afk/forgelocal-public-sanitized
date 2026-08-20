# T25 — Import/export de fixtures cookie synthétiques

## Portée locale et sûre

T25 gère exclusivement des **fixtures QA synthétiques** associées à un profil local. Il ne lit, n’écrit, ne copie ni n’exporte les cookies d’un navigateur, d’une session, d’un profil Chromium, d’un fichier `Cookies`, d’un MCP ou d’un runtime.

Une valeur entrante doit porter le préfixe `fixture:` et est immédiatement remplacée par son SHA-256 dans le store. La valeur ne peut donc ni être persistée, ni revenir dans l’export, les audits, les logs, les erreurs ou les preuves.

| Surface | Contrat |
|---|---|
| Import | `POST /api/profiles/{id}/cookie-fixtures/import` ; remplacement atomique d’un manifeste complet validé. |
| Export | `GET /api/profiles/{id}/cookie-fixtures/export` ; projection redacted avec `value_digest`, jamais de valeur. |
| Domaine | Domaines de test uniquement : suffixe `.test`. |
| Limites | 20 fixtures, 32 KiB de corps, noms ≤ 64, domaines ≤ 128, chemins ≤ 128. |
| Gardes | Bearer, loopback, Origin/Referer local, JSON strict et profil actif. |
| Atomicité | Toute fixture invalide refuse le manifeste entier ; le dernier manifeste valide reste intact. |

Les cookies réels, les identifiants, l’accès à un navigateur, un proxy réel, les secrets, le Dashboard et toute release sont hors périmètre.
