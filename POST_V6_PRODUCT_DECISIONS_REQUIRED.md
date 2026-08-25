# POST_V6_PRODUCT_DECISIONS_REQUIRED

**Statut :** `DECISION_REQUIRED_NO_PRODUCT_CODE_AUTHORIZED`

**Baseline conservée :** V6 gelé au tag `t00-t42-v6-local-qualified-2026-08-25`, commit `999374d99b7996504ba91e421850a2fe84afb78d`. Ce document est volontairement décisionnel : il ne choisit pas à la place du produit et n’autorise aucune implémentation T28, T29, T39, T40, T41 ou T42.

## Règles de réponse

Pour chaque ligne, le propriétaire produit doit fournir une valeur unique, une justification courte, les références de sécurité/UX nécessaires et une date d’approbation. Les valeurs indiquées dans la colonne « Choix fermés acceptables » définissent le contrat à choisir ; une réponse libre ou ambiguë ne constitue pas une décision exploitable. Toute décision doit préciser si elle s’applique au MVP seulement ou à la cible ultérieure.

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| G-01 | Autoriser le code produit post-V6 | `AUCUN_CODE`, `LOT_FUTUR_ISOLE_APRES_APPROBATION` | `DECISION_REQUIRED` | Product / Security |
| G-02 | Maintenir les gates pendant les décisions | `MAINTENIR_TOUTES_LES_GATES`, `MODIFICATION_EXPLICITE_PAR_GATE` | `MAINTENIR_TOUTES_LES_GATES` par défaut | Product governance |
| G-03 | Niveau de données accepté pour les essais futurs | `SYNTHETIQUE_SEULEMENT`, `TEST_CONTROLLE_AVEC_APPROBATION`, `DONNEE_REELLE_INTERDITE` | `DECISION_REQUIRED` | Security / Privacy |

## T28 — Extensions

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| T28-01 | Source des extensions | `REGISTRE_INTERNE_SIGNE`, `REGISTRE_PUBLIC_APPROUVE`, `AUCUNE_SOURCE_MVP` | `DECISION_REQUIRED` | Product / Security |
| T28-02 | Provenance minimale | `SIGNATURE_ED25519_ET_DIGEST`, `SIGNATURE_PKIX_ET_DIGEST`, `DIGEST_SEUL_AVEC_APPROBATION`, `REFUS_MVP` | `DECISION_REQUIRED` | Security |
| T28-03 | Allowlist | `ALLOWLIST_EXPLICITE_PAR_ID_VERSION`, `ALLOWLIST_PAR_EDITEUR_CERTIFIE`, `AUCUNE_EXTENSION_MVP` | `DECISION_REQUIRED` | Product / Security |
| T28-04 | Permissions d’extension | `CAPABILITES_DECLARATIVES_ET_MINIMALES`, `SANDBOX_PAR_EXTENSION`, `REFUS_DES_PERMISSIONS_PRIVILEGIEES`, `AUCUNE_EXTENSION_MVP` | `DECISION_REQUIRED` | Security / Architecture |
| T28-05 | Stockage | `STORE_CHIFFRE_LOCAL`, `STORE_VERSIONNE_SEPARE`, `AUCUN_STOCKAGE_MVP` | `DECISION_REQUIRED` | Architecture |
| T28-06 | Lifecycle | `INSTALL_UPDATE_DISABLE_REMOVE`, `VERSIONS_IMMUABLES_ET_APPROUVEES`, `AUCUN_LIFECYCLE_MVP` | `DECISION_REQUIRED` | Product |
| T28-07 | Rollback | `ROLLBACK_VERSION_PRECEDENTE_VERIFIEE`, `ROLLBACK_SNAPSHOT_ATOMIQUE`, `AUCUN_ROLLBACK_MVP_DONC_REFUS_UPDATE` | `DECISION_REQUIRED` | Reliability |
| T28-08 | Politique de refus | `REFUS_FAIL_CLOSED`, `REFUS_ET_MESSAGE_ACTIONNABLE`, `REFUS_SILENCIEUX_INTERDIT`, `AUCUNE_EXTENSION_MVP` | `DECISION_REQUIRED` | Product / Security |
| T28-09 | Audit | `EVENEMENTS_REDACTED_LOCAUX`, `EVENEMENTS_SIGNES_EXPORTABLES`, `AUCUN_AUDIT_MVP_DONC_REFUS` | `DECISION_REQUIRED` | Security / Legal |

