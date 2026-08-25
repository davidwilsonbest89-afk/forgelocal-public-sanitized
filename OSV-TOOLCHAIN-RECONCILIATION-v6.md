# OSV-TOOLCHAIN-RECONCILIATION — lot séparé

**Décision :** `OSV_RESULTS_RECONCILED_AS_INDIVIDUAL_VERSION_MODELING_EXCEPTIONS_PENDING_COMPATIBLE_SCANNER`

**Base contrôlée :** audit OSV publié au commit `89dea8888fefd50422d1bd509f88079833fafbdc`. La branche de suivi est indépendante du tag V6 gelé et ne modifie ni code produit, ni dépendance, ni wrapper historique.

Le résultat OSV Scanner `1.9.2` contient **46 advisories individuelles** pour le package `stdlib`, détecté en version `1.25.0` depuis la directive de module. Le binaire effectif de compilation et de test est `go1.25.13`. Cette différence de patch n’est pas représentée par le résultat du scanner ; elle est donc traitée individuellement dans `V6_OSV_RECONCILIATION_MATRIX.md`, avec advisory, alias/CVE, package, version détectée, version corrigée, import/surface, exposition, comparaison et décision.

La comparaison conservée dans les artefacts est la suivante : Govulncheck `v1.1.4` retourne zéro vulnérabilité sur le code compilé avec Go `1.25.13`, et Grype `0.117.0` retourne zéro match sur les SBOM propres CycloneDX et SPDX produits par Syft `1.51.0`. Aucun advisory n’est supprimé, masqué, globalement désactivé ou transformé en PASS. Chaque ligne reste une exception individuelle `INDIVIDUAL_EXCEPTION_VERSION_MODELING_DIVERGENCE` jusqu’à l’obtention d’un scanner OSV compatible avec le patch effectif ou d’une preuve reproductible de la stdlib réellement embarquée.

Une CVE effectivement applicable au binaire ou à une dépendance livrée déclencherait la correction minimale, les tests ciblés et la régénération des SBOM dans un lot distinct. Ce lot ne lance aucun runtime, ne démarre pas T28/T29/T39/T40/T41/T42, Camoufox, proxy réel, cookies, données utilisateur ou SystemVault natif, et ne constitue pas une release.

**Owner :** mainteneurs Go et sécurité dépendances. **Condition de levée :** scan OSV compatible avec Go `1.25.13` ou preuve indépendante par advisory établissant l’exposition effective.
