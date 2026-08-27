# Rapport de remédiation critique — ForgeLocal

**Date d’exécution :** 26 août 2026 (UTC)
**Dépôt :** `davidwilsonbest89-afk/forgelocal-public-sanitized`
**Branche corrective :** `validation/final-secret-remediation`
**Point de départ figé :** `1d0fe24da0edbddc5eab30d971628df8fe5a92d3` sur `validation/final-environment-qualification`

## Conclusion exécutive

Une divulgation critique réelle a été confirmée dans la branche de qualification : la sortie de `docker/entrypoint.sh` interpolait le contenu du token API, et `scripts/start.sh` contenait une seconde divulgation runtime indépendante. La preuve originale a été conservée sous forme **redacted** avec les chemins, lignes, commit source, permissions et hashes ; aucune valeur de token n’a été conservée ou redistribuée.

Le correctif appliqué est volontairement minimal. Les deux lanceurs ne lisent plus le fichier de token et ne l’impriment plus. L’entrypoint conserve uniquement le test d’existence du fichier et affiche un état générique redacted ; le lanceur local supprime entièrement le bloc de lecture et d’affichage. Un binaire de test suivi par Git, classé artefact généré contenant des chaînes de fixture bearer-like, a également été retiré de la branche corrective ; seul son hash historique est conservé dans la preuve.

Le contrôle comportemental avec sentinelle synthétique privée, le contrôle de permissions `0600`, le cas fichier absent, le nettoyage des processus et les scans secrets ciblés sont PASS. Une installation contrôlée a ensuite réussi : Docker Engine `29.1.3` et Buildx `0.30.1` ont été installés depuis les paquets Ubuntu disponibles. Le daemon réel a démarré via systemd et `sudo docker info` a confirmé le serveur actif ; aucun ajout au groupe Docker n’a été effectué. Le premier build bridge a échoué avec l’erreur noyau/iptables `can't initialize iptables table raw`; un second build utilisant le réseau host du daemon réel a réussi sans modifier le code ni simuler Docker. Après le cycle, l’image, les caches, les conteneurs, les services et le socket ont été nettoyés ou arrêtés/désactivés. Le verdict de release reste bloqué par mandat et par les findings restants : `GOSEC_R7_BLOCKED_CRITICAL_FINDING`, `PUBLIC_RELEASE_BLOCKED`, `GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS`, `FORGELOCAL_PRODUCTION_READY=false`.

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
| Image secret scan Trivy | `OPEN_FINDING` | 1 clé privée snakeoil générée par Ubuntu dans `/etc/ssl/private/ssl-cert-snakeoil.key`, sans lien avec la sentinelle |
| Cleanup Docker | PASS | Aucun conteneur/processus résiduel ; daemon et socket arrêtés/supprimés |
| Archives historiques R7/V2 | Héritées, chaîne de conservation séparée | `INHERITED_FROM_R7`, non modifiées et non repackagées |

## Correctif appliqué

Le diff fonctionnel est limité à la suppression de la lecture du token dans `docker/entrypoint.sh`, au remplacement de l’ancien affichage par un message non secret, et à la suppression du bloc équivalent dans `scripts/start.sh`. Le test versionné `scripts/test-entrypoint-token-redaction.sh` reçoit la sentinelle uniquement par variable d’environnement du processus privé ; il ne contient aucune valeur de sentinelle.

Aucun `set -x`, aucune substitution de commande contenant le token, aucun message d’erreur indirect et aucune valeur réelle ou synthétique n’a été ajoutée aux rapports, aux sidecars, aux archives ou aux traces publiques.

## Résultats ciblés post-correctif

