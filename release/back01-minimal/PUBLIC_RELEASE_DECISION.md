# Décision de release publique — BACK-01

**Date :** 14 août 2026
**Branche :** `forgelocal-back01`
**Décision :** `PUBLIC_RELEASE_BLOCKED`

> L’artefact minimal `forgelocal-back01-core-0.1.0-back01-07f603d-linux-amd64.tar.gz` reste approuvé uniquement pour un **pilote local contrôlé**. Il n’est pas autorisé à devenir une release publique à ce stade.

## Éléments validés

| Contrôle | État | Preuve |
|---|---|---|
| AC-BACK-01 de bout en bout | Vert | `validation_back01_integration/final/ac_back_01_with_explicit_chromium.out` |
| Relance Chromium locale explicitement journalisée | Vert | binaire, version, PID éphémère, profil, `about:blank`, arrêt et locks |
| Suite Go complète | Verte | `validation_back01_integration/final/go_test_all_after_runtime_gate.out` |
| Périmètre minimal avec tests et `gosec` | Vert, 0 alerte | `minimal_artifact_runtime_gate_build.out` |
| Checksum et scan de secrets de l’archive extraite | Verts, 0 fuite | `minimal_artifact_runtime_gate_validation.out` et `gitleaks_runtime_gated_archive.log` |
| Runtime incorporé dans l’archive | Non | manifeste `runtime_included: false` |
| Candidat Chromium `151.0.7922.108` | Vert techniquement, non public | `RUNTIME_CANDIDATE_CHROMIUM_151.0.7922.108.json` et E2E candidat |
| SBOM de l’artefact candidat | Vert | `SBOM.spdx.json` SPDX-2.3, hashé dans les deux manifestes |
| Manifeste externe de release | Vert mais non signé | hash d’archive vérifié ; signature mainteneur encore requise |

## Bloquants publics

| Gate | État | Condition de levée |
|---|---|---|
| SystemVault natif | Bloquant | Exécuter et compléter `SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md` sur chaque OS annoncé, dans une session desktop déverrouillée, hors conteneur et sans `sudo` |
| Cas de révocation et coffre verrouillé | Bloquant | Montrer des refus contrôlés, sans fallback en clair et sans fuite |
| Anti-fuite intégrée | Bloquant | Fournir `systemvault-anti-leak.json` vert après un flux profil → backup → restauration réel |
| Paquet runtime QA exact | Bloquant | Archiver le `.deb` exact, son SHA-256, `InRelease`, le keyring, l’empreinte de clé et le contrôle d’index signé |
| Pin runtime reproductible | Bloquant | Conserver le verrou `RUNTIME_RELEASE_LOCK.json` à l’état complet ou revalider entièrement un nouveau candidat |
| Signature du manifeste externe | Bloquant | Signer le manifeste d’archive avec une clé mainteneur approuvée et publier la clé de vérification |
| Revue licence/distribution runtime | Bloquant | Vérifier les droits de redistribution du runtime externe avant toute publication |
| Portée et OS annoncés | Bloquant | Restreindre la communication à `RELEASE_SCOPE_AND_OS_MATRIX.md` et ne revendiquer que les environnements testés |

## Constat du runtime QA

Le Chromium qui a servi au pilote local initial est `151.0.7922.71-1xtradeb1.2404.1`. Cette version n’est plus disponible dans l’index APT courant et son paquet n’était ni présent dans le cache local ni disponible à l’URL directe testée. Cette situation interdit toute substitution implicite.

Le candidat distinct `151.0.7922.108-1xtradeb1.2404.1` a été capturé avec ses deux paquets requis, `InRelease`, keyring, `Packages.gz` et contrôles de checksum. Son E2E AC-BACK-01 est vert. Il reste **techniquement qualifié mais non public** jusqu’à la levée de tous les gates SystemVault, signature, licence et portée OS. Voir `RUNTIME_CANDIDATE_CHROMIUM_151.0.7922.108.json` et `RELEASE_SCOPE_AND_OS_MATRIX.md`.

## Autorité de changement

Cette décision ne peut être remplacée par `PUBLIC_RELEASE_APPROVED` que lorsque **tous** les fichiers de preuve exigés par `SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md` et `RUNTIME_RELEASE_LOCK.json` sont versionnés, contrôlés et revus. Camoufox demeure hors de cette décision tant qu’il n’a pas reçu une validation de runtime séparée.
