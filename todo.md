# Addendum v2.1 — Corrections avant figement

- [ ] Vérifier les documents canoniques, manifestes et statut Git de la branche produit.
- [x] Créer l’addendum v2.1 au chemin documentaire canonique avec identifiant et date.
- [x] Ajouter le registre Camoflox traçable, la politique de ports par runtime et les contrôles API redacted.
- [x] Clarifier le Bearer token local, les en-têtes de requête et le parallélisme CAMO-CORE-01 / lecture seule.
- [x] Calculer le SHA-256, produire le manifeste, vérifier le delta et créer les commits documentaires isolés.

# Fondations CAMO-CORE-01 et lecture seule redacted

- [x] Auditer les sources Camoflox et les surfaces API Core existantes sans les intégrer.
- [x] Publier le registre CAMO-CORE-01 avec hash, commit, dépendances, décision et responsable par module.
- [x] Ajouter les contrats API Core de lecture seule, redacted et paginés, avec tests.
- [x] Préparer le client React mémoire seule et ses états lecture seule sans token persistant.
- [x] Vérifier les contrats, scans, builds et commits isolés sans modifier le RC ou les gates publics.

# Addendum v2.3 — Provenance des composants

- [x] Auditer les scripts CI et inventaires de dépendances avant d’ajouter le contrôle de provenance.
- [x] Rédiger l’addendum v2.3 avec la séparation CSP/dashboard et rate limiting/Core.
- [x] Créer le registre JSON canonique et sa vue Markdown, avec GoLogin déclaré `denied` / `écarter`.
- [x] Implémenter la vérification CI des statuts `authorized` et `not_required` first-party uniquement.
- [x] Tester les cas autorisé, refusé, inconnu, absent et first-party, puis attester et versionner le lot.

# Exécution CI release — provenance v2.3

- [x] Auditer le workflow de release et les capacités d’archivage déjà disponibles.
- [x] Exécuter obligatoirement les contrôles de provenance dans le pipeline de release.
- [x] Produire et archiver le registre JSON validé avec son SHA-256 comme artefact CI.
- [x] Vérifier la syntaxe du workflow, les contrôles locaux et le delta hors RC avant versionnage.

# Non-contournement CI — merge et release

- [x] Auditer les workflows de merge/release et les capacités GitHub disponibles dans cette session. Résultat : remote `nczz/BrowseForge`, identité GitHub en lecture seule ; API protection `main` inaccessible (403).
- [x] Ajouter les garde-fous versionnables qui empêchent un workflow release de sauter la provenance.
- [x] Vérifier la protection de branche et les required checks externes ; configuration impossible sans droits administrateur sur le dépôt ForgeLocal maintenu.
- [x] Tester les workflows, scans et la configuration sans modifier le RC ni les gates publics. Contrôles locaux, test API Go 1.25.13, format Git et scan Gitleaks du delta réussis.

# Revue des fichiers sensibles et approbation de release

- [x] Confirmer les responsables GitHub réels : sécurité `@boucheriechefimane-cmd`, release `@davidwilsonbest89-afk`.
- [x] Ajouter le fichier `.github/CODEOWNERS` couvrant les workflows, validateurs, registre, dépendances et répertoires de release sensibles.
- [x] Rattacher l’environnement GitHub protégé `production-release` au seul job de publication après `provenance → verify → build`.
- [x] Documenter les réglages ruleset obligatoires : revue Code Owners, auteur distinct du dernier push, approbations obsolètes, absence de bypass et tags `v*` restreints.
- [x] Vérifier le delta, scanner les changements et les versionner sans modifier le RC ni les gates publics. Contrôles locaux, test API Go 1.25.13, format Git et scan Gitleaks du delta réussis.

# Revue renforcée des changements de release sensibles

- [x] Confirmer la relectrice indépendante réelle : `@hajarbenmlih91-cloud`.
- [x] Étendre la politique CODEOWNERS des chemins critiques à la relectrice indépendante, sans remplacer les rôles Sécurité et Release.
- [x] Documenter la règle distante requise : deux approbations, Code Owners, dernier push approuvé par une autre personne et zéro bypass.
- [x] Vérifier la cohérence, scanner le delta et versionner le lot sans modifier le RC ni les gates publics. Contrôles locaux, test API Go 1.25.13, format Git et scan Gitleaks du delta réussis.

# Contrôle CI de séparation Sécurité / Release

- [x] Définir les quatre scénarios d’approbation et les entrées JSON strictement nécessaires au contrôle.
- [x] Ajouter un validateur déterministe et ses fixtures, sans appel réseau ni exécution de code de PR.
- [x] Ajouter un workflow dédié sur les métadonnées de pull request, avec permissions lecture seule et sans checkout de la branche PR.
- [x] Documenter le check requis, le futur usage GitHub App à permission minimale et la preuve redacted.
- [x] Tester, scanner le delta et versionner sans modifier le RC ni les gates publics. Contrôles locaux, test API Go 1.25.13, format Git et scan Gitleaks du delta réussis.

# Continuité de l’environnement `production-release`

- [x] Formaliser les deux approbatrices requises : Release `@davidwilsonbest89-afk` et indépendante `@hajarbenmlih91-cloud`.
- [x] Mettre à jour la procédure mainteneur et le contrôle versionné de l’environnement déclaré dans le workflow.
- [x] Vérifier, scanner et versionner sans modifier le RC ni les gates publics ; conserver la configuration GitHub effective comme preuve distante PENDING. Contrôles locaux, test API Go 1.25.13 et format Git réussis avant scan final.

# Dépôt ForgeLocal administré par le propriétaire

- [x] Vérifier que l’identité GitHub active est `@boucheriechefimane-cmd` et que le nom privé `IPcache` est disponible.
- [x] Créer le dépôt privé `IPcache`, préserver `upstream` BrowseForge et rattacher `origin` au dépôt administré.
- [x] Vérifier les remotes et l’absence de publication de release ; documenter les réglages d’administration GitHub restants. La branche produit est publiée, sans tag ni release.

# Progression produit locale pendant les validations externes

- [x] Auditer les contrats Core et le dashboard pour identifier le premier incrément sûr et livrable.
- [x] Implémenter le bootstrap local de session mémoire seule, loopback-only et ses tests de sécurité.
- [x] Raccorder les vues lecture seule du dashboard sans rendre disponible une mutation, un lancement ou un runtime candidat.
- [x] Définir les contrats préparatoires Groupes, Runtimes, Proxys, Backups et Audit avec états redacted et sans secrets.
- [x] Poursuivre uniquement le triage documentaire de `generic-api-key` hors RC, sans relancer SystemVault ni reconstruire une release. Les documents existants confirment le statut `SCAN_BLOCKED_UNKNOWN` et ne permettent aucune action sur le candidat.
- [x] Exécuter les contrôles, scanner le delta et publier les commits produit isolés vers IPcache. Tests Core/CLI, contrôle CI, vérification dashboard, format Git et scans Gitleaks des deux deltas réussis.

# BOOTSTRAP-RO-01 — validation locale lecture seule

- [x] Préparer une matrice d’acceptation assainie et un harness local sans impression de code ni de token.
- [x] Prouver : code usage unique, échange loopback autorisé, refus de seconde utilisation et refus après expiration.
- [x] Prouver les refus d’origine, host et port non autorisés.
- [x] Prouver : aucune persistance navigateur, aucun Bearer dans URL, logs, analytics ou exports, invalidation immédiate après `401`/expiration.
- [x] Prouver les lectures redacted `health`, `dashboard/summary` et `profiles`, sans aucune écriture ni opération runtime.
- [x] Exécuter les tests ciblés et `go test -race`, scanner les preuves, versionner la décision `BOOTSTRAP_RO_APPROVED`.

# BOOTSTRAP-RO-01 — preuves d’exécution strictes FL-STRICT-EXEC-1.0

- [x] Reclasser la décision publiée en `DECLARED_PASS_EVIDENCE_PENDING` jusqu’à revue du dossier d’exécution rejouable.
- [x] Créer les commandes reproductibles `core-dev`, `dashboard-dev` et `test-bootstrap-ro`, sans activation de runtime ni inclusion de secret.
- [x] Exécuter séquentiellement T00 à T04 avec listes de tests, sorties JSON, race detector et fichiers de preuve assainis.
- [x] Exécuter T05 avec deux serveurs locaux loopback, inspection des sockets et E2E navigateur automatisé couvrant les dix contrôles.
- [x] Hacher toutes les preuves, scanner le dossier, vérifier le diff des chemins RC gelés et versionner les contrôles sans committer les données d’exécution.
- [x] Rendre une décision finale uniquement à partir des sorties T00 à T05 et de leurs hashes : `BOOTSTRAP_RO_APPROVED_VERIFIABLE`.

# BOOTSTRAP-RO-01 — clôture de traçabilité

- [x] Générer un manifeste `SHA256SUMS` portable, strictement relatif, vérifiable directement après extraction de l’archive.
- [x] Intégrer le refus hors loopback au résultat global T05 avec un statut `PASS` et une référence explicite à ses deux preuves autoritatives.
- [x] Archiver et hacher les sorties T00 : `git status --short`, `git diff --check` et diff des chemins RC gelés.
- [x] Exécuter le run corrigé, vérifier `19/19` hashes après extraction et scanner le delta ainsi que l’archive extraite.
- [x] Versionner exclusivement les scripts et documents de preuve, sans démarrer T06, modifier le RC, créer un tag ou une release.

# T06 — Groupes et Runtimes lecture seule

- [x] P0 audit indépendant : identifier les hashes Git source de base et cible du dashboard React T06, puis inventorier un delta dashboard non vide.
- [x] P0 audit indépendant : scanner avec Gitleaks 8.18.4 le snapshot des fichiers dashboard modifiés, avec journal redacted et JSON `[]` en l’absence de détection.
- [x] P0 audit indépendant : assembler une archive globale unique couvrant les preuves Core et dashboard, avec manifeste relatif et rescan après extraction ; ne pas démarrer T07 avant validation.
- [x] P0 audit indépendant : rejouer Gitleaks 8.18.4 sur la plage exacte `ae4622f1d49cfbe4a1872d0c833d21bc7aa25afb..37c34a65889ddc24aade417db75ffe9e600f5bee` avec une preuve de couverture non nulle, redacted et sans modifier le code produit.
- [x] P0 audit indépendant : constituer une mini-archive du seul correctif avec commande exacte, baseline, cible, inventaire de fichiers/commits, JSON `[]` sans détection, manifeste relatif et vérification après extraction.
- [x] P0 audit indépendant : soumettre la mini-archive à vérification avant toute décision `T06_APPROVED_VERIFIABLE` ; ne pas démarrer T07.
- [ ] Préflight : vérifier le commit propre, les chemins RC inchangés, les versions verrouillées et le dossier de preuves T06.
- [ ] T06-1 : compléter et passer les contrats API v1 Groups/Runtimes redacted, paginés, limités, authentifiés et sans mutation, y compris `-race`.
- [ ] T06-2 : créer les fixtures SQLite synthétiques via le chemin Core, vérifier l’intégrité et l’absence de sentinelles dans les réponses.
- [ ] T06-3 : raccorder les vues React Groupes/Runtimes au client mémoire seule et prouver le parcours local Playwright (chargement, vide, erreur, timeout, `401`, réessai).
- [ ] T06-4 : prouver redaction, absence des sentinelles dans DOM/réponses/logs/preuves et absence ou refus des mutations.
- [ ] T06-5 : produire l’archive hashée et portable, le rapport Gitleaks redacted du delta et de l’archive extraite, puis vérifier zéro delta RC avant revue indépendante.

# T07 — Provenance des modules Camoflox candidats

- [x] T07-1 : inventorier les modules Camoflox candidats et leurs sources sans importer, exécuter ni porter de code.
- [x] T07-2 : vérifier pour chaque candidat les droits, licence, propriétaire, révision exacte, hash, dépendances et décision de périmètre ; résultat `PARTIAL/BLOCKED` consigné sans exception.
- [x] T07-3 : compléter le registre machine-readable et produire une SBOM de provenance sans secret ni donnée de runtime.
- [x] T07-4 : exécuter et documenter les contrôles PROV-01 à PROV-07, y compris les refus `unknown` et `denied`.
- [x] T07-5 : produire une archive de preuves T07 redacted et portable pour audit indépendant ; ne pas commencer T08.

# T07-R — Remédiation de preuve et de sécurité (T08 interdit)

