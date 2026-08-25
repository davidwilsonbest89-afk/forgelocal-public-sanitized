# ForgeLocal — Requalification intégrée profil → session → proxy

**Périmètre.** Cette requalification couvre exclusivement la jonction entre l’affectation proxy persistée d’un profil, `POST /api/sessions` et le lancement réel de Chromium. Elle n’évalue pas T28 historique, T29, les comptes réels, les cookies réels, Camoufox, Docker/Buildx, SystemVault natif ou une release de production.

> **Verdict exact : `SMOKE_INTEGRATED_PROXY_PASS`**

Le parcours intégré local a été rejoué après le correctif. Le Dashboard a relié le Core, a sélectionné le profil alpha, a créé et affecté l’entrée proxy alpha, puis le Core a créé la session Chromium. La navigation Core vers la destination synthétique a renvoyé HTTP 200 et la destination a observé `destination_via_proxy=1`. Après arrêt du proxy alpha, Chromium a échoué avec `ERR_PROXY_CONNECTION_FAILED` et le compteur de destination est resté inchangé, ce qui démontre l’absence de repli direct dans ce scénario.

## Résultats observés

| Contrôle | Résultat | Preuve redacted |
|---|---:|---|
| Dashboard → Core readonly/admin | PASS | événement `dashboard_admin_linked`, état `linked` |
| Affectation alpha persistée | PASS | `alpha_assignment_persisted.proxy_id_present=true` |
| `POST /api/sessions` alpha | PASS | `create_status=201` |
| Navigation Core alpha | PASS | `navigate_status=200` |
| Hop proxy observé par la destination | PASS | `destination_via_proxy=1` |
| Proxy alpha arrêté | PASS | `result=fail-closed`, `destination_unchanged=true`, erreur `ERR_PROXY_CONNECTION_FAILED` |
| Proxy nominal indépendant | PASS | HTTP 200, forwards proxy alpha observés |
| Mauvaise authentification indépendante | PASS | réponse rejetée, proxy synthetic bad-auth avec rejets, aucune destination atteinte |
| Isolation beta | PASS | HTTP 200, forwards beta observés, destination beta distincte |
| Révocation du token | PASS | révocation HTTP 204, sonde protégée HTTP 401, raison `revoked` |
| Profil sans affectation | PASS ciblé | comportement direct explicite couvert par test unitaire ; non inclus dans le parcours Core réseau V5 pour éviter toute geo externe |
| Valeurs credential dans le log authoritative | PASS | aucun littéral sensible détecté |
| Cleanup | PASS | tous les ports 3001, 19281 et 19282–19287 libres, aucun processus smoke résiduel |

Les compteurs redacted du smoke V5 sont : proxy alpha `accepted=5, rejected=0`, proxy beta `accepted=3, rejected=0`, proxy bad-auth `accepted=0, rejected=3`. Les valeurs d’identifiants et de mots de passe ne sont pas présentes dans le log authoritative.

## Correctif livré

Lors de `POST /api/sessions`, le Core lit maintenant l’identifiant d’affectation canonique, récupère et revalide l’entrée proxy, refuse les affectations inconnues ou incohérentes, résout une référence `proxy.ref.*` via le vault configuré uniquement en mémoire de lancement, puis transmet un `LaunchProxy` éphémère à Chromium. Une affectation déclarée mais non résoluble ne peut donc pas devenir silencieusement une session directe. Sans affectation de registre, le comportement existant direct/profil/groupe est conservé explicitement. L’override éphémère prend la précédence sur une politique de groupe uniquement pour ce lancement et ne peut pas être persisté par JSON.

Pour les fixtures loopback, la détection geo utilise le fallback de région local et n’appelle pas les fournisseurs geo externes. Chromium reçoit en outre `--proxy-bypass-list=<-loopback>` lorsqu’un proxy effectif est configuré, afin que la destination synthétique loopback ne soit pas contournée.

Les régressions unitaires couvrent alpha, beta, isolation, no-proxy, affectation inconnue, état proxy incohérent, secret indisponible, absence de fuite JSON et détection d’hôte loopback. Le test réseau bad-auth reste un test Chromium local indépendant : avec l’architecture de registre actuelle, un credential `secret_ref` non résolu est refusé par le Core en `PROXY_CREDENTIALS_UNAVAILABLE` avant lancement, plutôt que transformé en tentative directe.

## Gates et limites

| Gate | Statut |
|---|---|
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `git diff --check` | PASS |
| Gitleaks | PASS sur cette exécution, aucun leak détecté |
| Govulncheck | PASS outil disponible, aucune vulnérabilité signalée |
| Gosec | FAIL/GATE INCHANGÉ : 177 findings historiques dans le scan courant, non masqués |
| OSV, Docker/Buildx, Camoufox, SystemVault natif | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Catalogue runtime readonly Dashboard | `NOT_IMPLEMENTED_PLACEHOLDER` pour la création de profil Dashboard directe ; les profils V5 ont été préparés par Core comme précondition documentée |
| Production-ready | NON DÉCLARÉ |

Le smoke V5 a été exécuté le 25 août 2026 UTC sur Chromium système `/usr/bin/chromium`, avec Core et Dashboard loopback, proxy HTTP synthétique et destinations HTTP synthétiques. Aucun site externe n’a été requis par le smoke après le garde-fou loopback/geo.

## Fichiers de preuve

Le log brut authoritative est [`SMOKE_INTEGRATED_PROXY_FIX_AUTHORITATIVE_RAW.log`](./SMOKE_INTEGRATED_PROXY_FIX_AUTHORITATIVE_RAW.log). Le log de commandes et gates est [`GATES_POST_SMOKE_RAW.log`](./GATES_POST_SMOKE_RAW.log), et la preuve de cleanup est [`CLEANUP_V5_RAW.log`](./CLEANUP_V5_RAW.log). La baseline de reprise est [`BASELINE_DISCOVERY_RAW_POSTSTOP.log`](./BASELINE_DISCOVERY_RAW_POSTSTOP.log). Le log HTTP Core redacted est [`CORE_V5_HTTP_RAW_REDACTED.log`](./CORE_V5_HTTP_RAW_REDACTED.log).

Les preuves historiques du défaut initial sont conservées séparément dans `evidence/SMOKE_INTEGRATED_PROXY/` et ne sont pas réécrites.
