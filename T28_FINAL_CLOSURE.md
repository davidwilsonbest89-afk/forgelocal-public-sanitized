# T28-FINAL-CLOSURE-PASS

## Décision locale

> `T28_APPROVED_VERIFIABLE_LOCAL`

T28 est clôturé au périmètre **Core local contrôlé** et vérifiable localement. Cette décision ne vaut pas approbation d’un runtime navigateur, d’une extension chargée ou exécutée, de Camoufox/Chromium, de SystemVault natif, de proxy/cookies réels, de migration, de production ou de release publique.

La décision est soumise à l’acceptation finale du propriétaire. Elle ne démarre aucun lot suivant.

## HEAD et lignée

La branche GitHub est [`feature/t28-local-extensions-controlled`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/feature/t28-local-extensions-controlled). Le HEAD distant contrôlé avant la présente correction était `15f3c4a125e5a587fb99cd0851b10369964db30b`. La correction de clôture est préparée depuis ce HEAD ; son commit de publication sera ajouté au même historique et ne doit contenir que le correctif d’intégrité, les tests ciblés et les preuves/documentations associées.

La baseline V6 reste `999374d99b7996504ba91e421850a2fe84afb78d`, l’implémentation T28 initiale `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578`, et le package R1 reste inchangé : ZIP `7be89af8e093b9b35c7924bc0ac3d8f0268cba9a2682db0d558f9b8573d9158d`, bundle `442c1ff49b4f62b41edbe7cc2ee686fe8eaf15553b56e30ac39b4930e6a3944f`. Aucun artefact R1 cohérent n’a été recréé ou supprimé.

## Défaut concret corrigé

La passe ciblée a ajouté un test reproductible qui modifiait le blob ZIP géré après import. Avant correction, `Approve` ne revérifiait pas le digest stocké et pouvait accepter ce package altéré. La correction ajoute `verifyBlobIntegrity`, compare la taille et le SHA-256 avant `Approve`, `Assign` et `Rollback`, et mappe l’échec API vers `INTEGRITY_MISMATCH`. Ce changement est le seul changement métier de la passe finale et est couvert par le test `TestT28RejectsPackageModifiedAfterImport`.

## Couverture fonctionnelle finale

| Exigence | Preuve |
|---|---|
| ZIP corrompu, manifest absent/invalide, JSON concaténé | `TestT28RejectsUnsafeArchivesAndCompensatesDatabaseFailure`, `TestT28RejectsCorruptAndOversizedArchives` |
| Traversal, symlink et limites | tests repository T28, refus avant mutation |
| Toutes permissions sensibles et host patterns | `TestT28PreservesAllAuthorizedPermissionsAndIgnoresUpdateURL` ; conservation et `HIGH_RISK` |
| `update_url`/`updateURL` | ignoré, non suivi, non téléchargé et non exécuté |
| Acknowledgement exact et high-risk | `TestT28ImportPreservesPermissionsAndRequiresExactHighRiskAcknowledgement`, tests API |
| Affectation | refus avant approval, profil absent, révocation/quarantaine ; concurrence déterministe |
| Immutabilité/update/rollback | `TestT28LifecycleAssignmentUpdateRollbackRevokeAndPersistence` |
| Purge | `TestT28PurgeRequiresSafeLifecycleState` ; purge uniquement après état sûr et explicite |
| Authentification et surface locale | bearer obligatoire, loopback-only et origin guard |
| Redaction | projections/audit sans package, token, cookie, chemin complet ou payload sensible |
| Non-exécution | aucun navigateur, runtime, proxy, téléchargement ou processus externe |

## Tests et scans

La commande ciblée exacte `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'` a retourné `0` après correction. `go vet ./internal/extensions ./internal/api`, `go build ./...` et `git diff --check` ont aussi retourné `0`. Les tests globaux baseline/HEAD R1 restent les preuves historiques fournies : PASS/PASS, sans reproduction du finding runtime historique.

OSV v1.9.2 a réellement retourné code `1` avec 46 identifiants sur chaque `go.mod`; ces avis ne sont pas masqués et restent une réserve de dépendances hors portée de cette correction. Gosec ciblé après correction retourne 6 findings historiques dans `internal/api`; le package `internal/extensions` n’ajoute aucun finding nouveau, et les suppressions sont locales et justifiées. Gitleaks R1 sur les huit commits, l’extraction ZIP et le package final n’a trouvé aucun leak ; un contrôle Gitleaks de la clôture sera également conservé après publication du correctif.

## Artefacts R1 réutilisés

Le ZIP et le bundle R1 ont été revérifiés physiquement depuis un clone public neuf. Le ZIP a passé SHA-256, sidecars distribué/neutre, `unzip -t`, extraction, manifeste non auto-référentiel, checksums internes et audit de contenu. Le bundle a passé SHA-256, sidecars, `git bundle verify`, prerequisite baseline, clone seedé, checkout de référence, `fsck` et worktree propre. Le clone bundle seul en code `128` est le comportement attendu d’un delta qui exige la baseline.

## Gates et arrêt

Les valeurs restent strictement inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false`. Aucun navigateur, Chromium, Camoufox, runtime d’extension, proxy réel, cookie réel, SystemVault natif, téléchargement externe, migration utilisateur, environnement de production ou release n’a été lancé.

T29, T39, T40, T41 et T42 ne commencent pas avant l’acceptation finale du propriétaire sur T28.

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/feature/t28-local-extensions-controlled "Branche T28"
[2]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/feature/t28-local-extensions-controlled/evidence/T28/t28-evidence-qualification-r1.zip "ZIP R1"
[3]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/blob/feature/t28-local-extensions-controlled/evidence/T28/t28-evidence-qualification-r1.delta.bundle "Bundle R1"
