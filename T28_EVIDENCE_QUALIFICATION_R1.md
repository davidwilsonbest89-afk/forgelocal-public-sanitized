# T28-EVIDENCE-QUALIFICATION-R1

## Verdict

> `T28_EVIDENCE_QUALIFICATION_R1_READY_FOR_INDEPENDENT_REVIEW`

Le lot T28 reste techniquement au statut `T28_IMPLEMENTED_VERIFIABLE_LOCAL_PENDING_INDEPENDENT_REVIEW`. Cette qualification porte uniquement sur la cohérence et la conservation des preuves. Elle ne prononce pas `T28_APPROVED_VERIFIABLE_LOCAL`, qui appartient à la revue indépendante des artefacts livrés.

## Baseline et lignée

La qualification utilise le dépôt public `forgelocal-public-sanitized`, la branche `feature/t28-local-extensions-controlled` et un clone R1 neuf, avec LFS smudge désactivé. Le tag `t00-t42-v6-local-qualified-2026-08-25` résout vers `999374d99b7996504ba91e421850a2fe84afb78d`. Le tag `t00-t27-complete-20260820` résout vers `72d54110c89583beacc556bb103f881b667d8137`, tandis que l’objet historique demandé est `69411e65c880d168832a65fc8475cc97d562a9ad`; le tag et l’objet sont contrôlés séparément dans le log brut. Après déshallow ciblé, `69411e6` est ancêtre de `999374d`, puis `999374d` est ancêtre du HEAD T28.

| Commit | Rôle | Parent | Fichiers/statut | Artefact associé |
|---|---|---|---|---|
| `69411e65c880d168832a65fc8475cc97d562a9ad` | Commit historique T00–T27 demandé | `76551a98335a26b0e6ef312e04adeeb5bd1a6b2f` | Archive historique | Packages T00–T27 |
| `999374d99b7996504ba91e421850a2fe84afb78d` | Baseline V6 gelée | `65aaf76a13b6fd6c81b286af3a88043686336332` | Baseline sans modification R1 | Tag V6 |
| `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578` | Commit d’implémentation cité par les scans | `9a15cc019a37383e8d8d8cf5fecd88f4b040e823` | Code T28 qualifié localement | Contenu T28 |
| `a8a014d361c8b364a8baf51b20f0e566231a138a` | Commit de contenu/package annoncé | `df9784ca8c06eddcbcd190163a13729a4be09748` | Rapport synchronisé et contenu source | ZIP/bundle T28 |
| `6476a21313f14511b95e4037f782947fc5e96e83` | Publication de preuve | `a8a014d361c8b364a8baf51b20f0e566231a138a` | ZIP, bundle, sidecars et hashes | ZIP `072eb921…3723b20`; bundle `60c78aff…b4adbdb` |
| `8b96981e9ffae245307d2b5cb279d6013c6dc11e` | HEAD distant qualifié | `6476a21313f14511b95e4037f782947fc5e96e83` | Journaux de vérification publique | Package inchangé |

Le diff exact `4f0f6201..8b96981` ne contient aucun fichier `internal/` modifié après le commit d’implémentation qualifié : uniquement documentation, rapports, scans, SBOM, package et journaux de preuve. Aucune modification métier n’a été introduite après `4f0f6201`.

## Tests Go globaux et ciblés

Deux worktrees propres ont exécuté exactement `go test -count=1 -race ./...` : baseline V6 (`999374d`) et HEAD T28 (`8b96981`). Les deux commandes ont retourné `corrected_exit=0`, tous les packages présents ont été testés et aucun nouveau finding n’est apparu. Le finding runtime historique `TestNewRegistryLoadsBrowseForgeChromiumFromDefaultConfig` n’a pas été reproduit dans cette exécution R1 ; aucune allowlist, skip, modification de test ou modification de code n’a été utilisée. Il n’est donc pas honnête de prétendre avoir démontré une erreur identique : la comparaison exacte est « baseline PASS / HEAD PASS / signatures FAIL absentes ».

Les tests ciblés `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'` ont retourné le code 0 sur le HEAD T28.

## Scans

