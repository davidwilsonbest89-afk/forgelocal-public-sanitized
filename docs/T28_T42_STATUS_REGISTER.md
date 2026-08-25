# ForgeLocal T28–T42 — registre canonique de clôture

**Date de validation :** 2026-08-24
**Baseline immuable :** tag `t00-t27-complete-20260820`, commit `69411e65c880d168832a65fc8475cc97d562a9ad`
**Règle :** une validation locale ne vaut pas autorisation de release.

| Lot | Verdict strict | Référence GitHub / commit | Limite principale |
|---|---|---|---|
| T28 | `BLOCKED` | audit indépendant `08107b3ef2d59c8054720fdf4dd81554d4984afb` | Contrat documentaire, pas d’extension ni autorisation produit |
| T29 | `BLOCKED` | audit indépendant `08107b3ef2d59c8054720fdf4dd81554d4984afb` | Contrat documentaire, SystemVault natif non testé |
| T30 | `PENDING_REMOTE_EVIDENCE_RECONCILIATION` | commit `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb`, [commit GitHub](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/commit/cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb) ; ZIP `c07321fbdf5f16948484264cf9677831cea6f3fd53ee54c0e72273ebea36304d`, bundle `c4f514ffe4bc24c3adfefff3cfb3a6b07db4a4e19c4c764c16bc9a395c867f14` | Commit accessible, mais aucune branche GitHub/kit T30 canonique rattaché à la clôture finale ; preuve distante à réconcilier |
| T31 | `APPROVED_VERIFIABLE_LOCAL` | `work/t31-canvas-webgl-audio` `9dad98703a748af61cbd3573b00e2bb036739a69` | Projections `UNSUPPORTED`, aucune observation réelle |
| T32 | `APPROVED_VERIFIABLE_LOCAL` | `work/t32-client-rects` `d7279e81dd724ba2278a65838bc65aaa16912007` | Projection `UNSUPPORTED`, aucune géométrie DOM |
| T33 | `APPROVED_VERIFIABLE_LOCAL` | `work/t33-synthetic-geolocation` `693632791041fde14db14ec8982b8bff1060a8d3` | Géolocalisation synthétique uniquement |
| T34 | `APPROVED_VERIFIABLE_LOCAL` | `work/t34-hardware-diagnostics` `c7f66da0da6f547813d10826dfd7772ad7e0f4b6` | Diagnostics projetés, aucun accès hôte |
| T35 | `APPROVED_VERIFIABLE_LOCAL` | `work/t35-font-bundle` `891b2d166a30a96d5ca872b473251ed1a3706ba3` | Aucun binaire ou font système fourni |
| T36 | `APPROVED_VERIFIABLE_LOCAL` | `work/t36-drift-detection` `099ff7bbf384ff900ec44fc3e9aafd1fc273f0dd` | Comparaison redacted uniquement |
| T37 | `APPROVED_VERIFIABLE_LOCAL` | `work/t37-profile-health` `c0ad051c77e153b3ec4435fbc7ff98e30b96969b` | Santé agrégée redacted, non observationnelle |
| T38 | `APPROVED_VERIFIABLE_LOCAL` | `work/t38-session-lifecycle` `f8fcfa02935a62fa66c3e22bb6bd5a01493f2f62` | Lifecycle mémoire local, aucun runtime |
| T39 | `BLOCKED` | `work/t39-secret-import-export-blocked` `08061f1e55a0bba01a4a28df66dc852e8f345ade` | T28/T29 non autorisés, SystemVault non testé |
| T40 | `BLOCKED` | `work/t40-integration-gated` `077ba90b2c522415aeefdc2d651c457b8c59683d` | Intégration runtime interdite par dépendances |
| T41 | `BLOCKED` | `work/t41-release-readiness-blocked` `944376fd1c9d22dad44730854ce4b2d6203c743b` | Release publique interdite ; tentative oversized conservée hors dépôt |
| T42 | `BLOCKED` | [work/t42-final-closure](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/work/t42-final-closure) `6489af39a4ac4f91f9f7dc1435f10b2bd10dfdc0` ; livraison [`evidence/T42_DELIVERY.md`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/work/t42-final-closure/evidence/T42_DELIVERY.md) | Clôture technique produite, clôture produit/release impossible |

