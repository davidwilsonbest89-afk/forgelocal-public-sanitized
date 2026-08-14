# Runbook — Gate SystemVault natif sur hôte cible

**Objectif.** Exécuter le dernier gate de coffre système de ForgeLocal sur un hôte réellement ciblé par la release. Ce runbook ne valide pas le sandbox, un conteneur, une session `sudo`, ni un simple mock mémoire.

> **Règle de décision :** le résultat reste `PUBLIC_RELEASE_BLOCKED` tant qu’un cas obligatoire est absent, échoue, révèle une valeur sentinelle ou utilise un fallback en clair.

## Préconditions

L’opérateur exécute cette procédure depuis le compte utilisateur desktop qui utilisera ForgeLocal. Le coffre natif doit être déverrouillé avant le lancement. Sous Linux, le backend attendu est **Secret Service**. Aucune valeur de secret, aucun token, aucun nom personnel de compte et aucune commande `sudo` ne doivent apparaître dans les preuves.

| Contrôle | Condition attendue |
|---|---|
| Compte | Identifiant de classe uniquement : `non-root-desktop-session` |
| Privilèges | `EUID != 0`, aucune commande `sudo` |
| Environnement | Hôte cible, hors conteneur, D-Bus de session présent |
| Coffre | Déverrouillé au départ ; backend natif accessible |
| Données | Répertoire ForgeLocal isolé, non productif |
| Secrets de test | Sentinelle non productive dans un fichier `0600` hors répertoire de données |

## 1. Matrice automatisée native

Depuis la racine du dépôt, lancer :

```bash
chmod 0755 scripts/run-systemvault-native-gate.sh scripts/check-systemvault-anti-leak.sh
OUT_DIR="$PWD/systemvault-native-evidence" \
FORGELOCAL_VAULT_SERVICE="ForgeLocal.Back01.ReleaseCandidate" \
scripts/run-systemvault-native-gate.sh
```

La sortie `systemvault-matrix.json` est assainie : elle ne contient que des booléens et l’identifiant de service. Elle doit indiquer `true` pour `created_key`, `read_key`, `restart_read`, `created_secret`, `read_secret`, `deleted` et `absent_verified`.

| Cas automatisé | Preuve attendue |
|---|---|
| Création de clé | `created_key: true` |
| Lecture immédiate | `read_key: true` |
| Lecture après nouvelle instance Core | `restart_read: true` |
| Secret proxy séparé | `created_secret: true`, `read_secret: true` |
| Suppression | `deleted: true` |
| Clé et secret absents après suppression | `absent_verified: true` |

## 2. Révocation externe et clé absente

Créer de nouveau un élément via le diagnostic, puis supprimer **manuellement dans le gestionnaire natif du coffre** l’élément de clé et l’élément de secret du service de test. Relancer une lecture avec le Core ou le diagnostic instrumenté.

La preuve acceptable montre uniquement l’opération et le résultat : `ErrNotFound` ou refus explicite. Une restauration ou une sauvegarde dépendant de cette clé doit échouer de façon contrôlée. Toute création implicite de clé, toute valeur dans SQLite ou tout fallback en clair invalide le gate.

## 3. Coffre verrouillé ou permissions insuffisantes

Verrouiller le coffre dans la session desktop ou utiliser une session de test sans accès au coffre. Exécuter une opération de lecture, de création de backup et de restauration. La sortie doit indiquer un refus contrôlé, sans valeur secrète. Répéter ensuite depuis la session normalement autorisée pour vérifier que le fonctionnement revient à la normale une fois le coffre déverrouillé.

## 4. Contrôle anti-fuite intégré

Créer une sentinelle non productive dans un fichier de permissions strictes, **en dehors** du répertoire ForgeLocal :

```bash
umask 077
sentinel_file="$(mktemp)"
openssl rand -hex 32 > "$sentinel_file"
chmod 0600 "$sentinel_file"
```

Injecter cette sentinelle exclusivement via le coffre natif pendant un scénario intégré de profil, backup et restauration. Ne jamais l’exporter en variable ni l’ajouter aux arguments de processus. Une fois les opérations terminées :

```bash
FORGELOCAL_DATA_DIR="$PWD/forge-local-test-data" \
FORGELOCAL_TEST_SENTINEL_FILE="$sentinel_file" \
OUT_FILE="$PWD/systemvault-native-evidence/systemvault-anti-leak.json" \
scripts/check-systemvault-anti-leak.sh
rm -f "$sentinel_file"
```

Le résultat accepté est `{"anti_leak":true,...}`. Les fichiers inspectés incluent `metadata.db`, les profils, les logs et les backups `.flbackup` sous le répertoire de données. En cas de détection, ne pas publier la correspondance ni la valeur ; ouvrir un incident et supprimer la donnée de test.

## 5. Dossier de preuve et décision

Le dossier de gate doit contenir les fichiers suivants, plus aucune donnée personnelle ou valeur secrète :

| Fichier | Rôle |
|---|---|
| `systemvault-host-context.env` | OS, architecture, backend attendu, classe de session et exécution sans `sudo` |
| `systemvault-matrix.json` | Création, lecture, redémarrage, suppression et absence vérifiée |
| `SYSTEMVAULT_NATIVE_GATE_STATUS` | Résultat automatisé et cas manuels restants |
| `systemvault-anti-leak.json` | Résultat anti-fuite booléen |
| `runtime-release-evidence/` | `.deb`, hashes, `InRelease`, keyring, empreinte et contrôle APT, lorsque disponible |
| `PUBLIC_RELEASE_DECISION.md` | Une seule décision finale et ses références |

La conclusion de `PUBLIC_RELEASE_DECISION.md` est obligatoirement l’une des valeurs ci-dessous.

| Décision | Usage |
|---|---|
| `PUBLIC_RELEASE_APPROVED` | Tous les cas SystemVault sont verts et la provenance runtime est complète, vérifiée et rejouable |
| `PUBLIC_RELEASE_BLOCKED` | Un cas échoue, manque, fuit une valeur ou le `.deb`/la signature exacts ne sont pas archivés |

## Limites connues du candidat pilote

Au 14 août 2026, le runtime QA Chromium local est `151.0.7922.71-1xtradeb1.2404.1`, mais le paquet exact n’est plus exposé par l’index APT courant. Il reste donc verrouillé pour le pilote, sans substitution. Le candidat APT plus récent doit être traité comme une nouvelle validation, jamais comme un remplacement automatique.
