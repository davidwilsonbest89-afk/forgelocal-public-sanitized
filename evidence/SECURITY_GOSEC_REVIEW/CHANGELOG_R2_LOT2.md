# Changelog — GOSEC-REVIEW-R2 Lot 2

## 2026-08-26 — subprocess et réseau

Le lot ajoute une policy CLI HTTP(S) loopback pour les appels GET/POST, metadata backup/restore et `open --base-url`. Elle refuse hôte externe, userinfo, query, fragment et ports invalides, et bloque les redirections externes avant dial.

Le pont WebSocket conserve la policy `ws` loopback et bénéficie d’un dial borné et d’une deadline de handshake de cinq secondes. Le subprocess macOS `xattr` utilise un contexte de dix secondes et des arguments séparés.

Les tests couvrent IPv4 loopback, IPv6 loopback, localhost, URL externe, schémas invalides, ports, userinfo, query, fragment, redirect externe, API GET/POST, metadata, open externe, nominal local et timeout local. Le bridge Playwright complet et les subprocess GUI/Darwin ne sont pas exécutés faute d’environnement autorisé; ils ne sont pas déclarés PASS.

Gosec source-only passe de 155 à 152 findings; G204 passe de 6 à 5 et G704 reste à 7. Les findings statiques restent ouverts. OSV `go.mod` reste ouvert avec 46 avis; OSV Dashboard, Govulncheck, Gitleaks et Trivy passent.

Source commit : `701c5949261de261d2044cbff3e125b88c56f1a2`.
