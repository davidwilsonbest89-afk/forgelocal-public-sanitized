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
- [ ] T07-R1 : consigner la chaîne de liaison vérifiable `révision privée attestée → hash du snapshot étudié → décision de portage/réimplémentation/écarter → futur commit ForgeLocal → tests`, sans source privée ni clé dans Git.
- [ ] T07-R2 : obtenir ou référencer sous contrôle d’accès une preuve indépendante de propriété/droits, une licence racine ou un accord attesté et les notices tierces applicables.
- [ ] T07-R3 : soumettre l’alerte redacted de `tests/smoke.test.js:24` à un mainteneur et une relectrice indépendante ; conserver `UNKNOWN` et le blocage tant qu’aucune décision `REAL_SECRET` ou `FALSE_POSITIVE` n’est prouvée.
- [ ] T07-R4 : si une décision `FALSE_POSITIVE` est produite, exiger un snapshot candidat nouveau, redacted, rescané et hashé ; si `REAL_SECRET`, exiger révocation/rotation avant toute nouvelle preuve.
- [ ] T07-R5 : reconstruire une archive T07 redacted avec registre, CI, SBOM ciblée, scans et manifeste portable, puis demander une revue indépendante ; ne pas commencer T08 avant `PROV-01`, `PROV-04` et `PROV-06` à `PASS`.
