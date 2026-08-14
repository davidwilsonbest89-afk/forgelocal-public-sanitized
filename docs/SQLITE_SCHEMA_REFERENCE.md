# Référence du schéma SQLite ForgeLocal

**Statut :** baseline versionnée du Lot Produit v0.3. Cette référence décrit le schéma réellement porté par la migration `internal/backup/migrations/0002_product.sql` et sa migration additive `0003_proxy_reference_indexes.sql`, appliquées après `0001_back01.sql` par `internal/backup/migrations.go`. Elles ne modifient ni le candidat RC BACK-01 figé, ni son runtime, ni son SBOM, ni ses gates de release. La migration des données JSON elle-même reste à implémenter et doit respecter le plan de migration défini dans ce document.

> **Principe d’architecture.** SQLite conserve l’état métier local de ForgeLocal. Le Core Go est son seul écrivain. Les données de navigation restent dans les répertoires `browser-data` privés ; les mots de passe, jetons, clés API et identifiants proxy restent exclusivement dans le coffre système.

## Portée, versionnement et garanties

La table historique `schema_migrations` créée par BACK-01 reste l’unique registre de versions. La migration BACK-01 est la version **1** ; le schéma métier est la version **2** ; les index de références proxy sont la version additive **3**. Le chargeur applique les migrations manquantes dans l’ordre et dans une transaction : le DDL et l’inscription de la version sont validés ensemble. Une base déjà au niveau 1 est donc mise à niveau vers 3 une seule fois ; une nouvelle exécution est idempotente.

| Élément | Décision versionnée | Motivation |
|---|---|---|
| Registre des migrations | `schema_migrations` existant | Une base locale, un historique ordonné, aucun second moteur de migration. |
| Migration BACK-01 | `0001_back01.sql` | Backups, restauration et audit déjà qualifiés. |
| Migration Produit | `0002_product.sql` | Métadonnées de profils, groupes, runtime et proxy générique ; types, FK et contrôles de format des références proxy. |
| Migration Produit additive | `0003_proxy_reference_indexes.sql` | Index partiels sur `proxy_provider_id` et `proxy_secret_ref` de `profiles` et `groups`, sans ajouter de secret ni modifier les tables v2. |
| Ordonnancement | `backup.Migrate` | Le Core Go demeure le seul responsable des mutations SQLite. |
| Modes SQLite | WAL, `foreign_keys=ON`, `busy_timeout=5000` | Cohérence locale, contraintes référentielles et comportement concurrent prévisible. |
| Compatibilité de release | Aucun impact sur le RC BACK-01 gelé | Le travail produit reste isolé sur `forgelocal-product-v0.3`. |

## Invariants de sécurité

Aucune table de ce schéma ne contient de colonne pour un mot de passe, un nom d’utilisateur, un jeton ou une clé API. Les seules références liées à un coffre sont les champs `*_secret_ref`, contrôlés par format et interprétés uniquement par l’adaptateur de coffre du Core. Le contenu d’une référence n’est pas une capacité d’authentification et ne permet pas de récupérer le secret sans accès au coffre système de l’utilisateur.

Les champs JSON (`identity_json`, `fingerprint_json`, `metadata_json`, `public_config_json`, `evidence_json`, `summary_json`, `details_json`) sont tous soumis à `json_valid`. Les services qui les écrivent doivent appliquer un contrat supplémentaire : les données d’audit et de preuve sont **redacted**, et les secrets ne doivent jamais y être sérialisés.

| Règle | Application dans le schéma | Contrôle applicatif complémentaire |
|---|---|---|
| Aucun secret SQLite | Aucune colonne `password`, `username`, `token` ou `api_key` | Tests anti-fuite et analyse de toutes les sorties avant archivage. |
| Référence de coffre bornée | `proxy.<id>`, `proxy.group.<id>`, `proxy.provider.<id>` | Le Core génère/valide la référence ; l’API ne choisit jamais une référence arbitraire. |
| Aucun effacement silencieux | Clés étrangères `ON DELETE RESTRICT` pour runtime, fournisseur et groupe référencés | Migration explicite, suppression contrôlée et auditée seulement. |
| Données navigateur hors SQLite | `profile_dir` est une référence de chemin unique | Validation canonique de chemin et permissions 0700 côté Core. |
| Runtime explicite | `profiles.runtime_id` référence `runtime_candidates` | Le lancement est autorisé par une politique de runtime, non par la seule présence d’une ligne. |

