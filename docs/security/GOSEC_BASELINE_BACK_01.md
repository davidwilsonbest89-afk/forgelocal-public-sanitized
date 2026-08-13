# Classification Gosec — baseline ForgeLocal / BACK-01

**État :** actif à partir du commit `97465bf` (à compléter par le commit de classification).  
**Portée :** branche `forgelocal-back01`, scan `gosec ./...` exécuté avec Go `1.25.13`.

## Règle de non-régression

> Toute modification introduisant une alerte Gosec dans `internal/backup` ou `internal/secrets` bloque le merge. La baseline BACK-01 est de **zéro alerte**.

Le scan ciblé du socle BACK-01 a trouvé **0 alerte** après correction des bornes de longueur, des chemins internes et de l’usage du nonce AES-GCM. Les `#nosec` restants sont documentés dans le code et limités à cinq cas dont les préconditions sont contrôlées.

## Répartition de la dette globale

| Classe | Nombre | Règles | Décision |
|---|---:|---|---|
| **Nouvelle alerte BACK-01** | 0 | — | Bloquante : toute nouvelle alerte est refusée. |
| **P0 — à corriger avant exposition réseau, import ZIP ou activation de fournisseurs externes** | 36 | G107 (1), G110 (3), G120 (2), G122 (1), G204 (6), G703 (15), G704 (7), G705 (1) | Traiter avant toute API non strictement locale, tout import non chiffré ou tout test proxy/navigation externe. |
| **Dette historique P1 — à corriger avant publication d’une beta distribuable** | 98 | G112 (1), G115 (4), G301 (35), G302 (4), G304 (24), G305 (3), G306 (10), G404 (17) | Réduire avec des tickets ciblés ; aucune nouvelle occurrence hors baseline. |
| **Dette reportée P2 — qualité et gestion d’erreurs** | 68 | G104 (68) | Journaliser chaque erreur ignorée ou justifier une suppression volontaire ; échéance avant RC. |
| **Total baseline** | **202** | 18 règles | Aucun élément de cette baseline ne vaut acceptation de BACK-01. |

## Justification de priorité

| Règle | Risque | Classe | Action exigée |
|---|---|---|---|
| G703 / G304 / G305 / G122 | Traversée de chemin, archive hostile ou course sur système de fichiers | P0 pour les flux atteignables ; P1 sinon | Normaliser, borner et résoudre les chemins sous une racine contrôlée ; remplacer les extractions ZIP existantes avant de les exposer. |
| G704 / G107 | SSRF ou URL contrôlée par l’entrée utilisateur | P0 | Bloquer les IP locales et réservées, imposer allowlist et validation DNS/redirects avant toute sortie réseau. |
| G110 / G120 | Décompression ou multipart non borné | P0 | Utiliser limites de taille, nombre d’entrées, ratio de compression et streaming borné. |
| G204 | Exécution de sous-processus avec argument dynamique | P0 | Arguments structurés, allowlist de runtime, jamais de shell ; tests d’injection. |
| G705 | XSS dans dashboard local | P0 | Échappement contextualisé, CSP stricte et interdiction de rendu HTML non fiable. |
| G112 / G115 | Épuisement de connexion ou conversion entière risquée | P1 | Délais serveur et contrôles de bornes avant conversion. |
| G301 / G302 / G306 | Permissions trop ouvertes | P1 | Répertoires `0700`, fichiers et backups `0600`, vérification via tests. |
| G404 | Aléa faible | P1 | `crypto/rand` pour identifiants, tokens et secrets ; autorisation écrite uniquement pour temporisations UX. |
| G104 | Erreurs ignorées | P2 reporté | Corriger ou annoter chaque occurrence avec motif, test et date d’expiration. |

## Gouvernance

Chaque ticket de correction doit référencer la règle, le fichier, le scénario d’attaque, un test de régression et un propriétaire. Les exceptions `#nosec` doivent nommer la règle, décrire la précondition vérifiée et être revues à chaque modification de la zone concernée. Aucun `#nosec` général, aucune exclusion globale de règle et aucune baisse de sévérité ne sont autorisés pour rendre le pipeline vert.

La reconnaissance de cette baseline est limitée au commit de référence et doit être recalculée sur chaque mise à jour majeure de Gosec, du toolchain ou des dépendances.
