# Exigences de protection de branche et de release

**Statut documentaire :** requis avant toute release publique ; configuration distante à confirmer par un administrateur GitHub.
**Périmètre :** branches de livraison et tags `v*` du dépôt ForgeLocal.
**Dernière mise à jour :** 15 août 2026.

## Objet

Le contrôle de provenance est présent dans les deux chemins automatisés : le job `test` du workflow **CI** et le job `Provenance gate` du workflow **Build and Release**. Le job distinct `release-separation` du workflow **CI** vérifie en outre les rôles d’approbation des pull requests qui modifient un chemin critique. Dans la release, `verify` dépend explicitement de `provenance`, puis `build` dépend de `verify` et `release` dépend de `build`. Un échec de provenance empêche donc la construction, la publication des artefacts et la création automatisée de la release.

> Cette chaîne contrôle la release automatisée. Elle ne remplace ni la protection distante des branches et des tags, ni la règle d’accès qui doit interdire à un acteur humain de publier une release directement hors du workflow.

## Réglage mainteneur obligatoire

Un administrateur GitHub doit créer un **ruleset de branche actif** ciblant au minimum `main` et `forgelocal-product-v0.3` tant que cette dernière reste la branche produit. La règle doit imposer une pull request avant merge, **au moins deux approbations indépendantes**, le rejet des approbations obsolètes, la résolution des conversations et l’interdiction des bypass, y compris pour les administrateurs. Les force-pushes et suppressions de branches doivent rester interdits.

Le même ruleset doit activer **Require review from Code Owners** et **Require approval of the most recent reviewable push**. Cette combinaison exige la revue d’un owner des fichiers sensibles et empêche le dernier auteur d’un push de s’auto-approuver. Le rejet des approbations obsolètes doit rester actif. Un fichier `CODEOWNERS` ne suffit pas à lui seul : GitHub n’en rend la revue bloquante que si la règle distante l’exige. [3]

La règle « deux approbations » s’applique à la branche entière, car les protections GitHub standard ne permettent pas de la limiter de façon native à un sous-ensemble de chemins. C’est volontairement plus strict. Les chemins critiques — `.github/CODEOWNERS`, `.github/workflows/`, les trois scripts de provenance, le registre JSON, `release/` et `dist/` — sont en plus assignés aux trois relecteurs dans `CODEOWNERS`.

Les status checks GitHub à rendre obligatoires sont **`CI / test`** et **`CI / release-separation`**. Le premier porte les contrôles `check-component-rights.mjs`, `test-component-rights.mjs`, les fixtures du validateur de séparation, la génération de preuve redacted et son archivage pendant 90 jours. Le second évalue uniquement les métadonnées de pull request sur les chemins critiques. Les deux checks doivent être configurés en mode strict afin que le commit à merger soit à jour de la branche cible.

| Cible | Contrôle requis | Paramètres non négociables | Finalité |
|---|---|---|---|
| `main` | `CI / test` et `CI / release-separation` | Pull request, **deux approbations**, checks stricts, revues obsolètes invalidées, Code Owners, dernier push approuvé par une autre personne, pas de bypass | Empêcher l’intégration d’un changement qui aurait retiré ou contourné le contrôle de provenance ou la séparation Sécurité/Release. |
| `forgelocal-product-v0.3` | `CI / test` et `CI / release-separation` | Même politique tant que la branche produit existe | Conserver la même garantie avant la promotion vers la chaîne de release. |
| Tags `v*` | Ruleset de tag et restriction de création/mise à jour/suppression | Créateurs limités aux mainteneurs de release ; aucun bypass permanent | Empêcher un tag de release d’être créé ou modifié hors de la politique de livraison. |

Les noms de job doivent rester uniques à l’échelle des workflows afin que GitHub n’associe pas un check requis au mauvais workflow. GitHub permet d’exiger des status checks avant modification d’une branche protégée et recommande de désigner leur source lorsque cette option est disponible. [1]

## Owners des fichiers sensibles

Le fichier [`.github/CODEOWNERS`](../.github/CODEOWNERS) est la politique versionnée des zones sensibles. Il attribue les chemins sensibles à **sécurité** `@boucheriechefimane-cmd` et **release** `@davidwilsonbest89-afk` ; les chemins critiques ajoutent la relectrice indépendante `@hajarbenmlih91-cloud`. Les trois comptes doivent recevoir un accès **Write** au futur dépôt ForgeLocal ; GitHub n’assigne pas un owner qui n’existe pas ou n’a pas cette permission. [3]

