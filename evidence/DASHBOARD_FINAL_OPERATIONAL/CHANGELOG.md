# Changelog — Dashboard final

## Corrections et implémentations

Le Dashboard comporte désormais des panneaux navigables pour les espaces de travail, le journal d’audit de session, les réglages d’interface, l’aide, les notifications et les filtres avancés. Les états locaux sont explicitement indiqués comme mémoire de session lorsqu’ils ne sont pas persistés dans le Core.

Le menu de ligne profil expose des actions reliées aux contrats Core existants : archive/réouverture, duplication, suppression et export local. Le panneau d’extensions expose l’import ZIP synthétique, l’inspection redacted, l’allowlist de permissions, l’acceptation HIGH_RISK, l’approbation, l’affectation, la révocation/quarantaine, le rollback et la purge.

Les erreurs 403, 404, 409 et 500 sont rendues par des messages lisibles et un feedback persistant. Une réponse 403 retire le contrôle d’écriture et fait réapparaître le formulaire de reconnexion ; cette transition fail-closed est couverte. Les tests R2 expired/revoked restent inclus dans la non-régression.

Les tests ajoutent cinq scénarios Dashboard final et sont rejoués avec Chromium système en worker unique. Le run combiné comprend six scénarios passés. Axe est exécuté sur desktop et mobile, avec contrôle de clavier, console, page errors, requêtes échouées et réponses HTTP.

## Non-modifications et limites

Aucun code métier T28 historique n’a été modifié. T29 n’a pas été démarré. T31–T38 restent inchangés. Aucun test Camoufox, SystemVault natif ou Docker n’a été falsifié ou remplacé par une substitution silencieuse.

La vérification cryptographique de signature et l’attestation de provenance ne sont pas prétendues fonctionnelles : le contrat Core actuel ne les expose pas. Gitleaks, Gosec et OSV conservent leurs findings et bloquent la clôture sécurité correspondante.
