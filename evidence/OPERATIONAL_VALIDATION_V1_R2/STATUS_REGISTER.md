# ForgeLocal — registre de statut opérationnel V1 R2

**Branche :** `validation/operational-v1`
**HEAD R2 :** `1e128bd3b2f1cb9b668afff25c7c155316fd0267`
**Baseline R2 :** `84f7bd2bd7c956a4766d1bd1ff2808c59f9c6eed`
**Verdict global :** `FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE`

## Statut courant

| Domaine | Statut | Référence |
|---|---|---|
| Core — cycle de vie token admin | `PASS` | `V-CORE-R2-AUTH_RAW.log` |
| Dashboard — périmètre Playwright couvert | `PASS` | `V-DASHBOARD-API_RAW.log` |
| Dashboard — Axe/console/réseau | `PASS` | `V-DASHBOARD-API_A11Y_CONSOLE_RAW.log` |
| SQLite crash recovery | `PASS` | `V-SQLITE-CRASH-RECOVERY_RAW.log` |
| Contrat runtime T28 | `PASS` | `V-T28-RUNTIME-CONTRACT_RAW.log` |
| Proxy/cookies synthétiques | `PASS` | `V-PROXY-COOKIES-SYNTHETIQUES_RAW.log` |
| Installation fraîche | `PASS` | `V-INSTALL-FRESH_RAW.log` |
| Chromium synthétique | `PASS` | `V-RUNTIME-SYNTHETIQUE_RAW.log` |
| Camoufox | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | `V-RUNTIME-SYNTHETIQUE_RAW.log` |
| MemoryVault, double de test | `PASS` — double uniquement | `V-SYSTEMVAULT-R2_RAW.log` |
| SystemVault natif | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | `V-SYSTEMVAULT-R2_RAW.log` |
| Docker/Buildx | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | `V-DOCKER-R2_RAW.log` |
| Gosec | `RESERVED_FINDINGS` | `GOSEC_OPERATIONAL.json`, `gosec.out` |
| OSV | `RESERVED_ADVISORIES` | `osv.go.out` |
| Semgrep/Grype/Shellcheck/Yamllint | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | `V-SECURITY_RAW.log` |

## Core et Dashboard

Le Core accepte un token valide, refuse les états `missing`, `malformed` et `invalid`, refuse l’alias `/api/v1` avec un token invalide, révoque par `POST /api/auth/revoke` en 204, puis refuse immédiatement et après redémarrage avec `reason=revoked`. Une métadonnée expirée après redémarrage produit `reason=expired`. Les tests race, la redaction et le cleanup sont passés.

Le Dashboard a obtenu 6 specs Playwright passées, dont les scénarios R2 `expired` et `revoked`, avec typecheck/build passés. Le contrôle Axe a obtenu 0 violation, dont 0 serious/critical ; la console a obtenu 0 erreur ; le réseau a obtenu 0 requête échouée et 0 réponse mauvaise.

Ces résultats ne ferment pas les zones explicitement non implémentées. Les actions d’espace de travail, journal d’audit, réglages, aide, notifications et filtres avancés restent des feedbacks `unavailable`. Les workflows UI T28 d’extension — installation, signature/provenance, allowlist, permissions, rollback et lifecycle — sont `NOT_IMPLEMENTED_PLACEHOLDER` et ne sont pas déclarés fonctionnels. Le contrat API T28 reste validé séparément, sans réouverture de T28.

## Sécurité, intégrité et release

`go test -race`, `go vet`, `go build` et `git diff --check` ont le code de sortie 0. Gitleaks ne relève aucun finding. `govulncheck` retourne 0 et « No vulnerabilities found. ». Trivy filesystem retourne 0 et Syft produit un SBOM de 747 artefacts. Les findings Gosec historiques et advisories OSV sont conservés sans suppression globale et maintiennent le gate `SCAN_BLOCKED_UNKNOWN`.

Les gates suivants restent inchangés : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. Le verdict n’autorise ni release ni production.

## Lots hors périmètre

T29 n’a pas été démarré. T31–T38 n’ont pas été modifiés. T28 n’a pas été rouvert ; son contrat a seulement été rejoué pour détecter une régression.
