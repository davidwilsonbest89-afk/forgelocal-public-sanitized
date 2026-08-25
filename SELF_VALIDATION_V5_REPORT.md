# SELF_VALIDATION_V5_REPORT — T00–T42

**Date d’exécution :** 2026-08-25 UTC
**Dépôt :** [forgelocal-public-sanitized][1]
**Branche source vérifiée :** `audit/t00-t42-self-validation-synthetic-e2e`
**HEAD source vérifié :** `b34fa5c02ff20144abfb5d240db1c67ad1f038f9`
**Baseline demandée :** tag `t00-t27-complete-20260820`, résolu dans le clone neuf vers `72d54110c89583beacc556bb103f881b667d8137`
**Nombre de commits baseline..HEAD :** 58
**Branche de livraison prévue :** `audit/t00-t42-self-validation-enhanced-v5`

## Objet et périmètre

Cette campagne renforce la self-validation sans revue humaine externe et sans modifier le code produit. Le clone v5 a été créé depuis la branche v4 publiée avec `GIT_LFS_SKIP_SMUDGE=1`. La baseline a été récupérée explicitement, son ancêtre commun et le nombre de commits ont été prouvés, puis Gitleaks a été rejoué sur la plage demandée, sur le checkout frais et, en complément nécessaire, sur chacun des 58 arbres de commits.

Les contrôles Core et Dashboard restent strictement locaux. Aucun Camoufox, proxy réel, cookie réel, profil utilisateur, donnée utilisateur, SystemVault natif, migration, runtime de production ou release n’a été lancé. Le test Axe utilise un Dashboard loopback synthétique ; il a pour politique d’échouer dès qu’une violation WCAG de sévérité `serious` ou `critical` est détectée.

## Résumé des résultats

| Contrôle | Résultat | Résultat vérifiable |
|---|---|---|
| Clone neuf et baseline | PASS | Ancêtre confirmé, 58 commits listés, HEAD `b34fa5c` |
| Gitleaks demandé avec `--log-opts` | INCOMPLET | La version installée annonce `0 commits scanned`; ce n’est pas présenté comme PASS |
| Gitleaks arbres explicites | FINDINGS | 58/58 arbres scannés, 348 détections cumulées `generic-api-key`, redacted |
| Gitleaks checkout frais `--no-git` | FINDINGS | 6 détections redacted, code 1 |
| GolangCI-Lint 2.13.1 | FINDINGS | Baseline 90, HEAD 90, 2 nouveaux, 2 résolus ; binaire compilé avec Go 1.27.0 et exécution réussie sur projet Go 1.25 |
| Go shuffle | PASS | `go test -shuffle=on -count=3 ./...` code 0 |
| Go shuffle + race | PASS | `go test -shuffle=on -count=3 -race ./...` code 0 |
| Semgrep 1.174.0 | FINDINGS | 18 findings issus des règles locales Go/TypeScript, tous à classer |
| Grype CycloneDX | FINDINGS | Scan code 0, 2 correspondances `High` sur `golang.org/x/mod` |
| Grype SPDX | FINDINGS | Scan code 0, 2 correspondances `High` sur `golang.org/x/mod` |
| Axe Playwright | BLOCKED_BY_FINDINGS | 2 violations : 1 `serious` de contraste, 1 `moderate` de viewport ; le test échoue volontairement sur la première |
| Cleanup Axe | PASS | Aucun processus, token, base SQLite ou répertoire temporaire résiduel |

## 1. Gitleaks obligatoire et couverture historique

La commande demandée a été exécutée depuis un clone neuf avec `BASE=t00-t27-complete-20260820`, récupération du tag, test d’ancêtre, `git rev-list --count`, liste complète des hashes et `gitleaks detect --log-opts`. La commande Gitleaks historique termine techniquement avec code 0 mais annonce `0 commits scanned`; ce résultat est classé **INCOMPLET**, jamais comme PASS.

Pour fournir une couverture réelle de la plage, un second contrôle a extrait séparément les arbres Git des **58 commits** dans des répertoires temporaires en conservant les pointeurs LFS, puis a exécuté Gitleaks `--no-git` sur chaque arbre. Les 58 scans ont retourné des findings, pour 348 détections cumulées, toutes sous la règle `generic-api-key`. Les valeurs sont redacted et les JSON individuels sont livrés. Le checkout frais a également été scanné avec `--no-git` et a retourné 6 détections redacted, code 1.

La conclusion correcte est donc : **la plage est non vide et couverte par un scan explicite, mais elle n’est pas propre**. Les findings historiques de fixtures et preuves sont conservés pour revue ; aucun faux positif n’est déclaré automatiquement et aucun secret en clair n’est livré.

## 2. GolangCI-Lint obligatoire

GolangCI-Lint `2.13.1` a été téléchargé depuis sa release officielle et vérifié par checksum. Le binaire indique une compilation avec `go1.27.0`; il s’est exécuté sans l’incompatibilité rencontrée avec la version 1.61.0. Le projet a été contrôlé avec le toolchain `go1.25.13` et `GOTOOLCHAIN=local`.

| Comparaison | Nombre |
|---|---:|
| Findings baseline | 90 |
| Findings HEAD | 90 |
| Findings nouveaux | 2 |
| Findings absents au HEAD | 2 |

