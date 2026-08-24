# ForgeLocal — classification des findings de la checklist finale

**Date d’exécution des contrôles :** 2026-08-24 (UTC)
**Clone et commit contrôlés :** clone neuf `/home/ubuntu/forgelocal-prehuman-fresh-20260824`, HEAD `6ae02e4ceed239b9310fbf3fccb1b5170117251e`
**Baseline de comparaison :** tag `t00-t27-complete-20260820`, commit `69411e65c880d168832a65fc8475cc97d562a9ad`
**Nature de cet ajout :** preuves et documentation uniquement ; aucun fichier de code produit, test métier, gate, runtime, secret ou release n’a été ajouté ou modifié dans cet addendum.

## Décision de lecture

La checklist finale **n’est pas un passage technique global**. Les contrôles déterministes qui ont retourné `exit_code=0` sont conservés comme réussis. Les outils qui ont trouvé des éléments sont conservés avec leur code de sortie réel et une qualification explicite. Les gates et les absences de configuration protégée empêchent toute requalification implicite en PASS.

| Domaine | Résultat brut | Classification stricte | Conséquence |
|---|---:|---|---|
| Intégrité Git, LFS, sidecars, ZIP et bundle de la livraison gelée | Contrôles antérieurs réussis dans le paquet original | `APPROVED_VERIFIABLE_LOCAL` pour la chaîne d’artefacts | Ne vaut pas autorisation produit ou release. |
| `go mod verify`, `go list`, tests race, `go vet`, `go build` | `exit_code=0` | `PASS_APPLICABLE` | Aucun échec détecté par ces contrôles. |
| `govulncheck` et OSV | `exit_code=0`, aucun problème signalé | `PASS_APPLICABLE_NO_FINDINGS` | Aucun avis de vulnérabilité détecté dans ces analyses. |
| Syft CycloneDX et SPDX | `exit_code=0` | `PASS_ARTIFACT_GENERATED` | Les deux SBOM sont conservés séparément. |
| Dashboard `pnpm install --frozen-lockfile`, TypeScript, build, audit | `exit_code=0` | `PASS_APPLICABLE` | Aucun échec détecté sur ces contrôles Dashboard. |
| Inventaire de licences production Dashboard | `exit_code=0`, JSON valide | `PASS_INVENTORY_GENERATED` | L’inventaire complet est joint ; il ne constitue pas une décision juridique de compatibilité. |
| Trivy baseline/head | 6 findings contre 6, `new=0`, `resolved=0` | `BLOCKED_BY_EXISTING_MISCONFIGURATIONS` | Six misconfigurations Docker historiques restent ouvertes ; pas de vulnérabilité Go/pnpm ni secret signalé dans ce rapport. |
| Gosec baseline/head | Scanner `exit_code=1`; 194 contre 194 ; `new=0`, `resolved=0` | `BLOCKED_BY_EXISTING_SECURITY_FINDINGS` | Les 194 findings existants restent ouverts ; aucune nouveauté différentielle. |
| Staticcheck | 36 contre 36 ; `new=0`, `resolved=0`; scanner `exit_code=1` | `BLOCKED_BY_EXISTING_QUALITY_FINDINGS` | Findings qualité présents et non corrigés, conformément à l’interdiction de modifier le produit. |
| GolangCI-Lint | 82 baseline, 83 head ; 13 nouveaux, 12 résolus ; scanner `exit_code=1` | `NOT_APPROVED_PENDING_REMEDIATION` | Le différentiel contient 13 findings head non présents dans la baseline ; aucun PASS ne doit être déclaré. |
| Gitleaks cumulatif baseline..HEAD | `exit_code=1`, marqueur historique `APi=REDACTED` | `SCAN_BLOCKED_UNKNOWN` | Ce signal n’est pas présenté comme un PASS et maintient la gate correspondante. |
| Playwright/T10 | Configuration synthétique protégée absente : `FORGELOCAL_CORE_BASE_URL` et token/config requis | `NOT_APPLICABLE_UNDER_CURRENT_GATES` / `BLOCKED_BY_REQUIRED_PROTECTED_CONFIGURATION` | Aucun Core runtime, token, proxy réel, cookie réel ou navigateur réel n’a été lancé. |

## Findings explicitement retenus

### Gitleaks

Le scan no-git du nouveau dossier d’addendum retourne `exit_code=0` avec un rapport vide `[]` [6]. Cela signifie uniquement qu’aucun nouveau finding n’a été détecté dans les fichiers ajoutés par cet addendum ; ce résultat local ne neutralise pas le signal cumulatif historique.

Le scan cumulatif sur la plage baseline..HEAD conserve exactement un signal générique historique, `APi=REDACTED`, avec `RuleID=generic-api-key` et `exit_code=1`. La valeur est déjà rédactée dans la preuve. Ce résultat est classé **`SCAN_BLOCKED_UNKNOWN`**, et non `PASS`; il est conservé dans `PREHUMAN_FINAL_GITLEAKS.json` [1].

### Gosec

Les sorties scanner baseline et head contiennent chacune **194 findings** et le scanner retourne `exit_code=1`. Le normalisateur corrigé, qui utilise les bons noms de fichiers et ramène les chemins aux chemins relatifs, établit `baseline=194`, `head=194`, `new_findings=0` et `resolved_findings=0` [2]. Les findings existants n’ont pas été supprimés ou minimisés, car toute correction du code produit était hors périmètre.

