# Revue ciblée Gosec — ForgeLocal operational-v1

## Verdict exact

```text
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
GOSEC_REVIEW_CLASSIFIED_NOT_CLOSED
FORGELOCAL_PRODUCTION_READY=false
```

Cette revue est une **revue automatisée/agent du code et des chemins d’exécution**, et non une revue humaine indépendante. Aucun finding Gosec n’est déclaré fermé, faux positif, accepté ou masqué par une suppression globale.

## Périmètre et baseline

La baseline a été prise sur le HEAD publié `e8581550b226dfcb60eb0c6b74b004901148a0e8`. Le correctif source est publié dans `ca00a0ffaf70c667e0583faafc3e477885a4c67f`, vérifié sur `origin/validation/operational-v1`. Le scan faisant autorité est :

```text
gosec -fmt json -out evidence/SECURITY_GOSEC_REVIEW/gosec_source_only.json ./cmd/... ./internal/    # exit_code=1, 177 findings
```

Le glob `./...` n’est pas utilisé pour les conclusions Gosec et Govulncheck : le dépôt contient dans `artifacts/` des snippets Go portables et partiels qui ne constituent pas des packages compilables. Un scan source-only évite cette contamination sans retirer aucun fichier source du périmètre.

La classification générée est conservée dans `GOSEC_CLASSIFICATION.md`, la liste complète dans `GOSEC_FINDINGS.tsv`, et la revue d’exploitabilité dans `GOSEC_TRIAGE.md`.

## Revue d’exploitabilité

| Famille | Nombre | Disposition actuelle | Conclusion opérationnelle |
|---|---:|---|---|
| `G704` — destinations réseau variables | 7 | `NEEDS_MANUAL_REVIEW` | Le pont WebSocket et les helpers CLI doivent rester limités aux usages internes/local-first ; les contrôles applicatifs réduisent le chemin, mais le scanner reste ouvert. |
| `G703` — chemins issus de données variables | 14 | `NEEDS_MANUAL_REVIEW` | Runtime/download/restore : revue confinement, symlinks, traversal et provenance encore requise. |
| `G304`/`G305` — ouverture/archives avec chemin variable | 25 | `NEEDS_MANUAL_REVIEW` | Les chemins runtime/backup/download nécessitent des tests de traversal, symlink et extraction hors racine. |
| `G110` — décompression sans limite | 3 | `NEEDS_MANUAL_REVIEW` | Risque conditionnel de consommation de ressources ; aucune fermeture déclarée sans borne démontrée. |
| `G204` — subprocess variables | 6 | `NEEDS_MANUAL_REVIEW` | Distinguer les arguments contrôlés par l’opérateur local des entrées HTTP ; revoir absence de shell et allowlists. |
| `G122` — parcours filesystem | 1 | `NEEDS_MANUAL_REVIEW` | Vérifier le confinement de la racine et le comportement symlink. |
| `G115` — conversions entières | 8 | `NEEDS_MANUAL_REVIEW` | Vérifier bornes et impact sur identifiants, tailles et ports ; aucun finding supprimé. |
| `G301`/`G302`/`G306` — permissions | 39 | `NEEDS_MANUAL_REVIEW` | Revoir modes effectifs pour secrets, tokens, profils, artefacts et répertoires. |
| `G112` — timeout HTTP | 1 | `CONFIRMED_HARDENING_GAP` | Le serveur doit ajouter des timeouts de lecture adaptés ; le finding reste ouvert dans ce lot. |
| `G404` — pseudo-aléatoire | 17 | `NEEDS_MANUAL_REVIEW` | Peut être acceptable pour humanisation/fingerprint non cryptographique ; interdit pour décisions de sécurité. |
| `G104` — erreurs ignorées | 53 | `NEEDS_MANUAL_REVIEW` | Prioriser les chemins flush/close/delete/audit ; ne pas traiter globalement comme bénin. |
| `G101` — motif secret | 1 | `NEEDS_MANUAL_REVIEW` | Vérifier le contexte de `.api-token.meta`; aucune valeur secrète réelle n’a été ajoutée. |

Les dispositions ci-dessus sont des états de triage, pas des clôtures de sécurité. Les findings historiques ou hors du chemin Dashboard/proxy courant restent référencés et ouverts jusqu’à une revue dédiée.

