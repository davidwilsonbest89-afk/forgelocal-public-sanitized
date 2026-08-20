# Audit indépendant — ForgeLocal R4

- [x] T27-R1-FOUNDATION : AM-01 à AM-10 intégrés dans le contrat d’exécution, `TOOLCHAIN.lock` réconcilié depuis les exécutables relevés et BASELINE_DISCOVERY T26 fraîche préservée avant code.
- [x] CR-01-RUNTIME-GATE : Camoufox bloqué fail-closed ; refus sans binaire/Playwright, API, MCP, téléchargement et bootstrap vérifiés au commit `df50a2b`.
- [x] CR-01-SCAN-RECOVERY : trois tentatives Trivy globales conservées comme non qualifiées ; Trivy `secret`/`misconfig` borné sur le delta passe, aucune dépendance ne change et Govulncheck/OSV sont joints.
- [x] CR-01-CANONICAL-PRESERVATION : bundle, sidecars, clone neuf, `fsck`, kit, manifeste non auto-référentiel, extraction et registre produits sous `/home/ubuntu/forgelocal-t27-r1-canonical/CR01/`.
- [x] CR-02-WORKFLOW-LOOPBACK : workflow HTTP/MCP désactivé fail-closed, bind loopback littéral imposé, tests réseau/auth/origine et conservation delta CR-01→CR-02 qualifiés au commit `d993289`.
- [x] CR-03-REDACTION-COOKIES-MCP : collecteur Dashboard sans corps ni headers libres, cookies MCP désactivés fail-closed, sentinelles synthétiques et conservation delta CR-02→CR-03 qualifiées au commit `a912cd6`.
- [x] CR-04-DASHBOARD-SUPPLY-CHAIN : lockfile Dashboard régénéré, audit pnpm ramené de 2 critiques/45 hautes à 0 vulnérabilité, build/TypeScript/sentinelle qualifiés et conservation delta CR-03→CR-04 au commit `967d293`.
- [x] CR-05-API-V1-OPENAPI : façade `/api/v1` protégée, index OpenAPI et comparaison routeur vivant qualifiés au commit `9386e5c`; bundle delta et clone neuf vérifiés, kit d’évidence complet restant à assembler.
- [x] CR-06-EXTERNAL-STATE-MACHINE : machine d’état documentaire externe créée sans écriture de store.
- [x] CR-07-WRITTEN-AUTHORIZATION : autorisation déjà validée par le propriétaire et référencée par la décision T27-R1 ; elle ne modifie pas les interdictions runtime, proxy réel, cookies réels ou release.
- [x] CR-08-NATIVE-VALIDATION : statut `NOT_TESTED_NATIVE_ENVIRONMENT_REQUIRED` consigné ; aucun coffre natif touché.
- [x] CR-09-CI-PROVENANCE-PUBLICATION : plan CI/provenance/notices documenté ; aucune publication, secret CI ou release configuré.
- [ ] T27-R1-CR09-CLOSURE-TOOLCHAIN : rejouer CI locale isolée, produire SBOM CycloneDX/SPDX avant/après, inventorier notices/licences et signer une attestation de provenance sans publication.
- [x] T27-R1-CR07-AUTHORIZATION-REVIEW : autorisation existante reconnue dans la décision T27-R1 ; aucune nouvelle autorisation requise pour la clôture documentaire/technique locale.
- [ ] T27-R1-CR09-LOCAL-CLOSURE : produire SBOM, inventaire licences/notices, provenance locale et kit de clôture hashé ; aucune publication.
- [x] T27-R1-CR09-LOCAL-CLOSURE : SBOM CycloneDX/SPDX, provenance locale, manifestes et Gitleaks produits ; aucune publication.
- [ ] T27-R1-CANONICAL-CLOSURE-PACKAGE : créer ZIP canonique, sidecar, manifeste non auto-référentiel, extraction neuve, re-scan et inscription au registre.
- [ ] T27-R1-CONTINUATION-KIT : assembler code, bundles, kits, handover, changelog, registre, TOOLCHAIN.lock, CDC et procédures de reprise ; exclure secrets, cookies, DB, runtime et attestations privées hors périmètre.
- [ ] T27-R1-GITHUB-CONTINUITY-SYNC : synchroniser le kit vérifié sur le dépôt GitHub associé, vérifier le commit distant et documenter la branche/tag de reprise.
- [ ] FUTURE-LOT-BASELINE-DISCOVERY-REQUIRED : chaque futur lot doit commencer par `BASELINE_DISCOVERY_RAW.log` contenant commandes, chemins, UTC, codes de sortie et sorties brutes avant toute écriture de code.
- [ ] T27-R1-CR08-NATIVE-ENVIRONMENT-REVIEW : exécuter uniquement dans un environnement natif admis avec procédure SystemVault dédiée ; la sandbox ne peut pas lever ce gate.
- [ ] T27-R1-CORRECTIVE-CLOSURE : qualifier chaque CR, préserver bundles/kits/copies canoniques et ne commencer CR-05 qu’après clôture CR-01 à CR-04.

