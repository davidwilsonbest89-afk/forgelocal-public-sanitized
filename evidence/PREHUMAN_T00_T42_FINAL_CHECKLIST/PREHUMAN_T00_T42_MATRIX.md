# ForgeLocal — matrice exhaustive T00–T42

**Point de contrôle :** HEAD de validation `6ae02e4ceed239b9310fbf3fccb1b5170117251e`, depuis le clone neuf `/home/ubuntu/forgelocal-prehuman-fresh-20260824`.
**Baseline :** tag `t00-t27-complete-20260820` → `69411e65c880d168832a65fc8475cc97d562a9ad`.
**Branche de livraison des preuves :** `audit/t00-t42-prehuman-validation`.
**Répertoire de l’addendum :** `evidence/PREHUMAN_T00_T42_FINAL_CHECKLIST/`.

La matrice distingue la vérifiabilité de la chaîne d’artefacts d’une autorisation produit. Pour T00–T23, aucune archive unitaire n’est revendiquée : la preuve est la chaîne combinée reconstruite à partir des morceaux LFS historiques. Le paquet gelé original conserve le nom `forgelocal-t28-t42-prehuman-validation.*`; l’addendum ne le renomme pas et ne le recrée pas.

| Lot | Statut strict | Référence connue | Preuve / contrôle applicable | Limitation et gate |
|---|---|---|---|---|
| T00 | `APPROVED_VERIFIABLE_LOCAL` | Baseline `t00-t27-complete-20260820` → `69411e65…` | Chaîne combinée T00–T23, morceaux LFS reconstruits, hash et extraction conformes | `NOT_IN_SCOPE` : pas de ZIP unitaire ; aucune autorisation produit |
| T01 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | `NOT_IN_SCOPE` : pas d’artefact unitaire revendiqué |
| T02 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de runtime réel |
| T03 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de données réelles |
| T04 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de release |
| T05 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de gate levée |
| T06 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de runtime réel |
| T07 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de credentials réels |
| T08 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | `t08_authorized=false`; Camoufox interdit |
| T09 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de proxy/cookie réels |
| T10 | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne ; exécution métier finale `NOT_APPLICABLE_UNDER_CURRENT_GATES` | Même chaîne combinée T00–T23 ; checklist addendum | Contrôles statiques et Dashboard exécutés ; Playwright arrêté avant métier faute de configuration T10 protégée | `FORGELOCAL_CORE_BASE_URL` et token/config absents ; aucun Core lancé |
| T11 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas d’observation runtime |
| T12 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de données personnelles |
| T13 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de migration |
| T14 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de release |
| T15 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de gate levée |
| T16 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de runtime réel |
| T17 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de credentials réels |
| T18 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de proxy/cookie réels |
| T19 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de SystemVault natif |
| T20 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de secret importé/exporté |
| T21 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas d’intégration runtime |
| T22 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Vérification globale du paquet historique | Pas de release |
| T23 | `APPROVED_VERIFIABLE_LOCAL` | Même chaîne combinée T00–T23 | Hash SHA-256, `unzip -t` et extraction fraîche conformes | Pas d’archive unitaire ; preuve limitée à la chaîne combinée |
| T24 | `APPROVED_VERIFIABLE_LOCAL` | Dossier `evidence/PREHUMAN_T00_T42/` | ZIP LFS, hash et manifeste global conformes | Vérification d’artefact ; aucune autorisation produit |
| T25 | `APPROVED_VERIFIABLE_LOCAL` | Dossier `evidence/PREHUMAN_T00_T42/` | ZIP LFS, hash et manifeste global conformes | Vérification d’artefact ; aucune autorisation produit |
| T26 | `APPROVED_VERIFIABLE_LOCAL` | Dossier `evidence/PREHUMAN_T00_T42/` | ZIP LFS, hash et manifeste global conformes | Vérification d’artefact ; aucune autorisation produit |
| T27 | `APPROVED_VERIFIABLE_LOCAL` | Dossier `evidence/PREHUMAN_T00_T42/` | Tarball, ZIP, bundles, sidecars et logs vérifiés localement | Aucune autorisation produit ou release déduite |
| T28 | `BLOCKED` | Audit `08107b3ef2d59c8054720fdf4dd81554d4984afb` | Documentation et audit conservés | Contrat/autorisation d’extension absents ; ne pas implémenter |
| T29 | `BLOCKED` | Audit `08107b3ef2d59c8054720fdf4dd81554d4984afb` | Documentation conservée | SystemVault natif non testé et import/export non autorisé |
| T30 | `PENDING_REMOTE_EVIDENCE_RECONCILIATION` | Commit `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb`; ZIP `c07321fb…`; bundle `c4f514ff…` | Commit et hashes accessibles | Branche/kit distant canonique non rattaché à la clôture finale |
| T31 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t31-canvas-webgl-audio` → `9dad98703a748af61cbd3573b00e2bb036739a69` | Code redacted, tests, ZIP/bundle et sidecars conservés | `BASELINE_DISCOVERY_RAW.log` historique absent ; reconstruction postérieure explicitement étiquetée ; projections `UNSUPPORTED` |
| T32 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t32-client-rects` → `d7279e81dd724ba2278a65838bc65aaa16912007` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; aucune géométrie DOM réelle |
| T33 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t33-synthetic-geolocation` → `693632791041fde14db14ec8982b8bff1060a8d3` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; géolocalisation synthétique seulement |
| T34 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t34-hardware-diagnostics` → `c7f66da0da6f547813d10826dfd7772ad7e0f4b6` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; aucun accès hôte |
| T35 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t35-font-bundle` → `891b2d166a30a96d5ca872b473251ed1a3706ba3` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; aucun binaire/font système fourni |
| T36 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t36-drift-detection` → `099ff7bbf384ff900ec44fc3e9aafd1fc273f0dd` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; comparaison redacted seulement |
| T37 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t37-profile-health` → `c0ad051c77e153b3ec4435fbc7ff98e30b96969b` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; santé agrégée redacted, non observationnelle |
| T38 | `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` | `work/t38-session-lifecycle` → `f8fcfa02935a62fa66c3e22bb6bd5a01493f2f62` | Code redacted, tests, ZIP/bundle et sidecars conservés | Reconstruction postérieure ; lifecycle mémoire local, aucun runtime |
| T39 | `BLOCKED` | `work/t39-secret-import-export-blocked` → `08061f1e55a0bba01a4a28df66dc852e8f345ade` | Preuves documentaires conservées | Dépend de T28/T29 ; SystemVault natif non qualifié |
| T40 | `BLOCKED` | `work/t40-integration-gated` → `077ba90b2c522415aeefdc2d651c457b8c59683d` | Preuves documentaires conservées | Intégration runtime interdite par les dépendances et gates |
| T41 | `BLOCKED` | `work/t41-release-readiness-blocked` → `944376fd1c9d22dad44730854ce4b2d6203c743b` | Preuves documentaires conservées | Release publique interdite ; aucun artefact de release exécuté |
| T42 | `BLOCKED` | `work/t42-final-closure` → `6489af39a4ac4f91f9f7dc1435f10b2bd10dfdc0` | Livraison T42 et documentation conservées | Clôture technique possible ; clôture produit/release impossible |

## Lots de correction CR01–CR05

| Lot | Statut | Preuve et limite |
|---|---|---|
| CR01 | `APPROVED_VERIFIABLE_LOCAL` | Correction documentaire/chaîne d’artefacts vérifiée ; aucune approbation produit |
| CR02 | `APPROVED_VERIFIABLE_LOCAL` | Sidecars et vérification de transport vérifiés ; aucune levée de gate |
| CR03 | `APPROVED_VERIFIABLE_LOCAL` | Requalification et journalisation conservées ; pas de runtime réel |
| CR04 | `APPROVED_VERIFIABLE_LOCAL` | Contrôles historiques et cohérence documentaire vérifiés |
| CR05 | `APPROVED_VERIFIABLE_LOCAL` | Livraison de preuves gelée ; release toujours bloquée |

## Identité des artefacts et gates globales

Le paquet original reste inchangé : `forgelocal-t28-t42-prehuman-validation.zip` a le SHA-256 `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80`, et `forgelocal-t28-t42-prehuman-validation.bundle` a le SHA-256 `14ef76cb68e7f64ff49fdc649cbcf96c5c69b0e9c410c5824a0592b7e33d1d14`. Son manifeste, ses sidecars portables et ses journaux restent dans `evidence/PREHUMAN_T00_T42/`. Le nom historique `t28-t42` est conservé malgré la couverture T00–T42, afin de ne pas modifier la sémantique du paquet gelé.

Les gates restent : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. La checklist finale a été exécutée sur le commit de validation `6ae02e4…`; les commits ultérieurs de livraison de cet addendum ne contiennent que des preuves et de la documentation.

## Références

[1]: ../PREHUMAN_T00_T42/PREHUMAN_T00_T42_SUMMARY.md "Résumé de la prévalidation gelée"
[2]: ../PREHUMAN_T00_T42/SHA256SUMS "Hashes du paquet gelé"
[3]: ./PREHUMAN_FINAL_FINDINGS_CLASSIFICATION.md "Classification des findings finaux"