## Tables du schéma métier

La table s’appelle volontairement **`groups`**, et non `profile_groups`. Le code existant emploie déjà ce concept dans `internal/groups/store.go` et dans `groups.json`; conserver ce nom évite un second vocabulaire pour la même entité. Elle n’entre pas en conflit avec les tables BACK-01 existantes.

| Table | Clé et relations | Rôle local | Contraintes essentielles |
|---|---|---|---|
| `proxy_providers` | `id`; référencée par profils et groupes | Instance optionnelle d’un adaptateur `ProxyProvider` générique, par exemple `manual` ou un futur adaptateur Decodo. | `display_name` unique, configuration publique JSON, `enabled` booléen, aucune dépendance à un fournisseur dans le cœur. |
| `runtime_candidates` | `id`; référencée par `profiles.runtime_id` | Runtime disponible localement et sa provenance binaire. | Unicité du tuple nom/version/architecture/chemin et du SHA-256 ; états `candidate`, `validated`, `quarantined`, `retired`. |
| `groups` | `id`; références facultatives au fournisseur | Politique proxy par groupe, avec héritage possible par profil. | Type `http`/`socks5`, hôte et port cohérents ; mode `enforced` impossible sans proxy complet. |
| `profiles` | `id`; FK runtime, groupe et fournisseur | Métadonnées de profil, répertoire local, identité, empreinte et proxy direct facultatif. | `profile_dir` unique ; runtime requis ; proxy complet ou absent ; cycle `active`/`archived`/`quarantined`. |
| `profile_tags` | `id`; `name` unique | Dictionnaire de tags. | Nom sans doublon indépendant de la casse. |
| `profile_tag_assignments` | PK composée profile/tag | Relation N–N des tags de profil. | Suppression d’un profil en cascade ; suppression d’un tag référencé refusée. |
| `proxy_test_runs` | `id`; référence exactement un profil ou un groupe | Résultat redacted d’un test de proxy. | Une seule cible, résultat borné, latence non négative et code d’erreur sans secret. |
| `profile_import_operations` | `id` | Journal de dry-run, validation, import, rollback et erreur de migration JSON. | Source hachée, état borné, rapport JSON assaini et `correlation_id`. |
| `profile_json_parity_checks` | `id`; FK profil | Preuve de parité entre une source JSON et l’enregistrement SQLite canonique. | Hash source, hash canonique, résultat explicite et corrélation. |
| `product_audit_events` | clé auto-incrémentée | Audit métier redacted. | Type d’entité, corrélation et JSON valide ; jamais de secret. |

Les tables BACK-01 `backup_operations`, `backups`, `restore_operations` et `audit_events` restent inchangées. Elles conservent leur responsabilité de sauvegarde/restauration et ne deviennent pas le modèle métier des profils.

## Modèle de profil et correspondance JSON

Le modèle JSON actuel est défini par `internal/profile/store.go`. La table `profiles` conserve les éléments structurants tandis que les objets extensibles restent en JSON valide. Cette répartition évite la sérialisation d’un objet de profil entier dans une seule colonne tout en préservant l’évolution contrôlée de l’empreinte.

| Champ JSON existant | Destination SQLite | Traitement de migration |
|---|---|---|
| `id` | `profiles.id` | Conservé sans régénération. Les IDs invalides ou dupliqués bloquent la validation. |
| `name` | `profiles.name` | Conservé après validation d’unicité locale prévue par le service. |
| `runtime_id` | `profiles.runtime_id` et `runtime_candidates` | Créer ou réconcilier d’abord le runtime candidat ; aucun lancement implicite. |
| `identity` | `profiles.identity_json` | Canonicalisation JSON avant hash de parité. |
| `group` | `profiles.group_id` via `groups.name` | Résolution par nom normalisé. Une référence absente produit un conflit de dry-run. |
| `tags[]` | `profile_tags` + `profile_tag_assignments` | Déduplication insensible à la casse, conservation de la relation N–N. |
| `created_at`, `last_used` | `created_at`, `last_used_at` | ISO-8601 UTC ; une valeur absente est signalée dans le rapport. |
| `fingerprint` | `fingerprint_json` | Canonicalisation JSON ; aucune donnée secrète admise. |
| `fingerprint_seed` | `fingerprint_seed` | Copie entière bornée par la validation applicative. |
| `container_id` | `container_id` | Valeur unique lorsqu’elle est non vide. |
| `profile_dir` | `profile_dir` | Chemin canonique géré par le Core ; jamais accepté comme chemin arbitraire côté API. |
| `proxy.type/host/port/region` | Colonnes proxy de `profiles` | Copie seulement de données non secrètes après validation type/hôte/port. |
| `proxy.secret_ref` | `profiles.proxy_secret_ref` | Normalisation vers `proxy.<profile_id>` ; une référence arbitraire est un conflit bloquant. |
| `proxy.username/password` | Nulle part dans SQLite | Lecture/écriture exclusivement dans le coffre. Ils ne figurent pas dans les rapports de parité. |

