# Rapport de remédiation critique — ForgeLocal

**Date d’exécution :** 26 août 2026 (UTC)
**Dépôt :** `davidwilsonbest89-afk/forgelocal-public-sanitized`
**Branche corrective :** `validation/final-secret-remediation`
**Point de départ figé :** `1d0fe24da0edbddc5eab30d971628df8fe5a92d3` sur `validation/final-environment-qualification`

## Conclusion exécutive

Une divulgation critique réelle a été confirmée dans la branche de qualification : la sortie de `docker/entrypoint.sh` interpolait le contenu du token API, et `scripts/start.sh` contenait une seconde divulgation runtime indépendante. La preuve originale a été conservée sous forme **redacted** avec les chemins, lignes, commit source, permissions et hashes ; aucune valeur de token n’a été conservée ou redistribuée.

Le correctif appliqué est volontairement minimal. Les deux lanceurs ne lisent plus le fichier de token et ne l’impriment plus. L’entrypoint conserve uniquement le test d’existence du fichier et affiche un état générique redacted ; le lanceur local supprime entièrement le bloc de lecture et d’affichage. Un binaire de test suivi par Git, classé artefact généré contenant des chaînes de fixture bearer-like, a également été retiré de la branche corrective ; seul son hash historique est conservé dans la preuve.

Le contrôle comportemental avec sentinelle synthétique privée, le contrôle de permissions `0600`, le cas fichier absent, le nettoyage des processus et les scans secrets ciblés sont PASS. La requête authentifiée synthétique sur la route protégée `/api/runtimes` retourne HTTP 200 ; les tokens missing, invalid, expired et revoked retournent HTTP 401 avec les raisons attendues, sans écho de secret. Une installation contrôlée a ensuite réussi : Docker Engine `29.1.3` et Buildx `0.30.1` ont été installés depuis les paquets Ubuntu disponibles. Le daemon réel a démarré via systemd et `sudo docker info` a confirmé le serveur actif ; aucun ajout au groupe Docker n’a été effectué. Le premier build bridge a échoué avec l’erreur noyau/iptables `can't initialize iptables table raw`; un build host-network réel a réussi pour la qualification ponctuelle. Après le cycle, l’image, les caches, les conteneurs, les services et le socket ont été nettoyés ou arrêtés/désactivés. Le chemin de divulgation réel est confirmé, mais aucune valeur n’a été retenue : `REAL_TOKEN_DISCLOSURE_PATH_CONFIRMED`, `SECRET_VALUE_NOT_RETAINED`. Aucun token réel actif n’est démontré dans les éléments conservés. Cette mission n’a effectué aucune rotation, mais l’absence d’usage réel reste à confirmer par l’owner : `NO_REAL_SECRET_ROTATION_PERFORMED=TRUE`, `SECRET_REAL_USE_STATUS=OWNER_CONFIRMATION_REQUIRED`. Le verdict de release reste bloqué par mandat et par les findings restants : `PUBLIC_RELEASE_BLOCKED`, `GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS`, `FORGELOCAL_PRODUCTION_READY=false`.

## Preuve et audit de contamination

La preuve redacted distribuable se trouve dans `evidence/FINAL_SECRET_REMEDIATION/source-entrypoint-redacted.txt`. Son sidecar SHA-256 est fourni séparément. Les deux sources runtime confirmées sont documentées sans contenu secret : `docker/entrypoint.sh` à la ligne 102 dans le commit de départ et `scripts/start.sh` à la ligne 29 dans le même commit. Les fixtures de tests sont classées séparément et ne sont pas traitées comme des logs runtime. `launch.test` est classé comme artefact binaire suivi par Git ; aucune chaîne n’a été extraite ni publiée, et son hash historique seul est conservé.

