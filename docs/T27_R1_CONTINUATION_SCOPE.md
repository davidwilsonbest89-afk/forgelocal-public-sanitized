# Périmètre du dossier de continuité T27-R1

## Inclus

Le kit contient le bundle source de la lignée T27-R1, les bundles delta et kits de preuves CR-01 à CR-05, les SBOM/provenance CR-09, `TOOLCHAIN.lock`, le contrat correctif, la politique `BASELINE_DISCOVERY`, le registre canonique, la checklist et un guide de reprise.

## Exclus par sécurité et périmètre

Le kit ne contient aucun token, cookie, `secret_ref` résoluble, fichier `.env`, base SQLite, profil utilisateur, runtime téléchargé, résultat Playwright, `node_modules`, build généré, donnée de navigateur ni attestation privée T07 non nécessaire à la reprise technique.

## Règle de continuité

Chaque lot futur doit vérifier un bundle/sidecar, cloner dans un répertoire neuf, exécuter `git fsck --full`, produire `BASELINE_DISCOVERY_RAW.log` avec commandes, chemins, UTC, codes de sortie et sorties brutes, puis seulement modifier le code. La copie canonique hashée et son inscription au registre précèdent tout nettoyage de répertoire temporaire.