> Les lots T31 à T38 disposent de résultats de tests et de bundles réellement conservés. Leurs journaux `BASELINE_DISCOVERY_RAW.log` historiques n’étaient pas présents dans la branche finale ; les fichiers `BASELINE_RECONSTRUCTION_POSTHOC_RAW.log` sont des reconstructions postérieures explicitement étiquetées et ne doivent pas être présentées comme des preuves contemporaines.

## Gates permanentes

Les valeurs suivantes restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Aucun runtime réel, Camoufox, proxy réel, cookie réel, migration utilisateur, récupération de géolocalisation réelle, SystemVault natif ou release n’a été exécuté.

Les archives originales, les copies de correction et les archives rejetées pour taille sont conservées. Les paquets GitHub sont des preuves auditables et ne constituent pas une approbation produit.


## Livraison prévalidation humaine T00–T42

- **Branche de livraison :** `audit/t00-t42-prehuman-validation`
- **Commit d’artefacts :** `cf280858b345e2fd566d391590f23d8cfa6bbe6d`
- **Dossier GitHub :** `evidence/PREHUMAN_T00_T42/`
- **ZIP :** `forgelocal-t28-t42-prehuman-validation.zip`
- **SHA-256 ZIP :** `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80`
- **Bundle :** `forgelocal-t28-t42-prehuman-validation.bundle`
- **SHA-256 bundle :** `14ef76cb68e7f64ff49fdc649cbcf96c5c69b0e9c410c5824a0592b7e33d1d14`
- **Décision :** `T00_T42_PREHUMAN_VALIDATION_DELIVERED_PENDING_INDEPENDENT_REVIEW`

Cette livraison fige la chaîne de preuve pour revue humaine. Elle ne constitue ni une release, ni une autorisation produit, ni une levée de gate. Les statuts T28, T29, T39, T40, T41 et T42 restent `BLOCKED`; T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`; T31–T38 restent `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION`.

Le journal brut de conservation ZIP/bundle est inclus dans le commit de livraison `cd6a95c66c029adf2140784491b06b8d9bf64fce`, qui est le head publié de la branche au moment du gel. Le bundle delta requiert explicitement la baseline ; le clone standalone sans baseline est conservé comme sortie attendue, tandis que la réhydratation seeded avec la baseline et le `git fsck --full` ont réussi.


## 2026-08-24 — Finalisation des findings T00–T42-PREHUMAN-FINDINGS-FINALIZATION

La dernière passe de prévalidation a individualisé les 13 findings GolangCI-Lint signalés comme nouveaux. Les 13 disposent maintenant d’une règle, d’un chemin, d’une ligne, du message brut, d’une sévérité explicitement déclarée non renseignée par le JSON du scanner, d’un lot ou rattachement prudent, d’un risque, d’un propriétaire et d’une condition de levée dans `evidence/PREHUMAN_T00_T42_FINAL_CHECKLIST/PREHUMAN_FINAL_FINDINGS_CLASSIFICATION.md`. Les 12 lignes situées dans des fichiers inchangés entre la baseline et HEAD sont classées comme différentiel scanner/contexte non réconcilié et ne sont pas présentées comme des régressions HEAD; le finding SA9003 de T38 est classé comme exception de test ouverte. Aucun code produit, test métier, configuration de lint ou gate n’a été modifié.

La classification inclut également les 36 findings Staticcheck historiques, les 6 misconfigurations Trivy historiques et l’inventaire de licences. Playwright/T10 dispose d’une preuve `NOT_APPLICABLE_UNDER_CURRENT_GATES` complète; aucun Core, token ou runtime réel n’a été créé ou lancé. La sortie attendue est `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW`, et non une approbation complète.

Un wrapper append-only est préparé sous `evidence/forgelocal-t00-t42-prehuman-final-review-wrapper-v2.zip` avec son sidecar portable. Il contient une copie intacte du ZIP historique `forgelocal-t28-t42-prehuman-validation.zip`, dont le hash reste `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80`, ainsi que l’addendum, les journaux bruts, la matrice, les SBOM, l’inventaire de licences et les classifications. Le hash du wrapper est joint dans son sidecar et sera confirmé avec le commit de publication. Les contrôles `unzip -t`, manifeste non auto-référentiel, extraction fraîche, `sha256sum -c` et re-scan Gitleaks d’extraction sont conservés dans l’addendum.

Les statuts restent strictement inchangés : T28, T29, T39, T40, T41 et T42 `BLOCKED`; T30 `PENDING_REMOTE_EVIDENCE_RECONCILIATION`; T31–T38 `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION`. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent maintenues.

## 2026-08-24 — Amendement senior : correction des findings réellement défectueux

La consigne de finalisation a été appliquée : les 13 findings GolangCI-Lint ont été analysés individuellement, puis corrigés lorsqu’ils révélaient une gestion d’erreur absente, un risque de transfert réseau silencieux, une gestion transactionnelle incomplète ou une assertion de régression vide. Les commits sources sont `6ee0840a7b264343be3840998df2a8903b511722` et `e0c9352710eb3710eaf0ea5d71614f2731a7051c`; ils sont rattachés à cette branche par `db0dd08` et `57d21b0`. La preuve ciblée confirme zéro finding ciblé résiduel.

La qualification post-correctif depuis un clone neuf au HEAD `e0c9352710eb3710eaf0ea5d71614f2731a7051c` a produit les logs `POSTFIX_QUALIFICATION_RAW.log` et `POSTFIX_TARGETED_REGRESSION_RAW.log`. `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `go mod verify`, OSV, govulncheck, Trivy, SBOM, Dashboard TypeScript/build/audit et le contrôle ciblé des 13 findings sont à zéro. Les findings historiques et les contrôles protégés restent classés explicitement; aucun scan n’est déclaré globalement vert.

