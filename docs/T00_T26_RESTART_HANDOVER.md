# ForgeLocal — Handover de reprise T00–T26

> **Point de reprise de code :** tag `t26-simulated-proxy-provider-2026-08-20`, commit `930003ca95a934fd996c94ae897693ffb6be21fb`.

Ce document est le point d’entrée de reprise ForgeLocal. Le README racine BrowseForge est historique et ne définit pas les autorisations, gates ou périmètres ForgeLocal actuels.

## Artefacts obligatoires

| Ordre | Artefact | Source privée |
|---:|---|---|
| 1 | Wrapper T00–T23 + sidecar | Release `t00-t23-public-continuity-window-20260820`, désormais privée. |
| 2 | Kit T24-SR + sidecar | Release `t24-sr-dependency-remediation-2026-08-20`. |
| 3 | Kit T25 + sidecar | Release `t25-synthetic-cookie-fixtures-2026-08-20`. |
| 4 | Kit T26 + sidecar et bundle T26 + sidecar | Release `t26-simulated-proxy-provider-2026-08-20`. |
| 5 | CDC, politique et registres | Wrapper T00–T23, `docs/BASELINE_DISCOVERY_POLICY.md`, contrats et rapports de releases. |

Les empreintes attendues figurent dans `CONTINUITY_PRIMARY_SHA256SUMS`. Les releases sont privées : un repreneur doit disposer de l’autorisation GitHub correspondante.

## État qualifié

| Période | État |
|---|---|
| T00–T23 | Historique et preuves dans le wrapper, tête T23 `df96953`. |
| T24 / T24-SR | `T24_APPROVED_VERIFIABLE_LOCAL`. |
| T25 | `T25_APPROVED_VERIFIABLE_LOCAL`, fixtures de cookies synthétiques uniquement. |
| T26 | `T26_APPROVED_VERIFIABLE_LOCAL`, catalogue fournisseur strictement simulé. |
| T27 | Non démarré ; décision produit et baseline distinctes obligatoires. |

## Reprise

Télécharger les assets privés dans un répertoire d’artefacts, vérifier les sidecars puis lancer :

```bash
bash scripts/replay-t00-t26.sh /chemin/vers/artefacts /tmp/forgelocal-t00-t26-replay
```

Le script vérifie la chaîne, clone le bundle T26, checkout le tag, exécute `git fsck --full`, `go test -count=1 -race ./...`, vet et build ; il journalise les commandes, chemins, UTC et codes de sortie.

## Limites permanentes

Les cookies réels, secrets, profils, tokens, sessions, bases SQLite, runtime, proxy réel, fournisseur réseau et release ne sont pas des entrées du kit de reprise. T25/T26 ne les autorisent pas.

Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` demeurent obligatoires.

Chaque futur lot doit ouvrir son propre `BASELINE_DISCOVERY_RAW.log` avant toute écriture.
