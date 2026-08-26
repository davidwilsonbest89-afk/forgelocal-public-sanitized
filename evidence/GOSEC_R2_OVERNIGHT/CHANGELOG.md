# Changelog — GOSEC R2 overnight

## 2026-08-26 — hardening local et conservation

La lignée Lot 1/Lot 2 a été réconciliée depuis un clone neuf. `31385c2` est un descendant direct de `20e5181`, puis `701c594`, `2091480`, `aab0cca`, `18367a6`, `7f3f5be` et `f796299` forment la suite publiée sur `validation/operational-v1`.

Le filesystem complémentaire a reçu la normalisation correcte des séparateurs Windows et la protection root-scoped des artifacts HTTP. Les permissions des logs, configs, backups, user-data, préférences et captures ont été resserrées. Le workflow HTTP est loopback-only, timeouté et bloque les redirections externes. La directive Go `1.25.13` supprime les 46 avis OSV correspondant au `go 1.25.0` déclaré, sans suppression de scanner.

Les tests source-only race/vet/build, tests ciblés, Dashboard check, Govulncheck, Gitleaks, OSV actuel, Trivy et Syft sont consignés. Gosec source-only reste à 132 findings et `exit_code=1`. Semgrep, Grype, Shellcheck et Yamllint restent indisponibles. Camoufox/SystemVault/Docker/release restent non exécutés.

Verdict : `GOSEC_R2_OVERNIGHT_HARDENING_COMPLETE_WITH_OPEN_FINDINGS`.
