# Rapport consolidé — réimplémentation clean-room T12–T16/G15

> **Statut de dossier : `REIMPLEMENTATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`.** Ce document décrit uniquement la nouvelle lignée contrôlée `forgelocal-baseline-2026-08-17`. Il ne prétend ni restaurer ni remplacer le snapshot historique T17, qui demeure non récupérable comme source complet.

| Champ | Valeur vérifiée |
|---|---|
| 1. Identifiant | `REIMPL-T12-T16-G15-R3` |
| 2. Date de contrôle | 2026-08-18 |
| 3. Dépôt source | `boucheriechefimane-cmd/IPcache` |
| 4. Baseline contrôlée | `forgelocal-baseline-2026-08-17` |
| 5. Commit source antérieur au delta | `8c1766b4113ccbe8f3886dbeaf1c300f18de8584` (`t15-reimpl-2026-08-18`) |
| 6. Périmètre | T12 LocalVault, T13 Environment, T14 Runtime, T15 Automation locale, G15-A/B et synchronisation T16 dashboard |
| 7. Statut technique | Implémentation et contrôles terminés ; revue indépendante encore requise |
| 8. État de release | `PUBLIC_RELEASE_BLOCKED` maintenu |

## 9. IMPLEMENTED

La lignée clean-room contient **T12 LocalVault**, avec chiffrement AES-256-GCM, AAD, journal durable et permissions restrictives ; **T13 Environment**, avec diagnostics projetés et redacted ; **T14 Runtime**, avec qualification Chromium locale, machine d’états et persistance SQLite ; et **T15**, avec validation URL fail-closed (`file://` et loopback seulement) avant effet de bord. Les jalons T12, T13, T14 et T15 sont référencés par les tags immuables respectifs dans `metadata/GIT_METADATA.log`.

Le delta final rend les projections T13/T14 effectivement accessibles au dashboard uniquement par le groupe administrateur authentifié. Il ajoute la qualification au démarrage du Core et ne projette que l’identifiant, l’état, la version et la date de qualification. Les chemins de binaires, ports de débogage et hashes d’intégrité restent confinés au Core. Le dashboard synchronisé normalise les états du Core et ne tente plus d’afficher une architecture absente du contrat redacted.

| Sous-lot | Implémentation constatée |
|---|---|
| T12 | LocalVault AES-256-GCM/AAD, journal JSONL fsync, fail-closed, permissions `0700/0600` |
| T13 | Diagnostic projeté, refus explicite profil inconnu, concurrence batch, endpoint admin-only |
| T14 | Découverte/sonde headless, version, état SQLite, catalogue API minimal redacted |
| T15 | URL locale fail-closed, session CDP locale, timeout/annulation/idempotence existants, code lecture seule vérifié par E2E |
| G15-A | Projection runtime sans chemin, port ni hash |
| G15-B | CORS/origine boucle locale et garde des mutations `POST`/`PUT`/`PATCH`/`DELETE` |
| G15-C | Verrous par profil et contention couverts par la suite Core existante sous `-race` ; aucun portage de runtime réel n’a été entrepris |
| T16 | Sources React synchronisées dans `forge-dashboard/`, contrat Core redacted réconcilié, E2E complète exécutée |

## 10. TESTED

Les journaux bruts ne contiennent aucun secret et sont emballés dans l’archive de preuves. Ils établissent les résultats suivants.

| Commande / contrôle | Résultat |
|---|---|
| `go test -count=1 -race ./...` | **PASS** ; packages Core testés, aucune course signalée |
| `go vet ./...` | **PASS** |
| `go build ./...` | **PASS** |
| `pnpm exec tsc --noEmit` | **PASS** |
| `pnpm run build` | **PASS** ; avertissements Vite non bloquants sur taille de chunk et résolution runtime d’un asset manus-storage |
| `npx playwright test --workers=1` | **PASS : 22 passed (10.9m)** |
| T14 isolé | **PASS : 3 passed (4.6s)** |
| `git diff --check` | **PASS** |

