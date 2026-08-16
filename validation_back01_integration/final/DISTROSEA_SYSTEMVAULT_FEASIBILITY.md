# Faisabilité DistroSea pour le gate SystemVault — constat initial

## Observations recueillies

La page Ubuntu 24.04 de DistroSea présente un environnement graphique distant servi via noVNC et placé en file d’attente. Le service se décrit publiquement comme un moyen de tester des distributions Linux dans le navigateur, sans installation locale. Les résultats de recherche disponibles décrivent également les sessions comme temporaires et signalent que la connectivité réseau de la machine invitée peut être limitée ou absente.

| Critère du gate | État DistroSea connu | Conséquence |
|---|---|---|
| Bureau graphique | Oui, via noVNC | Peut permettre une vérification visuelle et l’ouverture d’une session utilisateur. |
| Session utilisateur locale réelle | Incertain | À vérifier dans la machine invitée. |
| Persistance entre redémarrages | Incertaine / sessions annoncées temporaires | Risque majeur pour le cas de lecture après redémarrage du Core et pour l’archivage de preuves. |
| Coffre Secret Service déverrouillé | Inconnu | Doit être vérifié avec `systemvault-doctor`; aucun fallback en clair n’est acceptable. |
| Droit d’installer ou d’exécuter les outils nécessaires | Inconnu | Les scripts doivent fonctionner sans `sudo`; l’absence d’outillage est un blocage, pas une dérogation. |
| Réseau de la machine invitée | Potentiellement limité | Ne pas dépendre du téléchargement ou de GitHub pendant l’exécution. |
| Isolation réelle, hors conteneur | Inconnue | Doit être attestée par les commandes du runbook; un conteneur invalide le gate. |

## Décision provisoire

DistroSea peut être utilisé **uniquement comme candidat de qualification** si la session obtenue prouve sur place les prérequis du runbook : Ubuntu 24.04 amd64, session graphique utilisateur non root, Secret Service natif déverrouillé, absence de conteneur, exécution possible sans `sudo`, et export de preuves assainies.

Elle ne doit pas être utilisée pour importer un secret réel, une clé privée, un token, un profil utilisateur réel, un proxy privé ou un compte GitHub. Si un seul prérequis échoue, le résultat est `PENDING` ou `FAILED` et la cible DistroSea ne peut pas lever `SYSTEMVAULT_NATIVE_PER_TARGET`.

## Sources consultées

- Page DistroSea Ubuntu 24.04 : https://distrosea.com/fr/start/ubuntu-24.04-default/
- Résultats de recherche publics sur la nature temporaire et les limitations de session : consultés le 2026-08-14.

## État de session observé le 2026-08-14

La page DistroSea indique que le démarrage d’Ubuntu est en attente, avec sept utilisateurs devant la session et une estimation d’une minute. Elle mentionne également que l’ouverture d’une session sert à accéder à Internet dans les systèmes d’exploitation. Aucun bureau ni terminal invité n’était encore disponible lors du constat.

La présence d’une file d’attente ne constitue ni une preuve de persistance ni une preuve qu’un coffre Secret Service est utilisable. Dès l’ouverture du bureau, les prérequis du runbook devront être vérifiés avant toute exécution : utilisateur non root, session graphique, absence de conteneur, backend natif réellement disponible et aucune valeur secrète réelle.

Source : https://distrosea.com/fr/start/ubuntu-24.04-default/ (consultée le 2026-08-14).

## Alternative évaluée : VM Ubuntu Desktop sur Azure

Une offre Azure Marketplace « Virtual Desktop on Ubuntu 24.04 » décrit une machine virtuelle Ubuntu 24.04 avec bureau GNOME complet et accès RDP. L’offre déclare inclure notamment GNOME, un terminal et un accès RDP ; cela peut constituer une cible de qualification préférable à une démo navigateur partagée, sous réserve de contrôler le compte utilisateur, D-Bus de session, Secret Service, persistance, version exacte et absence de conteneur avant de déclarer un résultat.

Source : https://marketplace.microsoft.com/en-us/product/cloud-infrastructure-services.gnome-ubuntu-desktop?tab=Overview (consultée le 2026-08-14).

## Alternatives cloud consultées sans PC local

- AWS Marketplace publie une AMI x86 64 bits « Ubuntu Desktop 24.04 - Web and RDP », avec Ubuntu 24.04, GNOME, accès via RDP ou web, un essai annoncé de cinq jours et une facturation à l’usage après essai. Il s’agit d’une véritable image de machine virtuelle, non d’un émulateur navigateur, mais la description est fournie par le vendeur et chaque prérequis SystemVault doit être contrôlé dans la session avant qualification.
  Source : https://aws.amazon.com/marketplace/pp/prodview-ovus3zg6v3wc6 (consultée le 2026-08-14).