- [x] T07-RV1 : préparer un paquet de revue redacted avec attestation à compléter, fiche de synthèse, checklist de confirmation indépendante et consignes de retour sans secret.
- [x] T07-RV2 : vérifier que le paquet ne contient ni code, ni archive, ni clé, ni token, ni valeur d’alerte, ni décision automatique `PASS`.
- [x] T07-RV3 : transmettre le paquet prêt à compléter au créateur/détenteur des droits et à la relectrice indépendante ; attendre leurs références redacted. Issue privée GitHub `#1` créée le 2026-08-15.
- [x] T07-RV4 : inviter `@hajarbenmlih91-cloud` comme collaboratrice du dépôt privé `IPcache` et créer l’Issue privée redacted de demande de revue, sur autorisation explicite du propriétaire. Invitation GitHub en attente d’acceptation avec accès `read`.
- [x] T07-RV5 : diagnostiquer la visibilité de l’Issue privée `#1` pour `@hajarbenmlih91-cloud` après son acceptation de l’invitation, puis corriger le seul accès requis. Accès `read` actif vérifié ; Issue ouverte et assignée à la relectrice.
- [x] T07-RV6 : examiner la réponse publiée par `@hajarbenmlih91-cloud` dans l’Issue privée `#1` et statuer uniquement sur sa complétude redacted. Six couvertures `true`, référence fournie et décision indépendante `FALSE_POSITIVE` valides ; attestation du détenteur et remédiations exigées pour `FALSE_POSITIVE` encore absentes.
- [ ] T07-RV7 : vérifier l’attestation PDF redacted fournie par le détenteur, puis la joindre à l’Issue privée `#1` avec une demande de confirmation sur autorisation explicite. Contrôle effectué : PDF sans contenu sensible visible mais présenté comme brouillon et comportant des placeholders bloquants ; attente du choix du détenteur entre complétion et revue limitée de brouillon.
- [x] T07-RV8 : contrôler la version remplie `T07-R_—_Attestation_redacted_du_détenteur_des_droits(1).pdf` et la comparer à la revue indépendante `REV-T07R-2026-001`. Droits `yes/yes/granted`, lien de snapshot et décision `FALSE_POSITIVE` cohérents ; références d’obligations tierces/notices et informations de signature encore absentes, donc pas d’approbation finale possible.
- [ ] T07-RV9 : publier dans l’Issue privée `#1` le statut `ATTESTATION_REDACTED_PENDING_SIGNATURE_AND_EXTERNAL_REFERENCES` et demander uniquement les références obligations/notices, les éléments de signature et la confirmation limitée du nouveau snapshot/re-scan ; aucun changement de statut T07 ni de registre. Commentaire publié : `issuecomment-5303080327`, avec empreinte PDF `d36b5a11761ff82e20b52cd96cb49651fdcb5bc1b9e7c96e9321f4b956ef70ef`. Le PDF reste hors Git et doit être transmis par un canal privé contrôlé ou ajouté manuellement à l’Issue par le propriétaire.
- [x] T07-RV10 : notifier `@hajarbenmlih91-cloud` dans l’Issue privée `#1` afin qu’elle réponde seulement au commentaire `issuecomment-5303080327` sur les trois confirmations limitées. Mention publiée et vérifiée : `issuecomment-5303101953`.
- [x] T07-RV11 : examiner la nouvelle réponse de `@hajarbenmlih91-cloud` à la demande limitée de l’Issue `#1` et vérifier les trois confirmations attendues. Commentaire `issuecomment-5303108320` : obligations/notices, signature/validation/approbation et cohérence snapshot/SHA/re-scan déclarées `true`, statut limité `LIMITED_CONFIRMED`. Les références factuelles restent toutefois à renseigner dans l’attestation avant tout contrôle final automatisé.
- [x] T07-RV12 : contrôler la version `T07-R_—_Attestation_redacted_du_détenteur_des_droits(2).pdf` contre les références et confirmations limitées T07-R. Références droits, obligations tierces, notices, signature, snapshot redacted et re-scan présentes ; le document indique toujours `PENDING_EXPLICIT_HOLDER_CONFIRMATION`, donc la complétude documentaire finale reste en attente de confirmation explicite du détenteur.
- [x] T07-RV13 : vérifier le commentaire de confirmation explicite publié par `@boucheriechefimane-cmd` dans l’Issue privée `#1` et comparer sa portée à l’attestation `ATT-T07R-CAMOFLOX-001`. Commentaire `issuecomment-5303135777` conforme : détenteur, droits, références redacted, triage `FALSE_POSITIVE`, snapshot/re-scan et limites d’autorisation confirmés.
- [x] T07-RV14 : examiner la nouvelle réponse de `@hajarbenmlih91-cloud` dans l’Issue privée `#1`. Commentaire `issuecomment-5303140628` : confirmation finale limitée `REV-T07R-2026-002` du snapshot redacted, de son SHA-256, du re-scan Gitleaks et du maintien de `FALSE_POSITIVE`, sans autorisation d’intégration, de T08 ni de release.
- [x] T07-RV15 : créer dans l’emplacement privé l’attestation JSON redacted à partir du modèle et des confirmations T07-R, avec statut `external_evidence_received_pending_independent_audit` et T08 interdit.
- [x] T07-RV16 : exécuter le validateur T07-R sans le modifier, puis consigner le résultat de complétude et de cohérence. Résultat : `complete_for_independent_review_only`, `t08_authorized=false`, triage `FALSE_POSITIVE`, distribution future soumise à revue indépendante.
- [x] T07-RV17 : collecter une archive redacted portable, vérifier son manifeste après extraction et la soumettre uniquement à l’audit indépendant. Archive `t07-r-evidence-20260815.zip` SHA-256 `98af87596ba344292151f6e5109fb9a392803d14df58f2d51ef7eca037e37241` ; manifeste 14/14 vérifié, trois scans Gitleaks à `0` finding, aucune copie de l’attestation privée.
- [x] T07-RV18 : publier dans l’Issue privée `#1` la demande de revue finale limitée du paquet `t07-r-evidence-20260815.zip`, notifier `@hajarbenmlih91-cloud` et attendre l’une des décisions `REVIEW_ACCEPTED_FOR_T07_DECISION`, `REVIEW_INCOMPLETE` ou `REVIEW_REJECTED`. Commentaire publié et vérifié : `issuecomment-5303217806`.
- [x] T07-RV19 : récupérer et examiner la décision finale de `@hajarbenmlih91-cloud` dans l’Issue privée `#1`. Commentaire `issuecomment-5303254116` : `REVIEW_ACCEPTED_FOR_T07_DECISION`, archive/manifeste/validateur/T08/scans/références privées confirmés ; autorise uniquement l’audit final de décision T07, sans T08 ni release.
- [x] T07-FD1 : assembler un manifeste de décision finale T07 redacted avec les références attestation, revues, archives, snapshots, triage, droits et contraintes de périmètre.
- [x] T07-FD2 : vérifier les hashes, la cohérence des références, les gates de provenance, le validateur T07-R et l’intégrité de l’archive source de preuves. Manifeste T07-R `14/14 OK`, validateur `complete_for_independent_review_only`, composant Camoflox toujours `provenance-qualification-blocked`.
- [x] T07-FD3 : scanner le delta documentaire et le dossier de décision redacted, puis prouver l’absence de changement sous les chemins RC gelés. Delta documentaire : 20 fichiers, Gitleaks 8.18.4 JSON `[]` ; dossier : JSON `[]` ; RC committé et de travail : 0 fichier.
- [x] T07-FD4 : produire et vérifier l’archive portable du dossier final au statut `T07_FINAL_DECISION_EVIDENCE_READY_FOR_INDEPENDENT_AUDIT`. Archive SHA-256 `a10087c8b112949797ffd9f270693f8ddddc399234e445167c3126ad040041c3`, manifeste 19/19, re-scan JSON `[]`, aucune attestation privée ni archive source.
- [x] T07-RP1a : exiger dans l’attestation et le validateur que la revue indépendante couvre explicitement révision/snapshot, hash d’archive, portée des droits, licence/accord, notices et triage.
- [x] T07-RH1 : séparer strictement le modèle JSON d’attestation et la checklist Markdown, puis vérifier que le JSON est parseable sans contenu annexe.
- [x] T07-RH2 : exiger `internal_use=yes` et `modification=yes` pour toute attestation complète susceptible d’être soumise à revue.
- [x] T07-RH3 : exiger deux décisions de triage identiques et appartenant à l’ensemble autorisé ; toute absence ou divergence doit produire `UNKNOWN` opérationnel.
- [x] T07-RH4 : pour `REAL_SECRET`, exiger une référence de rotation/révocation, un nouveau snapshot candidat hashé et un re-scan avant toute future revue.
- [x] T07-RH5 : pour `redistribution=not_granted`, émettre `future_distribution=blocked` tout en laissant l’audit passif possible.
- [x] T07-RP1 : créer un emplacement privé contrôlé pour les attestations, licences, notices et décisions redacted, interdit au dépôt Git et aux archives de preuves.
- [x] T07-RP2 : préparer une checklist et un validateur qui refusent les attestations incomplètes, hashes divergents, révisions absentes, redistribution implicite, licence/notices absentes ou triages divergents.
- [x] T07-RP3 : préparer le collecteur de la future archive T07-R afin d’inclure uniquement des métadonnées redacted, jamais les documents bruts privés, sources, clés, tokens ou valeurs d’alerte.
- [x] T07-RP4 : vérifier que l’état d’intégration reste `provenance-qualification-blocked`, que les chemins produit et RC restent inchangés et que T08 est interdit.
- [x] T07-R0 : appliquer les critères de collecte renforcés : lien explicite révision/snapshot → SHA-256 candidat, redistribution déclarée `granted` ou `not_granted`, double triage redacted et revue privée réellement consultable.
- [x] T07-R1 : consigner la chaîne de liaison vérifiable `révision privée attestée → hash du snapshot étudié → décision de portage/réimplémentation/écarter → futur commit ForgeLocal → tests`, sans source privée ni clé dans Git. Chaîne consignée dans le suivi et le registre canonique : `camoflox-source.zip` SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` → `lib/concurrency.js` hash `b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7` → décision `reimplementer` (registre, avant tout code) → commit ForgeLocal (futur) → tests (futur).
- [ ] T07-R2 : obtenir ou référencer sous contrôle d’accès une preuve indépendante de propriété/droits, une licence racine ou un accord attesté et les notices tierces applicables.
- [ ] T07-R3 : soumettre l’alerte redacted de `tests/smoke.test.js:24` à un mainteneur et une relectrice indépendante ; conserver `UNKNOWN` et le blocage tant qu’aucune décision `REAL_SECRET` ou `FALSE_POSITIVE` n’est prouvée.
- [ ] T07-R4 : si une décision `FALSE_POSITIVE` est produite, exiger un snapshot candidat nouveau, redacted, rescané et hashé ; si `REAL_SECRET`, exiger révocation/rotation avant toute nouvelle preuve.
- [ ] T07-R5 : reconstruire une archive T07 redacted avec registre, CI, SBOM ciblée, scans et manifeste portable, puis demander une revue indépendante ; ne pas commencer T08 avant `PROV-01`, `PROV-04` et `PROV-06` à `PASS`.

# T07-R1 à T08-F1 — décision d'audit T07 reçue (avant enregistrement)

**Source :** décision d'audit transmise par l'utilisateur (`pasted_content_17.txt`), dossier `t07-final-decision-evidence-20260815.zip` (SHA-256 `a10087c8b112949797ffd9f270693f8ddddc399234e445167c3126ad040041c3`), identifiant `T07-DECISION-20260815-001`.

**Décision :** `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION`. Provenance approuvée uniquement pour étudier un module Camoflox hashé à la fois et le réimplémenter proprement en Go, sans copier, importer ni exécuter le code candidat, sans runtime, queue, lock, port ou lancement Camoflox, et sans lever aucun gate BACK-01 ou de release.

**Statuts PROV mis à jour :** PROV-01 PASS (snapshot privé identifié et attesté) ; PROV-02 PASS (inventaire/SBOM ; toute nouvelle dépendance = revue avant ajout) ; PROV-03 PASS (candidat non intégré, aucun second control plane) ; PROV-04 PASS (snapshot redacted précis + rescan ; tout nouveau snapshot = nouveau scan) ; PROV-05 DEFERRED_TO_T08 (aucune fiabilité du code Camoflox conclue par T07) ; PROV-06 PASS (distribution future = revue de droits séparée) ; PROV-07 DEFERRED_TO_T08 (valeur produit à prouver par réimplémentation Go testée).

**Chaîne T07-R1 :** snapshot privé `camoflox-source.zip` (SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2`) → fichiers étudiés (hash enregistrés par module) → décision `reimplementer` (registre, avant tout code) → commit ForgeLocal (futur) → tests (futur). Attestation `ATT-T07R-CAMOFLOX-001` ; revues `REV-T07R-2026-001` / `REV-T07R-2026-002` ; décision finale `REVIEW_ACCEPTED_FOR_T07_DECISION`.

**T08 autorisé — périmètre strict :** un seul module Camoflox déjà hashé ; spécification ForgeLocal indépendante avant tout code ; réimplémentation Go pure sans importer le code source. Premier choix : `lib/concurrency.js` (hash `b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7`) comme inspiration de contrat de concurrence seulement. Tests requis : queue bornée, limite globale, lock par profil, timeout, annulation `context`, concurrence `-race`, cleanup, recovery après crash, journal d'audit redacted, absence de lancement/runtime. Interdit : ports, Camoufox, cycle de vie navigateur, proxies, backups, restaurations, mutations UI, runtime lançable.

**Contraintes maintenues :** import direct interdit ; backend Node/Electron interdit ; `PUBLIC_RELEASE_BLOCKED` maintenu ; `SCAN_BLOCKED_UNKNOWN` BACK-01 maintenu (distinct de T07) ; pilote BACK-01 suspendu ; cinq gates publics en attente ; écritures UI/proxy/backup/restauration interdites jusqu'à leurs jalons propres.

**Actions en cours :** T07-R1 consigné. T08-F1 enregistrer décision `reimplementer` pour `lib/concurrency.js`. TODO. T08-F2 spécification ForgeLocal indépendante du contrat de concurrence. TODO. T08-F3 réimplémentation Go pure avec tests `-race`. TODO. T08-F4 relecture sécurité puis release avant intégration. TODO.

# T08-F0 — mise à jour du registre canonique avant code T08

- [x] T08-F0a : passer le registre JSON canonique à `schema_version: 1.1` (`FL-COMP-RIGHTS-20260814-R1`) avec décision `reimplementer-selective`, état `t08-concurrency-in-progress`, bloc de provenance T07 complet (statut, identifiant de décision, archives, attestation, revues, triage concordant `FALSE_POSITIVE`, distribution `not_granted` + `future_distribution=blocked`, PROV-01/02/03/04/06 `PASS`, PROV-05/07 `DEFERRED_TO_T08`).
- [x] T08-F0b : enregistrer dans le registre la décision `reimplementer` pour `lib/concurrency.js` avant tout code : raison, rôle `concurrency-contract-inspiration-only`, exclusions explicites, chaîne T07-R1, et `lib/global-action-limiter.js` en `deferred`.
- [x] T08-F0c : ajouter le bloc `release_gates` au registre (PUBLIC_RELEASE_BLOCKED, SCAN_BLOCKED_UNKNOWN, pilote suspendu, cinq gates publics, contraintes T08).
- [x] T08-F0d : valider la syntaxe JSON du registre (parse sans erreur) et calculer son SHA-256 `3723b45f3b46a3a1d59e6185126a8576ed089bf35f5ee7863c15c34f77d467e3`.
- [x] T08-F0e : synchroniser la vue Markdown `COMPONENT_RIGHTS_REGISTER.md` avec le registre JSON et la décision T07.

- [x] T08-F1 : enregistrer dans le registre canonique et la vue Markdown la décision `reimplementer` pour `lib/concurrency.js` (hash `b055a3e1…a36085d7`, snapshot `dcf668d4…e4851c2`), avec raison, rôle d'inspiration de contrat, exclusions explicites et chaîne T07-R1 ; `lib/global-action-limiter.js` en `deferred` ; registre validé JSON, SHA-256 `3723b45f3b46a3a1d59e6185126a8576ed089bf35f5ee7863c15c34f77d467e3`.
- [x] T08-F2 : rédiger la spécification ForgeLocal indépendante `docs/T08_CONCURRENCY_SPEC.md` (`T08-SPEC-20260815-001`) : queue bornée, limite globale, verrou par profil, timeout, contexte, cleanup, reprise après plantage, journal d'audit redacted, critères d'acceptation et interdiction de tout lancement/runtime ; clean room, sans import du candidat.

- [x] T08-V1 : scanner le delta documentaire T08 avec Gitleaks 8.18.4 `--no-git` (3 fichiers : registre JSON, vue Markdown, spécification) ; résultat JSON `[]`, zéro détection, sortie 0 ; rapport et fichiers hachés (registre `3723b45f3b46a3a1d59e6185126a8576ed089bf35f5ee7863c15c34f77d467e3`, vue `9edbbb9ded4cf4ea8968986b7601b4f80e680f35436500582e7298b33190e135`, spécification `63eff82b2ac88493657b42e29018f9551b46653cb26cc51b87472e391ce567d3`, rapport `37517e5f3dc66819f61f5a7bb8ace1921282415f10551d2defa5c3eb0985b570`).
- [x] T08-V2 : vérifier la cohérence d'ensemble — registres, spécification, suivi, décision T07 : même hashes source, même exclusions, `PUBLIC_RELEASE_BLOCKED` et `SCAN_BLOCKED_UNKNOWN` déclarés maintenus dans le registre et les documents, aucun chemin RC modifié, aucun code produit changé.

# CDC v3.9.7 — réception et validation du cahier des charges actualisé

- [x] CDC-1 : cahier `FL-CDC-3.9.7-20260815` enregistré comme document directeur du lot dans `docs/CAHIER_DES_CHARGES_v3_9_7.md` (SHA-256 `58de4650bd14d71fe774ba75ccb9d2b8cb2468a85ee8f4e85d6313b784a145b8`), avec préambule d’enregistrement et annexe du texte intégral.
- [x] CDC-2 : cohérence vérifiée — le cahier reprend fidèlement la décision T07 (`§7.3`), la table des gates PROV, le périmètre T08 (`§9`, AC-CAMO-01 à AC-CAMO-05 coïncident avec `docs/T08_CONCURRENCY_SPEC.md`), les exclusions définitives et les statuts release maintenus. Aucun conflit détecté avec le registre JSON canonique.
- [x] CDC-3 : conciliation documentaire réalisée — `CAMO_CORE_01_REGISTER.md` mis à jour (état T07 clôturé, décisions de modules actualisées : `lib/concurrency.js` AUTORISÉ T08, autres modules reportés/hors périmètre) ; `T07_CAMOFLOX_PROVENANCE.md` addendum de clôture ajouté (table PROV historique → statuts après décision) ; cahier v1.0 addendum de supersession ajouté (v3.9.7 prévaut, SHA-256 `58de4650bd14d71fe774ba75ccb9d2b8cb2468a85ee8f4e85d6313b784a145b8`). RC et gates publics non touchés.
- [x] CDC-4 : scan Gitleaks 8.18.4 `--no-git` du delta documentaire CDC (6 fichiers) : JSON `[]`, zéro détection, sortie 0 ; registre JSON canonique validé ; aucun fichier sous `release/` ou `dist/` modifié depuis le début du lot.
- [x] CDC-5 : verdict de validation du cahier livré (cohérent, conforme aux registres et à la décision T07) avec le cadre d’exécution T08 : `LaunchManager`/`SessionManager` Go clean room, AC-CAMO-01 à AC-CAMO-05, exclusions T08 (ports runtime réels, Camoufox, lancement navigateur, proxy, backup, restauration, import, mutation UI, release).

# T08 implémentation complète (demande utilisateur, livrable unique)

- [ ] T08-P0 : préparer l’environnement Go 1.25.13 (GOTOOLCHAIN=local) dans le Core ForgeLocal et créer le scaffold `internal/launch/`. TODO
- [ ] T08-P1 : implémenter queue bornée, limite globale, sérialisation par profil, timeout, `context.Context`, annulation, cleanup idempotent, crash recovery et audit redacted dans `internal/launch/` (clean room, sans code candidat). TODO
- [ ] T08-P2 : tests unitaires positifs, négatifs, de concurrence et de redaction (aucun lancement navigateur/runtime/Camoufox, aucun proxy, aucune UI, aucun backup/restore/import). TODO
- [ ] T08-P3 : exécuter tous les contrôles — `go test -list`, tests réellement sélectionnés et verts, `go test -race`, `go vet`, scan de sécurité, `git diff --check`, état Git propre sur chemins non RC. TODO
- [ ] T08-P4 : archive de preuves redacted + manifeste portable + SHA-256. TODO
- [ ] T08-P5 : rapport final unique au format exact (TASK/GAP/IMPLEMENTED/NOT IMPLEMENTED/FILES CHANGED/API CHANGED/DATABASE CHANGED/UI CHANGED/TESTS WRITTEN/TESTS EXECUTED/TEST RESULTS/RACE RESULT/SECURITY SCAN/EVIDENCE/LIMITATIONS/CURRENT STATUS/NEXT ALLOWED STEP), statut exact, commit inclus. TODO

