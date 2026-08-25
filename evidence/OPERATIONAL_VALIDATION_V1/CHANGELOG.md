# Changelog — OPERATIONAL_VALIDATION_V1

## 2026-08-25 — campagne post-T28

La branche dédiée `validation/operational-v1` a été créée depuis le HEAD T28 GitHub `053fd8f9612eb028c04983e0a23d311a6a099e29`. La baseline, le Core, SQLite, le contrat T28, Chromium synthétique, proxy/cookies locaux, MemoryVault, Dashboard/API et les contrôles de sécurité ont été exécutés avec données temporaires synthétiques.

Le défaut reproductible d’installation des répertoires runtime en `0755` a été corrigé dans `b1559ca53852c493ba15e4a06ad89b0c171c7938`. Le correctif impose et répare `0700` pour `profiles`, `data`, `logs` et `browsers` dans `init`, `serve` et `mcp-stdio`, avec `TestEnsureRuntimeDirsOwnerOnly`.

Résultats importants : SQLite PASS ; contrat T28 PASS ; proxy/cookies synthétiques PASS ; MemoryVault local PASS ; pnpm production audit PASS sans vulnérabilité ; Gitleaks PASS avec 0 finding. V-CORE reste FAIL non critique pour absence d’expiration/révocation du token admin. V-DASHBOARD-API reste FAIL non critique pour la violation Axe `color-contrast` et les erreurs d’assets/analytics locales. Camoufox, SystemVault natif, Docker/Buildx et plusieurs scanners ne sont pas disponibles et sont classés `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`.

Aucun `FAIL_CRITICAL` n’a été observé. T28 n’a pas été rouvert. Aucun compte, secret, cookie réel, proxy commercial, migration, release ou lot T29/T39–T42 n’a été exécuté.
