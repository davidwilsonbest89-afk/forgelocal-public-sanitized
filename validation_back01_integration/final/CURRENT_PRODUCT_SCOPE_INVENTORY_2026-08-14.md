# Inventaire de périmètre produit — 2026-08-14

## Constat sur la documentation héritée

Les documents `docs/plan.md` et `docs/wbs.md` sont des documents historiques BrowseForge. Ils indiquent explicitement qu’ils sont archivés et qu’ils décrivent une approche initiale centrée sur Camoufox, les containers Firefox, une WebExtension, Playwright et des promesses de compatibilité Windows/macOS.

Cette approche ne doit pas être utilisée comme cahier des charges de release ForgeLocal BACK-01. Elle est incompatible avec les décisions actuelles suivantes :

| Sujet | Périmètre ForgeLocal actuel |
|---|---|
| Release en qualification | Core/API BACK-01 minimal uniquement. |
| Source de vérité | BrowseForge Core Go unique, local-first. |
| Runtime de la chaîne RC | Chromium `151.0.7922.108-1xtradeb1.2404.1` externe à l’artefact. |
| Cible annoncée | Ubuntu 24.04.4 LTS amd64 uniquement. |
| Camoufox | Runtime candidat produit ultérieur, non inclus dans BACK-01 et interdit pendant la qualification RC. |
| GUI React / Tauri | Chantiers distincts, non livrés par l’artefact minimal BACK-01. |
| Publication publique | `PUBLIC_RELEASE_BLOCKED` tant que les cinq gates ne sont pas tous `PASSED` avec preuve fraîche et revue indépendante. |

## Tests et contrôles déjà identifiés comme reproductibles dans le sandbox

1. `go test ./...` pour la suite Go complète.
2. Test API BACK-01 ciblé : `TestBackupV1CreateModifyRestoreIsolation`.
3. Validateur de traçabilité : `scripts/validate-release-traceability.py`.
4. Validateur de gate public : `scripts/check-public-release-gate.py`.
5. Test négatif de métadonnées de gate : `scripts/test-public-release-gate-metadata.py`.
6. Contrôle de refus SystemVault du script dans le conteneur ; ce contrôle confirme la sécurité du refus, mais ne qualifie pas le coffre natif.
7. Intégrité de l’archive RC et des documents de release via SHA-256.

## Limite non substituable

Le test SystemVault natif et le test anti-fuite intégré ne peuvent pas être promus depuis le sandbox conteneurisé. Ils exigent une session graphique Ubuntu 24.04.4 amd64 hors conteneur, sans `sudo`, avec Secret Service déverrouillé. La préparation et les tests locaux peuvent progresser, mais les gates publics correspondants restent `PENDING`.

## Action à réaliser

Créer un cahier des charges ForgeLocal distinct, en français, qui sépare :

- exigences livrées et testables pour BACK-01 ;
- exigences de sécurité et gates publics ;
- exigences produit ultérieures (SQLite métier, dashboard React, connecteurs proxy, runtimes additionnels, Tauri) ;
- exigences explicitement hors périmètre ou interdites pendant la qualification.

## Gates publics applicables au candidat RC

| Gate | Décision actuelle | Preuve manquante ou exigée |
|---|---|---|
| `SYSTEMVAULT_NATIVE_PER_TARGET` | `PENDING` | Matrice native réussie sur Ubuntu 24.04.4 LTS amd64, sans sudo, conteneur ni fallback en clair. |
| `SYSTEMVAULT_ANTI_LEAK_INTEGRATED_FLOW` | `PENDING` | `systemvault-anti-leak.json` provenant d’un flux réel profil → backup chiffré → restauration isolée, sans révéler la sentinelle. |
| `MAINTAINER_MANIFEST_SIGNATURE` | `PENDING` | Signature détachée, clé publique publiée séparément, empreinte approuvée et vérification indépendante. |
| `RUNTIME_LICENSE_AND_REDISTRIBUTION_REVIEW` | `PENDING` | Revue licence/redistribution des paquets Chromium exacts du candidat. |
| `OS_COMPATIBILITY_EVIDENCE` | `PENDING` | Matrice limitée aux configurations disposant des preuves runtime et SystemVault natives complètes. |

Une décision `PASSED` est invalide si le commit, le hash de l’artefact, la version du runtime, la cible OS/architecture ou le hash de preuve change. Une preuve ne peut jamais être réutilisée entre deux chaînes runtime/artefact/commit/configuration distinctes.

## Portée release actuellement permise

Le paquet BACK-01 est seulement un Core/API local minimal. Il inclut API locale authentifiée, backups chiffrés, restauration isolée, métadonnées SQLite, audit local et coffre système sous gate. Chromium est externe à l’artefact et utilisé uniquement pendant la QA. Camoufox, le dashboard React, Tauri, migration métier JSON→SQLite, fingerprinting, humanization, MCP, extensions et workflows sont exclus de la release RC.

L’unique cible à qualifier est Ubuntu 24.04.4 LTS amd64. Le sandbox permet les tests Core/API et runtime E2E, mais son caractère conteneurisé interdit de le présenter comme preuve SystemVault native.

