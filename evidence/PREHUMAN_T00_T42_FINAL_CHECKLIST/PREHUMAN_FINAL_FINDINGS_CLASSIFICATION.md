# ForgeLocal — finalisation des findings T00–T42

**Exécution :** 2026-08-24, clone neuf `/home/ubuntu/forgelocal-prehuman-fresh-20260824`, HEAD `6ae02e4ceed239b9310fbf3fccb1b5170117251e`.
**Baseline :** `t00-t27-complete-20260820` → `69411e65c880d168832a65fc8475cc97d562a9ad`.
**Principe :** cette finalisation ne modifie aucun fichier de code produit ou test métier. Chaque finding est soit corrigé dans une passe autorisée, soit couvert par une exception explicite avec propriétaire, risque et condition de levée.

## Décision de finalisation

Les 13 findings GolangCI-Lint nouveaux du différentiel sont désormais tous individualisés. Aucun n’est laissé sous la seule étiquette « connu » : les 13 portent une exception ouverte et traçable, car leur correction nécessiterait soit une modification de code Core/API/BACK-01, soit une modification du test T38, ce qui élargirait la passe gelée. Le résultat est donc `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW`, avec risques ouverts explicitement documentés.

| Contrôle | Baseline | HEAD | Nouveau/résolu | Décision |
|---|---:|---:|---:|---|
| GolangCI-Lint | 82 | 83 | 13 nouveaux / 12 résolus | 13 exceptions exhaustives, aucune correction de code dans cette passe |
| Staticcheck | 36 | 36 | 0 / 0 | 36 exceptions historiques exhaustives |
| Gosec | 194 | 194 | 0 / 0 | Findings historiques conservés, normalisation jointe |
| Trivy | 6 | 6 | 0 / 0 | 6 misconfigurations historiques exhaustives |
| Gitleaks | signal historique | signal historique | `APi=REDACTED`, scanner 1 | `SCAN_BLOCKED_UNKNOWN` maintenu |

## 13 findings GolangCI-Lint nouveaux

La sévérité n’est pas renseignée dans le JSON GolangCI-Lint; aucune sévérité numérique n’est donc inventée. Les messages bruts, chemins et lignes proviennent de la sortie JSON. Les causes listées distinguent les lignes historiquement présentes mais absentes de la sortie baseline des lignes effectivement introduites par T38.

| Règle | Fichier | Ligne | Message brut | Sévérité | Baseline/head | Lot et cause | Risque | Décision | Propriétaire ; condition de levée |
|---|---|---:|---|---|---|---|---|---|---|
| staticcheck | internal/sessiontrack/tracker_test.go | 35 | SA9003: empty branch | non renseignée dans JSON | absent baseline / présent HEAD | T38 ; La ligne appartient à 57fc811… introduit dans T38 après la baseline; c’est le seul finding rattachable directement à un lot HEAD. | Branche de test vide pouvant réduire la lisibilité ou masquer une intention. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Propriétaire T38 ; Décider explicitement la forme de la branche de test et rejouer T38; aucune modification dans cette passe. |
| errcheck | internal/backup/store.go | 109 | Error return value of `tx.Rollback` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | BACK-01 / historique antérieur à la baseline ; La ligne appartient à 97465bf… du socle backup antérieur à la baseline; fichier inchangé dans baseline..HEAD. | Rollback différé non contrôlé. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur BACK-01 ; Autorisation BACK-01 et décision sur les erreurs de rollback différé. |
| errcheck | internal/backup/store.go | 125 | Error return value of `tx.Rollback` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | BACK-01 / historique antérieur à la baseline ; La ligne appartient à 97465bf… du socle backup antérieur à la baseline; fichier inchangé dans baseline..HEAD. | Rollback différé non contrôlé. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur BACK-01 ; Autorisation BACK-01 et décision sur les erreurs de rollback différé. |
| errcheck | internal/backup/store.go | 144 | Error return value of `tx.Rollback` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | BACK-01 / historique antérieur à la baseline ; La ligne appartient à 97465bf… du socle backup antérieur à la baseline; fichier inchangé dans baseline..HEAD. | Rollback différé non contrôlé. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur BACK-01 ; Autorisation BACK-01 et décision sur les erreurs de rollback différé. |
| errcheck | internal/api/sessions.go | 238 | Error return value of `w.Write` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / API legacy, lot précis non établi ; La ligne appartient à cc554320… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Écriture HTTP non contrôlée. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction API et test métier associé. |
| errcheck | internal/api/sessions.go | 397 | Error return value of `clientConn.Write` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / WebSocket legacy, lot précis non établi ; La ligne appartient à 649f803… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Réponse 502 potentiellement non écrite. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction du proxy WebSocket et test dédié. |
| errcheck | internal/api/sessions.go | 411 | Error return value of `backendConn.Write` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / WebSocket legacy, lot précis non établi ; La ligne appartient à 649f803… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Requête de montée de tunnel potentiellement non écrite. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction du proxy WebSocket et test dédié. |
| errcheck | internal/api/sessions.go | 417 | Error return value of `clientConn.Write` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / WebSocket legacy, lot précis non établi ; La ligne appartient à 649f803… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Réponse HTTP d’erreur potentiellement non écrite. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction du proxy WebSocket et test dédié. |
| errcheck | internal/api/sessions.go | 422 | Error return value of `clientConn.Write` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / WebSocket legacy, lot précis non établi ; La ligne appartient à 58a5e79… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Handshake 101 potentiellement non écrit. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction du proxy WebSocket et test dédié. |
| errcheck | internal/api/sessions.go | 432 | Error return value of `io.Copy` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / WebSocket legacy, lot précis non établi ; La ligne appartient à 58a5e79… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Copie client→backend potentiellement interrompue sans signal. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction du proxy WebSocket et test dédié. |
| errcheck | internal/api/sessions.go | 433 | Error return value of `io.Copy` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / WebSocket legacy, lot précis non établi ; La ligne appartient à 58a5e79… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absent de la sortie baseline. | Copie backend→client potentiellement interrompue sans signal. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur API ; Autorisation de correction du proxy WebSocket et test dédié. |
| errcheck | cmd/server/main.go | 469 | Error return value of `srv.Shutdown` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / Core legacy, lot précis non établi ; La ligne appartient à cc554320… antérieur à la baseline; fichier inchangé dans baseline..HEAD; absence du finding dans la sortie baseline, cause scanner/contexte non résolue. | Perte possible de l’erreur d’arrêt du serveur. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur Core ; Rerun linter sous contexte strictement identique ou autorisation de correction Core. |
| errcheck | cmd/server/main.go | 567 | Error return value of `cmd.Start` is not checked | non renseignée dans JSON | absent baseline / présent HEAD | T00–T30 / Core legacy, lot précis non établi ; La ligne appartient à cc554320… antérieur à la baseline; fichier inchangé dans baseline..HEAD; différentiel non attribuable à une modification HEAD. | Échec de démarrage potentiellement non propagé. | EXCEPTION_OPEN_DOCUMENTED_NO_CODE_CHANGE | Mainteneur Core ; Rerun linter identique ou autorisation de correction Core. |

