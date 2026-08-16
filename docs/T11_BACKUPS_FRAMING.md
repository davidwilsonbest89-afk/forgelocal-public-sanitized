# Cadrage T11 — Backups et restauration (document de préparation non normatif)

**Référence :** CDC ForgeLocal v3.9.7, sections 3.4 (import/export/backup/restauration) et BACK-01 · cadrage T10 `docs/T10_PROXIES_FRAMING.md` (T10-FRAMING-20260816-001)
**Date :** 16 août 2026
**Nature :** document de préparation non normatif. Il ne crée aucun nouveau statut, aucune nouvelle exigence, aucun nouveau certificat. Il borne uniquement le périmètre prévu de T11 pour un démarrage immédiat dès autorisation explicite, conformément à la cadence validée pour T08, T09 et T10.
**Identifiant :** T11-FRAMING-20260816-001

## Contexte et fondations déjà livrées

Les jalons précédents ont établi les fondations que T11 réutilisera sans modification : le contrat d'écritures T09 (création, modification, archivage, tags, validation serveur, `correlation_id`, audit redacted, mutations loopback uniquement, erreurs machine-readable), le référentiel proxy T10 (affectation profil↔proxy, `secret_ref` par référence seule, jamais de credential en clair), et la fiabilité de concurrence T08 (queue bornée, sérialisation par profil, timeout, annulation, cleanup borné, `RecoveryRecords`). Le schéma BACK-01 existe déjà en partie dans `metadata.db` (`backup_operations`, `backups`, `restore_operations`, `audit_events` selon le schéma canonique du CDC).

## Périmètre prévu de T11 (à confirmer par instruction explicite avant démarrage)

Le lot T11 couvrirait le contrat réel de **backup et restauration** via le Core Go unique, dans les limites suivantes :

| Axe | Contenu prévu | Source CDC |
|---|---|---|
| Backup `.flbackup` | Conteneur unique `FLBK … FLEND`, AES-256-GCM, AAD, nonce `crypto/rand`, publication atomique, `fsync`, permissions restrictives | CDC v3.9.7 §3.4, BACK-01 |
| Chiffrement | Clef dérivée localement ; fallback chiffré documenté, jamais en clair ; SQLite ne contient que des références opaques | CDC v3.9.7 §3.3 |
| Restauration isolée | Restauration vers un **nouvel identifiant** (jamais d'écrasement sans action explicite et audité), réconciliation au démarrage, quarantaine des artefacts malformés | CDC v3.9.7 §3.4, BACK-01 |
| Validation | Vérification complète du format, de l'authenticité et des chemins avant restauration ; archives traitées comme non fiables | CDC v3.9.7 §3.4 |
| Journalisation | Transitions durables `started → validated/committed/failed` ; crash marqué `INTERRUPTED_BEFORE_COMPLETION` sans reprise implicite | CDC v3.9.7 §3.4 |
| Dashboard | UI backup/restauration **client du Core** (mémoire seule, aucun accès direct SQLite), écran de gestion des backups, journal d'audit | CDC v3.9.7 §5 |
| Garde-fous zéro-trust | Loopback requis (403 hors boucle), token admin distinct, `correlation_id`, audit redacted, erreurs machine-readable | T09/T10 |

## Dépendances et conditions d'ouverture

La dépendance critique reste la qualification du coffre `internal/secrets` (chiffrement/verrouillage de type SystemVault) : tant qu'elle n'est pas démontrée, la clef de backup doit rester un fallback local chiffré documenté conforme BACK-01, jamais en clair, et aucune preuve ne doit utiliser de secret réel. La dépendance structurelle est la migration `0004_profile_import_operations_started.sql` (journal d'opération durable) et, le cas échéant, une migration `0007_t11_backups.sql` pour compléter le schéma BACK-01 — toute migration sera soumise à la même discipline de preuves que `0005_t09_profile_writes.sql`.

## Critères d'acceptation prévus (indicatifs, non normatifs)

Backup d'un profil valide ; refus des entrées invalides ; restauration isolée vers nouvel identifiant ; restauration d'une archive malformée refusée sans contamination (quarantaine) ; restauration d'une archive corrompue/altérée refusée (intégrité AES-GCM) ; journalisation durable des opérations avec crash recovery ; zéro secret en clair dans l'archive, les réponses, les logs ou SQLite ; concurrence sur le même profil ; comportement après erreur ; tests négatifs, concurrence (`go test -race`), E2E Playwright, Gitleaks/Gosec sur le delta, `git diff --check`, chemins RC inchangés.

## Interdictions confirmées

Aucun coffre système natif activé avant sa qualification ; aucun secret réel dans les preuves (valeurs synthétiques uniquement) ; aucun lancement runtime, navigateur ou Camoufox ; aucun proxy réseau réel ; aucun import de masse hors contrat transactionnel T07/qualifié ; aucune modification du RC ; aucune release.

## Livraison attendue

Comme pour T08, T09 et T10 : un rapport unique final au format 16 champs (TASK → NEXT ALLOWED STEP) à la fin du jalon, une archive de preuves redacted et portable avec manifeste SHA256SUMS vérifié, le hash du ZIP annoncé dans le rapport et calculé une seule fois après constitution finale, le rapport livré séparément de l'archive. Aucune validation intermédiaire n'est demandée pendant T11 ; T12 ne doit pas commencer dans le même lot.

## Statuts inchangés

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN` (findings `validation_back01_integration/` préexistants, règle `generic-api-key`), pilote suspendu, cinq gates publics en attente. Le démarrage de T11 nécessite une autorisation explicite du valideur.
