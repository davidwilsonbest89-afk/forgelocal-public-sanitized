# ForgeLocal BACK-01 — Signature mainteneur et vérification indépendante

**Statut actuel :** `UNSIGNED_REQUIRES_MAINTAINER_KEY`. Ce document ne constitue pas une approbation de publication.

> La clé privée de release est hors dépôt, hors archive et hors sandbox. Le dépôt ne contient que les scripts de signature et de vérification, l’empreinte publique approuvée et les attestations de vérification.

## Artefact couvert

La procédure concerne uniquement le candidat `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz` et son manifeste externe `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz.release-manifest.json`. Le manifeste contient déjà le hash de l’archive, du manifeste interne et du SBOM ; il ne doit plus être modifié une fois signé.

| Objet | Contrôle attendu |
|---|---|
| Archive RC | SHA-256 conforme au manifeste externe |
| Manifeste externe | JSON valide, source et SBOM cohérents |
| Signature détachée | Fichier `.asc` du manifeste, jamais de signature implicite |
| Clé publique | Export ASCII séparé, empreinte de 40 hexadécimaux annoncée par un canal de confiance |
| Attestation | Hash du manifeste, de la signature et de la clé publique, plus résultat de vérification |

## Signature hors dépôt

Le mainteneur réalise cette étape dans son propre trousseau OpenPGP, hors de ce dépôt et avec sa clé privée protégée. Il fixe l’empreinte de la clé approuvée, puis signe le manifeste sans modifier l’archive ou le JSON signé.

```bash
export FORGELOCAL_RELEASE_SIGNING_FINGERPRINT='EMPREINTE_PUBLIQUE_40_HEXA'
/path/to/ForgeLocal/scripts/sign-release-manifest.sh \
  forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz.release-manifest.json
```

Le résultat attendu est un fichier adjacent `.release-manifest.json.asc`. Cette action doit échouer si la clé privée approuvée n’est pas disponible dans le trousseau du mainteneur. La clé privée, un export secret, un répertoire `GNUPGHOME` ou une phrase de passe ne doivent jamais être ajoutés au dépôt, au SBOM, à l’archive ou aux journaux.

## Vérification indépendante par clé publique

Un vérificateur récupère séparément le manifeste, sa signature et la clé publique exportée. Il compare d’abord l’empreinte publique annoncée à un canal de confiance puis utilise un trousseau temporaire ne contenant aucune clé privée.

```bash
export FORGELOCAL_RELEASE_SIGNING_FINGERPRINT='EMPREINTE_PUBLIQUE_40_HEXA'
/path/to/ForgeLocal/scripts/verify-release-manifest.sh \
  forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz.release-manifest.json \
  forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz.release-manifest.json.asc \
  forgelocal-back01-maintainer-public-key.asc
```

Le script doit afficher `private_keys_present=false` et l’empreinte approuvée. Il échoue si la signature ne correspond pas, si la clé publique ne correspond pas à l’empreinte déclarée, ou si le trousseau temporaire contient une clé privée.

## Attestation de signature

Après réussite de la vérification indépendante, un mainteneur ajoute une **attestation externe**, sans modifier le manifeste signé. Elle doit contenir le nom et le SHA-256 du manifeste, de la signature et de la clé publique, l’empreinte vérifiée, le résultat de vérification et la date. Elle ne doit jamais inclure une clé privée, un token, une phrase de passe ou une valeur de coffre.

L’attestation ne lève pas les autres gates : la décision reste `PUBLIC_RELEASE_BLOCKED` tant que SystemVault natif, le flux anti-fuite, la revue de licence/runtime et la matrice OS ne sont pas intégralement verts.
