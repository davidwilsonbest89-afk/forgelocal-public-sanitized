# GOSEC-R3 — vérification overnight et hardening ciblé

## Objet et limites

Cette mission a suivi la correction de HEAD demandée pour la Phase 0. Elle a vérifié la conservation du package overnight v4 depuis un clone neuf, puis a traité trois lots fermés de findings Gosec ouverts. Aucun compte réel, cookie, proxy externe, secret, donnée utilisateur, runtime de production, Camoufox natif, SystemVault natif, Docker/Buildx ou release n’a été utilisé.

Les lots T28 et T29 n’ont pas été rouverts ou démarrés. Les lots T31–T38 n’ont pas été modifiés. Le statut de production reste explicitement `FORGELOCAL_PRODUCTION_READY=false`.

## Phase 0 — lineage et package overnight

Le clone public neuf a été effectué avec `GIT_LFS_SKIP_SMUDGE=1` au clone et au checkout explicite. Le HEAD distant vérifié avant R3 était `e9ad9d41afd2c80505d312efcac9f2c4e961abc7`. La chaîne vérifiée par merge-base est :

```text
20e5181554c52a92d6e1acad15feb426e8804621
→ 31385c26bc6ca8944d68b50be71fec8c0783d590
→ 701c5949261de261d2044cbff3e125b88c56f1a2
→ 209148020e8254d7af903bc09485dbaed4014948
→ aab0cca666e470ed312aabcdb5e369157ba4f204
→ 18367a6d651657e8e48afc006ad75bfa95aa46ea
→ 7f3f5bef4e8dba800c7548532e12fea09a876b46
→ f796299ef12c7c701937d53afd6e020088d95c3c
→ ee45a147919e8d0e9c39e2f276f0e590e770a77e
→ 90b292ee615fe0f4ae78418121375ccb577a442e
→ 9bbc2beed7f2f8f3aa79f2447bbb348a10eb8225
→ e9ad9d41afd2c80505d312efcac9f2c4e961abc7
```

`f796299` est bien le dernier commit source de l’overnight initial, `90b292e` porte le package v4 canonique et `e9ad9d4` porte la vérification publique finale. Le manifeste v4 déclare donc honnêtement `source_head=90b292e`; le log public ajouté ensuite est rattaché séparément au HEAD final. Les sidecars, l’extraction ZIP/TAR, les hashes internes, `unzip -t`, `git bundle verify` et `git fsck --full` ont réussi. Les défauts de packaging et de CWD historiques sont conservés dans les logs bruts et expliqués par les compagnons corrigés.

La toolchain réelle du clone neuf était `go1.25.13` avec `GOTOOLCHAIN=local`. La première invocation OSV R3 a échoué par syntaxe de chemin (`scan source --lockfile go.mod`); elle a été conservée et rejouée correctement avec `scan --lockfile go.mod`, exit code 0.

## Baseline R3

La baseline source-only a été exécutée avec `gosec ./cmd/... ./internal/...` au HEAD overnight final. Elle contient **132 findings** : G101=1, G104=36, G107=1, G115=8, G118=1, G122=1, G204=5, G301=17, G302=4, G304=20, G305=1, G306=1, G404=17, G703=12 et G704=7. Aucun finding n’a été supprimé par `nosec`, `nolint`, skip ou allowlist globale.

La matrice complète est `GOSEC_R3_BASELINE_MATRIX.tsv` et sa version lisible est `GOSEC_R3_BASELINE_MATRIX.md`.

## Lot R3-A — filesystem et archives

Le correctif `233b1eefee9c057b975f4293d0b38cec9463cd2c` remplace l’ouverture directe d’un fichier régulier dans le callback de backup tar par `os.Root.Open`, conserve le rejet des types spéciaux et propage l’erreur de fermeture. Un test synthétique vérifie qu’un symlink externe reste une entrée symlink et que son contenu externe n’est pas dereferencé.

Le scan après R3-A est passé de G122=1 à 0 et de G304=20 à 19. Les 34 entrées R3-A sont classées dans `GOSEC_R3_A_MATRIX.tsv` : deux `CORRECTED_AND_VERIFIED`, une `MITIGATED_CONTROL_SCANNER_OPEN` pour G305, et les entrées restantes `NEEDS_MANUAL_REVIEW` car Gosec conserve les alertes taint/path sur les arguments CLI explicitement contrôlés. Le package de restore conserve confinement lexical, séparation Windows, limites, rejet des hardlinks/types spéciaux, staging et rollback.

