# Cahier des charges — ForgeLocal

**Version :** 0.2 — consolidation BACK-01 et feuille de route produit  
**Date :** 14 août 2026  
**Auteur :** Manus AI  
**Statut :** document de référence interne ; aucune autorisation de publication publique

## 1. Objet et principes de produit

ForgeLocal vise à devenir un outil **local-first** de gestion de profils navigateur : profils conservés sous le contrôle de l’utilisateur, nombre de profils non limité par un abonnement cloud, API locale authentifiée, intégrations proxy sous contrôle de l’utilisateur et sauvegardes chiffrées. Le produit ne doit jamais promettre une « non-détection » ni contourner illégalement les protections d’un service tiers.

> **Principe directeur :** BrowseForge Core en Go est l’unique source de vérité des mutations métier. Une interface React, Tauri ou une intégration fournisseur ne peut pas devenir un second backend.

La présente version sépare explicitement le **lot BACK-01 minimal**, qui est un composant Core/API de sauvegarde et restauration, de l’application ForgeLocal complète visée ultérieurement. Le lot BACK-01 ne doit pas être présenté comme un navigateur anti-détection complet, comme un produit desktop final, ni comme une distribution de Chromium ou Camoufox.

| Axe | Exigence de produit |
|---|---|
| Modèle de déploiement | Local par défaut ; aucune dépendance cloud obligatoire pour les données, profils, secrets ou backups. |
| Contrôle des données | Les données métier, les profils et les preuves restent sur la machine de l’utilisateur, sauf export explicite. |
| Sécurité | Secrets uniquement dans le coffre système ; jamais dans SQLite, JSON, logs, profils, archives non chiffrées ou interface web. |
| Architecture | Un seul Core Go propriétaire des écritures ; API locale liée par défaut à `127.0.0.1`. |
| Licence cible | Composants propres sous licence permissive MIT ou Apache-2.0 ; toute dépendance à obligations particulières doit être revue avant redistribution. |
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

L’autorisation locale temporaire a été suspendue en mode défaillant fermé après la détection redigée d’une occurrence non classifiée dans une preuve de provenance incluse dans l’archive. Il ne faut ni exposer la valeur détectée, ni modifier silencieusement l’archive pour la retirer. Une revue mainteneur doit classifier formellement l’occurrence comme faux positif borné ou comme secret réel avant toute reprise du pilote.

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

## 6. Qualité, tests et sécurité de la chaîne de livraison

Les contrôles suivants sont reproductibles dans le sandbox et doivent être exécutés à chaque changement de code BACK-01 ou de documentation de release pertinente.

| Contrôle | Commande de référence | Exigence |
|---|---|---|
| Tests ciblés API | `go test ./internal/api -list 'Backup|Restore'` puis `go test ./internal/api -run 'TestBackupV1CreateModifyRestoreIsolation' -count=1 -v` | Le test AC-BACK-01 doit être effectivement sélectionné et vert. |
| Suite complète | `go test ./... -count=1` | Vert. |
| Courses BACK-01 | `go test -race ./internal/backup ./internal/profile ./internal/secrets ./cmd/back01-core -count=1` | Vert. |
| Analyse statique Go | `go vet ./...` | Vert. |
| Gosec minimal | `gosec ./internal/backup/... ./internal/profile/... ./internal/secrets/... ./cmd/back01-core/...` | Aucune alerte nouvelle non justifiée. |
| Dépendances minimales | `go list -deps ./cmd/back01-core` | Aucun package interne interdit. |
| Traçabilité | `scripts/validate-release-traceability.py` | Vert, avec chaînes indépendantes. |
| Décision publique | `scripts/check-public-release-gate.py` | Doit retourner le statut attendu, actuellement bloqué. |
| Scan de secrets | Scan redigé du dépôt, de l’arborescence staged et de l’archive extraite. | Zéro fuite réelle ; toute exception est versionnée et bornée. |