| Contrôle | Résultat | Observations |
|---|---|---|
| Test entrypoint privé, fichier présent | `PASS` | Sentinelle absente de stdout/stderr ; permission vérifiée à `0600` |
| Test entrypoint privé, fichier absent | `PASS` | Aucun affichage de token ; comportement d’attente contrôlé |
| Nettoyage processus et répertoire temporaire | `PASS` | Processus stub et fichiers temporaires nettoyés |
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
| ShellCheck des deux lanceurs | `FAIL` au sens exit code | Seulement deux warnings SC2034 sur l’index de boucle inutilisé ; aucun diagnostic secret ou erreur |
| `git diff --check` | `PASS` | Aucun whitespace error |
| `git fsck --full` local | `PASS` | Intégrité du clone locale vérifiée |
| Docker version/info/Buildx/context | `PASS` via sudo | Engine `29.1.3`, Buildx `0.30.1`; accès direct utilisateur refusé par socket `root:docker`, sans modification du groupe |
| Docker build bridge | `FAIL` contrôlé | Échec iptables table `raw`; aucun contournement du code |
| Docker build host-network | `PASS` | Build réel de l’image runtime v2.1.12 |
| Docker run/logs/layers/inspection/cleanup | `PASS` | Sentinelle absente logs, history et layers ; fichier `0600`, cas absent/invalide, cleanup PASS |
| Trivy image | `PASS` de scan avec `OPEN_FINDING` | 123 vulnérabilités, 1 secret image (clé privée snakeoil Ubuntu), 0 misconfiguration |
| Syft SBOM CycloneDX image | `PASS` | Syft `1.51.0`, SBOM produit, 5.801.674 octets |
| Grype SBOM image | `PASS` avec résultats ouverts | Grype `0.117.0`, 379 matches ; warning de précision Go documenté |
| Docker services/socket final | `PASS` cleanup | Docker, docker.socket et containerd inactifs/désactivés ; socket supprimé |
| Grype SBOM source précédent | `NOT_EXECUTED` | Remplacé par le scan Grype du SBOM image réel |
| Firefox/Camoufox natifs, SystemVault, Windows/macOS | `BLOCKED_ENVIRONMENT_REQUIRED` / `NATIVE_SYSTEMVAULT_NOT_TESTED` | Environnements non disponibles |
| Références et packages R7 historiques | `INHERITED_FROM_R7` | Non présentés comme nouvelle exécution |

Le replay additionnel prescrit par la pièce jointe a reconfirmé `bash -n`, Go race/vet/build, `git diff --check`, Gitleaks source-only, Trivy filesystem et `git fsck --full`. Le cycle Docker réel a ajouté les contrôles image/layers/logs, Syft et Grype ; les assertions ne rapportent que l’absence de sentinelle. ShellCheck conserve uniquement les warnings SC2034 déjà classés. Les sorties brutes d’installation, build, cycle, scans et cleanup, ainsi que les rapports JSON, sont conservés sous `evidence/FINAL_SECRET_REMEDIATION/`. Aucun journal ne contient la sentinelle synthétique.

## Package et vérification

Le package final comprend un ZIP, un TAR, leurs sidecars SHA-256, un manifeste non auto-référentiel et un bundle Git. Le manifeste liste les hashes des composants sans inclure son propre hash. Les identifiants exacts du HEAD final, du commit d’évidence, du commit package, des fichiers et du bundle sont consignés dans `PUBLIC_VERIFICATION.log` après validation par clone neuf ; aucune référence R7/V2 n’est réétiquetée comme nouvelle preuve.

La vérification publique exige un clone frais, un checkout explicite de `validation/final-secret-remediation`, la vérification des SHA-256 du ZIP/TAR/bundle, l’extraction fraîche du ZIP/TAR, `git bundle verify` et `git fsck --full`. Le journal de vérification sera livré avec les sidecars et restera sans valeur de token.

## Verdict obligatoire

> `GOSEC_R7_BLOCKED_CRITICAL_FINDING`
> `PUBLIC_RELEASE_BLOCKED`
> `GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS`
> `FORGELOCAL_PRODUCTION_READY=false`

Le correctif des deux divulgations runtime identifiées est démontré sur le périmètre testable. Cette démonstration ne constitue pas une autorisation de release : le verdict global demeure bloqué jusqu’à la clôture formelle des autres findings et à l’exécution des contrôles dépendant des environnements indisponibles.

## Références

[1]: https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized ForgeLocal public sanitized repository.
[2]: https://github.com/gitleaks/gitleaks Gitleaks project documentation and release source.
[3]: https://trivy.dev/latest/ Trivy documentation.
[4]: https://www.shellcheck.net/ ShellCheck project documentation.