La sortie est `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW_WITH_CODE_FIXES`. Les statuts T28/T29/T39/T40/T41/T42 `BLOCKED`, T30 `PENDING_REMOTE_EVIDENCE_RECONCILIATION` et T31–T38 `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION` restent inchangés. Les gates restent maintenues.

## 2026-08-24 — Publication du paquet code-fixed final

Le paquet code-fixed final est publié comme ajout append-only. Le wrapper `evidence/forgelocal-t00-t42-prehuman-final-review-wrapper-v3-code-fixed.zip` a le SHA-256 `5b31ae7afcdc032bec46785a2573ceab9ec797261c3e0d31490f3c6fb9dfbe2b`; son manifeste V3 contient 67 entrées et son scan Gitleaks source/extraction retourne un rapport vide avec code `0`. Il contient le ZIP historique original intact, dont le hash reste `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80`.

L’addendum code-fixed `evidence/forgelocal-t00-t42-final-exit-checklist-code-fixed.zip` a le SHA-256 `c7b83a2a55f7c623280fad574a6dfdeb0714ff8c51215b0bc33636b5c9daaa77`. Le correctif de code est porté par `6ee0840a7b264343be3840998df2a8903b511722` et `e0c9352710eb3710eaf0ea5d71614f2731a7051c`; les preuves post-correctif confirment zéro des 13 findings ciblés restants.

Le bundle de finalisation reste un delta nécessitant `631605ba136ff864d23e9674ca6adb4a8df0b740`; sa vérification seeded et `git fsck --full` sont à zéro. La décision est `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW_WITH_CODE_FIXES`, sans levée de gate ni déclaration de release.

## 2026-08-24 — Self-validation v4 avec E2E synthétique locale

Une auto-vérification renforcée a été exécutée depuis un clone neuf sur le HEAD `861880e56f13866346cf974110a01c8a890b86e2`. Elle couvre sidecars portables (31/31), ZIP et extractions fraîches (19/19), bundles (18/18), `git fsck --full`, LFS ciblé, qualification Go, Dashboard, Gitleaks explicite sur la plage demandée, OSV/Govulncheck, Trivy, SBOM CycloneDX/SPDX, licences et E2E Playwright synthétique loopback. L’E2E est vert : 1 test, 52 requêtes loopback, stockage navigateur vide, rejeu refusé et cleanup vérifié.

