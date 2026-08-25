# STATIC-DEBT-TRIAGE — lot documentaire séparé

**Base :** tag `t00-t42-v6-local-qualified-2026-08-25` / `999374d99b7996504ba91e421850a2fe84afb78d`  
**Décision :** `STATIC_DEBT_OPEN_INDIVIDUAL_OWNER_REVIEW`

Les matrices `V6_STATIC_DEBT_TRIAGE.md` contiennent une ligne individuelle pour chaque diagnostic : 34 Staticcheck et 89 GolangCI-Lint. Les champs couvrent règle, fichier, ligne, message, risque, correction possible, propriétaire et lot futur. Les rapports JSON/bruts utilisés restent conservés dans le dossier de preuve.

Aucun diagnostic n’est supprimé, masqué par `nolint`, transformé en succès ou mélangé avec le gel V6. Les défauts à impact élevé sont à prioriser uniquement après revue du propriétaire, notamment les erreurs non gérées sur IO/secrets et les API dépréciées sur des chemins réellement exposés. Chaque correction future devra être un commit séparé accompagné de tests ciblés, puis d’un rerun Staticcheck/GolangCI-Lint. Les autres lignes restent dette qualité/maintenabilité jusqu’à décision individuelle.

Le propriétaire de décision est l’équipe mainteneurs Go. La condition de levée est une classification individuelle, une correction testée ou une justification acceptée, sans modifier les gates de release.
