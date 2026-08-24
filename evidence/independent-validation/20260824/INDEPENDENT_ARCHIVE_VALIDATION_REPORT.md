# ForgeLocal — rapport de validation indépendante des archives

**Date de l’audit :** 24 août 2026
**Autorité de baseline :** tag `t00-t27-complete-20260820`, commit `69411e65c880d168832a65fc8475cc97d562a9ad`
**Dépôt de référence :** [forgelocal-public-sanitized](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized)
**Branche de passation consultée :** `handover/t28-t42-new-session`
**Principe appliqué :** aucune archive n’est considérée comme validée sur la seule base du rapport d’un autre agent ; chaque paquet a été extrait en quarantaine, vérifié, réhydraté par bundle lorsque possible, puis testé dans une sandbox équipée d’un véritable environnement de compilation.

## Conclusion exécutive

Les cinq archives ont été ouvertes, inventoriées, testées et comparées à la baseline. **T30 est le seul lot pouvant recevoir un verdict local positif complet sur les contrôles applicables**, car son bundle mène bien au commit annoncé, son manifeste maître est vérifiable, son delta compile, ses tests race globaux passent et le script de replay fourni retourne `exit_code=0` dans un environnement disposant d’environ 33,9 GiB libres.

**T28, T29, CR06 et CR08 ne sont pas des lots de fonctionnalité native validés.** Ils contiennent principalement des contrats documentaires et des preuves. T28, T29 et CR06 restent soumis à autorisation produit et ne modifient pas le code applicatif. CR08 reste explicitement non exécuté dans l’environnement natif requis. Plusieurs sidecars de bundle contiennent en outre des chemins absolus dépendants de la machine d’origine ; ils ne sont donc pas portables même lorsque le hash binaire calculé localement correspond.

Les gates permanents restent inchangés : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Aucun secret réel, navigateur, proxy réel, cookie, Camoufox, migration utilisateur, release ou qualification native SystemVault n’a été lancé.

## Environnement réellement utilisé

| Élément | Résultat observé |
|---|---|
| Espace libre pendant les tests | Environ `35 GiB` sur `/home/ubuntu` avant installation ; environ `33,9 GiB` disponibles dans le replay T30 |
| Go | `go1.25.13` |
| Compilateur CGO | GCC Ubuntu `13.3.0`, installé avant les tests `-race` |
| Git LFS | Disponible ; les clones de qualification utilisaient `GIT_LFS_SKIP_SMUDGE=1` pour ne pas confondre absence d’objets LFS hérités et validité du delta |
| Gitleaks | Disponible et exécuté sur les deltas binaires via `git diff --binary … | gitleaks detect --pipe`, ainsi qu’en scan complet de l’arbre |
| Tests Go | `CGO_ENABLED=1 go test -count=1 -race ./...` exécuté sur chaque bundle livrant un checkout Go |
| Contrôles complémentaires | `go vet ./...`, `go build ./...`, `git fsck --full`, `git diff --check`, `git bundle verify`, tests ZIP et manifestes SHA-256 |

Les commandes et sorties brutes sont conservées dans les cinq journaux indépendants joints à ce rapport.

## Matrice des verdicts