- [x] AVIS-VALIDATION-CDC-CORRECTIF : AM-01 à AM-09 retenus comme obligatoires ; AM-10 retenu sous réserve que les versions exactes soient vérifiées depuis une toolchain qualifiée.
- [x] AVIS-VALIDATION-CDC-CORRECTIF : séquencement validé sous statut `APPROVED_WITH_REQUIRED_AMENDMENTS`; aucune implémentation corrective avant intégration textuelle et BASELINE_DISCOVERY T26 fraîche.

- [x] CDC-CORRECTIF-POST-AUDIT : cahier des charges P0/P1 rédigé pour Camoufox, workflow, loopback, logs, cookies/MCP, dépendances dashboard, API et architecture SQLite.
- [x] CDC-CORRECTIF-POST-AUDIT : lots CR-01 à CR-08, critères d’entrée/sortie, tests dynamiques, scans, preuves et gates définis sans code.
- [x] CDC-CORRECTIF-POST-AUDIT : cahier livré, SHA-256 `c5cdd0f7…`, Gitleaks `[]`; aucun T27 n’est démarré par cette rédaction.

- [x] AUDIT-CONFORMITE-CDC-CONTREVERIFICATION : constats rapprochés du commit public T23 et de la lignée canonique T24–T26.
- [x] AUDIT-CONFORMITE-CDC-CONTREVERIFICATION : Camoufox, stores hybrides, workflow, collecteur, cookies/MCP, dépendances dashboard et CI/provenance vérifiés ; limites de preuve consignées.
- [x] AUDIT-CONFORMITE-CDC-CONTREVERIFICATION : verdict rendu sans modification de code : audit public T23 substantiellement valide, avec statut post-T23 et prochaine étape datés.

- [x] T00-T26-SENIOR-CONTINUITY-AUDIT : bundles, kits, sidecars, registres, CDC, politiques, rapports, procédures de reprise et branches privées accessibles inventoriés.
- [x] T00-T26-SENIOR-CONTINUITY-AUDIT : chaîne T23→T24-SR→T25→T26 vérifiée depuis bundle, clone neuf, tags et `git fsck --full`; tests `-race`, vet et build T26 rejoués.
- [x] T00-T26-SENIOR-CONTINUITY-AUDIT : README de reprise, BASELINE_DISCOVERY, logs bruts, manifestes et limites contrôlés ; lacune du README historique corrigée par handover dédié et script de replay.
- [x] T00-T26-SENIOR-CONTINUITY-AUDIT : verdict, manifeste, scan, copie canonique hashée et handover versionné dans la branche privée `continuity-t00-t26-handover`.
- [x] GITHUB-PRIVATE-T00-T26-HANDOVER-SYNC : handover, script et manifeste primaire poussés au commit `1a5112b6e9b0206c30cebe6fc5d37e820555f3e5`; branche privée vérifiée.

- [x] GITHUB-PRIVATE-T24-LINEAGE-SYNC : lignée T24/T24-SR publiée sur la branche privée `continuity-t24-t24sr` (`7a375cc`), tags poussés, release privée `t24-sr-dependency-remediation-2026-08-20` avec 8 assets vérifiés.
- [x] T25-COOKIE-IMPORT-EXPORT : BASELINE_DISCOVERY exécutée ; contrat local de fixtures synthétiques établi avec validation stricte, redaction, limites, atomicité et exclusions de cookies réels/secrets.
- [x] T25-COOKIE-IMPORT-EXPORT : unités, API loopback dynamique, concurrence, `-race`, Gitleaks, Trivy, Gosec delta, Govulncheck, OSV, SBOM, bundle, clone neuf et copie canonique qualifiés ; statut `T25_APPROVED_VERIFIABLE_LOCAL`.
- [x] T26-PROXY-PROVIDER-INTEGRATION : BASELINE_DISCOVERY après T25 exécutée ; fournisseur local simulé à référence de secret seule, sans connexion fournisseur/proxy réel, qualifié par tests dynamiques et copie canonique.
- [x] T26-PROXY-PROVIDER-INTEGRATION : branche, tag et release privés synchronisés ; T26 est `T26_APPROVED_VERIFIABLE_LOCAL` et T27 reste interdit sans instruction distincte.

