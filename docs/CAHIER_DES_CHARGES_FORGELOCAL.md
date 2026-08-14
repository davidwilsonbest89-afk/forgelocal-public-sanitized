# Cahier des charges — ForgeLocal

**Version :** 1.0 — spécification produit versionnée ; release BACK-01 suspendue
**Identifiant du document :** `FL-CDC-1.0-20260814`
**Date :** 14 août 2026  
**Auteur :** Manus AI  
**Statut :** document de référence interne ; aucune autorisation de publication publique
**Empreinte vérifiable :** consignée dans `docs/CAHIER_DES_CHARGES_FORGELOCAL.v1.0.manifest.json`, avec le commit qui la porte.

## 1. Objet et principes de produit

ForgeLocal vise à devenir un outil **local-first** de gestion de profils navigateur : profils conservés sous le contrôle de l’utilisateur, API locale authentifiée, intégrations proxy sous contrôle de l’utilisateur et sauvegardes chiffrées. ForgeLocal n’impose **aucun quota commercial ou artificiel de profils**. La capacité effective dépend toutefois du stockage, de la mémoire, du CPU, des limites du système d’exploitation et de la charge du runtime ; elle doit être mesurée par tests de charge, sans promesse absolue d’« illimité ». Le produit ne doit jamais promettre une « non-détection » ni contourner illégalement les protections d’un service tiers.

> **Principe directeur :** BrowseForge Core en Go est l’unique source de vérité des mutations métier. Une interface React, Tauri ou une intégration fournisseur ne peut pas devenir un second backend.

La présente version sépare explicitement le **lot BACK-01 minimal**, qui est un composant Core/API de sauvegarde et restauration, de l’application ForgeLocal complète visée ultérieurement. Le lot BACK-01 ne doit pas être présenté comme un navigateur anti-détection complet, comme un produit desktop final, ni comme une distribution de Chromium ou Camoufox.

| Axe | Exigence de produit |
|---|---|
| Modèle de déploiement | Local par défaut ; aucune dépendance cloud obligatoire pour les données, profils, secrets ou backups. |
| Contrôle des données | Les données métier, les profils et les preuves restent sur la machine de l’utilisateur, sauf export explicite. |
| Capacité de profils | Aucun quota commercial ou artificiel ; la limite mesurée est celle des ressources de l’hôte et du runtime. |
| Sécurité | Secrets uniquement dans le coffre système ; jamais dans SQLite, JSON, logs, profils, archives non chiffrées ou interface web. |
| Architecture | Un seul Core Go propriétaire des écritures ; API locale liée par défaut à `127.0.0.1`. |
| Licence cible | Le code propriétaire ForgeLocal est sous licence permissive MIT ou Apache-2.0 ; un inventaire complet des licences, notices et obligations de toutes les dépendances est obligatoire avant toute redistribution. |
| Communication produit | Aucune promesse de « zéro détection », aucune annonce d’OS ou runtime non validé. |

## 2. État de référence au 14 août 2026

Le candidat RC BACK-01 est figé. Son archive, son SBOM, son manifeste, son commit source et son runtime de QA ne doivent pas être modifiés pendant la qualification. Le statut public est strictement `PUBLIC_RELEASE_BLOCKED`.

| Élément | Référence actuelle | État |
|---|---|---:|
| Archive RC | `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz` | Figée |
| SHA-256 RC | `553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6` | Vérifiée localement |
| Commit source | `67a8dfd897e540a55fc10749e1f2ef85b8356a8b` | Référence de chaîne |
| Runtime QA externe | Chromium `151.0.7922.108-1xtradeb1.2404.1` | E2E technique vert, non public |
| Cible annoncable à terme | Ubuntu 24.04.4 LTS `amd64` seulement | Preuve SystemVault native manquante |
| Décision publique | `PUBLIC_RELEASE_BLOCKED` | Obligatoire |
| Pilote local temporaire | Suspendu par précaution, en attente du triage d’une alerte de scan d’archive | Non utilisable avant résolution |