| Élément audité | Classification | Action corrective ou statut |
|---|---|---|
| `docker/entrypoint.sh:102` | Source runtime critique | Lecture et interpolation supprimées ; test ciblé PASS |
| `scripts/start.sh:29` | Seconde source runtime critique | Bloc de lecture et d’affichage supprimé ; garde statique PASS |
| Fixtures de tests | Test/fixture, non-runtime | Conservées et distinguées dans la matrice anonymisée |
| `launch.test` | Binaire de test suivi, bearer-like fixture possible | Retiré de la branche corrective ; hash seul conservé |
| Logs Docker, couches d’image et `docker logs` | Contrôle runtime | Cycle réel PASS ; sentinelle absente |
| Installation Docker Engine/Buildx | PASS | Engine `29.1.3`, Buildx `0.30.1`; provenance et codes dans les raw logs |
| Premier build Docker bridge | FAIL contrôlé | Table iptables `raw` indisponible ; erreur exacte conservée |
| Build Docker host-network | PASS | Image ForgeLocal construite sans push |
| Historique et layers de l’image | PASS | Sentinelle absente de `docker history` et de l’export de layers |
| Image secret scan Trivy baseline | `TRIVY_IMAGE_SECRET_OPEN` | Clé snakeoil générée par `ssl-cert`, sans lien avec la sentinelle |
| Analyse snakeoil | `PASS` | `kasmvncserver` dépend de `ssl-cert`; KasmVNC référence la paire par défaut, mais ForgeLocal force `-sslOnly 0` et `-SecurityTypes None` |
| Image secret scan Trivy remédiée | `TRIVY_IMAGE_SECRET_REMEDIATED` | 0 secret sur l’image reconstruite exacte `sha256:c8a77b4acbb01794b14a84b100aa05a9ebfb9584c10781e800556e924070cca5`; image locale de régression supprimée après cleanup |
| Cleanup Docker | PASS | Aucun conteneur/processus résiduel ; daemon et socket arrêtés/supprimés |
| Test de non-régression snakeoil | PASS | `scripts/test-docker-image-hardening.sh` vérifie à chaque image l’absence de clé, certificat, variantes et symlink |
| Identité image par digest | PASS borné | Digest exact testé conservé ; aucune image courante ni digest de registry n’est revendiqué |
| Archives historiques R7/V2 | Héritées, chaîne de conservation séparée | `INHERITED_FROM_R7`, non modifiées et non repackagées |

## Correctif appliqué

Le diff fonctionnel comprend la suppression de la lecture du token dans `docker/entrypoint.sh`, la conservation d’un message non secret, la suppression du bloc équivalent dans `scripts/start.sh`, la suppression des artefacts snakeoil générés par `ssl-cert` dans `docker/Dockerfile.run`, et la correction des deux index de boucle ShellCheck par `_`. Les tests versionnés reçoivent la sentinelle uniquement par variable d’environnement du processus privé ; ils ne contiennent aucune valeur de sentinelle.

Aucun `set -x`, aucune substitution de commande contenant le token, aucun message d’erreur indirect et aucune valeur réelle ou synthétique n’a été ajoutée aux rapports, aux sidecars, aux archives ou aux traces publiques.

## Résultats ciblés post-correctif

