# T28-FINAL-INDEPENDENT-EVIDENCE-REVIEW

## Décision

> `T28_INDEPENDENT_EVIDENCE_REVIEW_COMPLETE_PENDING_OWNER_ACCEPTANCE`

La revue indépendante finale des preuves T28 est complète. Elle ne prononce pas `T28_APPROVED_VERIFIABLE_LOCAL` : l’acceptation finale appartient au propriétaire. Aucun nouveau code métier, test métier ou runtime n’a été créé ou lancé pendant cette revue.

La branche vérifiée est [`feature/t28-local-extensions-controlled`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/feature/t28-local-extensions-controlled). Le HEAD distant a été découvert par interrogation GitHub, sans sélection manuelle : `15f3c4a125e5a587fb99cd0851b10369964db30b`, également matérialisé dans un clone neuf.

## 1. Lignée GitHub réconciliée

Le clone indépendant a été créé avec LFS smudge désactivé et blobs filtrés. Le clone shallow initial ne permettait pas de prouver tous les ancêtres ; ce diagnostic est conservé dans `T28_FINAL_INDEPENDENT_LINEAGE_INITIAL_RAW.log`. Une reprise avec `git fetch --unshallow --filter=blob:none` a ensuite produit les contrôles d’ancêtre suivants : historique→baseline `0`, baseline→implémentation `0`, implémentation→HEAD distant `0`, `fsck=0` et clone propre.

| Référence | Rôle | Parent | Fichiers représentatifs | Nature |
|---|---|---|---|---|
| `69411e65c880d168832a65fc8475cc97d562a9ad` | Baseline historique T00–T27 | `76551a98335a26b0e6ef312e04adeeb5bd1a6b2f` | `.gitattributes` et archive historique | Historique |
| `999374d99b7996504ba91e421850a2fe84afb78d` | Baseline V6 gelée | `65aaf76a13b6fd6c81b286af3a88043686336332` | `evidence/V6_FREEZE/V6_FREEZE_BASELINE_DISCOVERY_RAW.log` | Baseline |
| `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578` | Implémentation T28 | `9a15cc019a37383e8d8d8cf5fecd88f4b040e823` | `internal/extensions/*`, `internal/api/extensions.go`, tests, OpenAPI et contrat | Code métier T28 |
| `a8a014d361c8b364a8baf51b20f0e566231a138a` | Commit package historique demandé | `df9784ca8c06eddcbcd190163a13729a4be09748` | `T28_FINAL_DECISION.md` | Preuve/rapport |
| `6476a21313f14511b95e4037f782947fc5e96e83` | Publication package historique | `a8a014d361c8b364a8baf51b20f0e566231a138a` | ZIP, bundle, sidecars, hashes | Preuve |
| `2099ec17a5ae99c59afe31f467d437319639782e` | Qualification R1 | `8b96981e9ffae245307d2b5cb279d6013c6dc11e` | baseline/lignée/tests/scans R1, contrats documentaires | Preuve/documentation |
| `bedf630139a8d63ec80419e071f0401f09cd54e8` | Contenu package R1 | `2099ec17a5ae99c59afe31f467d437319639782e` | rapport et rapports Gitleaks/verify | Preuve |
| `20da1afaa3b5a01381ee7c23ce84559725c3af4f` | Publication ZIP/bundle R1 | `bedf630139a8d63ec80419e071f0401f09cd54e8` | `t28-evidence-qualification-r1.zip`, bundle, sidecars | Preuve |
| `37b69bd4ea8772f8138572149fed23dc962788b3` | Vérification publique R1 | `20da1afaa3b5a01381ee7c23ce84559725c3af4f` | `T28_R1_FINAL_PACKAGE_VERIFY_CURRENT_RAW.log`, rapport Gitleaks | Preuve |
| `15f3c4a125e5a587fb99cd0851b10369964db30b` | HEAD distant observé | `37b69bd4ea8772f8138572149fed23dc962788b3` | rapport de lignée final uniquement | Preuve/documentation |

Le diff exact `4f0f6201..15f3c4a` limité à `internal/` est vide. Les commits postérieurs à l’implémentation ne modifient donc pas le code métier T28. Le package R1 est volontairement un artefact de contenu attaché à `bedf630`; le HEAD GitHub courant `15f3c4a` ajoute seulement de la preuve et de la documentation. Cet écart de référence est documenté, non corrigé par recréation de l’artefact.

## 2. ZIP R1

