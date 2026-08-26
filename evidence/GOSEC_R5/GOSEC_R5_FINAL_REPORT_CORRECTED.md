# GOSEC-R5 — rapport final corrigé après package et vérification publique

Le rapport `GOSEC_R5_FINAL_REPORT.md` et tous les raw antérieurs sont conservés. Cette version ajoute uniquement les résultats post-package et corrige le rattachement final.

## HEAD et package

| Référence | Valeur |
|---|---|
| Branche | `validation/gosec-r5` |
| Commit source | `54ed3a4964806eeb4880c9ebb3949d410c335174` |
| HEAD d’évidence utilisé pour le package | `517b55557297a6372bb5ef2dcb6d8c2b80a8db70` |
| Commit package GitHub | `b2fdecf97e3c28b29c374cf89f9bdbe4077d0bd7` |
| Package | `forgelocal-gosec-r5-final-v1.zip` et `.tar.gz` |
| Bundle | `forgelocal-gosec-r5-delta-096e3f5-517b555.bundle` |

Le bundle delta a été créé avec une référence temporaire locale et vérifié par `git bundle verify`. Le manifest est non auto-référentiel : il rattache le contenu au HEAD d’évidence `517b555` et ne dépend pas du commit qui contient les artefacts du package.

## Vérification post-publication

`R5_PUBLIC_VERIFICATION_RAW.log` documente les résultats suivants : hashes ZIP/TAR PASS, extraction fraîche TAR/ZIP PASS, comparaison des manifestes PASS, absence de `SMOKE_INTEGRATED_PROXY` dans les deux extractions, bundle verify PASS, clone public neuf PASS, checkout détaché PASS, `git fsck --full` avec `exit_code=0` PASS et vérification des deux HEAD distants PASS.

La ligne `PACKAGE_SMOKE_EXCLUSION_EXIT_CODE=1` signifie que `grep` n’a trouvé aucune occurrence, car la commande était préfixée par `!`; l’assertion d’exclusion est donc PASS et non un échec de package.

## Verdict inchangé

Le scan Gosec final reste à **59 findings ouverts** : G101=1, G115=3, G204=5, G302=5, G304=11, G404=17, G703=9 et G704=7. Les tests Go race/vet/build, Gitleaks, OSV Go/pnpm, Trivy, Syft, Dashboard check et Govulncheck source-only sont documentés PASS; les outils absents et environnements non exécutés restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`.

Les verdicts restent :

```text
GOSEC_R5_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R5_PARTIAL_ENVIRONMENT_UNAVAILABLE
FORGELOCAL_PRODUCTION_READY=false
```

Les invariants restent `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`. T29 n’a pas été démarré, T28 n’a pas été rouvert et T31–T38 n’ont pas été modifiés.
