# ForgeLocal — Passation autonome pour nouvelle session

## Mission

Cette passation permet à une nouvelle session de reprendre **sans supposer qu’un rapport est vrai**. Elle doit auditer les artefacts reçus, appliquer les gates permanents, puis développer les lots restants du CDC dans l’ordre documenté ci-dessous.

> **Règle absolue :** aucun lot ne peut être déclaré terminé sur la base d’un récit. Chaque conclusion doit être rattachée à une sortie brute, une empreinte, un commit, un bundle et une extraction fraîche vérifiable.

## Dépôt et baseline obligatoire

| Élément | Valeur |
|---|---|
| Dépôt GitHub public | <https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized> |
| Branche d’archive | `continuation-t00-t27-canonical` |
| Tag annoté de référence | `t00-t27-complete-20260820` |
| Objet de tag | `72d54110c89583beacc556bb103f881b667d8137` |
| Commit déréférencé | `69411e65c880d168832a65fc8475cc97d562a9ad` |
| Ancienne branche T27-R1 | `continuation-t27-r1-canonical` — ne pas l’utiliser comme baseline de nouveau développement sans comparaison avec le tag T00–T27. |

La nouvelle session doit toujours partir du **tag `t00-t27-complete-20260820`**, puis appliquer un bundle de lot vérifié ou une branche de lot validée. Ne jamais repartir d’un worktree local non identifié, d’un `HEAD` détaché inconnu ou d’un bundle historique seul.

## Pré-requis d’environnement

Utiliser une session ou machine avec **au moins 10 GiB libres** pour les tests Go complets ; **30 GiB ou plus** est recommandé si les archives LFS T00–T27 doivent être réhydratées et si plusieurs preuves sont conservées localement.

L’environnement précédent n’avait qu’environ 39 MiB libres. Il ne faut ni supprimer une copie canonique locale, ni effacer un bundle/ZIP de preuve pour pallier ce manque. Préférer un clone ou worktree temporaire, puis supprimer uniquement le répertoire temporaire après une vérification fraîche réussie et la préservation canonique.

Versions attendues : Go `1.25.13` avec `GOTOOLCHAIN=auto`, Git LFS, Node `22.13.0`, pnpm `10.4.1`, Gitleaks, Gosec, `staticcheck`, `golangci-lint`, `govulncheck`, Syft, Trivy et OSV Scanner. Confirmer les versions réellement installées au début de chaque lot.

## Démarrage propre

```bash
export GOTOOLCHAIN=auto
export GIT_LFS_SKIP_SMUDGE=1

gh repo clone davidwilsonbest89-afk/forgelocal-public-sanitized forgelocal
cd forgelocal
git fetch --tags --force
git checkout --detach t00-t27-complete-20260820
git rev-parse HEAD
# attendu : 69411e65c880d168832a65fc8475cc97d562a9ad
git fsck --full
```

N’activer `git lfs pull` que pour les artefacts réellement nécessaires. Les fragments du grand paquet T00–T23 sont reconstruisibles mais ne doivent pas être téléchargés par réflexe.

## État réellement vérifié

| Périmètre | État de reprise | Référence et réserve |
|---|---|---|
| T00–T23 | Copie distante conservée | Paquet de 2,42 Go en fragments LFS avec ordre et hash de reconstruction documentés. |
| T24–T26 | Code, contrats et archives de preuve présents dans la copie distante T00–T27 | Traiter toute requalification historique comme postérieure ; ne pas prétendre retrouver des logs originaux si seuls bundles/ZIP sont disponibles. |
| T27-R1 / CR-01 à CR-05 | Clôture locale vérifiable documentée | Gates permanents conservés ; CR-08 reste non testé en environnement SystemVault natif. |
| T28 Extensions autorisées | **Non approuvé** | L’archive Agent 03 existe côté utilisateur mais n’a pas pu être téléchargée dans la sandbox saturée. Commencer par un audit physique, sans code. |
| T29 Password Manager | Contrat / modèle de menace attendu de l’Agent 04 | Aucune implémentation avant politique de stockage, redaction et décision SystemVault. |
| T30 Diagnostics environnement | Code local finalisé mais non encore intégré | Head local `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb`; bundle/kit de handover disponibles chez le propriétaire. Seul le replay global `go test -count=1 -race ./...` manque avant registre et push. |
| T31–T37 | Non commencés | Contrats à écrire ou recevoir avant code. |
| T38–T42 | Non commencés | Dépendent des diagnostics et contrats précédents. |

## Gates permanents, non négociables

```text
PUBLIC_RELEASE_BLOCKED
SCAN_BLOCKED_UNKNOWN
NATIVE_SYSTEMVAULT_NOT_TESTED
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
```

Sont interdits sans une autorisation distincte, écrite et bornée : runtime navigateur réel, Camoufox, proxy réel, cookie réel, migration utilisateur, SystemVault natif et release publique. Les tests dynamiques exigés sont donc des tests locaux synthétiques et négatifs, jamais un contournement des gates.

## Procédure obligatoire `BASELINE_DISCOVERY`

Avant **toute écriture de code**, créer `evidence/Txx/BASELINE_DISCOVERY_RAW.log`. Le fichier doit contenir, pour chaque commande : date UTC de début et de fin, CWD absolu, commande complète, sortie brute, exit code et référence Git observée.

Au minimum :

