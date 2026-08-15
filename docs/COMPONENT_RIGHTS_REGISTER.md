# Registre de provenance des composants ForgeLocal

> **Source de contrôle :** [`component-rights-register.json`](component-rights-register.json). Ce document est une vue humaine synchronisée ; les preuves originales restent hors dépôt dans l’espace privé indiqué par `evidence_ref`.

| Composant | Statut | Décision | État d’intégration | Révision consommée | Responsable |
|---|---|---|---|---|---|
| BrowseForge Core | `authorized` | `direct` | intégré | `cc554320d937aeb6714aaecac3f73a8f4d599a44` | ForgeLocal maintainer |
| ForgeLocal first-party | `not_required` | `direct` | intégré | `f5deb10ea914ba906f89634914cee88a1148b7c6` | ForgeLocal maintainer |
| Camoflox | `authorized` | qualification de provenance T07 | qualification bloquée, non intégré | `sha256:dcf668…e4851c2` | ForgeLocal maintainer |
| GoLogin | `denied` | `écarter` | non intégré | non consommée | ForgeLocal maintainer |
| Persona Studio | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| DonutBrowser | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| ShardBrowser / ShardX | `authorized` | `portage` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| CloakBrowser | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |

La source Camoflox ne satisfait pas encore les conditions d’intégration : l’archive autorisée a une empreinte exacte, mais aucun commit source vérifiable ni licence racine déclarée, et son scan contient une alerte `generic-api-key` non classifiée. T07 limite donc l’examen conceptuel à `lib/concurrency.js` et `lib/global-action-limiter.js`. Les modules de queue, d’isolation, de ports et de cycle de vie runtime restent explicitement hors périmètre. GoLogin demeure un benchmark de marché public seulement et ne peut jamais entrer dans la chaîne technique.

Une source non sélectionnée ou dont la qualification est bloquée n’est pas une dépendance autorisée : avant import, elle doit déclarer une révision exacte et franchir les gates `PROV-01` à `PROV-07`.
