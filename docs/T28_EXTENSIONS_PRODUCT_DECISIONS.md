# T28 — Décisions produit des extensions locales contrôlées

**Autorisation appliquée :** T28 uniquement.  
**Décision de sécurité structurante :** toutes les permissions sont importables et autorisables, mais aucune n’est tacitement accordée et aucune extension n’est téléchargée, chargée ou exécutée.

## Décisions fermées

| Sujet | Décision T28 |
|---|---|
| Source | ZIP local fourni explicitement par le client Core ; aucune URL applicative, store ou mise à jour Internet. |
| Format | ZIP avec `manifest.json` à la racine ; CRX, unpacked directory et archive distante refusés. |
| Provenance | Identité locale par SHA-256 du fichier, format, taille et manifest extrait ; aucune prétention d’identité navigateur. |
| Série/version | IDs ForgeLocal générés par le serveur ; version immuable dans une série ; une seule version active. |
| Permissions | `permissions`, `optional_permissions`, `host_permissions`, `optional_host_permissions` et `matches` conservés après normalisation. |
| Autorisation | Toutes les catégories, y compris sensibles et inconnues, sont acceptées à l’import ; les valeurs sensibles ou larges sont `HIGH_RISK`; les inconnues `UNCLASSIFIED_HIGH_RISK`. |
| Approbation | Acknowledgement exact de la liste normalisée et `accept_high_risk=true` lorsqu’un risque est présent. |
| Affectation | Opération distincte, uniquement vers un profil existant et depuis une version approuvée. |
| Stockage | Blob ZIP immuable sous `--base-dir`, arborescence dérivée du digest ; métadonnées et états en SQLite dédié. |
| Profile Store | `profile.json` canonique conservé ; aucune migration vers SQLite ; GAP-002 reste ouvert. |
| Lifecycle | importé → approuvé → affecté/prêt ; update crée un blob distinct ; rollback vers version approuvée ; revoke/quarantine interdit les nouvelles affectations ; purge explicite et auditée. |
| Rollback | Repointer logiquement la série vers une version approuvée antérieure, sans remplacer ni réécrire les blobs. |
| Refus | Refuser archive dangereuse, absence/invalidité de manifest, corruption, zip-slip, symlink, limites dépassées, auth/origine/loopback invalides, état ou profil incompatibles. |
| Concurrence | Sérialisation par série et profil ; mutations fail-closed ; aucune double affectation. |
| Audit | UTC, action, IDs ForgeLocal, digest tronqué, catégories, profil pseudonymisé, résultat et code stable ; redaction stricte. |
| Runtime | Interdit dans T28 : navigateur, Chromium, Camoufox, proxy, moteur d’extension, processus externe, store et réseau applicatif. |

## Interprétation de `ready`

`ready` signifie seulement que la configuration locale a satisfait les validations T28. Ce mot ne signifie ni que le navigateur reconnaît l’extension, ni qu’un manifest fournit un ID runtime stable, ni que l’extension est compatible avec un navigateur, ni qu’elle sera lancée.

## Autorisations non accordées automatiquement

La liste exhaustive de permissions autorisables constitue une capacité de conservation et d’approbation, pas une allowlist d’exécution. Chaque approbation doit être précédée d’une vue de la liste normalisée et d’un acknowledgement transmis explicitement. Une omission ou une divergence d’un seul élément est bloquante pour l’approbation.

## Hors décision T28

Les décisions T29 Password Manager, T39 import/export de secrets, T40 API finale, T41 Dashboard et T42 release ne sont pas prises ni implémentées par ce lot.
