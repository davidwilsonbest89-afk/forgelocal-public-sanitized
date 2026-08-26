# GOSEC-R4 — threat model initial

## Périmètre

Le modèle couvre les 99 findings Gosec source-only observés sur la branche dédiée `validation/gosec-r4`, issue du HEAD R3 découvert dynamiquement. Il ne couvre pas les comptes réels, secrets réels, proxies externes, données utilisateur, runtime de production, release, Camoufox natif, SystemVault natif ou Docker/Buildx actif.

> Le scan R3 avait été résumé précédemment comme 94 findings, mais le JSON Gosec publié et rejoué contient réellement 99 lignes. R4 conserve ce comptage observé et documente la divergence plutôt que de la masquer.

## Actifs

| Actif | Valeur | Exposition principale |
|---|---|---|
| Profils, cookies et métadonnées locales | Confidentialité et intégrité | fichiers, archives, restore, permissions, artifacts |
| Backups et preuves d’audit | Intégrité, disponibilité et traçabilité | I/O partiel, fermeture, rollback, chemins et archives |
| Core loopback et ponts locaux | Intégrité et disponibilité | HTTP, WebSocket, redirections, timeouts |
| Runtime navigateur et exécutables | Intégrité et disponibilité | subprocess, extraction, remplacement, permissions |
| Tokens synthétiques et métadonnées admin | Confidentialité et authentification | redaction, logs, motifs G101 |
| Matériel de preuve GitHub | Traçabilité | commits, bundles, sidecars, manifests, LFS |

## Menaces prioritaires

| Menace | Précondition | Impact | Contrôles déjà présents | Preuve R4 exigée |
|---|---|---|---|---|
| Lecture ou écriture hors racine | chemin contrôlé par l’utilisateur, symlink, séparateurs Windows ou race | divulgation ou corruption de profils/backups | validation loopback/path, `os.Root` sur chemins corrigés, staging | tests de traversal, symlink, hardlink, TOCTOU et racine inaccessible |
| Archive malformée | entrée spéciale, taille/profondeur excessive ou restore interrompu | exécution, corruption, DoS ou perte de rollback | limites d’entrées/octets/profondeur, rejet types spéciaux, staging/rollback | ZIP/TAR positifs et négatifs, corruption, interruption et cleanup |
| I/O partiel non signalé | `Write`, `Close`, `Rename`, `Sync`, transaction ou audit en erreur | état partiel ou faux succès | contrôles présents selon chemin | writers/closers fault-injectés, permissions refusées, audit indisponible |
| Subprocess ou réseau détourné | argument/URL tainted, redirection, timeout absent | exécution externe, SSRF, processus orphelin | validation loopback, deadlines, allowlists de chemins | tests loopback/externe, redirect, expiré/révoqué et processus résiduel |
| Permission excessive | dossier ou fichier créé avec mode trop large | lecture/remplacement par autre utilisateur | modes 0700/0600 sur nombreux chemins R3 | assertions de mode et tentative d’accès hors propriétaire |
| Faible aléatoire utilisé à mauvais escient | `math/rand` réutilisé pour token, clé, auth ou décision sécurité | prédictibilité d’un secret ou bypass | revue initiale : usages humanisation/fingerprint | recherche d’appels et test de séparation crypto/non-crypto |
| Secret embarqué ou journalisé | token réel dans code, fixture, archive, rapport ou log | divulgation d’identifiants | fixtures synthétiques, redaction, Gitleaks | scan redacted et inspection ciblée G101/G107 |
| Perte d’élément de preuve | manifeste auto-référentiel, mauvais CWD, LFS smudge ou bundle incomplet | impossibilité de reproduire ou audit trompeur | manifestes non auto-référentiels, sidecars portables, clone neuf | extraction, hashes, bundle verify, fsck et journal des défauts |

## Critères d’arrêt immédiat

La campagne doit s’arrêter et produire `GOSEC_R4_BLOCKED_CRITICAL_FINDING` si elle découvre une fuite de secret, une corruption non réversible, un bypass d’authentification, une sortie de confinement, une exécution arbitraire, une connexion externe non contrôlée ou un mélange de profils.

Un contrôle applicatif ne clôt jamais à lui seul une ligne Gosec : il doit être accompagné d’un test positif et négatif, d’une explication du chemin exécuté et d’une preuve versionnée. Les outils absents et les plateformes absentes restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` ou `BLOCKED_ENVIRONMENT_REQUIRED`; ils ne sont pas assimilés à un PASS.

## Référence opérationnelle

La matrice détaillée par finding se trouve dans `GOSEC_R4_BASELINE_MATRIX.tsv`; le scan brut source-only est `GOSEC_R4_BASELINE_RAW.log`. Les statuts autorisés sont `CORRECTED_AND_VERIFIED`, `MITIGATED_CONTROL_SCANNER_OPEN`, `NEEDS_MANUAL_REVIEW`, `HISTORICAL_NOT_REACHABLE` et `BLOCKED_ENVIRONMENT_REQUIRED`.
