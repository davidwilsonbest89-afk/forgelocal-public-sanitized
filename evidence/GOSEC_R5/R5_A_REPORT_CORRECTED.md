# GOSEC-R5 Lot R5-A — correction de rattachement des preuves

Le rapport initial `R5_A_REPORT.md` est conservé comme artefact publié. Cette version corrigée précise la chronologie observée dans le raw : `R5_A_FINAL_RAW.log` a été exécuté à `HEAD=096e3f591f273801cf735669d7a135daf6a7f601` avec les modifications de configuration et de marqueur runtime alors non committées. Il ne contient donc pas encore le contrôle `.api-token` ajouté ensuite dans `internal/api/router.go`.

Le commit source R5-A finalement publié est `54ed3a4964806eeb4880c9ebb3949d410c335174`; il contient le hardening configuration, marqueur runtime et token bootstrap. La mesure Gosec propre au raw R5-A reste **63 → 60**, avec trois G304 supprimés : deux dans `internal/config/config.go` et une dans `internal/browser/download.go`. Le contrôle token est mesuré séparément dans le raw R5-B, qui part du scan R5-A et observe **60 → 59**.

Aucun artefact antérieur n’est supprimé ni réécrit. La correction évite d’attribuer au scan R5-A un contrôle qui n’était pas présent au moment de son exécution.
