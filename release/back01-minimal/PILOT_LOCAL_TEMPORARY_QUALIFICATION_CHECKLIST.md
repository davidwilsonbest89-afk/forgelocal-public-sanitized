# Checklist opérationnelle — pilote local temporaire et qualification Ubuntu native

**Statut de ce document :** procédure de test contrôlé. Il ne constitue ni une publication, ni une promesse de compatibilité, ni une levée de gate.

> **Autorisation applicable :** `PILOT_LOCAL_APPROVED — TEMPORARY / SYSTEMVAULT_PENDING`.
>
> Le statut public demeure sans exception **`PUBLIC_RELEASE_BLOCKED`**.

## 1. Limites du pilote autorisé

| Point | Exigence |
|---|---|
| Cible unique | Ubuntu **24.04.4 LTS** `amd64`. |
| Participants | Personnel explicitement contrôlé par le mainteneur du projet. |
| Environnement | Local, isolé, non productif, avec accès restreint. |
| Usage | Validation technique limitée et préparation de la qualification native SystemVault. |
| Interdits | Publication, distribution publique, accès utilisateur externe, fallback secret en clair, ajout de Camoufox, annonce Windows/macOS. |
| État public | `PUBLIC_RELEASE_BLOCKED` jusqu’à cinq gates `PASSED` frais et revus indépendamment. |

L’autorisation est liée à la chaîne suivante et cesse automatiquement à la moindre divergence :

```text
artifact_sha256=553095461c94a44fd4f4d8c4040590134ca344b3d1a86cb1a5e9d400245b16d6
source_commit=67a8dfd897e540a55fc10749e1f2ef85b8356a8b
runtime=Chromium 151.0.7922.108-1xtradeb1.2404.1 amd64
```

## 2. Contrôle pré-pilote obligatoire

Avant chaque session, le responsable compare l’archive, le commit, le runtime et les métadonnées au fichier `PILOT_LOCAL_TEMPORARY_AUTHORIZATION.json`. Toute différence signifie **suspension immédiate** ; il ne faut pas essayer de modifier les preuves pour les adapter.

| Contrôle | Résultat à exiger | Si échec |
|---|---|---|
| Archive | SHA-256 identique à la valeur autorisée | Suspendre. |
| Runtime | Chromium et version identiques | Suspendre et requalifier. |
| Source référencée | Commit identique | Suspendre. |
| Accès | Pas d’utilisateur externe, pas de publication | Suspendre l’accès. |
| Secrets | Coffre uniquement, aucun fallback clair | Ouvrir un incident et suspendre. |
| Isolation | Répertoires et locks distincts | Suspendre et conserver les faits assainis. |

## 3. Préconditions du test Ubuntu réel

Le test doit être exécuté depuis une **session graphique de l’utilisateur desktop** qui emploiera ForgeLocal. Il ne peut pas être exécuté depuis root, `sudo`, un conteneur, un terminal non graphique, une session distante sans D-Bus utilisateur ou un mock en mémoire.

Exécuter ce préflight non sensible dans le terminal de la session graphique :

```bash
printf 'euid=%s\n' "$EUID"
printf 'dbus=%s\n' "${DBUS_SESSION_BUS_ADDRESS:+PRESENT}"
printf 'xdg_runtime=%s\n' "${XDG_RUNTIME_DIR:+PRESENT}"
printf 'arch=%s\n' "$(uname -m)"
printf 'os='
. /etc/os-release && printf '%s %s\n' "$ID" "$VERSION_ID"
if [[ -f /.dockerenv ]] || grep -Eq '/(docker|containerd|kubepods|lxc)/' /proc/1/cgroup 2>/dev/null; then
  printf 'container=DETECTED\n'
else
  printf 'container=not_detected\n'
fi
command -v secret-tool >/dev/null && printf 'secret_tool=PRESENT\n' || printf 'secret_tool=ABSENT\n'
```

Le préflight doit indiquer : utilisateur non root, `dbus=PRESENT`, `xdg_runtime=PRESENT`, `arch=x86_64`, `os=ubuntu 24.04`, `container=not_detected` et `secret_tool=PRESENT`. Le coffre Secret Service doit être déverrouillé avant le lancement. Toute condition manquante maintient le gate `SYSTEMVAULT_NATIVE_PER_TARGET` à `PENDING`.

## 4. Matrice SystemVault obligatoire

Depuis un clone de travail non productif, sans `sudo` :

