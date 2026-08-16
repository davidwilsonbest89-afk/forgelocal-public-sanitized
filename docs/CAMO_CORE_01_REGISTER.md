# Registre CAMO-CORE-01 — Audit passif et décisions de portage

| Champ | Valeur |
|---|---|
| Identifiant | `CAMO-CORE-01` |
| État | T07 clôturé au statut `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION` (décision `T07-DECISION-20260815-001`) ; aucun module Camoflox n’est intégré ni exécuté par ForgeLocal ; seul `lib/concurrency.js` est autorisé en réimplémentation Go sélective (T08). |
| Décision T07 | `T07-DECISION-20260815-001`, 2026-08-15 ; dossier de décision redacted SHA-256 `a10087c8b112949797ffd9f270693f8ddddc399234e445167c3126ad040041c3` ; attestation `ATT-T07R-CAMOFLOX-001` ; revues indépendantes `REV-T07R-2026-001` et `REV-T07R-2026-002`, décision finale `REVIEW_ACCEPTED_FOR_T07_DECISION`. |
| Sécurité | Alerte `generic-api-key` (tests/smoke.test.js:24) : triage concordant `FALSE_POSITIVE` (détenteur + relectrice) ; snapshot redacted `b8e71d3b2fdad98de8346c8063aa438caa0c1b0578a71528feab8fe5852326cf` rescanné à zéro détection. |
| Source auditée | `camoflox-FINAL.zip`, SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` ; archive fournie hors Git, commit source non fourni donc non vérifiable. |
| Décision globale | **Réimplémenter en Go après revue indépendante** ; ne pas copier ni lancer le code JavaScript. |
| Scan passif | Une occurrence Gitleaks est présente dans l’archive extraite ; sa valeur est volontairement non affichée. La source reste **non admissible au portage** jusqu’à classification indépendante. |
| Date d’audit passif | 2026-08-14 |
| Responsable de décision | Non assigné ; un mainteneur ForgeLocal doit être nommé avant tout portage. |

## Modules examinés

| Module source | Hash source | Dépendances observées | Décision | Justification et prochaine preuve |
|---|---|---|---|---|
| `lib/concurrency.js` | `b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7` | Promesses JavaScript | **`reimplementer` — AUTORISÉ T08** (2026-08-15) | Contrat de concurrence Go pur (queue bornée, limite globale, verrou par profil, timeout, contexte) ; exclusions explicites : code Node/Electron, lancement runtime, cycle de vie navigateur, ports, queue de lancement, isolation de processus, activation Camoufox ; rôle `concurrency-contract-inspiration-only` enregistré avant tout code. |
| `lib/global-action-limiter.js` | `a5f6838be5731ff64836f013b3b27fa280874de43afd817829de92052a487d05` | État mémoire JavaScript | Reporté | T08 interdit hors périmètre concurrence ; à réexaminer uniquement après clôture de T08, si autorisé. |
| `lib/profile-launch-queue.js` | `bf547b09299eafebfe541401aa9cb104e9679bf0ac79de016eec06400e3577cb` | Queue JavaScript | Hors périmètre | Interdit dans ce lot (lancement) ; futurs invariants re-spécifiés séparément. |
| `lib/process-isolation.js` | `10ac1cc010121afa1a895117b5d435e6231b90e9bc59253b284d4b499a73d062` | Fichiers lock, PID, ports | Écarter comme implémentation | Les locks fichiers et la vérification de PID ne suffisent pas au contrat ForgeLocal ; seuls les invariants seront re-spécifiés. |
| `lib/browser-lifecycle.js` | `9aefb2e999108ed719d19f7061032192ab097f0c73e4c2579bf8c3660b2cc4f4` | Cycle de vie navigateur | Hors périmètre | Interdit (cycle de vie navigateur) ; futur jalon runtime séparé. |

> Le registre est exécutoire pour `lib/concurrency.js` depuis la décision T07 du 2026-08-15 : hash source enregistré, décision `reimplementer` enregistrée avant tout code, exclusions explicites, triage de sécurité concordant et snapshot redacted rescanné. Pour tout autre module, aucune réimplémentation ne peut commencer sans une nouvelle décision et des tests propres.

## Invariants non négociables pour la réimplémentation autorisée (T08)

Le Core Go doit posséder le cycle de vie complet. Toute réservation de port doit être définie par runtime, limitée par timeout, confirmée par l’endpoint effectivement attaché au processus, puis libérée par un cleanup idempotent. Une reprise après crash ne doit jamais relancer silencieusement une instance inconnue.

Le dashboard ne possède aucun lock, aucune file de lancement et aucune décision de runtime. Camoufox reste un candidat non lançable jusqu’à qualification séparée et ne peut pas recevoir de preuve d’un autre runtime.
