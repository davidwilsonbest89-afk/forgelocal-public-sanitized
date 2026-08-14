# Rapport final de validation — ForgeLocal BACK-01

**Date de validation :** 14 août 2026

**Branche :** `forgelocal-back01`

**État :** pilote local contrôlé approuvé ; publication publique bloquée par SystemVault natif et la provenance/signature runtime complète.

## Décision

Le flux **AC-BACK-01** est désormais prouvé de bout en bout dans le sandbox local : création d’un profil, sauvegarde chiffrée, modification de la source, restauration sous un nouvel identifiant, contrôle de l’isolation, puis relance effective du répertoire `browser-data` restauré dans Chromium local en mode headless sur `about:blank`.

L’artefact minimal BACK-01 a été reconstruit avec un manifeste de provenance, les migrations, les notices de licences, les checksums et les preuves de relance. Le runtime Chromium utilisé pour le test est **externe à l’artefact** ; l’artefact ne distribue ni navigateur, ni runtime Camoufox, ni modules de fingerprinting, humanization, MCP, extensions ou workflow.

| Domaine | Résultat | Preuve principale |
|---|---|---|
| Création, chiffrement et restauration isolée | **Vert** | `ac_back_01_with_explicit_chromium.out` |
| Relance Chromium du profil restauré | **Vert, explicitement journalisée** | binaire, version, PID, profil, `about:blank`, arrêt et locks vérifiés |
| Tests BACK-01 ciblés avec détecteur de course | **Vert** | build minimal et sorties versionnées |
| Analyse Gosec du périmètre minimal | **0 alerte** | `minimal_artifact_runtime_gate_build.out` |
| Scan de secrets de l’archive extraite | **0 fuite** | `gitleaks_runtime_gated_archive.log` |
| Checksums internes et checksum de l’archive | **Verts** | `minimal_artifact_runtime_gate_validation.out` |
| Suite Go complète | **Verte** | `go_test_all_after_runtime_gate.out` |
| Matrice SystemVault native | **Non concluante dans ce sandbox** | hôte sans session de coffre utilisateur déverrouillée |
| Provenance/signature runtime de release | **Incomplète** | canal et empreintes documentés ; signature et paquet source à archiver |

## Preuve de relance runtime

Le test suivant a été exécuté avec succès :

```bash
export GOTOOLCHAIN=go1.25.13
go test ./internal/api -run '^TestBackupV1CreateModifyRestoreIsolation$' -count=1 -v
```

Le test démarre Chromium avec les options `--headless=new`, `--no-sandbox`, `--user-data-dir=<profil-restauré>`, `--dump-dom` et `about:blank`. Il n’effectue donc aucune navigation vers un service externe et n’ouvre aucun endpoint de débogage. Sa sortie journalise explicitement le binaire, la version, le PID éphémère, le profil cible, le répertoire de données temporaire, l’absence attendue d’endpoint, `about:blank`, l’arrêt propre et l’absence des locks `Singleton*` à la fin. Le runtime évalué est Chromium `151.0.7922.71`, avec le binaire réel `/usr/lib/chromium/chromium` et l’empreinte SHA-256 `ad69c6632131d3a8b0e7395c3bb36d05cad6a2c650ecfa7ebe2e8dcba955c2de`.

> Cette approbation est limitée à la preuve E2E locale et à cette version exacte du runtime. Elle ne vaut pas approbation de Camoufox ni autorisation de distribuer Chromium dans ForgeLocal.

## Artefact minimal final

| Attribut | Valeur |
|---|---|
| Archive | `forgelocal-back01-core-0.1.0-back01-07f603d-linux-amd64.tar.gz` |
| Commit de provenance du contenu | `07f603d8d8314988b8b3279c78eceae645610c19` |
| Outil de compilation | `go1.25.13` |
| SHA-256 de l’archive | `66f785eb82dd1bef90a142e618974f1e895cf5389deb0040475877dbc596c045` |
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
| `07f603d` | Preuve de relance explicite, provenance runtime et gate SystemVault + Runtime |

## Condition restante avant publication publique

La publication reste **bloquée** par deux gates cumulatifs : la validation native du `SystemVault` sur chaque OS réellement pris en charge, dans une session utilisateur déverrouillée, ainsi que la provenance complète du runtime de release (dépôt, empreinte de clé/signature, paquet source `.deb`, pin de version et empreinte). Le sandbox headless ne fournit pas une collection Secret Service utilisable ; il ne permet donc pas de conclure les scénarios création, lecture après redémarrage, absence/révocation, permissions insuffisantes et absence de fuite dans SQLite, `profile.json`, logs ou backups.

Le dossier de gate désormais prêt comprend `SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md`, `RUNTIME_RELEASE_LOCK.json`, `PUBLIC_RELEASE_DECISION.md` et les scripts de capture. Le paquet Chromium QA `151.0.7922.71-1xtradeb1.2404.1` n’est plus disponible dans l’index APT courant ; il doit être récupéré depuis une archive de confiance et vérifié, ou remplacé explicitement par un nouveau candidat qui repasse l’E2E complet.

La migration des métadonnées de profils JSON vers SQLite métier et le branchement du dashboard React vers les routes BACK-01 demeurent également des travaux produit séparés. Ils ne bloquent pas la preuve cryptographique et de restauration BACK-01, mais empêchent de présenter ForgeLocal comme un panneau desktop complet finalisé.

## Fichiers de preuve

| Fichier | Contenu |
|---|---|
| `validation_back01_integration/final/ac_back_01_with_explicit_chromium.out` | E2E API avec binaire, version, PID, profil, `about:blank`, arrêt et locks explicitement journalisés |
| `validation_back01_integration/final/chromium_runtime_release_provenance.out` | OS, paquet, canal APT, mainteneur et empreintes runtime QA |
| `validation_back01_integration/final/minimal_artifact_runtime_gate_build.out` | Tests minimaux et Gosec `Issues : 0` |
| `validation_back01_integration/final/minimal_artifact_runtime_gate_validation.out` | Checksum de l’archive, checksums internes, manifeste et Gitleaks |
| `validation_back01_integration/final/go_test_all_after_runtime_gate.out` | Suite `go test ./...` entièrement verte |
| `release/back01-minimal/RUNTIME_QA_APPROVAL.md` | Périmètre exact de l’approbation runtime QA |
| `release/back01-minimal/SYSTEMVAULT_RUNTIME_RELEASE_GATE.md` | Décision pilote/public, conditions SystemVault et provenance runtime |
| `release/back01-minimal/SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md` | Procédure native sans `sudo`, hors conteneur, incluant révocation, coffre verrouillé et anti-fuite |
| `release/back01-minimal/RUNTIME_RELEASE_LOCK.json` | Verrou machine-readable de version, source et preuves runtime exigées |
| `release/back01-minimal/PUBLIC_RELEASE_DECISION.md` | Décision actuelle `PUBLIC_RELEASE_BLOCKED` et conditions de levée |
| `docs/security/SYSTEM_VAULT_NATIVE_MATRIX.md` | Matrice de validation native à exécuter sur hôte cible |