## 11. VERIFIED

Le test navigateur complet effectue les opérations dashboard/Core via boucle locale et inclut les scénarios T05, T06, T09, T10, T11, T13, T14 et T15. Son journal final indique exactement `22 passed`; les résultats antérieurs échoués ne sont pas utilisés comme preuve de passage.

La politique T15 est vérifiée avant l’accès à une session : une URL externe est refusée avec une erreur machine-readable, tandis que `file://` et les hôtes loopback font partie du domaine local admis. La garde G15-B exige aussi une origine ou un référent loopback pour les mutations, sans assouplir les lectures, `HEAD` ni `OPTIONS`.

## 12. SCANS DE SÉCURITÉ

| Contrôle | Résultat et interprétation |
|---|---|
| Gitleaks delta | **PASS, 0 fuite** après exclusion des métadonnées de plateforme non versionnables ; voir `scans/GITLEAKS_DELTA.json` (`[]`). |
| Gosec packages concernés | Outil installé puis exécuté sur `./cmd/server ./internal/api`. Il relève **72 findings** dans le périmètre total des packages, dont des zones historiques de CLI/backup/session. |
| Gosec sur lignes ajoutées au delta | **0 finding** : le rapprochement automatique entre le diff `HEAD` et `GOSEC_SCANNED_PACKAGES.json` est archivé dans `scans/GOSEC_CURRENT_DELTA.json`. |
| Statut global des scans | `SCAN_BLOCKED_UNKNOWN` **reste obligatoire** : zéro finding sur ce delta n’efface pas les 72 findings historiques. |

## 13. EVIDENCE_ARCHIVED

L’archive de preuves comprend les journaux de contrôle Core et frontend, le journal E2E complet, les résultats JSON des scanners, les métadonnées Git, le patch du delta et les documents de contexte de baseline. Les données temporaires, le jeton E2E, les navigateurs, les dépendances `node_modules`, les artefacts de build et les rapports HTML Playwright sont exclus.

## 14. MANIFESTE ET REPRODUCTIBILITÉ

Le manifeste `SHA256SUMS` est généré **après** la constitution des contenus et n’inclut pas son propre fichier. Le hash final de l’archive est déposé dans un fichier compagnon `forgelocal-reimplementation-evidence-r3.zip.sha256`, hors de l’archive, afin d’éviter toute auto-référence. La reproduction exige Go 1.25.13, les dépendances dashboard verrouillées, un Chromium local qualifiable et les navigateurs Playwright installés.

## 15. BLOCKED

Les restrictions permanentes ne changent pas : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Le lot T18 conserve séparément ses statuts `T18_TARGETED_IMPLEMENTATION_VERIFIED` pour `internal/launch` et `T18_BLOCKED_BASELINE_SOURCE_NOT_FOUND` pour la validation globale historique.

## 16. NOT_IN_SCOPE

Ce dossier n’effectue ni lancement Camoufox, ni proxy réseau réel, ni qualification SystemVault native Ubuntu, ni publication de release, ni restauration artificielle de T17. Il ne convertit pas la validation locale en autorisation de pilote ou de release publique.

## 17. CONDITIONS DE CLÔTURE

La prochaine étape autorisée est une **revue indépendante** de l’archive et de son manifeste. Une fois cette revue positive, le code peut être enregistré comme réimplémentation vérifiée sur cette baseline uniquement. Toute évolution SystemVault native, Camoufox, proxy réel ou release exige une instruction distincte et les gates publics encore manquants.

## Références locales

- [1] `metadata/GIT_METADATA.log`
- [2] `logs/CORE_GLOBAL_CHECKS.log`
- [3] `logs/DASHBOARD_BUILD_CHECKS.log`
- [4] `logs/PLAYWRIGHT_FULL_22_22.log`
- [5] `scans/GITLEAKS_DELTA.json`
- [6] `scans/GOSEC_CURRENT_DELTA.json`
