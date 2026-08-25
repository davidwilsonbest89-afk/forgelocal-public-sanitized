# T00–T42 V6 — matrice individuelle Semgrep

La règle locale `go-math-rand-read` a produit **18 occurrences**. L’inspection de l’import réellement résolu montre que les 18 appels utilisent `crypto/rand`, et non `math/rand`. Aucun finding ne protège une session, un code ou un token avec un générateur pseudo-aléatoire.

| ID | Règle | Fichier | Ligne | Import réel | Usage | Impact | Décision | Test / condition de levée |
|---:|---|---|---:|---|---|---|---|---|
| 1 | go-math-rand-read | internal/api/correlation.go | 25 | crypto/rand | ID de corrélation non secret | Faible | Faux positif contextuel ; conserver | Test du package API et rescan Semgrep |
| 2 | go-math-rand-read | internal/api/readonly_session.go | 88 | crypto/rand | Secret de session de lecture seule | Élevé si faible aléa | Correct | Test du package API, race et rescan |
| 3 | go-math-rand-read | internal/api/router.go | 516 | crypto/rand | Seed de fingerprint | Moyen | Faux positif contextuel ; conserver | Tests router et rescan |
| 4 | go-math-rand-read | internal/api/router.go | 728 | crypto/rand | ID de requête | Faible | Faux positif contextuel ; conserver | Tests router et rescan |
| 5 | go-math-rand-read | internal/api/router.go | 814 | crypto/rand | Token API | Élevé | Correct | Tests token/API, race et rescan |
| 6 | go-math-rand-read | internal/backup/service.go | 369 | crypto/rand | Nonce AES-GCM | Élevé | Correct | Tests backup, race et rescan |
| 7 | go-math-rand-read | internal/backup/service.go | 462 | crypto/rand | ID d’artefact backup | Faible | Faux positif contextuel ; vérifier l’erreur non bloquante | Tests backup et rescan |
| 8 | go-math-rand-read | internal/launch/id.go | 25 | crypto/rand | ID de lancement avec fallback de test | Moyen | Faux positif contextuel ; ne pas remplacer par math/rand | Tests launch et rescan |
| 9 | go-math-rand-read | internal/localvault/localvault.go | 182 | crypto/rand | Sel de dérivation/coffre local | Élevé | Correct | Tests localvault, race et rescan |
| 10 | go-math-rand-read | internal/localvault/localvault.go | 294 | crypto/rand | Nonce AES-GCM | Élevé | Correct | Tests localvault, race et rescan |
| 11 | go-math-rand-read | internal/mcp/screenshot_artifacts.go | 267 | crypto/rand | Nom d’artefact temporaire | Faible | Faux positif contextuel ; conserver | Tests MCP et rescan |
| 12 | go-math-rand-read | internal/mcp/web_session_pool.go | 557 | crypto/rand | ID de session de recherche | Moyen | Correct | Tests MCP/session pool et rescan |
| 13 | go-math-rand-read | internal/profile/store.go | 1050 | crypto/rand | ID de profil | Moyen | Faux positif contextuel ; conserver | Tests profile et rescan |
| 14 | go-math-rand-read | internal/profile/store.go | 1085 | crypto/rand | ID de profil | Moyen | Faux positif contextuel ; conserver | Tests profile et rescan |
| 15 | go-math-rand-read | internal/profilemigration/migrator.go | 1048 | crypto/rand | ID d’opération migration | Moyen | Faux positif contextuel ; conserver | Tests migration et rescan |
| 16 | go-math-rand-read | internal/proxies/store.go | 477 | crypto/rand | Octets aléatoires injectables de test | Moyen | Correct ; wrapper explicitement crypto/rand | Tests proxies et rescan |
| 17 | go-math-rand-read | internal/secrets/keyring.go | 39 | crypto/rand | Clé de backup | Élevé | Correct | Tests secrets/keyring, race et rescan |
| 18 | go-math-rand-read | internal/templates/store.go | 513 | crypto/rand | ID de template | Moyen | Faux positif contextuel ; conserver | Tests templates et rescan |

**Conclusion.** Aucun des 18 findings n’est un usage de `math/rand`. Il n’y a donc pas de remplacement crypto à appliquer. La règle locale est trop large pour distinguer l’import qualifié ; elle reste utile comme contrôle de revue, mais ces occurrences ne doivent pas être présentées comme des usages non sûrs. Le rescan V6 doit conserver le JSON brut et cette qualification.
