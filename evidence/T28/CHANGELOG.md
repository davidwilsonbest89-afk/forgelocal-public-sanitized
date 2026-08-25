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
