# T10 — Proxies : document de cadrage (non normatif, préparatoire)

**Projet :** ForgeLocal — Core Go + dashboard React (client du Core)
**Référence :** CDC ForgeLocal, jalon T10 (contrat des proxies)
**Statut :** `T10_FRAMING_ONLY` — **aucun code T10 produit, démarrage interdit avant autorisation explicite** (instruction utilisateur ou du valideur, après la validation unique finale de T09)
**Date :** 16 août 2026

Ce document est un cadrage documentaire non normatif : il ne crée aucun nouveau statut, aucune nouvelle exigence normative, aucun nouveau certificat, et ne modifie ni le registre canonique JSON ni les statuts de release. Il prépare uniquement le contenu du jalon T10 pour que, dès autorisation, le travail démarre dans un périmètre déjà borné.

---

## État du jalon

| Élément | État |
|---|---|
| T09 (Profile Writes) | `T09_APPROVED_VERIFIABLE_LOCAL` — clôturé le 16/08/2026 |
| T07 (Provenance Camoflox) | `T07_PROVENANCE_APPROVED_FOR_SELECTIVE_GO_REIMPLEMENTATION` ; statut T07-R intermédiaire `ATTESTATION_REDACTED_PENDING_SIGNATURE_AND_EXTERNAL_REFERENCES` (références obligations tierces/notices et signature finale en attente dans l'Issue privée `#1`) |
| T08 (Concurrency) | `T08_APPROVED_VERIFIABLE_LOCAL` |
| T10 (Proxies) | **Non démarré** — cadrage seulement |
| Release | `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, pilote suspendu, cinq gates publics en attente |

## Fondations déjà livrées (réutilisables par T10, sans modification)

T10 ne repart pas de zéro : trois fondations de jalons précédents couvrent déjà une partie de son contrat. `Profile.Proxy` existe dans `internal/profile/store.go` (`ProxyConfig` : `socks5`/`http`, `host`, `port`, `username`/`password` sérialisés `json:"-"`, `secret_ref`, `region`), avec un vault local `internal/secrets` et `persistProxySecret`/`loadProxySecret` qui stockent les identifiants uniquement par référence et n'exposent jamais les valeurs en clair. Le PUT T09 (`internal/api/profiles_write.go`) accepte et valide déjà un champ `proxy_config` à la création et à la modification d'un profil. Enfin, l'infrastructure zéro-trust T09 (loopback 403 `LOOPBACK_REQUIRED`, audit redacted `internal/api/audit.go`, `correlation_id`, erreurs machine-readable) s'applique de plein droit aux futures routes proxy.

| Fondation | Chemin | Rôle pour T10 |
|---|---|---|
| Type proxy | `internal/profile/store.go` | Contrat de sérialisation des proxies (valeurs secrètes exclues du JSON) |
| Vault local | `internal/secrets/` | Stockage des identifiants proxy par `secret_ref` (qualification SystemVault = dépendance critique T10) |
| Validation profil | `internal/api/profiles_write.go` | `proxy_config` déjà validé à la création/modification de profil |
| Zéro-trust T09 | `audit.go`, `correlation.go`, `readonly_session.go` | Loopback, audit redacted, corrélation — réutilisables tels quels |

## Périmètre prévu du jalon T10

Le cœur de T10 est un **référentiel proxy indépendant** géré par le Core Go unique : création, modification, retrait et listing de proxies avec validation serveur stricte (type `socks5`/`http`, hôte, port borné, région optionnelle, identifiants jamais renvoyés en clair), affectation profil↔proxy et désaffectation, et application du même contrat zéro-trust que les écritures T09 (boucle uniquement, token d'administration, erreurs machine-readable, `X-Correlation-Id`, audit redacted). Le dashboard devient client de ce contrat via un client mémoire seule (aucune écriture SQLite directe). Une migration SQLite portera les index de performance du référentiel proxy.

Une **dépendance critique** a été identifiée dans le cadrage : la qualification du vault `internal/secrets`. Le CDC exige un chiffrement/verrouillage de type SystemVault ; tant que cette qualification n'est pas démontrée, aucun `secret_ref` réel ne doit être activé pour des identifiants de production — les preuves T10 utiliseront exclusivement des identifiants synthétiques non secrets. T10 est un contrat local de gestion de référentiel : il ne configure, n'utilise, n'intègre et ne lance aucun proxy réseau réel, ne contient aucune intégration de fournisseur (Decodo ou autre) et ne constitue pas un connecteur réseau.

## Critères d'acceptation présumés (à confirmer à l'ouverture)

AC-PROXY-01 : création d'un proxy valide avec validation serveur et erreurs machine-readable ; AC-PROXY-02 : refus des proxies invalides (type inconnu, port hors bornes, hôte invalide, identifiants en clair dans les réponses) ; AC-PROXY-03 : affectation/désaffectation profil↔proxy avec audit redacted ; AC-PROXY-04 : `secret_ref` seul retourné, identifiants jamais exposés (contrat `json:"-"` + tests) ; AC-PROXY-05 : loopback requis sur les mutations, refus 403 hors loopback prouvé E2E ; AC-PROXY-06 : concurrence et isolation sur les allocations profil↔proxy (mutex par profil déjà en place) ; AC-PROXY-07 : dashboard client mémoire seule, build production et `tsc --noEmit` propres, E2E Playwright ; AC-PROXY-08 : suite complète sous `go test -race` sans `DATA RACE`, `go vet`, Gitleaks du delta `[]`, `git diff --check` propre, chemins RC inchangés.

## Prérequis d'ouverture

L'ouverture de T10 exige deux conditions, déjà observées au niveau jalon : la validation unique finale de T09 (donnée : `T09_APPROVED_VERIFIABLE_LOCAL`) et une **autorisation explicite d'ouverture T10** (instruction utilisateur ou du valideur). Même après ouverture, T10 est interdit aux domaines suivants : proxy réseau réel, runtime/Camoufox actif, lancement navigateur, intégration fournisseur, backup/restore/import UI et release. Le lot sera livré comme un rapport unique au format 16 champs à la fin du jalon, sans demande de validation intermédiaire, selon la pratique établie T08/T09.

## Interdictions maintenues

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN` (les findings `validation_back01_integration/` sont antérieurs à T09 et maintenus classés), pilote suspendu, cinq gates publics en attente. Aucun code T10 ne doit être produit avant l'autorisation explicite ; ce document et son commit sont des preuves préparatoires documentaires uniquement.