## Correctif ciblé publié dans le worktree

Le pont `playwrightWSProxy` validait auparavant uniquement la syntaxe URL puis construisait une cible TCP loopback à partir du port. Le worktree ajoute maintenant une validation défensive avant tout dial : schéma `ws`, hôte loopback, port 1–65535, absence d’utilisateur/query/fragment, chemin `/api/playwright/ws/{session_id}`, puis normalisation du dial vers `127.0.0.1`.

Régression ajoutée : `internal/api/playwright_ws_security_test.go`. Elle couvre IPv4 loopback, IPv6 loopback et `localhost`, ainsi que hôte externe, IP externe, mauvais schéma, port absent/invalide, query, chemin différent et userinfo.

Le finding G704 est volontairement **toujours visible** : le code applicatif ne prétend pas que le scanner a disparu. Baseline/post-correctif sur le code du commit `ca00a0ffaf70c667e0583faafc3e477885a4c67f` : **177 → 177**, `new_findings=0`, aucune suppression globale.

## Tests et gates exécutés

| Gate | Résultat | Preuve |
|---|---|---|
| Tests ciblés API + test endpoint interne | `PASS` | `GOSEC_HARDENING_FOCUSED_TESTS_RAW.log` — exit 0 |
| Syntaxe runner Playwright externe | `PASS` | même log — `node --check` exit 0 |
| `go test ./cmd/... ./internal/...` | `PASS` | `SOURCE_GATES_RAW.log` |
| `go vet ./cmd/... ./internal/...` | `PASS` | `SOURCE_GATES_RAW.log` |
| `go build ./cmd/... ./internal/...` | `PASS` | `SOURCE_GATES_RAW.log` |
| `pnpm run check` Dashboard | `PASS` | `FULL_VERIFICATION_RAW.log` — exit 0 |
| Gitleaks | `PASS` | `FULL_VERIFICATION_RAW.log` — 0 leaks |
| Govulncheck source-only | `PASS` | `SOURCE_GATES_RAW.log` — no vulnerabilities |
| Gosec source-only post-correctif | `FAIL_GATE_UNCHANGED` | `GOSEC_POST_FIX_RAW.log`, `gosec_source_only_post_fix.json` — exit 1, 177 findings |
| `go test ./...`, `go vet ./...`, `go build ./...`, Govulncheck `./...` | `NOT_APPLICABLE_AS_GLOBAL_GATE` | échec de compilation dans les snippets partiels `artifacts/.../source/internal/api`; les gates source-only passent |

## Runner Playwright externe

Le runner reste **un artefact externe non publié** et n’est pas inclus dans `ca00a0ffaf70c667e0583faafc3e477885a4c67f`. Il remplace les attentes fixes de 3 000 ms et 1 500 ms par un polling borné de `/api/sessions` filtré par `profile_id`, avec événements de diagnostic minimaux par tentative. Il conserve aussi une assertion explicite `external_forward_observed === false`. Aucun credential de fixture n’est écrit dans les preuves du dépôt.

Le runner utilise uniquement Chromium système autorisé pour le smoke local. Cela ne valide ni Camoufox, ni SystemVault natif, ni Docker/Buildx, ni un proxy ou cookie réel.

## Statut fonctionnel conservé

Les preuves Dashboard déjà publiées restent l’autorité fonctionnelle :

```text
SMOKE_DASHBOARD_PROFILE_CREATE_PASS
SMOKE_DASHBOARD_PROFILE_RESTART_PASS
SMOKE_PROXY_LOCAL_PASS
```

La revue actuelle n’a pas recommencé ces parcours ; elle a ajouté la protection ciblée et les régressions unitaires. Les gates environnementales et le gate Gosec restent ouverts. T28 n’est pas redémarré, T29 n’est pas commencé et T31–T38 ne sont pas touchés.

## Suite autorisée

La prochaine étape est la publication de ce lot et la poursuite des tests de contrat proxy. Les revues Camoufox, SystemVault natif, proxy/cookies de test et Docker/Buildx doivent être exécutées séparément, uniquement dans des environnements réels autorisés. Aucun statut de production ne peut être déduit de ce lot.
