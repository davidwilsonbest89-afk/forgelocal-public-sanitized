# CHANGELOG — T28 Extensions locales contrôlées

## 2026-08-25

1. Baseline V6 capturée avant code depuis `999374d99b7996504ba91e421850a2fe84afb78d`.
2. Contrat, décisions produit et modèle de menace T28 versionnés avant implémentation.
3. Repository SQLite local-first ajouté : ZIP borné, manifest parser strict, traversal/symlink refusal, staging atomique, objets immuables, lifecycle, audit redacted et compensation après échec transactionnel.
4. API Core ajoutée sous bearer, loopback et origin guard : import, list/detail redacted, approve, assign, update, rollback, revoke/quarantine et purge.
5. Correction découverte par tests : acknowledgement comparé à toutes les permissions/host/content matches normalisées, plutôt qu’aux seules catégories high-risk.
6. Correction découverte par tests : scan SQL d’approbation aligné sur les onze colonnes sélectionnées.
7. Correction découverte par test d’intégration : fermeture des rows SQLite avant les sous-requêtes de `List`, supprimant un deadlock avec `SetMaxOpenConns(1)`.
8. Classification high-risk renforcée pour les host patterns globaux.
9. Tests T28 repository/API ajoutés : positif, refus auth/origin/loopback, permissions, high-risk, lifecycle, assignation, rollback/revoke, persistence, concurrency, traversal, symlink, JSON concaténé, compensation et redaction.
10. Gosec T28 repository corrigé à `found=0` avec une annotation locale justifiée pour les permissions owner-only 0700. Gitleaks extraction/diff corrigés sans leak. Govulncheck corrigé depuis le module : aucune vulnérabilité trouvée. SBOM SPDX Syft généré.
11. Suite globale Go conserve un finding runtime V6 préexistant lié à la configuration BrowseForge Chromium/Docker-GHCR ; aucun navigateur/runtime n’a été lancé par T28.
12. Branche publiée : `feature/t28-local-extensions-controlled`, HEAD `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578`.

## Non exécuté volontairement

Aucun téléchargement de package d’extension, chargement/exécution, Camoufox, Chromium, proxy/cookie réel, SystemVault natif, migration, production runtime, release, T29, T39, T40, T41 ou T42.


## 2026-08-25 — T28-EVIDENCE-QUALIFICATION-R1

1. Baseline R1 créée depuis un clone public sparse neuf ; la première passe CWD incorrecte est conservée, puis la reprise depuis le clone a confirmé tags, objets, lignée après déshallow ciblé et `fsck=0`.
2. `go test -count=1 -race ./...` exécuté exactement sur deux worktrees propres baseline/HEAD ; les deux runs corrigés ont retourné code 0. Aucun finding runtime historique n’a été reproduit ; aucune allowlist ou modification de test n’a été utilisée.
3. Tests T28 ciblés sous race exécutés avec code 0.
4. OSV Scanner v1.9.2 installé explicitement et exécuté réellement sur `go.mod` baseline/head ; 46 identifiants uniques par côté, code 1. Les tentatives v2 incompatibles avec Go 1.25.13 et le code 127 initial sont conservés comme diagnostics, sans faux PASS.
5. Gitleaks R1 a inspecté 8 commits réels de la plage baseline→HEAD par arbres de chemins modifiés, tous code 0 sans leak. L’extraction ZIP publiée a aussi passé `--no-git --redact`, code 0. L’ancien rapport `0 commits scanned` n’est pas supprimé.
6. Gosec R1 a trouvé six findings de chaque côté ; la comparaison normalisée par règle/fichier/détail donne `new_findings=0` et `resolved_findings=0`. Aucun finding T28 n’est masqué par allowlist globale.
7. Documentation corrigée uniquement : `update_url`/`updateURL` explicitement ignoré, non suivi et non exécuté ; gate exact `camoflox_execution_authorized=false` maintenu ; absence de navigateur/extension/proxy/processus externe rappelée.
8. Verdict probatoire : `T28_EVIDENCE_QUALIFICATION_R1_READY_FOR_INDEPENDENT_REVIEW`.


## 2026-08-25 — T28-FINAL-CLOSURE-PASS

