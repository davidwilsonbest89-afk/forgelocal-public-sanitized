# MAX LOCAL — Rapport final court

## Périmètre

Campagne exécutée dans un clone neuf du dépôt `davidwilsonbest89-afk/forgelocal-public-sanitized`, sur une branche locale dédiée `validation/max-local-execution` créée depuis le HEAD réel `9aba6ed109c54539b1e2fd8b083ec2c5c7e727e3`. La campagne a exécuté 249 tests ciblés, dont la suite, la race et vet. La branche corrective distante `validation/final-secret-remediation` n’a pas été remplacée. Aucun package historique n’a été reconstruit.

## Résultats exécutés

| Contrôle | Résultat | Statut |
|---|---|---|
| Lignée T07/T08 | Dépôt historique `boucheriechefimane-cmd/IPcache` non résolu par GitHub CLI ; commits et arbres non vérifiables | `HISTORICAL_EVIDENCE_UNVERIFIED` |
| Baseline code | `go test -race ./...`, `go vet ./...` et `go build ./...` échouent uniquement sur un arbre historique sous `artifacts/`; le code actuel ciblé passe | `BASELINE_FAIL_HISTORICAL_ARTIFACT_SCOPE` |
| Token anti-fuite | authentification synthétique et sentinelle absente | `PASS` |
| Authentification | 6/6 états : valid/missing/invalid/expired/revoked/0600, permissions et cleanup | `PASS` |
| Isolation profils | 5 contrôles ciblés et suite Go/race exécutés ; 249 tests ciblés au total | `PASS_DANS_LE_PÉRIMÈTRE_EXÉCUTÉ` |
| Proxy loopback | 8 scénarios ciblés : loopback, arrêté, port invalide, externe, redirection, timeout, query autorisée/non prévue | `PASS_DANS_LE_PÉRIMÈTRE_EXÉCUTÉ` |
| Core/Dashboard | 10 scénarios fonctionnels couverts par les tests ciblés ; i18n, typecheck et build dashboard | `PASS` |
| Docker host | daemon arrêté, socket absent, 7,5 Go au preflight ; aucun nouveau cycle image lancé par prudence | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Docker bridge | non exécuté dans cette campagne ; blocage bridge historique conservé | `DOCKER_BRIDGE_BUILD_BLOCKED_BY_ENVIRONMENT` |
| Trivy secrets | filesystem scan : 0 secrets | `PASS` |
| Grype Critical/High | source : 46/46 matches, dont 24 High/Critical ouverts ; image gate CI : 43/43 High/Critical non approuvées | `FAIL_BLOCKING` |
| Gosec | outil absent | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Semgrep | 46/46 findings listés individuellement ; 46 restent `OPEN_MANUAL_REVIEW_REQUIRED` après recalcul courant | `OPEN_MANUAL_REVIEW_REQUIRED` |
| ShellCheck/Yamllint | ShellCheck exit 0, Yamllint exit 0 | `PASS` |
| Chromium | job qualifié skipped ; aucun PASS natif local dans cette campagne | `BLOCKED_ENVIRONMENT_REQUIRED` |
| Firefox | binaire absent | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Camoufox | binaire absent, aucun téléchargement non vérifié | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| SystemVault natif | Secret Service/keyring natif indisponible | `NATIVE_SYSTEMVAULT_NOT_TESTED` |
| CI distante | run `33099867635` sur `9aba6ed` : test success, sécurité failure sur Grype image policy, navigateur skipped, conclusion globale failure | `CI_REMOTE_EXECUTION_FAIL` |
| Protection branche | API GitHub `404 Branch not protected` | `BRANCH_PROTECTION_ENFORCEMENT=NOT_VERIFIED` |

## Gouvernance

La matrice prioritaire existante reste ouverte : 453 lignes dédupliquées et 85 Critical/High, sans owner, échéance ou exploitabilité inventés. `SECRET_REAL_USE_STATUS=OWNER_CONFIRMATION_REQUIRED`. La revue indépendante n’est pas attestée. Le package historique reste inchangé et est seulement vérifié par ses références existantes.

> `PUBLIC_RELEASE_BLOCKED`
>
> `FORGELOCAL_PRODUCTION_READY=false`

## Verdict unique

`MAX_LOCAL_EXECUTION_VALIDATION_BLOCKED_ENVIRONMENT`