- [x] BASELINE-DISCOVERY-PERMANENT-GOVERNANCE : règle obligatoire intégrée au commit `7a375cc`, avec commandes, chemins, UTC, codes de sortie et sorties brutes requis pour chaque futur lot.
- [x] BASELINE-DISCOVERY-PERMANENT-GOVERNANCE : bundle, kit, sidecars, manifeste, copie canonique hashée et registre de continuité préservés.
- [x] GITHUB-PUBLIC-MAINTAINED-ON-OWNER-INSTRUCTION : dépôt `forgelocal-public-sanitized` repassé en privé sur instruction explicite ; `isPrivate: true` vérifié.
- [x] GITHUB-PRIVATE-BASELINE-POLICY-SYNC : politique et référence README intégrées et poussées au commit distant `f09af4b2a5c6bf6914a214e1220a0353c09a6fb0` ; visibilité privée vérifiée.

- [x] T24-SR-DEPENDENCY-REMEDIATION : BASELINE_DISCOVERY T24 auditée vérifiée par sidecar, bundle, clone, tag, HEAD, `fsck` et journal brut.
- [x] T24-SR-DEPENDENCY-REMEDIATION : Chi mis à niveau vers `v5.3.1` et x/net vers `v0.58.0`, avec régénération déterministe de `go.sum` ; aucun code métier modifié.
- [x] T24-SR-DEPENDENCY-REMEDIATION : `-race`, vet, build, diff-check, Gitleaks, Trivy, Gosec, Govulncheck ciblé, OSV-Scanner et SBOM rejoués ; dépendances vulnérables précédentes absentes des résultats ciblés.
- [x] T24-SR-DEPENDENCY-REMEDIATION : bundle, kit, sidecars, manifeste, copie canonique et verdict T24 révisé préservés ; T25 reste interdit sans instruction distincte.

- [x] T24-SENIOR-INDEPENDENT-AUDIT : kit, sidecars, manifeste, bundle, clone neuf, tag et `git fsck --full` vérifiés depuis des répertoires d’audit neufs.
- [x] T24-SENIOR-INDEPENDENT-AUDIT : diff T23→T24 et comportements bulk, gardes, concurrence, annulation, History et redaction rejoués sans modifier le code métier.
- [x] T24-SENIOR-INDEPENDENT-AUDIT : Gitleaks/Trivy secrets passent ; Govulncheck ciblé et OSV sont concluants et signalent la remédiation de dépendances requise ; Gosec/Staticcheck restent documentés comme dette historique hors delta.
- [x] T24-SENIOR-INDEPENDENT-AUDIT : verdict documenté et registre mis à jour ; T25 reste interdit sans instruction distincte.
- [x] T24-DEPENDENCY-SECURITY-REMEDIATION : remédiation T24-SR livrée et qualifiée ; T24 est `T24_APPROVED_VERIFIABLE_LOCAL`.

- [x] T24-BULK-OPERATIONS : BASELINE_DISCOVERY exécutée depuis le bundle T23 final ; sidecar, bundle verify, clone neuf, tag, HEAD et `fsck` consignés avant écriture.
- [x] T24-BULK-OPERATIONS : contrat Core local fermé défini pour archivage/réouverture en lot, changement de groupe et ajout/retrait de tags ; runtime, proxy réel, secrets réels, Dashboard, cloud et release exclus.
- [x] T24-BULK-OPERATIONS : limites, no-op, sérialisation par profil, annulation, erreur par profil, audit redacted et capture History durable implémentés au commit `d811a5c`.
- [x] T24-BULK-OPERATIONS : unités, API loopback, refus hors loopback/origine/authentification, concurrence, History unique, `-race`, vet, build, diff-check, Gitleaks, Trivy secrets, Gosec delta, SBOM, bundle et clone neuf qualifiés.
- [ ] T24-BULK-OPERATIONS : rejouer Govulncheck et OSV-Scanner sur un hôte à ressources suffisantes ; la sandbox a interrompu les deux scans (`143`) avant résultat, statut T24 maintenu `PENDING_SECURITY_REVIEW`.
- [x] T24-BULK-OPERATIONS : kit, sidecars, manifeste, preuves brutes et copie canonique préservés ; T25 interdit sans instruction distincte.