| Scan | Résultat R1 | Décision |
|---|---|---|
| OSV Scanner v1.9.2 sur `go.mod` baseline/head | code 1, 46 identifiants uniques de chaque côté | Scan réel qualifié ; findings dépendances/std-lib non masqués |
| OSV Scanner v1.9.2 récursif sur HEAD | code 1, JSON produit | Scan réel conservé séparément des résultats ciblés |
| Gitleaks plage baseline→HEAD | 8 commits réels, arbres des chemins modifiés, code 0 pour chaque, aucun leak | Qualifié ; ancien diagnostic `0 commits scanned` conservé |
| Gitleaks extraction ZIP | `--no-git --redact`, aucun leak | Qualifié |
| Gosec baseline/head | 6 findings chacun ; `new_findings=0`, `resolved_findings=0` après normalisation règle/fichier/détail | Findings historiques identiques ; aucun finding T28 |

OSV v2.5.1/v2.4.0/v2.3.8 n’étaient pas compatibles avec Go 1.25.13 local ; cette limitation et l’installation réussie de v1.9.2 sont dans `OSV_R1_RAW.log` et `OSV_R1_HEAD_RAW.log`. Aucun résultat OSV n’a été fabriqué.

## Périmètre documentaire et non-exécution

La valeur exacte maintenue est `camoflox_execution_authorized=false`, avec `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `t08_authorized=false` et `release_authorized=false`. Un éventuel `update_url`/`updateURL` du manifest est ignoré, non suivi et non exécuté. T28 ne lance ni navigateur, ni extension, ni proxy, ni processus externe et ne télécharge aucun package.

## Conservation

Les preuves R1 comprennent baseline brute, lignée brute/table, tests baseline/head/ciblés, OSV JSON/raw, Gitleaks par commit et extraction, Gosec baseline/head/comparaison normalisée, registre, changelog, todo et package T28. Le package strict final a été généré depuis le commit de contenu `bedf630139a8d63ec80419e071f0401f09cd54e8`. Le HEAD de publication GitHub de la branche est `37b69bd4ea8772f8138572149fed23dc962788b3` ; ses commits postérieurs au contenu package ne contiennent que les journaux de vérification R1.

| Artefact | SHA-256 | Vérification |
|---|---|---|
| `evidence/T28/t28-evidence-qualification-r1.zip` | `7be89af8e093b9b35c7924bc0ac3d8f0268cba9a2682db0d558f9b8573d9158d` | sidecar distribué + dossier neutre, `unzip -t`, extraction et manifeste/checksums |
| `evidence/T28/t28-evidence-qualification-r1.delta.bundle` | `442c1ff49b4f62b41edbe7cc2ee686fe8eaf15553b56e30ac39b4930e6a3944f` | sidecar distribué + dossier neutre, `git bundle verify`, clone bundle seedé, checkout explicite et fsck |

La vérification finale a retourné code 0 pour les sidecars, `unzip -t`, extraction, manifeste non auto-référentiel, 70 checksums internes, absence de `.git`/`node_modules`/noms de DB-cookie-token-secret et absence de motifs de payloads secrets. Le clone bundle seul a échoué en code 128 comme prévu faute du prerequisite baseline ; le clone seedé puis l’attachement du bundle ont réussi avec checkout de la ref T28, `fsck=0` et worktree propre. Gitleaks sur l’extraction finale a retourné code 0 avec rapport vide. Les logs bruts sont `T28_R1_FINAL_PACKAGE_VERIFY_RAW.log`, `T28_R1_FINAL_PACKAGE_VERIFY_INITIAL_RAW.log` et `T28_R1_FINAL_PACKAGE_VERIFY_CURRENT_RAW.log` ; le dernier contrôle a vérifié le HEAD public `20da1afaa3b5a01381ee7c23ce84559725c3af4f` et le package source `bedf630139a8d63ec80419e071f0401f09cd54e8`. Le rapport Gitleaks correspondant est `GITLEAKS_R1_FINAL_PACKAGE_CURRENT.json`.

## Gates et arrêt

Aucune gate n’est levée. T29, T39, T40, T41 et T42 ne sont pas démarrés. La prochaine action autorisée est une revue indépendante des artefacts T28 ; aucun runtime navigateur, Camoufox, Chromium, extension, proxy réel, cookie réel, SystemVault natif, migration utilisateur ou release ne doit être exécuté dans cette mission.
