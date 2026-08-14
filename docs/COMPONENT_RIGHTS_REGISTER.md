# Registre de provenance des composants ForgeLocal

> **Source de contrôle :** [`component-rights-register.json`](component-rights-register.json). Ce document est une vue humaine synchronisée ; les preuves originales restent hors dépôt dans l’espace privé indiqué par `evidence_ref`.

| Composant | Statut | Décision | État d’intégration | Révision consommée | Responsable |
|---|---|---|---|---|---|
| BrowseForge Core | `authorized` | `direct` | intégré | `cc554320d937aeb6714aaecac3f73a8f4d599a44` | ForgeLocal maintainer |
| ForgeLocal first-party | `not_required` | `direct` | intégré | `f5deb10ea914ba906f89634914cee88a1148b7c6` | ForgeLocal maintainer |
| Camoflox | `authorized` | `portage` | non intégré | `sha256:dcf668…e4851c2` | ForgeLocal maintainer |
| GoLogin | `denied` | `écarter` | non intégré | non consommée | ForgeLocal maintainer |
| Persona Studio | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| DonutBrowser | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| ShardBrowser / ShardX | `authorized` | `portage` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| CloakBrowser | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |

Une source non sélectionnée n’est pas une dépendance autorisée : avant import, elle doit déclarer une révision exacte et franchir les gates `PROV-01` à `PROV-07`. GoLogin demeure un benchmark de marché public seulement et ne peut jamais entrer dans la chaîne technique.