| Contrôle | Résultat | Observations |
|---|---|---|
| Test entrypoint privé, fichier présent | `PASS` | Sentinelle absente de stdout/stderr ; permission vérifiée à `0600` |
| Test entrypoint privé, fichier absent | `PASS` | Aucun affichage de token ; comportement d’attente contrôlé |
| Requête authentifiée synthétique | `PASS` | Route protégée `/api/runtimes`, HTTP 200, réponse sans sentinelle |
| Token absent | `PASS` | HTTP 401, raison `missing`, réponse sans sentinelle |
| Token invalide | `PASS` | HTTP 401, raison `invalid`, réponse sans sentinelle |
| Token expiré | `PASS` | HTTP 401, raison `expired`, réponse sans sentinelle |
| Token révoqué | `PASS` | Révocation HTTP 204 puis HTTP 401, raison `revoked`, réponse sans sentinelle |
| Variables d’environnement et artefacts temporaires | `PASS` | Sentinelle absente de l’environnement du serveur et après cleanup |
| Nettoyage processus et répertoire temporaire | `PASS` | Processus serveur et fichiers temporaires nettoyés |
| `bash -n docker/entrypoint.sh scripts/start.sh` | `PASS` | Syntaxe valide |
| Garde statique contre log/lecture runtime | `PASS` | Aucun motif runtime résiduel identifié |
| Go `test -count=1 -race ./cmd/... ./internal/...` | `PASS` | `CGO_ENABLED=1`, toolchain locale |
| Go `vet ./cmd/... ./internal/...` | `PASS` | Aucun diagnostic bloquant |
| Go `build ./cmd/... ./internal/...` | `PASS` | Build réussi |
| Dashboard `pnpm install --frozen-lockfile` | `PASS` | Lockfile respecté |
| Dashboard `pnpm run check` | `PASS` | Type/check réussi |
| Dashboard `pnpm run build` | `PASS` | Build réussi |
| Gitleaks source-only | `PASS` | 0 finding ; scan redacted |
| Gitleaks extraction sans historique | `PASS` | 0 finding ; scan de répertoire, rapport JSON vide |
| Trivy filesystem secrets | `PASS` | 0 secret détecté |
| Trivy filesystem complet | `PASS` | 0 secret ; 6 misconfigurations Docker restantes à classer |
| ShellCheck des deux lanceurs et bancs de test | `SHELLCHECK_PASS` | Exit code nul après correction des index et du contrôle `/proc` |
| `git diff --check` | `PASS` | Aucun whitespace error |
| `git fsck --full` local | `PASS` | Intégrité du clone locale vérifiée |
| Docker version/info/Buildx/context | `PASS` via sudo | Engine `29.1.3`, Buildx `0.30.1`; accès direct utilisateur refusé par socket `root:docker`, sans modification du groupe |
| Docker build bridge | `FAIL` contrôlé | Échec iptables table `raw`; aucun contournement du code |
| Docker build host-network | `PASS` | Build réel de l’image runtime v2.1.12 |
| Docker run/logs/layers/inspection/cleanup | `PASS` | Sentinelle absente logs, history et layers ; fichier `0600`, cas absent/invalide, cleanup PASS |
| Trivy image baseline | `TRIVY_IMAGE_SECRET_OPEN` | 123 vulnérabilités, 1 clé snakeoil, 0 misconfiguration |
| Trivy image remédiée | `TRIVY_IMAGE_SECRET_REMEDIATED` | 123 vulnérabilités, 0 secret, 0 misconfiguration |
| Syft SBOM CycloneDX image remédiée | `PASS` | Syft `1.51.0`, SBOM produit, 5.801.675 octets |
| Grype SBOM image remédiée | `GRYPE_FINDINGS_TRIAGE_PENDING` | Grype `0.117.0`, 379 matches ; aucun ignore global |
| Grype policy gate unitaire | `PASS` | La fixture Critical synthétique non approuvée échoue, puis passe uniquement avec owner/justification/expiration |
| Grype policy gate sur rapport image réel | `FAIL` attendu | 43 Critical/High non approuvées, 0 exception approuvée ; la gate bloque effectivement |
| Matrice dédupliquée Trivy/Grype | `IMAGE_VULNERABILITIES_TRIAGE_PENDING` | 453 lignes uniques ; 85 Critical/High prioritaires ; exploitabilité, action, owner et échéance restent à renseigner |
| Docker services/socket final | `PASS` cleanup | Docker, docker.socket et containerd inactifs/désactivés ; socket supprimé ; 25 Go libres |
| Misconfigurations Docker | `IMAGE_MISCONFIGURATIONS_TRIAGE_PENDING` | 6 findings initiaux ; 3 corrigés et absents du rescan réel c8a77 ; 3 restent ouverts : DS-0002 build, DS-0026 build et DS-0002 runtime ; matrice dédiée conservée |
| Revue indépendante | `INDEPENDENT_REVIEW_PENDING` | Aucune revue humaine indépendante n’est attestée dans les preuves |
| Rescan image post-correction | `PASS` partiel | Digest `sha256:c8a77b4acbb01794b14a84b100aa05a9ebfb9584c10781e800556e924070cca5`; Trivy : 0 secret, 0 misconfiguration, 123 vulnérabilités ; Syft : 17 474 composants ; Grype : 379 matches dont 43 Critical/High bloquants |
| Authentification image post-correction | `PASS` | Route `/api/runtimes` HTTP 200, fichier 0600, sentinelle absente des réponses, logs et inspection, cleanup PASS |
| Protection de branche | `NOT_ACTIVE` | Vérification antérieure : API GitHub 404 `Branch not protected`; re-vérification ultérieure non exécutée après HTTP 401 d’authentification ; job sécurité obligatoire non vérifié |
| Cleanup final documenté | `PASS` | Aucun conteneur, aucune image temporaire, aucun cache, aucun processus ou socket résiduel |
| Grype SBOM source précédent | `NOT_EXECUTED` | Remplacé par le scan Grype du SBOM image réel |
| Firefox/Camoufox natifs, SystemVault, Windows/macOS | `BLOCKED_ENVIRONMENT_REQUIRED` / `NATIVE_SYSTEMVAULT_NOT_TESTED` | Environnements non disponibles |
| Références et packages R7 historiques | `INHERITED_FROM_R7` | Non présentés comme nouvelle exécution |

