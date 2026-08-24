# Prévalidation humaine — chaîne historique T00–T27

## Périmètre et méthode

L’audit historique a été réalisé depuis le clone neuf `/home/ubuntu/forgelocal-prehuman-fresh-20260824`, figé au commit `6ae02e4ceed239b9310fbf3fccb1b5170117251e`. Le tag de conservation `t00-t27-complete-20260820` a été résolu vers `69411e65c880d168832a65fc8475cc97d562a9ad`. Les journaux historiques existants n’ont pas été rejoués ni réinterprétés : les résultats ci-dessous proviennent de contrôles actuels séparés, et toute opération postérieure est explicitement identifiée comme telle.

## Résultats directs

| Périmètre | Contrôle | Résultat |
|---|---|---|
| T00–T23 | Réhydratation LFS ciblée des deux morceaux, concaténation temporaire et SHA-256 | `0046e99e22ad1498a6307f4acd43d1cebbb7c317d47edee16329fc38b59b924d`, conforme |
| T00–T23 | `unzip -t` de l’archive reconstruite | `exit_code=0` |
| T24 | Hash du ZIP LFS et manifest global | conforme |
| T25 | Hash du ZIP LFS et manifest global | conforme |
| T26 | Hash du ZIP LFS et manifest global | conforme |
| T27-R1 | Hash du tarball LFS et manifest global | conforme |
| CR01–CR05 | ZIP, bundle et sidecars historiques | contrôles d’intégrité à `exit_code=0` |
| T00–T27 | `git lfs fsck` après fetch ciblé et checkout local | `Git LFS fsck OK`, `exit_code=0` |

Les huit objets LFS suivis ont été récupérés un à un avec `git lfs fetch origin --include=<chemin>`, et non par `git lfs pull` global. L’espace libre était contrôlé avant et après chaque objet ; le clone et les artefacts ont ensuite été traités sans dépasser la règle d’arrêt sous 10 Go.

## Classification de provenance

Les archives et preuves présentes sous `continuation/T00-T27/` et `continuation/T27-R1/` sont les artefacts historiques originaux versionnés ou leurs sidecars historiques. L’archive T00–T23 est une reconstruction technique postérieure à partir de deux morceaux LFS originaux ; son hash et son intégrité ZIP ont été vérifiés. Cette reconstruction ne constitue pas un nouveau log historique.

Les contrôles SHA-256, `unzip -t`, `git bundle verify`, `git lfs fsck` et les clones de vérification réalisés pendant cette session sont des validations postérieures. Ils ne remplacent aucun journal historique absent. Les sidecars historiques restent conservés tels quels ; lorsqu’un sidecar ancien n’est pas portable hors de son répertoire d’origine, cela est distinct des sidecars compagnons `*.portable.sha256` ajoutés pour les lots récents.

Les lots T00–T23 sont distribués comme un paquet combiné, et non comme vingt-quatre ZIP indépendants. L’audit ne prétend donc pas avoir validé des artefacts unitaires T00, T01, etc. qui ne sont pas présents séparément dans la branche.

## Conclusion historique

La chaîne d’artefacts historiques T00–T27 est **réhydratable et vérifiable localement** dans les limites décrites ci-dessus. Cette conclusion porte sur l’intégrité et la provenance des preuves, pas sur une autorisation produit, une exécution runtime ou une release. Les contrôles bruts sont conservés dans `PREHUMAN_T00_T42_HISTORICAL_AUDIT_RAW.log` et `PREHUMAN_T00_T42_LFS_REHYDRATION_RAW.log`.
