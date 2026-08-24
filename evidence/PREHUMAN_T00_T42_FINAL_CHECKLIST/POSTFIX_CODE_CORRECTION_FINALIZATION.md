# ForgeLocal — finalisation des 13 findings GolangCI-Lint

## Identité

Le correctif est basé sur le clone neuf validé au commit `6ae02e4ceed239b9310fbf3fccb1b5170117251e`. Deux commits ont été créés sur la branche `fix/t00-t42-golangci-findings`, puis rattachés à la branche de livraison : `6ee0840a7b264343be3840998df2a8903b511722` et `e0c9352710eb3710eaf0ea5d71614f2731a7051c`.

Aucun finding n’a été neutralisé par exclusion, configuration ou `nolint`. Les corrections traitent les retours d’erreur ou le comportement de transfert; le test T38 a été transformé d’une branche vide en assertion de non-régression. Le mapping exhaustif et les messages originaux sont conservés dans `PREHUMAN_FINAL_QUALITY_NORMALIZED.log`, `PREHUMAN_FINAL_GOLANGCI_NEW_FINDINGS_AUDIT_RAW.log` et `POSTFIX_TARGETED_FINDINGS_VERIFICATION.log`.

## Mapping exhaustif

| Findings d’origine | Fichier et correction | Test / preuve | Décision |
|---|---|---|---|
| T1 `srv.Shutdown` non vérifié | `cmd/server/main.go` : retour de `srv.Shutdown` journalisé sur stderr | `go test ./cmd/server`; qualification globale post-correctif | `CLOSED_WITH_CODE_FIX` |
| T2 `cmd.Start` non vérifié dans `openBrowser` | `cmd/server/main.go` : `openBrowser` retourne l’erreur; appel startup traité; `cli_runtime.go` retourne une erreur CLI | `go test ./cmd/server`; qualification globale post-correctif | `CLOSED_WITH_CODE_FIX` |
| T3 `w.Write` non vérifié | `internal/api/sessions.go` : helper `writeAll` traite les écritures partielles et les erreurs | `TestWriteAllHandlesPartialAndFailedWrites` | `CLOSED_WITH_CODE_FIX` |
| T4 `clientConn.Write` 502 non vérifié | `internal/api/sessions.go` : réponse 502 via `writeAll` | tests API ciblés et race | `CLOSED_WITH_CODE_FIX` |
| T5 `backendConn.Write` upgrade non vérifié | `internal/api/sessions.go` : requête d’upgrade via `writeAll` | tests API ciblés et race | `CLOSED_WITH_CODE_FIX` |
| T6 `clientConn.Write` 502 non vérifié | `internal/api/sessions.go` : seconde réponse 502 via `writeAll` | tests API ciblés et race | `CLOSED_WITH_CODE_FIX` |
| T7 `clientConn.Write` 101 non vérifié | `internal/api/sessions.go` : réponse 101 via `writeAll` | tests API ciblés et race | `CLOSED_WITH_CODE_FIX` |
| T8 `io.Copy` client→backend non vérifié | `internal/api/sessions.go` : `copyAll` retourne l’erreur, deux goroutines attendues | `TestCopyAllPropagatesReaderAndWriterResult` | `CLOSED_WITH_CODE_FIX` |
| T9 `io.Copy` backend→client non vérifié | `internal/api/sessions.go` : erreurs des deux copies attendues avant sortie | `TestCopyAllPropagatesReaderAndWriterResult` | `CLOSED_WITH_CODE_FIX` |
| T10 rollback `BeginBackup` non vérifié | `internal/backup/store.go` : rollback non terminal joint au résultat nommé | `TestBackupTransactionLifecycleRemainsAtomic` | `CLOSED_WITH_CODE_FIX` |
| T11 rollback `MarkPublished` non vérifié | `internal/backup/store.go` : rollback non terminal joint au résultat nommé | `TestBackupTransactionLifecycleRemainsAtomic` | `CLOSED_WITH_CODE_FIX` |
| T12 rollback `CommitBackup` non vérifié | `internal/backup/store.go` : rollback non terminal joint au résultat nommé | `TestBackupTransactionLifecycleRemainsAtomic` | `CLOSED_WITH_CODE_FIX` |
| T13 SA9003 branche vide T38 | `internal/sessiontrack/tracker_test.go` : assertion explicite sur une clé valide | test T38 corrigé et exécuté | `CLOSED_WITH_TEST_FIX` |

Les corrections de `internal/api/sessions.go` ne changent pas le protocole attendu; elles empêchent seulement de perdre silencieusement une écriture partielle ou une erreur de copie. Les corrections de transaction ignorent uniquement `sql.ErrTxDone`, qui est le retour normal d’un rollback après commit, et joignent toute autre erreur au résultat de la méthode.

## Résultats post-correctif

Le clone neuf indépendant au HEAD `e0c9352710eb3710eaf0ea5d71614f2731a7051c` a exécuté `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `go mod verify`, `go list -m -json all`, `govulncheck`, OSV, Trivy, Syft CycloneDX/SPDX, Dashboard install/tsc/build/audit et les scans de secrets. Les contrôles ciblés et globaux ont un code `0`, à l’exception documentée des scanners qui retournent `1` parce qu’ils conservent des findings historiques.

GolangCI-Lint post-correctif retourne encore `82` findings historiques, mais **aucun** des 13 tuples ciblés n’est présent. Staticcheck conserve `36` findings historiques; le SA9003 T38 ciblé a disparu. Gosec conserve `194` findings baseline et `182` HEAD par comparaison sensible aux lignes; la comparaison normalisée par règle, fichier et détail donne `new_count=0` et `resolved_count=2`, correspondant aux deux signaux de `internal/api/sessions.go` effectivement éliminés. Le détail complet figure dans les sorties JSON et le normalisateur.

Gitleaks conserve le signal cumulatif historique `APi=REDACTED` avec `exit_code=1`; il reste `SCAN_BLOCKED_UNKNOWN`. Le contrôle Playwright a été invoqué mais reste `NOT_APPLICABLE_UNDER_CURRENT_GATES` avec `exit_code=1` dans le bloc brut : `FORGELOCAL_CORE_BASE_URL`, token protégé et `FORGELOCAL_BINARY` manquent. Le script global se termine à `0` parce qu’il journalise les contrôles attendus sans interrompre la collecte; la décision Playwright n’est donc pas un succès.

Le `git lfs fsck` du clone de qualification sans objets LFS historiques globalement réhydratés retourne `1` pour objets manquants; cela est attendu sous la discipline LFS ciblée. Le `git fsck --full` du clone neuf est à `0`, et les objets LFS historiques avaient déjà été vérifiés par récupération séquentielle dans la validation canonique. Aucun `git lfs pull` global n’a été exécuté.

## Verdict

La sortie est `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW_WITH_CODE_FIXES`. Elle autorise la revue humaine de la chaîne corrigée, mais ne lève aucune gate et ne constitue pas une release. Les statuts T28, T29, T39, T40, T41 et T42 restent `BLOCKED`; T30 reste `PENDING_REMOTE_EVIDENCE_RECONCILIATION`; T31–T38 restent `APPROVED_VERIFIABLE_LOCAL_WITH_POSTHOC_BASELINE_RECONSTRUCTION`.

Les contrôles suivants restent obligatoirement ouverts : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`.
