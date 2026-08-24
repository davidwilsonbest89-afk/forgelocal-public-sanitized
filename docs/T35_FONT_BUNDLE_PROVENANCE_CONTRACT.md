# T35 — Font Bundle : inventaire, provenance et licences

**Statut :** `T35_FONT_BUNDLE_DOCUMENTARY_INVENTORY_APPROVED_VERIFIABLE_LOCAL`
**Prérequis :** T34 distant `c7f66da0da6f547813d10826dfd7772ad7e0f4b6`.
**Mode :** métadonnées déclarées uniquement ; aucun binaire de police fourni.

## Périmètre fermé

T35 fournit un inventaire documentaire minimal `font-bundle` avec provenance `NOT_SUPPLIED`, licence `PENDING_REVIEW`, redistribution `BLOCKED` et état `NOT_INCLUDED`. L’inventaire est déterministe et ne lit aucune police installée sur l’hôte.

Le validateur refuse les identifiants vides, absolus, traversant un répertoire ou contenant un séparateur de chemin. Une redistribution `GRANTED` est refusée tant que la licence n’est pas `APPROVED`. Les licences absentes ou `UNKNOWN` sont bloquantes.

## Limites de licence et de provenance

Aucune police, fonte, fichier de licence, notice tierce ou donnée de fournisseur n’est copiée dans ForgeLocal. `NOT_SUPPLIED` ne signifie ni propriété, ni permission, ni approbation. Toute future police doit fournir une provenance, une licence vérifiable, les notices requises et une décision explicite de redistribution avant ajout.

## Critères d’acceptation

| Critère | Attendu |
|---|---|
| Inventaire | Entrée déclarative stable, sans binaire |
| Provenance | `NOT_SUPPLIED` tant qu’aucune source n’est fournie |
| Licence | `PENDING_REVIEW` par défaut |
| Redistribution | `BLOCKED` par défaut et refusée sans licence approuvée |
| Sécurité | Chemins absolus et traversals refusés |
| Qualification | Tests race, suite globale, vet, build, format et Gitleaks |

T35 reste documentaire et ne constitue pas une validation de licence ni une autorisation de redistribution. T36 peut commencer après publication et vérification du commit T35.
