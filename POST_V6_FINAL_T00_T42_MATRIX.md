# ForgeLocal — matrice exhaustive T00–T42

**Point de contrôle :** V6 gelé au commit `999374d99b7996504ba91e421850a2fe84afb78d`, packaging publié au commit `727e94bcf2bf40bc9388ee2df82bdb241962e5b2`, depuis les clones publics sparse vérifiés le 2026-08-25.
**Baseline :** tag `t00-t27-complete-20260820` → `69411e65c880d168832a65fc8475cc97d562a9ad`.
**Branche de livraison des preuves :** `audit/t00-t42-v6-findings-remediation` et branches post-gel séparées listées ci-dessous.
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
| T30 | `T30_REMOTE_EVIDENCE_RECONCILED_PENDING_INDEPENDENT_REVIEW` | Branche `audit/t00-t42-t30-remote-evidence-reconciliation-v6` → `e5241c9076814adf2b4870e4a61b2f8702f1e1fa`; branche canonique V6 `audit/t00-t42-v6-findings-remediation` → `727e94b…`; tag `t00-t42-v6-local-qualified-2026-08-25` → `999374d…` | ZIP T30 `abd5b359…`; bundle T30 `f9efb212…`; clone, sidecars, extraction, manifeste, bundle, Gitleaks et fsck vérifiés | Revue indépendante encore requise ; aucune exécution T30 ni gate levée |
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

## Lots post-gel exécutés dans l’ordre senior

Les lots suivants sont strictement séparés de la branche V6 et ont chacun reçu une baseline brute, un registre, un changelog, un todo, un bundle delta, des sidecars, un ZIP, un manifeste non auto-référentiel et une vérification publique sparse. Les validations de transport ne constituent pas des releases et ne réactivent aucune fonctionnalité bloquée.