| Zone protégée | Owners requis | Raison |
|---|---|---|
| `.github/CODEOWNERS` et `.github/workflows/` | Sécurité, Release et relectrice indépendante | Protéger la politique d’ownership et les contrôles d’automatisation. |
| Les trois scripts de contrôle de provenance | Sécurité, Release et relectrice indépendante | Empêcher une neutralisation du validateur et de sa preuve. |
| `docs/component-rights-register.json` | Sécurité, Release et relectrice indépendante | Protéger la source de vérité de droits des composants. |
| `package*.json`, `pnpm-lock.yaml`, `go.mod`, `go.sum` | Sécurité ou Release | Revoir toute évolution de dépendance ou de chaîne de build. |
| `release/` et `dist/` | Sécurité, Release et relectrice indépendante | Empêcher une modification d’artefact ou de chaîne de livraison hors revue. |

GitHub accepte une approbation de l’un des owners déclarés pour un chemin donné : une ligne CODEOWNERS est une alternative, non une obligation d’approbation de chaque rôle. Le minimum distant de **deux approbations** et la règle du dernier push approuvé par une autre personne restent donc nécessaires. Le check obligatoire **`CI / release-separation`** transforme en plus la séparation Sécurité/Release en contrôle technique pour les chemins critiques. [3]

## Check `CI / release-separation`

Le workflow `.github/workflows/release-separation.yml` est déclenché par les événements de pull request, mais il n’exécute ni ne télécharge le code proposé. Il utilise `pull_request_target` exclusivement pour lire les métadonnées de la pull request, la liste des fichiers et les décisions de review avec un `GITHUB_TOKEN` limité à `contents: read` et `pull-requests: read`. Le validateur est récupéré à la révision de la branche de base, et non depuis la branche proposée ; aucun `checkout`, installateur de dépendance, build ni test du code de pull request n’est admis. Cette séparation évite le scénario où un code non fiable serait exécuté avec la confiance de la branche de base. [5]

| Auteur du changement critique | Approbations exigées par `CI / release-separation` |
|---|---|
| Ni Sécurité ni Release | Une approbation Sécurité et une approbation Release. |
| Sécurité | Une approbation Release et une approbation indépendante. |
| Release | Une approbation Sécurité et une approbation indépendante. |
| Relectrice indépendante | Une approbation Sécurité et une approbation Release. |

Le script `check-release-separation.mjs` utilise pour chaque reviewer sa décision la plus récente. Une décision plus récente autre que `APPROVED` invalide donc l’approbation antérieure. Sa preuve JSON redacted n’enregistre que la version de schéma, le résultat, le rôle de l’auteur, les rôles requis et satisfaits, les rôles manquants et le nombre de fichiers critiques ; elle n’exporte ni identifiant de pull request, ni login, ni nom de fichier, ni secret.

## Environnement `production-release`

Le job `release` du workflow est maintenant rattaché à l’environnement `production-release`, après `provenance → verify → build`. L’administrateur doit créer cet environnement dans **Settings → Environments**, ajouter `@davidwilsonbest89-afk` comme reviewer requis, cocher l’interdiction de self-review, désactiver le bypass administrateur si cette option est proposée et limiter les références de déploiement aux tags `v*`. Les secrets de signature et de publication doivent être ajoutés uniquement dans cet environnement, jamais dans le code, les secrets globaux du dépôt ou les logs.

> Un job référant un environnement avec approbation requise ne peut accéder aux secrets d’environnement qu’après cette approbation. Cette propriété protège les secrets de signature ou de publication, mais son activation dépend du réglage GitHub distant et des capacités du plan du dépôt. [4]

## Protection de la release manuelle

Le job `Provenance gate` protège le chemin GitHub Actions déclenché par un tag `v*`. Un fichier du dépôt ne peut toutefois pas retirer, à lui seul, le droit d’un administrateur GitHub de créer une release par l’interface : cette barrière est une **décision de droits distante**. Les mainteneurs doivent donc restreindre les rôles capables de créer des releases et des tags de version au compte ou à l’automatisation officiellement responsables de la release ; tout administrateur doit être soumis au ruleset, sans bypass permanent.

Si l’offre GitHub du dépôt ne permet pas de distinguer finement l’autorisation de créer une release des autres droits d’écriture, une release manuelle par un administrateur reste un contournement organisationnel possible. Dans ce cas, ForgeLocal doit conserver le statut `PUBLIC_RELEASE_BLOCKED` pour toute release dont l’attestation `component-provenance-<commit>` est absente, et l’accès administrateur doit être limité aux mainteneurs de release nommés. Il serait inexact de qualifier la protection de « non contournable » tant que ce réglage distant et cette restriction de rôles n’ont pas été vérifiés.

## Procédure de vérification et preuve