Les groupes JSON existants suivent la même règle. Leur `name`, `proxy_mode` et les paramètres publics de proxy vont dans `groups`; une éventuelle référence de coffre est normalisée vers `proxy.group.<group_id>`. Si les identifiants de coffre existants ne peuvent pas être prouvés comme appartenant au groupe migré, le dry-run doit échouer de façon sûre et demander une réparation explicite, sans copier de secret.

## Règles de proxy et de runtime

Le cœur ne connaît pas de fournisseur commercial obligatoire. `proxy_providers.adapter_id` identifie un adaptateur générique, et `public_config_json` ne peut contenir que sa configuration non sensible. Ainsi, un adaptateur manuel, Decodo ou tout autre fournisseur peut être ajouté sans coupler le schéma ou le Core à une marque particulière.

Un profil peut soit ne définir aucun proxy direct et hériter de son groupe, soit définir ses propres champs proxy. Si un groupe est en mode `enforced`, il doit posséder un proxy complet. L’algorithme actuel d’héritage (`enforced` → proxy groupe ; sinon proxy profil puis proxy groupe) doit être conservé par le futur repository SQLite afin de ne pas modifier le comportement produit sans décision explicite.

Le statut de `runtime_candidates` ne constitue pas une approbation de release et ne modifie pas les gates BACK-01. En particulier, l’entrée d’un runtime candidat peut être créée afin de migrer des métadonnées, mais le Core doit refuser de lancer un runtime `quarantined` ou `retired`, et la politique produit devra décider explicitement le traitement de `candidate`.

## Index, intégrité et suppression

Les index couvrent la liste des profils par groupe/état/date d’usage, la sélection par runtime, la relation tags, l’historique des tests proxy, les opérations d’import et l’audit. La version 3 ajoute les index partiels `idx_profiles_proxy_provider_id_not_null`, `idx_profiles_proxy_secret_ref_not_empty`, `idx_groups_proxy_provider_id_not_null` et `idx_groups_proxy_secret_ref_not_empty` : ils accélèrent les recherches par fournisseur ou référence opaque de coffre sans indexer une valeur secrète. Le test de contrat `TestProxyReferenceLookupPlansUsePartialIndexes` impose par `EXPLAIN QUERY PLAN` que les quatre requêtes canoniques de recherche par fournisseur ou référence de coffre utilisent exactement ces index ; chaque requête inclut explicitement le prédicat de son index partiel. L’index partiel unique sur `container_id` ne s’applique qu’aux valeurs non vides : plusieurs profils non conteneurisés restent autorisés, tandis qu’un même conteneur ne peut pas être associé à deux profils.

Une suppression de runtime, fournisseur ou groupe référencé est refusée par SQLite. Le futur GUI doit proposer une transaction métier explicite : réaffecter ou archiver les profils concernés, vérifier les références de coffre, écrire un audit redacted, puis seulement supprimer l’entité devenue libre. Cette règle évite d’introduire des profils orphelins ou des mutations implicites.

## Plan de migration JSON → SQLite obligatoire

La migration de données est une fonctionnalité distincte de la seule migration DDL. L’importeur interne `internal/profilemigration` fonctionne en mode `dry-run` par défaut, exige en mode `apply` une sauvegarde pré-migration chiffrée et vérifiée, relit les lignes SQLite réellement écrites pour établir la parité, et effectue toutes les mutations métier dans une transaction unique. Aucune suppression de `profile.json` ou `groups.json` n’est autorisée lors de la première migration.