Les deux findings nouveaux sont conservés individuellement : `ineffassign` sur `cmd/server/cli_runtime.go:240` pour une affectation inefficace à `cfg`, et `staticcheck` `SA1019` sur `internal/api/sessions.go:310` pour l’utilisation d’une API Playwright dépréciée. Aucun `nolint`, exclusion globale ou modification produit n’a été ajouté. Les deux findings résolus sont conservés comme variation baseline/HEAD, sans que cela constitue une approbation globale.

## 3. Flakiness Go

Les deux campagnes demandées ont terminé avec code 0 : `go test -shuffle=on -count=3 ./...` et `go test -shuffle=on -count=3 -race ./...`. Elles ont été exécutées avec `CGO_ENABLED=1`, Go 1.25.13 et `GOTOOLCHAIN=local`. Aucun fichier produit n’a été modifié pendant ces tests.

## 4. Semgrep indépendant

Semgrep `1.174.0` a été installé dans un environnement virtuel isolé. Le scan utilise quatre règles locales ciblées Go/TypeScript, sans téléchargement de règles distantes. Il retourne 18 findings, tous issus de la règle de revue de `rand.Read` dans des chemins Go. Ces diagnostics sont des signaux de contexte, non des preuves autonomes de vulnérabilité ; chacun dispose dans l’analyse jointe d’un fichier, d’une ligne, d’un risque, d’un propriétaire et d’une condition de levée.

Le fichier d’inventaire des fichiers scannés et le JSON brut Semgrep sont livrés. La configuration de règles est également livrée afin que le contrôle soit rejouable exactement.

## 5. Grype sur les SBOM

Grype `0.117.0` a été installé depuis sa release officielle et vérifié par checksum. Les deux SBOM déjà produits ont été scannés séparément. Chaque scan termine avec code 0 mais retourne deux correspondances de sévérité `High` sur `golang.org/x/mod`. Les JSON bruts sont conservés et les correspondances restent à trier par CVE, version affectée, version corrigée et contexte d’utilisation.

Grype signale aussi que certains binaires Go ne portent pas de symboles de fonction et que la distribution OS de certains packages ne peut pas être déterminée. Ces avertissements sont conservés ; ils empêchent de présenter le résultat comme une absence universelle de vulnérabilités.

## 6. E2E Playwright synthétique avec Axe

Le scénario a utilisé Core et Dashboard sur loopback, ports temporaires, base temporaire, token éphémère à permissions restrictives et aucun stockage navigateur. Les appels HTTP hors origines loopback sont surveillés. Le test a contrôlé l’absence de code dans les URL, l’absence de token dans les paramètres, l’absence de stockage navigateur et la politique de rejeu héritée du bootstrap local.

Axe `@axe-core/playwright` a trouvé deux violations : une violation `color-contrast` d’impact `serious` sur plusieurs métriques du rail d’observabilité, et une violation `meta-viewport` d’impact `moderate`. Conformément au mandat, le test échoue sur toute violation `serious` ou `critical`, avec code Playwright 1. Le JSON Axe complet est conservé avant le cleanup. Le produit n’a pas été corrigé dans cette campagne, car la consigne impose de ne pas modifier le produit ; ces findings doivent être traités dans un lot UI autorisé puis soumis à un nouveau run.

Le cleanup vérifié après l’échec est propre : aucun processus temporaire, token, base SQLite ou répertoire temporaire ne subsiste. Le test Axe a été retiré du checkout et les manifests Dashboard ont été restaurés à leur contenu Git exact.

## 7. Classification et décision

Les classifications détaillées sont livrées dans `SELF_VALIDATION_V5_MANDATORY_ANALYSIS.md` et `SELF_VALIDATION_V5_COMPLEMENTS_ANALYSIS.md`, complétées par les JSON bruts par finding. Les catégories sont : findings historiques Gitleaks à revue, findings nouveaux GolangCI-Lint à correction ou justification ciblée, signaux Semgrep à revue de contexte, correspondances Grype à triage CVE/package et violations Axe à correction UI.

La campagne renforce substantiellement la couverture mais ne permet pas de déclarer une campagne totalement verte. Gitleaks historique et frais produisent des findings, GolangCI-Lint produit deux findings nouveaux, Semgrep produit 18 findings, Grype produit quatre correspondances `High` sur deux SBOM et Axe bloque volontairement l’E2E par une violation sérieuse.

## Statut exact

`T00_T42_SELF_VALIDATION_ENHANCED_PENDING_INDEPENDENT_REVIEW`

Ce statut ne constitue ni une release, ni une approbation produit, ni une levée de gate. Les statuts et interdictions restent inchangés : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`.

## Prochaine étape autorisée

Une revue humaine indépendante peut examiner les JSON bruts et les classifications. Après autorisation explicite, les propriétaires peuvent traiter les findings UI, SAST, lint et dépendances dans des lots séparés, puis rejouer les contrôles affectés. Aucun runtime réel, migration, donnée utilisateur ou release ne doit être lancé dans cette étape.

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/audit/t00-t42-self-validation-synthetic-e2e "Branche source v4 publiée"
[2]: https://github.com/golangci/golangci-lint/releases/tag/v2.13.1 "Release officielle GolangCI-Lint 2.13.1"
[3]: https://semgrep.dev/docs/ "Documentation officielle Semgrep"
[4]: https://github.com/anchore/grype/releases/tag/v0.117.0 "Release officielle Grype 0.117.0"
[5]: https://github.com/dequelabs/axe-core-npm/tree/develop/packages/playwright "Package officiel axe-core Playwright"