### Trivy

Trivy produit six misconfigurations dans la baseline et les six mêmes dans HEAD : trois sur `Dockerfile` et trois sur `docker/Dockerfile.run`. Le différentiel normalisé est `new_findings=0` et `resolved_findings=0` [3]. Les règles sont les suivantes, pour chacun des deux fichiers : `DS-0002` (**HIGH**, utilisateur d’image root), `DS-0026` (**LOW**, absence de `HEALTHCHECK`) et `DS-0029` (**HIGH**, `apt-get` sans `--no-install-recommends`). Elles sont classées **historiques/préexistantes** et restent ouvertes ; aucun correctif produit n’a été introduit. Les rapports ne signalent aucune vulnérabilité Go/pnpm ou secret de dépendance dans cette exécution.

### Staticcheck et GolangCI-Lint

Staticcheck retourne **36 findings en baseline et 36 en HEAD**, avec zéro nouveau et zéro résolu après normalisation des chemins [4]. Le scanner retourne `exit_code=1`; le résultat est donc un blocage qualité existant, pas un PASS.

GolangCI-Lint a été reconstruit avec la version compatible Go 1.25.13 (`v1.64.8`) puis exécuté en JSON. Il retourne 82 findings baseline, 83 findings HEAD, **13 nouveaux** et **12 résolus** selon la comparaison chemin/linter/ligne/message [5]. Les 13 findings nouveaux sont tous conservés dans la sortie normalisée ; ils comprennent notamment des retours non vérifiés `errcheck` dans `cmd/server/main.go`, `internal/api/sessions.go` et `internal/backup/store.go`, ainsi qu’un `staticcheck` `SA9003` dans `internal/sessiontrack/tracker_test.go`. Cette différence est classée **`NOT_APPROVED_PENDING_REMEDIATION`**. Aucun finding n’a été masqué et aucun code produit n’a été changé pour améliorer artificiellement le résultat.

### Playwright et configuration protégée

Le harness Playwright n’a pas atteint un test métier : il s’est arrêté sur `CONFIGURATION_T10_ABSENTE`, car l’environnement autorisé ne fournissait ni `FORGELOCAL_CORE_BASE_URL` ni la configuration/token requis. Ce contrôle est donc **non applicable sous les gates courantes et bloqué par configuration protégée manquante**. Le Dashboard a néanmoins réussi ses contrôles d’installation gelée, de typage, de build et d’audit. Il serait incorrect de transformer l’absence de runtime et de credentials en réussite Playwright.

## Gates maintenues

Les valeurs suivantes restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Aucun runtime réel, Camoufox, proxy réel, cookie réel, SystemVault natif, migration, géolocalisation réelle ou release n’a été exécuté.

> **Verdict technique strict :** `READY_FOR_INDEPENDENT_REVIEW_WITH_OPEN_EXISTING_FINDINGS`.
>
> **Verdict de livraison conservé :** `T00_T42_PREHUMAN_VALIDATION_DELIVERED_PENDING_INDEPENDENT_REVIEW`.
>
> Ces deux formulations signifient que la chaîne de preuves est livrée pour revue humaine avec ses limites exposées. Elles ne signifient ni release readiness, ni approbation produit, ni levée de gate.

## Fichiers sources de preuve

Les fichiers bruts et structurés joints dans ce dossier sont les sources de vérité de cette classification. Le fichier `PREHUMAN_FINAL_EXIT_CHECKLIST_RAW.log` contient les commandes complètes, UTC, CWD, HEAD, sorties et codes de sortie de la checklist principale. Les normalisations qualité, Gosec et Trivy sont jointes séparément afin que les différences ne soient pas déduites d’un résumé textuel.

## Références

[1]: ./PREHUMAN_FINAL_GITLEAKS.json "Rapport Gitleaks final"
[2]: ./PREHUMAN_FINAL_GOSEC_BASELINE.json "Gosec baseline" ; ./PREHUMAN_FINAL_GOSEC_HEAD.json "Gosec head" ; ./PREHUMAN_FINAL_GOSEC_NORMALIZED.log "Différentiel Gosec normalisé"
[3]: ./PREHUMAN_FINAL_TRIVY_BASELINE.json "Trivy baseline" ; ./PREHUMAN_FINAL_TRIVY_HEAD.json "Trivy head" ; ./PREHUMAN_FINAL_TRIVY_NORMALIZED.log "Différentiel Trivy normalisé"
[4]: ./PREHUMAN_FINAL_STATICCHECK_BASELINE.jsonl "Staticcheck baseline" ; ./PREHUMAN_FINAL_STATICCHECK_HEAD.jsonl "Staticcheck head" ; ./PREHUMAN_FINAL_QUALITY_NORMALIZED.log "Différentiel qualité normalisé"
[5]: ./PREHUMAN_FINAL_GOLANGCI_BASELINE.json "GolangCI-Lint baseline" ; ./PREHUMAN_FINAL_GOLANGCI_HEAD.json "GolangCI-Lint head" ; ./PREHUMAN_FINAL_QUALITY_NORMALIZED.log "Différentiel qualité normalisé"
[6]: ./PREHUMAN_FINAL_GITLEAKS_ADDENDUM.json "Scan Gitleaks no-git de l’addendum"
