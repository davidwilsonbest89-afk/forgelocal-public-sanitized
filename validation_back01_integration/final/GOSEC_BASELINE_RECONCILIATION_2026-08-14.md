# Réconciliation de la baseline Gosec — 14 août 2026

## Objet

Ce document corrige la traçabilité de la dette Gosec mentionnée dans la spécification ForgeLocal v1.0. Il ne clôt aucune alerte, ne modifie aucun gate de release et ne qualifie aucune publication.

## Rapport historique à 189 résultats

| Attribut | Valeur vérifiée |
|---|---|
| Rapport | `sandbox-gosec-20260814.json` |
| Date d’exécution observée | 14 août 2026, 14:01 UTC |
| Nombre de résultats | `189` |
| SHA-256 du rapport | `332ae84056ec9ad5d15a965674540f0d5f21215bbeb998dd6bf614516c65b978` |
| Version déclarée par le rapport | `dev` |
| Commit source déclaré | **Absent** |
| Décision | Preuve historique informative seulement ; non conforme à l’exigence de baseline reproductible. |

Le rapport contient `GosecVersion: dev` et aucun commit source. Il est par conséquent interdit de l’utiliser comme preuve que « 189 alertes » correspondent à un état versionné déterminé. Aucun résultat ne doit être ignoré pour autant : la dette historique doit être classifiée ou régénérée avec un outillage verrouillé.

## Baseline reproductible de contrôle

| Attribut | Valeur vérifiée |
|---|---|
| Date d’exécution | 14 août 2026 |
| Commit source | `64bede39dc3355e0db2c4871cf4de7eb46410265` |
| Version Go | `go1.25.13 linux/amd64` avec `GOTOOLCHAIN=local` |
| Version Gosec | `2.21.4` (`v2.21.4`) |
| Archive Gosec vérifiée | `gosec_2.21.4_linux_amd64.tar.gz` |
| SHA-256 de l’archive Gosec | `9229dbfdc092b176e628b9ea6e4210757373b819f47365cedd9f9e12d3b2c173` |
| Commande | `gosec -fmt=json -out gosec-baseline-64bede3-v2.21.4-20260814.json ./...` |
| Code de sortie | `1`, attendu lorsque des résultats sont présents |
| Nombre de résultats | `166` |
| SHA-256 du rapport | `ba42d3e2af1fe8d9a61407ed87d54c57b7d81cccfaddc9fa382b07f35e06ec9d` |

> La baseline de 166 résultats ne réduit pas mécaniquement la dette de 189 résultats. Elle établit seulement un point de comparaison reproductible avec la version verrouillée de Gosec.

## Règles applicables

Toute nouvelle baseline doit contenir, dans un même dossier de preuve, le rapport JSON, son SHA-256, le commit Git complet, la date UTC, la commande exacte, les versions de Go et Gosec ainsi que le SHA-256 de l’archive Gosec. Les changements de commit, de version d’outil, de cibles analysées ou de configuration invalident la comparaison numérique.

Les analyses du lot produit doivent être comparées à une baseline exécutée avec les mêmes paramètres. Une baisse de nombre n’est jamais une clôture : chaque résultat doit avoir une décision, un risque, un propriétaire, une échéance et, lorsque pertinent, un test de non-régression. Les gates `PUBLIC_RELEASE_BLOCKED`, le pilote suspendu et l’alerte `generic-api-key` restent inchangés.