**Conséquence avant décision T28 :** aucune API, UI, stockage, chargeur ou migration d’extension ne doit être codé. Une extension non approuvée doit être refusée, pas traitée comme `NOT_APPLICABLE`.

## T29 — Password Manager

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| T29-01 | Types de secrets autorisés | `AUCUN_SECRET_MVP`, `IDENTIFIANTS_DE_TEST_SYNTHETIQUES`, `SECRETS_UTILISATEUR_APPROUVES` | `DECISION_REQUIRED` | Product / Privacy |
| T29-02 | Repository | `STORE_LOCAL_CHIFFRE`, `KEYCHAIN_SYSTEME`, `VAULT_DISTANT_APPROUVE`, `AUCUN_REPOSITORY_MVP` | `DECISION_REQUIRED` | Architecture / Security |
| T29-03 | Chiffrement au repos | `AEAD_AVEC_CLE_SYSTEMVAULT`, `AEAD_AVEC_CLE_DERIVEE`, `NON_APPLICABLE_CAR_REFUS_MVP` | `DECISION_REQUIRED` | Security |
| T29-04 | SystemVault | `OBLIGATOIRE_AVEC_IMPLEMENTATION_NATIVE_APPROUVEE`, `ABSTRACTION_PORTABLE_APPROUVEE`, `INTERDIT_MVP` | `DECISION_REQUIRED` | Security / Platform |
| T29-05 | Lecture | `LECTURE_EXPLICITEMENT_CONFIRMEE`, `LECTURE_AUTORISEE_PAR_SCOPE`, `AUCUNE_LECTURE_MVP` | `DECISION_REQUIRED` | Product / Security |
| T29-06 | Écriture | `ECRITURE_EXPLICITEMENT_CONFIRMEE`, `ECRITURE_AUTORISEE_PAR_SCOPE`, `AUCUNE_ECRITURE_MVP` | `DECISION_REQUIRED` | Product / Security |
| T29-07 | Export | `INTERDIT`, `EXPORT_CHIFFRE_AVEC_CONFIRMATION`, `EXPORT_REDACTED_METADATA_ONLY` | `DECISION_REQUIRED` | Product / Legal |
| T29-08 | Redaction | `VALEURS_JAMAIS_LOGGEES_ET_REDACTION_STRUCTURELLE`, `METADONNEES_SEULEMENT`, `AUCUNE_FONCTION_MVP` | `DECISION_REQUIRED` | Security / Privacy |
| T29-09 | Rotation / suppression | `ROTATION_ET_SUPPRESSION_AUDITEES`, `SUPPRESSION_SEULE_AUDITEE`, `AUCUNE_GESTION_MVP` | `DECISION_REQUIRED` | Security |

**Conséquence avant décision T29 :** aucune lecture/écriture de secret, intégration SystemVault, import/export ou migration ne doit être codée. T29 reste `BLOCKED` et `NATIVE_SYSTEMVAULT_NOT_TESTED`.

