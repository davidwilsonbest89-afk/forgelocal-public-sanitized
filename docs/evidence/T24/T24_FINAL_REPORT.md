# Rapport final T24 — Opérations bulk bornées

## Décision locale proposée

> **`T24_IMPLEMENTED_TESTED_LOCAL_PENDING_SECURITY_REVIEW`**

T24 est implémenté et les tests fonctionnels locaux, la compilation et les contrôles de secret ont réussi. Cette décision ne constitue pas une approbation de release ni une levée de gate : les scans de vulnérabilités de dépendances complets ont été interrompus par les limites de ressources de la sandbox et doivent être rejoués sur une machine de qualification disposant de davantage de mémoire.

| Champ | Valeur |
|---|---|
| Baseline | `df969533cdd446be41d868bebea2c8a106f543d5` |
| Tag de baseline | `t23-scope-and-evidence-correction-2026-08-19` |
| Commit T24 | `d811a5c71e4b9398e358f4469408a1917b608a3d` |
| Branche | `t24-bulk-operations` |
| Contrat | `docs/T24_BULK_OPERATIONS_CONTRACT.md` |
| Journal BASELINE_DISCOVERY | `/home/ubuntu/forgelocal-t24-baseline-discovery/BASELINE_DISCOVERY_RAW.log` |

## Périmètre livré

Le Core expose `POST /api/profiles/bulk`, sous les mêmes middlewares d’authentification, de loopback et d’origine locale que les écritures unitaires. Les opérations admises sont `archive`, `reopen`, `add_tag`, `remove_tag`, `set_group` et `clear_group` sur au plus 50 profils distincts, dans l’ordre de la requête.

Les résultats sont explicitement par profil (`changed`, `noop`, `failed`) ; aucune transaction globale trompeuse n’est annoncée. La mutation passe par les primitives de store existantes, conserve les verrous par profil et les séquences History, et écrit des audits redacted avec `correlation_id`.

## Tests et contrôles passants

| Contrôle | Verdict | Preuve brute |
|---|---|---|
| Tests T24 ciblés, `-race` | PASS : 6 scénarios | `raw/go-test-race-final.log` couvre également la suite globale. |
| Suite Go complète, `-race` | PASS : tous les packages exécutés, aucun `DATA RACE` | `raw/go-test-race-final.log` |
| `go vet ./...` | PASS | `raw/go-vet-final.log` |
| `go build ./...` | PASS | `raw/go-build-final.log` |
| `git diff --check baseline..HEAD` | PASS | `raw/diff-check-final.log` |
| Gitleaks delta Git réel | PASS : `[]` | `raw/gitleaks-delta.log`, `raw/gitleaks-delta.json` |
| Trivy secrets filesystem | PASS : aucun secret détecté | `raw/trivy-secret-final.log` |
| SBOM Syft CycloneDX | PASS : 781 composants inventoriés | `raw/syft-sbom.log`, `raw/t24-sbom.cdx.json` |
| Gosec | 167 findings globaux, **0 dans les fichiers Go T24** | `raw/gosec-head.log`, `raw/gosec-head.json` |

La suite T24 couvre les transitions archive/réouverture, no-op idempotent, tags, groupe existant, `partial success`, erreurs par profil, JSON strict, limite de 50 cibles, authentification, loopback, Origin/Referer, annulation avant mutation, concurrence, History unique et redaction des audits.

## Scans non conclusifs dans cette sandbox

| Contrôle | État | Motif vérifiable | Suite requise |
|---|---|---|---|
| Govulncheck | Interrompu (`143`) | Le processus a été terminé avant résultat à deux reprises malgré l’exécution isolée ; logs `raw/govulncheck-head.log` et `raw/govulncheck-final.log`. | Rejouer `govulncheck ./...` sur un hôte de qualification à mémoire suffisante. |
| OSV-Scanner | Interrompu (`143`) | Le scan du projet a identifié 22 packages puis a été terminé avant résultat ; `raw/osv-scanner-final.log`. | Rejouer `osv-scanner scan source . --format json`. |
| Staticcheck global | Findings historiques | Les findings listés ne ciblent pas les fichiers T24 ; `raw/staticcheck-t24.log`. | Traiter séparément la dette historique, sans l’attribuer à T24. |

Ces éléments justifient le maintien de `SCAN_BLOCKED_UNKNOWN` et empêchent toute qualification de release.

## Invariants maintenus

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` demeurent inchangés. T24 n’exécute ni runtime, ni Camoufox, ni proxy réel, ni coffre natif, ni Dashboard, ni release.

## Prochaine action autorisée

Préserver le bundle et le kit T24, puis soumettre ce lot à une revue indépendante après le rejeu de Govulncheck et OSV-Scanner sur l’environnement approprié. **T25 ne démarre pas automatiquement.**
