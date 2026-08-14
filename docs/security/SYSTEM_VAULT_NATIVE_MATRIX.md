# Matrice native SystemVault — BACK-01

## Principe

Le coffre ne doit jamais être remplacé par un fichier local de secours lors d’un test de distribution. La validation doit utiliser le backend natif de l’OS cible : Secret Service/Keychain/Credential Manager.

Le binaire `cmd/systemvault-doctor` ne journalise ni clés, ni tokens, ni valeurs proxy. Il vérifie automatiquement la création, la lecture, la recréation du client (équivalent redémarrage du Core), la suppression et l’absence vérifiée.

## Exécution Linux native

```bash
export GOTOOLCHAIN=go1.25.13
export FORGELOCAL_VAULT_SERVICE="ForgeLocal.Back01.ReleaseCandidate"
go run ./cmd/systemvault-doctor > systemvault-matrix.json
```

Le test doit être exécuté **dans la session graphique ou utilisateur réellement utilisée pour ForgeLocal**, avec une collection Secret Service déverrouillée. Il ne doit pas être exécuté avec `sudo`, dans un conteneur, ni avec un backend mémoire.

## Matrice d’acceptation

| Cas | Méthode | Attendu | Statut sandbox |
|---|---|---|---|
| Création clé AES | `systemvault-doctor` | `created_key=true` | Non concluant : coffre headless sans collection déverrouillée |
| Lecture pendant lancement | `systemvault-doctor` puis API Core | clé lisible, jamais renvoyée par API | À exécuter sur hôte cible |
| Redémarrage Core | recréer `SystemVault`, redémarrer le Core | même `key_id`, valeur récupérable | À exécuter sur hôte cible |
| Secret proxy | `PutSecret`/`GetSecret` et lancement local | secret utilisable uniquement en mémoire | À exécuter sur hôte cible |
| Clé absente | supprimer l’item, lancer backup/restore | échec contrôlé, aucun fallback en clair | Couverte par doctor après suppression |
| Clé révoquée | supprimer/révoquer via UI OS, redémarrer Core | `ErrNotFound`, audit sans valeur | Manuel, hôte cible |
| Permissions insuffisantes | session OS sans accès au coffre | refus explicite, logs sans secret | Manuel, hôte cible |
| Absence de fuite | scanner `metadata.db`, `profile.json`, logs, `.flbackup` | aucune valeur de test trouvée | API/tests et scan release requis |

## Exécution de la vérification anti-fuite

Après une sauvegarde intégrée contenant le secret de contrôle connu uniquement du test :

```bash
! grep -R --binary-files=text --fixed-strings "$FORGELOCAL_TEST_SECRET" \
  "$FORGELOCAL_DATA_DIR/metadata.db" \
  "$FORGELOCAL_DATA_DIR/profiles" \
  "$FORGELOCAL_DATA_DIR/backups" \
  "$FORGELOCAL_DATA_DIR/logs"
```

Le secret de contrôle ne doit pas être saisi dans une ligne de commande, journal shell, capture CI ou rapport. Il doit être fourni à la session de test par un mécanisme de secret CI.

## Décision sandbox

Le sandbox possède les binaires `secret-tool` et `gnome-keyring`, mais ne possède pas une collection utilisateur durable et déverrouillée. Les tentatives de session D-Bus éphémère se bloquent lors de la création de collection ou échouent sans backend utilisable. Ceci confirme que la matrice ne peut pas être validée honnêtement dans ce contexte headless.

> Une matrice SystemVault n’est marquée **verte** que lorsque son JSON provient de l’OS cible, dans la session utilisateur qui exécutera réellement ForgeLocal. L’échec ou l’indisponibilité du coffre est un blocage de publication, jamais une autorisation de fallback en clair.
