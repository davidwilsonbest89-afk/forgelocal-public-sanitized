# CHANGELOG — T10-T15-INDEPENDENT-REVIEW

## 2026-08-25

Revue physique initiale : les artefacts, assertions, cleanup et Gitleaks étaient verts, mais les sidecars copiés en répertoire neutre conservaient le chemin de dépôt et la plage Git n’était pas matérialisée dans le clone. La revue n’a pas été qualifiée à cette étape.

Après correction, les sidecars ont été normalisés pour le répertoire neutre, le tag V6 a été fetché, la plage Git non vide a été vérifiée, puis Gitleaks a retourné zéro. Les contrôles ZIP/bundle, extraction, manifeste, checksums, clone, checkout exact, fsck, assertions et cleanup ont tous retourné zéro. Décision : `T10_T15_SYNTHETIC_E2E_APPROVED_VERIFIABLE_LOCAL`.