| Lot | Branche publiée / commit final | Preuve et hash | Statut strict | Limite / prochain owner |
|---|---|---|---|---|
| DOCKER-NATIVE-NONRUNTIME-VALIDATION | `audit/t00-t42-docker-native-nonruntime-validation` / `38a273de9e9ba239338d4661119cdd53abd9ad28` | ZIP SHA `3e989a014a8cd9950174e1c986ce94a9bcb44f357a89c84c41de0a1741067851`; bundle SHA `929b9681ce32ef8f73e81399b84a36e80cacbae6e221ad439495548a3291d457` | `DOCKER_NATIVE_NONRUNTIME_VALIDATION_BLOCKED_PENDING_DOCKER_BUILDX` | Docker/Buildx absent : build, history d’image, Trivy image, permissions runtime et cleanup container non exécutés. Owner : plateforme / sécurité conteneurisation. |
| LFS-CONSERVATION-RESTORATION | `audit/t00-t42-lfs-conservation-restoration-v6` / `1fa32f68dae6a745cd7f8fec2e883f13b2398aca` | ZIP SHA `344e1d7d6a03cddde1136d7e47c2e9c5e3db44cf6dd1b0e76df4a30317b02fa1`; bundle SHA `26c4d13f8d46b414f8e84af889b766076d4744c785370f21fedfde4742b59ced` | `LFS_CONSERVATION_RESTORED_BYTE_IDENTICAL_FSCK_ZERO` | Douze OID vérifiés individuellement par taille et SHA-256 ; `git lfs fsck=0`. Revue de provenance et réplication durable restantes. Owner : administrateur dépôt / preuves. |
| OSV-TOOLCHAIN-RECONCILIATION | `audit/t00-t42-osv-toolchain-reconciliation-v6` / `0f055e7fff5ac2f70885c937b2a036ffc45b5526` | ZIP SHA `08321e9dfd91587d0178a5fa40a55e5c5081f5eaf8510af4b11f3e12cf67d4d5`; bundle SHA `8fecaaa71b5109f06f7c148e5beba0f35450113dc2ac80136b3e3c13d48f2c2f` | `OSV_RESULTS_RECONCILED_AS_INDIVIDUAL_VERSION_MODELING_EXCEPTIONS_PENDING_COMPATIBLE_SCANNER` | 46 advisories conservés individuellement ; divergence directive Go 1.25.0 / toolchain 1.25.13. Owner : mainteneurs Go / sécurité dépendances. |
| LICENSE-COMPLIANCE-TRIAGE | `audit/t00-t42-license-compliance-triage-v6` / `0931ec8905b4b38ce9ce1f5d195287dae02267b3` | ZIP SHA `ab86ae3be472954048267633b697cdeb2d2f2a4886539408eb23d2f242a4f307`; bundle SHA `7c9b838ba5f822265f59d4a8e1cf0e114bcf4bd115e77e189cbd2b138cfc47e1` | `LICENSE_UNKNOWN_RETAINED_PENDING_OFFICIAL_SOURCE_REVIEW` | 749 composants courants : 3 identifiés, 746 UNKNOWN ; aucune licence devinée. Owner : Legal / OSS. |
| STATIC-DEBT-PRIORITIZATION | `audit/t00-t42-static-debt-prioritization-v6` / `78b6c49ccbe313ae4453c8ec6c042c4828faf296` | ZIP SHA `e0641cb1254ffac6085e5146251b856c8b40334a44b358864f196a7b6a447e71`; bundle SHA `1e54450bea38d5aa8b0417030412cec8434eee2e37bccccca6a87560e13ff1b5` | `STATIC_DEBT_INDIVIDUALLY_TRIAGED_NO_CODE_CHANGE` | 34 Staticcheck + 89 GolangCI-Lint ; corrections élevées réservées à des lots futurs testés. Owner : mainteneurs Go. |
| GITLEAKS-HISTORICAL-METHOD-HARDENING | `audit/t00-t42-gitleaks-historical-method-hardening-v6` / `6fd0e31f1342be8c4eb067b9c72afa587ecc412e` | ZIP SHA `11ea9935dd261c466007a9706c182d9174d820ad019af7a49582ee2b3b2a5eb5`; bundle SHA `67057075cf9465bca55d17f920a80eaf0780b60c878ccab2fce2eec1d5972936` | `GITLEAKS_HISTORICAL_TREE_BY_TREE_METHOD_HARDENED` | Quatre arbres V6 de contenu inventoriés, JSON individuels et configuration exacte. Tout intervalle vide est `INVALID_ZERO_COMMIT_RANGE`, jamais PASS. Owner : sécurité dépôt / preuves. |
| T30-REMOTE-EVIDENCE-RECONCILIATION | `audit/t00-t42-t30-remote-evidence-reconciliation-v6` / `e5241c9076814adf2b4870e4a61b2f8702f1e1fa` | ZIP SHA `abd5b359e4ed69719f224222d940c25b3b99635f28017b7393159ca4893abf8e`; bundle SHA `f9efb212024775c45da85d0c72cb8f7c5935bd84417a56c7136dd0a633da61ba` | `T30_REMOTE_EVIDENCE_RECONCILED_PENDING_INDEPENDENT_REVIEW` | Branche V6, tag, commit, bundle, ZIP, sidecars, manifeste, clone et fsck rattachés ; revue indépendante encore requise. Owner : responsable preuves distantes. |

## Résultat consolidé des contrôles publics

Le journal `POST_V6_LOTS_PUBLIC_VERIFY.log` montre, pour les sept branches, un clone distant correspondant au SHA annoncé, un `git fsck --full` initial et final à zéro, un fetch LFS ciblé, des hashes ZIP et bundle correspondants, `unzip -t`, checksums internes, `git bundle verify` et Gitleaks à zéro. Le premier essai Docker consolidé a utilisé un nom de branche inexistant et a été conservé séparément ; il a été corrigé avant la vérification finale. Le clone T30 a produit du bruit d’auto-packing lié à la pression disque, mais toutes les étapes ciblées ont retourné zéro et le clone temporaire a été supprimé.

## Garde-fous finaux

La décision V6 demeure **`V6_LOCAL_QUALIFIED_BASELINE_FROZEN`** et ne constitue pas une release. Les branches et tags V6 restent gelés ; les sept lots ci-dessus sont documentaires ou de validation isolée et ne modifient pas opportunistement le Core, le Dashboard métier ou les wrappers V3–V6 historiques.

Les gates restent strictement : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. T28, T29, T39, T40, T41 et T42 restent `BLOCKED`. T30 est rattaché à des preuves distantes mais reste `PENDING_INDEPENDENT_REVIEW`. Aucun Camoufox, proxy réel, cookie réel, donnée utilisateur, SystemVault natif, migration utilisateur, runtime de production ou release n’a été démarré.

**Prochain owner global :** revue indépendante de la chaîne de preuves, puis owners spécialisés selon chaque ligne du tableau. Aucune fonctionnalité bloquée ne doit débuter sans nouvelle instruction produit explicite.