- [x] SANDBOX-TOOLCHAIN-QUALIFICATION : compilateurs, analyseurs, scanners, SQLite, E2E, charge, SBOM, API et conteneur local inventoriés.
- [x] SANDBOX-TOOLCHAIN-QUALIFICATION : outils de qualification locale manquants installés (Go tooling, SQLite, HTTPie, Trivy, OSV-Scanner, Syft, k6, Podman, axe, Lighthouse CI, tcpdump, mitmproxy et linters), sans les présenter comme qualification native de production.
- [x] SANDBOX-TOOLCHAIN-QUALIFICATION : versions exécutables et limites de sandbox vérifiées ; rapport `SANDBOX_TOOLCHAIN_QUALIFICATION.md` hashé et scanné par Gitleaks `[]`.

- [ ] GITHUB-PUBLIC-MAINTAINED-ON-OWNER-INSTRUCTION : dépôt `forgelocal-public-sanitized` à maintenir public jusqu’à instruction explicite du propriétaire demandant le retour en privé.

- [x] CDC-CONSOLIDATED-DELIVERY : version v3.9.7 localisée dans quatre copies concordantes (SHA-256 `b1dddc4f…`), intégrité et lignée documentaire vérifiées.
- [x] CDC-CONSOLIDATED-DELIVERY : cahier rapproché de l’état réellement prouvé T00–T23, avec distinctions explicites entre preuves locales, revue restante et gates de production non levés.
- [x] CDC-CONSOLIDATED-DELIVERY : cahier des charges complet consolidé livré avec addendum de continuité T00–T23, registre canonique, politique BASELINE_DISCOVERY, manifeste 6/6 et scan Gitleaks `[]`.

- [ ] GITHUB-PUBLIC-FIVE-MINUTE-WINDOW : publier temporairement cinq minutes les preuves et archives autorisées, puis rétablir le dépôt en privé ; exclure attestations, certificats et valeurs techniques réelles.

- [ ] GITHUB-PUBLIC-FULL-CONTINUITY : publier toutes les preuves et archives T00–T23, y compris les sources T07, sauf les attestations et certificats ; exécuter uniquement le contrôle anti-fuite de secrets techniques réels.

- [ ] GITHUB-PUBLIC-CONTINUITY-EVIDENCE : publier les preuves T00–T23 non sensibles, en excluant attestations et archives privées T07, sources T07 privées, secrets, cookies, sessions, bases et artefacts sensibles.
- [ ] GITHUB-PUBLIC-CONTINUITY-EVIDENCE : inventorier, scanner, publier et vérifier le paquet de preuves public avant toute nouvelle exposition.

- [x] GITHUB-PUBLIC-REMEDIATION : première tentative retirée de l’exposition en basculant `forgelocal-public` en privé ; nouveau dépôt public assaini publié depuis un historique neuf.
- [x] GITHUB-PUBLIC-SANITIZED : dépôt distinct public créé avec code T23 et documentation publique sélectionnée ; T07 privé, archives, cookies, sessions, secrets et preuves sensibles exclus.
- [x] GITHUB-PUBLIC-SANITIZED : clone neuf du dépôt public audité : exclusions de noms/contenus et Gitleaks `[]` PASS.

- [x] GITHUB-PRIVATE-CONTINUATION : compte `davidwilsonbest89-afk` vérifié, dépôt privé créé et release T00–T23 publiée sans exposition publique.
- [x] GITHUB-PRIVATE-CONTINUATION : wrapper réparti en trois assets de 1 Go maximum, avec sidecar, manifeste de release et instructions de reconstitution vérifiés.

- [ ] CONTINUATION-T00-T23-SIZE : mesurer le poids de chaque composant, distinguer les duplications de conservation et proposer une variante légère sans perte de source T23.

- [x] CONTINUATION-T00-T23 : archives, bundles, sources, CDC, registre, logs et procédures accessibles inventoriés ; limites déclarées dans `CONTINUATION_HONEST_COVERAGE.md`.
- [x] CONTINUATION-T00-T23 : wrapper hashé, manifeste non auto-référentiel, sidecar, README de reprise et rapport honnête assemblés.
- [x] CONTINUATION-T00-T23 : sidecar, 75 entrées de manifeste et réhydratation de la lignée T23 depuis le bundle vérifiés ; clone neuf qualifié.

- [ ] T23-CLOSURE-RECEPTION : transmettre ensemble le sidecar du kit T23, le paquet ratification/distribution et son sidecar, la ratification et la décision de clôture réelle.
- [ ] T23-CLOSURE-RECEPTION : vérifier avant transmission les cinq artefacts reçus contre leurs hashes, sidecars et contenu, puis attendre le contrôle de réception.

- [x] T23-CLOSURE-FINAL-CHECK : sidecars ZIP, ratification postérieure non antidatée, paquets canoniques hashés, registre et invariants vérifiés ; clôture locale `T23_CLOSED_RATIFIED_VERIFIABLE_LOCAL` enregistrée.