Le replay post-correction a reconfirmé `bash -n`, Go race/vet/build, Dashboard check/build, Gitleaks source et extraction, Trivy filesystem, ShellCheck avec exit code nul, `git diff --check` et `git fsck --full`. Le test authentifié versionné `scripts/test-authenticated-synthetic.sh` est également PASS pour les états valid, missing, invalid, expired et revoked. Le cycle Docker réel a ajouté les contrôles image/layers/logs, token absent/invalide, permissions `0600`, Syft et Grype. Les assertions ne rapportent que l’absence de sentinelle. Les sorties brutes d’installation, build, cycle, scans et cleanup, ainsi que les rapports JSON et la matrice de triage, sont conservés sous `evidence/FINAL_SECRET_REMEDIATION/`. Le cleanup final est documenté dans `evidence/FINAL_SECRET_REMEDIATION/snakeoil/snakeoil-final-cleanup-raw.log`. Aucun journal ne contient la sentinelle synthétique.

## Gates CI ajoutées

Le workflow `.github/workflows/ci.yml` contient maintenant un job `security-scans` et une gate d’authentification synthétique. Les contrôles CI ajoutés couvrent Gitleaks source et historique pertinent, Trivy filesystem secrets, Syft source et image, Grype source et image, build de l’image runtime durcie, inspection hardening/layers, cycle anti-fuite, ShellCheck et test authentifié synthétique. Le job de release n’a pas été modifié et aucun push d’image n’est effectué par ces gates.

