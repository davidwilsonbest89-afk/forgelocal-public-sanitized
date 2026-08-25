# T28 — Modèle de menace des extensions locales contrôlées

## 1. Actifs à protéger

Les actifs sont le Profile Store canonique, les métadonnées SQLite T28, les blobs ZIP immuables, les IDs ForgeLocal, les relations d’affectation, les événements d’audit et les secrets d’authentification du Core. T28 ne doit pas transformer un package d’extension en identité navigateur ni exposer son contenu.

## 2. Adversaires et scénarios

| Menace | Exemple | Contrôle requis |
|---|---|---|
| Archive traversante | nom `../../profile.json` ou chemin absolu | normalisation et refus avant extraction |
| Symlink/hardlink ZIP | entrée Unix pointant vers un fichier local | détection des attributs et refus |
| Zip bomb | taille compressée faible, expansion excessive | limites taille entrée/cumul/entrées |
| Manifest hostile | JSON invalide, valeurs énormes, doublons | parse borné, normalisation, refus déterministe |
| Package non autorisé | extension sensible ou inconnue | conservation + `HIGH_RISK`, approbation explicite obligatoire |
| Fuite de confidentialité | ZIP, chemin, manifest complet, token dans log/API | redaction centralisée et projections minimales |
| Mutation concurrente | deux approvals, assignments ou rollbacks simultanés | verrou par série/profil et transaction fail-closed |
| État partiel | SQLite échoue après copie ou stockage échoue pendant import | compensation sans publication d’une ligne incohérente |
| Révocation contournée | assign après quarantine/revoke | précondition d’état dans la même transaction |
| Purge dangereuse | supprimer un blob cible de rollback ou affecté | garde de référence et purge explicite auditée |
| Runtime indirect | handler qui appelle launcher/proxy/processus | dépendances interdites, garde structurelle et tests |
| Accès API illégitime | bearer absent, origine étrangère ou adresse non loopback | middleware existant, refus 401/403 avant mutation |
| Confusion d’identité | manifest `name`, `key` ou version présenté comme ID runtime | IDs ForgeLocal générés par serveur et texte contractuel explicite |

## 3. Limites de confiance

Le client local authentifié est autorisé à proposer un fichier, mais le ZIP et son manifest sont des données non fiables. Le fichier ne doit être traité qu’en lecture bornée. Le Profile Store est une dépendance de lecture permettant de vérifier qu’un profil existe ; il reste hors migration. SQLite T28 est la source des états T28, mais ne peut publier une affectation avant les préconditions.

Aucun contenu du package n’est interprété comme code. Aucun chemin issu du client n’est réutilisé dans l’audit ou une réponse complète. Un digest tronqué suffit à corréler les événements redacted ; le digest complet reste interne au repository.

## 4. Risques liés aux permissions

Les permissions sensibles, host patterns larges et valeurs inconnues ne sont pas refusés à l’import afin de respecter la décision produit. Elles sont conservées exactement après normalisation et produisent `HIGH_RISK` ou `UNCLASSIFIED_HIGH_RISK`. L’approbation exige la confirmation exacte de la liste visible et l’acceptation high-risk lorsque nécessaire. Une affectation séparée garantit qu’un import ou une approbation ne modifie pas implicitement un profil.

## 5. Confidentialité et redaction

Les événements d’audit peuvent contenir UTC, action, IDs ForgeLocal, digest tronqué, catégories, profil pseudonymisé, résultat et code stable. Ils ne contiennent jamais le ZIP, les bytes, le manifest complet, un chemin utilisateur complet, token, bearer, cookie, header libre, presse-papiers ou données navigateur. Les erreurs API utilisent des codes stables et des messages génériques ; les détails de filesystem restent dans un journal développeur contrôlé seulement s’ils sont eux-mêmes redacted.

## 6. Concurrence et reprise

Une série est une unité de sérialisation pour import/update/rollback/revoke ; un profil est une unité de sérialisation pour assign/unassign. Une mutation prend le verrou logique avant la transaction SQLite et vérifie l’état après acquisition. Toute erreur débouche sur rollback logique et réponse d’erreur ; aucune transition partielle n’est renvoyée. Un restart relit SQLite et les blobs immuables sans modifier les états.

Lorsqu’un blob est copié avant un échec SQLite, il ne doit pas être supprimé aveuglément : il est conservé comme objet non référencé et signalé pour une purge explicite ultérieure. Le repository ne doit jamais considérer un blob référencé comme orphelin. Une erreur de copie ne crée aucune ligne publiée.

## 7. Preuve de non-exécution

Les packages ne sont jamais extraits intégralement pour exécution. Un éventuel champ `update_url` ou `updateURL` du manifest est ignoré, non suivi et non exécuté ; il ne constitue ni une source ni une instruction. Le package T28 ne doit importer aucun module de lancement navigateur, Camoufox, proxy ou processus externe. Les handlers appellent uniquement le parser, le repository, le Profile Store de vérification et l’audit redacted. Les tests doivent rechercher et instrumenter les frontières de lancement pour démontrer qu’aucun scénario T28 ne les atteint.

## 8. Critères d’échec

Le lot échoue si un ZIP dangereux est accepté, si une permission est perdue, si l’approbation contourne un acknowledgement, si une affectation existe avant approval ou après révocation, si deux versions deviennent actives, si un rollback remplace un blob, si une fuite apparaît dans API/audit/preuves, si une mutation concurrente publie un état partiel, si un profil absent est accepté, ou si un chemin T28 atteint un runtime navigateur/processus.