## Lot R3-B — permissions et erreurs I/O

Le correctif `25d2f8890203b1770e9702d7e97580b80229f19a` durcit les chemins réellement utilisés : répertoires de groupes et de runtime owner-only, output de backup metadata en 0700/0600, téléchargements temporaires en 0600, marqueurs de version en 0600, erreurs de `chmod`, suppression, écriture et fermeture propagées, modes d’archives et de restauration plafonnés à owner-only. Les tests couvrent les permissions du store de groupes, du backup metadata, du restore et des installations synthétiques.

Le commit `63a6175a71faf73d2071cbf80c38d5d264885e10` ferme les trois derniers G301 sur les répertoires de données Chromium et le parent de configuration. Après ces corrections, G301=0 et G306=0. G104 est passé de 36 à 27. G302 reste à 5 parce que Gosec conserve des alertes sur les exécutables explicitement chmodés 0755 à l’intérieur de répertoires 0700 et sur une permission signalée par le scanner; ces lignes restent ouvertes et classées.

## Lot R3-C — conversions, contexte et classification

Le correctif `57420297849949369bf53c1a66631e783f2d8908` remplace les quatre conversions implicites byte de l’identifiant opaque par `binary.BigEndian`, sans changer son format de 16 octets. Il remplace également `context.Background()` par le contexte attach borné pour la transition durable journalisée. Les tests de launch, fingerprint, browser et humanize passent.

G115 est passé de 8 à 4 et G118 de 1 à 0. Les quatre G115 restants sont bornés par des contrôles de taille/valeur et restent `MITIGATED_CONTROL_SCANNER_OPEN`. G404=17 est `NEEDS_MANUAL_REVIEW` : les usages concernent la simulation humanisée ou la sélection de fingerprint, pas une autorisation cryptographique; aucune substitution comportementale non validée n’a été faite. G101=1 reste `NEEDS_MANUAL_REVIEW` car le finding porte sur le contexte de métadonnées du token admin, pas sur un secret embarqué.

## Résultats finaux

Le scan source-only final au commit source `63a6175` contient 94 findings ouverts : G101=1, G104=27, G107=1, G115=4, G204=5, G302=5, G304=19, G305=1, G404=17, G703=12 et G704=7. Les règles G118, G122, G301 et G306 sont à zéro. La réduction constatée est donc de 132 à 94 findings visibles; elle ne constitue pas une clôture de sécurité.

| Classification individuelle de la baseline | Nombre |
|---|---:|
| `CORRECTED_AND_VERIFIED` | 25 |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 16 |
| `NEEDS_MANUAL_REVIEW` | 91 |
| `HISTORICAL_NOT_REACHABLE` | 0 |

Les 132 lignes et leurs justifications sont dans `GOSEC_R3_FINAL_MATRIX.tsv` et `GOSEC_R3_FINAL_MATRIX.md`.

Les gates finales ont réussi pour `go test -count=1 -race ./cmd/... ./internal/...`, `go vet`, `go build`, Govulncheck, Gitleaks, OSV Go, OSV pnpm sur `forge-dashboard/pnpm-lock.yaml`, Trivy filesystem et Syft. `forge-dashboard/pnpm run check` a réussi avec `tsc --noEmit`. Semgrep, Grype, Shellcheck et Yamllint restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`. Les scans globaux contaminés par snippets incomplets sous `artifacts/` ne sont pas utilisés comme scans source-only de référence.

## Verdict et gates maintenus

```text
GOSEC_R3_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R2_OVERNIGHT_HARDENING_COMPLETE_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R2_LOT2_PRESERVED_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
PUBLIC_RELEASE_BLOCKED
SCAN_BLOCKED_UNKNOWN
NATIVE_SYSTEMVAULT_NOT_TESTED
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```

Cette mission ne déclare ni production-ready, ni release, ni validation native Camoufox/SystemVault/Docker. La prochaine étape doit faire l’objet d’une nouvelle autorisation étroite; T29 ne démarre pas automatiquement.

## Références de preuve

Les preuves brutes, matrices, rapports scanner, sidecars et packages se trouvent dans le répertoire versionné [`evidence/GOSEC_R3/`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/validation/operational-v1/evidence/GOSEC_R3). Le package overnight v4 antérieur est conservé dans [`evidence/GOSEC_R2_OVERNIGHT/`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/validation/operational-v1/evidence/GOSEC_R2_OVERNIGHT).
