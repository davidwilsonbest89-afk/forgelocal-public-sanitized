# GITLEAKS-HISTORICAL-METHOD-HARDENING — lot séparé

**Décision :** `GITLEAKS_HISTORICAL_TREE_BY_TREE_METHOD_HARDENED`

**Base du lot :** tag V6 `t00-t42-v6-local-qualified-2026-08-25` / commit `999374d99b7996504ba91e421850a2fe84afb78d`. Le lot ne modifie aucune preuve V6 ; il formalise une méthode de contrôle dans une branche indépendante.

## Méthode obligatoire

La procédure doit d’abord exécuter `git rev-list --reverse BASE..HEAD`. Si la liste est vide, le résultat est `INVALID_ZERO_COMMIT_RANGE`, jamais `PASS`. Pour chaque commit de la liste, elle doit résoudre `git rev-parse COMMIT^{tree}`, écrire une ligne d’inventaire avec commit et arbre, exporter précisément cet arbre par `git archive COMMIT`, puis exécuter Gitleaks en mode filesystem (`--no-git --source TREE`) avec le fichier `.gitleaks.toml` versionné explicitement passé par `--config`. Le JSON et le code de sortie sont sauvegardés sous un nom dérivé du commit. La synthèse compte les arbres réellement traités et les échecs ; elle ne déduit pas un PASS à partir d’un code zéro sur une liste vide.

## Preuve disponible

Pour la plage V6 de contenu `b34fa5c02ff20144abfb5d240db1c67ad1f038f9..fc080456711dd7f2266911aaec55041fdb1b424c`, l’inventaire comprend quatre commits : `638a39f6a97e908f1b37fdffc1eecb59a3ee77b3`, `d3bb8940a462f66cc816ba0a3eec64468f69a5fc`, `e2fad936d49d35881760f1b15d72ee06802f7bd0` et `fc080456711dd7f2266911aaec55041fdb1b424c`. Les quatre arbres ont été traités individuellement et les JSON correspondants sont archivés dans le sous-dossier de preuve. Les résultats zéro de ces quatre arbres sont des contrôles individuels ; ils ne valent pas qualification de l’historique entier si une autre plage retourne zéro commit ou n’est pas inventoriée.

Le fichier de configuration exact est la politique Gitleaks versionnée du dépôt. Son allowlist est limitée au fingerprint OpenPGP public documenté et à ses chemins de fixture ; elle ne transforme pas un scan vide en succès et ne constitue pas une exclusion générale de l’historique.

Le lot ne lance aucun runtime, ne démarre pas T28/T29/T39/T40/T41/T42, Camoufox, proxy réel, cookies, données utilisateur ou SystemVault natif, et ne constitue pas une release.

**Owner :** sécurité dépôt / responsable des preuves. **Condition de levée :** toute plage annoncée possède un inventaire non vide, un arbre et un JSON par commit, une configuration hachée et une synthèse cohérente.
