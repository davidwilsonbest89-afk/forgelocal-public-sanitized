# GOSEC-R6 Lot B — rapport final post-package

Ce document compagnon complète `R6_B_REPORT.md` avec les références du package et de la vérification publique. Le package R6-B v1 et son rapport inclus sont conservés sans reconstruction.

## Résultats Lot B

La baseline recalculée depuis le HEAD R6-A contenait 12 findings : G204=5 et G704=7. Le scan post-correctif depuis `a436a68d1ca4e223b54c45d7b99efa77617e620b` contient encore G204=5 et G704=7. Les 12 lignes de la matrice restent `MITIGATED_CONTROL_SCANNER_OPEN`; les commandes natives `open`, `xdg-open`, `rundll32` et `xattr` restent en plus `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` lorsque leur plateforme n’est pas disponible.

Le seul changement de code du Lot B est l’allowlist workflow `full_page=true|false`; les queries non prévues ou supplémentaires sont refusées. Les contrôles loopback HTTP et WebSocket existants sont conservés. Les tests ciblés, race, vet, build, Govulncheck, OSV Go/pnpm, Trivy et Gitleaks corrigé sont PASS dans leurs périmètres documentés. Gosec retourne exit code 1 parce que les 12 findings restent signalés.

## Conservation et vérification publique

Le commit source est `a436a68d1ca4e223b54c45d7b99efa77617e620b`. Le commit d’évidence est `e0c9e16b879f0ef6ee1a6cf733d84eb8a172e854`. Le commit package est `f505f4a5e1a4d14df40c13555e8ec24b0e305f4c`. Le package est `forgelocal-gosec-r6b-final-v1.zip` et `.tar.gz`, avec sidecars de SHA-256; le bundle `forgelocal-gosec-r6b-delta-a436a68-e0c9e16.bundle` retourne PASS avec `git bundle verify`.

La vérification publique `R6B_PUBLIC_VERIFICATION_RAW.log` retourne PASS pour les hashes, extractions fraîches, manifestes, absence réelle de membre `SMOKE_INTEGRATED_PROXY`, bundle, clone GitHub neuf, checkout explicite du commit package et `git fsck --full` avec exit code 0.

## Limites et verdict

Aucun compte, cookie, secret, proxy commercial, site externe, runtime de production, Camoufox, SystemVault natif ou Docker/Buildx n’a été utilisé. Les plateformes Windows/macOS et les GUI natifs n’ont pas été simulés. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts.

Le verdict est `GOSEC_R6_LOT_B_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
