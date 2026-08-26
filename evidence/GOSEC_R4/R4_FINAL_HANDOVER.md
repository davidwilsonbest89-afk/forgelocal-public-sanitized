# GOSEC-R4 — Handover final

## Publication

La branche dédiée [`validation/gosec-r4`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/validation/gosec-r4) est publiée sans modification de `validation/operational-v1`.

| Référence | Valeur |
|---|---|
| HEAD source R4 | `3cd3ac0fb9eb684e61eafb8746c69dda9781df64` |
| Commit preuves R4 | `75edd5480a1d283046b54445c150f0a220e07d66` |
| Commit package R4 v1 | `6c949a7c093f7f59ebedc79c084b53389cd649c7` |
| Commit vérification publique | `681b24f47748285110e2b29ea7ffc2b7f2b14eb4` |
| HEAD distant final R4 | `681b24f47748285110e2b29ea7ffc2b7f2b14eb4` |
| HEAD distant `validation/operational-v1` | `80048180d4ec5241c08146ade698d53a3f29454d` |
| Package source/evidence head | `3cd3ac0` / `75edd54` |

Le package canonique est `forgelocal-gosec-r4-final-v1.zip` et son TAR associé. Il a été construit avant le commit de vérification publique. Il ne contient donc pas ce dernier commit; cette différence est intentionnelle et documentée, et aucun package v2 n’a été produit pour éviter une boucle de régénération.

## Vérification publique

Le premier clone public a échoué au niveau TLS/RPC et est conservé dans `PUBLIC_R4_VERIFICATION_RAW.log`. Un retry neuf, avec `GIT_LFS_SKIP_SMUDGE=1` au clone et au checkout, `--depth=1`, `--filter=blob:none` et HTTP/1.1, a réussi. Le HEAD du clone correspondait au remote; `git fsck --full`, les sidecars, `unzip -t`, l’extraction ZIP/TAR, les hashes internes et `git bundle verify` ont tous retourné exit 0. Le bundle requiert le HEAD R3 `8004818` et contient le HEAD de preuves `75edd54`.

## Statut sécurité

Le scan final source-only reste à 63 findings Gosec : G101=1, G115=3, G204=5, G302=5, G304=15, G305=1, G404=17, G703=9 et G704=7. G104 et G107 sont à zéro. Govulncheck, Gitleaks, OSV corrigé Go/pnpm, Trivy, Syft, tests race/vet/build et Dashboard check passent. Semgrep, Grype, Shellcheck et Yamllint restent indisponibles.

## Gates

```text
GOSEC_R4_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R4_PARTIAL_ENVIRONMENT_UNAVAILABLE
PUBLIC_RELEASE_BLOCKED
SCAN_BLOCKED_UNKNOWN
NATIVE_SYSTEMVAULT_NOT_TESTED
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```

T28 n’a pas été rouvert, T29 n’a pas commencé et T31–T38 n’ont pas été modifiés. Le répertoire historique non suivi `evidence/SMOKE_INTEGRATED_PROXY/` est conservé localement et exclu des commits et packages R4.
