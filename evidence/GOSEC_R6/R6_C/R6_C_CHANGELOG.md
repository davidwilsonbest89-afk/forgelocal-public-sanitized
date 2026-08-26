# GOSEC-R6 Lot C — changelog

Le Lot C a recalculé 26 findings : G302=5, G115=3, G404=17 et G101=1. Les contrôles existants ont été vérifiés; une régression supplémentaire confirme que `groups.json` est écrit en 0600 dans un répertoire 0700, et une régression confirme les bornes de seed et de taille d’archive.

Aucun finding G302/G115/G404/G101 n’est supprimé du scan. Les permissions et conversions restent `MITIGATED_CONTROL_SCANNER_OPEN`; les usages d’aléatoire de simulation/fingerprint et le motif de nom de fichier token restent `SCANNER_OPEN_MANUAL_REVIEW`.