| Archive | Commit de livraison constaté | Contrôles applicables réussis | Écarts constatés | Verdict indépendant |
|---|---:|---|---|---|
| `agent03-t28-delivery(1).zip` | `a3a6ff621d7d30dace02678631ac2dc701b77f07` | ZIP valide, bundle vérifié contre la baseline, checkout frais, `git fsck`, delta Gitleaks, scan complet Gitleaks, `go test -race ./...`, `go vet`, `go build` | Le sidecar contient `/home/ubuntu/agent03-delivery/agent03-t28.bundle`; le contrôle brut du manifeste échoue dans la structure ZIP d’origine, mais réussit après ré-enracinement documenté ; aucun code d’extension n’est livré | **Documentaire vérifiable localement, mais non approuvé comme implémentation T28** ; `T28_PRODUCT_AUTHORIZATION_PENDING` |
| `T29_COMPLETE_PACKAGE.zip` | `5e35e0d11bd833a10fa49c0866a7e421e23faf5a` | Manifestes externe et interne après ré-enracinement, ZIP valide, bundle vérifié, delta documentaire exact, Gitleaks, `go test -race ./...`, `go vet`, `go build` | Le sidecar contient `/home/ubuntu/AGENT_04_T29_COMPLETENESS_DELIVERY/…`; aucune implémentation du secret manager et aucune qualification native | **Contrat documentaire vérifiable localement, non autorisé** ; `T29_CONTRACT_PENDING_PRODUCT_AUTHORIZATION`, `NATIVE_SYSTEMVAULT_NOT_TESTED` |
| `forgelocal-t30-master-handover-cbf3a50(1).zip` | `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb` | Manifeste maître, sidecars R3 et bundle, source snapshot, bundle ancestry, checkout frais, Gitleaks delta, `go test -race ./...`, `go vet`, `go build`, script de replay complet | Gosec retourne le code 1 car il signale des constats historiques ; la comparaison indépendante baseline/head trouve exactement 16 constats identiques, zéro nouveau et zéro résolu | **`APPROVED_VERIFIABLE_LOCAL` pour le delta T30 et son replay local** ; pas de release ni de push automatique |
| `CR06-final-delivery.zip` | `4c1be659841019d0592bef8253414338ca404aa1` | Manifeste, bundle vérifié, checkout frais, delta limité aux documents/preuves, Gitleaks delta/complet, `go test -race ./...`, `go vet`, `go build` | Sidecar avec `/home/ubuntu/cr06-counter-audit-final/…`; aucune migration ou restauration réelle exécutée ; autorisation produit absente | **Contrat documentaire prêt sous réserves** ; `CR06_CANONICAL_STORAGE_CONTRACT_READY_PENDING_PRODUCT_AUTHORIZATION` |
| `agent01-cr08-senior-complete.zip` | `b98f0e035a03c413e279ebe80593d0e7b33a72da` | Bundle vérifié avec tag et HEAD, checkout frais au vrai HEAD, delta documentaire, Gitleaks, `go test -race ./...`, `go vet`, `go build`, garde anti-fausse-qualification native | Le manifeste déclaré ne correspond pas à l’archive extraite : 13 fichiers référencés sont absents et `DIFF_CHECK.log` a un hash différent ; sidecar avec `/home/ubuntu/agent01-senior-delivery/…`; SystemVault natif non exécuté | **`BLOCKED_CORRIGIBLE_PACKAGE`** ; `CR08_NOT_EXECUTED_NATIVE_ENVIRONMENT_REQUIRED` |

## T28 — Extensions Agent 03

Le bundle T28 est cryptographiquement lisible par Git et sa tête `a3a6ff6` est récupérable depuis la baseline. Le delta réel ne contient que le contrat `docs/T28_EXTENSIONS_CONTRACT_AND_TEST_MATRIX.md` et les preuves sous `evidence/AGENT_03/**`. Les tests Go globaux, `go vet`, `go build`, `git diff --check`, `git fsck --full` et les scans Gitleaks indépendants passent.

Cela ne constitue toutefois pas une implémentation d’extensions. Le rapport livré déclare lui-même `CODE_CHANGED: Non`, `TEST_EXECUTED: Non pour le produit` et des décisions produit ouvertes. Le sidecar du bundle est également non portable, car il référence un chemin absolu de la machine d’origine. Après reconstruction d’un répertoire de vérification adapté à la structure de l’archive, toutes les sommes du manifeste ont correspondu ; ce résultat ne corrige pas le défaut de portabilité du sidecar.

**Décision :** ne pas intégrer ni autoriser T28 Extensions. Conserver le lot comme contrat documentaire vérifiable et attendre l’autorisation produit/sécurité, l’allowlist, la provenance et l’implémentation future explicitement prévues.

## T29 — Secret manager et modèle de menace

Le manifeste ZIP externe et le manifeste interne textuel passent après extraction. Une tentative de lecture JSON du manifeste interne n’était pas applicable : il s’agit d’une liste textuelle SHA-256 ; le contrôle canonique `sha256sum -c` a retourné zéro après ré-enracinement.

Le bundle mène à `5e35e0d`, son delta est limité à `docs/T29_*` et `evidence/AGENT_04/**`, et les tests Go globaux passent. Le sidecar n’est pas portable en raison d’un chemin absolu. La politique interdit correctement le fallback en clair et maintient `T29_CONTRACT_PENDING_PRODUCT_AUTHORIZATION` ainsi que `NATIVE_SYSTEMVAULT_NOT_TESTED`. Aucune valeur secrète réelle et aucune résolution de coffre n’ont été testées.

**Décision :** accepter uniquement comme contrat de revue documentaire. Ne pas présenter T29 comme secret manager implémenté ou qualifié.

## T30 — Diagnostics d’environnement

Le bundle final vérifie exactement le commit annoncé `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb` et requiert correctement la baseline `69411e65c880d168832a65fc8475cc97d562a9ad`. Le manifeste maître, les sidecars R3, l’archive R3, la source snapshot et le bundle passent les contrôles indépendants.

