# GOSEC-R3 changelog

## 2026-08-26

Le package overnight v4 et sa lignée ont été vérifiés depuis un clone neuf avec Go 1.25.13 réel, `GOTOOLCHAIN=local`, sidecars, extraction ZIP/TAR, bundle et `git fsck --full`.

Le lot R3-A a root-scopé les lectures de backup tar via `os.Root.Open`, ajouté une régression symlink externe et fait passer G122 de 1 à 0 ainsi que G304 de 20 à 19.

Le lot R3-B a durci les permissions des groupes, runtime, archives, restore et téléchargements, rendu plusieurs erreurs I/O observables, et fait passer G301 de 17 à 0, G306 de 1 à 0 et G104 de 36 à 27.

Le lot R3-C a remplacé les conversions byte implicites de l’identifiant par `binary.BigEndian` et supprimé le G118 en transmettant le contexte attach borné à la transition durable. G115 est passé de 8 à 4 et G118 de 1 à 0.

Le scan Gosec source-only final reste ouvert à 94 findings. Aucun masquage ou allowlist globale n’a été ajouté. Les outils absents et les gates Camoufox, SystemVault, Docker/Buildx et release restent ouverts.
