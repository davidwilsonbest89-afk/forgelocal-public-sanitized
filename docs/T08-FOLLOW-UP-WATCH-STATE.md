# État de suivi pendant l'attente T07-R — document non normatif

**Référence :** `T08-WATCH-20260816-001` · **Date :** 16 août 2026
**Nature :** document informatif, non normatif. Il ne crée aucun statut, aucune exigence et
aucun certificat. Les statuts canoniques restent ceux du CDC v3.9.7 et du registre des jalons.

## 1. État des jalons T00–T08

| Jalon | Objet | Statut canonique | Preuve de référence |
|---|---|---|---|
| T00 | Intégrité workspace, RC gelé et outils | Validé (préflight) | Proofs T05/T06 |
| T01 | Baseline Go et tests sélectionnés | Validé | Proofs bootstrap |
| T02 | Migrations SQLite | Validées | Flux produit T05/T06 |
| T03 | API bootstrap et loopback | Validé | Archive r10 |
| T04 | Build React, absence de persistance token | Validé | Proofs bootstrap |
| T05 | `BOOTSTRAP-RO-01` dashboard → Core loopback | **`BOOTSTRAP_RO_APPROVED_VERIFIABLE`** | Archive r10 + Playwright |
| T06 | Groupes/Runtimes lecture seule SQLite/Core | **`T06_APPROVED_VERIFIABLE`** | Archive T06 globale (Core + dashboard scannés) |
| T07 | Provenance Camoufox | **`T07_PROVENANCE_BLOCKED_PENDING_EVIDENCE`** | Demande T07-R transmise ; attestation propriétaire et revue indépendante en attente |
| T08 | Fiabilité Core : queue/locks/cleanup/recovery | **`T08_APPROVED_VERIFIABLE_LOCAL`** | `t08-r2-final.zip` SHA-256 `4918ac9876545904c822ff72fb3dfcc4f8b12f6fb2214452e308a39b4c0719bb` |

## 2. Preuves T08 déjà disponibles (figées, ne doivent pas être modifiées)

| Élément | Référence |
|---|---|
| Archive de preuves | `t08-r2-final.zip` — SHA-256 `4918ac98…19bb` (deux copies identiques vérifiées) |
| Code produit T08 | Commit `99a22f5` — `internal/launch/` (5 fichiers Go, clean room) |
| Commits documentaires T08 | `0e6a95f` (rapport initial), `e10a48c` (harmonisation spec + preuve cleanup), `903f6bd` (rapport final auto-cohérent), `f3a19df` (consignation de la validation) |
| Spécification | `docs/T08_CONCURRENCY_SPEC.md` (commande canonique `./internal/launch`) |
| Rapport final 16 champs | `docs/T08-R2-FINAL-REPORT.md` |
| Résultats tests | 13/13 PASS sous `go test -count=1 -race` (log `test-out.log`), `go vet` exit 0, `go build` exit 0 |
| Scan de sécurité | Gitleaks 8.18.4 JSON `[]` sur le snapshot 5 fichiers Go |
| Bug critiques fermés | Deadlock `ErrQueueFull`, data race `cancelSession`/`begin.func1` |
| Chaîne de traçabilité | Baseline `31a51e9` → code `99a22f5` → docs → consignation `f3a19df` |
| Branche / dépôt | `forgelocal-product-v0.3` sur `github.com/boucheriechefimane-cmd/IPcache` |

## 3. Dépendances bloquantes T07 (seul blocage empêchant la suite)

- Le registre Camoufox reste `integration_state=provenance-qualification-blocked` dans
  `docs/component-rights-register.json`.
- Les preuves externes T07-R attendues : attestation propriétaire redacted (snapshot,
  SHA-256 de l'archive, droits `internal_use`/`modification`/`redistribution`, licence/notices,
  six booléens de couverture de revue) et confirmation de la relectrice indépendante
  (@hajarbenmlih91-cloud) sur l'archive `camoflox-redacted.zip` et son rescan.
- Règle fail-closed : toute divergence de triage ou champ manquant donne `UNKNOWN` et maintient
  le blocage. Le validateur T07-R prépare la vérification de complétude, pas une décision.

## 4. Conditions d'ouverture de T09 (les deux sont nécessaires)

1. **T07 officiellement débloqué** par la revue indépendante, avec mise à jour du registre
   canonique (`docs/component-rights-register.json`) ;
2. **Autorisation explicite d'ouverture de T09** donnée par la relectrice indépendante.

En l'absence de ces deux conditions : T09–T21 restent bloqués.

## 5. Actions explicitement interdites pendant l'attente

- Modifier le code T08 (`internal/launch/`) ou les commits validés ;
- Recréer ou modifier les archives T08 (`t08-r2-final.zip`) ;
- Démarrer T09 ou toute implémentation runtime, port, proxy, backup/restore, import ;
- Lancer un navigateur, un runtime réel ou Camoufox (Camoufox candidat non lançable) ;
- Modifier l'UI dashboard au-delà de la préservation des preuves ;
- Modifier les répertoires RC gelés (`release/back01-minimal`, `dist/back01-minimal`) ;
- Changer le CDC au-delà de la préservation documentaire ;
- Interpréter seul les preuves externes T07-R ou changer le registre sans revue ;
- Produire automatiquement `T07_APPROVED` ou une autorisation T09 ;
- Toute release publique ou mutation du statut release.

## 6. Actions autorisées pendant l'attente

- Préserver les preuves existantes (aucune modification, uniquement protection) ;
- Réceptionner les références externes T07-R redacted et produire un **rapport de
  complétude brut** (champs reçus/manquants, cohérence des hashes, droits déclarés,
  redistribution, références licence/notices, booléens de revue, deux décisions de triage) ;
- Transmettre ce rapport pour revue indépendante, sans écrire de décision.

## 7. Statuts globaux inchangés

`PUBLIC_RELEASE_BLOCKED` · `SCAN_BLOCKED_UNKNOWN` · pilote suspendu · cinq gates publics
actifs · aucun composant Camoufox intégré ni exécuté.
