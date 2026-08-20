# Registre canonique — T22 Profile History

## T27-R1 — Dossier de continuité canonique

| Champ | Valeur |
|---|---|
| Décision locale | `T27_R1_TECHNICAL_CLOSURE_APPROVED_VERIFIABLE_LOCAL_PENDING_CR08_NATIVE_VALIDATION` |
| Branche GitHub privée | `continuation-t27-r1-canonical` sur `davidwilsonbest89-afk/forgelocal-public-sanitized` |
| Commit distant | `282fb0a28bf48a15465341b02f82c83e09e2fd92` |
| Manifeste kit | SHA-256 `5e5defbb24b3314bde1c43802b3c81c900603823cac82ab702eb1ec7d3021dab` |
| Artefacts LFS | Bundle et ZIP CR-01, 308 Mo transférés et pointeurs Git vérifiés localement. |
| Contenu | Lignée CR-01→CR-05, bundles/sidecars/kits, `TOOLCHAIN.lock`, CDC, registre, checklist, politique baseline, SBOM/provenance CR-09 et guide de reprise. |
| Exclusions | Secrets, cookies, DB, profils, runtime, builds, `node_modules` et attestations privées hors périmètre. |
| Règle future | Aucun lot sans `BASELINE_DISCOVERY_RAW.log` avec commandes, chemins, UTC, exit codes et sorties brutes ; copie canonique hashée avant nettoyage. |

Les gates de release et CR-08 restent inchangés.

## T27-R1 — CR-01 Gate Camoufox/runtime

| Champ | Valeur |
|---|---|
| Baseline T26 | `930003ca95a934fd996c94ae897693ffb6be21fb` — `t26-simulated-proxy-provider-2026-08-20` |
| Fondation amendée | `982ca28bbb51db845fc1ea8b0d28812bd3bd272d` |
| Commit CR-01 | `df50a2b242f74558d2e405f099bdfe6db6437c9c` |
| Décision locale | `CR01_APPROVED_VERIFIABLE_LOCAL_WITH_HISTORICAL_SCAN_GATE` |
| Bundle canonique | `forgelocal-t27-r1-cr01-df50a2b.bundle` — SHA-256 `ab57d11f59fa0d13db7ded91e0a468413a094104ae8778761df919d103cbd271` |
| Kit canonique | `forgelocal-t27-r1-cr01-evidence-df50a2b.zip` — SHA-256 `b1cdcf44d081ef6e30f2f5524156765c88d1dc2a9e24a08048cad025935d741b` |
| Conservation | `/home/ubuntu/forgelocal-t27-r1-canonical/CR01/` ; sidecars portables, manifeste non auto-référentiel et scan Gitleaks extraction `[]`. |
| Réhydratation | `git bundle verify`, clone neuf à `df50a2b`, `git fsck --full` et test négatif CR-01 depuis le clone : PASS. |
| Qualification | Test négatif sans binaire/Playwright, `go test -count=1 -race ./...`, vet, build et diff-check : PASS. |
| Scans | Gitleaks delta `[]`; Trivy delta borné `secret` et `misconfig` : PASS. Les trois essais Trivy vulnérabilités globaux sont archivés comme `NOT_VERIFIED_TOOL_TIMEOUT`; aucune dépendance ne change, Govulncheck/OSV sont joints. |