L’autorisation locale temporaire a été suspendue en mode défaillant fermé après la détection **assainie/redacted** d’une occurrence non classifiée dans une preuve de provenance incluse dans l’archive. Il ne faut ni exposer la valeur détectée, ni modifier silencieusement l’archive pour la retirer. Une revue mainteneur doit classifier formellement l’occurrence comme faux positif borné ou comme secret réel avant toute reprise du pilote.

## 3. Périmètre BACK-01 minimal

Le lot BACK-01 couvre seulement le noyau de sauvegarde chiffrée et restauration de profils. Il ne couvre ni l’expérience produit complète, ni les fonctions d’automatisation et de personnalisation navigateur présentes historiquement dans le dépôt.

| Fonction | Exigence | Statut de référence |
|---|---|---:|
| API locale authentifiée | Écoute locale par défaut sur `127.0.0.1:45100`; Bearer token obligatoire. | Inclus |
| Snapshot de profil | Snapshot verrouillé de `browser-data`, avec rejet des liens symboliques, chemins hostiles et TOCTOU. | Inclus |
| Backup | Format `FLBK … FLEND`, AES-256-GCM, AAD, checksum et publication atomique. | Inclus |
| Métadonnées | SQLite local, WAL, clés étrangères, réconciliation `published_unregistered`. | Inclus |
| Restauration | Restauration sous nouvel identifiant ou nouvel emplacement, sans écrasement implicite. | Inclus |
| Coffre système | `SecretRef` déterministe et clé hors stockage applicatif ; aucun fallback en clair. | Inclus sous gate natif |
| Audit local | Événements utiles, sans secret ni token en clair. | Inclus |
| Runtime navigateur | Runtime externe au binaire ; aucun runtime n’est distribué par BACK-01. | Hors artefact |
| Dashboard React / Tauri | Interface produit à construire séparément. | Hors périmètre |
| Proxy fournisseur | Connecteurs et modèle `proxy_connections` à construire séparément. | Hors périmètre |
| Fingerprinting, humanization, MCP, extensions | Exclus du build minimal et des revendications de release. | Exclus |
| Camoufox | Candidat produit ultérieur, à qualifier séparément. | Interdit pendant la qualification RC |

### 3.1 Critères d’acceptation BACK-01

| Identifiant | Critère vérifiable |
|---|---|
| AC-BACK-01 | Créer un profil, créer un backup chiffré, modifier la source, restaurer sous un nouvel ID, relancer le runtime approuvé avec le répertoire restauré, confirmer l’arrêt propre et l’absence de lock partagé. |
| AC-BACK-02 | Rejeter une archive corrompue, tronquée, à checksum invalide, AAD invalide, contenant un chemin hostile ou un lien symbolique. |
| AC-BACK-03 | Vérifier le verrou exclusif de snapshot et l’absence de partage de répertoire entre profil source et cible. |
| AC-BACK-04 | Simuler un arrêt après publication du fichier et vérifier la réconciliation SQLite `published_unregistered` au redémarrage. |
| AC-BACK-05 | Vérifier qu’aucun secret, clé de backup, token ou valeur proxy n’est présent dans SQLite, le JSON de profil, les logs ou le backup. |
| AC-BACK-06 | Confirmer l’échec explicite et sûr du coffre lorsqu’il n’est pas natif, déverrouillé ou accessible, sans fallback en clair. |

## 4. Exigences de sécurité et de confidentialité

### 4.1 Coffre système et secrets

Les secrets proxy, tokens de fournisseur et clés de chiffrement doivent être résolus via une référence (`secret_ref` ou `key_id`) et le backend de coffre propre au système. Aucune API, interface utilisateur ou migration ne doit accepter l’écriture d’un secret applicatif dans un champ métier en clair.

La validation SystemVault doit se faire dans une session graphique utilisateur réelle, non-root, hors conteneur, avec D-Bus de session, `XDG_RUNTIME_DIR` et Secret Service déverrouillé. Les cas obligatoires sont création/lecture, lecture après redémarrage du Core, clé absente, révocation, coffre verrouillé et permissions insuffisantes. Un refus contrôlé est attendu quand le coffre n’est pas utilisable.

