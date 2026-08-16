# Constat de validation sandbox — SystemVault et BACK-01

**Date d’exécution :** 2026-08-14 UTC  
**Branche observée :** `forgelocal-back01`  
**Commit observé :** `4f99af24f042cb74d280b7aae87434104b5181f8`  
**Statut de publication :** `PUBLIC_RELEASE_BLOCKED` — inchangé.

> Ce document est une **preuve de contrôle local et de refus sécurisé**, non une preuve native SystemVault. Il ne permet pas de lever `SYSTEMVAULT_NATIVE_PER_TARGET` ni aucun des cinq gates publics.

## Résumé de décision

Le sandbox est bien Ubuntu 24.04 `x86_64` et s’exécute sous un utilisateur non root. Toutefois, il contient `/.dockerenv` et ne fournit pas `XDG_RUNTIME_DIR`. Le script officiel refuse donc correctement de s’exécuter, avec le code attendu `3`, avant toute opération de coffre. Cette protection est conforme au runbook : aucune tentative de contournement, de simulation du coffre, de fallback en clair ou de promotion de gate n’a été réalisée.

| Contrôle | Résultat | Interprétation |
|---|---:|---|
| Utilisateur non root | `PASS` | EUID `1000`, mais ce seul prérequis ne suffit pas. |
| Ubuntu 24.04 `x86_64` | `PASS` | Cohérent avec la cible annoncée. |
| Indicateur conteneur | `FAIL` pour une preuve native | `/.dockerenv=true`. |
| D-Bus de session | Présent | Insuffisant sans session desktop complète. |
| `XDG_RUNTIME_DIR` | Absent | Précondition native non satisfaite. |
| `secret-tool` | Présent | Sa présence ne prouve ni accès ni déverrouillage du Secret Service. |
| Refus du script dans le conteneur | `PASS` | Code `3`, refus explicite validé. |
| Test ciblé BACK-01 | `PASS` | Flux API backup → restauration isolée → relance Chromium observé. |
| Suite Go complète | `PASS` | `go test ./...` vert. |
| Traçabilité RC | `PASS` | Deux chaînes cohérentes et décision publique bloquée. |
| Gate public | `PASS` en tant qu’état bloqué valide | Les cinq gates restent `PENDING`. |

## Intégrité du candidat RC

L’archive candidate n’a pas été modifiée pendant les contrôles. L’empreinte observée avant et après est identique :

```text
553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6
```

Elle correspond à `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz`.

## Contrôles effectués

### Préflight du sandbox

Le préflight assaini a produit les valeurs suivantes :

```text
os=ubuntu 24.04
arch=x86_64
euid=1000
dockerenv=true
dbus_session=present
xdg_runtime=absent
secret_tool=present
```

Le contrôle de conteneur fondé sur le cgroup ne détecte pas de marqueur, mais le script officiel vérifie également `/.dockerenv`. La présence de ce fichier suffit à disqualifier le sandbox pour une validation SystemVault native.

### Refus protégé du gate SystemVault

La commande officielle `scripts/run-systemvault-native-gate.sh` a été appelée sans élévation ni modification. Elle a retourné :

```text
refusing SystemVault release gate in a container; run on the target OS host
exit_code=3
refusal_check=passed
```

Ce résultat confirme que le garde-fou est actif. Il ne s’agit pas d’un échec produit ni d’une matrice SystemVault exécutée.

### BACK-01 ciblé et relance du runtime

La commande de sélection a identifié exactement trois tests dans `./internal/api` :

```text
TestBackupV1CreateModifyRestoreIsolation
TestLegacyRestoreEndpointRetired
TestLegacyBackupEndpointRetired
```

Le scénario complet a ensuite été exécuté explicitement :

```bash
GOTOOLCHAIN=go1.25.13 go test ./internal/api \
  -run '^TestBackupV1CreateModifyRestoreIsolation$' -count=1 -v
```

Il est passé en `0.338s`. La sortie journalise la création du backup, la restauration isolée et la relance/arrêt propre de Chromium `151.0.7922.108`, avec nettoyage du verrou de profil. La preuve préexistante E2E du candidat RC reste la seule référence de qualification runtime ; ce test sandbox n’est pas réutilisé comme preuve native de coffre.

La suite entière est également passée :

```bash
GOTOOLCHAIN=go1.25.13 go test ./...
```

### Traçabilité et décision publique

Le validateur de traçabilité a retourné :

```json
{"chains": 2, "public_release_decision": "PUBLIC_RELEASE_BLOCKED", "valid": true}
```

Le validateur de gate a confirmé une métadonnée valide avec la décision `PUBLIC_RELEASE_BLOCKED`. Les cinq gates demeurent `PENDING` : `SYSTEMVAULT_NATIVE_PER_TARGET`, `SYSTEMVAULT_ANTI_LEAK_INTEGRATED_FLOW`, `MAINTAINER_MANIFEST_SIGNATURE`, `RUNTIME_LICENSE_AND_REDISTRIBUTION_REVIEW` et `OS_COMPATIBILITY_EVIDENCE`.

> Un code de sortie nul du validateur signifie que le fichier de décision est cohérent ; il ne signifie pas que la publication est approuvée.

## Fichiers de preuve locaux

| Fichier | SHA-256 |
|---|---|
| `sandbox-systemvault-preflight-20260814.txt` | `7fb7b98374d8fc4f264e92fa331c8ef432f1e20c1a1b453f8a78755f6d816fda` |
| `sandbox-systemvault-gate-refusal-20260814.txt` | `fbe6c994d40d64fd8a122ff9ac8da2669458a727dbeea029030a676f7f6a8ed4` |
| `sandbox-back01-test-list-20260814.txt` | `6a8bdef39e1cd8c1bd675c90e66bdda2caaf62be3afb3195893da5a86b1dc0e5` |
| `sandbox-back01-targeted-test-20260814.txt` | `d90e30db79c86087b1d7563e2019e855b51aa426b6a1d43647027ef33b313250` |
| `sandbox-go-test-all-20260814.txt` | `ebefd7732a909328be5b1bd477d2ce21411b5ee856963cb6d9f4c0457238b467` |
| `sandbox-release-gates-20260814.txt` | `5bcef15a4bdaa95876bf0fd56fa19568448e1a853bf70dd6101ad81c02c627e6` |

## Limite irréductible et prochaine action

Pour lever `SYSTEMVAULT_NATIVE_PER_TARGET`, il faut toujours une session graphique utilisateur **hors conteneur**, avec Secret Service déverrouillé et les variables D-Bus de session présentes. Le sandbox ne peut pas remplacer cette preuve. La prochaine action reste l’exécution du runbook sur une VM Ubuntu 24.04 amd64 native/persistante — par exemple l’option Oracle Always Free documentée dans `ORACLE_FREE_SYSTEMVAULT_PROTOCOL.md` — suivie de l’export des quatre fichiers assainis attendus.

Aucun champ de `PUBLIC_RELEASE_GATE_STATE.json`, aucun artefact candidat et aucun manifeste n’a été modifié par ces contrôles.
