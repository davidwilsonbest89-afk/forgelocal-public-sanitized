# V6 — findings résiduels et conditions de revue

Ce document distingue les contrôles corrigés des résultats non nuls conservés. Aucun résultat ci-dessous n’est transformé en PASS implicite.

| ID | Contrôle | Résultat | Nature | Propriétaire | Condition de levée | Date de revue |
|---:|---|---:|---|---|---|---|
| R1 | Gitleaks plage Git | `0 commits scanned` avec `--log-opts=BASE..HEAD` | Limite reproductible de Gitleaks 8.18.4 ; non-preuve de couverture | Équipe sécurité/release | Utiliser une version corrigée ou confirmer indépendamment le scan par arbre ; conserver 58 arbres V5 + 4 arbres V6 | 2026-09-08 |
| R2 | Gitleaks checkout | code 0 avec exception exacte | Faux positif confirmé : empreinte publique PPA dans six artefacts de provenance | Équipe sécurité/release | Revue indépendante du chemin/règle/empreinte ; aucune extension de l’allowlist | 2026-09-08 |
| R3 | Semgrep `rand.Read` | 18 résultats | Faux positifs contextuels : imports réels `crypto/rand`, usages documentés individuellement | Mainteneurs Go | Remplacer la règle par une règle qualifiée par import ou refaire la revue de contexte | 2026-09-08 |
| R4 | staticcheck | 34 diagnostics, code 1 | Diagnostics historiques hors périmètre des deux corrections V6 ciblées | Mainteneurs Go | Corriger ou accepter individuellement par fichier ; aucune suppression globale | 2026-09-08 |
| R5 | GolangCI-Lint | 89 findings, code 1 (`errcheck` 50, `staticcheck` 22, `unused` 17) | Findings historiques ; `ineffassign` et `sessions.go:310` corrigés | Mainteneurs Go | Revue individuelle des findings restants et tests affectés | 2026-09-08 |
| R6 | OSV Scanner v1.9.2 | 46 résultats stdlib | La v1.9.2 lit `go 1.25.0` et ne reflète pas le patch toolchain `go1.25.13` ; `govulncheck` effectif est à 0 | Build/security | Rejouer avec un scanner compatible avec le patch toolchain ou joindre une preuve de version stdlib effective | 2026-09-08 |
| R7 | Trivy | 0 vulnérabilité, 0 secret, 6 misconfigurations Docker | Root user, HEALTHCHECK absent et `--no-install-recommends` absent dans deux Dockerfiles | Propriétaire images/release | Revue et test d’image autorisés ; aucune modification runtime risquée dans V6 | 2026-09-08 |
| R8 | Git LFS | `git lfs fsck` code 1 ; 14 objets historiques absents du remote accessible | Limite de disponibilité des objets LFS historiques ; fetch ciblé exécuté sans pull global | Administrateur dépôt | Restaurer les 14 objets LFS ou fournir une preuve canonique distante | 2026-09-08 |
| R9 | Licences | 744 composants, 741 `UNKNOWN` | Métadonnées de licence absentes dans le SBOM, non reclassées automatiquement | Legal/OSS review | Revue humaine des 741 composants inconnus | 2026-09-08 |
| R10 | Grype répertoire brut | 46 matches | Artefacts `node_modules/@esbuild` inclus par scan large ; SBOM propres excluant `node_modules` : 0 match CycloneDX et SPDX | Build/release | Conserver les SBOM propres comme périmètre de livraison et revoir le périmètre du scan brut | 2026-09-08 |

Le résultat final reste `T00_T42_V6_FINDINGS_REMEDIATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`. Les gates ne sont pas modifiées.
