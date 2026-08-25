# T00–T42 V6 — matrice individuelle Grype

Le scan des SBOM CycloneDX et SPDX V5 a signalé deux correspondances High sur le même module sélectionné `golang.org/x/mod v0.37.0`. Il s’agit de deux advisories distincts, corrigés par `v0.40.0`.

| ID | Advisory / CVE | Sévérité | Version affectée | Version corrigée | Directe ou transitive | Chemin observé | Décision V6 |
|---:|---|---|---|---|---|---|---|
| 1 | GO-2026-6180 / CVE-2026-56864 | High | x/mod `< v0.40.0` ; sélectionnée `v0.37.0` | `v0.40.0` | Transitive, module indirect | `forgelocal → modernc.org/sqlite → modernc.org/libc → golang.org/x/mod` ; symbole `sumdb.Client.Lookup` | Mettre à jour au minimum `golang.org/x/mod v0.40.0`, puis tidy/verify/tests/rescans |
| 2 | GO-2026-6179 / CVE-2026-56865 | High | x/mod `< v0.40.0` ; sélectionnée `v0.37.0` | `v0.40.0` | Transitive, module indirect | `forgelocal → modernc.org/sqlite → modernc.org/libc → golang.org/x/mod` ; vérification des tuiles `sumdb/tlog` | Mettre à jour au minimum `golang.org/x/mod v0.40.0`, puis tidy/verify/tests/rescans |

`go mod why -m golang.org/x/mod` confirme l’utilisation via `modernc.org/libc` et `golang.org/x/mod/semver`. `go mod graph` montre que le module sélectionné est `v0.37.0`, tandis que plusieurs anciennes versions apparaissent comme contraintes historiques de modules transitifs ; la version MVS effectivement utilisée est celle du module sélectionné.

Les advisories officielles indiquent que **GO-2026-6180 / CVE-2026-56864** concerne la possibilité pour un GOSUMDB malveillant de servir du contenu de module non présent dans le journal de transparence [1], et que **GO-2026-6179 / CVE-2026-56865** concerne un contournement de vérification de tuiles sumdb via un GOPROXY malveillant [2]. Les deux rapports indiquent `golang.org/x/mod` avant `v0.40.0` comme affecté et `v0.40.0` comme version corrigée [1] [2].

La correction V6 est donc une mise à jour minimale et compatible avec le module Go `1.25.0` du projet. Une exception individuelle n’est pas retenue puisque la version corrigée est disponible.

## Références

[1]: https://pkg.go.dev/vuln/GO-2026-6180 "Go Vulnerability Database — GO-2026-6180 / CVE-2026-56864"
[2]: https://pkg.go.dev/vuln/GO-2026-6179 "Go Vulnerability Database — GO-2026-6179 / CVE-2026-56865"
