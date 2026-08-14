# Qualification QA runtime — Chromium 151.0.7922.108

**Statut :** `TECHNICAL_RUNTIME_E2E_PASSED_PUBLIC_RELEASE_BLOCKED`
**Date de qualification :** 14 août 2026
**Plate-forme observée :** Ubuntu 24.04.4 LTS, `amd64`

> Ce document qualifie Chromium uniquement comme **runtime externe de test** pour AC-BACK-01. Chromium n’est pas distribué dans l’artefact minimal ForgeLocal BACK-01 et cette qualification n’autorise pas une publication publique.

| Élément | Valeur vérifiée |
|---|---|
| Paquet principal | `chromium` `151.0.7922.108-1xtradeb1.2404.1` |
| Paquet requis | `chromium-common` `151.0.7922.108-1xtradeb1.2404.1` |
| Dépôt | XtraDeb Apps PPA, Ubuntu Noble, composant `main` [1] |
| Empreinte PGP publique | `5301FA4FD93244FBC6F6149982BB6851C64F6880` |
| Chaîne vérifiée | `InRelease` signé → `Packages.gz` haché → checksum des deux `.deb` |
| Binaire testé | `/usr/bin/chromium` → `/usr/lib/chromium/chromium` |
| Version exécutée | `Chromium 151.0.7922.108 built on Ubuntu 24.04.4 LTS` |
| AC-BACK-01 | **Réussi** : backup chiffré, altération source, restauration isolée, relance directe `about:blank`, arrêt propre et nettoyage des locks |

Les paquets exacts, les index signés et leurs empreintes sont archivés sous `validation_back01_integration/candidates/`. Le lock machine-readable, incluant les checksums des deux paquets, est `RUNTIME_CANDIDATE_CHROMIUM_151.0.7922.108.json`.

## Décision

La qualification technique de ce candidat est **verte** pour le scénario AC-BACK-01 local. Elle ne transforme pas `PILOT_LOCAL_APPROVED` en autorisation de publication publique. Les gates suivants restent obligatoires : matrice native SystemVault sur chaque OS annoncé, preuve intégrée anti-fuite, revue de redistribution/licence du runtime, manifest de release signé par une clé mainteneur et matrice de compatibilité limitée aux OS réellement testés.

## Références

[1]: https://launchpad.net/~xtradeb/+archive/ubuntu/apps "XtraDeb Apps PPA sur Launchpad"