## Notes T08 (contexte opérationnel, persistées contre compaction)

- Go 1.25.13 installé dans `/usr/local/go1.25.13` (SHA-256 archive officiel `39042a078ea9ceebe3ecda4a7188f0f5b96e14a071d27923ba7f40b456e85ae3`, vérifié OK). PATH requis : `PATH=/usr/local/go1.25.13/bin:$PATH GOTOOLCHAIN=local`. `go version` → `go1.25.13 linux/amd64`.
- Dépôt Core : `/home/ubuntu/ForgeLocal`, module `forgelocal`, Go 1.25.13 via go.mod. Packages : `internal/api`, `backup`, `browser`, `config`, `fingerprint`, `groups`, `humanize`, `mcp`, `productschema`, `profile`, `profilemigration`, `runtime`, `secrets`, `spike`, `workflow`. PAS de package `launch` avant T08 (nouveau package créé : `internal/launch/manager.go` + `launch.go`).
- Format rapport final exigé (un seul rapport, à la fin) : TASK/GAP/IMPLEMENTED/NOT IMPLEMENTED/FILES CHANGED/API CHANGED/DATABASE CHANGED/UI CHANGED/TESTS WRITTEN/TESTS EXECUTED/TEST RESULTS/RACE RESULT/SECURITY SCAN/EVIDENCE/LIMITATIONS/CURRENT STATUS/NEXT ALLOWED STEP. Statut 🟢/🟡/🟠/🔵/🔴 selon preuves. Code 0 sans test sélectionné = échec. Interdit : nouvelle documentation, T09 avant validation, transformer doc/interface/test non exécuté en preuve. `PUBLIC_RELEASE_BLOCKED` indépendant.
- Critères AC-CAMO-01..05 (identiques à docs/T08_CONCURRENCY_SPEC.md) : double demande même profil = une session ou refus audité ; limite globale respectée, attente expirée libère ; crash recovery sans session fantôme ; erreur attach sans lock ; passe sous -race sans secret dans logs/audit.
- Interdictions T08 : ports runtime réels, Camoufox, lancement navigateur, proxy, backup, restauration, import, mutation UI, release.
- Specs déjà existantes : docs/T08_CONCURRENCY_SPEC.md (T08-SPEC-20260815-001, SHA 63eff82b…). Registre JSON SHA 3723b45f… Validé JSON. CDC v3.9.7 SHA 58de4650…
- Gitleaks 8.18.4 installé dans /usr/local/bin (vérifier via `gitleaks version` qui affiche "version is set by build process").
- Décisions T07 : `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION`, module autorisé `lib/concurrency.js` hash b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7, snapshot dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2, redistribution not_granted/future_distribution=blocked.
- L'implémentation actuelle (manager.go, launch.go) définit : Manager, Options (GlobalLimit 4, MaxQueueDepth 32, WaitDeadline 30s, StartTimeout 45s), Request, Stop, Recover, Status, SessionForProfile, Launcher interface, AuditEvent, Recoverer/RecoveredSession, états queued/starting/running/stopping/stopped/error/interrupted, erreurs ErrProfileAlreadyRunning/ErrGlobalLimitReached/ErrQueueFull/ErrWaitExpired/ErrCancelled/ErrContextDone/ErrInvalidProfile. ATTENTION : `newID()`, `redacted()` et éventuellement `wakeNext/passthrough` doivent être définis dans un fichier `id.go`/`redact.go` (pas encore écrits). `SessionCopy()` est méthode sur *Session. Le passthrough actuel reply avec Session de begin — à valider aux tests (le reply channel du waiter devrait recevoir la session running).

## État T08 au moment du diagnostic de tests (sauvegarde anti-compaction)

Fichiers créés : `internal/launch/manager.go` (types, Options, NewManager, Status, SessionForProfile, snapshot), `internal/launch/launch.go` (Request, begin, wakeNext, removeRequest, Stop, stopOne, Recover, await, cancelSession, record), `internal/launch/id.go` (newID), `internal/launch/redact.go` (redacted, splitTokens, isSecretish), `internal/launch/launch_test.go` (12 tests : SingleSessionPerProfile, InvalidProfile, GlobalLimit, QueueFull, CancelledWhileQueued, AttachFailure_Cleanup, Recover_NoGhostSessions, Stop_Idempotent, Stop_ReleaseAllProfiles, Audit_Redacted, ReuseAfterStop, Status_Bounds).

Signatures actuelles : `begin(ctx, sess, launcher, notify chan<- Session)` ; `failLocked(sess, err)` ; `await(ctx, sess)` attend `ctx.Done()` (ErrCancelled/ErrContextDone) ou `done` ; le goroutine attach utilise `ctx` parent avec `WithTimeout(StartTimeout 45s)` si pas de deadline ; `notify` reçoit le snapshot de session quand l'attach résout.

Diagnostic du blocage restant : TestRequest_QueueFull et GlobalLimit bloquent sur le Request d'un profil mis en queue avec launcher bloquant — `await` attend indéfiniment (ctx Background sans deadline) jusqu'au timeout 45s de l'attach, puis retourne ErrContextDone au lieu de nil, ce qui fait échouer le test. Correctif prévu : (1) QueueFull → lancer p2 en goroutine (ne pas attendre son retour), vérifier que p3 retourne immédiatement ErrQueueFull ; (2) GlobalLimit → idem pour le waiter (le waiter attend ErrWaitExpired après 2s, OK) mais vérifier ensuite que le slot libéré admet bien un nouveau profile : relancer `m.Request` après releaseAll + sleep et vérifier StateRunning au lieu de compter `Status().Running != 2` qui dépend du waiter ; (3) SingleSessionPerProfile attend `releaseAll` puis vérifie running — OK car release unbloque l'attach.

Autres points connus : wakeNext appelle begin avec qr.ctx (qui peut être expiré → attach fail immédiatement). Le test Audit_Redacted utilise redacted(suspicious) directement — les tokens courts (<25) comme "bearer ..." passent isSecretish par préfixe ; OK.

Commandes preuves à exécuter : `go test -list`, `go test -count=1 -v -race ./internal/launch/`, `go vet ./internal/launch/`, gitleaks `--no-git --source ./internal/launch --report-path out.json`, `git diff --check`, `git status`. Rapport final unique exigé (format 16 champs, voir phase T08 du plan). Go 1.25.13 dans `/usr/local/go1.25.13`, PATH=/usr/local/go1.25.13/bin:$PATH, GOTOOLCHAIN=local.

## Diagnostics restants T08 (avant compaction, sauvegarde)

Erreurs actuelles (run complet) et correctifs prévus :
1. TestAudit_Redacted : token base64 découpé par splitTokens sur "=" → "eyJhbGciOiJIUzI1NiJ9" (20 chars) passait le seuil 24. CORRECTION FAITE : seuil baissé à 16.
2. SingleSessionPerProfile l.114 "duplicate request was not audited" : le refus ne produit pas d'AuditEvent avec Event=request_refused. Cause : Request retourne directement ErrProfileAlreadyRunning sans enregistrer d'événement audit. CORRECTION À FAIRE : dans Request, avant retour ErrProfileAlreadyRunning, enregistrer AuditEvent{Event: "request_refused", Profile: p, From: StateQueued, Reason: "session already active for profile"}.
3. GlobalLimit l.150 "expected ErrWaitExpired, got <nil>" : le waiter (goroutine) est promu alors que la limite est pleine. Cause probable : wakeNext est appelé après chaque fin d'attach ; pour les 2 occupying en parallèle, le waiter est promu quand l'attach d'un occupying démarre mais la limite est pleine. Vérifier wakeNext : il promeut si len(running) < GlobalLimit — mais l'attach des 2 occupying ajoute running=2 dès begin (avant Attach). Le waiter est mis en queue à l'instant t=50ms ; à ce moment running=2 déjà (les 2 goroutines ont appelé begin). Le waiter attend. Mais wakeNext est aussi appelé dans le goroutine du waiter lui-même ? Non. L'attach du waiter ne démarre que si wakeNext promu. L'attach réussit (nil err) → waiter retourne nil au lieu d'ErrWaitExpired. Cause : après 2s de deadline, le goroutine waiter n'expire pas ? Le waiter est promu AVANT que la deadline expire car… non. Vérifier : releaseAll n'est appelé qu'après. Hmm : les 2 occupying ne bloquent pas m.running pendant 50ms, mais begin met running à jour. Le waiter mis en queue ; wakeNext appelle begin sur le waiter avec ctx du waiter (deadline 2s). L'attach du waiter avec blockingLauncher bloque jusqu'à releaseAll. La deadline expire, ctx.Done → Request devrait retourner ErrWaitExpired… sauf si le waiter a été admis ? Oui : wakeNext promu quand len(running) < 2 : mais running=2 (les 2 occupying sont dans running dès begin). wakeNext ne devrait pas promouvoir. Le waiter retourne nil — donc il a reçu un snapshot de running via notify. Les 2 occupying, après releaseAll, passent à running (final=running) et notify nil (reply fast path). Pas de notify vers le waiter. ALORS le nil vient de… le waiter a été promu et son attach… non releaseAll est appelé après l'assert. Mystère : peut-être que les occupying N'ont pas atteint begin à t=50ms (race) et que running < 2 au moment de la mise en queue ; le waiter passe en fast path avec ctx deadline 2s, commence l'attach blocking, deadline expire → ErrWaitExpired attendu. Mais il retourne nil → l'attach a réussi avant la deadline, ce qui est impossible sans releaseAll. Sauf si le waiter ctx (sans deadline parente) : Request dérive WithDeadline 2s, puis begin : hasDeadline=true donc cancel=func(){} → l'attach NE reçoit pas de deadline ? Non, le ctx passé à begin est qr.ctx qui a bien deadline. Attach reçoit ctx → blocking → ne revient jamais sans releaseAll. Mais il retourne nil. Possible : le waiter ctx expire, Request retourne ErrWaitExpired MAIS le goroutine du test lit err nil… non. Attendez : le test lit err du waiter : "got <nil>". Le seul chemin nil est `case out := <-qr.result: return out, nil` ou le fast path reply. Le waiter n'a pas de reply channel fast path… il a qr.result, rempli par notify dans le goroutine attach uniquement après l'attach. Donc l'attach doit avoir résolu. Impossible sans releaseAll SAUF si wakeNext est appelé par un autre mécanisme. HYPOTHÈSE FINALE : les 2 occupying ne remplissent PAS m.running car begin met running MAINTENANT ; mais entre le moment où les 2 goroutines occupying lancent Request et le sleep 50ms, les 2 begin ont bien mis running=2. Le waiter est mis en queue. wakeNext : running=2 >= 2 → pas de promotion. OK. Puis rien. Deadline 2s → Request retourne ErrWaitExpired. Mais le test lit nil. ERREUR DE LECTURE DU LOG ? Non. AUTRE HYPOTHÈSE : les 2 occupying en parallèle font Request qui appelle begin pour chacun. Le premier begin fait wakeNext ? Non, wakeNext est appelé APRÈS l'attach (fin de goroutine). Donc après attach du premier occupying : running encore 2 (le second est dans running). wakeNext : running=2 → pas de promotion. Après le second occupying : même. Donc le waiter reste en queue. Timeout → ErrWaitExpired. Le test retourne nil → le waiter a fait le chemin `out := <-qr.result`… qr.result jamais écrit sans begin. IMPOSSIBLE SAUF concurrent wakeNext d'un STOP ou recover. Donc : mon analyse ignore un appel de wakeNext supplémentaire. Chercher wakeNext dans tout le fichier (Stop ? stopOne n'appelle pas wakeNext ; cancelSession non). SEUL appel : fin du goroutine attach + wakeNext. Le goroutine waiter attach : begin fait goroutine attach, après begin PAS de wakeNext. Hmm. Je re-lance avec -v -race pour tracer. Peut-être que les 2 occupying n'ont PAS démarré à t=50ms : une des deux est en queue ! Si un occupying n'a pas atteint begin à t=50ms, running=1, le waiter passe fast path (deadline 2s), begin waiter, running=2. Attach blocker pour 3 (1 occupying + waiter). Le 2e occupying arrive, byProfile libre ? Non. Il attend dans Request ? Non, running=2 → mis en queue. Puis un occupying termine son attach après releaseAll → wakeNext → promu le 2e occupying → son attach → etc. Le waiter avec deadline 2s expire → retourne ErrWaitExpired. Toujours pas nil. SAUF : releaseAll débloque le waiter attach (deadline non expirée) → nil. Si releaseAll est appelé avant 2s ? Le test : waiter démarre, puis `select case err:=<-done: ... case time.After(5s)` ; releaseAll est appelé APRÈS le select. Donc releaseAll après la réponse. Le done reçoit nil en <2s → impossible sans releaseAll… À INVESTIGUER : je vais tracer avec -race et prints.

4. AttachFailure l.256 "profile not released after attach failure" (5s timeout) : failOnce.Swap(true) dans le goroutine Request… mais swap retourne true à chaque appel après le premier ; le second attach (launcher2) échoue aussi ! Le failOnce reste true pour le second Request. CORRECTION À FAIRE : reset failOnce après le premier échec (ou recharger un nouveau launcher pour le second déjà fait — launcher2 est nouveau, failOnce par défaut false). Mais l'erreur vient du SECOND done timeout 5s → le second Request bloque 5s. Le second Request (fail-profile) : l'attach de launcher2 (failOnce=false) bloque jusqu'à releaseAll ? Non, releaseAll est dans le cas succès du second. Le premier done reçoit après 5s → l'attach du premier (failOnce true) ne retourne pas immédiatement ? context.Canceled err immédiate → failLocked → notify → first done reçu. Mais le premier select a timeout 5s : donc notify jamais écrit. Cause : failLocked supprime m.running avant que le goroutine waiter puisse… notify est écrit après failLocked, devrait marcher. SAUF race avec ctx du waiter : le waiter Request(fail-profile) met en queue si running=1 ? Le fail-profile n'est pas encore dans running au premier Request… first Request : byProfile vide, running=0 → fast path, reply channel. Le goroutine attach de first : err (failOnce true) → failLocked → notify(reply, final StateError). done du test reçoit StateError. Le log dit timeout 5s donc reply jamais écrit. Pourquoi ? failOnce.Store(true) ; Swap(true) → first call retourne true → err = context.Canceled → ok err non-nil → failLocked → notify. Ça devrait marcher. Peut-être que le first select a reçu mais State != StateError → t.Fatal "expected error state"? Non le log dit "profile not released" = second select timeout. Donc premier OK (StateError reçu) mais second bloque. Second Request : byProfile est vide (failLocked l'a retiré), running=0 → fast path reply launcher2. attach launcher2 bloque (failOnce false) jusqu'à releaseAll (qui est appelé dans le select du second). DEADLOCK : le second attend done2, releaseAll après. Correctif : releaseAll avant, puis drain done2 (comme SingleSessionPerProfile).
5. Recover l.293 : même pattern — le Request du ghost (launcher blocking) attend releaseAll jamais appelé avant. Je drain après releaseAll.
6. Stop_ReleaseAllProfiles l.335 : le second Request (après Stop) retourne state error. Le launcher2 attach bloque jusqu'à releaseAll… le test attend done avec timeout 5s puis releaseAll. Mais "expected starting/running after stop, got error" → le retour est state=error. Cause : le waiter est promu après releaseAll par wakeNext (appelé après l'attach de l'occupying, mais l'occupying est stoppé par Stop qui attend done via stopOne). Stop attend stopOne qui attend m.attached[id] closé → done closé à la fin de l'attach. L'attach (blocking) bloque jusqu'à releaseAll. Donc Stop ne retourne pas avant releaseAll. Le test fait Stop(stopCtx 2s) → attend 2s timeout → le goroutine attach de l'occupying continue en arrière-plan. Puis le second Request : running encore 1 (occupying toujours dans running) → queue (deadline 5s du WaitDefault 30s? options nil → WaitDeadline 30s). Le waiter attend. Mais done retourne en 5s avec state error : le waiter promu par wakeNext ? Stop ne libère pas les slots car l'attach continue… stopOne delete running après done. Quand releaseAll arrive : l'attach de l'occupying réussit, fait fail? non success → running StateRunning, notify nil (fast path reply non utilisé car stop en cours). wakeNext → pas de queue ? le second est en queue → promu. running=1 (l'occupying est encore running) + waiter promu → running=2 → final waiter starting. Done2 reçoit state Starting. Mais le test lit state error. HYPOTHÈSE : le goroutine du waiter est promu et son attach (launcher2 blocking) bloque ; la deadline 30s ne passe pas en 5s… Le test lit "error" → final StateError. Pourquoi error ? le ctx du waiter (qr.ctx dérivé du ctx du test, pas annulé) n'expire pas en 5s. Autre possibilité : Stop, à la fin de stopCtx (2s), stopOne timeout → done closée (defer close) → stopOne delete running → le goroutine attach de l'occupying continue et fera wakeNext en succès → mais notify nil. Le waiter : running=0 après stopOne → wakeNext promu waiter → attach waiter blocking → done2 timeout 5s → le test lit err "profile not released"? Non "expected starting/running after stop, got error". Donc done2 a reçu un Session avec StateError. Le waiter est promu, son attach : err nil après releaseAll → final StateRunning. Mais done2 reçoit en 5s (timeout du select) un state error… Le select 5s expire → t.Fatal "reuse admission did not resolve in time" → mais le log dit got error (pas timeout). Donc done2 a reçu. Le waiter promoted, son attach bloque, releaseAll débloque → running. Je dois tracer en réel. PLAN : relancer avec prints temporaires OU accepter de simplifier : utiliser un launcher à auto-release (non blocking) avec gate manuel par test. Je réécris le launcher pour avoir des tests plus déterministes : `gateLauncher` avec un canal de libération par attach.