- [x] T24-GATE : précondition de clôture T23 satisfaite ; T24 reste non démarré et exigera une instruction distincte avec BASELINE_DISCOVERY.

- [x] T23-RATIFICATION-CANONICAL-PRESERVATION : paquet canonique hashé créé, sidecar et manifeste vérifiés, puis inscrit au registre.

- [x] T23-DISTRIBUTION-AND-RATIFICATION : sidecar portable transmis séparément et vérifié dans le répertoire de livraison par `sha256sum -c`; copie canonique synchronisée.
- [x] T23-DISTRIBUTION-AND-RATIFICATION : ratification propriétaire non antidatée reçue le 2026-08-19 UTC du périmètre métier T23 Archive/Restore tel qu’implémenté, distincte de l’autorisation du sous-lot de preuve.

- [x] T23-SCOPE-AND-EVIDENCE-CORRECTION : autorisation explicite du propriétaire reçue le 2026-08-19 UTC pour régulariser uniquement T23 Archive/Restore Core local et ses exclusions, sans antidatation.
- [x] T23-SCOPE-AND-EVIDENCE-CORRECTION : preuves directes de redémarrage post-reopen, absence de double History, interleaving archive/reopen/mutation et guards NewRouter ajoutées.
- [x] T23-SCOPE-AND-EVIDENCE-CORRECTION : commit tests/documentation, sidecar ZIP, logs autonomes, bundle et clone neuf produits ; soumis à revue, sans auto-approbation T23.

- [x] T23-ARCHIVE-RESTORE : BASELINE_DISCOVERY depuis la copie canonique T22, avec bundle, sidecar, clone neuf et sorties brutes.
- [x] T23-ARCHIVE-RESTORE : contrat du cycle `active → archived → active`, exclusions et protocole durable rédigés avant le code.
- [x] T23-ARCHIVE-RESTORE : transitions durables, audit redacted, récupération et sérialisation par profil implémentés.
- [x] T23-ARCHIVE-RESTORE : pannes, redémarrage, guards et concurrence couverts ; requalification et bundle réalisés, T24 non démarré.

- [ ] T23-CANDIDATE : relire le CDC et les écarts prouvés, puis proposer un lot local fermé sans démarrer son implémentation.

- [x] CONSERVATION-PERMANENTE : copie canonique hashée créée sous `/home/ubuntu/forgelocal-t22-canonical/`, sidecars portables et ZIP vérifiés.
- [x] CONSERVATION-PERMANENTE : exigence BASELINE_DISCOVERY inscrite au registre et dans `BASELINE_DISCOVERY_POLICY.md` avec commandes, chemins, UTC, codes de sortie et sorties brutes obligatoires.

- [x] T22-PENDING-DISTRIBUTION-CHECK : sidecar externe corrigé en nom relatif seul et vérifié contre le ZIP exact (`sha256sum -c` : PASS), sans modification de code.

- [x] T22-PENDING-OPERATION-CORRECTION : reprendre la baseline finale `e4a26fe…`, consigner le défaut booléen/opération et corriger les espaces finaux de l’addendum.
- [x] T22-PENDING-OPERATION-CORRECTION : remplacer le booléen seul par un marqueur owner-only (`operation_id`, action, date, digest redacted) et rendre son effacement conditionnel.
- [x] T22-PENDING-OPERATION-CORRECTION : sérialiser la séquence Profile → History SQLite → effacement conditionnel par profil.
- [x] T22-PENDING-OPERATION-CORRECTION : ajouter les tests déterministes d’entrelacement A/B, restauration concurrente, échec SQLite et échec d’effacement.
- [x] T22-PENDING-OPERATION-CORRECTION : requalifier, classifier les scans et régénérer bundle, kit, registre et sidecars sans démarrer T23 — statut `T22_PENDING_OPERATION_CORRECTION_APPROVED_VERIFIABLE_LOCAL`.

- [x] T22-CONSISTENCY FIX : qualifier le commit T22 et établir le journal de baseline de correction avant modification — `BASELINE_DISCOVERY.log`, tag T22 annoté résolu, `fsck` PASS.
- [x] T22-CONSISTENCY FIX : écrire l’addendum de récupération durable entre Profile JSON et SQLite History avant le code — `T22_CONSISTENCY_AND_RECOVERY_ADDENDUM.md`.
- [x] T22-CONSISTENCY FIX : implémenter un journal de restauration/pending operation et une reprise déterministe, avec compensation testée — marqueur `HistoryPending`, récupération router startup et test d’échec injecté.
- [x] T22-CONSISTENCY FIX : compléter les tests proxy sensible, diff/pagination, lectures sans écriture/audit, guards et concurrence restore/mutation — matrice `internal/api/profile_history_test.go` et stores.
- [x] T22-CONSISTENCY FIX : requalifier, classifier Gosec base→head, committer, bundler et livrer la copie canonique corrigée — tag final `e4a26fe…`, clone neuf et scans PASS sans nouveau finding.