Le workflow reste soumis à la politique de provenance existante, validée localement par `node scripts/check-ci-provenance-workflow.mjs`. Les vulnérabilités et matches scanners ne sont pas ignorés globalement. Les étapes Grype `report` conservent les rapports comme artefacts et peuvent être `continue-on-error` uniquement pour permettre la génération du rapport ; elles sont immédiatement suivies d’une `Grype policy gate` bloquante. Cette policy échoue pour tout Critical/High non approuvé par une exception exacte comportant owner, justification et expiration. Sur le rapport image réel, elle retourne exit code 1 avec 43 Critical/High ouverts et 0 exception approuvée. La gate Trivy secrets échoue si un secret est présent dans l’image. La définition locale est `CI_GATES_DEFINED_LOCAL_VERIFICATION_PASS`. Des exécutions GitHub Actions distantes réelles ont été observées : la première run `33033705812` a échoué sur la référence inexistante `aquasec/trivy:0.74.2`, puis le workflow a été corrigé vers `0.74.0`. La dernière run ayant exécuté les changements fonctionnels est `33069090849`, sur le HEAD `e3e1e24448b4bdf28f8b09e13f03bd21f5a98fd1` : elle est terminée en échec. Les runs des commits package et documentaire suivants, `33070413940` sur `01c3bdb3e4b4785584edbacec91efe8840aecb0c` et `33071664971` sur `c1c70beb4c46448eb136af101fed8c35a000e62f`, sont également terminées en échec et conservées dans l’évidence CI. Le job sécurité échoue sur les 43 Critical/High non approuvées, ce qui est le comportement bloquant attendu ; le job `test` échoue sur `TestBackupV1CreateModifyRestoreIsolation` parce que le Chromium fourni par le runner GitHub ne se termine pas après 60 secondes malgré le probe headless local et plusieurs variantes testées, ce qui est classé `BLOCKED_ENVIRONMENT_REQUIRED` et non présenté comme un succès. Le statut est `CI_REMOTE_EXECUTION_OBSERVED=TRUE`, `CI_REMOTE_EXECUTION_PASS=FALSE`, `CI_REMOTE_EXECUTION_NOT_CONFIRMED=FALSE`, `CI_CURRENT_HEAD_CI_RUN=33071664971`, `CI_CURRENT_HEAD_CI_CONCLUSION=FAILURE`, `CI_REMOTE_EXECUTION_CONCLUSION=FAILURE`, `BRANCH_PROTECTION_ENFORCEMENT=NOT_ACTIVE`, `SECURITY_JOB_REQUIRED_FOR_PR=NOT_VERIFIED` et `INDEPENDENT_REVIEW_PENDING`. Une re-vérification GitHub CLI antérieure avait retourné HTTP 401 `Bad credentials`, puis l’accès a été rétabli : `CI_REMOTE_FINAL_RECHECK=PASS`, `CI_CURRENT_PASTED13_HEAD=e3e1e24448b4bdf28f8b09e13f03bd21f5a98fd1`. La PR et la protection de branche ont été vérifiées après réauthentification ; la protection reste inactive. Les compléments de pasted_content_13 sont désormais publiés sur la branche corrective ; la CI du commit package est observée en échec sur la policy Grype et le timeout Chromium distant. Aucun changement n’a été effectué sur `main` et aucune image n’a été poussée.

## Triage et décisions ouvertes

La matrice finale `evidence/FINAL_SECRET_REMEDIATION/triage/image-trivy-grype-matrix.csv` déduplique par `vulnerability_id + component + installed_version`. Elle contient **453 lignes uniques**, dont **85 Critical/High prioritaires**. Chaque ligne indique le composant, la version installée, la présence dans l’image runtime, l’étape `build_or_runtime`, l’identifiant, la sévérité, le scanner, les versions corrigées rapportées par les scanners, l’exploitabilité, l’action, l’owner, l’échéance, l’état d’exception et une décision `OPEN_*_MANUAL_REVIEW_REQUIRED`. Les 453 lignes sont actuellement `RUNTIME_IMAGE`, `OWNER_UNASSIGNED` et `DUE_DATE_OWNER_REQUIRED` : ces marqueurs rendent l’absence de triage explicite et ne constituent pas des décisions positives. L’exploitabilité n’est pas inventée : elle reste `NOT_ASSESSED_SCANNER_DATA_UNAVAILABLE`. La policy Grype a été testée avec une fixture synthétique et exécutée contre le rapport réel : 43 Critical/High non approuvées bloquent effectivement la gate, avec 0 exception active.

Le fichier `evidence/FINAL_SECRET_REMEDIATION/triage/image-secret-triage.json` documente le finding snakeoil baseline, son créateur `ssl-cert`/`make-ssl-cert`, les références KasmVNC, la correction appliquée et le résultat Trivy post-correction à zéro secret. Les vulnérabilités Trivy et les matches Grype restent donc `IMAGE_VULNERABILITIES_TRIAGE_PENDING` et `GRYPE_FINDINGS_TRIAGE_PENDING` jusqu’à une revue finding par finding avec exploitabilité et propriétaire. Les six misconfigurations initiales sont suivies séparément dans `evidence/FINAL_SECRET_REMEDIATION/triage/docker-misconfigurations.csv` : trois corrections sont vérifiées par le rescan image c8a77, tandis que trois restent ouvertes et soumises à revue de compatibilité ou d’applicabilité. Les vulnérabilités Trivy et les matches Grype restent inchangés dans leur matrice : 453 lignes uniques, 85 Critical/High, exploitabilité non évaluée et décision pending.

## Package et vérification