```bash
cd /chemin/vers/ForgeLocal
chmod 0755 scripts/run-systemvault-native-gate.sh scripts/check-systemvault-anti-leak.sh
OUT_DIR="$PWD/systemvault-native-evidence" \
FORGELOCAL_VAULT_SERVICE="ForgeLocal.Back01.ReleaseCandidate" \
scripts/run-systemvault-native-gate.sh
```

| Cas | Exigence de réussite |
|---|---|
| Création de clé | `created_key: true` |
| Lecture immédiate | `read_key: true` |
| Lecture après redémarrage du Core | `restart_read: true` |
| Secret proxy séparé | `created_secret: true` et `read_secret: true` |
| Suppression | `deleted: true` |
| Absence après suppression | `absent_verified: true` |

## 5. Cas manuels bloquants

Après la matrice automatisée, exécuter et consigner de façon assainie les cas suivants :

| Cas | Comportement obligatoire | Conséquence si échec |
|---|---|---|
| Clé absente | `ErrNotFound` ou refus explicite ; aucune recréation implicite | Gate `PENDING` ou `FAILED`. |
| Révocation externe | Suppression dans le gestionnaire de coffre, puis échec contrôlé de lecture/backup/restore | Gate `PENDING` ou `FAILED`. |
| Coffre verrouillé | Refus contrôlé sans secret ni fallback | Gate `PENDING` ou `FAILED`. |
| Permissions insuffisantes | Refus contrôlé, puis retour au fonctionnement normal après rétablissement | Gate `PENDING` ou `FAILED`. |
| Flux intégré | Profil → backup chiffré → modification → restauration isolée | Gate anti-fuite `PENDING` ou `FAILED`. |
| Anti-fuite | `anti_leak: true`, sans sentinelle dans SQLite, profils, browser-data, logs ou `.flbackup` | Incident sécurité et suspension immédiate. |

La sentinelle de test est non productive, créée dans un fichier externe avec permissions `0600`. Elle ne doit jamais être affichée, ajoutée aux arguments de processus, exportée comme variable ou ajoutée à un historique shell.

## 6. Preuves autorisées et revue

Seuls les fichiers assainis suivants peuvent être ajoutés au dossier de preuve :

| Fichier | Rôle |
|---|---|
| `systemvault-host-context.env` | OS, architecture, backend attendu et classe de session. |
| `systemvault-matrix.json` | Résultats booléens de la matrice. |
| `SYSTEMVAULT_NATIVE_GATE_STATUS` | Statut automatisé et cas manuels restants. |
| `systemvault-anti-leak.json` | Résultat anti-fuite booléen. |

Chaque fichier est haché et associé au même artefact, au même runtime, au même commit et à la même cible Ubuntu 24.04.4 amd64. Le responsable indépendant doit vérifier ces liens avant de rendre une décision de gate. Aucun journal contenant une valeur secrète, une sentinelle, un token, une clé privée, un identifiant personnel ou un historique shell ne doit être versé.

## 7. Suspension immédiate et réponse

La permission de pilote est automatiquement suspendue si l’un des événements ci-dessous survient : changement du commit, de l’archive, du runtime, du SBOM, du manifeste ou d’une configuration non tracée ; fuite ou fallback clair d’un secret ; échec de backup, restauration ou isolation ; erreur du coffre ; incohérence d’empreinte.

À la suspension, arrêter le pilote, supprimer les accès externes éventuels, préserver uniquement des constats assainis, ne pas recopier de secret dans un ticket, et conserver `PUBLIC_RELEASE_BLOCKED`. Une reprise exige une nouvelle décision tracée et, lorsque la chaîne candidate a changé, des preuves E2E et SystemVault fraîches.

## 8. Règle de sortie

La qualification Ubuntu ne rend **pas** automatiquement la release publique. `PUBLIC_RELEASE_APPROVED` ne peut être examiné que lorsque les cinq gates de `PUBLIC_RELEASE_GATE_STATE.json` sont explicitement `PASSED`, possèdent des preuves versionnées fraîches et une revue indépendante. Tant qu’un seul gate est `PENDING` ou `FAILED`, le statut public reste `PUBLIC_RELEASE_BLOCKED`.

## Références internes

- `PILOT_LOCAL_TEMPORARY_AUTHORIZATION.json`
- `PUBLIC_RELEASE_GATE_STATE.json`
- `RELEASE_TRACEABILITY_INDEX.json`
- `SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md`
- `SANDBOX_SYSTEMVAULT_LIMITATION_REPORT_2026-08-14.md`
