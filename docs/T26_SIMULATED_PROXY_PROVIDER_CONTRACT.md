# T26 — Fournisseur Proxy simulé local

## Portée fermée

T26 fournit un catalogue local de fournisseurs **simulés** afin de tester le contrat de sélection de proxy sans communiquer avec un fournisseur, sans créer de socket sortante et sans affecter de proxy au runtime.

| Surface | Contrat |
|---|---|
| Enregistrement | `POST /api/proxy-providers` avec `id`, `name` et une référence exacte `provider.ref.<id>`. |
| Catalogue | `GET /api/proxy-providers` ; projection sans valeur secrète. |
| Résolution simulée | `POST /api/proxy-providers/{id}/simulate-resolve` ; produit une proposition `*.provider.test` et ne contacte aucun réseau. |
| Identifiants | Références de secret uniquement ; clés, tokens, URLs de fournisseur, hôtes réels et mots de passe sont refusés par JSON strict. |
| Profils | La résolution exige un profil actif, mais ne modifie ni ProxyConfig, ni vault, ni runtime. |
| Gardes | Bearer, loopback, Origin/Referer local, JSON strict, taille de requête et audit redacted. |

Les fournisseurs externes, requêtes HTTP sortantes, proxies réels, stockage de secrets, Dashboard et release sont hors périmètre. `mode` est toujours `simulated` dans T26.
