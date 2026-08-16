# T07-R — Guide de complétion pour le détenteur des droits

Ce guide vous permet de compléter l’attestation redacted sans être développeur. Vous êtes identifié dans le brouillon comme `@boucheriechefimane-cmd`, conformément à votre déclaration de créateur et détenteur des droits de Camoflox. Les éléments manquants sont volontairement laissés incomplets : ils ne doivent jamais être devinés.

> Ne copiez pas le code Camoflox, l’archive, une clé, un token, une valeur trouvée dans `tests/smoke.test.js:24`, une licence privée complète ou une capture de votre espace privé dans le JSON.

## Les cinq décisions ou références à fournir

| Champ à compléter | Ce que vous indiquez | Format attendu | Exemple de forme redacted |
|---|---|---|---|
| `private_repository_or_attestation_id` | L’identifiant non secret de l’emplacement que vous contrôlez | Identifiant stable, pas un mot de passe | `private-git:owner/camoflox-history` |
| `revision_identifier` et `snapshot_sha256` | La version privée exacte liée au snapshot étudié et son hash SHA-256 | Identifiant + 64 caractères hexadécimaux | `release-v28-stage2` + `a1b2…` |
| `redistribution` | Votre décision explicite | Exactement `granted` ou `not_granted` | `not_granted` |
| Références de licence, notices et obligations tierces | Références consultables par une relectrice autorisée | Référence redacted stable, pas le document brut | `private-license-record:CAMO-LIC-01` |
| `maintainer_decision` | Votre classification de l’alerte sans révéler sa valeur | `REAL_SECRET`, `FALSE_POSITIVE` ou `UNKNOWN` | `UNKNOWN` tant qu’aucune conclusion n’est établie |

Pour le hash du snapshot privé, vous pouvez calculer localement `sha256sum <nom-du-fichier>`. Ne transmettez que le résultat de 64 caractères et l’identifiant du snapshot, jamais le fichier lui-même dans ce paquet.

## Règles de remplissage essentielles

La valeur de `revision_kind` est déjà fixée à `immutable_snapshot`, car le candidat étudié est une archive. Les deux droits de travail sont déjà déclarés `yes` : `internal_use` et `modification`. Vous devez toutefois renseigner explicitement la redistribution ; elle ne peut jamais être supposée.

Si vous estimez que l’alerte est `REAL_SECRET`, ne déclarez pas de conclusion de déblocage. Il faudra fournir des références redacted de révocation ou rotation, un nouveau snapshot candidat, son hash et un re-scan. Si vous l’estimez `FALSE_POSITIVE`, un nouveau snapshot redacted, son hash et un re-scan restent obligatoires. Si vous n’êtes pas certain, choisissez `UNKNOWN` ; c’est le comportement sûr et T07 demeure bloqué.

Une fois vos champs renseignés, remettez à la relectrice une copie redacted du JSON avec le message prêt à envoyer. Elle seule doit positionner les six booléens de couverture à `true` après vérification réelle. Tant que ces booléens sont à `false`, l’attestation est intentionnellement incomplète.