La fabrique `NewEncryptedPreMigrationBackup` construit une enveloppe JSON versionnée des sources validées, la chiffre dans un artefact BACK-01 `FLBK … FLEND`, vérifie cryptographiquement cet artefact avant toute écriture SQLite, puis ne transmet à l’importeur que l’identifiant de backup et son SHA-256. Le rapport d’import ne contient ni payload, ni clé, ni valeur de coffre. La fonction est actuellement une dépendance interne : son invocation doit être câblée au futur Core produit, jamais ajoutée à l’artefact RC BACK-01 gelé.

| Étape | Mode et transaction | Preuve attendue | Échec ou interruption |
|---|---|---|---|
| 1. Préflight | Lecture seule | Chemins, permissions, version SQLite et inventaire JSON assaini. | Arrêt sans écrire SQLite. |
| 2. Dry-run | Lecture source et journal minimal | Nombre de profils/groupes/tags, conflits, hash de chaque source et plan de mutations. | Deux entrées `validated` redacted, aucune table métier modifiée ni backup créé. |
| 3. Backup préimage (apply uniquement) | Chiffrement BACK-01 + vérification | ID et SHA-256 de l’artefact contrôlé couvrant `groups.json` et tous les `profile.json`. | Arrêt avant transaction SQLite si la source évolue, si le coffre échoue ou si la vérification échoue. |
| 4. Import | Transaction SQLite unique et journal `profile_import_operations` | Deux entrées `committed`, compteurs, audit redacted et hash de l’enregistrement canonique. | Rollback de toutes les écritures métier, y compris le journal et l’audit ; fichiers JSON intacts. |
| 5. Parité | Lecture des lignes SQLite dans la transaction | Ligne `profile_json_parity_checks` par profil, uniquement après égalité des hash canoniques. | Divergence = rollback total ; aucun `match` ne peut être écrit sur confiance. |
| 6. Reprise contrôlée | Nouveau dry-run après correction de la cause | Nouvelle corrélation, nouveaux hashes source et nouveau backup préimage. | Aucun « reprendre » implicite à partir d’un état partiel. |
| 7. Bascule contrôlée | Feature flag local explicite, à livrer ultérieurement | Chargement SQLite + fallback de lecture JSON documenté, sans double écriture non contrôlée. | Retour au lecteur JSON, données SQLite conservées pour diagnostic. |
| 8. Nettoyage ultérieur | Décision mainteneur séparée | Backup vérifié, parité durable et audit. | Aucun nettoyage automatique. |

La reprise après interruption ne réutilise jamais un état partiellement importé. Un `dry-run` journalise deux sources dans l’état `validated`; un `apply` réussi inscrit ces deux sources dans l’état `committed`. Une erreur de préflight n’écrit rien, tandis qu’une erreur dans l’import déclenche le rollback SQL de l’ensemble des lignes produit et du journal de l’opération. L’opérateur corrige la cause, exécute un nouveau `dry-run`, produit un nouveau backup préimage chiffré, puis relance l’`apply` avec une nouvelle corrélation.

## Couverture automatisée actuelle

Les tests exécutés sur la branche produit valident l’application conjointe BACK-01 + Produit, la migration d’une base déjà au niveau 1 vers le niveau 3, l’idempotence du chargeur, les contraintes runtime/groupe/proxy, les références étrangères, les quatre index de références proxy et l’absence de colonnes de secrets en clair. Ils couvrent aussi désormais l’import JSON : `dry-run` sans écriture métier, refus d’un `apply` sans backup vérifié, backup préimage BACK-01 chiffré, parité relue depuis SQLite, rollback lors d’un conflit d’intégrité, interruption après publication du préimage sans écritures produit partielles, reprise après correction, et refus de toute donnée proxy en clair avant journalisation ou sauvegarde.

```bash
go test ./internal/backup ./internal/productschema ./internal/profilemigration -count=1 -v
```

La bascule du lecteur métier JSON vers SQLite, son feature flag local et la désactivation éventuelle du lecteur JSON restent des travaux ultérieurs. Ils ne doivent être entrepris qu’après revue de ce flux, y compris sur une base existante représentative, et sans toucher au candidat BACK-01 gelé.