Le wrapper V3 historique reste inchangé et est inclus dans le wrapper v4 append-only. Les limites demeurent : Staticcheck 36 diagnostics historiques, Gosec 182 findings HEAD contre 194 baseline avec le même outil, golangci-lint 1.61.0 incompatible avec la cible Go 1.25, espaces finaux historiques dans `git diff --check` et huit objets LFS historiques indisponibles. La livraison ne lève aucune gate et ne constitue pas une release.

Statut exact : `T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW`.
T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`.

## 2026-08-24 — Livraison v4 append-only et branche dédiée

- **Branche :** `audit/t00-t42-self-validation-synthetic-e2e`
- **Commit d’artefacts et preuves :** `ad41afd71498fa5dda8eacc6a6ae0b47dbc865fd`
- **Commit des manifestes et bundle delta :** `c905f884ad9a84228985add6b7f77391e12b7b03`
- **Wrapper V4 :** `evidence/forgelocal-t00-t42-self-validation-v4-synthetic-e2e.zip`
- **SHA-256 wrapper V4 :** `429e683472b484076938d71428b5f52e1eb28794da1ad92526b670aa152f706b`
- **Bundle delta :** `evidence/forgelocal-t00-t42-self-validation-v4.delta.bundle`
- **SHA-256 bundle delta :** `0e4159703d453d8bea37617fe8e89460026b0a3118e57257d471af1777fd743e`
- **E2E :** 1 test Playwright, 52 requêtes loopback, stockage vide, rejeu refusé, cleanup PASS
- **Décision :** `T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW`

Les gates permanentes, T30 `PENDING_REMOTE_EVIDENCE_RECONCILIATION` et les findings historiques restent inchangés. Cette livraison n’est pas une release.

## 2026-08-24 — Référence distante finale de la livraison v4

La branche dédiée audit/t00-t42-self-validation-synthetic-e2e est publiée et sa référence distante vérifiée : 5e174dba6dddc35865f5bd943383d988ea12170c. Le wrapper V4 est evidence/forgelocal-t00-t42-self-validation-v4-synthetic-e2e.zip, avec SHA-256 f6544091783c2a4d4694d4b5f02c5dd5f0c70d22dab5efbb90abfd81418019bc. Le bundle delta final cible la révision 5e174dba6dddc35865f5bd943383d988ea12170c et requiert 861880e56f13866346cf974110a01c8a890b86e2 ; son SHA-256 est 10059c3c610d5a1b1ade88f936c8bb52ed893741a596326d8ba532f6f415e2fe.

Un clone neuf de cette branche retourne git fsck --full avec code 0. Le statut exact reste T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW. Les gates permanentes restent inchangées et aucune release n’est autorisée.

## 2026-08-24 — HEAD publié après synchronisation des preuves

La branche audit/t00-t42-self-validation-synthetic-e2e est maintenant au HEAD b4a04e4b9b489c22f3a86986c6faa1cbb9bf77c5. Ce commit synchronise les preuves de bundle final, manifestes et hashes avec le contenu publié. Le wrapper V4 et le bundle delta restent vérifiés par leurs sidecars ; aucun code produit, gate ou statut de release n’a été modifié.


## 2026-08-25 — Remédiation des findings réels et qualification V6

La branche dédiée `audit/t00-t42-v6-findings-remediation` part du HEAD V5 `b34fa5c02ff20144abfb5d240db1c67ad1f038f9`. Les deux findings GolangCI-Lint obligatoires ont été corrigés dans un commit séparé, les deux violations Axe ont été corrigées avec un test E2E non-régressif, et `golang.org/x/mod` a été mis à jour de `v0.37.0` à `v0.40.0` pour éliminer les deux advisories High Grype. Les 18 findings Semgrep ont été individualisés et qualifiés comme usages `crypto/rand`. Les 348 findings Gitleaks historiques et les 6 du checkout frais restent redacted, individualisés et classés comme empreinte publique PPA dans six chemins de provenance ; aucune allowlist globale n’est utilisée.

La requalification V6 confirme `go test -shuffle=on -count=3 ./...`, la variante race, `go vet`, `go build`, `govulncheck`, les deux Grype sur SBOM propres et l’E2E Axe loopback à zéro. Restent explicitement ouverts : 34 diagnostics Staticcheck historiques, 89 findings GolangCI-Lint historiques, 46 résultats OSV liés à la lecture `go 1.25.0` de la directive par OSV Scanner v1.9.2 malgré le toolchain effectif `go1.25.13`, six misconfigurations Docker Trivy, 741 licences `UNKNOWN`, la limite Gitleaks `0 commits scanned` sur `--log-opts` et 14 objets LFS historiques indisponibles pour `git lfs fsck`. Ces résultats sont conservés dans `V6_REMAINING_FINDINGS.md` et ne sont pas présentés comme un PASS global.

**Statut exact :** `T00_T42_V6_FINDINGS_REMEDIATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`. Les statuts T28, T29, T39, T40, T41 et T42 restent `BLOCKED`; T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangées.


## 2026-08-25 — Publication distante V6 vérifiée

La branche [audit/t00-t42-v6-findings-remediation](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/audit/t00-t42-v6-findings-remediation) a été publiée. Le HEAD distant vérifié avant cette entrée est `8e26bfb0c8bf6e92c09d645dd84ec854320c01f9`. Un clone neuf avec `GIT_LFS_SKIP_SMUDGE=1` correspond exactement à ce SHA ; `git fsck --full` retourne 0 avant et après le fetch LFS ciblé des deux artefacts V6. Le wrapper et le bundle ciblés correspondent respectivement à `ce722915d70e0aa528927b753c6f18efa5706fc9fa8703ef6f449b6728a5fab6` et `ad4484e795b80eb5b7655228012e695dc4b260d43057477a97ae145d164614c2`.

Le bundle delta publié exige `b34fa5c02ff20144abfb5d240db1c67ad1f038f9` et cible le commit de contenu `fc080456711dd7f2266911aaec55041fdb1b424c`; le HEAD de packaging est distinct par construction. Les gates et les lots bloqués restent inchangés.

## 2026-08-25 — Implémentation T28 Extensions locales contrôlées

La branche dédiée `feature/t28-local-extensions-controlled` implémente le contrat T28 depuis le tag V6 immuable `t00-t42-v6-local-qualified-2026-08-25` / baseline `999374d99b7996504ba91e421850a2fe84afb78d`. Le HEAD publié de l’implémentation et des corrections est `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578`.

Le verdict strict est `T28_IMPLEMENTED_VERIFIABLE_LOCAL_PENDING_INDEPENDENT_REVIEW`. Les tests ciblés T28 sous `-race`, `go vet` ciblé, `go build ./...`, `govulncheck ./...` et Gitleaks extraction/diff sont documentés ; la suite globale `go test -count=1 -race ./...` conserve le finding runtime V6 préexistant concernant la configuration BrowseForge Chromium/Docker-GHCR et ne doit pas être attribué à T28. Gosec ciblé `internal/extensions` est à zéro ; le package API conserve uniquement les findings historiques hors fichiers T28.

T28 reste strictement local-first : import ZIP explicite borné, extraction sans traversal/symlink, objet immuable, permissions 0700/0600, SQLite local, projection redacted, acknowledgement complet des permissions, high-risk explicite, affectation uniquement à un profil existant, update immuable, rollback, revoke/quarantine, purge contrôlée et audit sans package/token/chemin/secret. Aucun runtime d’extension n’est lancé.

Les statuts T29/T39/T40/T41/T42 restent `BLOCKED`; T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION` / revue indépendante selon son registre antérieur. Les gates permanentes restent inchangées et aucune release n’est autorisée.