## T39 — Import / export

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| T39-01 | Formats d’import | `AUCUN_IMPORT_MVP`, `JSON_SCHEMA_VERSIONNE`, `CSV_REDACTED_SEULEMENT`, `FORMAT_PASSWORD_MANAGER_APPROUVE` | `DECISION_REQUIRED` | Product / Security |
| T39-02 | Formats d’export | `AUCUN_EXPORT_MVP`, `JSON_CHIFFRE_VERSIONNE`, `ARCHIVE_CHIFFREE_VERSIONNEE`, `METADATA_REDACTED_ONLY` | `DECISION_REQUIRED` | Product / Legal |
| T39-03 | Chiffrement | `AEAD_AVEC_SYSTEMVAULT`, `AEAD_AVEC_CLE_DERIVEE_ET_CONFIRMATION`, `EXPORT_INTERDIT` | `DECISION_REQUIRED` | Security |
| T39-04 | Confirmation utilisateur | `CONFIRMATION_EXPLICITE_AVANT_LECTURE_ET_ECRITURE`, `CONFIRMATION_AVANT_EXPORT_SEULEMENT`, `AUCUN_FLUX_MVP` | `DECISION_REQUIRED` | Product / UX |
| T39-05 | Rollback | `TRANSACTION_ATOMIQUE`, `SNAPSHOT_AVANT_IMPORT`, `REFUS_IMPORT_SANS_ROLLBACK` | `DECISION_REQUIRED` | Reliability |
| T39-06 | Audit | `AUDIT_REDACTED_LOCAL`, `AUDIT_SIGNE_EXPORTABLE`, `AUCUN_IMPORT_EXPORT_MVP` | `DECISION_REQUIRED` | Security / Legal |
| T39-07 | Stockage intermédiaire | `MEMOIRE_SEULE`, `FICHIER_CHIFFRE_EPHEMERE`, `AUCUN_FLUX_MVP` | `DECISION_REQUIRED` | Security |
| T39-08 | Politique d’erreur | `FAIL_CLOSED_ET_AUCUNE_ECRITURE_PARTIELLE`, `IMPORT_PARTIEL_INTERDIT`, `AUCUN_FLUX_MVP` | `DECISION_REQUIRED` | Product / Reliability |

## T40 — API finale

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| T40-01 | Versioning | `URL_V1_SEMVER`, `HEADER_VERSIONNE`, `VERSION_UNIQUE_MVP` | `DECISION_REQUIRED` | API product |
| T40-02 | Compatibilité | `BACKWARD_COMPATIBLE_UNE_VERSION`, `DEPRECATION_POLICY`, `BREAKING_CHANGE_MAJEURE_SEULEMENT` | `DECISION_REQUIRED` | API / Product |
| T40-03 | Authentification loopback | `BEARER_EPHEMERE_PLUS_ORIGIN_REFERER`, `MUTUAL_AUTH_LOCALE`, `NO_AUTH_INTERDIT` | `DECISION_REQUIRED` | Security |
| T40-04 | Autorité et scopes | `SCOPES_PAR_ENDPOINT`, `ADMIN_EXPLICITE_ET_READONLY_SEPARE`, `AUCUN_ENDPOINT_SENSIBLE_MVP` | `DECISION_REQUIRED` | Security / API |
| T40-05 | Erreurs | `SCHEMA_ERREUR_VERSIONNE_CORRELATION_ID`, `HTTP_STANDARD_ET_CODE_DOMAINE`, `AUCUN_ENDPOINT_SENSIBLE_MVP` | `DECISION_REQUIRED` | API |
| T40-06 | OpenAPI | `OPENAPI_VERSIONNEE_GENERATED_AND_REVIEWED`, `OPENAPI_MANUELLE_REVIEWED`, `AUCUNE_API_FINALE_MVP` | `DECISION_REQUIRED` | API / Security |
| T40-07 | Limites et timeouts | `LIMITES_EXPLICITES_PAR_ENDPOINT`, `TIMEOUTS_ET_TAILLES_VERSIONNES`, `AUCUN_ENDPOINT_SENSIBLE_MVP` | `DECISION_REQUIRED` | Reliability |
| T40-08 | Compatibilité client | `CONTRACT_TESTS_VERSIONNES`, `CLIENT_GENERATION_OPENAPI`, `AUCUN_CLIENT_FOURNI_MVP` | `DECISION_REQUIRED` | API / Dashboard |

