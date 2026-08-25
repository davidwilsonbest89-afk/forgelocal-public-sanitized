# T28-POSTFIX-FINAL-CLOSURE

## Décision

> `T28_APPROVED_VERIFIABLE_LOCAL`

La requalification post-correctif T28 est terminée et cohérente. T28 est approuvé uniquement comme **Core local contrôlé vérifiable localement** et ses nouveaux artefacts post-correctif sont publiés. Cette décision ne couvre aucun runtime navigateur, extension chargée ou exécutée, Camoufox/Chromium, SystemVault natif, proxy/cookies réels, migration, production ou release publique.

## Publication et lignée

La branche GitHub est [`feature/t28-local-extensions-controlled`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/feature/t28-local-extensions-controlled). Le HEAD réellement retourné par GitHub lors de la vérification fraîche est `dff092cd5eeb27ddff6e331c22594a03c5b4e1a7`.

| Référence | Rôle |
|---|---|
| `999374d99b7996504ba91e421850a2fe84afb78d` | Baseline V6 |
| `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578` | Implémentation T28 initiale |
| `f0701da849ce0f9073397bb42ded5e2e76b29ef1` | Commit identifiable du correctif d’intégrité taille/SHA-256 |
| `e806d4e915d1b702362389411d8ac823551df044` | Commit des tests de régression Approve/Assign/Rollback |
| `66f3bf09d3139e22f1885f7336c9edd879a32ede` | Commit source du nouveau package post-correctif |
| `dff092cd5eeb27ddff6e331c22594a03c5b4e1a7` | HEAD GitHub final, incluant ZIP/bundle/sidecars |

Le diff post-implémentation limité à `internal/` est attendu et documenté : il contient uniquement le correctif d’intégrité dans `internal/extensions/store.go`, son mapping API dans `internal/api/extensions.go` et les régressions dans `internal/extensions/store_test.go`. Aucun autre lot produit n’a été modifié.

## Correctif d’intégrité et tests

Le défaut concret était reproductible : un blob ZIP géré pouvait être modifié après import, puis passer `Approve`, `Assign` ou `Rollback` sans vérification du digest stocké. `verifyBlobIntegrity` compare désormais la taille et le SHA-256 avant chacune de ces transitions. L’API expose l’échec sous le code stable `INTEGRITY_MISMATCH`.

Le test `TestT28RejectsModifiedPackageBeforeApproveAssignAndRollback` démontre les trois refus et vérifie l’absence d’état partiel : la version reste `imported` avant approval, aucune affectation ni activation n’est créée avant assignation, et le rollback n’altère pas la version active. Les tests supplémentaires couvrent ZIP corrompu, limites, manifest absent/invalide, traversal, symlink, permissions sensibles, host patterns larges, `update_url`, purge, révocation et profil inexistant.

Depuis le code publié au commit `e806d4e`, la commande exacte `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'` a retourné `0`, ainsi que `go vet ./internal/extensions ./internal/api`, `go build ./...` et `git diff --check`. Le journal de requalification est `evidence/T28/T28_POSTFIX_REQUALIFICATION_RAW.log`.

## Nouveaux artefacts post-correctif

Le ZIP n’est pas le ZIP R1 réutilisé. Le nouveau fichier est `evidence/T28/t28-postfix-evidence-v1.zip`, généré depuis le snapshot corrigé `66f3bf0`. Son SHA-256 vérifié depuis le dossier distribué, un dossier neutre et le clone GitHub frais est :

`4efda01771ed7af135769dfa68caa8bdc6f226ca7cad5bf894dcfce05f5c8923`

Le nouveau bundle est `evidence/T28/t28-postfix-evidence.delta.bundle`. Son SHA-256 vérifié dans les mêmes conditions est :

`e9f65a5b9a734933f20ecf13b05f73136e604dc006b50dfc0d40286d91262097`

`unzip -t`, extraction fraîche, manifeste non auto-référentiel, checksums internes, sidecars portable frais/neutre et `git bundle verify` ont retourné `0`. Le bundle exige la baseline V6, puis a été chargé dans un clone seedé, checkouté explicitement au commit source corrigé `66f3bf0`, et vérifié avec `git fsck --full=0` et worktree propre. Le clone du bundle seul n’est pas utilisé comme preuve autonome car il s’agit d’un delta exigeant sa baseline.

## Gitleaks et réserves de scans

L’extraction du nouveau ZIP a passé Gitleaks `--no-git --redact` avec code `0` et aucun leak. Le scan initial avec `--log-opts` a retourné `0 commits scanned`, diagnostic conservé et non présenté comme validation de plage. La méthode corrective non vide a extrait les 15 fichiers réellement modifiés entre le parent du correctif et le commit de régression, puis Gitleaks a retourné code `0` et aucun leak. Les logs et JSON sont distincts.

Les avis OSV v1.9.2 déjà obtenus restent à traiter séparément : 46 identifiants sur chaque `go.mod`, code `1`. Les findings Gosec historiques restent conservés ; la correction locale d’ouverture de blob est explicitement annotée et le Gosec ciblé post-correctif reste documenté. Aucun avis ou finding n’est masqué.

## Gates et arrêt

Les gates restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Aucun navigateur, runtime d’extension, proxy réel, cookie réel, Camoufox, SystemVault natif, téléchargement externe, migration utilisateur, production ou release n’a été lancé.

T29, T39, T40, T41 et T42 ne sont pas démarrés. L’approbation présente ne signifie pas que le runtime navigateur est qualifié ; elle clôt uniquement le périmètre Core local T28 et ses preuves post-correctif.

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/feature/t28-local-extensions-controlled "Branche GitHub T28"
[2]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/feature/t28-local-extensions-controlled/evidence/T28/t28-postfix-evidence-v1.zip "ZIP post-correctif T28"
[3]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/feature/t28-local-extensions-controlled/evidence/T28/t28-postfix-evidence.delta.bundle "Bundle post-correctif T28"