Le package final comprend un ZIP, un TAR, leurs sidecars SHA-256, un manifeste non auto-référentiel et un bundle Git. Le manifeste liste les hashes des composants sans inclure son propre hash. Les identifiants exacts du HEAD publié, du commit package, des fichiers et du bundle sont consignés dans `PUBLIC_VERIFICATION.log`; la vérification par clone neuf porte sur le dernier HEAD public confirmé. Aucune référence R7/V2 n’est réétiquetée comme nouvelle preuve.

La vérification publique exige un clone frais, un checkout explicite de `validation/final-secret-remediation`, la vérification des SHA-256 du ZIP/TAR/bundle, l’extraction fraîche du ZIP/TAR, `git bundle verify` et `git fsck --full`. Le journal de vérification sera livré avec les sidecars et restera sans valeur de token.

## Verdict obligatoire

```text
REAL_TOKEN_DISCLOSURE_PATH_CONFIRMED
SECRET_VALUE_NOT_RETAINED
NO_REAL_SECRET_ROTATION_PERFORMED=TRUE (OWNER_CONFIRMATION_REQUIRED)
SECRET_REAL_USE_STATUS=OWNER_CONFIRMATION_REQUIRED
SECRET_REMEDIATION_VALIDATED_IN_REAL_DOCKER_CYCLE
TRIVY_IMAGE_SECRET_REMEDIATED
AUTHENTICATED_SYNTHETIC_REQUEST=PASS
AUTHENTICATED_SYNTHETIC_NEGATIVE_CASES=PASS
IMAGE_VULNERABILITIES_TRIAGE_PENDING
GRYPE_FINDINGS_TRIAGE_PENDING
GRYPE_POLICY_GATE_BLOCKING=TRUE
GRYPE_POLICY_REAL_REPORT=BLOCKED_OPEN_CRITICAL_HIGH_43
CI_GATES_DEFINED_LOCAL_VERIFICATION_PASS
CI_REMOTE_EXECUTION_OBSERVED=TRUE
CI_REMOTE_EXECUTION_PASS=FALSE
CI_REMOTE_EXECUTION_NOT_CONFIRMED=FALSE
CI_REMOTE_EXECUTION_CONCLUSION=FAILURE
CI_REMOTE_FINAL_RECHECK=PASS
CI_CURRENT_PASTED13_HEAD=e3e1e24448b4bdf28f8b09e13f03bd21f5a98fd1
CI_REMOTE_RUN_ID=33069090849
CI_REMOTE_RUN_URL=https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/actions/runs/33069090849
CI_REMOTE_SECURITY_JOB_ID=98506432000
CI_REMOTE_SECURITY_JOB_CONCLUSION=FAILURE_GRYPE_POLICY_43_UNAPPROVED_CRITICAL_HIGH
CI_REMOTE_TEST_JOB_ID=98506431843
CI_REMOTE_TEST_JOB_CONCLUSION=FAILURE_CHROMIUM_HOSTED_RUNNER_TIMEOUT
BRANCH_PROTECTION_ENFORCEMENT=NOT_ACTIVE
SECURITY_JOB_REQUIRED_FOR_PR=NOT_VERIFIED
INDEPENDENT_REVIEW_PENDING
SHELLCHECK_PASS
DOCKER_HOST_NETWORK_BUILD_PASS
DOCKER_BRIDGE_BUILD_BLOCKED_BY_ENVIRONMENT
PUBLIC_RELEASE_BLOCKED
FORGELOCAL_PRODUCTION_READY=false
```

Le correctif des deux divulgations runtime, la requête authentifiée synthétique et la suppression de la clé snakeoil sont démontrés sur le périmètre testable. Cette démonstration ne constitue pas une autorisation de release : le verdict global demeure bloqué tant que les 85 Critical/High de la matrice, les 43 Critical/High bloquants de la policy Grype, les 379 matches Grype, la revue d’exploitabilité et la validation bridge ne sont pas clôturés formellement.

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized ForgeLocal public sanitized repository.
[2]: https://github.com/gitleaks/gitleaks Gitleaks project documentation and release source.
[3]: https://trivy.dev/latest/ Trivy documentation.
[4]: https://www.shellcheck.net/ ShellCheck project documentation.
