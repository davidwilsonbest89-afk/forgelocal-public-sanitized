# GOSEC-R5 — Threat model

## Portée

Ce threat model couvre les 63 findings Gosec observés sur le périmètre source-only `./cmd/... ./internal/...` au commit de baseline `079c452444b6bc55d885afafdd845a72bffd7ab4`. Il ne constitue pas une analyse de production et ne couvre pas les environnements absents ou explicitement bloqués.

## Actifs protégés

Les actifs principaux sont les profils et leurs données, les tokens et métadonnées d’authentification, les groupes de proxies, les backups et archives, les fichiers runtime, les logs et preuves d’audit, les processus locaux, les endpoints HTTP/WebSocket loopback et l’intégrité des transitions d’état.

| Menace | Règles concernées | Précondition | Impact possible | Contrôles R4 existants | Preuve R5 attendue |
|---|---|---|---|---|---|
| Lecture ou écriture hors racine | G703, G304, G305 | Entrée de chemin contrôlable, symlink, hardlink ou course | Confidentialité/intégrité des backups, profils ou artefacts | Validation lexicale, `os.Root`, refus symlinks, limites archives | Tests positifs/négatifs, TOCTOU si réalisable, extraction fraîche |
| Exécution ou dial non contrôlé | G204, G704 | Entrée URL/commande atteignable | Exécution arbitraire, SSRF, disponibilité | Allowlist de commandes, loopback, timeout, redirections refusées | Tests d’arguments, refus externe, arrêt/cleanup |
| Permission trop large | G302 | Fichier/dossier runtime créé ou restauré | Exposition locale de données sensibles | Modes 0700/0600; exécutables 0755 sous racine privée | Mode effectif, non-substitution, restauration |
| Conversion numérique dangereuse | G115 | Taille, index, port, seed ou compteur hors borne | Panique, troncation, corruption, bypass de limite | Bornes explicites sur chemins déjà durcis | Tests min/max, négatif, overflow |
| Aléatoire inadapté | G404 | Valeur aléatoire utilisée comme contrôle de sécurité | Prévisibilité si usage auth; bruit comportemental sinon | Séparation humanisation/fingerprint vs secrets | Démonstration d’absence d’usage sécurité |
| Motif secret détecté | G101 | Scanner confond nom/fixture et credential | Fausse assurance ou fuite réelle | Redaction et stockage de hash/métadonnées | Inspection code, fixtures, logs, manifests, bundles |

## Principes de décision

Un finding ne peut être classé `HISTORICAL_NOT_REACHABLE` que si une garde ou un chemin non atteignable est démontré par le code et un test négatif. Une mitigation applicative ne supprime pas automatiquement un finding statique : elle est classée `MITIGATED_CONTROL_SCANNER_OPEN` jusqu’à une preuve suffisante et une revue explicite.

Les lots R5-A, R5-B et R5-C sont indépendants. Chaque lot doit conserver son baseline, ses tests ciblés, son scan post-correctif, sa matrice mise à jour et ses artefacts. Une fuite de secret, une sortie de confinement, une exécution arbitraire, un bypass d’authentification, un accès externe non contrôlé, un mélange de profils ou une mutation partielle impose l’arrêt `GOSEC_R5_BLOCKED_CRITICAL_FINDING`.

## Limites environnementales

Docker/Buildx, Camoufox natif, Darwin/xattr/GUI et SystemVault natif ne sont pas simulés sous Linux. Le proxy réel, les cookies réels, les comptes externes et les données utilisateur sont exclus. Ces éléments restent `BLOCKED_ENVIRONMENT_REQUIRED` ou `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` selon la preuve de disponibilité.
