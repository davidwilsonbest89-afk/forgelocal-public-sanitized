# Revue indépendante T10–T15 E2E synthétique

**Décision exacte :** `T10_T15_SYNTHETIC_E2E_APPROVED_VERIFIABLE_LOCAL`

Cette décision signifie que la chaîne de preuve locale publiée est physiquement vérifiable. Elle ne constitue ni une approbation produit générale, ni une levée de gate, ni une release.

## Contrôles physiques

| Contrôle | Résultat |
|---|---|
| Branche distante | `audit/t00-t42-t10-t15-e2e-validation-v6` résolue au commit `8c38830b155f293c8b05b2789e60f7de4c45f565` |
| Clone neuf | clone sparse neuf ; HEAD égal au commit attendu |
| Checkout exact | checkout détaché du commit annoncé = 0 |
| Git | `git fsck --full` avant/après = 0 ; worktree propre |
| ZIP | sidecar portable contrôlé depuis un répertoire neutre = 0 |
| Archive | `unzip -t` = 0 ; extraction fraîche = 0 |
| Manifeste | non auto-référentiel = 0 |
| Checksums internes | `sha256sum -c SHA256SUMS` = 0 |
| Bundle | `git bundle verify` = 0 ; sidecar bundle neutre = 0 |
| Gitleaks kit extrait | = 0, aucun leak |
| Gitleaks plage Git | = 0 sur `999374d99b7996504ba91e421850a2fe84afb78d..8c38830b155f293c8b05b2789e60f7de4c45f565` ; une plage non vide a été vérifiée |
| Logs E2E | UTC/CWD/commande/versions/HEAD/exit codes archivés ; `Running 7 tests using 1 worker`, `7 passed (15.6s)` |

## Assertions auditées

Les preuves et les sources de test couvrent la création T10 valide, le refus du port invalide, le refus du profil inexistant avec correlation ID, le refus des écritures hors loopback, le listing redacted sans valeur de credential, la navigation locale fail-closed, la projection redacted/digest sans credential, la fermeture de session T15 et le montage du panneau d’automatisation Dashboard.

La preuve de cleanup contient `token_file_removed=yes`, `base_dir_removed=yes`, `run_root_removed=yes`, `port_19280_after_cleanup=closed` et `port_3000_after_cleanup=closed`. Aucun token temporaire n’est présent dans le kit extrait. Camoufox, proxy réel, cookie réel, donnée utilisateur, SystemVault natif, migration, runtime de production et release n’ont pas été utilisés.

## Artefacts

| Artefact | SHA-256 |
|---|---|
| ZIP `t10-t15-e2e-validation-v6.zip` | `93c0224f01fbde6be0d1e5cb9a11c9a20362647116023a0d2e2ff244c1260081` |
| Bundle `t10-t15-e2e-validation-v6.delta.bundle` | `28872679155097942f620e31144f84f83f98bb93b7a21871d04b12a16b51611a` |

**Vérification publique finale :** la branche de revue `audit/t00-t42-t10-t15-independent-review-v6` a été contrôlée depuis un clone neuf sparse au commit `8b2b5f567b4bf19ca1172c3331b84a697e7059c3` ; clone, fsck, checkout exact, fetch LFS ciblé, sidecars en répertoire neutre, `unzip -t`, extraction, checksums internes, manifeste, bundle, worktree propre, plage Git non vide et Gitleaks ont tous retourné zéro.

**Owner suivant :** revue indépendante de gouvernance qualité. **Gate produit :** aucune gate V6 n’est levée par cette décision.
