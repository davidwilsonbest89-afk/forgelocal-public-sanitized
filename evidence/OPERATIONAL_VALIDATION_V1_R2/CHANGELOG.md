# Changelog — OPV1 R2

## 2026-08-25 — `1e128bd3b2f1cb9b668afff25c7c155316fd0267`

### Corrigé

Le cycle de vie du token admin est désormais versionné et persistant : métadonnées sans valeur en clair, expiration TTL de 15 minutes, révocation persistante, validation centralisée sur les routes protégées et raisons d’erreur distinctes. L’endpoint loopback `POST /api/auth/revoke` révoque immédiatement le token courant et répond 204 en cas de succès.

Le Dashboard React sonde maintenant une route protégée réelle, distingue les états d’authentification expiré/révoqué/malformé/absent/invalide, propage `error.reason` dans les chemins d’écriture et déconnecte le contrôle local en mémoire après expiration ou révocation. Les contrastes Axe signalés, le focus clavier, les assets externes non disponibles et le chargement analytique placeholder ont été corrigés.

### Ajouté

Ajout des tests Core de cycle de vie, de concurrence/race, de redaction et d’intégration endpoint, ainsi que `forge-dashboard/tests/r2-auth-ui.spec.ts` pour les parcours UI expired/revoked avec APIs synthétiques routées en mémoire.

Ajout des preuves R2 : baseline redacted, journaux opérationnels, sorties Gitleaks/Gosec/OSV/govulncheck/Trivy, SBOM Syft, rapport final, registre, todo et manifeste/package associés.

### Maintenu volontairement

Le contrat et le code métier T28 n’ont pas été modifiés ni rouverts. T29 et T31–T38 n’ont pas été démarrés ni modifiés. Aucun runtime Camoufox, SystemVault natif, daemon Docker, compte réel, secret réel, cookie réel, proxy réel, release ou production n’a été utilisé.

### Verdict

`FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE` — `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés.
