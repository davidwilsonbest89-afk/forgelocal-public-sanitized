# T20-NCF — Notes & Custom Fields Core Foundation

**Base :** `cc2bfbd274eb584f5d2076d32ee451b7f79b8b51`
**Périmètre sélectionné :** premier lot fonctionnel local suivant l’ordre CDC.
**État de travail :** `T20_NCF_IMPLEMENTED_PENDING_FULL_EVIDENCE`

## Autorisation et ordre CDC

L’ordre du CDC impose **Notes + Custom Fields** avant les templates. L’instruction explicite du propriétaire après l’audit T20-CANDIDATE autorise l’exécution de ce premier lot local. Le lot suivant, Templates, reste exclu jusqu’à la clôture probante de T20-NCF.

## Périmètre fermé

T20-NCF ajoute seulement des métadonnées non sensibles de profil dans le Core existant : une note et des champs `text`, `number`, `boolean` ou `select`. Les mutations passent par une route locale authentifiée, protégée par le guard loopback, le verrou par profil et l’audit redacted.

| Inclus | Exclu explicitement |
|---|---|
| Validation typée et bornée | Templates, clone, historique et bulk |
| Persistance dans le store Profile canonique actuel | Migration globale du Profile Store vers SQLite |
| Endpoint Core local audité | Dashboard, token persistant, secrets et proxies |
| Tests positifs, négatifs, lifecycle, redaction et `-race` | Import/export global, backup/restore et runtime |

## Écart CDC déclaré, non masqué

Le CDC GAP-002 vise à terme une persistance exportable intégrée à l’architecture SQLite métier. La lignée locale `cc2bfbd` conserve encore le modèle Profile canonique sous `profile.json`. T20-NCF **ne prétend pas fermer cet écart d’architecture** : il implémente et qualifie le contrat de métadonnées dans le store canonique réellement utilisé, tout en laissant `SQLite migration/export` au statut `NOT_IN_SCOPE` de ce sous-lot. Il est donc interdit de présenter T20-NCF seul comme `GAP-002 COMPLETE`.

## Critères de sortie

La preuve T20-NCF exige : mutation unique par profil, persistance après réouverture du store, validation de tous les types et des options `select`, refus sur profil archivé, refus hors loopback, erreur machine-readable, audit sans note ni valeur de champ, tests `-race`, `go vet`, build, scan Gitleaks de la plage exacte, commit, tag, bundle, clone neuf et manifeste.

## Invariants

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés. T20-NCF ne lance aucun navigateur, Camoufox, proxy réel ni coffre natif.
