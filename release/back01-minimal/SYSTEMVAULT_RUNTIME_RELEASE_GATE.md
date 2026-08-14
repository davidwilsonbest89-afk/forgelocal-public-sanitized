# SystemVault + Runtime Release Gate — ForgeLocal BACK-01

## Décision de gate

> **Statut actuel : pilote local contrôlé autorisé ; publication publique refusée.**

La validation AC-BACK-01 locale est concluante : le profil restauré a été relancé avec Chromium local, sans navigation externe, puis le processus s’est arrêté proprement sans lock de profil résiduel. Cette preuve ne transfère pas l’approbation au runtime Camoufox et n’autorise pas à distribuer Chromium dans l’artefact minimal.

La publication publique exige cumulativement une matrice SystemVault native verte sur l’OS cible et une provenance runtime vérifiée, avec paquet/version/empreinte/signature documentés pour le binaire réellement évalué.

| Gate | Pilote local contrôlé | Publication publique | État actuel |
|---|---:|---:|---|
| AC-BACK-01, restauration isolée | Obligatoire | Obligatoire | **Vert** |
| Relance runtime explicite | Obligatoire | Obligatoire | **Vert, Chromium QA local** |
| Artefact minimal, checksums, Gosec, scan secrets | Obligatoire | Obligatoire | **Vert** |
| SystemVault natif sur OS cible | Non | Obligatoire | **Bloquant** |
| Provenance et signature runtime complète | Non | Obligatoire | **Bloquant** |
| Camoufox validé séparément | Non | Obligatoire si distribué/proposé | **Non évalué** |

## Preuve de relance Chromium

La commande de validation est :

```bash
export GOTOOLCHAIN=go1.25.13
go test ./internal/api -run '^TestBackupV1CreateModifyRestoreIsolation$' -count=1 -v
```

La sortie de référence doit contenir les marqueurs suivants, sans secret :

```text
AC-BACK-01 runtime relaunch started:
  binary=/usr/bin/chromium
  version=Chromium 151.0.7922.71 built on Ubuntu 24.04.4 LTS
  pid=<pid éphémère>
  target_profile_id=target-api
  user_data_dir=<répertoire temporaire de test>
  endpoint=none (--dump-dom direct process)
  navigation=about:blank
AC-BACK-01 runtime relaunch stopped cleanly:
  profile_lock_cleanup=verified
```

Le test lance directement Chromium avec `--dump-dom about:blank`; il n’ouvre donc pas d’endpoint de débogage. La valeur explicite `endpoint=none` est le comportement attendu. Le PID est un identifiant de processus éphémère, présenté uniquement comme preuve de lancement/arrêt dans le log de test.

## Provenance runtime QA observée

| Attribut | Valeur observée |
|---|---|
| OS de validation | Ubuntu 24.04.4 LTS, `amd64` |
| Runtime | Chromium standard Ubuntu, utilisé uniquement pour QA locale |
| Wrapper | `/usr/bin/chromium` |
| Binaire réel | `/usr/lib/chromium/chromium` |
| Version installée | `151.0.7922.71-1xtradeb1.2404.1` |
| Version Chromium | `151.0.7922.71` |
| Canal configuré | `http://ppa.launchpad.net/xtradeb/apps/ubuntu noble main` |
| Mainteneur de paquet | XtraDeb Team `<team@xtradeb.net>` |
| SHA-256 wrapper | `36cbbb620daeb933ae7861de07fcff05b5e1f7527303b5992459b8aa6707b845` |
| SHA-256 binaire réel | `ad69c6632131d3a8b0e7395c3bb36d05cad6a2c650ecfa7ebe2e8dcba955c2de` |
| Runtime incorporé dans l’artefact minimal | Non |

Le candidat APT observé était `151.0.7922.108-1xtradeb1.2404.1`, différent de la version installée et testée. La version QA historique `151.0.7922.71-1xtradeb1.2404.1` n’est plus présente dans l’index APT courant, n’est pas dans le cache local et l’URL directe historique a répondu `404`. Elle ne doit donc jamais être remplacée implicitement : récupérer l’artefact exact depuis une archive de confiance, ou nommer le candidat plus récent et refaire l’E2E complet.

