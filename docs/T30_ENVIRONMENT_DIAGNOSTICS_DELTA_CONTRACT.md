# T30 — Diagnostics d’environnement : delta read-only

## Décision de périmètre

Le Core contient déjà le diagnostic projeté T13 : métadonnées de profil et qualification du runtime. T30 ne le réimplémente pas. Il rend son niveau de couverture explicite, versionné et déterministe afin qu’aucune surface ne présente des diagnostics non exécutés comme des observations réelles.

## Comportement ajouté

La réponse de diagnostic porte une version et le mode `PROJECTED_METADATA_ONLY`. Le verdict existant demeure fondé sur les deux contrôles réellement disponibles : complétude des métadonnées et liaison à un runtime qualifié. Les catégories qui exigeraient une exécution de navigateur sont présentes en état `UNSUPPORTED`, avec une note fixe redacted.

| Catégorie déclarée | État T30 | Motif |
|---|---|---|
| Navigator, Battery, Storage, Plugins/MIME, Input, Canvas/WebGL/Audio, Fonts, Performance, Permissions | `UNSUPPORTED` | Observation runtime non implémentée et non autorisée dans ce lot. |
| Network | `UNSUPPORTED` | Aucun diagnostic réseau ou proxy réel n’est exécuté. |
| WebRTC | `UNSUPPORTED` | Aucun runtime ni test ICE/STUN/TURN n’est exécuté. |
| Métadonnées de profil et qualification runtime | Projection existante | Données Core déjà disponibles, sans lancement de runtime. |

## Exclusions et gates

T30 ne lance aucun navigateur, Camoufox ou runtime ; il ne contacte aucun réseau ; il ne lit ni secret, cookie, IP, User-Agent, hash Canvas, chemin système, identifiant matériel ou valeur proxy. Il ne modifie aucun store, profil, configuration, migration, gate ou statut de release.

Les diagnostics observés Canvas/WebGL/Audio, WebRTC, géolocalisation, matériel, font rendering et performance restent des lots futurs dépendants d’un runtime qualifié et de leurs contrats propres. `NATIVE_SYSTEMVAULT_NOT_TESTED`, `PUBLIC_RELEASE_BLOCKED` et les autres gates permanents restent inchangés.

## Critères T30

Le format de sortie est stable et machine-readable. Les contrôles non observés sont explicitement déclarés `UNSUPPORTED`. Les valeurs brutes ne sont pas retournées. Les IDs invalides et inconnus restent refusés comme dans T13. Les tests vérifient la version, le mode projeté, le nombre de capacités non supportées et l’absence de note dynamique.
