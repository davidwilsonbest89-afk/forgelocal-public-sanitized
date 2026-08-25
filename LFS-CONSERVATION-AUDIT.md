# LFS-CONSERVATION-AUDIT — lot séparé

**Base gelée :** `t00-t42-v6-local-qualified-2026-08-25` / `999374d99b7996504ba91e421850a2fe84afb78d`  
**Branche :** `audit/t00-t42-lfs-conservation-audit`  
**Décision :** `LFS_CONSERVATION_OPEN_PENDING_CANONICAL_RECOVERY`

La baseline V6 a signalé 14 objets LFS historiques absents du stockage accessible. La matrice individuelle est dans `V6_LFS_MATRIX.md`. Le contrôle courant `git lfs fsck` retourne 1 et montre 12 objets toujours indisponibles ; deux objets de la liste initiale ont été récupérés par fetch ciblé : le wrapper V3 historique et le wrapper V4 historique. La différence 14 → 12 est donc documentée, non ignorée.

Aucun des 14 objets n’est déclaré non critique sans preuve. Les archives de continuation T00–T27, T27-R1/CR01 et PREHUMAN portent des éléments de preuve historiques nécessaires à la traçabilité T00–T42. Pour les 12 objets encore absents, aucune copie canonique byte-identique n’a été vérifiée dans le remote accessible. Les artefacts V5/V6 publiés peuvent remplacer certaines fonctions de preuve mais ne sont pas déclarés byte-identiques aux objets manquants.

Les tentatives de récupération ont été limitées à des fetch/pulls LFS ciblés, sans `git lfs pull` global. Les quatre artefacts critiques de livraison V5/V6 déjà récupérés et contrôlés séparément sont : wrapper V3 historique inclus, wrapper V4 historique inclus, bundle V5 delta et wrapper V5/V6 de livraison selon leurs hashes documentés dans le manifeste V6. La vérification de ce lot ne réhydrate aucun objet historique absent.

Le propriétaire de récupération est l’administrateur du dépôt / responsable des preuves release. La condition de levée est la restauration des 12 OID encore absents, ou une preuve canonique distante et byte-identique pour chacun, suivie d’un `git lfs fsck` complet à code 0. Les gates restent inchangées et le gel V6 n’est pas modifié.