Le verrou `RUNTIME_RELEASE_LOCK.json` fixe ce constat, l’empreinte de l’artefact pilote, les empreintes runtime observées et les preuves encore obligatoires. Une release publique reste refusée tant que ce fichier indique `PUBLIC_RELEASE_BLOCKED`.

### Conditions supplémentaires de provenance avant publication publique

1. Exécuter `scripts/capture-runtime-release-evidence.sh` avec le paquet/version exacts. Le script n’exige pas `sudo`, n’installe pas et n’exécute pas le paquet ; il refuse une version historique indisponible plutôt que de la remplacer par le candidat courant.
2. Enregistrer l’URL de dépôt, l’empreinte de clé de signature et l’état de vérification APT dans le dossier de release.
3. Archiver le nom, la version, l’architecture et le SHA-256 du paquet `.deb` source, pas seulement ceux du wrapper et du binaire décompressé.
4. Vérifier que le SHA-256 du `.deb` est celui de l’entrée de l’index APT authentifié par `InRelease`, puis conserver `InRelease`, le keyring et l’empreinte de clé.
5. Installer explicitement la version validée, ou valider la nouvelle version candidate, avec un pin de version reproductible.
6. Réviser licence, notices, conditions de redistribution et mise à jour de sécurité du runtime choisi.
7. Si Camoufox est proposé, appliquer la même procédure à ses artefacts, sa licence MPL-2.0, son dépôt officiel, ses signatures et son E2E ; aucune équivalence avec Chromium n’est présumée.

## Matrice SystemVault native obligatoire

L’exécution doit se faire dans la session utilisateur réelle qui exécutera ForgeLocal, sans `sudo`, sans conteneur et avec le coffre natif déverrouillé. Aucune valeur de test ne doit figurer dans une ligne de commande, un historique shell, une sortie CI ou un rapport.

```bash
export GOTOOLCHAIN=go1.25.13
export FORGELOCAL_VAULT_SERVICE="ForgeLocal.Back01.ReleaseCandidate"
go run ./cmd/systemvault-doctor > systemvault-matrix.json
```

| Cas | Contrôle requis | Preuve assainie attendue |
|---|---|---|
| Backend natif | Identifier Secret Service, Keychain ou Credential Manager | OS, backend, service et compte ; jamais la valeur |
| Création et lecture | `systemvault-doctor` | création/lecture réussies, `key_id` non sensible |
| Redémarrage Core | recréer `SystemVault`, lancer à nouveau Core puis backup/restore | même `key_id`, lecture réussie après redémarrage |
| Secret proxy | `PutSecret`/`GetSecret`, lancement local contrôlé | référence non escaladable ; valeur absente des logs |
| Clé absente/révoquée | suppression via backend puis backup/restore | échec explicite, pas de fallback en clair |
| Coffre verrouillé / permissions | session sans accès au coffre | refus explicite, aucun dump de secret |
| Absence de fuite | recherche binaire contrôlée après scénario | valeur de contrôle absente de SQLite, profils, logs et `.flbackup` |

Pour le contrôle anti-fuite, fournir le secret de test uniquement par un mécanisme de secret CI ou un environnement éphémère non journalisé, puis vérifier :

```bash
! grep -R --binary-files=text --fixed-strings "$FORGELOCAL_TEST_SECRET" \
  "$FORGELOCAL_DATA_DIR/metadata.db" \
  "$FORGELOCAL_DATA_DIR/profiles" \
  "$FORGELOCAL_DATA_DIR/backups" \
  "$FORGELOCAL_DATA_DIR/logs"
```

## Sortie de gate attendue

Le rapport de release doit se conclure par une seule de ces décisions :

| Décision | Conditions |
|---|---|
| `PILOT_LOCAL_APPROVED` | AC-BACK-01, artefact minimal et preuve runtime locale verts ; SystemVault public non requis pour le pilote isolé |
| `PUBLIC_RELEASE_APPROVED` | Toutes les lignes SystemVault sont vertes sur l’OS cible **et** la provenance/signature/version runtime est complète et reproductible |
| `PUBLIC_RELEASE_BLOCKED` | Toute ligne SystemVault, provenance, signature ou nouvelle version runtime non validée |

Le sandbox headless reste classé `PUBLIC_RELEASE_BLOCKED` : il n’expose pas de collection Secret Service durable et déverrouillée. Cette indisponibilité ne doit jamais déclencher un fallback de secrets en clair.
