# Registre CAMO-CORE-01 — Audit passif et décisions de portage

| Champ | Valeur |
|---|---|
| Identifiant | `CAMO-CORE-01` |
| État | Prévalidation documentaire ; aucun module Camoflox n’est intégré ni exécuté par ForgeLocal. |
| Source auditée | `camoflox-FINAL.zip`, SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` ; archive fournie hors Git, commit source non fourni donc non vérifiable. |
| Décision globale | **Réimplémenter en Go après revue indépendante** ; ne pas copier ni lancer le code JavaScript. |
| Scan passif | Une occurrence Gitleaks est présente dans l’archive extraite ; sa valeur est volontairement non affichée. La source reste **non admissible au portage** jusqu’à classification indépendante. |
| Date d’audit passif | 2026-08-14 |
| Responsable de décision | Non assigné ; un mainteneur ForgeLocal doit être nommé avant tout portage. |

## Modules examinés

| Module source | Hash source | Dépendances observées | Décision | Justification et prochaine preuve |
|---|---|---|---|---|
| `lib/concurrency.js` | `b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7` | Promesses JavaScript | Réimplémenter | Le pattern de borne de concurrence est utile, mais l’implémentation Go doit utiliser `context.Context`, tests race et annulation. |
| `lib/global-action-limiter.js` | `a5f6838be5731ff64836f013b3b27fa280874de43afd817829de92052a487d05` | État mémoire JavaScript | Réimplémenter | À remplacer par un limiteur Core Go, borné, observable et testé. |
| `lib/profile-launch-queue.js` | `bf547b09299eafebfe541401aa9cb104e9679bf0ac79de016eec06400e3577cb` | Queue JavaScript | Réimplémenter | À transformer en queue durable/fail-closed ; aucun lancement dans ce lot. |
| `lib/process-isolation.js` | `10ac1cc010121afa1a895117b5d435e6231b90e9bc59253b284d4b499a73d062` | Fichiers lock, PID, ports | Écarter comme implémentation | Les locks fichiers et la vérification de PID ne suffisent pas au contrat ForgeLocal ; seuls les invariants seront re-spécifiés. |
| `lib/browser-lifecycle.js` | `9aefb2e999108ed719d19f7061032192ab097f0c73e4c2579bf8c3660b2cc4f4` | Cycle de vie navigateur | Réimplémenter | À réécrire après stratégie de port par runtime, cleanup idempotent et test crash recovery. |

> Le registre devient exécutoire uniquement lorsque chaque hash, le commit autoritatif, la date de revue, le responsable nommé, les licences transitives et la classification du scan sont ajoutés. Avant cela, **aucun portage ne peut commencer**.

## Invariants non négociables pour un futur portage

Le Core Go doit posséder le cycle de vie complet. Toute réservation de port doit être définie par runtime, limitée par timeout, confirmée par l’endpoint effectivement attaché au processus, puis libérée par un cleanup idempotent. Une reprise après crash ne doit jamais relancer silencieusement une instance inconnue.

Le dashboard ne possède aucun lock, aucune file de lancement et aucune décision de runtime. Camoufox reste un candidat non lançable jusqu’à qualification séparée et ne peut pas recevoir de preuve d’un autre runtime.
