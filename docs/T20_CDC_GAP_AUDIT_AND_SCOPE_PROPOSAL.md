# T20-CANDIDATE — CDC Gap Audit & Scope Proposal

**Base figée examinée :** `cc2bfbd274eb584f5d2076d32ee451b7f79b8b51`
**Tag :** `t19rr-harness-evidence-2026-08-19`
**État de preuve de départ :** `T19-RR_CLOSED_VERIFIABLE_LOCAL`
**Conclusion d’audit :** `T20_CANDIDATE_SCOPE_PROPOSAL_PENDING_PRODUCT_AUTHORIZATION`

## Méthode et limites de classification

Les statuts ci-dessous portent sur les preuves réellement distribuées, non sur les intentions du CDC. `PROVEN_LOCAL` signifie qu’un comportement est couvert par une preuve locale identifiable ; `PARTIAL` signifie qu’une fondation existe sans satisfaire tout le contrat CDC ; `NOT_PROVEN` signifie qu’aucune preuve de contrat, code ou test correspondante n’est présente ; `BLOCKED` nécessite une dépendance ou un environnement non disponible ; `NOT_IN_SCOPE` couvre une capacité reportée ou exclue.

| Exigence CDC | État réel | Preuve existante | Dépendance ou risque |
|---|---|---|---|
| Core Go unique, Dashboard mémoire seule, SQLite métier et redaction | `PROVEN_LOCAL` | T12–T19-RR, drill `cc2bfbd` 22/22 | Les gates de production restent séparés. |
| Runtime Chromium local qualifié et automation bornée | `PROVEN_LOCAL` | T14 3/3, T15 5/5, Playwright 22/22 au commit `cc2bfbd` | Aucun lancement Camoufox autorisé. |
| Notes et Custom Fields `text`, `number`, `boolean`, `select` persistés, validés, audités, exportables sans secret — GAP-002 | `NOT_PROVEN` | Le modèle `internal/profile` courant ne contient pas de modèle Notes/Custom Fields ; aucun contrat ou test dédié n’est livré | Migration SQLite, validation serveur, redaction et audit à concevoir. |
| Templates locaux versionnés, validés, clonables sans secret — GAP-003 | `NOT_PROVEN` | Aucun modèle/versioning/template dédié prouvé | Dépend du schéma de métadonnées et du contrat GAP-002. |
| Archive/restore transactionnel — GAP-005 / GAP-024 | `PARTIAL` | Service backup existant et preuves historiques | Doit rester après templates dans l’ordre CDC ; qualification coffre native externe non disponible. |
| ProxyProvider health — GAP-009 | `PARTIAL` | T10 historique, références proxy redacted | Les secrets réels restent dépendants de SystemVault qualifié. |
| SystemVault natif | `BLOCKED` | `NATIVE_SYSTEMVAULT_NOT_TESTED` | Ubuntu avec keyring déverrouillé requis ; gate de production, non lot fonctionnel local. |
| Classification Gosec globale et protection de branche/remote | `BLOCKED` | `SCAN_BLOCKED_UNKNOWN`, kit local sans push distant | Gates de production, non fonctionnalités CDC. |
| Cloud, RBAC, 2FA, Android, Enterprise, marketplace | `NOT_IN_SCOPE` | CDC §0.2 et §15 | P2 ou exclu du MVP local. |
| Camoufox, proxy réel, bypass ou fraude | `NOT_IN_SCOPE` | Limitations et exclusions CDC | Interdits ou non autorisés. |

## Lots candidats classés

| Rang | Lot candidat | Faisabilité locale | Périmètre minimal et exclusions | Qualification exigée | Décision requise |
|---:|---|---|---|---|---|
| 1 | **T20-NCF — Notes & Custom Fields Core Foundation** | Élevée | Métadonnées non sensibles de profil : note et champs `text/number/boolean/select`, migration SQLite, API Core locale, audit redacted et export de métadonnées. **Exclut** templates, clone, UI Dashboard, secrets, import global, backup, proxy réel et runtime. | Tests Go positifs/négatifs, validation type/longueur/valeurs select, migration, persistance, export redacted, audit, API loopback, `-race`, `vet`, build, Gitleaks delta, manifest. | Autorisation explicite de modifier modèle profil, schéma SQLite, API locale et tests Core. |
| 2 | **T21-PT — Profile Templates versionnés** | Moyenne après T20-NCF | Templates locaux versionnés utilisant uniquement les métadonnées déjà supportées ; validation, clone indépendant, migration et audit. **Exclut** secrets, session, stockage navigateur, proxy credentials et runtime. | Tests de versioning, validation, clone non contaminant, migration, audit, redaction et export. | Ne s’ouvre qu’après T20-NCF prouvé. |
| 3 | **G-PROD-01 — Préparation de qualification production** | Faible localement | Inventaire et classification Gosec, plan remote/protection de branche, procédure SystemVault native. **Exclut** modifications fonctionnelles et release. | Rapports de scan bornés, plan de qualification externe et preuves de configuration distante. | Environnement Ubuntu/keyring et dépôt distant canonique requis. |

## Sélection et cadrage de départ

L'ordre obligatoire du CDC commence par **Notes + Custom Fields**, avant les templates. Le lot candidat prioritaire est donc **T20-NCF**. Il est réalisable sur la lignée locale, ne nécessite aucun runtime non qualifié et débloque directement le lot Templates suivant sans traiter prématurément les gates de production.

La présente conclusion d’audit reste `T20_CANDIDATE_SCOPE_PROPOSAL_PENDING_PRODUCT_AUTHORIZATION`. L’instruction utilisateur reçue après l’audit autorise séparément le périmètre T20-NCF ; son cadrage et ses limites sont enregistrés dans `docs/T20_NCF_SCOPE_AND_ACCEPTANCE.md`. Cette autorisation ne s’étend ni aux Templates ni à un autre lot CDC.

## Invariants

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangés. Aucun nouveau lot ne peut présenter sa progression comme une autorisation de release.