> **Exigence de persistance :** une VM Ubuntu 24.04 persistante ou une installation dédiée est requise pour la qualification complète, notamment les cas de redémarrage, révocation et coffre verrouillé. Une session éphémère « Try Ubuntu » peut servir au préflight, mais ne constitue pas seule une preuve native complète.

Le dossier de preuve emploie **exclusivement** les noms contractuels du runbook `release/back01-minimal/SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md` : `systemvault-host-context.env`, `systemvault-matrix.json`, `SYSTEMVAULT_NATIVE_GATE_STATUS` et `systemvault-anti-leak.json`. Malgré son suffixe `.env`, `systemvault-host-context.env` est un **fichier de preuve généré et assaini/redacted**, jamais un fichier de configuration à sourcer : il ne contient aucun token, secret, mot de passe, clé privée, valeur proxy, variable sensible ni historique de session. Le répertoire optionnel `runtime-release-evidence/` et `PUBLIC_RELEASE_DECISION.md` complètent ce dossier lorsque la provenance runtime est disponible. Aucun nom alternatif ne constitue une preuve de gate.

### 4.2 Backups et restauration

Le flux `profil → backup chiffré → modification → restauration isolée` est le flux de référence. La restauration doit échouer de manière atomique ou être entièrement récupérée ; elle ne doit jamais écraser un profil existant sans une action explicite et enregistrée. Les archives exportées doivent être traitées comme non fiables jusqu’à vérification complète du format, de l’authenticité et des chemins.

### 4.3 API locale et interface future

L’API doit rester liée par défaut à la boucle locale et exiger un Bearer token envoyé explicitement dans l’en-tête `Authorization`. L’interface React future ne doit jamais stocker de token ni de secret dans le navigateur de façon persistante. Les opérations mutantes doivent journaliser l’identité locale, l’action, l’horodatage, le profil concerné et le résultat, sans enregistrer de données secrètes.

## 5. Gates obligatoires de publication

Les cinq gates ci-dessous sont tous requis. Une seule décision `PENDING` ou `FAILED` impose `PUBLIC_RELEASE_BLOCKED`.

| Gate | Responsable à désigner | Preuve requise | Décision actuelle |
|---|---|---|---:|
| `SYSTEMVAULT_NATIVE_PER_TARGET` | Platform Security Owner | Matrice native réussie par OS/version/architecture annoncés. | `PENDING` |
| `SYSTEMVAULT_ANTI_LEAK_INTEGRATED_FLOW` | Security QA Owner | `systemvault-anti-leak.json` produit par le flux réel sans révéler la sentinelle. | `PENDING` |
| `MAINTAINER_MANIFEST_SIGNATURE` | Release Maintainer | Signature détachée, clé publique séparée, empreinte approuvée et vérification indépendante. | `PENDING` |
| `RUNTIME_LICENSE_AND_REDISTRIBUTION_REVIEW` | Release Compliance Owner | Revue des paquets runtime exacts, licences et droit de redistribution. | `PENDING` |
| `OS_COMPATIBILITY_EVIDENCE` | Release QA Owner | Matrice limitée aux cibles avec runtime et coffre natif entièrement qualifiés. | `PENDING` |

Toute preuve `PASSED` devient invalide si le commit testé, le SHA-256 de l’archive, la version du runtime, l’OS, l’architecture, la configuration cible ou le hash de preuve change. Les preuves d’un runtime ou d’un artefact ne peuvent jamais être réutilisées pour un autre candidat.

### 5.1 Distinction impérative entre gates internes et gates publics

Les gates de développement internes `G0` à `G6` du lot produit servent uniquement à organiser les jalons techniques. Ils ne remplacent jamais les cinq gates publics machine-readable ci-dessus et ne peuvent ni les passer automatiquement, ni transformer `PUBLIC_RELEASE_BLOCKED` en approbation. Seul `PUBLIC_RELEASE_GATE_STATE.json`, avec des preuves cohérentes et une revue indépendante, décide du statut public.

## 6. Qualité, tests et sécurité de la chaîne de livraison

### 6.1 Toolchain et outillage reproductibles

