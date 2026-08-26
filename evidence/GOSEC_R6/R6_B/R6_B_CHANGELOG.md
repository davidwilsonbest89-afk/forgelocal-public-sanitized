# GOSEC-R6 Lot B — changelog

Le commit source `a436a68` ajoute une allowlist de query workflow : `full_page=true|false` est autorisée, les queries non prévues sont refusées. Les gardes CLI/WebSocket existants sont conservés.

Le Lot B ne ferme aucun G204/G704 statique : la baseline et le scan final restent à G204=5, G704=7. Les 12 lignes sont classées individuellement dans `R6_B_FINDING_MATRIX.tsv`. Les contrôles applicatifs sont documentés comme mitigations; les chemins GUI/macOS/Windows non exécutés restent indisponibles.
