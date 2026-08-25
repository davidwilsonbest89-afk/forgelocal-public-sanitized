# POST-V6 — protocoles de validations externes

**Statut :** `PROTOCOLS_PREPARED_NOT_EXECUTED`

Ce document prépare quatre protocoles indépendants. Il ne constitue pas une autorisation d’exécution, une approbation produit ou une release. Aucun protocole ne doit commencer sans une autorisation écrite séparée qui nomme exactement le protocole, l’environnement, les données, l’opérateur, la durée, le rollback et les gates soumises à revue.

## Autorisation commune obligatoire

L’autorisation doit contenir : `protocol_id`, propriétaire produit, propriétaire sécurité, environnement isolé, versions exactes, données synthétiques autorisées, absence de comptes ou cookies réels, durée maximale, chemin de logs redacted, procédure d’arrêt, procédure de rollback, critères de succès/échec et gates candidates. Une autorisation globale ou implicite ne suffit pas ; les quatre protocoles doivent être autorisés séparément.

| Règle commune | Exigence |
|---|---|
| Données | fixtures synthétiques et identifiants de test non réutilisables ; aucune donnée utilisateur réelle |
| Secrets | valeurs éphémères, non journalisées, redaction vérifiée avant archivage |
| Réseau | loopback ou réseau de test explicitement isolé ; aucun endpoint de production |
| Rollback | arrêt, suppression des artefacts temporaires, restauration du snapshot et preuve d’absence de résidu |
| Logs | événements structurés redacted ; jamais de secret, cookie, valeur de credential ou payload sensible |
| Gate | aucune gate levée automatiquement ; seule la revue autorisée peut proposer une transition |
| Arrêt immédiat | fuite, donnée non synthétique, endpoint inattendu, comportement non déterministe ou rollback incomplet |

## P-C1 — SystemVault natif

**Objectif.** Vérifier l’intégration native au coffre système uniquement si T29/T39 ont reçu une décision produit et une autorisation dédiée. Le protocole ne doit pas être exécuté dans le sandbox actuel.

| Domaine | Protocole |
|---|---|
| Environnement | VM dédiée hors production, OS/version explicitement approuvés, compte de test isolé, SystemVault natif disponible, réseau coupé sauf dépendances locales nécessaires, build identifié par commit |
| Données de test | deux secrets synthétiques générés pour le run (`vault-test-A`, `vault-test-B`), métadonnées fictives, aucun mot de passe réel ; valeurs détruites en fin de run |
| Séquence | vérifier la disponibilité du backend ; créer une entrée de test ; lire dans le scope autorisé ; mettre à jour ; demander une confirmation ; supprimer ; vérifier l’absence après suppression ; tester refus hors scope |
| Redaction/logs | journaliser uniquement opération, identifiant opaque, scope, résultat, code d’erreur et correlation ID ; ne jamais écrire les valeurs, clés, dumps ou exports |
| Rollback | supprimer les entrées synthétiques, effacer le workspace et caches, révoquer la clé/session de test, restaurer le snapshot VM si l’état natif diverge, vérifier zéro résidu |
| Succès | backend natif détecté ; chiffrement et scope conformes à la décision produit ; lecture/écriture/rotation/suppression confirmées ; refus hors scope ; aucun secret dans logs/mémoire persistée observable ; rollback zéro résidu |
| Échec | backend indisponible ; fallback silencieux ; secret dans logs/UI/artefact ; scope élargi ; écriture partielle ; suppression non vérifiable ; rollback incomplet |
| Gate candidate | `NATIVE_SYSTEMVAULT_NOT_TESTED` peut seulement être proposée à revue pour T29/T39 ; aucune levée de `PUBLIC_RELEASE_BLOCKED` ou `release_authorized` par ce protocole |
| Owner | Platform Security + Product, avec revue indépendante |

**Interdiction.** Ne pas contourner SystemVault par un fichier plaintext, une variable persistante ou un store utilisateur. Un échec doit rester `BLOCKED`, jamais `NOT_APPLICABLE`.

## P-C2 — Camoufox / runtime ciblé

**Objectif.** Vérifier un runtime ciblé uniquement après autorisation distincte. Il ne doit pas être confondu avec les E2E T10/T15 loopback, qui ont utilisé des fixtures synthétiques et Chromium système local.

| Domaine | Protocole |
|---|---|
| Environnement | VM ou runner dédié, runtime et binaire hashés, profil éphémère, réseau deny-by-default, aucun compte utilisateur, aucun proxy réel, aucun cookie réel, aucun endpoint production |
| Données de test | page locale synthétique `file://` et serveur loopback de test ; paramètres fictifs ; aucun contenu importé depuis un utilisateur |
| Séquence | vérifier signature/digest du runtime ; démarrer avec profil vierge ; exécuter navigation locale, content redacted, screenshot hashée et fermeture ; tester refus des URL externes ; vérifier absence d’ouverture réseau non autorisée |
| Redaction/logs | conserver version, hash binaire, événements, URL synthétiques, codes de refus et digests ; supprimer stdout contenant payload, chemin sensible, token ou cookie |
| Rollback | fermer le runtime et descendants ; supprimer profil/cache/temp ; restaurer snapshot runner ; vérifier ports/processus fermés et absence de fichiers persistants |
| Succès | binaire autorisé et hash conforme ; local-only respecté ; fixture synthétique traitée ; fermeture propre ; aucun réseau externe, secret, cookie ou résidu ; logs redacted complets |
| Échec | binaire non vérifié ; exécution Camoufox sans autorisation ; sortie réseau ; profil persistant ; fuite de contenu ; processus zombie ; rollback incomplet |
| Gate candidate | `camoflox_execution_authorized=false` peut seulement être soumise à revue pour le runtime ciblé ; aucune modification de T08, release ou gate produit par ce protocole |
| Owner | Runtime Security + Product, avec revue indépendante |

