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
