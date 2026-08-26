# GOSEC-R6 Lot A — changelog

## Commit source

`142477ae0d576eae937b16660899fd973d6f2464` publie exclusivement le hardening des chemins et les tests de régression. Les fichiers de preuve ne sont pas inclus dans ce commit source.

Les changements ajoutent des ouvertures `os.Root` pour le token bootstrap, les profils migrés, les workflows, les fingerprints, les préférences Chromium, les téléchargements, les entrées TAR et les sorties de restore/export. Les symlinks et types non réguliers sont refusés lorsque le contrôle local le permet. Les bornes d’archives et le rollback existants sont conservés.

## Résultats

La baseline Lot A était G703=9, G304=11 et G305=1. Le scan post-correctif retourne G703=7, G304=0 et G305=1. Le scan global passe de 59 à 46 findings. Les 11 G304 et deux sinks G703 supprimés sont rattachés aux changements root-scoped; les sept G703 et le G305 restants restent `MITIGATED_CONTROL_SCANNER_OPEN`.

Les tests ciblés, la suite race, vet, build et `git diff --check` sont PASS. Gosec conserve un exit code non nul parce que des findings restent ouverts; ce n’est pas un échec caché ni une suppression de findings.
