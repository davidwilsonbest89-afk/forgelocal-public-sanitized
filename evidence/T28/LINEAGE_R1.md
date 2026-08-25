# T28 — Lignée Git R1

La qualification R1 utilise le dépôt public , la branche  et la baseline V6 .

| Commit | Rôle exact | Parent | Fichiers / statut | Package associé |
|---|---|---|---|---|
|  | Ancienne baseline historique T00–T27, tag  |  | Archive historique ; aucun code T28 | Packages T00–T27 historiques |
|  | Baseline V6 gelée, tag  |  | Baseline de comparaison ; aucune modification R1 | Baseline V6 |
|  | Commit d’implémentation/hardening cité par les scans |  | Fonctionnalité T28 et contrôles associés ; commit qualifié localement | Inclus dans le contenu repris par le package T28 |
|  | Commit de contenu/package annoncé |  | Rapport final synchronisé ; contenu source du ZIP/bundle | ZIP , bundle  |
|  | Publication du rapport et des artefacts de preuve |  | ZIP, bundle, sidecars et hashes publiés | ZIP/bundle ci-dessus |
|  | HEAD distant actuel et preuve de vérification publique |  | Journaux de clone sparse, extraction, checksums, fsck et bundle verify | Package inchangé ; preuves R1 ajoutées après publication |

## Contrôle de non-régression métier

Le diff exact  est inspecté dans . Il ne doit contenir que des documents et preuves postérieurs à l’implémentation qualifiée ; aucune modification de ,  ou autre code métier ne doit apparaître après .

Le package ZIP et le bundle correspondent au commit de contenu ; le HEAD  n’a ajouté que les journaux de vérification publique. Les sidecars sont vérifiés depuis le dossier distribué et les copies de qualification R1.

## Références brutes

- 
- 
- 
- 