Décision architecture : remplacer blockingLauncher par un launcher qui retourne immédiatement l'erreur via un hook, et contrôle l'état attach via un gate global plus fin. Mais plus simple : garder blockingLauncher avec releaseAll ET respecter l'ordre release-avant-drain dans TOUS les tests. Les échecs restants viennent de l'ordre des appels.

## État T08 (sauvegarde complète avant compaction)

Le package `internal/launch` (Go clean room, Manager/SessionManager) est implémenté dans `internal/launch/{manager.go,launch.go,id.go,redact.go,launch_test.go}` avec les AC-CAMO-01 à 05. L'environnement : Go 1.25.13 installé dans `/usr/local/go1.25.13` (PATH en tête), `GOTOOLCHAIN=local`.

Progression des tests (dernière exécution) : sur ~12 tests, il reste un seul échec, `TestRequest_GlobalLimit` (ligne 167 : « expected ErrWaitExpired, got <nil> »). Les autres tests passent désormais, y compris `TestAudit_Redacted` (corrigé par le seuil `isSecretish` abaissé de 24 à 16 caractères dans `redact.go`), `TestRequest_SingleSessionPerProfile` (corrigé par l'ajout de l'`AuditEvent` `request_refused` avant le retour `ErrProfileAlreadyRunning` dans `launch.go`), et les tests de concurrence/Stop (corrigés par `waitForRunning`, `reqCtx` bornés et `releaseAll` idempotent via `sync.Once`).

Diagnostic du dernier échec : les 2 occupying en goroutine ne sont pas tous les deux dans `m.running` au moment où le waiter est lancé en goroutine (le log montre `running=1 queued=0`). Le waiter passe donc en fast path avec un ctx à deadline 2s, ce qui bloque le test. Correctif prévu : utiliser `waitForRunning(t, m, 2)` au lieu du sleep 50ms, comme dans les autres tests. Le fichier contient deux prints temporaires (lignes 156-157) à retirer après correction.

Prochaines étapes : (1) corriger GlobalLimit, (2) exécuter la suite complète + `go test -race` + `go vet`, (3) scanner Gitleaks 8.18.4 du delta, (4) produire l'archive de preuves redacted avec manifeste portable et SHA-256, (5) livrer le rapport final au format exact exigé (TASK/GAP/IMPLEMENTED/NOT IMPLEMENTED/FILES CHANGED/API CHANGED/DATABASE CHANGED/UI CHANGED/TESTS WRITTEN/TESTS EXECUTED/TEST RESULTS/RACE RESULT/SECURITY SCAN/EVIDENCE/LIMITATIONS/CURRENT STATUS/NEXT ALLOWED STEP) avec le statut exact.

Contraintes à maintenir : aucun lancement de navigateur/Camoufox, pas de proxy, pas de mutation UI, pas de backup/restore/import, pas de release ; `PUBLIC_RELEASE_BLOCKED` et `SCAN_BLOCKED_UNKNOWN` (BACK-01) maintenus. Le dashboard reste au checkpoint T06 (`manus-webdev://2b50697e`).

## État T08 (sauvegarde — avant compaction, reprise)

Le package `internal/launch` (Go clean room) est implémenté dans `internal/launch/{manager.go,launch.go,id.go,redact.go,launch_test.go}` : Manager/SessionManager, queue bornée, sérialisation par profil (`byProfile`), timeout (`StartTimeout` 45s par défaut, `WaitDeadline`), `context.Context`, annulation (`ErrCancelled`/`ErrContextDone`), cleanup idempotent (`Stop`, `cancelSession`), crash recovery (`Recover` sans session fantôme), audit redacted (`redact.go`, seuil 16, sentinelles). AC-CAMO-01 à 05 couverts.

Environnement : Go 1.25.13 dans `/usr/local/go1.25.13` (ajouter au PATH en tête), `GOTOOLCHAIN=local`. Commande de test : `export PATH=/usr/local/go1.25.13/bin:$PATH && export GOTOOLCHAIN=local && cd /home/ubuntu/ForgeLocal && go test -count=1 -timeout=110s ./internal/launch/`.

Correctifs appliqués aux tests : (1) `waitForRunning` pour remplacer les sleep ; (2) `reqCtx` borné (5s) + goroutine + `releaseAll` avant drain pour tout Request avec launcher blocking ; (3) `failOnce` single-shot → un `newBlockingLauncher()` distinct par session (sinon le 2e attach échoue immédiatement) ; (4) `releaseAll` idempotente via `sync.Once` ; (5) capture de variable de boucle corrigée (`go func(idx)` ou profils `-a`/`-b`) ; (6) refus audit : `AuditEvent{Event:"request_refused"}` ajouté avant `ErrProfileAlreadyRunning` ; (7) seuil `isSecretish` abaissé de 24 à 16.

Dernier état : build échoue avec « cannot use s as error value in send » ligne 198 — le `s` du done est maintenant masqué par le `st` de Status ; mais l'erreur persiste : c'est l'inverse, le `s` du done shadow le `s` de Status dans une autre ligne, ou le sed précédent a cassé une autre occurrence (`st.Running` appliqué là où `s` devait rester). Vérifier `sed -n '176,212p' internal/launch/launch_test.go` pour voir les lignes 198 et 204 : le `s` du case done <- s doit être renommé en `snap`, et le `s` du goroutine Request (ligne ~193 `s, err := m.Request`) renommé en `snap` aussi.

Ensuite : relancer la suite, puis `go test -race`, `go vet`, Gitleaks 8.18.4 sur le delta (`gitleaks --no-git --source internal/launch`), puis archive de preuves redacted avec manifeste portable (SHA-256 relatif), puis rapport final au format exact exigé (TASK/GAP/IMPLEMENTED/NOT IMPLEMENTED/FILES CHANGED/API CHANGED/DATABASE CHANGED/UI CHANGED/TESTS WRITTEN/TESTS EXECUTED/TEST RESULTS/RACE RESULT/SECURITY SCAN/EVIDENCE/LIMITATIONS/CURRENT STATUS/NEXT ALLOWED STEP) avec statut exact. Dashboard inchangé (`manus-webdev://2b50697e`). Contraintes : pas de navigateur/Camoufox/proxy/mutation UI/backup/release ; `PUBLIC_RELEASE_BLOCKED` maintenu.

## État T08 (sauvegarde — suite)

Tous les 12 tests passent sans -race (8.35s). Avec -race, seul `TestStop_Idempotent` échoue : DATA RACE entre `stopOne` (launch.go:260, écriture `st.State = StateStopping`) et `begin.func1` (launch.go:137, lecture/écriture `st.State = StateRunning`), avec `Stop.func1` (launch.go:244) et le goroutine attach de S1.