## 36 findings Staticcheck historiques

Les 36 entrées ci-dessous sont présentes dans la baseline et dans HEAD, avec zéro nouveau et zéro résolu. Le scanner retourne `exit_code=1`; elles restent des exceptions qualité historiques, non un PASS.

| Règle | Fichier | Ligne | Message brut | Sévérité | Baseline/head | Risque | Décision | Condition de levée |
|---|---|---:|---|---|---|---|---|---|---|
| S1016 | cmd/server/cli.go | 250 | should convert global (type cliGlobal) to mcpStdioOptions instead of using struct literal | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA4006 | cmd/server/cli_runtime.go | 237 | this value of cfg is never used | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/backup_v1.go | 35 | func (*handler).profileHasLiveSession is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/proxies_test.go | 30 | type loopbackNet is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/proxies_test.go | 32 | func loopbackNet.Dial is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/proxies_test.go | 33 | func loopbackNet.DialTimeout is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/proxies_test.go | 34 | func loopbackNet.Listen is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/proxies_test.go | 37 | func loopbackNet.ListenPacket is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/proxies_test.go | 38 | func loopbackNet.ResolveTCPAddr is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1012 | internal/api/proxies_test.go | 325 | do not pass a nil Context, even if a function permits it; pass context.TODO if you are unsure about which Context to use | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/api/sessions.go | 289 | page.WaitForSelector is deprecated: Use web assertions that assert visibility or a locator-based [Locator.WaitFor] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/api/templates_test.go | 32 | func templateRequest is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| ST1005 | internal/browser/launch_helpers.go | 29 | error strings should not be capitalized | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| ST1005 | internal/browser/launch_helpers.go | 33 | error strings should not be capitalized | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/history/store.go | 24 | const defaultPageSize is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/humanize/keyboard.go | 27 | page.Fill is deprecated: Use locator-based [Locator.Fill] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/humanize/keyboard.go | 33 | page.Click is deprecated: Use locator-based [Locator.Click] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/humanize/mouse.go | 21 | page.Click is deprecated: Use locator-based [Locator.Click] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/humanize/mouse.go | 56 | page.Click is deprecated: Use locator-based [Locator.Click] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/launch/launch.go | 497 | func (*Manager).await is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/launch/launch_test.go | 37 | field attached is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/advanced_tools.go | 155 | target.page.WaitForSelector is deprecated: Use web assertions that assert visibility or a locator-based [Locator.WaitFor] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/advanced_tools.go | 326 | target.page.SelectOption is deprecated: Use locator-based [Locator.SelectOption] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/advanced_tools.go | 356 | target.page.Check is deprecated: Use locator-based [Locator.Check] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/advanced_tools.go | 360 | target.page.Uncheck is deprecated: Use locator-based [Locator.Uncheck] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/mcp/results.go | 9 | func imageResult is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/mcp/screenshot_artifacts.go | 17 | const defaultScreenshotURLTTL is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/mcp/search_provider.go | 48 | func supportedSearchProviderNames is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/search_provider.go | 66 | page.WaitForSelector is deprecated: Use web assertions that assert visibility or a locator-based [Locator.WaitFor] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/search_provider.go | 163 | page.WaitForSelector is deprecated: Use web assertions that assert visibility or a locator-based [Locator.WaitFor] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/mcp/search_provider.go | 248 | page.WaitForSelector is deprecated: Use web assertions that assert visibility or a locator-based [Locator.WaitFor] instead. Read more about [locators]. | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/mcp/server.go | 36 | field reqID is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA1019 | internal/profile/backup_snapshot.go | 229 | tar.TypeRegA has been deprecated since Go 1.11 and an alternative has been available since Go 1.1: Use TypeReg instead.  | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/profile/errors.go | 106 | func wrap is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| U1000 | internal/proxies/errors.go | 82 | func wrap is unused | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |
| SA4006 | internal/proxies/store.go | 213 | this value of p is never used | error | présent / présent | Qualité statique; impact précis à revoir au remédiateur | EXCEPTION_HISTORICAL_OPEN | Autorisation de remédiation et rerun baseline/head sans masquer le finding |

