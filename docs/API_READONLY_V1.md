# Contrat API Core v1 — Lecture seule redacted

Les endpoints ci-dessous sont uniquement des lectures. Ils exigent le Bearer token local du Core, retournent `X-Request-ID` et n’exposent ni secret, ni valeur proxy, ni chemin, ni empreinte complète.

| Endpoint | Réponse | Règles |
|---|---|---|
| `GET /api/v1/readonly/health` | état et version Core | aucune donnée métier |
| `GET /api/v1/readonly/summary` | compteurs profils, groupes, runtimes | aucun détail sensible |
| `GET /api/v1/readonly/profiles?limit=&cursor=` | DTO profil redacted | curseur opaque, limite 1–100 |
| `GET /api/v1/readonly/groups?limit=&cursor=` | DTO groupe redacted | jamais de host, port ou référence proxy |
| `GET /api/v1/readonly/runtimes` | descripteurs sûrs | chemin binaire et capacités détaillées omis ; Camoufox `candidate=true`, `launchable=false` |

Le token est détenu seulement en mémoire par le client local. Il ne doit jamais apparaître dans le DOM, le stockage navigateur, les logs, les analytics ou les paramètres d’URL. Le bootstrap du token demeure un contrat Core à définir et doit être court, local et lié à l’origine loopback autorisée.

`X-Request-ID` sert exclusivement à relier une requête technique à ses logs assainis. Les futures mutations Core créeront leur propre `correlation_id` d’audit métier ; elles ne sont pas incluses dans cette version.
