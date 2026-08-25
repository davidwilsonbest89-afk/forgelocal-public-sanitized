# Changelog — OPERATIONAL_VALIDATION_V1

## 2026-08-25 — campagne post-T28

La branche dédiée `validation/operational-v1` a été créée depuis le HEAD T28 GitHub `053fd8f9612eb028c04983e0a23d311a6a099e29`. La baseline, le Core, SQLite, le contrat T28, Chromium synthétique, proxy/cookies locaux, MemoryVault, Dashboard/API et les contrôles de sécurité ont été exécutés avec données temporaires synthétiques.

Le défaut reproductible d’installation des répertoires runtime en `0755` a été corrigé dans `b1559ca53852c493ba15e4a06ad89b0c171c7938`. Le correctif impose et répare `0700` pour `profiles`, `data`, `logs` et `browsers` dans `init`, `serve` et `mcp-stdio`, avec `TestEnsureRuntimeDirsOwnerOnly`.

Résultats importants : SQLite PASS ; contrat T28 PASS ; proxy/cookies synthétiques PASS ; MemoryVault local PASS ; pnpm production audit PASS sans vulnérabilité ; Gitleaks PASS avec 0 finding. V-CORE reste FAIL non critique pour absence d’expiration/révocation du token admin. V-DASHBOARD-API reste FAIL non critique pour la violation Axe `color-contrast` et les erreurs d’assets/analytics locales. Camoufox, SystemVault natif, Docker/Buildx et plusieurs scanners ne sont pas disponibles et sont classés `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`.

Aucun `FAIL_CRITICAL` n’a été observé. T28 n’a pas été rouvert. Aucun compte, secret, cookie réel, proxy commercial, migration, release ou lot T29/T39–T42 n’a été exécuté.


## 2026-08-25 — publication et vérification des preuves OPV1

Le package d’évidence a été généré depuis `b594f236a28483cc975dd054575d6cfb171d4f86` et publié séparément dans `9cc6f5e45fc1aff9ba7a6ca06740cb6ac17538a2`.

| Artefact | SHA-256 |
|---|---|
| `forgelocal-operational-validation-v1.zip` | `664fba54efe36a20269e60bba007082458a5d90c45ca10e9728cdf15fc7fedf5` |
| `forgelocal-operational-validation-v1.delta.bundle` | `67cce18e08f51540fceea71ba8e2dcb3c71344c24d10c4990e5e5e6460efb46f` |

La vérification publique fraîche a confirmé le HEAD `9cc6f5e…`, les quatre sidecars, les hashes, l’extraction ZIP et son manifeste non auto-référentiel, Gitleaks extraction sans finding, `git bundle verify` exit 0, refus standalone pour baseline absente, import du bundle après checkout de la baseline, checkout du source `b594f23…`, `git fsck --full` exit 0 et worktree propre. Le refus standalone a retourné exit 1 dans Git 2.43 avec l’erreur explicite de prérequis manquant ; il est conservé comme refus attendu et non comme PASS.

Le verdict est `FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE`. T28 reste clôturé localement et non rouvert ; T29/T39–T42, SystemVault natif, Camoufox, Docker et release restent non exécutés.
