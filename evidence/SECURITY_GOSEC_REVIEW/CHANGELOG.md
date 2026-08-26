# Changelog — SECURITY_GOSEC_REVIEW

## 2026-08-26 — revue ciblée et hardening

Le lot ajoute une revue d’exploitabilité des 177 findings Gosec source-only, une classification P0/P1/P2 et un rapport de gate. Le pont WebSocket Playwright interne valide désormais le schéma, l’hôte loopback, le port, le chemin de session et l’absence de composants URL dangereux avant tout dial TCP. Les tests négatifs couvrent hôtes externes, schéma incorrect, port, query, chemin et userinfo.

Le runner Playwright externe a été durci localement avec polling borné par profil, diagnostics minimaux et assertion `external_forward_observed=false`. Il reste volontairement hors dépôt et hors package public de preuves.

Le gate Gosec reste ouvert : baseline/post-correctif `177 → 177`, `new_findings=0`, sans suppression ni allowlist globale. Les gates source-only Go, Dashboard, Gitleaks et Govulncheck sont documentés dans le rapport. Les erreurs du glob `./...` proviennent des snippets Go partiels sous `artifacts/` et sont conservées comme réserve de packaging, non converties en PASS.

T28 n’est pas redémarré, T29 n’est pas commencé et T31–T38 ne sont pas touchés. Aucune validation Camoufox, SystemVault natif, Docker/Buildx, proxy commercial, cookie ou site externe n’a été exécutée.