## 6 misconfigurations Trivy historiques

Les six entrées sont présentes à l’identique en baseline et HEAD. Le différentiel est zéro nouveau et zéro résolu. Les rapports ne signalent aucune vulnérabilité Go/pnpm ni secret de dépendance dans cette exécution.

| Cible | Règle | Sévérité | Titre brut | Baseline/head | Risque | Décision | Condition de levée |
|---|---|---|---|---|---|---|---|
| Dockerfile | DS-0002 | HIGH | Image user should not be 'root' | présent / présent | Durcissement d’image à évaluer | EXCEPTION_HISTORICAL_OPEN | Revue et correction Dockerfile autorisées, puis rerun Trivy |
| Dockerfile | DS-0026 | LOW | No HEALTHCHECK defined | présent / présent | Durcissement d’image à évaluer | EXCEPTION_HISTORICAL_OPEN | Revue et correction Dockerfile autorisées, puis rerun Trivy |
| Dockerfile | DS-0029 | HIGH | 'apt-get' missing '--no-install-recommends' | présent / présent | Durcissement d’image à évaluer | EXCEPTION_HISTORICAL_OPEN | Revue et correction Dockerfile autorisées, puis rerun Trivy |
| docker/Dockerfile.run | DS-0002 | HIGH | Image user should not be 'root' | présent / présent | Durcissement d’image à évaluer | EXCEPTION_HISTORICAL_OPEN | Revue et correction Dockerfile autorisées, puis rerun Trivy |
| docker/Dockerfile.run | DS-0026 | LOW | No HEALTHCHECK defined | présent / présent | Durcissement d’image à évaluer | EXCEPTION_HISTORICAL_OPEN | Revue et correction Dockerfile autorisées, puis rerun Trivy |
| docker/Dockerfile.run | DS-0029 | HIGH | 'apt-get' missing '--no-install-recommends' | présent / présent | Durcissement d’image à évaluer | EXCEPTION_HISTORICAL_OPEN | Revue et correction Dockerfile autorisées, puis rerun Trivy |

## Inventaire de licences de production

L’inventaire JSON de production contient **52 entrées** et a été généré avec `exit_code=0`. La liste exhaustive package par package est jointe dans `PREHUMAN_FINAL_LICENSE_INVENTORY.json`; le tableau ci-dessous donne le regroupement exact retourné par l’outil. La compatibilité juridique de chaque licence reste une décision de revue humaine, non une déduction automatique de l’inventaire.

| Valeur de licence retournée | Nombre d’entrées | Décision | Condition de levée |
|---|---:|---|---|
| MIT | 48 | INVENTORIED_NOT_LEGAL_APPROVAL | Validation humaine de compatibilité avec la politique de distribution |
| Apache-2.0 | 2 | INVENTORIED_NOT_LEGAL_APPROVAL | Validation humaine de compatibilité avec la politique de distribution |
| ISC | 1 | INVENTORIED_NOT_LEGAL_APPROVAL | Validation humaine de compatibilité avec la politique de distribution |
| Unlicense | 1 | INVENTORIED_NOT_LEGAL_APPROVAL | Validation humaine de compatibilité avec la politique de distribution |

