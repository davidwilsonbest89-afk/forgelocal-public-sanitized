# GOSEC-R6 — threat model de triage exploitable

## Baseline autoritative

La baseline R6 corrigée contient **59 findings** et reprend exactement le JSON final R5 dont le SHA-256 est `f24905f47bee87164511232caf0538d551e2199460678775390f789b7a2c3e9f`. La matrice `R6_FINDING_PRIORITY_MATRIX.tsv` contient une ligne par finding, y compris **G305=1**, qui avait été omis dans le texte du mandat initial.

| Règle | Count | Priorité par défaut | Phase |
|---|---:|---|---|
| G304 | 11 | P0 | paths/confinement |
| G703 | 9 | P0 | paths/confinement |
| G305 | 1 | P0 | paths/confinement |
| G204 | 5 | P1 | subprocess/réseau |
| G704 | 7 | P1 | subprocess/réseau |
| G302 | 5 | P2 | permissions/bornes |
| G115 | 3 | P2 | permissions/bornes |
| G404 | 17 | P2 | aléatoire |
| G101 | 1 | P2 | secrets |
| **Total** | **59** |  |  |

## Modèle de menace

R6 évalue chaque finding selon la source de l’entrée, le chemin de données réellement exécuté, la racine attendue, la possibilité de traversal, le traitement des symlinks/hardlinks/fichiers spéciaux, les courses TOCTOU, le rollback, les permissions effectives, l’exposition réseau, les arguments subprocess, les secrets et l’impact confidentialité/intégrité/disponibilité.

Les chemins P0 sont traités en premier. Une alerte Gosec qui persiste malgré un garde applicatif est classée `MITIGATED_CONTROL_SCANNER_OPEN` uniquement avec test négatif, preuve de l’appelant et limites documentées. Elle n’est pas déclarée close par convention. Un chemin réellement non atteignable doit être démontré par le graphe d’appel et un test ou une preuve reproductible.

Les chemins Linux sont exécutés uniquement sur des fixtures synthétiques loopback. Les chemins Windows/macOS restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` sans simulation. Aucun compte réel, cookie réel, secret réel, proxy commercial, site externe, runtime de production, Camoufox natif ou SystemVault natif n’est utilisé.

Les tests de non-régression T28 sont strictement ceux déjà existants; aucune modification ou réouverture fonctionnelle de T28 n’est autorisée. T29 reste interdit, et T31–T38 restent intacts.

## Gates

Les outils absents restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`. Les statuts de campagne attendus sont `GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`. En présence d’un problème critique réellement démontré, le statut devient `GOSEC_R6_BLOCKED_CRITICAL_FINDING` sans masquer les autres findings.