**Interdiction.** Ce protocole ne doit pas être lancé par défaut, ne doit pas télécharger un runtime non approuvé et ne doit pas utiliser de session de navigateur de l’utilisateur.

## P-C3 — proxy de test et cookies de test

**Objectif.** Vérifier le chemin de proxy et de cookie uniquement dans un réseau de test isolé, avec doubles synthétiques, jamais contre un fournisseur réel ou une session utilisateur.

| Domaine | Protocole |
|---|---|
| Environnement | réseau local de test avec mock proxy contrôlé, serveur HTTP synthétique, certificats de test si nécessaire, egress externe bloqué, compte opérateur de test |
| Données de test | proxy fictif `198.51.100.10:8080` ou mock local, cookie `forge_test_cookie=synthetic-only`, valeur aléatoire à usage unique ; aucun domaine réel |
| Séquence | créer le proxy de test ; vérifier validation type/host/port ; envoyer une requête vers la fixture locale via le mock ; vérifier redaction du cookie ; tester port invalide, host externe et cookie expiré ; supprimer le proxy et le cookie |
| Redaction/logs | conserver type, host de documentation, port, code de refus, digest du cookie et correlation ID ; ne jamais conserver la valeur du cookie, header Authorization, payload ou trafic brut |
| Rollback | supprimer l’entrée proxy, purger le cookie et le profil ; fermer le mock ; nettoyer certificats et règles réseau ; vérifier aucun port et aucun fichier temporaire |
| Succès | seules les destinations de test sont acceptées ; refus fail-closed des destinations externes ; cookie synthétique jamais exposé ; opérations et erreurs corrélées ; rollback complet |
| Échec | egress externe ; cookie brut dans log/UI ; proxy réel contacté ; acceptation d’un port/host invalide ; persistance inattendue ; rollback incomplet |
| Gate candidate | toute évolution T10/T39/T40 doit être revue séparément ; ce protocole ne lève ni `PUBLIC_RELEASE_BLOCKED`, ni `release_authorized`, ni l’autorisation Camoufox |
| Owner | Security Networking + Product, avec revue indépendante |

**Interdiction.** Les valeurs RFC 5737 et le nom `synthetic-only` ne sont pas une autorisation de contact réseau ; ils représentent uniquement les fixtures attendues.

## P-C4 — release

**Objectif.** Vérifier le processus de release sans publier ni distribuer tant que toutes les gates ne sont pas formellement approuvées.

| Domaine | Protocole |
|---|---|
| Environnement | runner CI de pré-release isolé, dépôt miroir ou tag candidat immuable, credentials de publication absents par défaut, stockage d’artefacts de test séparé de la distribution |
| Données de test | artefacts synthétiques construits depuis un commit explicitement approuvé, SBOM de test, notices/licences approuvées, checksums et signatures de test ; aucun artefact utilisateur ou secret de production |
| Séquence | vérifier commit/tag ; exécuter build reproductible ; tests unit/integration/E2E autorisés ; scans sécurité/SBOM/licences/LFS ; générer bundle/ZIP/sidecars/manifeste ; vérifier extraction et clone ; effectuer un dry-run de publication sans endpoint de distribution |
| Redaction/logs | conserver versions, commits, hashes, codes de sortie, verdicts et signatures de test ; exclure credentials, tokens de registry, payloads et données personnelles |
| Rollback | supprimer les artefacts de staging, révoquer les credentials temporaires, supprimer le tag candidat si la politique l’exige, restaurer le snapshot CI, vérifier absence de publication et de release visible |
| Succès | toutes les gates candidates approuvées par leurs owners ; tests et scans sans exception non approuvée ; artefacts reproductibles ; signature et provenance vérifiées ; dry-run sans publication ; rollback testé |
| Échec | gate inconnue/non approuvée ; SBOM ou licence UNKNOWN non traitée ; test manquant ; hash divergent ; credential exposé ; publication involontaire ; rollback incomplet |
| Gate candidate | `PUBLIC_RELEASE_BLOCKED` et `release_authorized=false` ne peuvent être proposés à transition qu’après approbations formelles T41/T42 et revue indépendante ; aucune publication dans ce protocole préparatoire |
| Owner | Release governance + Security + Product, avec revue indépendante |

**Interdiction.** Aucun tag de release, upload, distribution, déploiement ou runtime de production ne doit être effectué dans le présent lot.

## Modèle d’autorisation séparée

```text
protocol_id: P-C1 | P-C2 | P-C3 | P-C4
product_owner:
security_owner:
independent_reviewer:
environment:
exact_commit_or_build:
synthetic_data_scope:
real_accounts_cookies_proxy_data: FORBIDDEN
start_utc:
stop_utc:
rollback_owner:
redacted_log_path:
success_criteria_ack:
failure_stop_criteria_ack:
gates_submitted_for_review:
written_authorization:
date:
```

**Statut final du document :** quatre protocoles préparés, aucun exécuté ; chaque protocole attend son autorisation écrite séparée.