La suite actuelle passe dans le sandbox pour les tests Go, le détecteur de courses sur les modules BACK-01 ciblés, `go vet`, la fermeture de dépendances minimale et le Gosec limité au graphe de distribution. Le scan Gosec sur l’ensemble du dépôt retourne 189 résultats : ils sont de la dette héritée et ne doivent pas être ignorés. La plupart se concentrent dans `internal/browser` et `cmd/server`, qui sont hors du binaire BACK-01 minimal ; les 17 alertes G404 de génération pseudo-aléatoire sont concentrées dans les fonctionnalités historiques de humanization/fingerprint, elles aussi exclues du lot.

> **Règle de qualité :** toute fonctionnalité BACK-01 modifiée doit avoir des tests unitaires et d’intégration qui couvrent les chemins de succès, erreur et récupération. Les commandes `cmd/back01-core` et `cmd/systemvault-doctor`, ainsi que `internal/secrets`, doivent recevoir des tests dédiés avant toute extension du périmètre de release.

## 7. Feuille de route produit après BACK-01

### 7.1 Lot P1 — noyau métier SQLite et profils

Le noyau doit migrer l’état métier actuellement hérité vers SQLite de manière réversible. La migration doit inclure un mode `dry-run`, un rapport de parité JSON/SQLite, une sauvegarde pré-migration et un rollback documenté. Le Core Go reste l’unique écrivain ; un processus frontend ne modifie jamais directement la base.

Les tables cibles comprennent au minimum `profiles`, `profile_groups`, `proxy_connections`, `profile_proxy_assignments`, `audit_events`, `runtime_candidates`, `backup_operations`, `backups` et `restore_operations`. Les données secrètes doivent rester hors de ces tables : `proxy_connections` contient uniquement une référence de coffre et des champs non secrets.

### 7.2 Lot P2 — dashboard React local

Le dashboard React doit être une interface locale qui consomme l’API authentifiée du Core. Les écrans minimaux sont la liste/recherche de profils, création/édition, groupes, statut de runtime, lancement/arrêt, backups, restauration, journal d’audit, diagnostic local et gestion des références de secrets.

Chaque écran mutateur doit gérer la concurrence, afficher les erreurs de façon actionnable, exiger une confirmation pour les suppressions/restaurations destructives et ne jamais afficher une valeur secrète après son enregistrement. Une page de diagnostic doit distinguer les tests locaux, les tests de coffre natif et les gates de publication.

### 7.3 Lot P3 — connecteurs proxy sous contrôle utilisateur

Les fournisseurs tels que Decodo doivent être intégrés derrière un adaptateur versionné. Les identifiants API et mots de passe de proxy restent dans le coffre système ; le Core ne conserve que des références. Les fonctions attendues sont découverte des offres, validation de connectivité, import contrôlé d’endpoint, association explicite à un profil et diagnostic sans révélation de secret.

Aucune interface ne doit importer automatiquement des proxies, créer une facturation fournisseur ou activer une rotation de proxy sans consentement explicite de l’utilisateur. La documentation doit préciser que les règles et contrats de chaque fournisseur s’appliquent.

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
| P0 | Triage mainteneur de l’alerte de scan de l’archive RC | Faux positif borné et scan redigé vert, ou secret révoqué avec nouveau candidat complet. |
| P0 | Exécuter le runbook SystemVault sur Ubuntu 24.04.4 LTS `amd64` hors conteneur | Quatre sorties assainies versionnées et revue indépendante. |
| P0 | Produire l’anti-fuite du flux intégré | `systemvault-anti-leak.json` vert, même chaîne artefact/runtime/commit. |
| P0 | Revue de licence et redistribution Chromium | Décision écrite sur les paquets exacts et leur distribution. |
| P0 | Signature mainteneur | Manifeste signé avec clé publique séparée et vérification indépendante. |
| P1 | Réduire et classer la dette Gosec du dépôt historique | Registre de décisions par règle, fichier, risque, propriétaire, échéance et test de non-régression. |
| P1 | Implémenter la migration SQLite métier | Dry-run, parité, rollback, tests de migration. |
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

> Ce cahier des charges est un document de pilotage. Il ne remplace pas les validateurs machine-readable de release et ne peut pas à lui seul lever un gate de publication.
