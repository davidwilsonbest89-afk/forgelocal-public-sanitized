# Portée de release et matrice de compatibilité — BACK-01 minimal

**Décision actuelle :** `PUBLIC_RELEASE_BLOCKED`
**Artefact pilote :** `forgelocal-back01-core-0.1.0-back01-07f603d-linux-amd64.tar.gz`
**Artefact candidat non public :** `forgelocal-back01-core-0.1.0-back01-rc1-chromium151108-linux-amd64.tar.gz`

> Une éventuelle publication future concerne exclusivement le **Core/API BACK-01 minimal**. Elle ne doit jamais être décrite comme ForgeLocal complet, un dashboard desktop final, un navigateur anti-détection, une garantie de non-détection, ni une distribution de Chromium ou Camoufox.

## Portée autorisée

| Élément | Statut | Formulation autorisée |
|---|---:|---|
| API locale authentifiée | Inclus | API locale BACK-01 liée par défaut à `127.0.0.1` |
| Backups chiffrés et restauration isolée | Inclus | Sauvegarde/restauration locale de profils avec contrôles BACK-01 |
| SQLite, recovery et audit local | Inclus | Métadonnées BACK-01 locales et réconciliation contrôlée |
| Coffre système | Inclus sous gate | Secrets référencés par le coffre OS ; validation native obligatoire avant publication |
| Chromium externe de QA | Hors artefact | Runtime de test seulement, avec version et provenance verrouillées |
| Camoufox | Hors gate | Aucun runtime Camoufox n’est approuvé dans cette release |
| Dashboard React / Tauri | Hors périmètre | Non livré, non validé, non annoncé |
| Migration métier JSON vers SQLite | Hors périmètre | Chantier produit distinct, non annoncé comme terminé |
| Fingerprinting, humanization, MCP, extensions, workflow | Exclus | Non embarqués dans le Core/API minimal |

## Matrice OS réellement testée

| OS / architecture | Core/API minimal | Runtime externe AC-BACK-01 | SystemVault natif | Statut revendicable |
|---|---:|---:|---:|---|
| Ubuntu 24.04.4 LTS `amd64` | Vert dans le sandbox | Vert avec Chromium `151.0.7922.108-1xtradeb1.2404.1` | Non validé : sandbox conteneurisé | Pilote local contrôlé seulement |
| Ubuntu autre version / autre architecture | Non testé | Non testé | Non testé | Aucune compatibilité annoncée |
| macOS | Non testé | Non testé | Non testé | Aucune compatibilité annoncée |
| Windows | Non testé | Non testé | Non testé | Aucune compatibilité annoncée |
| Autres distributions Linux | Non testé | Non testé | Non testé | Aucune compatibilité annoncée |

Le seul environnement avec un scénario runtime exécuté est l’Ubuntu explicitement listé. Il ne peut pas être extrapolé à une autre version, une autre architecture, une autre distribution ou un autre coffre système.

## Gates obligatoires avant annonce d’un OS

Pour chaque OS et architecture qui serait annoncée publiquement, il faut exécuter sur une session utilisateur graphique réelle et non privilégiée : création/lecture après redémarrage du Core, clé absente, clé révoquée, coffre verrouillé, permissions insuffisantes, preuve du backend natif, flux réel profil → backup → restauration, génération de `systemvault-anti-leak.json`, et vérification qu’aucune valeur secrète ne se trouve dans SQLite, profil, logs ou backup.

Le runtime externe associé doit être un candidat explicitement verrouillé avec son ou ses paquets exacts, leur checksum comparé à un index signé, la clé de signature vérifiée, et l’E2E AC-BACK-01 complet. Une mise à jour de runtime est un nouveau candidat, jamais un remplacement silencieux.

## Conditions de passage à la release publique

La décision peut passer à `PUBLIC_RELEASE_APPROVED` seulement lorsque tous les éléments suivants sont disponibles pour la portée réellement annoncée : preuves SystemVault natives vertes, `systemvault-anti-leak.json` intégré, runtime approuvé avec chaîne de provenance complète, SBOM SPDX cohérent, manifeste externe correspondant au hash de l’archive, signature détachée par la clé mainteneur publiée, et revue de licence/distribution des composants externes.