CR-01 refuse Camoufox avant téléchargement, résolution de binaire, création de répertoire ou lancement. `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.

## T27-R1 — CR-02 Workflow désactivé et bind loopback

| Champ | Valeur |
|---|---|
| Baseline | `df50a2b242f74558d2e405f099bdfe6db6437c9c` — CR-01 canonique |
| Commit CR-02 | `d993289b415ff96ed68a1fde52b50d39db63c0c4` |
| Décision locale | `CR02_APPROVED_VERIFIABLE_LOCAL_WITH_HISTORICAL_SCAN_GATE` |
| Bundle canonique | `forgelocal-t27-r1-cr02-d993289.delta.bundle` — SHA-256 `7f81f07f78b4f852909a51522c26bd341a1473c5b54bfec911ae084f4d3be9fc` |
| Prérequis bundle | Bundle CR-01 au commit `df50a2b`; `git bundle verify`, fetch incrémental, checkout `d993289` et `fsck` : PASS. |
| Kit canonique | `forgelocal-t27-r1-cr02-evidence-d993289.zip` — SHA-256 `bf82242685a0b8ca2ebabd3371b878744f0915cd1d11ce3da961457a38b61bd9` |
| Conservation | `/home/ubuntu/forgelocal-t27-r1-canonical/CR02/` ; sidecars, manifeste non auto-référentiel, extraction et Gitleaks `[]` : PASS. |
| Qualification | Tests Go ciblés et complets `-race`, vet, build, diff-check ; refus réseau externe, écoute loopback, refus workflow HTTP/MCP : PASS. |
| Scans | Gitleaks delta explicite, Trivy `secret`/`misconfig` et OSV : PASS. Gosec SSA et Govulncheck terminés par l’outillage/sandbox et sont classés non qualifiés, non masqués. |

CR-02 ne lance aucun workflow et ne peut plus ouvrir de bind non loopback littéral. Les gates permanents demeurent inchangés.

## T27-R1 — CR-03 Redaction Dashboard, cookies et MCP

| Champ | Valeur |
|---|---|
| Baseline | `d993289b415ff96ed68a1fde52b50d39db63c0c4` — CR-02 canonique |
| Commit CR-03 | `a912cd6ee77ea853a1374e94a5d21932d04e0356` |
| Décision locale | `CR03_APPROVED_VERIFIABLE_LOCAL_WITH_HISTORICAL_SCAN_GATE` |
| Bundle canonique | `forgelocal-t27-r1-cr03-a912cd6.delta.bundle` — SHA-256 `58b7e604b8c09f1c643ce18472e3c6b582da47cccf36adee10ccf73cf8d66423` |
| Prérequis bundle | Bundle CR-02 au commit `d993289`; `git bundle verify`, fetch incrémental, checkout `a912cd6` et `fsck` : PASS. |
| Kit canonique | `forgelocal-t27-r1-cr03-evidence-a912cd6.zip` — SHA-256 `ea85737bdfb9bb4ba2cdd02bf5cb25abcdc556f7f202ceeb97e6d93f4bf3c535` |
| Conservation | `/home/ubuntu/forgelocal-t27-r1-canonical/CR03/` ; sidecars, manifeste non auto-référentiel, extraction et Gitleaks `[]` : PASS. |
| Qualification | Sentinelles Dashboard exécutées, MCP cookies refusé avant contexte, `-race`, vet, build et diff-check : PASS. |
| Scans | Gitleaks et Trivy `secret`/`misconfig` du delta : PASS. Dépendances inchangées ; limitations Gosec/Govulncheck de la lignée maintenues. |

CR-03 ne capture plus de corps réseau ni d’en-têtes non allowlistés et ne permet aucun accès MCP aux cookies. Les gates permanents demeurent inchangés.

## T27-R1 — CR-04 Chaîne d’approvisionnement Dashboard

| Champ | Valeur |
|---|---|
| Baseline | `a912cd6ee77ea853a1374e94a5d21932d04e0356` — CR-03 canonique |
| Commit CR-04 | `967d29311e7d39a9bf41cbaa9f0ec3c0e48e830a` |
| Décision locale | `CR04_APPROVED_VERIFIABLE_LOCAL_WITH_HISTORICAL_SCAN_GATE` |
| Audit pnpm | Avant : 2 critiques, 45 hautes, 74 modérées et 9 basses ; après régénération du lockfile : **0 vulnérabilité** sur 697 dépendances. |
| Bundle canonique | `forgelocal-t27-r1-cr04-967d293.delta.bundle` — SHA-256 `1693c40d214f659aad9fdef4366e22f40b7ac1e7967b618d472c5e55f3c32757` |
| Prérequis bundle | Bundle CR-03 au commit `a912cd6`; `git bundle verify`, fetch incrémental, checkout `967d293` et `fsck` : PASS. |
| Kit canonique | `forgelocal-t27-r1-cr04-evidence-967d293.zip` — SHA-256 `08c07a4c7923124eff18abd7c95718cce3bedcbb17e9163bfa43e57c636e1433` |
| Conservation | `/home/ubuntu/forgelocal-t27-r1-canonical/CR04/` ; sidecars, manifeste non auto-référentiel, extraction et Gitleaks `[]` : PASS. |
| Qualification | Installation `--frozen-lockfile --ignore-scripts`, TypeScript, build, audit zéro, sentinelle CR-03 et `git diff --check` : PASS. |
| Playwright | Les 22 scénarios sont découverts depuis un Core loopback `--no-runtime`; la suite E2E complète n’est pas rejouée dans ce lot de dépendances. |
| Scans | Gitleaks et Trivy `secret`/`misconfig` du delta : PASS. |

CR-04 ne contient ni `node_modules` ni build généré dans la conservation. Les gates permanents demeurent inchangés.

## T27-R1 — CR-05 API v1 et OpenAPI

| Champ | Valeur |
|---|---|
| Baseline | `967d29311e7d39a9bf41cbaa9f0ec3c0e48e830a` — CR-04 canonique |
| Commit | `9386e5c4fbc733fdda2b717ea9dd208c1da7872c` |
| Décision locale | `CR05_APPROVED_VERIFIABLE_LOCAL_WITH_HISTORICAL_SCAN_GATE` |
| Bundle delta | `forgelocal-t27-r1-cr05-9386e5c.delta.bundle` — SHA-256 `0b55e76745fd528cb16f05de5eb681770001299009d9d430b59341365258828f` |
| Kit | `forgelocal-t27-r1-cr05-evidence-9386e5c.zip` — SHA-256 `336280e8e4c51bd4e98d21f348bbaab433d986c4e90fa20bddee8dde73c1e1ac` |
| Vérifications | Alias `/api/v1`, OpenAPI, auth loopback, 404 route inconnue, `-race`, vet, build, diff-check, Gitleaks, bundle verify, clone neuf et `fsck` : PASS. |

CR-05 ne publie pas l’API et conserve les routes legacy sous leurs gardes existantes ; la façade v1 ajoute une garde avant délégation. Les gates permanents demeurent inchangés.

| Champ | Valeur |
|---|---|
| Jalon | T22 — Profile History |
| Commit | `5cfe7df3b5bb24c3d84ba455d3c32569555c4bdc` |
| Tag | `t22-profile-history-2026-08-19` |
| Baseline | `ee08ffd6d1f997c2f6e0fa8ffbbe783c83af80b8` |
| Bundle canonique | `forgelocal-core-t22-profile-history-5cfe7df.bundle` |
| SHA-256 bundle | Référencé par son sidecar portable |
| Kit canonique | `forgelocal-t22-profile-history-evidence-5cfe7df-r3.zip` |
| SHA-256 kit | `844168e6520815a1fd87894b45726bd1a840be14199f14795a7432a103106e0f` |
| Manifeste kit | Non auto-référentiel ; vérifié après extraction neuve |
| Clone neuf / `fsck` | PASS |
| Gitleaks delta | PASS, JSON `[]`, 57 298 octets de diff |
| Re-scan extraction | PASS, JSON `[]` avec allowlist strictement limitée à une empreinte runtime statique documentée |
| Gosec nouveau package History | 0 finding |
| Statut historique | `T22_PROFILE_HISTORY_EVIDENCE_READY_FOR_INDEPENDENT_AUDIT` |

Le kit est conservé sous `/home/ubuntu/forgelocal-t22-delivery/` avec son sidecar externe, son bundle et les logs de re-scan. Aucun lot ultérieur n’est autorisé par ce registre.

## Correction de cohérence et tests — référence canonique de clôture

| Champ | Valeur |
|---|---|
| Sous-lot | `T22-CONSISTENCY-AND-TESTS` |
| Baseline immuable | `5cfe7df3b5bb24c3d84ba455d3c32569555c4bdc` (`t22-profile-history-2026-08-19`) |
| Commits de correction | `7a95c616d7d80ae3c7a64559544c9866b28758c7`, `78ac349784e4d6a9d1bb27fc0449759b09f6c6a5`, `e4a26fede014210ff959d27d8d85a979ef92cfbf` |
| Tag final | `t22-consistency-and-tests-final-2026-08-19` → `e4a26fede014210ff959d27d8d85a979ef92cfbf` |
| Bundle canonique final | `forgelocal-core-t22-consistency-and-tests-e4a26fe.bundle` |
| SHA-256 bundle | `2cb3d8dd0f4ebb4f373c7b455717710961a558dae9d258009508c27adc83279e` ; sidecar portable adjacent |
| Réhydratation | `git bundle verify`, clone neuf au tag final et `git fsck --full` : PASS |
| Qualification clone neuf | `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check` : PASS |
| Gitleaks delta | Arbre exact des 23 fichiers `5cfe7df..e4a26fe`, JSON `[]`, exit 0 |
| Gosec base→head | 59 signatures règle/fichier de part et d’autre, aucun finding nouveau ; `SCAN_BLOCKED_UNKNOWN` historique maintenu |
| Statut de clôture locale | `T22_CONSISTENCY_AND_TESTS_APPROVED_VERIFIABLE_LOCAL` |

Le correctif n’affirme pas une transaction ACID inter-store. Il apporte un marqueur `HistoryPending` durable, une reprise déterministe au démarrage et des tests de panne après écriture Profile, redaction proxy, pagination, lectures, guards et concurrence. Les invariants restent `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. **T23 ne démarre pas sans instruction explicite.**

