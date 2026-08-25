# CHANGELOG — LFS-CONSERVATION-RESTORATION

## 2026-08-25

Création du lot séparé depuis l’audit LFS publié. La baseline brute reprend les douze OID encore indisponibles. Une tentative initiale de checkout a déclenché un smudge historique et saturé l’espace ; l’incident est conservé comme limitation de procédure. Une séquence contrôlée de douze fetchs filtrés par chemin a ensuite été exécutée, sans pull global. Les douze contenus ont été vérifiés byte-à-byte par taille et SHA-256, et `git lfs fsck` retourne zéro. La conservation durable et la revue indépendante restent sous la responsabilité de l’administrateur du dépôt.
