# ForgeLocal BACK-01 — Contrôles finaux de traçabilité et de gate public

**Décision contrôlée :** `PUBLIC_RELEASE_BLOCKED`.

Cette synthèse complète la qualification du candidat RC Chromium `151.0.7922.108`. Elle ne transforme ni le pilote historique ni le candidat RC en publication publique.

## Contrôles exécutés

| Contrôle | Méthode | Résultat |
|---|---|---|
| Reproductibilité de la traçabilité | Clone Git propre au commit `03c1fb4c875cc30b47bac15104f0d029b8cc9e93`, copie explicite des deux archives référencées et du manifeste externe RC, puis exécution du validateur | Vert ; aucun binaire livré n’a été exécuté |
| Signature mainteneur | Signature et vérification dans deux trousseaux temporaires, export de clé publique uniquement vers le vérificateur | Vert ; aucune clé privée ni signature de test n’est restée dans le dépôt |
| Gate public machine-readable | Contrôle de l’index de traçabilité, de la chaîne RC, des documents obligatoires et des gates | Vert ; `PUBLIC_RELEASE_BLOCKED` confirmé |

> Le test de signature utilise une clé de démonstration éphémère supprimée en fin d’exécution. Il vérifie le mécanisme ; il ne constitue pas une signature mainteneur de release.

## Gates obligatoires non levés

| Gate | État | Condition de clôture |
|---|---|---|
| `SYSTEMVAULT_NATIVE_PER_TARGET` | `PENDING_NATIVE_HOST` | Matrice native verte, distincte pour chaque OS, version et architecture annoncés, sans `sudo`, conteneur ni fallback clair |
| `SYSTEMVAULT_ANTI_LEAK_INTEGRATED_FLOW` | `PENDING_INTEGRATED_E2E` | Preuve anti-fuite issue d’un flux réel profil → backup chiffré → restauration isolée, sans sentinelle affichée |
| `MAINTAINER_MANIFEST_SIGNATURE` | `PENDING_MAINTAINER_SIGNATURE` | Signature détachée, clé publique publiée séparément, empreinte approuvée et vérification indépendante |
| `RUNTIME_LICENSE_AND_REDISTRIBUTION_REVIEW` | `PENDING_REVIEW` | Revue documentée de licence et de redistribution des paquets runtime exacts |
| `OS_COMPATIBILITY_EVIDENCE` | `PENDING_PER_TARGET_EVIDENCE` | Matrice limitée aux cibles possédant les preuves runtime et SystemVault complètes |

## Règle de promotion

Une promotion vers `PUBLIC_RELEASE_APPROVED` exige que les cinq gates soient explicitement marqués `PASSED`, chacun avec sa preuve versionnée et une revue indépendante. Une signature réussie ou un E2E Chromium vert, pris isolément, ne suffit pas.