```bash
date -u +%Y-%m-%dT%H:%M:%SZ
pwd
git status --short
git rev-parse HEAD
git show -s --format=fuller HEAD
git tag --points-at HEAD
git remote -v
git fsck --full
df -h .
go version
node --version
pnpm --version
git lfs version
gitleaks version
gosec -version
```

Si une baseline, un tag, un bundle, un ZIP ou un sidecar semble absent, effectuer d’abord une recherche dans le workspace courant, les uploads, les kits extraits, les worktrees historiques, les bundles locaux et le dépôt GitHub. Consigner la recherche. Déclarer `BLOCKED_MISSING_BASELINE` uniquement après cette procédure, en précisant exactement le fichier, hash ou ref absent.

## Discipline par lot

1. Créer une branche `work/tXX-<scope>` depuis le tag ou le head canonique validé.
2. Rédiger/valider un contrat de périmètre fermé, incluant exclusions, tests négatifs, redaction et décisions produit non déductibles.
3. Implémenter uniquement le delta autorisé.
4. Exécuter les tests ciblés, puis les qualifications globales prévues : `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, tests Dashboard si concerné, `git diff --check`.
5. Exécuter Gitleaks sur la plage Git immuable `baseline..HEAD`, puis Gosec baseline/head avec comparaison normalisée. Ne jamais présenter `git diff HEAD` sur un worktree propre comme scan de delta.
6. Créer un bundle delta depuis la baseline, un sidecar SHA-256 **portable** (nom relatif uniquement), un kit ZIP, un manifeste non auto-référentiel et un sidecar externe du ZIP.
7. Vérifier `sha256sum -c`, `unzip -t`, le manifeste après extraction fraîche, Gitleaks de l’extraction, `git bundle verify`, clone neuf, checkout du commit/tag et `git fsck --full`.
8. Mettre à jour le registre canonique, `todo.md`, le changelog et le document de clôture avec les hashes réels et les limites. Ne pas auto-valider une limitation environnementale comme PASS.
9. Pousser la branche du lot et ses artefacts vérifiés vers GitHub. Vérifier le commit distant et les pointeurs LFS avant le lot suivant.
10. Ne supprimer aucune copie canonique locale avant confirmation de réception et vérification par le propriétaire.

## Ordre d’exécution des lots restants

| Ordre | Lot | Prérequis | Sortie minimale attendue |
|---:|---|---|---|
| 0 | Audit Agent 03 | Archive T28 physiquement reçue | Verdict T28 précis : approuvé, corrigible ou bloqué. |
| 1 | T28 Extensions | Contrat Agent 03 approuvé + décisions produit | Allowlist, provenance, refus fail-closed, tests sans extension réelle. |
| 2 | T29 Password Manager | Contrat Agent 04 + modèle de menace | Politique redacted, stockage autorisé, tests synthétiques ; aucune clé réelle. |
| 3 | T30 | Kit T30-R3 + ≥5 GiB libres | Replay `-race ./...` vert, registre canonique, branche GitHub T30. |
| 4 | T31 | Contrat Canvas/WebGL/Audio | Diagnostics explicitement projetés/unsupported sans runtime interdit. |
| 5 | T32 | T31 stabilisé | Contrat ClientRects et tests de redaction. |
| 6 | T33 | T30/T31 stabilisés | QA géolocalisation synthétique, aucune position réelle. |
| 7 | T34 | Contrats diagnostic communs | Diagnostics matériel read-only/redacted. |
| 8 | T35 | T34 stabilisé | Font Bundle : inventaire, provenance et limites de licence. |
| 9 | T36 | T31–T35 stabilisés | Détection de dérive, baseline explicite et seuils documentés. |
| 10 | T37 | T36 stabilisé | Profile Health agrégé, read-only et explications redacted. |
| 11 | T38 | T37 stabilisé | Suivi local de session, lifecycle et redaction. |
| 12 | T39 | T28, T29, T38 stabilisés | Import/export complet synthétique, validation fail-closed et rollback. |
| 13 | T40 | T28–T39 stabilisés | API locale finalisée et OpenAPI cohérent. |
| 14 | T41 | T40 stabilisé | Dashboard final, E2E, TypeScript et build. |
| 15 | T42 | Tous lots intégrés | Suite finale complète, scans, SBOM/provenance et décision MVP. |

T31–T35 peuvent avoir des **contrats** préparés en parallèle, mais les implémentations et merges doivent rester séquentiels dès qu’elles touchent les mêmes modèles, API ou Dashboard.

## Artefacts que le propriétaire peut fournir

| Artefact | Utilité |
|---|---|
| `agent03-t28-delivery.zip` | Audit physique et décision T28. |
| `forgelocal-t30-master-handover-cbf3a50.zip` + `.sha256` | T30 R1–R3 complet, bundle final, script de replay, sources delta et README. |
| Paquets de continuité T00–T27 si nécessaire | Reconstitution historique hors Git LFS. |

## Définition de fin

Ne déclarer T42 clôturé que si chaque lot possède : baseline brute, commit, tests/scans réellement exécutés, bundle, ZIP, sidecars, manifeste, extraction fraîche, registre canonique, branche GitHub vérifiée et une décision explicite distinguant `PASS`, `BLOCKED`, `NOT_TESTED` et `NOT_IN_SCOPE`.

Les gates de release peuvent rester bloqués même si le MVP local est qualifié. Une clôture locale n’est jamais une autorisation de runtime réel ou de release.
