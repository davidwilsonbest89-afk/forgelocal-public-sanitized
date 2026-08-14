# Rapport final de validation — ForgeLocal BACK-01

**Date de validation :** 14 août 2026

**Branche :** `forgelocal-back01`

**État :** validation technique locale réussie, sous réserve de la validation native du coffre système sur l’OS cible.

## Décision

Le flux **AC-BACK-01** est désormais prouvé de bout en bout dans le sandbox local : création d’un profil, sauvegarde chiffrée, modification de la source, restauration sous un nouvel identifiant, contrôle de l’isolation, puis relance effective du répertoire `browser-data` restauré dans Chromium local en mode headless sur `about:blank`.

L’artefact minimal BACK-01 a été reconstruit avec un manifeste de provenance, les migrations, les notices de licences, les checksums et les preuves de relance. Le runtime Chromium utilisé pour le test est **externe à l’artefact** ; l’artefact ne distribue ni navigateur, ni runtime Camoufox, ni modules de fingerprinting, humanization, MCP, extensions ou workflow.

| Domaine | Résultat | Preuve principale |
|---|---|---|
| Création, chiffrement et restauration isolée | **Vert** | `ac_back_01_with_chromium.out` |
| Relance Chromium du profil restauré | **Vert** | `ac_back_01_with_chromium.out` |
| Tests BACK-01 ciblés avec détecteur de course | **Vert** | build minimal et sorties versionnées |
| Analyse Gosec du périmètre minimal | **0 alerte** | `minimal_artifact_build.out` |
| Scan de secrets du staged et de l’archive extraite | **0 fuite** | `gitleaks_minimal_*.log` |
| Checksums internes et checksum de l’archive | **Verts** | `minimal_artifact_final_validation.out` |
| Suite Go complète | **Verte** | `go_test_all_after_persona_fix.out` |
| Matrice SystemVault native | **Non concluante dans ce sandbox** | hôte sans session de coffre utilisateur déverrouillée |

## Preuve de relance runtime

Le test suivant a été exécuté avec succès :

```bash
export GOTOOLCHAIN=go1.25.13
go test ./internal/api -run '^TestBackupV1CreateModifyRestoreIsolation$' -count=1 -v
```

Le test démarre Chromium avec les options `--headless=new`, `--no-sandbox`, `--user-data-dir=<profil-restauré>`, `--dump-dom` et `about:blank`. Il n’effectue donc aucune navigation vers un service externe. Le runtime évalué est Chromium `151.0.7922.71`, avec le binaire réel `/usr/lib/chromium/chromium` et l’empreinte SHA-256 `ad69c6632131d3a8b0e7395c3bb36d05cad6a2c650ecfa7ebe2e8dcba955c2de`.

> Cette approbation est limitée à la preuve E2E locale et à cette version exacte du runtime. Elle ne vaut pas approbation de Camoufox ni autorisation de distribuer Chromium dans ForgeLocal.

## Artefact minimal final

| Attribut | Valeur |
|---|---|
| Archive | `forgelocal-back01-core-0.1.0-back01-bd1fcac-linux-amd64.tar.gz` |
| Commit de provenance du contenu | `bd1fcac238ac32aa49aa28c1cb9e58c46194838b` |
| Outil de compilation | `go1.25.13` |
| SHA-256 de l’archive | `bd923dfdf15e1a069bbde71f1542fd2067be3d22b3e0f85bc6b0f8a84fcdaa12` |
| Relance AC-BACK-01 dans le manifeste | `true` |
| Runtime embarqué | `false` |
| Dépendances internes autorisées | `internal/backup`, `internal/profile`, `internal/secrets` |
| Exclusions vérifiées | `browser`, `fingerprint`, `humanize`, `mcp`, `runtime`, `workflow` |

Le script `scripts/build-back01-minimal.sh` échoue désormais explicitement si l’analyseur de sécurité est introuvable, si la fermeture des dépendances internes sort du périmètre autorisé, si les preuves de relance manquent alors que la revendication est activée, ou si un fichier interdit est importé. Il produit également un fichier `SHA256SUMS` qui exclut son propre contenu afin que sa vérification soit déterministe.

## Correction du défaut historique Persona

Le défaut `BASE-PERSONA-UTC-US` a été corrigé dans un commit séparé de test. La fixture utilisait un proxy loopback volontairement inaccessible mais héritait de `TZ=UTC`, ce qui produisait une Persona incohérente avec la région proxy `us-ny`. La fixture vide désormais les replis de l’hôte afin que le repli de région proxy fournisse `America/New_York` avec la provenance attendue `proxy_region_fallback`.

Les contrôles suivants sont verts après cette correction :

```bash
go test ./internal/browser -run '^TestLaunchChromiumAssemblesProxyFingerprintArgsWithoutLaunchingBrowser$' -count=1 -v
go test ./internal/browser -count=1
go test ./... -count=1
```

## Commits livrés

| Commit | Objet |
|---|---|
| `1e8473f` | Binaire minimal, matrice SystemVault, profil de distribution et preuves de relance |
| `b02f143` | Résolution portable de l’analyseur Gosec dans le script de build |
| `bf642bd` | Retrait du graphe brut de dépendances du livrable afin d’éviter les faux positifs de scan |
| `bd1fcac` | Correction d’auto-inclusion du fichier de checksums |
| `6cfd3dc` | Correction isolée de la fixture Persona `UTC` / `US` |
| `5d00f93` | Preuves finales : build, scans, checksums et suite globale |

## Condition restante avant publication publique

La publication reste **bloquée** par la validation native du `SystemVault` sur chaque OS réellement pris en charge, dans une session utilisateur déverrouillée. Le sandbox headless ne fournit pas une collection Secret Service utilisable ; il ne permet donc pas de conclure les scénarios création, lecture après redémarrage, absence/révocation, permissions insuffisantes et absence de fuite dans SQLite, `profile.json`, logs ou backups.

La migration des métadonnées de profils JSON vers SQLite métier et le branchement du dashboard React vers les routes BACK-01 demeurent également des travaux produit séparés. Ils ne bloquent pas la preuve cryptographique et de restauration BACK-01, mais empêchent de présenter ForgeLocal comme un panneau desktop complet finalisé.

## Fichiers de preuve

| Fichier | Contenu |
|---|---|
| `validation_back01_integration/final/ac_back_01_with_chromium.out` | Exécution E2E API et relance Chromium réussie |
| `validation_back01_integration/final/chromium_runtime_provenance.out` | Version, paquet et empreintes du runtime QA |
| `validation_back01_integration/final/minimal_artifact_build.out` | Tests minimaux et Gosec `Issues : 0` |
| `validation_back01_integration/final/minimal_artifact_final_validation.out` | Checksum archive, scans Gitleaks et checksums internes |
| `validation_back01_integration/final/go_test_all_after_persona_fix.out` | Suite `go test ./...` entièrement verte |
| `release/back01-minimal/RUNTIME_QA_APPROVAL.md` | Périmètre exact de l’approbation runtime QA |
| `docs/security/SYSTEM_VAULT_NATIVE_MATRIX.md` | Matrice de validation native à exécuter sur hôte cible |