## 2026-08-25 — T28-EVIDENCE-QUALIFICATION-R1

**Verdict probatoire :** `T28_EVIDENCE_QUALIFICATION_R1_READY_FOR_INDEPENDENT_REVIEW`.

La R1 a vérifié la baseline V6 `t00-t42-v6-local-qualified-2026-08-25` (`999374d99b7996504ba91e421850a2fe84afb78d`), la lignée historique demandée, les tests globaux exacts sur deux worktrees propres baseline/HEAD, les tests T28 ciblés sous race, OSV Scanner v1.9.2, Gitleaks sur huit commits réels et extraction ZIP, ainsi que Gosec baseline/head normalisé (`new_findings=0`, `resolved_findings=0`). Les 46 avis OSV par `go.mod` et côté baseline/head sont conservés sans faux PASS.

Le finding runtime historique `TestNewRegistryLoadsBrowseForgeChromiumFromDefaultConfig` n’a pas été reproduit sur les deux worktrees R1 propres ; aucune allowlist, skip, modification de test ou modification de code n’a été utilisée. Une éventuelle divergence avec des sorties historiques est documentée, et non transformée artificiellement en échec identique.

La valeur exacte maintenue est `camoflox_execution_authorized=false`. Un éventuel `update_url`/`updateURL` du manifest est ignoré, non suivi et non exécuté. T28 ne lance ni navigateur, ni extension, ni proxy, ni processus externe.