Analyse : dans `Stop`, les sessions sont copiées par valeur (`*s` en argument), mais stopOne refait `st, ok := m.running[sess.ID]` (pointeur partagé avec begin). begin.func1 (fin d'attach) modifie st.State=StateRunning sous mu ; stopOne modifie StateStopping sous mu, MAIS stopOne lit aussi `m.attached[sess.ID]` hors lock (select) et prend/relâche mu plusieurs fois entre les deux accès, permettant un interleave avec le lock de begin.

Correctif décidé : (1) dans `Stop`, capturer `done := m.attached[id]` sous le lock et le passer à stopOne (plus de lecture de map hors lock) ; (2) dans stopOne, lire une seule fois sous lock les champs nécessaires (copier Session), puis faire les mutations sous lock en une seule section ; (3) idéalement un seul lock/section dans stopOne.

Commandes de vérification : `export PATH=/usr/local/go1.25.13/bin:$PATH && export GOTOOLCHAIN=local && cd /home/ubuntu/ForgeLocal && go test -race -count=1 -timeout=170s ./internal/launch/`. Tests : `go test -list . ./internal/launch/` (12 tests listés).

Ensuite : `go vet ./internal/launch/`, Gitleaks 8.18.4 `--no-git --source internal/launch`, archive de preuves redacted (/tmp/t08-evidence/) avec manifeste portable + SHA-256 (hashes relatifs), rapport final au format exact (TASK/GAP/IMPLEMENTED/NOT IMPLEMENTED/FILES CHANGED/API CHANGED/DATABASE CHANGED/UI CHANGED/TESTS WRITTEN/TESTS EXECUTED/TEST RESULTS/RACE RESULT/SECURITY SCAN/EVIDENCE/LIMITATIONS/CURRENT STATUS/NEXT ALLOWED STEP), statut exact selon preuves.

Gitleaks installé, version 8.18.4. Go installé dans /usr/local/go1.25.13.

## État T08 (suite — 2e fix race)

Les 3 races restantes : stopOne écrit `st.State`, `st.StoppedAt`, `st.Err` (lignes 269-271) pendant que begin.func1 lit le même `*st` via `final = *st` (ligne 151). Le pointeur `st := m.running[sess.ID]` est partagé entre le goroutine attach (begin) et stopOne (Stop) — les deux sont sous `m.mu` séparément, mais la race est réelle car ils peuvent s'entrelacer (attach réussit et stopOne nettoie).

Correction choisie (déterministe) : introduire un canal `stopped` par session : quand stopOne nettoie une session encore attachée, il attend le close de `done` (attach fini) avant d'écrire, ET begin à la fin : s'il est interrompu par Stop, le `m.running[sess.ID]` peut avoir été supprimé → branch `else` (StateInterrupted) — mais le `*st` lu dans `final = *st` doit être protégé : begin prend le lock, lit et copie. stopOne doit prendre le même lock. Les deux sont sous mu, donc la race devrait disparaître... MAIS le pointeur `st` du running est le même objet ; le read de begin (l.151 `final = *st`) se fait sous lock begin, stopOne écrit sous lock stopOne — les locks sont distincts (même mu !) — donc synchronisés. Le race detector signale quand même ? Non, si les deux sous mu, pas de race. La race réelle vient d'ailleurs : l'adresse 0x00c0000e6b60/... ce sont les champs de l'objet *Session pointé. Le read de begin (l.137 offset 0x244) = la closure capture `sess` et aussi... l.137 = `err := launcher.Attach(ctx, sess)` — le read du champ State du `sess` (copie) fait par Attach ? Attach(ctx, s Session) reçoit une copie. Pas de partage. Offset 0x244 dans la closure = probablement le `m.snapshot(&sess)` (rec = m.snapshot(&sess)) ligne 120 dans mon édition (l.118 avant) ? Le snapshot lit les champs du pointeur &sess (le même pointeur partagé avec stopOne). `m.snapshot(&sess)` se fait SOUS le lock begin (begin prend le lock au début). stopOne prend le même mu. Pas de race possible avec mu commun... SAUF si le snapshot est fait APRÈS le lock release : begin prend mu, fait m.running[sess.ID]=&sess, byProfile, snapshot, attached, unlock, puis goroutine attach. Le snapshot est sous lock. stopOne sous le même mu. Pas de race. Donc le race detector détecte une lecture/écriture réelle entre deux goroutines sans protection commune : stopOne écrit StateStopping/Stopped et le goroutine begin (attach) lit... La lecture du goroutine begin (attach, func1) : la seule lecture en dehors du lock est le `select <−m.attached[sess.ID]` ? Non, begin.func1 n'a pas ça. Le goroutine begin lit `ctx.Done`... hmm. Le read est à launch.go:137 offset 0x244. La ligne 137 dans le fichier actuel = `err := launcher.Attach(ctx, sess)` — mais le offset 0x244 suggère plus loin dans la closure (le code après Attach). Peut-être que la ligne source est 148 `st.State = StateRunning` (écrite après lock) — non. Je vais simplement corriger toutes les lectures partagées : (1) stopOne : attendre done AVANT de verrouiller (déjà fait via w.done) — fait ; (2) stopOne : utiliser une copie du pointeur n'est pas possible ; (3) le plus robuste : stopOne ne doit JAMAIS écrire dans le même objet que begin écrit à la fin. Solution : séparer les états — introduire le champ Stopping dans Session mais l'écrire dans un objet copié : stopOne copie la Session SOUS le lock (valeur) puis met à jour le copy, puis sous le lock : met à jour m.running (state final) et copie la version copiée. Mais la race vient du fait que begin écrit `st.State = StateRunning` sur le même pointeur que stopOne écrit `st.State = StateStopped`. S'ils sont tous les deux sous m.mu, pas de race réelle. DONC le race réel doit être : begin écrit st.State=StateRunning APRÈS que stopOne a déjà pris son lock et vérifié m.running (ok) puis relâché, puis begin prend le lock et écrit StateRunning — mais stopOne a DÉJÀ supprimé m.running[id] ? Non, stopOne prend le lock, supprime, unlock. Puis begin prend le lock, m.running[sess.ID] n'existe plus → branch else (StateInterrupted). Pas de partage d'écriture. Les races restantes doivent donc concerner un read hors lock : begin.func1 `launcher.Attach(ctx, sess)` — sess est la copie passée à la goroutine. Attach(ctx context.Context, s Session) — la copie du goroutine. Pas de partage. SAUF : le goroutine attach capture `sess` (la variable du scope de begin) — c'est la même copie. Pas de partage avec le running. La race est ailleurs : les offsets 0x00c0000e6b60, ...bb8 sont les champs State/StoppedAt/Err du même objet Session — celui dans m.running. Le read par goroutine 48 (begin.func1) à l.137 = le read de begin.func1 d'un champ de cet objet. La seule lecture de cet objet hors du lock begin est... `m.snapshot(&sess)` est dans begin (pas func1). Dans func1, les lectures de cet objet : `m.running[sess.ID]` (sous lock), `notifyNotifyLocked` (sous lock). HORS lock : `m.wakeNext()` (après unlock) — wakeNext lit les sessions running (sous lock dans wakeNext). Donc toutes les lectures sont sous mu. Je ne trouve pas la lecture hors lock. Décision : je vais reproduire avec un petit log d'adresse et de champ, ou mieux : je mets un `sync.Mutex` de session ? Trop complexe. Option pragmatique : le test Idempotent est le seul qui expose la race (Stop pendant que les attaches sont encore en cours, blocking, releaseAll après). Le pattern réel : Stop avec attaches en cours — les goroutines begin continuent, stopOne attend done. Les races viennent du fait que stopOne écrit dans l'objet pendant que begin écrit dans le même objet (les deux prennent mu mais PAS le même instance ? si stopOne et begin utilisent le même m.mu, les deux sections sont sérialisées — la race est impossible. À moins que le race detector signale deux accesses par des goroutines différentes sur le même objet, chacune protégée par un lock DIFFÉRENT — mais il n'y a qu'un mu. À moins que... stopOne et begin ont des m différents ? Non. Vérifions que stopOne et begin utilisent bien m.mu (même instance). Oui. DONC pas de race réelle, mais le race detector signale quand même → il doit y avoir un read hors lock que je ne vois pas. Je vais simplement instrumenter le code avec -race et le traceur pour localiser la lecture exacte.

## État T08 (suite — analyse race détaillée)

Le code begin actuel : begin (l.114-158) fait sous mu : m.running[sess.ID]=&sess, byProfile, snapshot(&sess), attached[id]=done, puis lance goroutine func1 qui appelle Attach puis sous mu : failLocked ou st.State=StateRunning/final=*st. stopOne (l.257-285) : select <−w.done (copie du canal), puis sous mu : m.running[id] → st.State=StateStopped/StoppedAt/Err/delete.

Le race : stopOne écrit st.State/StoppedAt/Err (l.269-271) pendant que begin.func1 LIT le même *st (l.148-151 : st=m.running, st.State=StateRunning est un write aussi, mais final=*st est un read de l'objet entier). Les deux sont sous mu. MAIS la race real : le goroutine Stop (l.253) lit/écrit `w` hors lock ? Non. Le « Previous read by goroutine 48 (begin.func1) » à l.137 offset 0x244 : je compte les octets — 0x244 dans la closure. La ligne 137 est `err := launcher.Attach(ctx, sess)` — Attach reçoit sess PAR VALEUR (copie). Le read de begin.func1 d'un champ de l'objet partagé peut être... `m.wakeNext()` après unlock (l.162) lit m.running sous lock dans wakeNext. Les writes de stopOne sont sous lock. Les reads de wakeNext sont sous lock. PAS DE RACE RÉELLE selon mon analyse.

MAIS le race detector signale quand même → un read hors lock existe quelque part. L'offset 0x244 après l.137 : dans la closure, le seul read hors lock d'un champ de l'objet *Session = AUCUN selon mon code. SAUF : `sess.State` dans `if err != nil { final = sess; final.State = StateError }` — sess est la copie locale de begin, pas le pointeur partagé. Hmm.

Dernière hypothèse : `failLocked(sess, err)` modifie `m.running[sess.ID]` (l'objet pointé) — failLocked sous lock. stopOne sous lock. OK.

Décision finale : je vais rendre stopOne SANS modifier le pointeur partagé du tout : stopOne attend done (close par begin.func1 à la fin de l'attach). Après done, begin.func1 a DÉJÀ fait son write final (StateRunning). Le conflit read/write n'est possible que si stopOne s'exécute PENDANT que begin.func1 est dans sa section critique. Mais done est close par defer close(done) AVANT le lock de begin.func1 ? Non, `defer close(done)` est au début de func1, donc close(done) se fait à la FIN, APRÈS le lock release (defer empilé : le dernier déclaré est exécuté en premier... `defer close(done)` déclaré en premier, exécuté en DERNIER, après `defer cancel()`). Donc done closé APRÈS tout le traitement de func1 (lock release inclus). Après done, stopOne peut prendre le lock et écrire StateStopped sans conflit avec func1 (qui a fini). Le read de func1 (`final=*st`) se fait PENDANT que stopOne attend done — stopOne n'a pas encore écrit. Le write de stopOne se fait APRÈS que func1 a fini. PAS DE CONFLIT.

DONC le race détecté est FAUX-POSITIF réel dû à... le pointeur `st` du running est lu (final=*st) et écrit (st.State=StateRunning) par func1, et stopOne écrit (st.State=StateStopped). Le goroutine func1 lit le champ State via final=*st AVANT le lock (non, après). Les deux locks sont le même mu. Le race detector ne devrait pas signaler.

Je relance avec le fix : je rends stopOne sans la lecture/écriture du pointeur partagé APRÈS le done : puisque done est closé par func1 après TOUT son traitement, stopOne n'a plus qu'à attendre done puis simplement delete m.running[id] (sans écrire State dans le pointeur partagé, mais en gardant une copie locale pour l'audit). Le write `st.State = StateStopped` sur le pointeur partagé disparaît : stopOne fait m.mu.Lock, st:=m.running[id], copie := *st (copie de valeur), puis delete, puis utilise copie pour l'audit. Le StateStopped est écrit dans la copie, pas dans le pointeur partagé. Comme ça, plus aucune écriture partagée par stopOne.

## T08 — preuves finalisées (2026-08-16)

12 tests écrits et exécutés, tous PASS (8.348s sans race, 9.355s avec -race, zéro DATA RACE) : SingleSessionPerProfile, InvalidProfile, GlobalLimit, QueueFull, CancelledWhileQueued, AttachFailure_Cleanup, Recover_NoGhostSessions, Stop_Idempotent, Stop_ReleaseAllProfiles, Audit_Redacted, ReuseAfterStop, Status_Bounds. `go vet` PASS. Gitleaks 8.18.4 `--no-git` sur `internal/launch/` : JSON `[]` (0 fuite). `git diff --check` OK, zéro fichier `release/`/`dist/` modifié.

Correctifs notables : race stopOne/begin corrigée (stopWork snapshot + copie, stopOne n'écrit plus dans le pointeur partagé), sentinelle de test fragmentée (redactor toujours défectueux détecté).

Archive preuves : `/tmp/T08-PROOF-20260816.zip`, SHA-256 `3d4a13847278779b4e5c696875bd817d86b2903913986aac79d97d988db2ea35`, ZIP intègre, rescan gitleaks de l'archive : 0 fuite. Manifeste portable dans l'archive (T08-MANIFEST.md) + SHA256SUMS relatifs (11/11 OK après extraction).

Rapport final : format exact TASK/GAP/.../NEXT ALLOWED STEP. Statut : 🟢 COMPLET si toutes preuves présentes — les preuves sont complètes pour le périmètre T08 spécifié (tests, race, vet, scan, manifeste), donc statut COMPLET pour T08. Next allowed step : T09 après validation finale T08 par l'utilisateur.

## Audit T08 reçu (16 août 2026) — 3 écarts à traiter

Verdict de l'audit : `T08_EVIDENCE_READY_WITH_BLOCKING_GAPS`. SHA/ZIP/manifeste/tests/race/vet/gitleaks/périmètre : PASS. Trois écarts :
1. **BLOQUANT** — `TestConcurrentStress` avec `t.Parallel` + goroutines multiples exigé par `T08_CONCURRENCY_SPEC.md` : absent. À ajouter et exécuter.
2. **BLOQUANT** — code non commité dans ForgeLocal. À commiter, puis rejouer tests/race/vet/gitleaks sur la révision cible et compléter T07-R1 (hash commit).
3. **À clarifier** — preuve de cleanup sans goroutine bloquée dans un délai borné (contrôle de terminaison explicite des goroutines).

Actions : (a) ajouter TestConcurrentStress (stress : beaucoup de profils/goroutines parallèles, queue pleine, annulations simultanées, stop pendant attach, puis join WaitGroup borné) ; (b) vérifier/ajouter dans le code la preuve de terminaison des goroutines (le Manager n'a pas de WaitGroup global ; ajouter un mécanisme de suivi/WaitGroup rejoint dans Stop avec délai borné, ou prouver dans le test que toutes les goroutines attach se terminent dans un délai borné après releaseAll/Stop) ; (c) commit dans ForgeLocal (branche ou main ? → commit sur la branche courante du dépôt local ForgeLocal, puis push ? L'utilisateur n'a pas fait push jusqu'ici — commit local + preuve du hash commit suffit pour T07-R1, push à confirmer) ; (d) nouvelle archive T08-R2 avec hash commit, logs complets, manifeste, rescan ; (e) rapport final.

Contraintes : pas de lancement navigateur/runtime/Camoufox, pas de proxy/UI/backup/release, PUBLIC_RELEASE_BLOCKED maintenu, dashboard T06 inchangé.

## Plan de traitement des 3 écarts (état avant implémentation)

Structure actuelle du code (relue) :
- manager.go : Manager {mu, opt, running, byProfile, queue, attached, sink} ; Options (GlobalLimit, MaxQueueDepth, WaitDeadline, StartTimeout, Recoverer) ; Launcher.Attach(ctx, s Session) error ; Session struct avec Err redacted.
- launch.go : begin(ctx, sess, launcher, notify) lance une goroutine attach qui prend mu à la fin (failLocked/notify final), puis wakeNext() HORS lock. Stop : snapshot stopWork sous lock, stopOne attend w.done ou ctx, lock final (delete), record.
- Le Manager n'a PAS de WaitGroup global de terminaison des goroutines attach. Chaque goroutine attach signale sa fin uniquement via close(done) et wakeNext.

Plan :
1. Ajouter à Manager un champ `attach sync.WaitGroup` (privé) : begin fait m.attach.Add(1) AVANT le lancement de la goroutine (sous mu, avec le pointeur Manager), la goroutine fait defer m.attach.Done() après tout (après wakeNext inclus — mettre Done à la toute fin). Stop : après wg.Wait() des stopOne, m.attach.Wait() prouve que toutes les goroutines attach sont terminées. ATTENTION : Stop doit d'abord signaler l'arrêt, puis attendre. Le ctx de Stop fait déjà fail/cancel les attaches restantes ? StopOne attend w.done — si l'attach continue (launcher bloquant), w.done jamais closé → stopOne attend ctx.Done (le ctx de Stop). Les attaches bloquantes ne se terminent jamais dans ce cas → attach.Wait() bloquerait. Solution : le Manager expose aussi Close/Stop qui annule le ctx des demandes en queue ET… le ctx des attaches est celui du Request (caller). Pour le stress test : utiliser des ctx avec deadline pour que toutes les attaches se terminent dans un délai borné, puis join. Le test prouvera la terminaison bornée (pas une fuite).
2. Ajouter dans Options ou API : m.Join(ctx) ou StopJoin : Stop + attend attach.Wait() avec un ctx borné (retourne si timeout). Ou simplement : Stop attend attach.Wait() après avoir annulé. Je vais ajouter `stopCtx` via un ctx global manager : NewManager crée un ctx interne + cancel ; Stop() : cancel interne (toutes les goroutines attach héritent ? Non, le ctx de l'attach vient du Request). Alternative simple et robuste : la goroutine attach reçoit en plus le ctx interne du manager (context.Any? Go 1.21+ context.AfterFunc, ou créer ctx = merge(ctx, m.ctxStop) via goroutine). Implémenter `ctxAny` : dérivé, cancel déclenché si m.ctxStop OU ctx d'origine. begin(ctxRequest...) → dans begin, créer ctxMerged = combine(ctxRequest, m.ctx) (goroutine qui select). La goroutine attach hérite ctxMerged. m.Stop : m.cancelStop() + stopOne. Toutes les attaches en cours voient ctxMerged fermé.
3. TestConcurrentStress : t.Parallel, N=100-200 goroutines avec 20-40 profils, limite 4, queue 32, mix (attach bloquant + fail + ctx annulés à mi-parcours), tout avec ctx internes bornés 3s ; après Stop, join borné 5s : prouve zéro goroutine attach active (m.attach.Wait() ne bloque pas), running=0, queued=0.
4. Commit ForgeLocal : git add internal/launch/ + go.mod? (pas de dépendance) ; commit message décrivant T08 ; noter le hash commit pour T07-R1 ; pas de push demandé explicitement mais l'utilisateur travaille avec un repo IPcache séparé pour preuves — le repo ForgeLocal est local. Je commite dans le repo local et donne le hash ; si un push est attendu, demander.
5. Rejouer preuves sur la révision : go test, go test -race, go vet, gitleaks (package + archive), manifeste.
6. Archive T08-R2 + rapport.

Commandes de référence : export PATH=/usr/local/go1.25.13/bin:$PATH && export GOTOOLCHAIN=local ; repo /home/ubuntu/ForgeLocal ; Gitleaks 8.18.4 ; go1.25.13.
Dashboard webdev : inchangé, checkpoint 2b50697e, auto-publish.

## Diagnostic TestConcurrentStress (deadlock)

Stacktrace SIGQUIT :
- ~104 goroutines [sync.Mutex.Lock] dans Manager.Request à launch.go:27 (m.mu.Lock initial) — toute la meute attend le lock.
- Stop : [sync.Mutex.Lock] à launch.go:255 (m.mu.Lock dans Stop) — attend le lock aussi.
- 4 goroutines begin.func1 à launch.go:150 : `m.mu.Lock` à la fin de l'attach (snapshot final / failLocked / notify).
- Aucune goroutine ne détient le lock en ce moment... toutes l'attendent ! C'est le symptôme d'un unlock oublié/panic, ou... toutes attendent le même mutex ET personne ne le tient = le détenteur a terminé sans libérer ? Impossible en Go sauf... AH : le goroutine mergeContexts : dans begin, `ctx = mergeContexts(ctx, m.ctx)` lance une goroutine par attach (120) qui select <−a.Done / <−b.Done puis cancel(). Ces goroutines merge ne bloquent pas le mutex.
- Revoir : 116 goroutines func2 (caller Request) bloquées sur m.mu.Lock à launch.go:27. 4 func1 bloquées à l.150. Stop bloqué à l.255. Le détenteur du lock ? Chercher une goroutine dans wakeNext (lock held) — aucune mention. Aucune goroutine ne tient le lock au moment du dump. Si personne ne le détient, c'est que toutes les goroutines qui pourraient l'obtenir sont au moment du dump justement toutes en attente du lock... mais qui l'a relâché en dernier ? Le lock est libre au moment du dump (sinon quelqu'un le détenait avec un état différent). S'il est libre, l'une des goroutines devrait l'obtenir. À moins que... toutes soient "runable" mais pas exécutées ? Non, il y a des threads idle. C'est impossible sauf... la goroutine du test func3 (Stop) attend le lock... et le dump montre stop à l.255. Le mutex est libre ? On ne sait pas.
- HYPOTHÈSE FORTE : le détenteur est une goroutine begin.func1 qui, après m.mu.Unlock (l.169 de mon code = après wakeNext), ne fait plus rien... non.
- AUTRE HYPOTHÈSE : wakeNext est appelé HORS lock à la fin de func1, mais prend le lock. Toutes les 4 func1 attendent le lock... mais une d'entre elles l'a peut-être et est en wakeNext dans le dump ? Non montrée.
- ATTENDS : le dump montre toutes les Request attendent m.mu.Lock à l.27. C'est normal si les attach sont bloquants : les goroutines Request bloquent dans le select (reply channel), PAS sur le lock. Elles sont [sync.Mutex.Lock] = elles bloquent sur le lock initial AVANT le select ! Donc elles n'ont même pas atteint le select. Le lock est pris par... quelqu'un continuellement. Qui ? wakeNext est rapide. begin prend le lock brièvement. 
- EN FAIT : 116 callers bloqués sur le lock + stop bloqué sur le lock + 4 attach bloqués sur le lock. Le détenteur probable : une goroutine... qui attend un canal tout en tenant le lock ? Chercher : begin prend le lock de 100 à 169 (avec la section m.running etc.) — si une goroutine attend le select reply EN TENANT LE LOCK... c'est le bug original ! Request : m.mu.Lock, puis m.begin (notify nil en fast path), m.mu.Unlock, select reply. OK unlock avant select.
- Mais cancelSession à l.36-41 : appelé depuis Request dans le case <-ctx.Done() APRÈS m.mu.Unlock. cancelSession prend le lock. OK.
- removeRequest : le waiter prend m.mu.Lock après removeRequest qr... 
- Je cherche le vrai détenteur : dans un dump SIGQUIT, si un goroutine détient le lock et attend un canal, il apparaîtra comme [chan receive] avec Lock dans sa pile AVANT le chan. Cherchons les goroutines [chan receive] dans launch.

## Diagnostic suite (2e passe)

Toutes les goroutines Request du stress sont [sync.Mutex.Lock] à l.27 ou [select] à l.49. Les 4 goroutines func1 sont [sync.Mutex.Lock] à l.150 (le lock final du notify). Aucun détenteur visible : toutes attendent le même mutex et AUCUNE goroutine ne le détient dans le dump.

Réinterprétation : le dump SIGQUIT arrête le monde au moment précis où la goroutine "détentrice" vient de relâcher ou était en cours de course. Mais le système est figé depuis 8s d'après le test (Stop n'est jamais revenu). Avec 8s d'inactivité totale, le détenteur devrait apparaître.

AUTRE PISTE : les 104 goroutines [sync.Mutex.Lock] — mais 116 goroutines func2 total. Donc 116-104=12 sont [select] à l.49 = elles ont le lock, l'ont relâché, attendent reply. OK. Les 104 n'ont jamais pris le lock. Si personne ne le détient et que 104 attendent, le runtime devrait en réveiller une... SAUF si elles attendent le mutex du record ? m.mu à l.27 est bien le mutex Manager. Hmm, m.record (audit) prend aussi le lock ? Vérifier record : si record prend m.mu, une boucle ? Non.

CONCLUSION ALTERNATIVE probable : la goroutine qui détient le lock est une goroutine mergeContexts qui... non, elle ne prend pas le lock.

Vérifier stopOne : elle prend m.mu.Lock à la fin. Si stopOne attend <−w.done tout en tenant le lock ? Le w.done est le canal attached — stopOne attend sur le canal COPIÉ hors du lock dans Stop (mon fix précédent). Vérifier Stop.

## Cause racine identifiée (hypothèse confirmée par analyse logique)

Dans le stress : N=120 goroutines, chacune bloque sur m.mu.Lock à l.150 (la fin de l'attach, après que l'attach a échoué avec m.ctx cancelé OU attend la gate). Scénario réel :

1. func1 finit l'attach (err via ctx cancel) → prend m.mu.Lock → failLocked (delete m.running etc.) → notify → m.mu.Unlock → m.wakeNext() (qui re-prend le lock).
2. 4 goroutines func1 bloquées à l.150 dans le dump : elles attendent le lock tenu par... une autre func1 en train de faire wakeNext ? wakeNext prend le lock 200ms max puis le libère.

MAIS le test échoue : Stop bloqué (l.255 = m.mu.Lock dans Stop). Stop attend le lock tenu par... au moment du dump, aucune goroutine ne détient le lock et toutes attendent → le runtime est figé → le dump capture un instantané où le détenteur était en train d'exécuter du code non-bloquant... non.

VRAIE CAUSE (la plus probable, classique) : **starvation par contention**. Tous les goroutines (func2 callers, func1 attaches, stopOne) battent le même mutex en boucle : les func1 font lock→failLocked→unlock→wakeNext→lock→unlock en continu (fail + ré-admission), les func2 lock→check→unlock, etc. Stop n'arrive jamais à prendre le lock car il est constamment contesté par 120+ goroutines. Le test timeout à 8s : c'est de la contention, pas un deadlock.

Fix : découpler la fin d'attach du mutexManager : failLocked+notify+wakeNext doivent être faits SANS tenir le lock pendant la wakeNext (ré-acquisition), et réduire la pression : par ex. wakeNext après unlock. C'est déjà le cas (unlock puis wakeNext). Mais la ré-admission : wakeNext promeut un waiter → le waiter begin (lock→begin→unlock) crée un nouvel attach qui refait le cycle.

EN RÉALITÉ : le vrai problème est que m.record prend aussi m.mu ? Vérifier record. Et cancelSession/record appelés depuis Request/Stop ajoutent du lock contention.

Fix pragmatique et sûr : rendre le lock plus court et séparer les canaux — OU mieux : accepter que la contention existe en prod mais le test stress est trop agressif. Le vrai test de preuve : réduire la contention en (a) ne pas appeler wakeNext depuis le chemin d'échec quand la queue est vidée, (b) batcher le nettoyage dans stopOne sans ré-acquisition multiple, (c) limiter la pression du test.

DÉCISION : le plus solide = faire que stopOne attende w.done (canal attach fini, qui ferme même après err) et que le mutex Manager ne soit jamais détenu longtemps. Le dump montre que le lock est LIBRE au moment du dump, mais Stop n'obtient jamais le lock pendant 8s de contention → la goroutine Stop n'est pas planifiée à temps. Go n'offre pas de garantie d'équité du mutex. Fix : limiter N dans le test, OU ajouter lock spinning... NON : fix du PRODUCER de contention : dans Request, le check initial du lock ne doit pas bloquer si la session est déjà en queue avec même profil (ré-admission spam). Et dans begin/failLocked : éviter le cycle lock-unlock-lock de wakeNext.

Fix retenu (minimal, propre) :
1. Dans le chemin de fin d'attach : déléguer wakeNext à une goroutine unique (un canal drain "promotion" consommé par une seule goroutine). Élimine 120 goroutines qui battent le lock.
2. Dans stopOne : une seule acquisition (déjà le cas).
3. Test stress : réduire N à 40, durée bornée.

Le point 1 est le vrai gain de robustesse : wakeNext appelé par N goroutines concurrentes = N acquisitions supplémentaires = contention. Avec un canal de promotion consommé par une goroutine dédiée, wakeNext est appelé UNE fois au lieu de N fois.

## Debug stress 3e passe (après promote-channel)

Le fix promote ne change rien : 106 goroutines [sync.Mutex.Lock] (func2 callers bloquées à launch.go:27), 17 [select] (func2 avancées à l.49), 4 func1 [sync.Mutex.Lock] à l.150, Stop bloqué à l.255. Toujours personne ne détient le mutex dans le dump.

CONCLUSION DÉFINITIVE : ce n'est PAS de la contention classique. Si personne ne détient le lock et que 106 goroutines l'attendent en permanence (8s+), c'est que l'une des goroutines qui prend le lock le garde pour toujours — mais le dump montrerait son état avec Lock dans la pile avant le blocage. Cherche une goroutine avec pile : Lock + chan receive (prend lock, attend canal). Les 2 [chan receive] sont candidates. Vérifier leurs piles.

Piste la plus probable restante : une goroutine mergeContexts ou begin bloquée en tenant le lock... mergeContexts ne prend pas le lock. begin prend le lock en entrant, libère avant de lancer le goroutine attach (defer unlock). Le lock de begin couvre l'ajout à running + création du canal attached + goroutine lancé. OK.

NOUVELLE PISTE : Request l.27-49 : après le lock initial, le fast path : si running<limit, appelle m.begin (qui prend le lock INTERNE ?) — begin est appelé AVEC m.mu.Lock DÉJÀ TENU (l.100: m.mu.Lock() puis l.108 m.begin(...)) ! Et begin commence par m.mu.Lock() — DOUBLE LOCK sur sync.Mutex = DEADLOCK INSTANTANÉ ! sync.Mutex n'est pas réentrant. Toute goroutine qui atteint le fast path en premier prend m.mu (l.27) puis appelle begin qui re-prend m.mu (l.108) → deadlock immédiat.

ATTENDS : comment les autres tests passent-ils alors ? Parce que Request est appelé avec mu NON pris normalement... je vérifie Request : l.27 prend m.mu.Lock(), l.28+ fait les checks, l.108 appelle m.begin()... NON : l.108 déverrouille d'abord ? Vérifier la pile Request exacte : l.27 lock puis l.49 select. Les checks entre 27 et 49 doivent faire unlock avant begin. Vérifier le code Request réel.

## Debug 4e passe — hypothèses restantes

1. Les func2 fast path s'ajoutent à running puis attendent reply. func1 en cours (attach timeout 45s via mergeContexts avec m.ctx cancelé) — les func1 devraient finir en <1s après m.cancel(), puis failLocked → running décrémenté → les func2 du select devraient recevoir qr.result/... non, les func2 fast path attendent reply (jamais écrit si failLocked écrit reply? failLocked ne touche pas reply; le goroutine func1 écrit notify = reply pour le fast path à l.166-167). OK failLocked modifie st.State mais notify envoie final (final=sess err). Reply reçu → func2 retourne. Donc après m.cancel(), toutes les attaches finissent en ~StartTimeout? Non : mergeContexts → ctx cancel immédiat → attach retourne → func1 failLocked → notify reply → func2 reçoit → fin. En <1s tout devrait finir.

MAIS les func1 bloquées à l.150 après 3s = elles n'ont même pas fini leur attach ? L'attach = blockingLauncher.Attach(ctx, sess) attend gate. Gate closé par releaseAll APRÈS le drain du test. m.cancel() ne ferme PAS la gate ! Donc l'attach ne retourne JAMAIS même avec ctx cancelé. C'EST ÇA. Le blockingLauncher ne surveille pas ctx ! L'attach bloque sur gate même si ctx est fermé.

Vérifier blockingLauncher : si attach ne respecte pas la convention ctx, c'est un bug du test HARNESS, pas du code. Je dois corriger le blockingLauncher pour observer ctx.Done (retourner ctx.Err() quand le ctx ferme, en plus de la gate).

## Debug 5e passe — blockingLauncher OK

blockingLauncher observe bien ctx.Done (l.55-59). Donc l'attach avec mergeContexts retourne dès m.cancel(). Les func1 terminent → failLocked → unlock → promote → attach.Done().

ALORS POURQUOI LE BLOCAGE ? Les func1 bloquées à l.150 dans le dump = elles attendent le lock pour failLocked. Mais elles finissent par le prendre. Le lock est libéré par tout détenteur.

Dernière hypothèse forte : les func2 fast path dans le select (l.49-61) — reply jamais écrit si func1 échoue avec err != nil : failLocked fait m.record qui prend le lock — record OK. notify(reply) envoie le final. Reply reçu → func2 retourne. OK.

MAIS attendez : failLocked ne touche PAS reply (le champ notify dans begin). Le goroutine func1 envoie `final` via m.notifyNotifyLocked(notify, final) à l.167. notify est le reply du fast path. Donc reply écrit. OK.

VRAIE RÉVÉLATION : le test Stress a 120 goroutines func2 + releaseAll après... le test fait releaseAll() AVANT Stop ? Vérifier TestConcurrentStress. Si releaseAll est appelé, les attaches non-échouées (failOnce=false) réussissent → func1 notify → les func2 du select reçoivent → fin. Puis releaseAll de nouveau, puis Stop. Hmm, mais les func2 fast path : elles sont dans running avec leur func1 en cours. releaseAll débloque toutes les func1 → tout finit. Puis Stop. Devrait passer.

Je relis TestConcurrentStress directement pour vérifier l'ordre.

## Debug 6e passe — pprof mutex

Test Stress n'appelle PAS releaseAll : les attaches se terminent uniquement via ctx (10ms staggered) → func1 failLocked/notify → func2 finissent. Expected: tout fini en <1s, puis Stop trivial.

Le test échoue à 8s → stop bloqué à lock. Tracer avec GODEBUG=mutexprofile ou tracer les lock counts.

## Debug 7e passe — mutex profile décisif

Le profil mutex montre ZÉRO contention sur m.mu (seulement 430µs cumulés, essentiellement fmt). Donc le lock n'est PAS le problème : Stop ne bloque PAS sur le lock pendant 8s — le dump SIGQUIT montrait une image figée instantanée (le lock libre à cet instant). Stop bloque SUR UN CANAL : soit <−w.done (stopOne), soit m.attach.Wait() (join), soit le wg des stopOne.

Relecture Stop : works créé sous lock, puis goroutines stopOne qui attendent <−w.done. w.done = m.attached[id] capturé sous lock. Le canal done est closé par `defer close(done)` dans func1. Func1 termine (ctx cancel) → close(done). StopOne reçoit. Puis stopOne lock → delete → unlock → record.

MAIS : stopOne attend `select { case <-ctx.Done(): case <-w.done: }` — le ctx de Stop est stopCtx 10s timeout. ET les goroutines stopOne sont lancées AVANT le cancel ? Non : Stop commence par m.cancel(), puis prend le lock, crée works, puis lance les goroutines. Donc m.ctx déjà cancelé. Les func1 observent le ctx cancel → terminent → close(done). stopOne reçoit done. OK.

ATTENDS : dans le dump, aucune goroutine stopOne n'apparaît ! Stop bloqué à l.255 (m.mu.Lock) dans le dump après 3s. Mais le profil mutex dit zéro contention sur m.mu... contradiction apparente. Le dump montre la pile de Stop comme `m.mu.Lock` à l.255 — c'est l'endroit où elle bloque. Mais si pas de contention... elle bloque parce qu'UNE AUTRE goroutine détient le lock AU MOMENT DU DUMP. Laquelle ? Le dump ne montre aucun détenteur (toutes en Lock ou select). Cela veut dire que le détenteur a libéré JUSTE AVANT le dump. Normal.

HYPOTHÈSE FINALE : Stop est dans la file d'attente du mutex, la file progresse lentement parce que m.record (audit) prend aussi le lock ET il y a une boucle : record appelle log.Printf(fmt.Sprintf...) — le profil montre fmt bloqué sur le lock... NON.

REGARDONS LE DUMP PLUS PRÉCISÉMENT : 106 goroutines func2 [sync.Mutex.Lock] à l.27 — elles attendent le lock. 17 dans le select à l.49 — elles ont fini le fast path et attendent reply. Mais reply est écrit par func1 à la fin de l'attach. Les func1 avec failOnce=false (la majorité, 1/20e échoue) : leur attach observe ctx.Done → retourne → func1 lock → notify → unlock → promote. Donc les func1 finissent, notify écrit reply, func2 sort du select, unlock, retour. Progression. Mais le dump les montre [sync.Mutex.Lock] à l.27 — elles n'ont PAS encore pris le lock pour leur premier passage. Elles attendent le lock tenu par... des func1 ? Les func1 prennent le lock à la fin. Mais les func1 aussi se terminent.

OÙ EST LE VRAI PROBLÈME : func2 fast path — reply jamais écrit si la func1 termine avec err≠nil ET notify==nil ? Le fast path passe notify=reply (l.47). Func1 utilise notify si non-nil (l.166). Reply écrit. OK.

DÉCIDÉ : je vais exécuter le test avec -trace 1s et analyser les goroutines bloquées à la fin, ou plus simple : un petit test dédié minimal qui reproduit : 120 goroutines demandent en parallèle avec cancellation ; je compte les Request résolus après 5s sans releaseAll. Le Manager devrait résoudre toutes les Request via ctx cancel (l.50-55 : ctx.Done → cancelSession → return ErrCancelled). 

ATTENTION — CANCELSESSION ! Dans le fast path du test, quand ctx.Done se ferme, Request appelle m.cancelSession(sess) puis retourne. cancelSession appelle removeRequest... qui supprime de la queue — mais le session est dans running (fast path admis) ! cancelSession doit aussi le retirer de running. Vérifier cancelSession : si elle ne retire pas de running, le slot est perdu → la limite se remplit de zombies → les autres Request tombent en queue avec deadline 30s → attendent 30s → timeout test.

Voilà probablement le vrai bug : cancelSession ne nettoie pas le running set dans le fast path !

## Debug 8e passe — BUG RÉEL IDENTIFIÉ

cancelSession nettoie bien running/byProfile/attached. MAIS la func1 attach continue en arrière-plan ! Quand l'attach finit (via ctx merge), func1 fait :
- m.mu.Lock()
- check m.running[sess.ID] → PAS PRÉSENT (supprimé par cancelSession) → branche else : final = sess, StateInterrupted, "attach interrupted"
- notify(reply) — écrit sur le reply du fast path
- m.mu.Unlock()
- promote

Le Request dans le fast path avait DÉJÀ retourné (ctx.Done → cancelSession → ErrCancelled). Le reply arrive après : personne ne le lit. OK pas de fuite (buffer 1).

Donc le fast path avec ctx cancelé : Request retourne vite. Les func2 dans le select attendent reply — reply arrive à la fin de l'attach (45s timeout ou ctx cancel 10ms). Tout devrait finir.

ALORS POURQUOI BLOQUÉ 8s ? Les func2 avec ctx annulé (10ms) retournent via la branche ctx.Done (l.50-55). Les func2 SANS annulation : leur ctx n'est jamais fermé ! Leur func1 attend la gate JAMAIS ouverte (pas de releaseAll dans ce test) → l'attach dure 45s (StartTimeout) → func1 failLocked (err context deadline) → notify → func2 reçoit → retourne. Donc les func2 sans annulation retournent à t=45s. Le test attend Stop qui bloque tant que les func1 ne sont pas finies... MAIS Stop fait m.cancel() → mergeContexts ferme tous les ctx func1 → les func1 finissent IMMÉDIATEMENT (avant les 45s). Func1 finit → func2 du select reçoit → retourne. Stop : stopOne attend done (closé par func1) → stopOne finit → attach.Wait() (func1 Done) → join fini → Stop retourne. Tout devrait finir en <100ms après m.cancel().

Sauf que : m.cancel() ANNULE AUSSI LES CTX DES FUNC2 qui attendent dans le select du fast path (leur ctx parent est passé à begin mais le Request garde son propre ctx)... le select fast path attend <-ctx.Done (le ctx func2) — pas m.ctx. OK.

HMMMM. Dernier angle : le test bloque parce que Stop prend le lock À L.255 — et les func1 en fin d'attach prennent le lock à l.150 aussi. Pas de deadlock mais... la func1 qui prend le lock fait failLocked (record prend... non, record sans lock) → notify → unlock → promote. Le promote est un chan buffer 1 avec default : si plein, perdu. La goroutine promote consomme 1 et appelle wakeNext. Si 120 func1 finissent en même temps, 1 seul promote passe. wakeNext : check running>=limit → return. OK.

Mais ATTENDS : les func2 dans le select (17 du dump) — reply écrit par func1 à l.167 (notifyNotifyLocked). notifyNotifyLocked : select notify<-sess default. Si le func2 a déjà quitté le select (via ctx.Done), le reply tombe dans le canal buffer 1. OK.

TOUT SEMBLE FINIR. Pourquoi ça ne finit pas ? Je vais instrumenter : un print dans func1 à chaque étape (attach fin, lock, unlock, promote) avec compteur atomique, pour voir ce qui n'arrive jamais.

## État complet debug stress (sauvegarde)

Contexte : audit T08 a demandé (1) TestConcurrentStress, (2) preuve de terminaison sans fuite. Le stress test bloque à 8s sur Stop.

Code actuel (après fix promote-channel) :
- launch.go:27-62 Request fast path (lock→check→begin→unlock→select ctx.Done/reply)
- launch.go:64-107 Request queue path (lock→queue→unlock→select)
- launch.go:114-179 begin (lock tenu par caller, running+byProfile+attached, goroutine func1 avec defer close(done), defer attach.Done(), ctx=mergeContexts(ctx,m.ctx), WithTimeout 45s si pas de deadline, Attach, puis lock→failLocked/notify→unlock→promote canal)
- manager.go:142 promote chan, goroutine dédiée consommant → wakeNext
- launch.go:211-231 wakeNext (lock, check, begin waiter)
- launch.go:314-346 stopOne (select ctx/w.done, lock, delete, record)
- launch.go:249-295 Stop (m.cancel, lock→works+drain queue, goroutines stopOne, wg, join attach.Wait avec deadline StartTimeout)
- launch.go:423-434 cancelSession (lock, StateInterrupted, delete running/byProfile/attached)
- blockingLauncher.Attach (failOnce, select ctx.Done/ gate) — respecte ctx

TestConcurrentStress (l.476-545) : 120 goroutines, 24 profils, 24 launchers distincts, 1/6 annulés à 10ms, PAS de releaseAll, Stop avec timeout 10s, test échoue si Stop pas fini en 8s.

Dump SIGQUIT : 106 func2 [Mutex.Lock l.27], 17 func2 [select l.49], 4 func1 [Mutex.Lock l.150], Stop [Mutex.Lock l.255]. mutex profile : quasi zéro contention sur m.mu.

Piste restante la plus forte : le ctx du fast path func2 = ctx parent passé par le test. begin fait ctx=mergeContexts(ctx,m.ctx). Merge :
```go
func mergeContexts(a, b context.Context) context.Context {
    merged, cancel := context.WithCancel(context.Background())
    go func() {
        select { case <-a.Done(): case <-b.Done(): }
        cancel()
    }()
    return merged
}
```
LEAK DE GOROUTINE : chaque mergeContexts lance une goroutine qui attend a.Done ou b.Done. Si ni a ni b ne se ferme jamais (func2 sans annulation, func1 avec gate) → goroutine fuite jusqu'à ce que attach timeout 45s. Pas le blocage mais un leak.

VRAI BLOCAGE PROBABLE : dans func1, après Attach retour (via ctx cancel 10ms ou 45s), le goroutine func1 fait lock→failLocked→unlock→promote. MAIS le lock est tenu par... La func1 bloque à l.150 après 3s dans le dump — elle attend le lock. Qui le tient ? Au dump, personne. Mais la func1 attend DEPUIS 3s → il y a bien un holder quelque part... OU la func1 n'est même pas encore arrivée à l.150 : le dump montre sa pile à l.150 parce que le fichier a été recompilé ? NON, le dump montre la pile réelle.

DÉCISION : instrumenter avec des compteurs atomiques (attachFinis, locksPris, promotes, wakeNexts, stopOnes, requestsRésolus) et logguer à la fin du test. Ça tranchera.

## État complet (avant instrumentation) — relire en priorité

Fichier à modifier : /home/ubuntu/ForgeLocal/internal/launch/launch.go et manager.go.

Le Manager est déclaré dans manager.go (l.126-145) avec ctx/cancel aux l.140-141.

Instrumentation prévue (atomic.Int64, import "sync/atomic") : CAttachDone, CNotifySent, CPromoteSign, CWakeNextRun, CStopOneDone, CReqResolved, CReqQueuePath, CByProfileHit, CLockTaken. Ajouter dans le struct Manager (manager.go après cancel), incrémenter : CLockTaken à chaque m.mu.Lock (Request l.27, begin caller lock l.100, wakeNext l.212, cancelSession l.424, stopOne l.324, removeRequest l.235, Status l.193), CByProfileHit à l.28-35 (déjà édité via edit 2 — à revérifier), CReqQueuePath après queue append, CAttachDone dans func1 defer, CNotifySent à notifyNotifyLocked success, CPromoteSign dans func1 après promote send, CWakeNextRun début wakeNext, CStopOneDone fin stopOne, CReqResolved avant chaque return de Request (5 sites : l.35, l.39, l.53, l.55, l.58-61, l.98, l.100, l.103-105).

Test : compter les compteurs dans TestConcurrentStress (m.C...) et t.Logf avant le Fatal.

Objectif : déterminer qui bloque vraiment. Le dump SIGQUIT + profil mutex excluent la contention mutex.

Note : edit 1 (commentaire import) déjà appliqué. Edit 2 et 3 rejetés (edit 2 : texte introuvable car le Manager est dans manager.go pas launch.go ; edit 3 : appliquer sur launch.go Request).

## Observation clé (16 août, ~après instrumentation)

Le test TestConcurrentStress PASSE maintenant seul (0.103s). Aucun changement
de logique n'a été fait depuis la dernière tentative ratée — seul ajout :
compteurs atomiques + unlock du mutex dans la branche ErrQueueFull (qui
retournait avec le lock TENU = bug réel, mais ce bug ne cause pas le stress
deadlock car la file n'était pas pleine ; en revanche il causait un deadlock
potentiel si la file était pleine).

Hypothèse : le fix précédent (promote goroutine dédiée) avait réduit la
pression ; le vrai blocage intermittent du join était causé par la race
promote : wakeNext sous lock, les attaches finissant en rafale → contention
massive + goroutine promote buffer 1 → perte de signaux de promotion →
les waiters en queue non promus, les attaches finissent, m.attach.Wait()
attend les N attach (qui terminent via m.ctx cancel) — donc le join devrait
passer. Sauf si des waiters restent dans m.queue avec ctx expirés mais la
goroutine Request de ce waiter bloque sur removeRequest puis le second lock.

Le blocage stress observé : Stop join 8s échouait. Le join attend
m.attach.Wait() = tous les attach. Les attach observent mergeContexts(ctx,
m.ctx) : m.ctx est cancel → ils terminent. CAttachDone devrait monter à
le nombre d'attach lancés. Si ça passe maintenant, c'est parce que la
correction `m.mu.Unlock()` dans la branche ErrQueueFull (absente avant)
était le vrai fix — sans elle, quand la file et la limite sont pleines,
Request retourne ErrQueueFull AVEC LE LOCK TENU → toutes les goroutines
suivantes meurent en lockSlow à jamais.

Le stress : 120 goroutines, 24 profils, limite 4, file 32. À la fin :
running=4, queue=32, tout le reste = refusés par la file pleine — qui
rentraient dans la branche ErrQueueFull avec lock tenu → DEADLOCK MASSIF.

✅ BUG CONFIRMÉ ET CORRIGÉ : ErrQueueFull retenait m.mu.
Le stress a toujours échoué en queue-full.

## Race identifiée (après fix ErrQueueFull)

La race : begin.func1 lance `launcher.Attach(ctx, sess)` où sess est une
COPIE locale du struct Session — PAS une race. MAIS le pointeur :
m.running[sess.ID] = &sess (pointeur vers la variable locale de la
goroutine Request !). Wait : begin est appelé avec sess=Session(valeur).
`m.running[sess.ID] = &sess` stocke l'adresse du paramètre de begin —
chaque goroutine begin a son propre frame.

cancelSession écrit st.State/StoppedAt/Err via m.running[sess.ID] (sous lock).
begin.func1 lit launcher.Attach(ctx, sess) — le read à launch.go:159 +0x3d3
est dans le frame de begin.func1. offset 0x3d3 ~ `err := launcher.Attach`.
Le compiler instrumente l'accès à sess (le struct, passé par valeur mais
l'instrumentation couvre la copie). La copie elle-même est locale → pas de
race réelle sur les données, MAIS le paramètre sess de begin est passé
par valeur — la copie est faite à l'appel.

ATTENDS : la race réelle : le paramètre `sess Session` de begin est une
copie. Mais m.running[sess.ID] = &sess stocke l'adresse du paramètre
copie. Deux goroutines différentes n'accèdent pas au même &sess.

La race détectée : cancelSession (goroutine 12 = Request ctx.Done →
m.cancelSession(sess)) écrit dans `st` = m.running[sess.ID]. begin.func1
(goroutine 59) lit... launch.go:159 +0x3d3. Ligne 159 = `err :=
launcher.Attach(ctx, sess)`. L'instrumentation race considère la lecture
du CHAMP State/StoppedAt/Err de sess dans l'appel Attach ? Non, Attach ne
lit que ctx et sess (copie) — mais le compiler trace la copie du paramètre
au moment de l'appel : c'est une écriture (copy into callee frame), pas
une lecture. Hmm.

Vraie lecture dans func1 à cet offset : peut-être le `ctx` capture ou le
read de `m.opt.StartTimeout` ? La ligne exacte est launch.go:159. Le `+0x3d3`
est l'offset dans le symbol. La lecture peut être le champ `State` de sess
lu par... aucun. Attendez : le compiler pourrait instrumenter le read du
champ `Err` dans `redacted(err.Error())` ? Non, c'est err.

Je pense que c'est le `var final Session; final = *st` (sous lock) ou
`final = sess; final.State = StateError` (hors lock ? non, sous lock aussi).

Le read/write concernent les adresses 0xac0b0/0f0/108 (3 champs consécutifs
de Session : State/StoppedAt/Err) — le même pointeur de Session partagé.
La goroutine 59 lit (begin.func1 à 159 = la ligne `err := launcher.Attach`).

Conclusion : la race est un FALSE POSITIVE sur la copie de paramètre OU un
vrai partage. Le paramètre de begin est une copie locale → pas de race.
Mais le compiler détecte un read sur ces adresses par func1... parce que
`sess` dans le scope func1 est... le paramètre de func1 ? func1 capture
`sess` de begin (par valeur). La lecture = le read de `sess` lors du
launcher.Attach(ctx, sess) au moment de l'appel (copy argument read).
Le write = cancelSession écrit dans `st` = m.running[sess.ID] qui est
&sess de begin (la copie) — c'est LA MÊME ADRESSE ! Le pointeur
m.running[sess.ID] pointe vers le paramètre copie de begin. cancelSession
écrit dans ce pointeur → modifie le paramètre copie → func1 lit ensuite
(ou avant) ce même paramètre copie → RACE RÉELLE : le paramètre copie est
partagé entre cancelSession (via m.running pointeur) et func1 (lecture
locale).

FIX : m.running doit stocker un pointeur vers une nouvelle variable heap
(ou &sess déjà heap car échappement), MAIS la vraie correction : ne pas
exposer le pointeur au paramètre. Solution simple : begin convertit sess
en pointeur heap ET func1 lit via ce pointeur (ou copie) — plus simple :
faire func1 copier sess localement avant tout, et cancelSession copier
avant d'écrire. OU : begin crée un pointeur dédié : `sp := &Session{...};
m.running[id] = sp; sp2 := *sp` dans func1. Le plus propre :
- `sp := &sess` dans begin → m.running[id] = sp
- func1 : `sessLocal := *sp` (copie locale) au début ; tout le reste
  utilise sessLocal
- cancelSession : écrire sur la copie, pas sur le pointeur partagé :
  lire st, copier, delete, puis appliquer les changements sur la copie et
  la republier via un canal... ou simplement écrire directement sur st
  puisque func1 ne lit plus &sess (il lit sessLocal).

## FIX FINAL race (confirmé)

Tests -race : OK (3.457s, toute la suite). Les deux bugs corrigés :
1. ErrQueueFull retenait m.mu (deadlock massif sous charge queue-full)
2. m.running stockait &sess (paramètre de begin) partagé avec begin.func1
   → cancelSession écrivait dedans pendant que func1 le lisait → race réelle
   (fix : pointeur heap dédié)

Reste à faire :
- go vet
- gitleaks sur le delta (4 fichiers)
- git status
- archive de preuves T08-R2 avec manifeste
- rapport final à l'utilisateur (statut 🟢 COMPLET si tout passe)

# Gel contrôlé post-T08 (16/08/2026)

- [x] T08 clôturé : `T08_APPROVED_VERIFIABLE_LOCAL` (CDC v3.9.7 commit `f3a19df`, rapport final commit `903f6bd`, code commit `99a22f5`).
- [x] Archive figée `t08-r2-final.zip` SHA-256 `4918ac9876545904c822ff72fb3dfcc4f8b12f6fb2214452e308a39b4c0719bb` (deux copies identiques vérifiées par la relectrice).
- [x] Document de suivi non normatif : `docs/T08-FOLLOW-UP-WATCH-STATE.md` (commit `0bcc94f`), sans nouveau statut/exigence/certificat.
- [x] Mode gel contrôlé : aucune modification code T08, aucune nouvelle archive, aucun T09 (avant autorisation), aucun runtime/Camoufox/proxy/backup/UI, pas de changement CDC hors préservation.
- [x] T07 : décision d'audit `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION` reçue ; attestation `ATT-T07R-CAMOFLOX-001`, revues `REV-T07R-2026-001`/`REV-T07R-2026-002`, `REVIEW_ACCEPTED_FOR_T07_DECISION` ; statut intermédiaire T07-R `ATTESTATION_REDACTED_PENDING_SIGNATURE_AND_EXTERNAL_REFERENCES` (références obligations tierces/notices et informations de signature finale en attente dans l'Issue `#1`).
- [ ] T09 exigeait (conditions levées) : (1) T07 débloqué par revue indépendante avec registre mis à jour, et (2) autorisation explicite d'ouverture T09. T09 clôturé `T09_APPROVED_VERIFIABLE_LOCAL`.
- Statuts : PUBLIC_RELEASE_BLOCKED, SCAN_BLOCKED_UNKNOWN (réserves pré-T09 maintenues), pilote suspendu, cinq gates publics actifs.

# T09 — Profile Writes — CLÔTURÉ

- [x] Clôture validée par le valideur : `T09_APPROVED_VERIFIABLE_LOCAL` (16/08/2026). Archive `t09-r1.zip` SHA-256 `1d032a53c516e7e61fad4ca3523a6d685d1f7b1a8cf605e79626359bd7108369`, manifeste 13/13, 74 tests Core sélectionnés (63 + 11 sous-tests) 0 FAIL 0 DATA RACE, E2E Playwright 2/2 (7 assertions nommées), Gitleaks delta T09 + delta UI `[]`, build tsc + prod propres.
- [x] Commits : `e66ba0e`, `fe1f91b`, `3dad8db`, `dc32d96` (HEAD) sur `forgelocal-product-v0.3` de `boucheriechefimane-cmd/IPcache`. Dashboard : checkpoint publié `bd96760e` → `forgelocal-d-c8wqrxmp.manus.space` (auto-publish).
- [x] Périmètre exact validé : création, modification, archivage, réouverture autorisée, tags, validation serveur, audit redacted, `correlation_id`, loopback 403 `LOOPBACK_REQUIRED`, client React mémoire seule connecté au Core. Rapports finaux 16 champs livrés (`T09-R1-FINAL-REPORT.md`).
- [x] Code livré : migration `0005_t09_profile_writes.sql`, `internal/api/profiles_write.go`, `audit.go`, `correlation.go`, `readonly_session.go` (requireLoopbackMiddleware), `internal/profile/store.go` (mutex par profil, cycle de vie), `errors.go` ; dashboard `coreWrite.ts`, `LocalCoreConnection.tsx`, `Home.tsx`, `tests/profile-writes-t09.spec.ts`.
- [x] T10 (Proxys) autorisé à préparer (cadrage seulement) mais interdit à démarrer dans le même lot : aucun code T10 produit. Aucune validation intermédiaire demandée pendant T10 ; rapport unique 16 champs à la fin.

# T10 — Proxies — OUVERT (autorisation T10_AUTHORIZED_TO_START, 16/08/2026)

- Décision : `T10_AUTHORIZED_TO_START` — cadrage validé par le valideur (AC-PROXY-01 à AC-PROXY-08 validés comme critères) ; aucun certificat/attestation/approbation de release requis pour démarrer ; rapport unique 16 champs à la fin ; T11 interdit dans le même lot.
- [x] T10-F1 : cadrage documentaire non normatif du lot T10. Cadrage produit : `docs/T10_PROXIES_FRAMING.md` (T10-FRAMING-20260816-001).
- [x] T10-F2 : autorisation explicite d'ouverture T10 reçue du valideur (validation du cadrage + instruction de démarrage) ; code T10 autorisé.
- Périmètre : référentiel proxy local Core Go (CRUD, validation http/socks5, hôte, port borné, région, secret_ref synthétique, affectation profil↔proxy, désaffectation, listing redacted, audit, correlation_id, loopback, erreurs machine-readable) ; dashboard client mémoire seule sans écriture SQLite directe. Interdictions : aucun proxy réseau réel, navigateur, runtime, Camoufox, intégration Decodo/fournisseur, backup/restore UI, release. `PUBLIC_RELEASE_BLOCKED` inchangé.
- [x] T10-1 : pas de migration SQLite (persisté en fichier `internal/proxies/proxystore`, découplé du catalogue T08) ; store proxy avec mutex par proxy, validation serveur.
- [x] T10-2 : API `GET/POST /api/proxies`, `GET/PUT/DELETE /api/proxies/{id}`, `POST/DELETE /api/proxies/{id}/assign` (404 `PROFILE_NOT_FOUND` sur profil fantôme), `GET /api/v1/readonly/proxies` redacted, `proxy_id` redacted exposé sur `getProfile`.
- [x] T10-3 : tests Core : AC-PROXY-01/02/03/04/05/06 (validation, refus, affectation, jamais en clair, loopback 403, concurrence), persisté en fichier (aucune migration SQLite), secrets synthétiques uniquement.
- [x] T10-4 : dashboard : client mémoire seule coreWrite étendu (routes proxy), UI référentiel + affectation, build prod + tsc + Playwright E2E T10.
- [x] T10-5 : contrôles finaux : `go test -race` (0 DATA RACE), `go vet`, `go build`, Gitleaks delta `[]`, `git diff --check`, RC inchangés.
- [x] T10-6 : archive de preuves redacted + manifeste portable + SHA-256 + rapport final 16 champs + commit/push.

# T10 — Proxies — CLÔTURÉ

- [x] Clôture validée par le valideur : `T10_APPROVED_VERIFIABLE_LOCAL` (16/08/2026). Archive `t10-r1.zip` SHA-256 `69723d4e55776e3a74408a3f9bb17418bbe94a49d9f4373ac775577edec6806b`, manifeste 17/17, 16 packages sous `-race` 0 FAIL 0 DATA RACE (dont 30 tests T10 nouveaux), E2E Playwright 2/2 (9 assertions nommées), Gitleaks delta T10 + arbre complet `[]`, Gosec 0 finding sur fichiers T10 (16 findings préexistants hors delta), build tsc + prod propres.
- [x] Commits : `8035609`, `38b3891`, `7a21e45`, `b6939d8`, `12a87ea` (HEAD) sur `forgelocal-product-v0.3` de `boucheriechefimane-cmd/IPcache`. Dashboard : checkpoint publié `6707097d` → `forgelocal-d-c8wqrxmp.manus.space` (auto-publish).
- [x] Périmètre exact validé : référentiel proxy local (CRUD, validation http/socks5 + port borné + nom unique), affectation/désaffectation profil↔proxy (refus profil fantôme), `secret_ref` référence seule (jamais de credential en clair), redaction, audit, `correlation_id`, loopback 403 `LOOPBACK_REQUIRED`, client React mémoire seule connecté au Core. Rapport final 16 champs livré (`T10-R1-FINAL-REPORT.md`).
- [x] Code livré : `internal/proxies/{errors,store,store_test}.go`, `internal/api/proxies.go`, `internal/api/proxies_test.go`, `internal/api/readonly.go` (readonlyProxies), `internal/api/router.go` (NewRouterWithProxyRegistry, proxy_id redacted), `cmd/server/main.go` (injection) ; dashboard `coreWrite.ts`, `ProxyRegistry.tsx`, `Home.tsx`, `index.css`, `tests/proxies-t10.spec.ts`.
- [x] Note historique : défaut de traçabilité initial (hash `aa56ed33…` annoncé avant inclusion du rapport dans l'archive) fermé par reconstruction canonique au pattern T09 : rapport livré séparément, hash `69723d4e…` stable.
- [x] Réserves de la validation : coffre `internal/secrets` non qualifié (aucun `secret_ref` réel activé, preuves synthétiques), connectivité proxy réelle/DNS/WebRTC/latence hors périmètre, écriture JSON non atomique et corruption de fichier non encore testées — à traiter si le CDC l'exige avant toute activation réelle.
- [x] T11 (Backups/restauration selon CDC v3.9.7 §3.4/BACK-01) autorisé à préparer (cadrage seulement) mais interdit à démarrer dans le même lot : aucun code T11 produit. Aucune validation intermédiaire demandée pendant T11 ; rapport unique 16 champs à la fin.
- [x] T11-F1 : cadrage documentaire non normatif du lot T11. Cadrage produit : `docs/T11_BACKUPS_FRAMING.md` (T11-FRAMING-20260816-001).
- [x] T11-F2 : autorisation explicite d'ouverture T11 reçue du valideur (`T11_AUTHORIZED_TO_START`, 16/08/2026) ; code T11 autorisé. Précision conservée : le fallback local chiffré a ses propres tests (absence de clé en clair, permissions, corruption, mauvais AAD, nonce non réutilisable, restauration interrompue, reprise fail-closed, absence de fuite logs/SQLite) sans être présenté comme équivalent SystemVault.
- Périmètre : contrat local Backup/Restore Core Go unique (.flbackup FLBK…FLEND, AES-256-GCM/AAD, nonce crypto/rand, publication atomique, fsync, restauration isolée vers nouvel identifiant, validation/quarantaine, journal durable, crash recovery fail-closed) ; dashboard client mémoire seule. Données/références synthétiques uniquement. Interdictions : secret réel, coffre non qualifié, runtime, navigateur, Camoufox, proxy réel, fournisseur, import de masse, release. `PUBLIC_RELEASE_BLOCKED` inchangé.
- [ ] T11-1 : package `internal/backup` : conteneur .flbackup (FLBK/FLEND), AES-256-GCM + AAD + nonce crypto/rand, permissions restrictives, publication atomique + fsync.
- [ ] T11-2 : restauration isolée vers nouvel identifiant + validation complète (format/intégrité/AAD/chemins) + quarantaine des artefacts malformés + jamais d'écrasement.
- [ ] T11-3 : journal durable backup_operations/restore_operations (started→validated/committed/failed), crash recovery INTERRUPTED_BEFORE_COMPLETION, reprise fail-closed.
- [ ] T11-4 : API loopback-only `POST /api/backups`, `GET /api/backups`, `GET /api/backups/{id}`, `POST /api/backups/{id}/restore`, suppression quarantaine ; audit redacted, correlation_id, erreurs machine-readable, jamais de clé/secret en clair.
- [ ] T11-5 : tests positifs + négatifs + corruption/troncature/AAD + chemins hostiles + concurrence + `-race` ; migration SQLite 0007 si nécessaire ; secrets synthétiques uniquement.
- [ ] T11-6 : dashboard : client mémoire seule étendu + UI backups/restore + build prod + tsc + Playwright E2E T11.
- [ ] T11-7 : contrôles finaux : `go test -race` 0 DATA RACE, `go vet`, `go build`, Gitleaks delta `[]`, Gosec delta 0, `git diff --check`, RC inchangés.
- [ ] T11-8 : archive de preuves canonique (rapport séparé du ZIP, SHA256SUMS stable, hash calculé une seule fois après constitution finale) + rapport final 16 champs + commit/push.
- Statuts inchangés : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN` (findings `validation_back01_integration/` préexistants, règle `generic-api-key`), pilote suspendu, cinq gates publics en attente.


# T28–T42 — clôture indépendante 2026-08-24

- [x] T28–T30 : auditer les archives reçues dans une sandbox réelle ; publier uniquement l’audit et les corrections de preuve non destructives ; conserver T28/T29 product-gated et T30 localement vérifiable.
- [x] T31–T38 : exécuter séquentiellement les lots redacted/localement vérifiables, avec tests race, vet, build, scans, bundles, manifestes et extraction fraîche ; publier les branches GitHub vérifiées.
- [x] T39 : documenter le blocage de l’import/export de secrets faute d’autorisation T28/T29 et de qualification native SystemVault ; aucune donnée secrète manipulée.
- [x] T40 : documenter le gate d’intégration sans fusion runtime ni activation de configuration.
- [x] T41 : documenter la readiness de release `BLOCKED` ; une archive cumulative oversized refusée par GitHub a été conservée hors dépôt et remplacée par une preuve delta compacte.
- [x] T42 : produire le registre canonique et la clôture technique avec verdicts stricts ; aucune clôture produit, runtime ou release déclarée.
- [ ] Lever `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` ou `release_authorized=false` : interdit dans cette passation.

# Correction de clôture T28–T42 — 2026-08-24

- [x] Vérifier indépendamment les écarts signalés sur la branche finale : logs baseline absents par lot, sidecars absolus, T30 non réconcilié et métadonnées T42 inexactes.
- [x] Ajouter les reconstructions postérieures explicitement nommées `BASELINE_RECONSTRUCTION_POSTHOC_RAW.log` pour T31–T38, sans les présenter comme des preuves contemporaines.
- [x] Ajouter des sidecars compagnons `*.portable.sha256` relatifs pour les ZIP et bundles T31–T38 et T42, sans modifier les sidecars historiques.
- [x] Réconcilier T30 avec son commit, son URL GitHub et les hashes d’archive ; maintenir `PENDING_REMOTE_EVIDENCE_RECONCILIATION` en l’absence de branche distante canonique.
- [x] Rejouer les contrôles complets depuis un clone neuf et conserver les sorties brutes, les codes de sortie et les timestamps UTC.
- [x] Conserver le signal Gitleaks cumulatif et les constats Gosec historiques ; ne pas transformer `SCAN_BLOCKED_UNKNOWN` en succès.
- [ ] Autoriser une release, un runtime réel, Camoufox, un proxy réel, un cookie réel, SystemVault natif ou une migration utilisateur : interdit.

# T00–T42-PREHUMAN-FINDINGS-FINALIZATION — 2026-08-24

- [x] Extraire les 13 findings GolangCI-Lint nouveaux avec règle, fichier, ligne, message brut, sévérité non renseignée, lot/rattachement, cause, risque, propriétaire et condition de levée.
- [x] Classer chaque finding GolangCI-Lint : 12 sont des lignes de fichiers inchangés entre baseline et HEAD avec différentiel scanner/contexte non réconcilié ; le SA9003 est rattaché à T38 ; aucun finding n’est laissé sous la seule mention « connu ».
- [x] Documenter les 36 findings Staticcheck historiques et les 6 misconfigurations Trivy avec baseline/head, risque, décision et condition de levée.
- [x] Joindre l’inventaire de licences production package par package et son regroupement exact.
- [x] Documenter Playwright/T10 comme `NOT_APPLICABLE_UNDER_CURRENT_GATES` avec commande, CWD, UTC, sortie brute, code de sortie et préconditions absentes ; aucun token ou runtime réel créé.
- [x] Préparer le wrapper append-only `forgelocal-t00-t42-prehuman-final-review-wrapper-v2.zip` contenant le ZIP historique intact et l’addendum complet ; vérifier son manifeste, son extraction, ses hashes et son re-scan Gitleaks.
- [x] Conserver la sortie `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW` pour revue humaine indépendante uniquement.
- [ ] Corriger les exceptions qualité ouvertes après autorisation explicite et périmètre de remédiation ; ne pas masquer par `nolint`, exclusion ou modification de configuration.
- [ ] Lever `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` ou `release_authorized=false` : interdit dans cette passation.

# T00–T42-PREHUMAN-CODE-FIX — 2026-08-24

- [x] Analyser individuellement les 13 findings GolangCI-Lint et distinguer erreur réelle, effet de différentiel et dette historique.
- [x] Corriger les retours `Shutdown`, `Start`, `Write`, `io.Copy` et `Rollback` réellement non traités.
- [x] Remplacer la branche vide SA9003 de T38 par une assertion de non-régression explicite.
- [x] Rejouer les tests ciblés, race, Go global, vet, build, linters, vulnérabilités, SBOM et Dashboard depuis un clone neuf.
- [x] Conserver les logs post-correctif et le mapping exhaustif sous `evidence/PREHUMAN_T00_T42_FINAL_CHECKLIST/POSTFIX_*`.
- [ ] Traiter les findings historiques non ciblés et décider leur remédiation avec les propriétaires concernés.
- [ ] Lever toute gate permanente ou déclarer une release : interdit dans cette passation.

# T00–T42-SELF-VALIDATION-SYNTHETIC-E2E-V4 — 2026-08-24
- [x] Rejouer la baseline depuis un clone neuf et consigner CWD, UTC, espace, versions, status, HEAD et refs.
- [x] Réhydrater uniquement les quatre artefacts LFS critiques et valider sidecars, ZIP, extractions, manifestes et bundles.
- [x] Rejouer Core, Dashboard, sécurité, SBOM et inventaire de licences ; conserver les codes de sortie réels.
- [x] Exécuter l’E2E Playwright synthétique loopback avec token éphémère, stockage vide, rejeu refusé et aucune requête tierce.
- [x] Vérifier le cleanup : aucun processus, token, SQLite ou répertoire temporaire résiduel.
- [x] Préparer le wrapper v4 append-only, son manifeste non auto-référentiel, ses sidecars, le bundle delta et la classification individuelle.
- [ ] Traiter les findings historiques avec leurs propriétaires et une revue humaine indépendante.
- [ ] Lever toute gate permanente ou déclarer une release : interdit.

# Publication V4 — 2026-08-24
- [x] Documenter dans le registre les commits `ad41afd`/`c905f88`, hashes wrapper/bundle, résultats E2E et limites.
- [x] Préparer la branche `audit/t00-t42-self-validation-synthetic-e2e` pour publication.
- [ ] Obtenir la revue humaine indépendante ; aucune gate ne doit être levée avant celle-ci.

# Publication distante finale V4 — 2026-08-24
- [x] Vérifier le HEAD distant 5e174dba6dddc35865f5bd943383d988ea12170c et un clone neuf avec fsck code 0.
- [x] Enregistrer les hashes finaux du wrapper et du bundle delta dans le registre et le CHANGELOG.
- [ ] Obtenir la revue humaine indépendante ; ne lever aucune gate avant cette revue.

# HEAD publié v4 — 2026-08-24
- [x] Enregistrer le HEAD final publié b4a04e4b9b489c22f3a86986c6faa1cbb9bf77c5 après synchronisation des preuves.
- [ ] Obtenir la revue humaine indépendante ; maintenir toutes les gates.

# T00–T42-V6-FINDINGS-REMEDIATION — 2026-08-25
- [x] Créer un clone neuf de la branche V4 et produire `V6_BASELINE_DISCOVERY_RAW.log` avant correction.
- [x] Individualiser les 348 findings Gitleaks redacted sur 58 arbres et les 6 findings du checkout frais.
- [x] Corriger `ineffassign` dans `cmd/server/cli_runtime.go` et `SA1019` ciblé dans `internal/api/sessions.go`, avec tests normaux/race et linter post-correction.
- [x] Corriger les deux violations Axe, ajouter le test Axe au scénario E2E et obtenir 0 violation après correction.
- [x] Mettre à jour `golang.org/x/mod` au minimum corrigé `v0.40.0`, puis exécuter tidy, verify, race, vet, build, govulncheck et Grype SBOM.
- [x] Trier individuellement les 18 findings Semgrep et confirmer l’usage réel de `crypto/rand`.
- [x] Rejouer shuffle normal/race, staticcheck, GolangCI-Lint compatible Go 1.25, OSV, Trivy, SBOM CycloneDX/SPDX et inventaire de licences.
- [x] Documenter les exceptions résiduelles avec propriétaire, condition de levée et date de revue dans `V6_REMAINING_FINDINGS.md`.
- [ ] Corriger les diagnostics Staticcheck/GolangCI-Lint historiques restants après attribution et revue des propriétaires.
- [ ] Réconcilier les 46 résultats OSV de version stdlib/directed-toolchain avec un scanner compatible `go1.25.13`.
- [ ] Revoir les six misconfigurations Docker et les 741 licences `UNKNOWN` avec les propriétaires images/OSS.
- [ ] Restaurer les 14 objets LFS historiques nécessaires à un `git lfs fsck` complet.
- [ ] Obtenir la revue indépendante ; ne lever aucune gate et ne pas déclarer de release.

# Publication distante V6 vérifiée — 2026-08-25
- [x] Publier `audit/t00-t42-v6-findings-remediation` sur GitHub.
- [x] Vérifier la référence distante et le SHA `8e26bfb0c8bf6e92c09d645dd84ec854320c01f9`.
- [x] Cloner la branche dans un répertoire neuf avec `GIT_LFS_SKIP_SMUDGE=1` et confirmer le HEAD local identique.
- [x] Exécuter `git fsck --full` avant et après réhydratation LFS ciblée ; code 0.
- [x] Vérifier par hash le wrapper V6 `ce722915d70e0aa528927b753c6f18efa5706fc9fa8703ef6f449b6728a5fab6` et le bundle `ad4484e795b80eb5b7655228012e695dc4b260d43057477a97ae145d164614c2`.
- [ ] Obtenir la revue indépendante ; maintenir les gates et l’interdiction de release.
