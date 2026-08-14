# Approbation QA de runtime — BACK-01

## Décision

Le runtime suivant est approuvé **uniquement pour la preuve E2E locale BACK-01 dans ce sandbox**. Il n’est pas inclus dans l’artefact minimal `forgelocal-back01-core`, n’est pas le runtime par défaut de ForgeLocal et ne constitue pas une approbation de Camoufox.

| Attribut | Valeur |
|---|---|
| Runtime | Chromium standard Ubuntu |
| Binaire wrapper | `/usr/bin/chromium` |
| Binaire réel | `/usr/lib/chromium/chromium` |
| Version rapportée | `Chromium 151.0.7922.71 built on Ubuntu 24.04.4 LTS` |
| Paquet | `chromium` `151.0.7922.71-1xtradeb1.2404.1` |
| Architecture | `amd64` |
| SHA-256 wrapper | `36cbbb620daeb933ae7861de07fcff05b5e1f7527303b5992459b8aa6707b845` |
| SHA-256 binaire réel | `ad69c6632131d3a8b0e7395c3bb36d05cad6a2c650ecfa7ebe2e8dcba955c2de` |
| Navigation externe | Aucune ; test limité à `about:blank` |

## Preuve E2E obtenue

La commande suivante a réussi :

```bash
go test -count=1 -run '^TestBackupV1CreateModifyRestoreIsolation$' -v ./internal/api
```

Le test crée un profil source, écrit un état local, crée une sauvegarde chiffrée, modifie l’original, restaure sous `target-api`, vérifie que source et cible restent distincts, puis démarre Chromium en mode headless sur le `browser-data` restauré avec `about:blank` uniquement. Le résultat est `PASS`.

## Exigences avant une approbation de release

1. Refaire l’inventaire de provenance, signature de paquet et licence sur l’hôte de release.
2. Vérifier le checksum exact du binaire livré, et non uniquement celui du wrapper.
3. Évaluer tout changement de version Chromium comme un nouveau candidat, avec nouveau checksum et E2E.
4. Garder Camoufox hors de cette approbation tant que sa licence, ses artefacts, sa provenance et sa signature ne sont pas validés séparément.
5. Rattacher la preuve SystemVault à une session OS utilisateur déverrouillée avant publication.
