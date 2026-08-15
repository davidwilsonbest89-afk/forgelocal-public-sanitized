# T08 — Spécification du module de concurrence ForgeLocal (avant tout code)

**Référence :** `T08-SPEC-20260815-001` · **Date :** 2026-08-15 · **Auteur :** ForgeLocal maintainer (décision enregistrée par Manus AI)
**Autorisation parente :** `T07-DECISION-20260815-001` — `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION`

## Décision de réimplémentation enregistrée (avant tout code)

La décision `reimplementer` est enregistrée dans le registre canonique `docs/component-rights-register.json` (registre `FL-COMP-RIGHTS-20260814-R1`, SHA-256 `3723b45f3b46a3a1d59e6185126a8576ed089bf35f5ee7863c15c34f77d467e3`) **avant toute ligne de code T08**, conformément à l'autorisation T07.

| Champ | Valeur |
|---|---|
| Module étudié | `lib/concurrency.js` |
| Hash source (dans `camoflox-source.zip`) | `b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7` |
| Snapshot source | `camoflox-source.zip` · SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2` |
| Décision | `reimplementer` |
| Raison | Contrat de concurrence Go pur (queue bornée, limite globale, verrous par profil, temporisations) : premier module autorisé par `T07-DECISION-20260815-001` |
| Rôle | inspiration de contrat uniquement (`concurrency-contract-inspiration-only`) |
| Exclusions explicites | code Node/Electron ; lancement runtime ; cycle de vie navigateur ; ports ; queue de lancement ; isolation de processus ; activation Camoufox |
| Chaîne T07-R1 | révision privée attestée → hash du fichier étudié → décision `reimplementer` (registre) → futur commit ForgeLocal → tests |

## Principe de clean room

Le module est écrit en Go sans importer, copier ni consulter le code candidat pendant la rédaction : la spécification ci-dessous décrit un **contrat fonctionnel ForgeLocal indépendant**. L'étude conceptuelle du hash enregistré s'est faite en amont ; le code produit ne reproduit aucune expression du candidat.

## Périmètre fonctionnel (spécification de contrat)

Le module `internal/concurrency` (chemin contrôlé existant) doit fournir un orchestrateur de concurrence par profil, strictement limité aux opérations internes au Core (aucun lancement, aucun réseau, aucun cycle de vie navigateur).

1. **Queue bornée.** Une file d'exécution par profil, de capacité configurable, rejette immédiatement une demande excédentaire avec une erreur explicite plutôt que de la mettre en attente illimitée.
2. **Limite globale.** Une limite concurrente globale, configurable, s'applique à l'ensemble des opérations orchestrées, en complément des files par profil.
3. **Verrou par profil.** Une opération par profil s'exécute sans interférence avec les opérations du même profil ; les profils différents restent indépendants.
4. **Temporisation (timeout).** Toute opération soumise possède un délai maximal ; une opération échue est signalée en erreur et libère ses ressources.
5. **Annulation par contexte.** Toute opération accepte un `context.Context` : annulation, expiration ou fermeture libère les verrous et termine la file.
6. **Nettoyage (cleanup).** À la fermeture du module ou du Core, aucune goroutine ne fuit : files drainées ou signalées, verrous libérés, `WaitGroup` rejoint.
7. **Reprise après plantage.** Si le processus orchestre hôte est redémarré, aucune opération marquée en cours n'est reprise sans vérification : la queue redémarre à vide et journalise les abandons (aucun état de queue persisté en SQLite — la persistance d'état d'orchestration reste interdite).
8. **Journal d'audit redacted.** Les événements (soumission, acceptation, rejet borné, expiration, annulation, abandon au redémarrage) sont journalisés sans secret, sans chemin absolu, sans valeur d'alerte ni trace de la source Camoflox.

## Critères d'acceptation et exigences de preuve

| Contrôle | Preuve exigée |
|---|---|
| Queue bornée | test `t.Parallel` : remplissage jusqu'à capacité, 17ᵉ requête refusée avec erreur dédiée |
| Limite globale | test : limite 3, 5 soumissions simultanées, exactement 3 en cours, retour au calme |
| Verrou par profil | test : deux profils A/B, opération A n'attend pas B, double soumission A sérialisée |
| Timeout | test : opération lente, expiration, erreur `timeout`, verrou libéré |
| Contexte | test : `context.WithCancel` annule, erreur `canceled`, ressources libérées |
| Race detector | `go test -count=1 -v -race ./internal/concurrency` sans warning ; test `TestConcurrentStress` avec `t.Parallel` et goroutines multiples |
| Cleanup | test : fermeture pendant exécution, `go leak` ou `WaitGroup` rejoint dans un délai borné |
| Recovery | test : redémarrage simulé, queue vide, journal d'abandon présent, aucune reprise implicite |
| Audit redacted | inspection du journal produit : zéro secret, zéro sentinelle, zéro référence Camoflox |
| Scan | Gitleaks 8.18.4 JSON `[]` sur le delta T08, avant toute intégration |

## Interdictions maintenues (non négociables)

Aucun port d'écoute, aucun Camoufox, aucun cycle de vie navigateur, aucun proxy, backup, restauration, mutation UI, lancement/runtime ou activation candidat. Aucun artefact de cette phase ne peut être présenté comme une qualification de runtime ni une autorisation de release. `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN` (BACK-01), pilote suspendu et les cinq gates publics restent en vigueur.

## Prochaines étapes contrôlées

T08-F2 : écrire le code Go de la spécification ci-dessus (clean room, sans import du candidat). T08-F3 : passer l'ensemble des critères de preuve, y compris `-race` et scan Gitleaks. T08-F4 : relecture sécurité (`@boucheriechefimane-cmd`) puis release (`@davidwilsonbest89-afk`) avant toute intégration au RC ; la relectrice indépendante est informée à chaque jalon.
