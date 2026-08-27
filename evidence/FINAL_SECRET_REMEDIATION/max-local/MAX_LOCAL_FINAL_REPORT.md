# MAX LOCAL — Rapport final court

## Périmètre

Campagne exécutée dans un clone neuf du dépôt `davidwilsonbest89-afk/forgelocal-public-sanitized`, sur une branche locale dédiée `validation/max-local-execution` créée depuis le HEAD réel `f685c88e1aec4a58dc04a3ba2a4af5f1e5d9e52a`. La branche corrective distante `validation/final-secret-remediation` n’a pas été remplacée. Aucun package historique n’a été reconstruit.

## Résultats exécutés

| Contrôle | Résultat | Statut |
|---|---|---|
| Lignée T07/T08 | Dépôt historique `boucheriechefimane-cmd/IPcache` non résolu par GitHub CLI ; commits et arbres non vérifiables | `HISTORICAL_EVIDENCE_UNVERIFIED` |
| Baseline code | `go test -race ./...`, `go vet ./...` et `go build ./...` échouent uniquement sur un arbre historique sous `artifacts/`; le code actuel ciblé passe | `BASELINE_FAIL_HISTORICAL_ARTIFACT_SCOPE` |
| Token anti-fuite | authentification synthétique et sentinelle absente | `PASS` |
| Authentification | tests valid/missing/invalid/expired/revoked, permissions et cleanup | `PASS` |
| Isolation profils | suite Go actuelle, race et tests d’isolation | `PASS_DANS_LE_PÉRIMÈTRE_EXÉCUTÉ` |
| Proxy loopback | tests de validation loopback, refus externe et isolation proxy | `PASS_DANS_LE_PÉRIMÈTRE_EXÉCUTÉ` |
| Core/Dashboard | tests Go ciblés, i18n, typecheck et build dashboard | `PASS` |
| Docker host | daemon arrêté, socket absent, 3,8 Go disponibles ; nouveau cycle non exécuté | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Docker bridge | non exécuté dans cette campagne ; blocage bridge historique conservé | `DOCKER_BRIDGE_BUILD_BLOCKED_BY_ENVIRONMENT` |
| Trivy secrets | filesystem scan : 0 secrets | `PASS` |
| Grype Critical/High | source : 46 matches, dont 24 High/Critical ; image gate CI : 43 High/Critical non approuvées | `FAIL_BLOCKING` |
| Gosec | outil absent | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Semgrep | 46 findings, 42 ouverts et 4 mitigations scanner ouvertes | `OPEN_MANUAL_REVIEW_REQUIRED` |
| ShellCheck/Yamllint | ShellCheck exit 0, Yamllint exit 0 | `PASS` |
| Chromium | job qualifié skipped ; aucun PASS natif local dans cette campagne | `BLOCKED_ENVIRONMENT_REQUIRED` |
| Firefox | binaire absent | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Camoufox | binaire absent, aucun téléchargement non vérifié | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| SystemVault natif | Secret Service/keyring natif indisponible | `NATIVE_SYSTEMVAULT_NOT_TESTED` |
| CI distante | run `33094861561` sur `9a1c01c` : test success, sécurité failure sur Grype image policy, navigateur skipped, conclusion globale failure | `CI_REMOTE_EXECUTION_FAIL` |
| Protection branche | API GitHub `404 Branch not protected` | `BRANCH_PROTECTION_ENFORCEMENT=NOT_VERIFIED` |

## Gouvernance

La matrice prioritaire existante reste ouverte : 453 lignes dédupliquées et 85 Critical/High, sans owner, échéance ou exploitabilité inventés. `SECRET_REAL_USE_STATUS=OWNER_CONFIRMATION_REQUIRED`. La revue indépendante n’est pas attestée. Le package historique reste inchangé et est seulement vérifié par ses références existantes.

> `PUBLIC_RELEASE_BLOCKED`
>
> `FORGELOCAL_PRODUCTION_READY=false`

## Verdict unique

`MAX_LOCAL_EXECUTION_VALIDATION_BLOCKED_ENVIRONMENT`
