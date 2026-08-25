# TODO — POST-V6-REVALIDATION-FINAL-MATRIX

| Priorité | Action | Condition de clôture | Owner |
|---|---|---|---|
| Haute | Revue indépendante T10/T15 | run reproduit, logs redacted et cleanup acceptés | QA / Security |
| Haute | Rejouer Docker image-validation | Docker/Buildx réels, build, inspect, Trivy image et HEALTHCHECK archivés | Plateforme |
| Haute | Recueillir décisions T28/T29/T39/T40/T41/T42 | document rempli et approuvé par Product/Security/Privacy/QA | Product |
| Haute | Autoriser séparément les protocoles externes | P-C1 à P-C4 autorisés un par un, ou maintenus non exécutés | Governance |
| Permanente | Maintenir les gates | aucune release, migration, runtime production ou fonctionnalité bloquée sans décision explicite | Tous owners |
