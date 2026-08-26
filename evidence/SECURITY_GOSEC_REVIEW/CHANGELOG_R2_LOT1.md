# Changelog — GOSEC-REVIEW-R2 Lot 1

## 2026-08-26 — filesystem, archives et provenance

Le lot ajoute un extracteur d’archives navigateur transactionnel et borné, avec confinement de chemin, refus des chemins absolus/traversal, refus des symlinks et types spéciaux, limites de fichiers/taille/profondeur et activation après staging complet.

La restauration CLI utilise désormais un staging temporaire, limite le nombre d’entrées et le volume restauré, refuse hardlinks/types spéciaux, accepte seulement les symlinks internes au backup et restaure les racines avec rollback si l’activation échoue.

Les tests positifs et négatifs couvrent ZIP/TAR, traversal, séparateurs Windows, chemins absolus, profondeur, symlinks internes/externes, hardlinks, archive partiellement valide, limites d’entrées et préservation de l’état existant.

Le scan Gosec source-only post-correctif compte 155 findings contre 176 avant le lot. Les résultats Gosec restent ouverts lorsque le scanner signale encore le sink. OSV `go.mod` reste ouvert avec 46 avis; le lockfile Dashboard reste à 0. Aucun outil absent n’est déclaré PASS.

Source commit : `cd0d2e61990ceb421765f75a26cfd986ad9dc558`.
