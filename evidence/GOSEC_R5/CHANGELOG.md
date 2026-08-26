# GOSEC-R5 — changelog

## 2026-08-26 — R5 finalisation en cours de conservation

La branche dédiée `validation/gosec-r5` a publié le hardening source R5-A dans `54ed3a4964806eeb4880c9ebb3949d410c335174`. Les contrôles ajoutés refusent les fichiers de configuration, marqueurs runtime et token bootstrap symlinkés et utilisent des ouvertures root-scoped. Les tests positifs et négatifs associés passent.

Les preuves R5-A, R5-B et R5-C ont été publiées séparément. Gosec source-only est passé de 63 à 59 findings; les 59 findings finaux restent visibles et ouverts. Les tests Go race, vet, build, Gitleaks, OSV, Trivy, Syft et le check TypeScript Dashboard passent dans les preuves finales. Govulncheck global a échoué sur des snapshots historiques non compilables; le rerun source-only passe.

Aucune branche historique n’a été modifiée. T28 n’a pas été rouvert, T29 n’a pas été démarré et T31–T38 n’ont pas été modifiés. Les gates production, Camoufox, SystemVault natif, proxy/cookies réels, Docker/Buildx et release restent ouvertes ou non exécutées.