## Correction d’opération pending — clôture du blocage de concurrence

| Champ | Valeur |
|---|---|
| Baseline de correction | `e4a26fede014210ff959d27d8d85a979ef92cfbf` |
| Commit correctif | `d428690aabe1032ad0d3073e3f809d1df8dbfb29` |
| Tag | `t22-pending-operation-correction-2026-08-19` |
| Protocole | Marqueur owner-only `operation_id`, `action`, `created_at`, `snapshot_digest`; version SQLite liée au même `operation_id`; clear conditionnel |
| Sérialisation | Verrou distinct par profil couvrant mutation handler → capture/recovery SQLite → clear conditionnel |
| Tests déterministes | Entrelacement capture A / mutation B / clear A / reprise; restauration A / mutation B / clear A; échec de clear; verrou de séquence |
| Bundle | `forgelocal-core-t22-pending-operation-d428690.bundle` |
| SHA-256 bundle | `8496861f9c43cb31a29ab41c205305b4add94923af53ab83e2842121193b2524` |
| Kit canonique | `forgelocal-t22-pending-operation-d428690-kit.zip` |
| SHA-256 kit | `cdc34d2a9ff07a6db0ea8adfdea6479579d74ad5a9ea30cc3f4ce8a265b0862c` ; sidecar portable adjacent |
| Manifeste et re-scan kit | 16 entrées non auto-référentielles, extraction neuve vérifiée ; Gitleaks JSON `[]` |
| Requalification | Bundle verify, clone neuf, `fsck`, `-race`, vet, build, diff-check et absence d’espaces finaux : PASS |
| Scans | Gitleaks delta JSON `[]`; Gosec base→head sans nouveau finding règle/fichier |
| Statut local | `T22_PENDING_OPERATION_CORRECTION_APPROVED_VERIFIABLE_LOCAL` |

