# Runbook de classification mainteneur — alerte de secret potentielle

**Statut du candidat au départ :** `PILOT_LOCAL_SUSPENDED_PENDING_SECRET_SCAN_TRIAGE`  
**Décision publique au départ :** `PUBLIC_RELEASE_BLOCKED`  
**Principe :** aucun acteur ne doit afficher, copier, décoder, transmettre ou tester la valeur détectée dans un service externe.

> Ce runbook ne constitue pas une autorisation de reprendre le pilote. Seule une décision versionnée, complétée par une revue indépendante, peut clôturer le triage.

## Préconditions communes

Le mainteneur travaille hors journal public avec le propriétaire autorisé de la preuve de provenance. Le candidat RC actuel, le runtime, les gates, le manifeste et le SBOM restent gelés. Les seuls éléments manipulés dans le dépôt sont des documents redacted ; la valeur brute n’y est jamais introduite.

| Contrôle préalable | Critère |
|---|---|
| Identité de l’archive | SHA-256 égal à `553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6` |
| Rapport de scan | Hash égal à `b9220e555155c055c40b3f8933d71fbbb4a780f6084ad8fcc129e681ac9e30ce` |
| Pilote | Toujours `PILOT_LOCAL_SUSPENDED_PENDING_SECRET_SCAN_TRIAGE` |
| Publication | Toujours `PUBLIC_RELEASE_BLOCKED` et cinq gates `PENDING` |
| Configuration de scan | Aucune exclusion globale ou fondée sur le seul chemin ; l’allowlist existante est limitée à une empreinte PPA publique exacte |

## Branche A — `REAL_SECRET`

La valeur est un `REAL_SECRET` si une source autoritative ou son propriétaire confirme qu’elle peut authentifier un accès, signer avec une clé privée, autoriser une API, protéger une ressource, ou être échangée contre un accès sensible. Dès cette conclusion, elle est traitée comme **potentiellement exposée** : une révocation ou rotation confirmée est obligatoire avant toute reconstruction.

| Étape | Exigence de clôture |
|---|---|
| 1. Confinement | Conserver le pilote suspendu et empêcher toute nouvelle distribution de l’archive. |
| 2. Révocation | Le propriétaire révoque, remplace ou désactive la valeur dans son système d’origine. Une confirmation redacted ou un identifiant de ticket doit être obtenu **avant toute reconstruction**. |
| 3. Portée | Examiner le Git et les artefacts dérivés avec des résultats de comptage et des chemins redacted, sans imprimer la valeur. |
| 4. Invalidation | Déclarer définitivement non publiable le RC actuel et ne jamais tenter de le corriger en place. |
| 5. Nouvelle chaîne | Construire un nouveau candidat à partir d’une preuve assainie, avec un nouveau SHA-256, SBOM, manifeste, index de traçabilité et contrôle de licences. |
| 6. Réexécution | Refaire les preuves E2E dépendantes et tous les gates applicables ; le test SystemVault ne reprend que sur ce nouveau candidat. |
| 7. Revue | Obtenir une revue indépendante de la révocation, de l’assainissement et des scans du nouveau candidat. |

## Branche B — `FALSE_POSITIVE`

Une conclusion `FALSE_POSITIVE` exige une preuve mainteneur indépendante que l’occurrence n’authentifie aucun accès. La décision redacted doit identifier le propriétaire, le rôle de la donnée de provenance, la source autoritative et une affirmation explicite d’absence de capacité d’authentification.

| Étape | Exigence de clôture |
|---|---|
| 1. Justification | Documenter la nature non secrète sans valeur brute, ainsi que l’autorité qui le confirme. |
| 2. Forme de preuve | Remplacer dans la source de provenance la ligne non structurée par une forme explicitement non secrète ou redacted. Une valeur qui déclenche à nouveau le scanner ne doit pas être conservée sous forme opaque. |
| 3. Candidat neuf | Ne jamais modifier l’archive RC gelée. Reconstruire une nouvelle archive avec un nouvel identifiant, manifeste, SBOM, checksums et index de traçabilité. |
| 4. Rescan | Un relecteur indépendant scanne la preuve redacted et l’archive reconstruite. Aucun motif équivalent ne doit être détecté et aucune exception Gitleaks globale, par chemin, large ou destinée à masquer le motif équivalent ne doit être ajoutée. |
| 5. Revue | Le relecteur indépendant atteste la justification, la transformation de preuve, le rescan propre et l’absence de nouvelle exception large. |
| 6. Suite | Reprendre seulement les gates rendus invalides par le nouvel artefact ou la nouvelle preuve. Aucun statut public ne devient approuvé par cette seule classification. |

## Branche C — `UNKNOWN`

Si le propriétaire, la source autoritative ou la revue indépendante sont indisponibles, la conclusion reste `UNKNOWN`.

| Règle | Effet |
|---|---|
| Pilote | Reste suspendu. |
| Candidat RC | Reste gelé, non distribué et non publiable. |
| SystemVault | Reste reporté pour ce candidat afin de ne pas investir dans une chaîne potentiellement invalide. |
| Publication publique | Reste `PUBLIC_RELEASE_BLOCKED`. |
| Reprise du triage | Possible uniquement lorsque les informations mainteneur nécessaires sont disponibles. |

## Dossier de clôture obligatoire

Le mainteneur complète `SECRET_SCAN_CLASSIFICATION_DOSSIER.template.json` dans un nouveau fichier de décision versionné. Il y associe une identité responsable, un horodatage UTC, une classification, une justification redacted, les hashes de l’archive et du rapport, un résultat de contrôle de l’historique, un rescan reproduit et une revue indépendante. Aucun champ ne doit contenir la valeur brute.

La mise à jour de l’autorisation de pilote est interdite avant que le dossier de décision soit complet et que la revue indépendante ait explicitement approuvé la conclusion. Une conclusion `REAL_SECRET` exige en plus une preuve redacted de révocation ou de rotation antérieure à la reconstruction ; une conclusion `FALSE_POSITIVE` exige un rescan indépendant propre de la nouvelle preuve et l’absence vérifiée d’exception Gitleaks large.
