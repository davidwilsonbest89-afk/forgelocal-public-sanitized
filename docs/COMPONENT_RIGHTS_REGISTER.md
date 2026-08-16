# Registre de provenance des composants ForgeLocal

> **Source de contrôle :** [`component-rights-register.json`](component-rights-register.json). Ce document est une vue humaine synchronisée ; les preuves originales restent hors dépôt dans l’espace privé indiqué par `evidence_ref`.

| Composant | Statut | Décision | État d’intégration | Révision consommée | Responsable |
|---|---|---|---|---|---|
| BrowseForge Core | `authorized` | `direct` | intégré | `cc554320d937aeb6714aaecac3f73a8f4d599a44` | ForgeLocal maintainer |
| ForgeLocal first-party | `not_required` | `direct` | intégré | `f5deb10ea914ba906f89634914cee88a1148b7c6` | ForgeLocal maintainer |
| Camoflox | `authorized` | `reimplementer-selective` | T08 concurrence en cours, non intégré | `sha256:dcf668…e4851c2` | ForgeLocal maintainer |
| GoLogin | `denied` | `écarter` | non intégré | non consommée | ForgeLocal maintainer |
| Persona Studio | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| DonutBrowser | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| ShardBrowser / ShardX | `authorized` | `portage` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |
| CloakBrowser | `authorized` | `adaptateur` | source non sélectionnée | à sélectionner avant import | ForgeLocal maintainer |

La provenance Camoflox a franchi le gate T07 au statut `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION` (décision `T07-DECISION-20260815-001`, dossier redacted `a10087c8…d040041c3`, attestation `ATT-T07R-CAMOFLOX-001`, revues indépendantes `REV-T07R-2026-001` et `REV-T07R-2026-002`, décision finale `REVIEW_ACCEPTED_FOR_T07_DECISION`). Le triage de l’alerte `generic-api-key` est confirmé `FALSE_POSITIVE` par le détenteur et la relectrice, avec snapshot redacted rescanné à zéro détection. Les contrôles de provenance sont passés : PROV-01 à PROV-04 et PROV-06 `PASS`, PROV-05 et PROV-07 reportés à T08 ; la distribution reste `not_granted` et toute distribution future est bloquée jusqu’à une revue de droits dédiée.

T08 est autorisé strictement pour la réimplémentation Go d’un module déjà hashé à la fois. La première décision enregistrée est `reimplementer` pour `lib/concurrency.js` (`sha256:b055a3e1…a36085d7`) comme inspiration de contrat de concurrence uniquement, avec exclusions explicites : code Node/Electron, lancement runtime, cycle de vie navigateur, ports, queue de lancement, isolation de processus et activation Camoufox. `lib/global-action-limiter.js` est reporté ; les modules `lib/profile-launch-queue.js`, `lib/process-isolation.js` et `lib/browser-lifecycle.js` restent explicitement hors périmètre. GoLogin demeure un benchmark de marché public seulement et ne peut jamais entrer dans la chaîne technique.

Cette approbation ne lève aucun gate de release : `PUBLIC_RELEASE_BLOCKED` et `SCAN_BLOCKED_UNKNOWN` (BACK-01) restent actifs, le pilote BACK-01 est suspendu et cinq gates publics demeurent en attente.

Une source non sélectionnée ou dont la qualification est bloquée n’est pas une dépendance autorisée : avant import, elle doit déclarer une révision exacte et franchir les gates `PROV-01` à `PROV-07`.