Cette correction résout le risque où le clear de l’opération A pouvait supprimer l’évidence durable de l’opération B. Les invariants et l’interdiction de démarrer T23 sans instruction explicite restent inchangés.

## Règle permanente de conservation et de reprise

Toute future ouverture de lot doit commencer par une section **`BASELINE_DISCOVERY`** dans son addendum ou son rapport de conservation. Cette section est un prérequis de modification, non une annexe facultative. Elle doit fournir les commandes complètes, les chemins consultés, les dates UTC de début et de fin, les codes de sortie, les sorties brutes et les références qualifiées (commit, tag, bundle, sidecar et hash) utilisés pour créer le worktree de travail.

Une copie canonique hashée de tout kit livré doit être conservée avec un sidecar externe portable qui ne contient que le hash et le nom de fichier. Avant toute clôture, le ZIP, son manifeste, son sidecar et l’extraction neuve doivent être vérifiés. La procédure normative de reprise est conservée dans `BASELINE_DISCOVERY_POLICY.md` à côté de ce registre.

### Copie canonique T22 actuellement conservée

| Élément | Référence |
|---|---|
| Répertoire immuable de conservation | `/home/ubuntu/forgelocal-t22-canonical/` |
| Index de reprise | `CANONICAL_INDEX.md` |
| Journal brut de découverte et vérification | `CANONICAL_BASELINE_DISCOVERY.log` |
| Kit et sidecar portable | `forgelocal-t22-pending-operation-d428690-kit.zip` ; SHA-256 `cdc34d2a9ff07a6db0ea8adfdea6479579d74ad5a9ea30cc3f4ce8a265b0862c` |
| Bundle et sidecar | `forgelocal-core-t22-pending-operation-d428690.bundle` ; SHA-256 `8496861f9c43cb31a29ab41c205305b4add94923af53ab83e2842121193b2524` |
| Vérification | Sidecars, intégrité ZIP et copies des rapports : PASS |

## T23 — Archivage et restauration logique

