# Addendum v2.5 — Surfaces produit lecture seule

**Statut :** contrat préparatoire local ; aucune autorisation de release, de mutation ou de lancement.

## Objectif

Le dashboard ForgeLocal peut progresser au-dessus du Core unique sans créer d’état parallèle. Les nouvelles surfaces sont exclusivement consultatives, utilisent le Bearer de session court émis par le bootstrap local et ne doivent jamais exposer de secrets, chemins hôte, valeurs proxy, identifiants de coffre ou corrélations métier.

## Catalogue de surfaces

| Surface | État | Endpoint cible | Champs affichables | Champs strictement exclus |
|---|---|---|---|---|
| Profils | Implémenté | `GET /api/v1/readonly/profiles` | ID, libellé, groupe, état, tags redacted | proxy brut, chemins, secrets, verrou interne |
| Groupes | Implémenté | `GET /api/v1/readonly/groups` | ID, libellé, nombre de profils | notes privées, chemins, références de coffre |
| Runtimes | Implémenté | `GET /api/v1/readonly/runtimes` | ID, libellé, état de qualification, lançable | ports, endpoint interne, arguments, empreinte de binaire non qualifié |
| Proxys | Contrat préparé | `GET /api/v1/readonly/proxy-providers` | ID, libellé, famille, état de test, horodatage de dernier test | hôte, port, identifiant, mot de passe, `proxy_secret_ref`, URL complète |
| Sauvegardes | Contrat préparé | `GET /api/v1/readonly/backups` | ID, profil, état, date, intégrité booléenne | chemin d’artefact, `key_id`, checksum complet, contenu d’archive |
| Audit | Contrat préparé | `GET /api/v1/readonly/audit-events` | type d’événement, type d’entité, date, résultat redacted | `correlation_id`, `details_json`, identité hôte, données de session |

## Règles de transport communes

Tous les endpoints préparés doivent respecter les règles suivantes avant implémentation :

1. Ils appartiennent au groupe `/api/v1/readonly/` et acceptent uniquement le Bearer principal ou le jeton de session court à portée lecture seule.
2. La pagination est par curseur opaque, avec une limite serveur de `1..100`, et sans offset libre.
3. Les réponses ajoutent `X-Request-ID`, mais ne renvoient jamais un `correlation_id` métier.
4. Les erreurs sont redacted et stables : `invalid_request`, `unauthorized`, `not_found`, `internal_error`.
5. Un endpoint vide renvoie `200` avec `items: []` et un curseur nul ; il ne doit jamais être remplacé par des données de démonstration.
6. Aucun endpoint de cette version ne crée, modifie, supprime, restaure, teste un proxy, isole un profil ou lance un runtime.

## État des parcours dashboard

Le dashboard doit présenter les états **chargement**, **vide**, **erreur avec réessai**, **Core indisponible** et **session expirée**. En dehors d’une session locale valide, il peut conserver la démonstration explicitement signalée mais ne doit jamais la présenter comme une donnée Core.

> **Camoufox reste un candidat non lançable.** Sa présence éventuelle dans la surface Runtimes est informative ; elle ne crée aucune capacité de lancement.

## Conditions de passage à l’implémentation

Chaque endpoint préparé exige, avant ajout au routeur : un DTO redacted, des tests de redaction et pagination, un test Bearer/session court, un test `X-Request-ID`, un test d’absence de mutation et une revue de provenance des dépendances ajoutées. Les actions métier correspondantes restent conditionnées à leurs contrats Core spécifiques et ne sont pas autorisées par cet addendum.

## Hors périmètre et release

Ce document ne modifie ni le candidat RC BACK-01, ni l’archive gelée, ni l’alerte `generic-api-key`, ni SystemVault, ni les cinq gates publics. Le statut demeure `PUBLIC_RELEASE_BLOCKED`.