Après activation, l’administrateur doit ouvrir une pull request de test qui modifie un fichier non sensible. Le merge doit être impossible tant que **`CI / test`** et **`CI / release-separation`** ne sont pas verts et que deux approbations ne sont pas présentes. Il doit ensuite ouvrir une seconde pull request qui modifie un chemin CODEOWNERS, sans la merger, et confirmer que l’owner est requis, que le dernier auteur ne peut pas auto-approuver, que deux approbations restent nécessaires et qu’une approbation antérieure devient obsolète après un nouveau push. Lorsque ni Sécurité ni Release n’est auteur, cette seconde pull request doit être revue par `@boucheriechefimane-cmd` et `@davidwilsonbest89-afk`. Si l’un de ces deux responsables est auteur, l’autre responsable et `@hajarbenmlih91-cloud` doivent fournir les deux approbations. Le check `CI / release-separation` doit échouer pour chaque combinaison de rôles incomplète du tableau précédent et produire une preuve redacted lorsqu’elle réussit. Enfin, il doit vérifier que le workflow de release exécute le job **`Provenance gate`**, que l’artefact `component-provenance-<SHA>` est présent avec une rétention de 90 jours et qu’aucune release automatisée n’est créée si ce job échoue.

Pour l’environnement, il doit déclencher une release de test non publique ou une exécution contrôlée, vérifier l’arrêt du job `release` dans l’attente du reviewer `@davidwilsonbest89-afk`, confirmer que l’auteur ne peut pas s’auto-approuver puis vérifier que les secrets ne sont disponibles qu’après approbation. Aucun secret ne doit être affiché dans la preuve.

La preuve d’activation à archiver dans le dossier de sécurité doit contenir la date, le dépôt, les branches et le motif de tags couverts, les noms exacts des deux checks requis, l’absence de bypass, le seuil de deux approbations, l’identité des trois relecteurs, la présence des droits Write, la configuration de l’environnement, ainsi que des captures ou exports redacted des règles. Elle doit également inclure l’artefact `release-separation-<head-sha>` d’un test réussi, sans modifier son contenu. Elle ne doit contenir aucun token, email personnel, chemin d’hôte, valeur proxy ni secret. Les règles GitHub permettent de cibler des branches ou tags et de rendre des status checks obligatoires avant les modifications couvertes. [1] [2]

## État de conformité

| Élément | État au 14 août 2026 | Preuve attendue |
|---|---|---|
| Contrôle de provenance dans le workflow CI | Implémenté dans le dépôt | Exécution `CI / test` réussie. |
| Contrôle de provenance dans le workflow de release | Implémenté dans le dépôt | Job `Provenance gate` et artefact associé. |
| Contrôle technique de séparation Sécurité/Release | Implémenté dans le dépôt ; application distante **PENDING** | `CI / release-separation` requis, test des quatre scénarios et artefact redacted associé. |
| CODEOWNERS des fichiers sensibles | Implémenté dans le dépôt ; application distante **PENDING** | Les trois comptes disposent de Write et une pull request de test exige la revue demandée. |
| Revue renforcée des chemins critiques | Politique versionnée ; application distante **PENDING** | Ruleset à deux approbations, Code Owners et test de pull request avec preuve des deux revues. |
| Environnement `production-release` | Déclaré dans le workflow ; configuration distante **PENDING** | Reviewer release, self-review interdite, bypass désactivé et secrets limités à l’environnement. |
| Protection distante des branches | **Non vérifiable dans cette session** : le remote pointe vers `nczz/BrowseForge` et l’identité GitHub disponible n’a que le droit `READ`; l’API de protection de `main` répond `403 Resource not accessible by integration`. | Export/capture redacted du ruleset actif depuis le dépôt ForgeLocal sous contrôle du mainteneur. |
| Restriction des tags et des releases manuelles | **Non vérifiable dans cette session** : aucun droit d’administration ou de publication n’est disponible sur le remote observé. | Ruleset `v*`, revue des rôles autorisés et preuve que la création manuelle est limitée. |

> **Décision de conformité :** le contrôle de provenance est bloquant dans les deux workflows versionnés, mais son caractère non contournable au niveau GitHub reste **PENDING**. Il ne doit pas être déclaré satisfait ni utilisé pour lever `PUBLIC_RELEASE_BLOCKED` tant que le dépôt ForgeLocal contrôlé par ses mainteneurs n’a pas une règle distante active et vérifiée.

## Références

[1] [GitHub Docs — About protected branches](https://docs.github.com/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
[2] [GitHub Docs — Available rules for rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets)
[3] [GitHub Docs — About code owners](https://docs.github.com/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
[4] [GitHub Docs — Deployments and environments](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments)
[5] [GitHub Docs — Securely using pull_request_target](https://docs.github.com/en/actions/reference/security/securely-using-pull_request_target)