## T41 — Dashboard final

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| T41-01 | Parcours principal | `ONBOARDING_ACTIONNEL`, `NAVIGATION_PAR_DOMAINES`, `PARCOURS_MINIMAL_MVP` | `DECISION_REQUIRED` | Product / UX |
| T41-02 | États UI | `LOADING_EMPTY_ERROR_SUCCESS_BLOCKED`, `ETATS_MINIMAUX_DEFINIS_PAR_ECRAN`, `AUCUN_ECRAN_SENSIBLE_MVP` | `DECISION_REQUIRED` | Product / UX |
| T41-03 | Accessibilité | `WCAG_2_2_AA_CIBLE`, `WCAG_2_1_AA_CIBLE`, `ACCESSIBILITE_MINIMALE_ET_PLAN` | `DECISION_REQUIRED` | UX / Accessibility |
| T41-04 | Responsive | `DESKTOP_ET_VIEWPORTS_DEFINIS`, `DESKTOP_SEULEMENT_MVP`, `VIEWPORTS_VERSIONNES` | `DECISION_REQUIRED` | Product / UX |
| T41-05 | Erreurs et refus | `MESSAGE_ACTIONNABLE_SANS_SECRET`, `DETAILS_REDACTED_ET_CORRELATION`, `REFUS_VISIBLES_ET_FAIL_CLOSED` | `DECISION_REQUIRED` | UX / Security |
| T41-06 | E2E | `PLAYWRIGHT_SEQUENTIEL_SYNTHETIQUE`, `MATRICE_VIEWPORTS_ET_AXE`, `AUCUN_E2E_SENSIBLE_MVP` | `DECISION_REQUIRED` | QA / Security |
| T41-07 | Données affichées | `REDACTED_PAR_DEFAUT`, `METADATA_ONLY`, `AUCUNE_DONNEE_SENSIBLE_MVP` | `DECISION_REQUIRED` | Privacy / Product |
| T41-08 | Observabilité | `LOGS_REDACTED_CORRELATION_ID`, `METRIQUES_SANS_PAYLOAD`, `AUCUNE_OBSERVABILITE_SENSIBLE_MVP` | `DECISION_REQUIRED` | Security / SRE |

## T42 — Sortie MVP et release

| ID | Décision attendue | Choix fermés acceptables | Réponse produit | Owner / approbateur |
|---|---|---|---|---|
| T42-01 | Définition de sortie MVP | `FONCTIONS_CORE_ET_DASHBOARD_DEFINIES`, `MVP_DOCUMENTAIRE_SANS_RELEASE`, `AUCUNE_SORTIE_MVP` | `DECISION_REQUIRED` | Product |
| T42-02 | Critères fonctionnels | `MATRICE_ACCEPTEE_PAR_LOT`, `CRITERES_MINIMAUX_ET_LIMITES_EXPLICITES`, `AUCUNE_FONCTIONNALITE_SENSIBLE_MVP` | `DECISION_REQUIRED` | Product / QA |
| T42-03 | Critères sécurité | `GATES_ZERO_ET_EXCEPTIONS_ACCEPTEES`, `REVIEW_SECURITY_FORMELLE`, `AUCUNE_RELEASE_MVP` | `DECISION_REQUIRED` | Security |
| T42-04 | Critères données | `DONNEES_SYNTHETIQUES_SEULEMENT`, `PRIVACY_REVIEW_APPROUVEE`, `AUCUNE_DONNEE_REELLE_MVP` | `DECISION_REQUIRED` | Privacy |
| T42-05 | Critères exploitation | `RUNBOOK_ROLLBACK_MONITORING`, `DEPLOIEMENT_CANARY_APPROUVE`, `AUCUN_RUNTIME_PRODUCTION_MVP` | `DECISION_REQUIRED` | SRE / Platform |
| T42-06 | Matrice de test | `UNIT_INTEGRATION_E2E_SECURITY_SBOM_LICENSE_LFS`, `MATRICE_REDUITE_APPROUVEE`, `AUCUNE_RELEASE_MVP` | `DECISION_REQUIRED` | QA / Security |
| T42-07 | Gates de release | `TOUTES_GATES_ZERO_ET_REVIEW_INDEPENDANTE`, `APPROVALS_FORMELLES_PAR_GATE`, `RELEASE_INTERDITE` | `DECISION_REQUIRED` | Release governance |
| T42-08 | Autorité finale | `PRODUCT_OWNER_ET_SECURITY_SIGNOFF`, `COMITE_RELEASE`, `AUCUNE_RELEASE_MVP` | `DECISION_REQUIRED` | Direction produit |

## Conditions de soumission

La réponse produit doit remplacer chaque `DECISION_REQUIRED` par une valeur unique, un owner et une approbation datée. Après soumission, un nouveau lot séparé pourra produire une baseline brute et un plan d’implémentation limité aux décisions approuvées. Avant cette étape, aucune modification de code T28, T29, T39, T40, T41 ou T42 n’est autorisée.

Les protocoles externes SystemVault, Camoufox/runtime ciblé, proxy/cookies et release restent séparés et nécessitent chacun une autorisation écrite distincte. Les gates globales restent inchangées : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false`.