- [x] T22-CONSISTENCY-AND-TESTS : vérifier et documenter le défaut de cohérence entre `profile.json` et SQLite History ; correction locale matérialisée.
- [x] T22-CONSISTENCY-AND-TESTS : fixer un contrat de récupération durable et d’audit redacted avant tout correctif.
- [x] T22-CONSISTENCY-AND-TESTS : ajouter reprise déterministe avec échec injecté, redaction proxy, pagination, lectures, guards et concurrence restore/mutation.
- [x] T22-CONSISTENCY-AND-TESTS : requalifier et préserver le correctif ; clôture `T22_CONSISTENCY_AND_TESTS_APPROVED_VERIFIABLE_LOCAL`, aucun T23 sans instruction explicite.

- [x] T22-Profile-History : produire `BASELINE_DISCOVERY` avec recherches, chemins, commandes, UTC, codes de sortie et sorties brutes avant toute écriture de code — ZIP R4, bundle, clone au tag et fsck qualifiés avant implémentation.
- [x] T22-Profile-History : définir le contrat local fermé et la matrice de tests avant implémentation — `docs/T22_PROFILE_HISTORY_CONTRACT.md`.
- [x] T22-Profile-History : implémenter versions immuables, lecture/diff/restauration transactionnelle et audit redacted avec tests — commit `5cfe7df…`.
- [x] T22-Profile-History : qualifier, scanner, committer, taguer, bundler et vérifier depuis un clone neuf — tag T22, bundle, clone/fsck, race/vet/build PASS.
- [x] T22-Profile-History : livrer une copie canonique hashée, sidecar, manifeste, registre et re-scan d’extraction — ZIP R3 `844168e6…`, registre `T22_CANONICAL_REGISTRY.md`.

- [ ] Verdict senior T0–T21 : distinguer ce qui est prêt pour poursuivre localement, ce qui est le prochain lot produit et ce qui relève seulement de la production.

- [x] Continuité privée R2 : qualifier les archives individuelles T09–T20 et les sidecars accessibles, y compris les ZIP T20/T21 joints — ZIP, manifeste interne et hash vérifiés pour T09–T21.
- [x] Continuité privée R2 : corriger les chemins `artifacts/current`, intégrer les sidecars et publier un registre de lacunes restant exact — chemin `current_t20_t21`, sidecars de transport régénérés et registre des gaps inclus.
- [x] Continuité privée R2 : assembler, vérifier et livrer la révision complète sans modification de code, commit ni tag — ZIP `1d0fbb1c…`, manifeste 58/58, sidecar et re-scan `[]` PASS.

- [x] Continuité privée T0–T21 : vérifier les sidecars externes, les chemins documentaires et les dépendances internes signalés par l’audit — corrections R2 terminées.
- [x] Continuité privée T0–T21 : qualifier l’absence ou la disponibilité des archives historiques individuelles, y compris T20-NCF — T09–T20 localisés et inclus ; réserves T01–T05, T08, T17/T18 historiques documentées.
- [x] Continuité privée T0–T21 : préparer une correction documentaire limitée si les constats sont confirmés — R2 livrée sans changement de code.

- [x] Continuité privée T07 : rechercher les documents de provenance privés accessibles et les identifier sans les réécrire — sept artefacts localisés et conservés séparément.
- [x] Continuité privée T07 : qualifier les fichiers trouvés, leurs hashes et leur emplacement avant toute inclusion — intégrité ZIP PASS et Gitleaks privé `[]`.
- [x] Continuité privée T07 : construire et vérifier une révision du kit qui les inclut comme artefacts privés séparés — ZIP privé `b2a7e71c…`, manifeste 32/32, sidecar et re-scan `[]` PASS.

- [ ] Audit de reprise T0–T21 : vérifier que code, bundle, docs, tests, dépendances et limites permettent une continuation externe sans hypothèse cachée.
- [ ] Audit de reprise T0–T21 : lister les informations, accès ou environnements que le repreneur devra encore obtenir hors du dossier.

- [ ] Continuité T0–T21 : comparer le kit précédent et le kit actuel, mesurer les ajouts/doublons et expliquer l’écart de taille.

