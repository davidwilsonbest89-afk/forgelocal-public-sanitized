# GOSEC-R4 — Préflights environnementaux

## Statut observé

Le préflight du 26 août 2026 confirme un environnement Linux x86_64. Docker, Buildx, xattr Darwin et Camoufox ne sont pas disponibles. Une surface SystemVault est détectée, mais aucun test natif n’est autorisé dans ce lot. Les ports réservés 19281, 3001 et 19282–19287 sont libres. Aucun service ForgeLocal n’est laissé actif.

| Environnement | Statut | Raison |
|---|---|---|
| Docker/Buildx | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Daemon/outils absents; aucun privilège forcé. |
| Camoufox natif | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Binaire/runtime ciblé absent. |
| Darwin/xattr/GUI | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Hôte Linux; les chemins Darwin restent statiques et scanner-ouverts. |
| SystemVault natif | `NATIVE_SYSTEMVAULT_NOT_TESTED` | Surface détectée, mais autorisation et environnement produit dédiés manquants. |
| Proxy/cookies réels | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Aucun compte, cookie, proxy ou donnée utilisateur autorisé. |
| Release | `release_authorized=false` | Aucune préparation ou publication de release autorisée. |

## Protocoles préparés sans exécution

Le protocole Docker/Buildx devra construire les Dockerfiles dans un daemon dédié, inspecter l’utilisateur effectif, les modes de fichiers, le `HEALTHCHECK`, l’historique et l’image Trivy, puis détruire les artefacts temporaires. Le protocole SystemVault devra utiliser un coffre natif de test, des secrets synthétiques, une redaction vérifiable, un rollback et une destruction du coffre. Le protocole Camoufox/Darwin devra utiliser une machine ciblée, un profil jetable, des archives synthétiques et une validation de fermeture sans cookie ni proxy réel. Le protocole proxy/cookies devra rester limité à des fixtures synthétiques et à un proxy de test explicitement autorisé. Le protocole release devra être exécuté séparément, avec revue des licences, provenance, SBOM, signature et rollback.

Aucun de ces protocoles ne peut lever sa gate sur la seule base de ce document ou des scans source-only. Les critères de succès, les journaux redacted et les décisions d’autorisation doivent être produits dans leur environnement réel.

## Gates conservées

```text
PUBLIC_RELEASE_BLOCKED=true
SCAN_BLOCKED_UNKNOWN=true
NATIVE_SYSTEMVAULT_NOT_TESTED=true
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```