Le delta constaté est cohérent avec la portée : contrat T30, test HTTP/OpenAPI, modification de l’OpenAPI, code `internal/environment`, tests d’environnement et script de replay. Le scan Gitleaks du diff binaire retourne zéro fuite. Le replay fourni mesure `available_kib=35533224`, contre `required_kib=5242880`, puis exécute réellement `go test -count=1 -race ./...` et retourne `exit_code=0`. Le même test global, `go vet` et `go build` ont été relancés indépendamment et passent.

Gosec signale 16 constats dans la baseline et 16 dans la tête ; la comparaison structurée montre `new_findings=[]`, `resolved_findings=[]` et `same_finding_set=true`. Le code de sortie 1 de Gosec est donc conservé comme résultat de l’outil, sans être transformé artificiellement en zéro ; seule la conclusion différentielle « aucun nouveau finding » est retenue.

**Décision :** T30 est vérifiable localement et son gate de replay global est maintenant passé dans cet environnement. Cela n’ouvre ni la release, ni les gates natifs, ni une publication GitHub automatique.

## CR06 — Stockage canonique

Le bundle CR06 mène à `4c1be659`, avec un delta limité aux deux contrats CR06 et à leurs preuves. Le manifeste, le bundle, le checkout frais, le `git fsck`, le scan Gitleaks et la qualification Go globale passent. Le sidecar échoue uniquement comme contrôle portable, car son nom de fichier contient un chemin absolu `/home/ubuntu/cr06-counter-audit-final/…` ; le hash du bundle calculé localement est cohérent avec la valeur déclarée.

Le contenu ne livre aucun SQL, aucune migration et aucun code applicatif. Les migrations JSON→SQLite, le cutover, la restauration réelle, le runtime, les données utilisateur et SystemVault restent non testés. **CR06 reste donc un contrat prêt sous réserves, pas une migration approuvée.**

## CR08 — Qualification native SystemVault

Le bundle CR08 est vérifiable et mène réellement à `b98f0e035a03c413e279ebe80593d0e7b33a72da`. Le delta est documentaire et les tests Go généraux de la baseline enrichie passent. La garde indépendante confirme qu’aucun texte de qualification native réussie n’a été introduit.

L’archive est néanmoins **corrigible mais non acceptable en l’état comme paquet de preuves** : son manifeste déclare des fichiers absents de l’extraction et un `DIFF_CHECK.log` dont le hash déclaré ne correspond pas au fichier présent. Le sidecar est non portable. Surtout, le paquet affirme explicitement `CR08_NOT_EXECUTED_NATIVE_ENVIRONMENT_REQUIRED`; aucune opération native SystemVault ne doit être simulée ou requalifiée depuis la sandbox.

## Actions autorisées ensuite

La correction prioritaire est de demander aux agents concernés des sidecars ne contenant que des noms relatifs et des manifestes recalculés à partir de l’archive finale. CR08 doit également être reconstruit avec un manifeste cohérent couvrant exactement les fichiers remis. T28, T29 et CR06 nécessitent une autorisation produit écrite avant toute implémentation ou migration.

Le prochain lot technique autorisable est T31 seulement après clôture documentaire des prérequis et décision explicite sur l’ordre de reprise. Aucun merge vers `continuation-t00-t27-canonical`, aucune release et aucune activation de capability sensible ne sont effectués par ce rapport.

## Preuves indépendantes conservées

| Fichier | Contenu |
|---|---|
| `T28_INDEPENDENT_VALIDATION_RAW.log` | Manifestes, bundle, clone frais, Gitleaks, Go race/vet/build et limitations T28 |
| `T29_INDEPENDENT_VALIDATION_RAW.log` | Manifestes T29, bundle, clone frais, Gitleaks et qualification Go |
| `T30_INDEPENDENT_VALIDATION_RAW.log` | Manifestes T30, bundle, clone frais, Gitleaks, Gosec et replay global |
| `T30_GOSEC_COMPARISON.json` | Comparaison structurée baseline/head : 16 contre 16, zéro nouveau finding |
| `CR06_INDEPENDENT_VALIDATION_RAW.log` | Manifestes CR06, bundle, clone frais, Gitleaks et qualification Go |
| `CR08_INDEPENDENT_VALIDATION_RAW.log` | Manifeste CR08, bundle, clone frais, Gitleaks, qualification Go et garde native |

Les journaux ci-dessus sont attachés séparément afin que chaque verdict puisse être rejoué sans dépendre d’une déclaration d’agent.

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized "Dépôt public ForgeLocal"
[2]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/handover/t28-t42-new-session/docs/NEW_SESSION_HANDOVER_T28_T42.md "Document de passation T28–T42"
