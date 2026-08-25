# V6 — matrice individuelle Gitleaks

La matrice CSV contient **348 lignes individuelles** correspondant aux **58 arbres** et aux **348 détections** de la campagne V5. Les valeurs `Match`/`Secret` ne sont jamais recopiées. Chaque ligne conserve le chemin, le commit, le type, la règle, la ligne, une empreinte redacted de localisation et un hash de contenu non secret.

| Catégorie | Occurrences | Décision |
|---|---:|---|
| Finding actuel au HEAD V4, contenu identique au fichier courant | 6 | Faux positif confirmé : empreinte publique de clé de signature PPA ; exception exacte documentée, sans incident secret |
| Findings historiques, contenu identique au fichier courant | 342 | Faux positifs historiques confirmés ; histoire conservée, aucune réécriture |
| Contenu divergent nécessitant arrêt et incident | 0 | Aucun dans cette comparaison |

Les six chemins uniques sont les artefacts de provenance du runtime candidat et du verrou de release sous `release/back01-minimal/` et `validation_back01_integration/`. Les six occurrences du checkout courant sont les mêmes valeurs de fingerprint public que les versions historiques. La configuration existante est une exception exacte par fingerprint ; le contrôle V6 doit la resserrer aux chemins de provenance si le parseur de configuration de l’outil le permet, puis conserver le finding non classifié en cas d’échec de ce resserrement.

Le résultat **n’est pas un PASS de scan** : il s’agit d’une qualification documentée de faux positifs publics, à confirmer par la revue indépendante.
