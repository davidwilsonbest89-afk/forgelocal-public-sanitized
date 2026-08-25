# ForgeLocal — Dashboard final — rapport opérationnel

## Conclusion

Le lot a remplacé les surfaces Dashboard précédemment classées `NOT_IMPLEMENTED_PLACEHOLDER` par des parcours locaux réellement cliquables et testés : espaces de travail en mémoire de session, journal d’audit local, réglages d’interface, aide embarquée, notifications, filtres avancés, actions de ligne profil et workflows d’extensions T28 reliés aux contrats Core existants. Le lot ne modifie pas le code métier T28 historique et ne démarre pas T29.

Le résultat est **fonctionnel sur le périmètre synthétique local couvert**, mais il ne constitue pas une déclaration production-ready. La signature cryptographique n’est pas exposée par le contrat Core présent ; l’interface affiche honnêtement cette limite et ne la remplace pas par une simulation. Les environnements Camoufox, SystemVault natif et Docker/Buildx restent indisponibles. Les réserves Gitleaks, Gosec et OSV restent ouvertes.

## Provenance et périmètre

| Élément | Valeur |
|---|---|
| Baseline du lot | `a0913cfb98bdf3cc8278a2731890a69ac32423cd` |
| Référence R2 précédente | `1e128bd3b2f1cb9b668afff25c7c155316fd0267` |
| Branche | `validation/operational-v1` |
| Données | Fixtures synthétiques, Core loopback, Chromium système, aucune donnée utilisateur |
| T28 historique | Non modifié, non rouvert |
| T29 | Non démarré |
| T31–T38 | Non modifiés |

## Matrice finale

| Statut | Éléments | Preuve |
|---|---|---|
| `PASS` | Typecheck Dashboard, build Dashboard, build Core, tests Go ciblés, vet, diff-check | `QUALITY_SECURITY_RAW.log`, sortie Playwright combinée |
| `PASS` | 5/5 scénarios Dashboard final ; 6/6 avec le scénario R2 auth expired/revoked | `PLAYWRIGHT_COMBINED_RAW.log` |
| `PASS` | Axe desktop 0, Axe mobile 0, serious/critical 0 | `A11Y_CONSOLE_NETWORK_RAW.log` |
| `PASS` | Erreurs console 0, warnings 0, page errors 0, requêtes échouées 0, mauvaises réponses 0 | `A11Y_CONSOLE_NETWORK_RAW.log` |
| `PASS` | Responsive 390×844 et navigation clavier | `PLAYWRIGHT_COMBINED_RAW.log`, `A11Y_CONSOLE_NETWORK_RAW.log` |
| `PASS` | Actions de ligne archive/réouverture, duplication, export ZIP local | `PLAYWRIGHT_COMBINED_RAW.log` |
| `PASS` | Import synthétique, inspection, allowlist, HIGH_RISK, approbation, affectation, révocation/quarantaine, rollback, purge | `PLAYWRIGHT_COMBINED_RAW.log` |
| `PASS` | Messages et états négatifs HTTP 403, 404, 409, 500 ; 401 expired/revoked couvert par R2 auth | `PLAYWRIGHT_COMBINED_RAW.log`, `PLAYWRIGHT_NEGATIVE_403_500_404_RAW.log` |
| `FAIL` | Gitleaks : 10 findings dans l’arbre complet ; aucune suppression par allowlist ajoutée | `gitleaks-final.json`, `SECURITY_CORRECT_CWD_RAW.log` |
| `FAIL` | Gosec : 128 findings non marqués `nosec` sur `internal/...` | `gosec-final.json`, `SECURITY_CORRECT_CWD_RAW.log` |
| `FAIL` | OSV : exit 1, 46 références advisory affichées, dont avis Go stdlib liés à la directive Go 1.25.0 | `osv-final.txt`, `SECURITY_CORRECT_CWD_RAW.log` |
| `PASS` | govulncheck : exit 0, aucun objet `finding` retourné | `govulncheck-final.json`, `SECURITY_CORRECT_CWD_RAW.log` |
| `PASS` | pnpm audit production : exit 0 | `pnpm-audit-final.json`, `SECURITY_CORRECT_CWD_RAW.log` |
| `PASS` | Trivy filesystem : exit 0, aucune vulnérabilité `VulnerabilityID` dans la sortie JSON | `trivy-fs.json` |
| `PASS` | Syft : SBOM JSON produit | `syft.json` |
| `NOT_IMPLEMENTED_PLACEHOLDER` | Vérification cryptographique de signature/provenance attestée : non exposée par le contrat Core ; seul le digest/intégrité du blob est affiché comme vérifié par le Core | composant T28 et test synthétique |
| `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Camoufox : archive vérifiée non fournie ; Chromium n’est pas une substitution | baseline et preuves runtime R2 |
| `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | SystemVault natif : Secret Service/session keyring absent ; aucun secret natif écrit | preuve SystemVault R2 |
| `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Docker/Buildx : client/daemon indisponibles ; aucun conteneur lancé | preuve Docker R2 |
| `FAIL_CRITICAL` | Aucun FAIL_CRITICAL observé dans les scénarios exécutés | journaux Playwright/Core et cleanup |

## Couverture des surfaces Dashboard

Les espaces, l’audit local, les réglages, l’aide, les notifications et les filtres avancés sont des contrôles de session Dashboard ; ils ne prétendent pas persister dans le Core. Chaque surface est ouverte par clic réel et l’état résultant est vérifié. Les actions de ligne profil utilisent les routes Core existantes pour lifecycle, duplication, suppression et export.

Le panneau T28 réalise les transitions disponibles dans le Core : import local, inspection de manifest/digest/provenance, reconnaissance de permissions, acceptation HIGH_RISK, approbation, affectation, révocation/quarantaine, rollback et purge. Les cas d’erreur sont affichés par un feedback persistant et par notification. Après un 403, le token d’écriture est retiré de la mémoire et le formulaire de reconnexion réapparaît ; ce comportement fail-closed est lui-même testé.

La **signature cryptographique** n’est pas inventée : le panneau indique qu’elle n’est pas exposée par le contrat Core actuel. La signature vérifiable, l’attestation de provenance et la politique de signature/allowlist de distribution restent donc `NOT_IMPLEMENTED_PLACEHOLDER` tant qu’un contrat Core explicite n’est pas fourni.

## Environnements et release

Le test navigateur utilise uniquement Chromium système pour le Dashboard et les fixtures locales. Il ne qualifie pas Camoufox. L’absence de Docker/Buildx et de Secret Service est enregistrée comme `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`, jamais comme PASS. Les gates restent inchangés : `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false`.

> Statut exact : **Dashboard final fonctionnel sur périmètre synthétique local couvert ; ForgeLocal non production-ready.**