- [x] Continuité T0–T21 : inventorier tous les ZIP, bundles, sidecars, rapports, sources et documents de reprise accessibles — inventaire brut et hashes inclus dans le kit.
- [x] Continuité T0–T21 : qualifier les bases Git récupérables, les statuts de jalons et les artefacts irrécupérables sans les remplacer — bundles T19 et T21, clones neufs et `fsck` PASS ; réserves historiques documentées.
- [x] Continuité T0–T21 : produire un handover, une matrice de couverture T0–T21, un changelog et une procédure de reprise externe — documents `README_START_HERE`, lignée, matrice, reproduction et conservation inclus.
- [x] Continuité T0–T21 : assembler un ZIP non auto-référentiel, sidecar, manifeste, vérification d’extraction et re-scan — ZIP `43d1de79…`, manifeste 34/34, sidecar et re-scan `[]` PASS.

- [x] T21-R4 remplacement : exécuter un re-scan post-extraction nouvellement daté du ZIP exact, avec log et JSON explicitement marqués comme preuve de remplacement — ZIP `151220fd…`, JSON `[]`, `exit_code=0`.
- [x] T21-R4 remplacement : livrer le log et le JSON sans les confondre avec le log original perdu — suffixe `rerun-20260819`, `evidence_type=REPLACEMENT_RESCAN_NOT_ORIGINAL_LOG`.

- [ ] T21-R4 log externe : rechercher le fichier original dans `/home/ubuntu/upload`, les téléchargements, les répertoires de livraison et les artefacts extraits accessibles, puis le transmettre inchangé s’il est trouvé.

- [x] T21-R4 restauration : qualifier le ZIP fourni, vérifier le manifeste et rechercher le log post-extraction sans recréer d’artefact — ZIP exact, manifeste 42/42 PASS.
- [x] T21-R4 restauration : transmettre le log existant si présent, ou consigner son absence exacte dans l’archive fournie — absent du ZIP et des archives accessibles ; preuve de remplacement autorisée ensuite.

- [ ] T21-EVIDENCE-R4 : vérifier et transmettre uniquement le log brut Gitleaks post-extraction R4, sans recréer ZIP, bundle, commit, tag ni code.

- [ ] Kit GitHub : vérifier le commit publié, l’archive, son sidecar et la cohérence entre hash annoncé et fichier téléchargeable.
- [ ] Kit GitHub : vérifier que `REPRODUCE.md` utilise le binaire `forge-core`, l’option `--base-dir` et une procédure token T19-R non secrète.
- [ ] Kit GitHub : vérifier les bundles, `LINEAGE.md` et les limitations afin de confirmer une reprise par un développeur externe.
- [ ] Kit GitHub : rendre la décision sur la version publiée.

- [ ] Kit de continuité : corriger le hash annoncé, joindre le sidecar ZIP externe et régénérer le manifeste si le ZIP est recomposé.
- [ ] Kit de continuité : corriger `REPRODUCE.md` (`-o forge-core`, `--base-dir`) et documenter la création non secrète du token éphémère T19-R.
- [x] Kit de continuité : vérifier l’intégrité ZIP, le manifeste et l’absence d’auto-référence ; la cohérence hash externe reste à corriger.
- [x] Kit de continuité : vérifier les bundles Core et dashboard, leurs sidecars et les clones neufs.
- [x] Kit de continuité : vérifier `LINEAGE.md`, limitations, dépendances, migrations et références de preuve.
- [x] Kit de continuité : identifier les prérequis non secrets pour rejouer T19-R et les commandes de reprise défectueuses.
- [x] Kit de continuité : rendre un verdict d’autonomie avec les compléments minimaux éventuels.

- [ ] T19-R : fournir un bundle/ZIP contenant `63e167…` ou normaliser le rapport, tag et sidecars sur le commit bundle `7d71ba6…`.
- [ ] T19-R : ajouter une procédure non secrète de replay des E2E, avec variables, ordre de démarrage et séparation `--no-runtime` / Chromium qualifié.
- [x] T19-R : vérifier les hashes ZIP/bundle contre leurs sidecars, l’intégrité, le manifeste et l’absence d’auto-référence.
- [x] T19-R : vérifier le bundle dans un dépôt neuf et identifier la divergence entre commit du rapport, tag et commit réellement cloné.
- [x] T19-R : examiner les E2E, l’inspection de stockage, le Core réellement lancé et la dérogation documentée au mode `--no-runtime`.
- [x] T19-R : rejouer les contrôles réalisables depuis le clone et auditer Gitleaks/Gosec de la plage exacte.
- [x] T19-R : rendre une décision sans confondre un parcours démontré au commit bundle avec le commit final annoncé.