| Champ | Valeur |
|---|---|
| Baseline | `d428690aabe1032ad0d3073e3f809d1df8dbfb29` (`t22-pending-operation-correction-2026-08-19`) |
| Commit | `be1d68341e601e406edc18d0763c466966105a4b` |
| Tag | `t23-archive-restore-2026-08-19` |
| Contrat | `docs/T23_ARCHIVE_RESTORE_CONTRACT.md`, avec section `BASELINE_DISCOVERY` et sortie brute associée |
| Bundle | `forgelocal-core-t23-archive-restore-be1d683.bundle` |
| SHA-256 bundle | `ab475a9434494ba81a963dd420b9a80250bad1dee83769ecbfefe1a9d707e565` |
| Kit | `forgelocal-t23-archive-restore-be1d683-kit.zip` |
| SHA-256 kit | `43c7e876225536b97502a55c6e47c8ac2d31581eb86cde94f8072b2eafcbb145` ; sidecar portable adjacent |
| Manifeste et re-scan | 18 entrées non auto-référentielles, extraction neuve vérifiée ; Gitleaks JSON `[]` |
| Qualification | Clone neuf, `fsck`, `-race`, vet, build et diff-check : PASS |
| Scans | Gitleaks delta `[]`; Gosec sans nouveau finding règle/fichier |
| Statut | `T23_ARCHIVE_RESTORE_APPROVED_VERIFIABLE_LOCAL` |

Le lot rend l’archive/reopen persistante avant publication mémoire, évite les double-captures idempotentes et réutilise le protocole pending-opération T22. Aucun T24 n’est autorisé sans instruction explicite.

### Régularisation de preuve T23

| Champ | Valeur |
|---|---|
| Commit | `df969533cdd446be41d868bebea2c8a106f543d5` |
| Tag | `t23-scope-and-evidence-correction-2026-08-19` |
| Objet | Autorisation explicite, tests complémentaires, logs autonomes et artefacts de transport ; aucun changement métier T23 |
| Bundle | `forgelocal-core-t23-scope-evidence-df96953.bundle` |
| SHA-256 bundle | `e679fe2898628a814bffb2aa0254f0792a2a2b1a137d320dfbd4a95950de3fd6` |
| Kit de révision | `forgelocal-t23-scope-evidence-df96953-kit.zip` |
| SHA-256 kit | `a42f1a67919ffa2f8fff76d0e1c9f2f01045acd7d4f28413bac7ac7c8f5e9dd0` ; sidecar portable séparé vérifié par `sha256sum -c` |
| Manifeste et re-scan | 28 entrées non auto-référentielles, extraction neuve, bundle inclus vérifié et Gitleaks JSON `[]` |
| Ratification métier postérieure | `T23_SCOPE_RATIFICATION.md`, `2026-08-19T23:02:23Z`, sans antidatation ni réécriture de la chronologie Git |
| État | `T23_SCOPE_AND_EVIDENCE_CORRECTION_READY_FOR_INDEPENDENT_REVIEW` |

Le statut T23 initial ne doit pas être tenu pour approuvé avant revue du kit de régularisation. T24 reste interdit sans instruction explicite.

### Paquet canonique de ratification et distribution

| Champ | Valeur |
|---|---|
| Paquet | `forgelocal-t23-ratification-distribution-20260819T230223Z.zip` |
| SHA-256 | `01357a92acd03777ab87871bed58c2389987d846cca09f2239faec321078a5ff` |
| Sidecar | Portable, externe et vérifié par `sha256sum -c` |
| Contenu | Ratification non antidatée, contrôle de distribution, sidecar du kit T23 et `BASELINE_DISCOVERY.log` |
| Manifeste | 6 entrées non auto-référentielles, vérifiées après extraction neuve |
| Copie canonique | `/home/ubuntu/forgelocal-t22-canonical/T23_SCOPE_AND_EVIDENCE_CORRECTION/` |

### Décision finale locale T23

| Champ | Valeur |
|---|---|
| Décision | `T23_CLOSED_RATIFIED_VERIFIABLE_LOCAL` |
| Justification | Sidecars externes reçus et vérifiés, ratification propriétaire postérieure non antidatée, paquets hashés et copies canoniques vérifiés |
| Décision formelle | `T23_FINAL_LOCAL_CLOSURE.md` |
| SHA-256 de la décision | `43cbe44c6583c2de4c6cdd24e9d2082b182bb0412229854ee8f6a1a1812350ca` ; sidecar portable vérifié |
| Réserve historique | L’autorisation produit n’était pas distribuée avant le commit métier ; la ratification postérieure est conservée honnêtement et ne prétend pas modifier cette chronologie |