- Azure Marketplace publie une image « Browser based Ubuntu 24.04 GUI Desktop accessible via HTTPS » de TechLatest, annoncée comme machine virtuelle avec bureau graphique accessible depuis un navigateur. Elle est à traiter comme une option de test : le besoin de session utilisateur, Secret Service, persistance et absence de conteneur demeure à vérifier, jamais à présumer.
  Source : https://marketplace.microsoft.com/it-it/marketplace/apps/techlatest.browser-based-ubuntu-2404?tab=overview (consultée le 2026-08-14).

Conséquence : pour un utilisateur sans PC local, une VM cloud persistante Ubuntu 24.04 amd64 accessible depuis le navigateur constitue le chemin réaliste. Les démos gratuites non persistantes restent utiles pour un smoke test, mais insuffisantes pour lever le gate SystemVault sans les contrôles du runbook.

## Recherche d’alternatives gratuites sans PC local — 2026-08-14

| Service | Offre gratuite déclarée | Compatibilité avec le gate Ubuntu 24.04 |
|---|---|---|
| OnWorks | Postes Linux gratuits accessibles dans un navigateur ; la sélection Ubuntu publiée expose notamment Ubuntu 20 | Non retenu : la version publiquement identifiée est Ubuntu 20.04, donc pas la cible Ubuntu 24.04 amd64 annoncée. |
| KodeKloud Playgrounds | Environnement Ubuntu 22.04 gratuit à durée limitée (une heure), décrit comme accès en ligne de commande | Non retenu : mauvais OS et absence de bureau graphique/Secret Service démontrable. |
| DistroSea | Démonstration Ubuntu dans le navigateur, session temporaire | Candidat de smoke test uniquement ; persistance et Secret Service restent non établis. |

Conclusion provisoire : aucun des services gratuits vérifiés n’est une base suffisante pour lever le gate `SYSTEMVAULT_NATIVE_PER_TARGET` du candidat RC Ubuntu 24.04. Les environnements gratuits restent utiles seulement pour vérifier le comportement de refus du script et préparer la procédure, jamais pour déclarer le gate vert.

Sources :
- https://www.onworks.net/ (consultée le 2026-08-14)
- https://www.onworks.net/os-distributions/ubuntu-based/free-ubuntu-online-version-20 (consultée le 2026-08-14)
- https://kodekloud.com/playgrounds/playground-ubuntu-22-04 (consultée le 2026-08-14)

## Options gratuites supplémentaires examinées — 2026-08-14

| Service | Gratuité déclarée | Conclusion pour SystemVault |
|---|---|---|
| Oracle Cloud Free Tier | Services « Always Free » disponibles sans limite de temps ; inscription nécessitant une carte bancaire ou de débit pour vérification d’identité, avec blocage d’autorisation temporaire possible | Meilleure option gratuite potentielle : une instance de calcul persistante peut être une vraie VM hors conteneur, mais Ubuntu 24.04, bureau graphique, D-Bus et Secret Service déverrouillé doivent être installés et prouvés avant qualification. Aucune ressource payante ne doit être sélectionnée. |
| GitHub Codespaces | Quota gratuit mensuel pour comptes personnels ; utilisable depuis navigateur | Incompatible avec le gate : la documentation officielle indique que l’utilisateur est placé dans un conteneur Docker. Le runbook doit le refuser. Utile uniquement pour analyses de code non natives. |

Sources :
- https://www.oracle.com/cloud/free/ (consultée le 2026-08-14)
- https://docs.github.com/codespaces/overview (consultée le 2026-08-14)
- https://github.com/features/codespaces (consultée le 2026-08-14)

## Recommandation gratuite retenue

La seule voie gratuite ayant une chance raisonnable de satisfaire les invariants du gate est une VM persistante Oracle Cloud Always Free **AMD** (`VM.Standard.E2.1.Micro`), car la documentation Oracle la décrit comme une VM à processeur AMD, donc compatible avec la cible amd64. Les instances Arm A1 ne sont pas utilisables pour valider la cible amd64, même si elles disposent de davantage de mémoire gratuite. L’offre AMD est très contrainte ; la capacité est de type micro et peut être insuffisante pour GNOME complet. Si l’installation ne fournit pas une session graphique Secret Service stable, ce n’est pas un motif de dérogation : le gate reste bloqué.

Le script exige précisément un utilisateur non root, l’absence de conteneur, `DBUS_SESSION_BUS_ADDRESS` et `XDG_RUNTIME_DIR`, puis une matrice booléenne complète. Une instance Oracle n’est admissible qu’après satisfaction de ces contrôles dans la session graphique réelle. Oracle prévient en outre que la capacité Always Free peut manquer dans une région et que les instances durablement inactives peuvent être récupérées ; la machine doit donc être détruite immédiatement après la qualification, une fois les sorties assainies récupérées.

Sources :
- https://docs.oracle.com/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm (consultée le 2026-08-14)
- https://www.oracle.com/cloud/free/ (consultée le 2026-08-14)
- `scripts/run-systemvault-native-gate.sh` du dépôt ForgeLocal, lu le 2026-08-14.