1. La passe ciblée a découvert un défaut concret et reproductible : après modification du blob ZIP géré, `Approve` ne revérifiait pas le digest stocké et pouvait accepter un package altéré.
2. Correction minimale appliquée : `verifyBlobIntegrity` compare taille et SHA-256 avant `Approve`, `Assign` et `Rollback`; l’API mappe l’échec vers `INTEGRITY_MISMATCH`. Aucun navigateur, runtime, réseau de package ou processus externe n’est impliqué.
3. Tests ciblés ajoutés pour ZIP corrompu, ZIP dépassant la limite, conservation de toutes les permissions sensibles et host patterns, ignorance de `update_url`, purge lifecycle et package modifié après import. Les tests T28 sous race, vet ciblé, build et diff check corrigés ont tous retourné code 0.
4. La clôture T28 est proposée comme `T28_APPROVED_VERIFIABLE_LOCAL`, limitée au Core local et aux artefacts de conservation. Cette approbation ne couvre aucun runtime navigateur, extension chargée/exécutée, SystemVault natif, proxy/cookies réels ou release.


## 2026-08-25 — T28-POSTFIX-FINAL-CLOSURE

1. Le correctif d’intégrité est publié dans `f0701da849ce0f9073397bb42ded5e2e76b29ef1`; les tests de régression `Approve`, `Assign` et `Rollback` sont publiés dans `e806d4e915d1b702362389411d8ac823551df044`.
2. La requalification post-correctif depuis le commit `66f3bf09d3139e22f1885f7336c9edd879a32ede` a passé les tests T28 sous race, vet, build et diff check avec code 0.
3. Gitleaks a passé le scan non vide des fichiers modifiés et l’extraction du nouveau ZIP avec code 0 et aucun leak. Le diagnostic séparé `0 commits scanned` est conservé et n’est pas présenté comme une validation de plage.
4. Le nouveau ZIP `t28-postfix-evidence-v1.zip` porte le SHA-256 `4efda01771ed7af135769dfa68caa8bdc6f226ca7cad5bf894dcfce05f5c8923`; le nouveau bundle `t28-postfix-evidence.delta.bundle` porte le SHA-256 `e9f65a5b9a734933f20ecf13b05f73136e604dc006b50dfc0d40286d91262097`.
5. `unzip -t`, extraction, manifeste/checksums, sidecars neutres, bundle verify, clone seedé par baseline et fsck ont passé la vérification fraîche.
6. Statut : `T28_APPROVED_VERIFIABLE_LOCAL`, limité au Core local et aux preuves post-correctif. Aucun runtime, navigateur, proxy/cookie réel, SystemVault natif, migration, production ou release n’a été exécuté.


## 2026-08-25 — Amendement administratif HEAD/Gosec T28-POSTFIX

Le HEAD GitHub actuel est `da04a2feb82437f5da78cdd7bb869d4136a9fbde`. Le package ZIP/bundle avait été vérifié fraîchement au HEAD intermédiaire `dff092cd5eeb27ddff6e331c22594a03c5b4e1a7`. Les commits postérieurs `19ebcbaf6f95b5cf5d5c16728e59746293a1f565` et `da04a2feb82437f5da78cdd7bb869d4136a9fbde` sont uniquement documentaires, de conservation et de preuve ; aucun fichier `internal/` T28 n’a changé entre ces références.

La comparaison Gosec normalisée est exactement : `baseline_findings=6`, `head_findings=6`, `new_findings=0`, `resolved_findings=0`, selon `evidence/T28/GOSEC_R1_NORMALIZED_COMPARISON.json` et `evidence/T28/GOSEC_R1_BASELINE_HEAD_RAW.log`. Le ciblage post-correctif donne `0` finding pour `internal/extensions` dans `GOSEC_T28_EXTENSIONS_POST.json` et `6` findings historiques conservés pour l’API dans `GOSEC_T28_API_PACKAGE_AFTER.json`.

La formulation de preuve est corrigée : `git fsck --full` a retourné `exit_code=0`. Le statut reste `T28_APPROVED_VERIFIABLE_LOCAL`, sans modification de code, de ZIP/bundle, de runtime ou de lot suivant.