> La transition à `PUBLIC_RELEASE_APPROVED` exige les cinq gates `PASSED`, des preuves fraîches et versionnées, ainsi qu’une revue indépendante.

## Résultats de vérification reproductible du sandbox

| Contrôle | Résultat | Interprétation |
|---|---:|---|
| SHA-256 de l’archive RC | Conforme | `553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6` est inchangé. |
| Tests API BACK-01 sélectionnés | 3 tests détectés | Le test d’acceptation de création → backup → modification → restauration isolée → relance runtime est présent. |
| Test ciblé `TestBackupV1CreateModifyRestoreIsolation` | PASS | Le flux API BACK-01 et la relance contrôlée de Chromium 151.0.7922.108 sont verts dans le sandbox. |
| `go test ./... -count=1` | PASS | Tous les paquets de test Go actuels sont verts. |
| `go vet ./...` | PASS | Aucun diagnostic de `go vet` n’a été remonté. |
| `go test -race ./internal/api ./internal/backup ./internal/profile -count=1` | PASS | Aucun détecteur de course dans les modules ciblés. |
| Test négatif de métadonnées de gate | PASS | Un gate `PASSED` incomplet est rejeté. |
| Validateur de traçabilité | PASS | Les deux chaînes de preuves restent distinctes et cohérentes. |
| Validateur public | PASS avec décision bloquée | Le validateur confirme `PUBLIC_RELEASE_BLOCKED` car les cinq gates restent `PENDING`. |
| Gitleaks (rapport redigé) | 0 fuite détectée | Aucun secret n’a été trouvé dans l’historique inspecté par le scan actuel. |
| Gosec | 189 résultats, sortie 1 attendue | Dette de sécurité à classer et à réduire ; ce résultat ne doit pas être ignoré. |

## Couverture et dette de tests

La couverture n’est pas homogène. Les modules directement liés à BACK-01 sont partiellement couverts (`internal/backup` 58,0 %, `internal/api` 35,1 %, `internal/browser` 56,5 %, `internal/profile` 70,3 %, `internal/runtime` 86,2 %). Les modules hors portée BACK-01 présentent des couvertures très faibles ou nulles, notamment `internal/secrets` 0 %, `internal/humanize` 0 %, `internal/mcp` 23,7 %, `internal/fingerprint` 19,3 % et `internal/workflow` 20,1 %.

Le cahier des charges doit donc imposer une couverture ciblée sur les flux de release plutôt qu’un seuil global artificiel qui serait gonflé par des modules hors périmètre. Les lignes de commande et le docteur SystemVault doivent gagner des tests propres avant une extension de périmètre.

## Classification initiale Gosec

Le scan actuel contient 189 résultats : 64 `G104` bas, 22 `G304` moyens, 32 `G301` moyens, 17 `G404` élevés, 14 `G703` élevés, 7 `G704` élevés, 4 `G115` élevés, 1 `G122` élevé et le solde réparti entre `G107`, `G110`, `G112`, `G204`, `G302`, `G305`, `G306` et `G705` moyens.

Les alertes doivent être traitées par zone de menace et non seulement par sévérité de l’outil : téléchargement de runtime et chemins de fichiers (`internal/browser`), exécution de commandes et extensions CLI (`cmd/server`), exposition de surfaces API, permissions de fichiers et sources d’aléa. Les modules `humanize`, MCP, workflows et Camoufox historiques sont hors BACK-01 ; ils doivent être isolés du build minimal et classés comme dette héritée, sans prétendre qu’ils sont validés par le candidat RC.

## Alerte de scan de l’archive candidate — à trier avant toute utilisation supplémentaire

Le scan Gitleaks sans Git de l’archive RC extraite a retourné une occurrence `generic-api-key` dans `evidence/RUNTIME_PROVENANCE.out`, ligne 8. Aucun contenu détecté n’a été affiché ni copié dans ce document. La chaîne a une longueur de 64 caractères, est référencée par une preuve de provenance historique, et n’est pas une simple empreinte SHA-256 hexadécimale. Son statut doit donc rester **SUSPECTED_SECRET_OR_FALSE_POSITIVE** tant qu’une revue mainteneur, hors diffusion, n’a pas établi sa nature.

Conséquences immédiates :

1. Ne pas modifier ni redistribuer l’archive RC figée.
2. Ne pas exposer la ligne, la chaîne ni le fichier de rapport non redigé.
3. Suspendre l’usage opérationnel du pilote local par précaution jusqu’à classification formelle.
4. Créer une preuve de triage redigée indiquant le chemin, le type de règle, la version du scanner, la décision et, si faux positif confirmé, une exception bornée au chemin et à la règle. Toute exception doit être vérifiée par un scanner redigé et une revue indépendante.
5. Si la chaîne est confirmée comme secret, invalider le candidat, révoquer la valeur concernée, générer un nouvel artefact et recommencer toutes les preuves dont la fraîcheur dépend de son hash.

Cette alerte ne change pas la décision déjà bloquée `PUBLIC_RELEASE_BLOCKED`, mais elle interdit de présenter le scan de l’archive comme vert jusqu’à résolution.
