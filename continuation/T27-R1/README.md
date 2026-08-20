# ForgeLocal — kit de continuité T27-R1

## Point de reprise

La tête de reprise est `03767204664d2bf5d359f59d103acd202edf6bfb`. La décision locale est `T27_R1_TECHNICAL_CLOSURE_APPROVED_VERIFIABLE_LOCAL_PENDING_CR08_NATIVE_VALIDATION`.

## Réhydratation obligatoire

1. Cloner cette branche avec Git LFS, puis exécuter `git lfs pull`.
2. Vérifier tous les fichiers `*.sha256` avec `sha256sum -c` dans leur répertoire.
3. Extraire `evidence/CR01/forgelocal-t27-r1-cr01-evidence-df50a2b.zip` dans un répertoire neuf et vérifier son manifeste.
4. Vérifier le bundle CR-01, puis appliquer les bundles delta CR-02, CR-03, CR-04 et CR-05 dans l’ordre indiqué par `evidence/CHAIN.md`.
5. Contrôler le commit obtenu avec `git rev-parse HEAD` et lancer `git fsck --full`.
6. Lire `TOOLCHAIN.lock`, le contrat T27-R1, le registre et `todo.md` avant toute écriture.

## Règle impérative pour tout futur lot

Avant toute modification de code, créer `BASELINE_DISCOVERY_RAW.log` qui contient les commandes exactes, les chemins, les dates UTC, les codes de sortie et les sorties brutes. Après qualification, créer une copie canonique hashée, son sidecar, un manifeste non auto-référentiel, une extraction neuve, un re-scan et une inscription au registre.

## Exclusions de sécurité

Ce kit exclut les secrets, tokens, cookies, profils, bases de données, runtimes, `node_modules`, builds générés et attestations privées hors périmètre de reprise. Aucun contenu inclus n’autorise une release, Camoufox, un proxy réel, un cookie réel ou SystemVault natif.
