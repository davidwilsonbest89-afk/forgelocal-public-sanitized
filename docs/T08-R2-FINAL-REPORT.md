# Rapport final T08-R2 — LaunchManager / SessionManager (reimplémentation clean room)

Référence : `docs/CAHIER_DES_CHARGES_v3_9_7.md`, jalon T08 (concurrency + launch management).
Périmètre strict : uniquement la surface `LaunchManager/SessionManager` en Go. Aucun navigateur,
aucun runtime, aucun Camoufox, aucun proxy, aucune UI, aucun backup/restore/import, aucune release.

## Rapport au format exigé

**TASK:** T08 — reimplémentation clean room du contrat de concurrence et de gestion de lancement
(inspiration sémantique Camoufox `lib/concurrency.js` seulement : queue bornée, limite globale,
sélection par profil, timeouts, contexte d'annulation, cleanup idempotent, crash recovery, audit redacted).
Aucune ligne de code Camoufox importée, copiée ou exécutée ; les tests utilisent un `Launcher` stub
local (`blockingLauncher` qui dort, `failingLauncher` qui échoue) — aucun navigateur réel lancé.

**GAP:** Deux bugs critiques découverts pendant les stress tests et corrigés dans ce jalon :
1. `ErrQueueFull` dans la boucle d'attente queue retenait `m.mu` (deadlock massif sous charge
   queue-full) ;
2. la map `running` stockait `&sess` (paramètre de `begin`), pointeur partagé avec
   `begin.func1` et `cancelSession` → **data race réelle** confirmée par le race detector
   (lecture du paramètre par la goroutine attach pendant écriture de `State/StoppedAt/Err`
   par `cancelSession`), corrigée par un pointeur heap dédié par session.
Le premier lot de correctifs avait également introduit un deadlock `ErrQueueFull` (mutex non
relâché sur erreur) et des deadlocks de harnais de test (launchers bloquants sans timeout) — tous
fermés avant ce rapport.

**IMPLEMENTED:**
- `internal/launch` : `Manager` (options `GlobalLimit`, `MaxQueueDepth`, `StartTimeout`,
  `RecoveryRecords`), `Request` (queue bornée, refus immédiat `ErrQueueFull` sans retenir le lock,
  sélection du profil, promotion asynchrone par une goroutine dédiée pour réduire la contention),
- sérialisation par profil : une session active par profil, les suivantes attendent dans la queue
  et refusées en attente si le profil est occupé (`session already active for profile`),
- limite globale de concurrence stricte, queue FIFO bornée, `context.Context` propagé au
  `Launcher.Attach`, timeout de démarrage appliqué uniquement si le contexte n'a pas de deadline,
- `Stop` idempotent : annule tous les attach en vol (via contexte manager + contexte de chaque
  attach), join prouvé borné (test `Stop` retourne sous deadline bornée), libère tous les profils,
- crash recovery : `Recover` reconstruit l'état depuis des enregistrements persistés, sans session
  fantôme (test `TestRecover_NoGhostSessions`),
- audit : sink redacted — toute erreur `Attach` est réécrite en `attach error: [redacted]`
  (seuls les motifs bénins de cycle de vie passent en clair), test `TestAudit_Redacted`,
- état Git propre : `git diff --check` propre, aucun secret dans le delta (Gitleaks 8.18.4,
  scan snapshot 5 fichiers, JSON `[]`), registre de droits mis à jour avec le module `launch`
  (MIT propre, clean room, pas de copiage).

**NOT IMPLEMENTED:** aucun lancement de navigateur réel, aucun pilote Camoufox activé, aucun
port, aucun proxy, aucune écriture SQLite depuis ce package, aucune UI, aucun backup/restore,
aucune release — statut `PUBLIC_RELEASE_BLOCKED` inchangé, `SCAN_BLOCKED_UNKNOWN` inchangé,
pilote suspendu, cinq gates publics actifs. Le lancement réel du runtime appartiendra à un
jalon ultérieur autorisé.

**FILES CHANGED:** 9 fichiers (7 en création dans le commit T08) :
`internal/launch/id.go`, `internal/launch/launch.go`, `internal/launch/launch_test.go`,
`internal/launch/manager.go`, `internal/launch/redact.go`, `docs/T08_CONCURRENCY_SPEC.md`,
`docs/CAHIER_DES_CHARGES_v3_9_7.md` (consolidation CDC déjà livrée), `docs/component-rights-register.json`,
`docs/CAHIER_DES_CHARGES_FORGELOCAL.md` (référencement T08).

**API CHANGED:** aucune API publique HTTP/REST modifiée ; le contrat API Core lecture seule est
inchangé. L'API interne nouvelle est `internal/launch` (package privé) :
`NewManager(opts)`, `m.Request(ctx, launcher, profileID, notify)`, `m.Stop(ctx)`,
`m.Recover(records)`, `m.Status()`, `m.Audit()`, `NewRedactor()`. Aucun impact sur
`docs/CAHIER_DES_CHARGES` API Core.

**DATABASE CHANGED:** aucune migration, aucun schéma SQLite touché. `RecoveryRecords` est un type
Go pur ; la persistance réelle (table `launch_sessions`) appartiendra au jalon ultérieur autorisé.