L’artefact exact contrôlé depuis le clone public neuf est `evidence/T28/t28-evidence-qualification-r1.zip`. Son SHA-256 direct et son sidecar correspondent à :

`7be89af8e093b9b35c7924bc0ac3d8f0268cba9a2682db0d558f9b8573d9158d`

Le sidecar a passé le contrôle depuis le clone frais et depuis un répertoire neutre. `unzip -t` et l’extraction fraîche ont retourné `0`. Le manifeste est non auto-référentiel (`manifest_self_reference=NOT_FOUND`) et tous ses checksums internes ont retourné `0`. L’audit de contenu a trouvé `0` répertoire `.git`/`node_modules`, `0` nom de fichier DB/cookie/token/secret et aucun motif de clé privée, token GitHub, bearer ou JWT. Le ZIP n’a pas été remplacé.

## 3. Bundle R1

L’artefact exact contrôlé est `evidence/T28/t28-evidence-qualification-r1.delta.bundle`. Son SHA-256 direct et son sidecar correspondent à :

`442c1ff49b4f62b41edbe7cc2ee686fe8eaf15553b56e30ac39b4930e6a3944f`

`git bundle verify` a retourné `0`. Le bundle expose `bedf630139a8d63ec80419e071f0401f09cd54e8 refs/heads/feature/t28-local-extensions-controlled` et exige explicitement la baseline `999374d99b7996504ba91e421850a2fe84afb78d`. Le clone du bundle seul a échoué en code `128` pour cette raison attendue. Un clone seedé par la baseline a ensuite attaché le bundle, effectué le checkout explicite de la référence T28, obtenu le HEAD `bedf630`, passé `git fsck --full` avec code `0` et conservé un worktree propre.

## 4. Périmètre code et non-exécution

Depuis le clone neuf, la replay indépendante a passé `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'`, `go vet ./...`, `go build ./...` et `git diff --check`, tous avec code `0`. L’inspection confirme les protections bearer, loopback et origin guard ; les permissions sont conservées, les permissions sensibles et host patterns larges produisent `HIGH_RISK`, et l’approbation exige l’acknowledgement exact ainsi que `accept_high_risk` lorsque nécessaire.

Aucun navigateur, Chromium, Camoufox, runtime d’extension, proxy, téléchargement externe ou processus du package n’a été lancé. Un éventuel `update_url`/`updateURL` est ignoré, non suivi et non exécuté. T28 ne charge pas de package dans un navigateur et ne transforme pas le manifest en identité runtime.

## 5. Tests et scans livrés

| Contrôle | Résultat vérifié | Réserve |
|---|---|---|
| Tests globaux R1 baseline/HEAD | Les logs correspondent aux commits annoncés et indiquent code `0` des deux côtés | Le finding runtime historique n’a pas été reproduit ; aucune erreur identique ne doit être prétendue |
| Tests T28 ciblés | Replay indépendante code `0` sous race | Aucun runtime démarré |
| OSV v1.9.2 | Scan réel ; code `1`, 46 identifiants dans chaque `go.mod` baseline/head | Avis dépendances/std-lib conservés, non masqués |
| Gitleaks plage | 8 arbres de commits réellement inspectés, rapports vides et code `0` | L’ancien diagnostic `0 commits scanned` est conservé séparément |
| Gitleaks extraction | Scan `--no-git --redact` du ZIP et du package final, code `0`, aucun leak | Aucun faux PASS déduit d’un outil absent |
| Gosec | 6 findings baseline et 6 HEAD ; comparaison normalisée `new_findings=0`, `resolved_findings=0` | Findings historiques maintenus |

## 6. Gates et décision propriétaire

Les valeurs suivantes restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. T29, T39, T40, T41 et T42 ne commencent pas avant la décision finale du propriétaire sur T28. Aucun runtime, migration, SystemVault natif, proxy/cookie réel, production ou release n’a été exécuté.

La décision soumise au propriétaire est donc exactement :

> `T28_INDEPENDENT_EVIDENCE_REVIEW_COMPLETE_PENDING_OWNER_ACCEPTANCE`

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/feature/t28-local-extensions-controlled "Branche GitHub T28"
[2]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/feature/t28-local-extensions-controlled/evidence/T28/t28-evidence-qualification-r1.zip "ZIP de qualification R1"
[3]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/feature/t28-local-extensions-controlled/evidence/T28/t28-evidence-qualification-r1.delta.bundle "Bundle de qualification R1"
