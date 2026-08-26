# GOSEC-R6 Lot B — todo

Les 12 findings G204/G704 restent ouverts sous `MITIGATED_CONTROL_SCANNER_OPEN`. Une validation native séparée est requise pour `open`, `xdg-open`, `rundll32` et `xattr` sur leurs plateformes respectives; ne pas les simuler sous Linux.

Avant le Lot C, conserver le package Lot B, ses preuves publiques et ses hashes. Ne pas rouvrir T28, démarrer T29 ou modifier T31–T38. Maintenir `FORGELOCAL_PRODUCTION_READY=false`.