La CI doit s’exécuter avec une distribution Go **exactement** égale à `1.25.13`, consignée dans l’environnement de build, avec `GOTOOLCHAIN=local`, et vérifiée au début du job par `go version` et `go env GOVERSION GOOS GOARCH`. Tout écart doit faire **échouer explicitement** le job ; aucun téléchargement implicite de toolchain différent n’est autorisé. Le manifeste machine-readable [`tools/TOOLCHAIN_LOCK.json`](../tools/TOOLCHAIN_LOCK.json) fige les sources et empreintes SHA-256. L’usage de `@latest` est interdit dans la CI et les preuves de release.

| Outil | Version verrouillée | Artefact Linux amd64 ou source | SHA-256 | Source de checksum |
|---|---:|---|---|---|
| Go | `1.25.13` | `go1.25.13.linux-amd64.tar.gz` | `39042a078ea9ceebe3ecda4a7188f0f5b96e14a071d27923ba7f40b456e85ae3` | Catalogue JSON [go.dev](https://go.dev/dl/?mode=json&include=all) |
| Gosec | `2.21.4` | `gosec_2.21.4_linux_amd64.tar.gz` | `9229dbfdc092b176e628b9ea6e4210757373b819f47365cedd9f9e12d3b2c173` | `gosec_2.21.4_checksums.txt` officiel |
| Govulncheck | `1.7.0` | Module `golang.org/x/vuln` v`1.7.0` | `a14bf913551ac09f00ae0e903c1b358713f71af911d7ddacc3fab8ce5c149a26` | Archive du proxy de modules Go |
| GolangCI-Lint | `1.61.0` | `golangci-lint-1.61.0-linux-amd64.tar.gz` | `77cb0af99379d9a21d5dc8c38364d060e864a01bd2f3e30b5e8cc550c3a54111` | `golangci-lint-1.61.0-checksums.txt` officiel |
| Gitleaks | `8.18.4` | `gitleaks_8.18.4_linux_x64.tar.gz` | `ba6dbb656933921c775ee5a2d1c13a91046e7952e9d919f9bac4cec61d628e7d` | `gitleaks_8.18.4_checksums.txt` officiel |
| SQLite CLI/source | `3.45.1` | `sqlite-autoconf-3450100.tar.gz` | `cd9c27841b7a5932c9897651e20b86c701dd740556989b01ca596fcfa3d49a0a` | Archive source SQLite officielle |

Avant installation, la CI calcule le SHA-256 local et le compare au manifeste ; elle refuse toute divergence. Pour `govulncheck`, l’archive de module est vérifiée avant la construction avec le Go verrouillé ; le sélecteur `@latest` n’est jamais admis.

Avant tout filtre de tests, le job doit enregistrer `go list ./...` et vérifier que chaque package ciblé existe dans le commit testé. Chaque commande doit ensuite prouver qu’au moins un test attendu est sélectionné ; un code de sortie `0` avec zéro test sélectionné est un échec de validation.

Les contrôles suivants sont reproductibles dans le sandbox et doivent être exécutés à chaque changement de code BACK-01 ou de documentation de release pertinente.

| Contrôle | Commande de référence | Exigence |
|---|---|---|
| Inventaire et tests ciblés API | Voir la procédure `T-API-AC-BACK-01` ci-dessous. | Les packages doivent exister ; le test AC-BACK-01 doit être effectivement sélectionné et vert. |
| Suite complète | `go test ./... -count=1` | Vert. |
| Courses BACK-01 | `go test -race ./internal/backup ./internal/profile ./internal/secrets ./cmd/back01-core -count=1` | Vert. |
| Analyse statique Go | `go vet ./...` | Vert. |
| Gosec minimal | `gosec ./internal/backup/... ./internal/profile/... ./internal/secrets/... ./cmd/back01-core/...` | Aucune alerte nouvelle non justifiée. |
| Dépendances minimales | `go list -deps ./cmd/back01-core` | Aucun package interne interdit. |
| Traçabilité | `scripts/validate-release-traceability.py` | Vert, avec chaînes indépendantes. |
| Décision publique | `scripts/check-public-release-gate.py` | Doit retourner le statut attendu, actuellement bloqué. |
| Scan de secrets | Scan effectué sur les **contenus originaux**, conservés localement et protégés : dépôt, historique Git pertinent, arborescence staged, archive extraite, SBOM, manifestes et journaux de build. Seules les sorties et preuves exportées sont **assainies/redacted**. | `SCAN_CLEAN` uniquement si aucune occurrence non classifiée ne subsiste ; toute exception doit être minimale, versionnée et revue. |

#### T-API-AC-BACK-01 — inventaire et restauration isolée

Les trois commandes suivantes sont distinctes, doivent être exécutées avec Go `1.25.13` et `GOTOOLCHAIN=local`, puis leurs sorties assainies doivent être archivées avec le commit testé :

```bash
export GOTOOLCHAIN=local
go version
go env GOVERSION GOOS GOARCH
go list ./...
go test ./internal/api -list 'Backup|Restore' -count=1
go test ./internal/api -run '^TestBackupV1CreateModifyRestoreIsolation$' -count=1 -v
```

Le préflight doit afficher exactement `go version go1.25.13` et ne doit déclencher aucun téléchargement implicite de toolchain. La commande d’inventaire `go list ./...` doit contenir exactement le package cible `forgelocal/internal/api`. La commande de sélection ciblée `go test ./internal/api -list 'Backup|Restore' -count=1` doit afficher les noms exacts des tests correspondants et le rapport doit consigner leur nombre. La commande d’exécution exacte exécute exclusivement `TestBackupV1CreateModifyRestoreIsolation` : le rapport doit compter une ligne `=== RUN` pour ce nom, une ligne `--- PASS` correspondante, zéro autre ligne `=== RUN`, et un résultat final `PASS`. Toute absence de package, sélection nulle, test supplémentaire ou échec rend le contrôle non conforme, même si le processus retourne le code `0`.

La suite actuelle passe dans le sandbox pour les tests Go, le détecteur de courses sur les modules BACK-01 ciblés, `go vet`, la fermeture de dépendances minimale et le Gosec limité au graphe de distribution. La dette Gosec possède deux références qui ne doivent pas être confondues : le rapport historique du **14 août 2026** contient 189 résultats (`validation_back01_integration/final/sandbox-gosec-20260814.json`, SHA-256 `332ae84056ec9ad5d15a965674540f0d5f21215bbeb998dd6bf614516c65b978`), mais il déclare `GosecVersion: dev` et n’associe pas un commit propre ; il est donc une **preuve historique non conforme** à la présente exigence de verrouillage. La baseline reproductible de contrôle, exécutée le 14 août 2026 avec Gosec `2.21.4` sur le commit `64bede39dc3355e0db2c4871cf4de7eb46410265`, contient 166 résultats (rapport SHA-256 `ba42d3e2af1fe8d9a61407ed87d54c57b7d81cccfaddc9fa382b07f35e06ec9d`). Aucun résultat n’est ignoré : le rapport historique à 189 résultats doit être régénéré sous l’outillage verrouillé avant toute décision fondée sur son nombre. Les conclusions de réduction doivent référencer à la fois le commit, la version Gosec, la date et le hash du rapport. Les concentrations historiques dans `internal/browser`, `cmd/server` et les fonctions humanization/fingerprint restent hors du binaire BACK-01 minimal, sans être dispensées de classification.

> **État de scan actuel : `SCAN_BLOCKED_UNKNOWN`.** Une occurrence `generic-api-key` non classifiée est présente dans une preuve incluse dans l’archive RC. Tant que le dossier mainteneur et sa revue indépendante ne concluent pas `REAL_SECRET` ou `FALSE_POSITIVE`, aucun scan ne peut être qualifié `SCAN_CLEAN`, le pilote reste suspendu et aucune qualification SystemVault ne reprend pour ce candidat.

> **Règle de qualité :** toute fonctionnalité BACK-01 modifiée doit avoir des tests unitaires et d’intégration qui couvrent les chemins de succès, erreur et récupération. Les commandes `cmd/back01-core` et `cmd/systemvault-doctor`, ainsi que `internal/secrets`, doivent recevoir des tests dédiés avant toute extension du périmètre de release.

## 7. Feuille de route produit après BACK-01

### 7.1 Lot P1 — noyau métier SQLite et profils

Le noyau doit migrer l’état métier actuellement hérité vers SQLite de manière réversible. La migration doit inclure un mode `dry-run`, un rapport de parité JSON/SQLite, une sauvegarde pré-migration et un rollback documenté. Le Core Go reste l’unique écrivain ; un processus frontend ne modifie jamais directement la base.

Le schéma canonique est la migration `internal/backup/migrations/0002_product.sql`, documentée dans [`SQLITE_SCHEMA_REFERENCE.md`](SQLITE_SCHEMA_REFERENCE.md). Il comprend `profiles`, `groups`, `profile_tags`, `profile_tag_assignments`, `proxy_providers`, `proxy_test_runs`, `runtime_candidates`, `profile_import_operations`, `profile_json_parity_checks`, `product_audit_events` ainsi que les tables BACK-01 `backup_operations`, `backups`, `restore_operations` et `audit_events`. Le schéma ne crée pas `profile_groups`, `proxy_connections` ni `profile_proxy_assignments` : les relations réellement versionnées sont `profiles.group_id`, les tables d’étiquettes et les références `proxy_provider_id`/`proxy_secret_ref`. Les données secrètes restent hors de toutes ces tables ; seuls des identifiants opaques de coffre sont persistés.

### 7.2 Lot P2 — dashboard React local

Le dashboard React doit être une interface locale qui consomme l’API authentifiée du Core. Les écrans minimaux sont la liste/recherche de profils, création/édition, groupes, statut de runtime, lancement/arrêt, backups, restauration, journal d’audit, diagnostic local et gestion des références de secrets.

Chaque écran mutateur doit gérer la concurrence, afficher les erreurs de façon actionnable, exiger une confirmation pour les suppressions/restaurations destructives et ne jamais afficher une valeur secrète après son enregistrement. Une page de diagnostic doit distinguer les tests locaux, les tests de coffre natif et les gates de publication.

### 7.3 Lot P3 — connecteurs proxy sous contrôle utilisateur

Les fournisseurs tels que Decodo doivent être intégrés derrière une interface générique et versionnée `ProxyProvider`, jamais directement dans le cœur métier. L’interface doit permettre à un adaptateur optionnel de découvrir des offres, valider la connectivité, importer un endpoint sous contrôle, signaler son état et associer explicitement un proxy à un profil. Les identifiants API et mots de passe de proxy restent dans le coffre système ; le Core ne conserve que des références non sensibles.

Aucune interface ne doit importer automatiquement des proxies, créer une facturation fournisseur ou activer une rotation de proxy sans consentement explicite de l’utilisateur. La documentation doit préciser que les règles et contrats de chaque fournisseur s’appliquent. ForgeLocal ne doit dépendre fonctionnellement d’aucun fournisseur unique.

### 7.4 Lot P4 — runtime et wrapper desktop

Camoufox est un candidat de runtime produit distinct. Sa licence MPL-2.0, ses modalités de distribution, ses mises à jour, ses signatures et son comportement SystemVault doivent être validés dans une chaîne de preuves propre avant toute annonce. Il ne peut pas être ajouté au RC BACK-01 figé.

Tauri devient envisageable seulement après stabilisation de l’API et du dashboard React. Il doit encapsuler le Core existant et ne jamais dupliquer la logique de backup, de coffre, de migration ou de gestion des profils.

## 8. Décisions explicitement hors périmètre ou interdites

| Sujet | Décision |
|---|---|
| Publication publique BACK-01 | Interdite tant que les cinq gates ne sont pas tous `PASSED` et revus indépendamment. |
| Support macOS, Windows ou autre distribution | Interdit de l’annoncer avant matrices natives propres par cible. |
| Ajout de Camoufox au candidat RC | Interdit pendant la qualification. |
| Remplacement silencieux du runtime Chromium | Interdit ; toute version est un candidat distinct avec nouvelle preuve E2E. |
| Fallback de secret en clair | Interdit sans exception. |
| Second backend dans React ou Tauri | Interdit. |
| Promesse de non-détection | Interdite. |
| Distribution de dépendances runtime sans revue | Interdite. |

## 9. Prochaines actions ordonnées

| Priorité | Action | Condition de clôture |
|---|---|---|
| P0-1 | Triage mainteneur indépendant de l’alerte de scan de l’archive RC | Conclusion `FALSE_POSITIVE` bornée avec rescan assaini/redacted vert, ou secret révoqué avant émission d’un nouveau candidat complet. |
| P0-2 | Émettre une nouvelle chaîne de release propre, si le triage impose une correction de preuve, d’artefact ou de provenance | Nouvel artefact, SBOM, manifestes, checksums et preuves cohérentes ; le RC concerné reste gelé et non qualifié. |
| P0-3 | Exécuter le runbook SystemVault sur une VM Ubuntu 24.04.4 LTS `amd64` persistante ou une installation dédiée, hors conteneur | Autorisé uniquement après clôture indépendante de P0-1 et contre une chaîne propre de P0-2 lorsque requise ; `systemvault-host-context.env`, `systemvault-matrix.json`, `SYSTEMVAULT_NATIVE_GATE_STATUS` et `systemvault-anti-leak.json` assainis, versionnés et revus indépendamment. |
| P0 | Produire l’anti-fuite du flux intégré | `systemvault-anti-leak.json` vert, même chaîne artefact/runtime/commit. |
| P0 | Revue de licence et redistribution Chromium | Décision écrite sur les paquets exacts et leur distribution. |
| P0 | Signature mainteneur | Manifeste signé avec clé publique séparée et vérification indépendante. |
| P1 | Réduire et classer la dette Gosec du dépôt historique | Registre de décisions par règle, fichier, risque, propriétaire, échéance et test de non-régression. |
| P1 | Produire l’inventaire de licences de distribution | Licences, notices et obligations de chaque dépendance et runtime évalués avant redistribution. |
| P1 | Intégrer la commande contrôlée de migration SQLite métier au Core | Importeur versionné disponible : dry-run par défaut, préimage BACK-01 chiffré et vérifié, parité lue depuis SQLite, rollback transactionnel, interruption et reprise testés. Aucun frontend ne modifie directement SQLite. |
| P1 | Construire le dashboard React local | Écrans de gestion de profils, backups, audit et diagnostic avec API unique. |
| P2 | Intégrer les fournisseurs proxy | Adaptateurs, coffre système, tests de connectivité et consentement explicite. |
| P2 | Qualifier Camoufox puis Tauri | Chaînes de preuve et revues de licence distinctes. |

## 10. Références internes

| Référence | Document |
|---|---|
| R1 | `release/back01-minimal/PROFILE.md` |
| R2 | `release/back01-minimal/PUBLIC_RELEASE_GATE_STATE.json` |
| R3 | `release/back01-minimal/RELEASE_SCOPE_AND_OS_MATRIX.md` |
| R4 | `release/back01-minimal/PUBLIC_RELEASE_DECISION.md` |
| R5 | `release/back01-minimal/SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md` |
| R6 | `release/back01-minimal/RELEASE_TRACEABILITY_INDEX.json` |
| R7 | `validation_back01_integration/final/CURRENT_PRODUCT_SCOPE_INVENTORY_2026-08-14.md` |
| R8 | `tools/TOOLCHAIN_LOCK.json` |
| R9 | `docs/SQLITE_SCHEMA_REFERENCE.md` |
| R10 | `validation_back01_integration/final/GOSEC_BASELINE_RECONCILIATION_2026-08-14.md` |
| R11 | `validation_back01_integration/final/gosec-baseline-64bede3-v2.21.4-20260814.json` |
| R12 | `docs/CAHIER_DES_CHARGES_FORGELOCAL.v1.0.manifest.json` |

> Ce cahier des charges est un document de pilotage. Il ne remplace pas les validateurs machine-readable de release et ne peut pas à lui seul lever un gate de publication.