**Gates inchangées :** `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false`. T29, T39, T40, T41 et T42 restent bloqués et ne sont pas démarrés. T30 conserve son statut antérieur.


## 2026-08-25 — T28-FINAL-CLOSURE-PASS

La dernière passe ciblée a démontré puis corrigé un défaut concret d’intégrité : un blob ZIP modifié après import n’était pas revérifié avant approbation. Le repository vérifie désormais taille et SHA-256 avant `Approve`, `Assign` et `Rollback`, et l’API expose `INTEGRITY_MISMATCH` de façon stable. Les tests T28 supplémentaires couvrent ZIP corrompu, limites, toutes permissions sensibles/host patterns, `update_url` ignoré, purge lifecycle et package altéré ; tests ciblés sous race, vet ciblé, build et diff check retournent `0`.

**Statut proposé à l’acceptation propriétaire :** `T28_APPROVED_VERIFIABLE_LOCAL`.

Cette valeur signifie uniquement que T28 est implémenté et vérifiable localement, que son périmètre Core local est clôturé et que ses artefacts de conservation R1 sont acceptés. Elle ne signifie pas runtime navigateur approuvé, extension chargée/exécutée, SystemVault natif validé, proxy/cookies réels validés ou release publique autorisée.

Les gates restent strictement inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false`. T29, T39, T40, T41 et T42 ne sont pas démarrés.


## 2026-08-25 — T28-POSTFIX-FINAL-CLOSURE

Le défaut concret d’intégrité du blob après import a été corrigé dans `f0701da849ce0f9073397bb42ded5e2e76b29ef1`, puis couvert par les régressions `Approve`, `Assign` et `Rollback` dans `e806d4e915d1b702362389411d8ac823551df044`. La requalification ciblée est passée : tests T28 sous race, vet, build et diff check retournent `0` ; Gitleaks du diff non vide et de l’extraction finale retournent `0` sans leak.

Les artefacts post-correctif distincts sont publiés depuis `66f3bf09d3139e22f1885f7336c9edd879a32ede` : ZIP SHA-256 `4efda01771ed7af135769dfa68caa8bdc6f226ca7cad5bf894dcfce05f5c8923`, bundle SHA-256 `e9f65a5b9a734933f20ecf13b05f73136e604dc006b50dfc0d40286d91262097`. Ils ont passé sidecars distribués/neutres, extraction, manifeste/checksums, bundle verify, clone seedé et fsck.

**Statut exact :** `T28_APPROVED_VERIFIABLE_LOCAL`.

L’approbation reste limitée au Core local et aux artefacts de conservation post-correctif. Elle ne qualifie pas de runtime navigateur, extension chargée/exécutée, SystemVault natif, proxy/cookies réels ou release publique. Gates inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false`. T29, T39, T40, T41 et T42 restent non démarrés.
