# Audit de gap T18-R — queue / recovery sur la lignée post-T17-R

**Baseline du lot :** commit `55776a4a802e5c9c36b4a49a98787a40bd215619` (tag `t17r-g15-real-handler-verified-2026-08-18`), branche `forgelocal-baseline-2026-08-17` du dépôt `boucheriechefimane-cmd/IPcache`. Le T17 historique reste `T17_SOURCE_SNAPSHOT_UNRECOVERABLE` et le dossier T18 historique n'est ni restauré ni modifié.

## Composants présents

Le composant queue/recovery est `internal/launch` (T18) avec 4 fichiers de production : `manager.go` (file bornée, limite globale, sérialisation par profil, budget de noms, `Recover`), `launch.go` (chaîne `Request` : vérification `ctx`, profil conflictuel, journal `Lookup`/`Reserve` idempotent, `begin`, attente `ctx.Done`/deadline, arrêt propre), `journal.go` et `journal_sqlite.go` (journal durable SQLite, transitions atomiques, colonne `correlation_id`, redaction `redactCorrelationID`/`redacted`) et `redact.go`/`id.go`.

## Tableau par comportement exigé

| Comportement exigé | Contrôles présents dans la lignée | Tests | Verdict |
|---|---|---|---|
| Queue bornée (MaxQueueDepth, ErrQueueFull, deadline de wait) | `launch.go:63` refuse `ErrQueueFull` quand limite globale et profondeur pleines | `TestRequest_QueueFull`, `TestT18Journal_QueueSaturationLeavesNoPartialOperation` | PRESENT_AND_TESTED |
| Limite globale de sessions | `launch.go:62` `len(m.running) >= m.opt.GlobalLimit` | `TestRequest_GlobalLimit` (montée puis descente sous la limite) | PRESENT_AND_TESTED |
| Sérialisation par profil / contention | `launch.go:59` `ErrProfileAlreadyRunning`, verrou par profil libéré dans les chemins d'erreur (`launch.go:137`) | `TestRequest_SingleSessionPerProfile`, `TestT18Journal_ActiveProfileContentionLeavesNoPartialOperation`, `TestT18Journal_DistinctProfilesReserveIndependently` | PRESENT_AND_TESTED |
| Annulation avant exécution (pendant la file) | `launch.go:27-33` refus précoce, libération sans attacher | `TestRequest_CancelledWhileQueued`, `TestT18Journal_CancelBeforeExecutionLeavesNoPartialWrite` | PRESENT_AND_TESTED |
| Annulation pendant exécution | `launch.go:98-105` `ctx.Done` pendant attach → `ErrCancelled`, profil libéré | `TestT18Journal_CancelDuringAttachPersistsInterrupted` (état durable `interrupted` vérifié en DB) | PRESENT_AND_TESTED |
| Timeout par opération | `launch.go:118` `WithDeadline`, finalisation en erreur | `TestT18Journal_OperationTimeoutPersistsError` (état durable `error`) | PRESENT_AND_TESTED |
| Arrêt propre / idempotence | `Stop` sans fuite, `ReleaseAllProfiles` | `TestStop_Idempotent`, `TestStop_ReleaseAllProfiles`, `TestRequest_ReuseAfterStop` | PRESENT_AND_TESTED |
| Reprise après crash (Recover) | `Options.Recoverer`, `Recover([]RecoveredSession)`, ghost sessions réconciliées et profils libérés | `TestRecover_NoGhostSessions` | PRESENT_AND_TESTED |
| Idempotence / refus de doublon | `journal.Lookup` avant `Reserve` (clé `idempotencyKey`) | `TestT18Journal_IdempotencePreventsDoubleAttachAfterCrash` (0 attach après reprise), `TestT18Journal_ManagerRestartRecoveryIsIdempotentAndDoesNotAttach` | PRESENT_AND_TESTED |
| Reprise idempotente du manager redémarré | `Recover` sans auto-attach | `TestT18Journal_ManagerRestartRecoveryIsIdempotentAndDoesNotAttach`, `TestT18Journal_CrashRecoveryAfterStartingStopsWithoutAttach` | PRESENT_AND_TESTED |
| Refus d'état incohérent (fail-closed) | `Reconcile` fail-closed, aucune écriture partielle | `TestT18Journal_ReconcileRejectsIncoherentStateWithoutPartialWrite`, `TestT18Journal_InvalidTransitionHasNoPartialAuditWrite` | PRESENT_AND_TESTED |
| Absence d'écriture partielle SQLite | Transitions et réserve atomiques, audit écrit après transition validée | `TestT18Journal_InvalidTransitionHasNoPartialAuditWrite`, `TestT18Journal_CrashRecoveryAfterStartingStopsWithoutAttach` | PRESENT_AND_TESTED |
| Audit redacted avec correlation_id | `correlation_id TEXT NOT NULL`, `redactCorrelationID` (longueur/charset → `[redacted]`), `redacted(Reason)` | `TestT18Journal_AuditRedactsSecretAndInvalidCorrelation`, `TestT18Journal_ReconcileKeepsCorrelationInCrashAudit`, `TestAudit_Redacted` | PRESENT_AND_TESTED |
| Nettoyage après échec d'attach | `begin` → cleanup en cas d'échec | `TestRequest_AttachFailure_Cleanup` | PRESENT_AND_TESTED |
| Concurrence réelle | 40 goroutines par profil + 120 goroutines sur 24 profils | `TestConcurrentStress` | PRESENT_AND_TESTED |
| Cohérence Status/limites | `Status()` borné, profils exclusifs par session | `TestStatus_Bounds`, `TestRequest_InvalidProfile` | PRESENT_AND_TESTED |

## Verdict

Aucun comportement T18 n'est ABSENT et aucun n'est PRESENT_UNTESTED dans la lignée post-T17-R. Tous les dix-sept critères exigés par la consigne T18-R sont couverts par du code de production et des tests existants qui s'exécutent sous `-race`. Aucune modification de code produit ni aucun ajout de test n'est nécessaire ; le lot consiste en la requalification complète de la lignée avec les preuves conservées.