## Gosec, Gitleaks et Playwright

Gosec conserve 194 findings en baseline et 194 en HEAD, avec `new_findings=0` et `resolved_findings=0`; le scanner retourne `exit_code=1`, ce qui reste documenté comme findings historiques ouverts. Gitleaks cumulatif conserve le marqueur historique `APi=REDACTED`, `exit_code=1`, et la gate `SCAN_BLOCKED_UNKNOWN` demeure obligatoire. Le scan Gitleaks no-git de l’addendum est vide et ne neutralise pas le signal cumulatif.

Playwright/T10 est `NOT_APPLICABLE_UNDER_CURRENT_GATES` et `BLOCKED_BY_REQUIRED_PROTECTED_CONFIGURATION`: la localisation et le message `CONFIGURATION_T10_ABSENTE`, l’absence de `FORGELOCAL_CORE_BASE_URL` et de token/config, la commande, le CWD, l’UTC et le code de sortie sont conservés dans `PREHUMAN_FINAL_EXIT_CHECKLIST_RAW.log` et dans la preuve dédiée `PREHUMAN_FINAL_PLAYWRIGHT_NOT_APPLICABLE.md` [7]. Il ne s’agit pas d’un oubli : aucun Core, token, proxy réel, cookie réel ou navigateur réel n’a été lancé.

## Gates et statut final

Les gates restent `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Les statuts T28, T29, T39, T40, T41 et T42 restent `BLOCKED`; T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`; T31–T38 restent `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION`.

> **Sortie attendue :** `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW`.

Cette sortie prépare la revue humaine et ne constitue ni une approbation complète, ni une release, ni une levée de gate.

## Références

[1]: ./PREHUMAN_FINAL_GOLANGCI_NEW_FINDINGS_AUDIT_RAW.log "Audit Git blame des 13 findings"
[2]: ./PREHUMAN_FINAL_QUALITY_NORMALIZED.log "Différentiel qualité normalisé"
[3]: ./PREHUMAN_FINAL_STATICCHECK_BASELINE.jsonl "Staticcheck baseline" ; ./PREHUMAN_FINAL_STATICCHECK_HEAD.jsonl "Staticcheck HEAD"
[4]: ./PREHUMAN_FINAL_TRIVY_BASELINE.json "Trivy baseline" ; ./PREHUMAN_FINAL_TRIVY_HEAD.json "Trivy HEAD"
[5]: ./PREHUMAN_FINAL_LICENSE_INVENTORY.json "Inventaire de licences production"
[6]: ./PREHUMAN_FINAL_EXIT_CHECKLIST_RAW.log "Journal brut final et preuve Playwright"
[7]: ./PREHUMAN_FINAL_PLAYWRIGHT_NOT_APPLICABLE.md "Preuve Playwright NOT_APPLICABLE dédiée"

## Amendement post-correctif — 2026-08-24

À la suite de la consigne senior, les 13 findings GolangCI-Lint précédemment classés comme exceptions ont été réanalysés et corrigés lorsqu’ils correspondaient à un retour d’erreur ou à une assertion de test manquante. Les commits de correction sont `6ee0840a7b264343be3840998df2a8903b511722` et `e0c9352710eb3710eaf0ea5d71614f2731a7051c`, rattachés à la branche de livraison par les commits cherry-pick `db0dd08` et `57d21b0`.

Le rapport `POSTFIX_CODE_CORRECTION_FINALIZATION.md` fournit le mapping exhaustif des 13 entrées vers les fichiers et les tests. `POSTFIX_TARGETED_FINDINGS_VERIFICATION.log` confirme `original_targeted_count=13`, `postfix_targeted_remaining_count=0` et `decision=ALL_13_TARGETED_FINDINGS_CLOSED`. Les tests ciblés et race sont à `0`; la qualification globale post-correctif est également à `0` pour `go test -count=1 -race ./...`, `go vet ./...` et `go build ./...`.

GolangCI-Lint retourne encore 82 findings non ciblés historiques; le finding SA9003 T38 ciblé a disparu. Staticcheck conserve 36 findings historiques, sans SA9003 ciblé. Gosec a `new_count=0` après comparaison normalisée par règle, fichier et détail, les variations de comptage ligne-sensible étant dues aux corrections et décalages de ligne. Les findings historiques ne sont ni masqués ni exclus.

Cette correction ne lève aucune gate. Gitleaks conserve `APi=REDACTED` et `SCAN_BLOCKED_UNKNOWN`; Playwright/T10 reste bloqué par `FORGELOCAL_CORE_BASE_URL`, token protégé et `FORGELOCAL_BINARY` absents; aucun runtime réel ou secret n’a été créé. Le clone de qualification a `git fsck --full=0`; son `git lfs fsck=1` est classé comme manque d’objets LFS historiques non réhydratés globalement, conformément à la discipline de récupération ciblée.
