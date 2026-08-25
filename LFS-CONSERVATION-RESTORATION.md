# LFS-CONSERVATION-RESTORATION — lot séparé

**Décision :** `LFS_CONSERVATION_RESTORED_BYTE_IDENTICAL_FSCK_ZERO`

**Base :** audit LFS publié au commit `a236291262bd11513e4c82474cce41d3f7c7a726`. La branche de restauration est dédiée et ne modifie ni la branche V6 gelée, ni le Core, ni le Dashboard métier, ni les wrappers historiques.

## Résultat

Les douze OID déclarés encore indisponibles dans la matrice V6 sont maintenant présents dans le store LFS partagé. Pour chacun, la taille locale correspond à la taille déclarée et le SHA-256 du contenu correspond exactement à l’OID LFS. Le contrôle `git lfs fsck` par défaut et le contrôle `git lfs fsck --objects` retournent tous deux zéro.

La présence des objets a été observée après une tentative de checkout sparse mal isolée qui a lancé le smudge LFS concurrent avant saturation de l’espace ; cette cause est conservée dans les journaux et n’est pas présentée comme une récupération volontaire parfaite. Une seconde procédure contrôlée a ensuite exécuté un fetch filtré par chemin pour chacun des douze chemins, sans jamais exécuter `git lfs pull` global, et a confirmé individuellement les OID, tailles et contenus. La branche ne réhydrate aucun fichier utilisateur ni runtime.

## Matrice individuelle

| # | OID | Taille | Chemin | Commit porteur | Criticité | Résultat |
|---:|---|---:|---|---|---|---|
| 1 | `a7e8729c65c8dfd1ef64e713331f6036daa210ac6cf9f4f5f1a6f38d87634387` | 1,992,294,400 | `continuation/T00-T27/archive-packages/forgelocal-continuation-t00-t23-20260819.zip.part.00` | `69411e65c880d168832a65fc8475cc97d562a9ad` | historique | `RESTORED_BYTE_IDENTICAL` |
| 2 | `20b32719150556255aebb4c9aa2413be9012197e178d5d7e841bc9b0b8f06d29` | 425,038,907 | `continuation/T00-T27/archive-packages/forgelocal-continuation-t00-t23-20260819.zip.part.01` | `69411e65c880d168832a65fc8475cc97d562a9ad` | historique | `RESTORED_BYTE_IDENTICAL` |
| 3 | `037eb5650a8fcefb1b2ab9bf747772ab3cf2ead9c698ca85675d433da56f4526` | 153,973,983 | `continuation/T00-T27/archive-packages/forgelocal-t24-bulk-evidence-7f63c71.zip` | `69411e65c880d168832a65fc8475cc97d562a9ad` | historique | `RESTORED_BYTE_IDENTICAL` |
| 4 | `9c9be6107e420e7a8c83ac64d6f3b80063bcad10b08f5a7c4c395f342bbd61f3` | 153,984,867 | `continuation/T00-T27/archive-packages/forgelocal-t25-synthetic-cookie-fixtures-a7ccfbc.zip` | `69411e65c880d168832a65fc8475cc97d562a9ad` | historique | `RESTORED_BYTE_IDENTICAL` |
| 5 | `bb984b9d991538862079b68d16804136e1912e689d89ac262866cc5221ac3f18` | 153,990,161 | `continuation/T00-T27/archive-packages/forgelocal-t26-simulated-provider-930003c.zip` | `69411e65c880d168832a65fc8475cc97d562a9ad` | historique | `RESTORED_BYTE_IDENTICAL` |
| 6 | `2a13180aa44db915313c4bd9a0d2ae8472a5cea6dcbddebcc6c70e222a78fb22` | 438,834,687 | `continuation/T00-T27/archive-packages/forgelocal-t27-r1-local-continuation-full.tar.gz` | `69411e65c880d168832a65fc8475cc97d562a9ad` | haute historique | `RESTORED_BYTE_IDENTICAL` |
| 7 | `ab57d11f59fa0d13db7ded91e0a468413a094104ae8778761df919d103cbd271` | 153,920,278 | `continuation/T27-R1/evidence/CR01/forgelocal-t27-r1-cr01-df50a2b.bundle` | `282fb0a28bf48a15465341b02f82c83e09e2fd92` | haute historique | `RESTORED_BYTE_IDENTICAL` |
| 8 | `b1cdcf44d081ef6e30f2f5524156765c88d1dc2a9e24a08048cad025935d741b` | 154,124,106 | `continuation/T27-R1/evidence/CR01/forgelocal-t27-r1-cr01-evidence-df50a2b.zip` | `282fb0a28bf48a15465341b02f82c83e09e2fd92` | haute historique | `RESTORED_BYTE_IDENTICAL` |
| 9 | `14ef76cb68e7f64ff49fdc649cbcf96c5c69b0e9c410c5824a0592b7e33d1d14` | 109,839,371 | `evidence/PREHUMAN_T00_T42/forgelocal-t28-t42-prehuman-validation.bundle` | `cf280858b345e2fd566d391590f23d8cfa6bbe6d` | haute historique | `RESTORED_BYTE_IDENTICAL` |
| 10 | `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80` | 109,998,883 | `evidence/PREHUMAN_T00_T42/forgelocal-t28-t42-prehuman-validation.zip` | `cf280858b345e2fd566d391590f23d8cfa6bbe6d` | haute historique | `RESTORED_BYTE_IDENTICAL` |
| 11 | `27598b6d55c767d338e0f61d7acf7f7d6786664d05752117cf57c7fd28fd48ec` | 110,452,290 | `evidence/forgelocal-t00-t42-prehuman-final-review-wrapper-v2.zip` | `696dd8a16a31978b5ec1ef32289c0897ca7e9cd6` | historique | `RESTORED_BYTE_IDENTICAL` |
| 12 | `10059c3c610d5a1b1ade88f936c8bb52ed893741a596326d8ba532f6f415e2fe` | 271,949 | `evidence/forgelocal-t00-t42-self-validation-v4.delta.bundle` | `b4a04e4b9b489c22f3a86986c6faa1cbb9bf77c5` | historique | `RESTORED_BYTE_IDENTICAL` |

Les deux objets initialement récupérés avant ce lot — wrapper V3 et wrapper V4 — restent documentés dans la matrice historique V6. Le comptage est donc 14 absents à la baseline, 2 récupérés avant le lot et 12 traités ici.

## Clôture

La condition technique de clôture est satisfaite par `git lfs fsck=0` et par la comparaison individuelle OID/taille/SHA-256. Une revue indépendante doit encore confirmer la provenance opérationnelle des objets et la conservation durable du store LFS. Owner : **administrateur du dépôt / responsable des preuves de release**.