T24 reste non démarré et exige une instruction séparée.

## Wrapper de continuation T00–T23

| Champ | Valeur |
|---|---|
| Wrapper | `forgelocal-continuation-t00-t23-20260819.zip` |
| SHA-256 | `0046e99e22ad1498a6307f4acd43d1cebbb7c317d47edee16329fc38b59b924d` |
| Manifeste | 75 entrées, non auto-référentiel, vérifiées en flux après construction |
| Reprise prouvée | Bundle T23 inclus, sidecar, `git bundle verify`, clone neuf au tag T23, `git fsck --full`, `go test -race`, vet, build et diff-check : PASS |
| Contenu | Kit privé T00–T21 R2 immuable, copies canoniques T22/T23, CDC, registre, politique BASELINE_DISCOVERY, procédures et bilan honnête |
| Limites | Les gates de release, secrets réels et preuves natives externes ne sont pas inclus ni déclarés levés ; voir `CONTINUATION_HONEST_COVERAGE.md` |

### Miroir GitHub privé de continuation

| Champ | Valeur |
|---|---|
| Dépôt privé | `davidwilsonbest89-afk/forgelocal-continuation-t00-t23-private` |
| URL | `https://github.com/davidwilsonbest89-afk/forgelocal-continuation-t00-t23-private` |
| Commit de documentation | `1ad79b56c0b19bdb6dc054b35d353fbbcaae2802` |
| Release | `t00-t23-continuation-20260819` |
| Assets | 8 assets privés : trois morceaux de wrapper, sidecar, manifeste et guides ; présence et tailles vérifiées via GitHub CLI |
| Confidentialité | Dépôt vérifié `private: true` ; ne pas rendre public en raison des documents T07 privés |

### Publication publique assainie

| Champ | Valeur |
|---|---|
| Dépôt public | `davidwilsonbest89-afk/forgelocal-public-sanitized` |
| URL | `https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized` |
| Commit | `14ec742f3be3df1028431eedc672d7201557f588` |
| Contenu | Code ForgeLocal T23, dépendances, scripts non T07, Dashboard et documentation publique sélectionnée |
| Exclusions contrôlées | Aucun fichier ou contenu T07/Camoufox, archive privée, secret, cookie, session, `.env`, SQLite, DB, ZIP ou bundle |
| Scan | Clone neuf public : Gitleaks `[]`, exclusion de chemins et contenus : PASS |
| Première tentative | `forgelocal-public` conservé **privé** après détection de références T07/Camoufox ; ne pas publier ce dépôt |

### Fenêtre publique temporaire de continuité complète T00–T23

| Champ | Valeur |
|---|---|
| Dépôt ciblé | `davidwilsonbest89-afk/forgelocal-public-sanitized` |
| Release | `t00-t23-public-continuity-window-20260820` |
| Ouverture UTC | `2026-08-20T00:43:00Z` |
| Retour privé UTC | `2026-08-20T00:48:02Z` |
| Fenêtre prévue | 300 secondes |
| Assets | 7 : deux archives de continuité, leurs sidecars, notes de portée et preuves de scan T22–T23 |
| Archive T00–T21 | `forgelocal-continuation-t00-t21-public-evidence-20260820.zip` ; SHA-256 `954fef1488b0e2cb365cb2b9e9147d2163534fffeb5ec3092c2c9dff6150d46b` |
| Archive T22–T23 | `forgelocal-continuation-t22-t23-public-evidence-20260820.zip` ; SHA-256 `563c4dbb37108e0944f8abe147be9de833237bd414d55ac47a6c188b01e5acb7` |
| Exclusions explicites | Attestations, déclarations d’attestation et certificats uniquement ; le paquet T07 de clôture a été dérivé sans ces documents, les autres sources/archives T07 autorisées restant incluses |
| Contrôle technique | Gitleaks de la fondation T00–T21 et de la copie canonique T22–T23 : `[]` / `no leaks found` |
| Visibilité finale | Contrôlée après retour : `private: true` |

Cette fenêtre de transport ne constitue pas une levée de gate ni une autorisation produit. Les invariants `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` sont expressément maintenus.
