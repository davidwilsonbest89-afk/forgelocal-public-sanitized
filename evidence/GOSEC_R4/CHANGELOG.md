# GOSEC-R4 changelog

## 2026-08-26

Le lot R4-A a rendu observables les erreurs d’I/O MCP, les erreurs de cleanup browser/relay, les erreurs keyboard et les erreurs de transport JSON-RPC. Le scan G104 est passé à 0 sur le périmètre source-only.

Le lot R4-B a root-scopé les accès aux fichiers de backup et de download avec `os.Root`, refusé les fichiers d’entrée/sortie symlinkés et ajouté des régressions positives et négatives pour les symlinks externes, les traversals et les sorties symlinkées.

Le lot R4-C a borné les conversions de taille d’archive, refusé les redirections du helper download et ajouté les tests de borne. G107 est passé à 0. Les G101, G115, G302 et G404 restent visibles et classés manuellement.

Le lot R4-D a conservé les contrôles loopback stricts du CLI et du WebSocket Playwright, les timeouts et le refus des redirections. Les G204 et G704 restent visibles, avec Darwin/GUI non exécutés.

Les scans finaux Go, Dashboard, Govulncheck, Gitleaks, OSV corrigé, Trivy et Syft ont été exécutés. Semgrep, Grype, Shellcheck et Yamllint restent indisponibles. Aucun environnement natif ou release n’a été forcé.