**UI CHANGED:** aucune. Dashboard inchangé (checkpoint `2b50697e`).

**TESTS WRITTEN:** 13 tests dans `internal/launch/launch_test.go` :
`TestRequest_SingleSessionPerProfile`, `TestRequest_InvalidProfile`, `TestRequest_GlobalLimit`,
`TestRequest_QueueFull`, `TestRequest_CancelledWhileQueued`, `TestRequest_AttachFailure_Cleanup`,
`TestRecover_NoGhostSessions`, `TestStop_Idempotent`, `TestStop_ReleaseAllProfiles`,
`TestAudit_Redacted`, `TestRequest_ReuseAfterStop`, `TestStatus_Bounds`, `TestConcurrentStress`
(120 goroutines, 24 profils, annulations échelonnées, Stop borné, résolution de tous les appelants,
audit redacted).

**TESTS EXECUTED:**
```text
GOTOOLCHAIN=local go test -count=1 -race -timeout 180s -v ./internal/launch
# exit 0, 13 sélectionnés, 13 PASS, 0 FAIL, 0 data race, 3.457s
go vet ./internal/launch   # exit 0
go build ./internal/launch # exit 0
```

**TEST RESULTS:** 13/13 PASS sous `-race` (liste exacte : voir archive `t08-r2.zip` /
`test-out.log`, 13 lignes `--- PASS`), couverture : chemin positif, chemin négatif
(profil invalide, queue pleine, annulation en queue, échec attach, timeout, stop), teardown
(réutilisation après stop, idempotence), concurrence (stress 120 goroutines / 24 profils /
annulations échelonnées), propriété d'audit (redaction de chaque motif).

**RACE RESULT:** `go test -race` : **zéro data race** (sortie propre, exit 0, suite complète).
La race précédemment identifiée entre `cancelSession` et `begin.func1` est fermée par le
pointeur heap dédié.

**SECURITY SCAN:** Gitleaks 8.18.4 `detect --no-git --source <snapshot 5 fichiers Go>` :
JSON `[]`, zéro détection, exit 0. Le delta git T08 (9 fichiers) ne contient aucun chemin RC
et aucune valeur de secret ; la preuve `gitleaks-t08-snap.json` est dans l'archive.
Pas d'introduction d'identifiants proxy, cookie, chemin absolu, valeur de coffre ni détail runtime.

**EVIDENCE:** archive `t08-r2.zip` — SHA-256 :
`9c904b1ef520d0828b8d9591ad0e6f795b9176d4fa22e57dffbf2e8e01fca335` ;
contenu vérifié : `unzip -t` sans erreur, `sha256sum -c SHA256SUMS` 10/10 OK après extraction ;
inclus : 5 sources Go, `T08_CONCURRENCY_SPEC.md`, `test-out.log` (-race -v complet),
`gitleaks-t08-snap.json` (`[]`), `source-info.txt` (commit `99a22f5106ebf0cef27e46c551757de8355e5cad`,
baseline `31a51e948e5975af4726db78c26b5da4a72e47c3`, branche `forgelocal-product-v0.3`,
dépôt `github.com/boucheriechefimane-cmd/IPcache`, Go 1.25.13).
Commit poussé et publié : `99a22f5`. Registre de droits mis à jour dans le même commit.
Chaîne T07-R1→T08 : le module `launch` est déclaré clean room (réimplémentation propre,
`integration_state` du candidat Camoufox inchangé, toujours `provenance-qualification-blocked`
dans le registre des composants externes).

**LIMITATIONS:**
- `Launcher.Attach` reste un stub de test (failing/blocking) : aucun vrai runtime n'est branché ;
- la persistance des `RecoveryRecords` (SQLite) n'existe pas encore dans ce package ;
- aucune limite de ports/ISOLATION réseau implémentée (hors périmètre T08) ;
- aucune garantie d'empreinte navigateur / fingerprint (hors périmètre) ;
- le stress test est borné à 120 goroutines / 24 profils ; des environnements très chargés
  devront être re-qualifiés avant intégration réelle ;
- le jalon ne débloque pas T09 : l'autorisation T09 reste soumise à votre validation unique
  finale de T08.

**CURRENT STATUS:** `🟢 COMPLET` pour le périmètre T08 (toutes les preuves T08 exigées sont
présentes : tests + négatifs + concurrence, `-race` propre, `go vet` propre, Gitleaks propre,
`git diff --check` propre, archive + manifeste + hashes, commit publié).
Statuts inchangés : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, pilote suspendu,
cinq gates publics actifs, aucun composant Camoufox intégré ni exécuté.

**NEXT ALLOWED STEP:** attente de votre validation unique finale de T08. Après validation,
le prochain jalon autorisé est **T09** (portage/runtime réel) si et seulement si T07 est
débloqué par les preuves externes indépendantes ; sinon, l'attente contrôlée T07-R continue.
Aucune mutation UI, aucun lancement de runtime, aucun proxy, aucun backup/restauration et
aucun changement RC ne sont autorisés avant ces validations.
