# OSV-TOOLCHAIN-RECONCILIATION — lot séparé

**Base gelée :** `t00-t42-v6-local-qualified-2026-08-25` / `999374d99b7996504ba91e421850a2fe84afb78d`  
**Branche :** `audit/t00-t42-osv-toolchain-reconciliation`  
**Décision :** `OSV_RESULTS_RECONCILED_AS_VERSION_MODELING_EXCEPTIONS_PENDING_COMPATIBLE_SCANNER`

Le rapport OSV Scanner v1.9.2 contient 46 advisories individuelles concernant `stdlib` en version détectée `1.25.0`, issue de la directive `go 1.25.0`. Le projet est compilé et testé avec le binaire effectif `go1.25.13`; cette différence de patch n’est pas représentée par le résultat OSV v1.9.2. La matrice `V6_OSV_RECONCILIATION_MATRIX.md` conserve une ligne par advisory avec ID, alias/CVE, package, version, versions fixes, import/surface, exposition et décision.

La comparaison indépendante est conservée dans les artefacts bruts : `govulncheck@v1.1.4` retourne 0 vulnérabilité sur le code compilé avec Go 1.25.13 ; Grype 0.117.0 retourne 0 match sur les SBOM propres CycloneDX et SPDX ; Syft 1.51.0 produit ces SBOM propres. Aucun résultat OSV n’est supprimé, ignoré globalement ou transformé en PASS. Chaque ligne est une exception individuelle de modélisation jusqu’à l’obtention d’un scanner compatible avec la version patch effective.

Les versions exactes sont : Go effectif `1.25.13`, directive `go 1.25.0`, OSV Scanner `1.9.2`, Govulncheck `v1.1.4`, Grype `0.117.0`, Syft `1.51.0`. Une CVE affectant réellement le binaire Go 1.25.13 ou une dépendance livrée déclencherait une mise à jour minimale, des tests ciblés et une nouvelle SBOM. Cette condition n’est pas satisfaite par les comparaisons disponibles ; la divergence de version reste ouverte à revue indépendante.

Le lot ne modifie aucun fichier produit, aucune dépendance, aucun wrapper historique ni aucune gate. La condition de levée est un scan OSV compatible avec le patch toolchain ou une preuve indépendante et reproductible établissant la version stdlib réellement compilée pour chacune des 46 lignes.
