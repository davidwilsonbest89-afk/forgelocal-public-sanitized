# Triage mainteneur redacted — alerte de scan de l’archive RC

**Date :** 14 août 2026  
**Statut :** `UNKNOWN — PILOT SUSPENDED`  
**Autorité de suspension :** ForgeLocal Release Maintainer — Repository Owner  
**Portée :** candidat RC BACK-01 Chromium `151.0.7922.108` uniquement

## Décision

> **Conclusion de triage : `UNKNOWN`.** La valeur détectée n’a pas été affichée, recopiée, décodée ou ajoutée à ce rapport. Les contrôles effectués ne suffisent pas à établir qu’il s’agit d’un faux positif non secret. Le pilote local temporaire reste donc suspendu en mode défaillant fermé.

Cette décision ne modifie pas l’archive RC, son runtime, son SBOM, son manifeste, les cinq gates ou la décision `PUBLIC_RELEASE_BLOCKED`.

## Identité gelée du candidat examiné

| Élément | Valeur |
|---|---|
| Archive | `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz` |
| SHA-256 archive | `553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6` |
| Commit source de chaîne | `67a8dfd897e540a55fc10749e1f2ef85b8356a8b` |
| Runtime QA | Chromium `151.0.7922.108-1xtradeb1.2404.1` (`amd64`) |
| Décision de publication | `PUBLIC_RELEASE_BLOCKED` |
| État pilote | `PILOT_LOCAL_SUSPENDED_PENDING_SECRET_SCAN_TRIAGE` |

## Occurrence redacted

| Champ | Valeur redacted |
|---|---|
| Détecteur | Gitleaks, build local identifié par l’outil comme version de build |
| Règle | `generic-api-key` |
| Chemin dans l’archive | `evidence/RUNTIME_PROVENANCE.out` |
| Ligne | `8` |
| Rapport de scan | `sandbox-gitleaks-rc-archive-20260814.json` |
| SHA-256 du rapport | `b9220e555155c055c40b3f8933d71fbbb4a780f6084ad8fcc129e681ac9e30ce` |
| Valeur brute | **Non affichée, non recopiée et non incluse** |

La chaîne signalée est une ligne de 64 caractères qui n’est pas une empreinte SHA-256 hexadécimale simple. Sa forme seule ne permet pas de déterminer s’il s’agit d’un identifiant public de provenance, d’un artefact non sensible ou d’une valeur secrète. Aucun test de décodage n’a été journalisé afin d’éviter de manipuler ou propager une valeur potentiellement sensible.

## Provenance du fichier et exposition constatée

Le script de build minimal copie une preuve runtime fournie dans l’entrée `RUNTIME_PROVENANCE` vers `evidence/RUNTIME_PROVENANCE.out` de l’archive. Pour le candidat concerné, la même occurrence est présente dans la preuve source `validation_back01_integration/final/chromium_runtime_candidate_108_provenance.out`.

| Contrôle redacted | Résultat |
|---|---:|
| Commit ayant introduit l’occurrence dans l’historique | 1 : `09df7d32247618950455729cbf5a240f7b40384c` (`release: qualify chromium candidate and add sbom gate`) |
| Chemins suivis à `HEAD` contenant l’occurrence | 1 : la preuve `chromium_runtime_candidate_108_provenance.out` |
| Occurrences dans les changements staged lors du contrôle | 0 |
| Occurrence dans l’archive RC | 1, au chemin redacted ci-dessus |
| Impression de la valeur dans les sorties de triage | 0 |

> Le contrôle montre que la valeur est déjà **committée** dans une preuve de provenance et **archivée** dans le candidat RC. Il ne permet pas de conclure que la valeur est inoffensive. Il serait donc faux d’affirmer que « rien de sensible n’a été committé ou archivé » avant une classification mainteneur indépendante.

## Classification et conditions de reprise

| Conclusion possible | Condition minimale | Action obligatoire |
|---|---|---|
| `REAL_SECRET` | Le propriétaire de la valeur ou une source autoritative confirme qu’elle authentifie un accès, une signature privée, un compte ou une ressource protégée. | Révoquer/remplacer la valeur ; inspecter l’historique ; reconstruire un nouveau candidat ; régénérer SBOM, checksums, manifeste, traçabilité et preuves E2E dépendantes. |
| `FALSE_POSITIVE` | Le mainteneur établit, hors rapport public, la nature non secrète de la valeur et son rôle précis de provenance. | Versionner une justification sans valeur ; réexécuter le scan ; fournir une exception **minimale et reproductible** seulement si elle n’occulte aucune nouvelle valeur sur le même chemin ; faire revoir la décision indépendamment. |
| `UNKNOWN` | La nature n’est pas démontrée de façon indépendante et redacted. | Maintenir la suspension du pilote, le gel RC et `PUBLIC_RELEASE_BLOCKED`. |

Aucune exclusion globale de Gitleaks n’est admise. Une exclusion limitée uniquement par chemin est insuffisante si elle masque tout futur motif `generic-api-key` dans le fichier. Si la seule exception techniquement possible exige d’inclure la valeur brute dans la configuration, le candidat doit être traité comme non publiable et remplacé par une nouvelle chaîne de preuves qui ne contient pas cette valeur.

## Conditions de clôture du triage

Le dossier ne pourra passer de `UNKNOWN` à `FALSE_POSITIVE` ou `REAL_SECRET` qu’avec une décision versionnée contenant : le propriétaire responsable, l’horodatage, la justification redacted, le hash de l’archive examinée, le hash du rapport de scan, le re-scan reproduit, le résultat du contrôle de l’historique et une revue indépendante.

Jusqu’à cette clôture, il est interdit de réactiver le pilote, de poursuivre SystemVault pour ce candidat comme s’il était propre, de modifier l’archive figée ou de lever un gate public.

## Références internes

| Référence | Fichier |
|---|---|
| R1 | `validation_back01_integration/final/secret-triage-freeze-metadata-20260814.txt` |
| R2 | `validation_back01_integration/final/secret-triage-exposure-inventory-20260814.txt` |
| R3 | `validation_back01_integration/final/sandbox-gitleaks-rc-archive-20260814.json` |
| R4 | `scripts/build-back01-minimal.sh` |
| R5 | `release/back01-minimal/PILOT_LOCAL_TEMPORARY_AUTHORIZATION.json` |
| R6 | `release/back01-minimal/PUBLIC_RELEASE_GATE_STATE.json` |
