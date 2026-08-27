# MAX LOCAL — État d’environnement

Date de campagne : 2026-08-27 UTC. Le dépôt ciblé est `davidwilsonbest89-afk/forgelocal-public-sanitized`, branche corrective `validation/final-secret-remediation`. Le HEAD de départ du clone dédié était `f685c88e1aec4a58dc04a3ba2a4af5f1e5d9e52a` ; le clone local de travail a été créé sur `validation/max-local-execution` et n’a pas été poussé.

| Élément | Statut réel | Preuve |
|---|---|---|
| Go système | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | `go version` exit 127 |
| Toolchain Go local 1.26.7 | disponible et utilisé | tests, race, vet et build ciblés exit 0 |
| Node / pnpm | disponibles | Node 22.13.0, pnpm 10.4.1 |
| Git LFS | absent | `git lfs version` exit 1 |
| Docker client / Buildx | client 29.1.3, Buildx 0.30.1 | daemon indisponible |
| Docker daemon | arrêté, socket absent | `docker info` exit 1 ; service inactive |
| Espace disque | insuffisant pour un nouveau cycle image sûr | 3,8 Go disponibles au contrôle Docker |
| Firefox | absent | exit 127 |
| Camoufox | absent | exit 127 |
| SystemVault natif | non disponible | `secret-tool` absent ; aucune simulation présentée comme native |
| Gosec / Govulncheck / OSV Scanner | indisponibles | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` |
| Semgrep / Gitleaks / Trivy / Syft / Grype / Yamllint | exécutés | codes et compteurs dans `MAX_LOCAL_SCANNERS_RAW.log` |

Le dépôt historique `boucheriechefimane-cmd/IPcache` n’a pas pu être résolu par GitHub CLI. La lignée T07/T08 est donc `HISTORICAL_EVIDENCE_UNVERIFIED`; aucun PASS historique n’est transféré au dépôt actuel.

Le package historique a été vérifié sans reconstruction ni remplacement. Le build Docker image et le bridge n’ont pas été relancés dans cette campagne en raison du daemon arrêté et de l’espace disponible ; ce résultat est `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`, non un PASS.