- [ ] T18-R complet : joindre le sidecar externe SHA-256 du ZIP à toute transmission ultérieure ; le hash réel est toutefois consigné.
- [x] T18-R complet : vérifier bundle, tag, commit, clone neuf et `git fsck --full` dans un dépôt indépendant.
- [x] T18-R complet : rejouer les tests queue/recovery ciblés et la qualification Core sur le clone bundle.
- [x] T18-R complet : vérifier Gitleaks, Gosec de plage immuable et absence de dérive hors de l’audit documentaire.
- [x] T18-R complet : rendre la décision de clôture et consigner les limitations qui restent externes.

- [ ] T18-R : obtenir l’archive de preuves, son sidecar SHA-256, le bundle et le manifeste avant toute validation indépendante.
- [ ] T18-R : vérifier dans un clone neuf les tests queue/recovery, la qualification Core et la chaîne bundle-clone.
- [ ] T18-R : auditer les logs Gitleaks/Gosec de la plage post-T17-R et confirmer que les 17 comportements ne sont pas seulement déclarés.
- [ ] T18-R : rendre la décision définitive ou documenter précisément les preuves manquantes.

- [ ] T17-R2 : joindre le sidecar externe SHA-256 du ZIP lors de toute prochaine transmission ; le hash réel est toutefois consigné.
- [x] T17-R2 : vérifier le hash, l’intégrité ZIP, le manifeste et la chaîne bundle-clone de la correction.
- [x] T17-R2 : confirmer que le test sessions instancie le handler ou routeur produit avec une fixture contenant des champs techniques non vides.
- [x] T17-R2 : rejouer le test ciblé et la qualification Core, puis vérifier scans et plage Git immuable.
- [x] T17-R2 : rendre la décision finale sur la clôture T17-R, sans modifier la distinction avec le T17 historique.

- [ ] T17-R : remplacer le test sessions synthétique par un test qui exerce `(*handler).listSessions` ou le routeur produit avec une session technique fixture.
- [x] T17-R : vérifier le hash, l’intégrité, le manifeste et la chaîne bundle-clone de l’archive reçue ; sidecar externe ZIP absent de la transmission.
- [x] T17-R : confirmer par le code et les tests la redaction G15-A, les refus Origin G15-B et les verrous G15-C ; l’ajout sessions demeure synthétique et doit être corrigé.
- [x] T17-R : rejouer les tests Core pertinents, vérifier le delta, Gitleaks et le filtre Gosec sur la plage immuable post-G6.
- [x] T17-R : confirmer explicitement que le verdict ne prétend pas restaurer le snapshot historique T17.
- [x] T17-R : rendre une décision locale vérifiable et lister les gates externes inchangés.

- [ ] G6 : joindre le sidecar externe du ZIP ; le hash réel est vérifié mais le sidecar de livraison n’était pas joint.
- [x] G6 : vérifier le bundle, le tag, le clone neuf et l’absence de dérive hors des fichiers LocalVault/Runtime annoncés.
- [x] G6 : lire le diff et les tests de confinement afin de confirmer que les sept findings sont corrigés par du code, non masqués.
- [x] G6 : vérifier les sorties Core, Gitleaks et Gosec sur la plage immuable `419f324…3b78c3c…`.
- [x] G6 : rendre une décision de sous-lot sans modifier les gates de release.

- [ ] R5 : transmettre le sidecar externe du ZIP R5 ; le hash réel est enregistré, mais aucun sidecar de livraison n’était joint.
- [x] R5 : recalculer le hash du bundle joint, exécuter `git bundle verify` dans un dépôt et vérifier un clone créé depuis ce bundle.
- [x] R5 : vérifier les logs de clone frais avec commandes, timestamps, commits et exit codes pour Core et dashboard.
- [x] R5 : vérifier que Gitleaks et Gosec utilisent la plage immuable baseline→commit et confirmer les sept findings Gosec annoncés.
- [x] R5 : rendre un verdict séparant les corrections R4 validées des findings Gosec toujours bloquants.

- [x] Calculer le SHA-256 réel de l’archive reçue et vérifier l’intégrité ZIP.
- [x] Extraire l’archive dans un répertoire neuf, identifier le manifeste et vérifier chaque entrée sans auto-référence.
- [x] Comparer les hashes annoncés, sidecars, bundles, commits et tags aux preuves réellement incluses.
- [x] Examiner les logs bruts Core, dashboard, Playwright, Gitleaks et Gosec pour confirmer les résultats annoncés.
- [x] Vérifier la chaîne de conservation : bundle, clone neuf, état Git, remote et contenu des sources associées.
- [x] Rendre un verdict nuancé : valide, valide avec réserves, ou nécessite correction.
