# Protocole de qualification gratuite — Oracle Cloud Always Free

**Statut :** procédure conditionnelle ; elle ne lève aucun gate tant que les preuves natives ne sont pas produites et revues.

## Décision et limites

La seule option gratuite identifiée qui puisse, sous conditions, héberger une cible **Ubuntu 24.04 amd64** persistante est une instance Oracle Cloud Always Free de type AMD. Oracle documente deux micro-instances AMD Always Free et les distingue des instances Ampere A1, qui sont à architecture Arm et ne sont donc pas admissibles pour le candidat Linux amd64 [1].

> Ce protocole n’autorise ni conteneur, ni Codespaces, ni DistroSea, ni OnWorks, ni une version Ubuntu différente de 24.04 à servir de preuve de release. Ces environnements peuvent aider à un smoke test, mais ne peuvent pas lever `SYSTEMVAULT_NATIVE_PER_TARGET`.

L’inscription Oracle est gratuite, mais exige habituellement un numéro de téléphone et une carte de paiement pour vérifier l’identité. Oracle indique qu’une empreinte d’autorisation temporaire peut être effectuée, sans débit lorsque le compte n’est pas converti en formule payante [2] [3]. **Sans compte gratuit vérifié, aucune solution gratuite vérifiée ne permet de lever ce gate à ce stade.**

## Contraintes à maintenir

| Élément | Règle |
|---|---|
| Candidat RC | Ne pas modifier l’archive, le runtime, le SBOM, le manifeste, son SHA-256 ou le commit référencé. |
| Portée | Ubuntu 24.04 amd64 uniquement ; aucune déclaration Windows, macOS ou Arm. |
| Coût | Choisir exclusivement des ressources marquées **Always Free** ; ne pas convertir le compte en payant ; ne pas sélectionner une machine non éligible. |
| Secrets | Aucun secret réel, token proxy, clé privée ou compte personnel dans la VM, les logs, l’archive ou la conversation. |
| Preuves | Exporter seulement les fichiers assainis prévus par le runbook. |

## Phase A — Création de la cible gratuite

Cette phase est réalisée par le titulaire du compte Oracle. Elle ne doit démarrer qu’après vérification visuelle de chaque libellé **Always Free** dans la console.

1. Créer un compte Oracle Cloud Free Tier et sélectionner avec soin la région d’origine : Oracle précise que les instances Always Free doivent être créées dans la région d’origine [3].
2. Créer une instance de calcul **AMD**, architecture x86_64, avec l’image Canonical **Ubuntu Server 24.04 LTS**. Ne pas sélectionner l’instance Arm A1, même si elle propose davantage de mémoire, car elle ne correspond pas à l’architecture annoncée.
3. Confirmer que l’instance affiche le statut **Always Free-eligible** avant création. Si la console annonce un coût, un tarif horaire, un essai limité ou une forme non éligible, annuler.
4. Utiliser un volume de démarrage restant dans les limites Always Free. Ne créer ni GPU, ni IP premium, ni stockage additionnel payant, ni service managé.
5. Créer un compte utilisateur desktop non privilégié destiné à la qualification. Le test de gate s’exécutera sous ce compte ; il ne s’exécutera jamais avec `sudo` ni avec `root`.

L’instance AMD gratuite est volontairement peu dotée. Oracle indique que la capacité Always Free peut être indisponible dans une région : si la création échoue pour manque de capacité, le résultat est un blocage d’infrastructure, non une raison d’assouplir le gate [1]. Si GNOME ou Secret Service ne sont pas utilisables de façon stable dans les ressources gratuites obtenues, le gate reste `PENDING`.

## Phase B — Préparation graphique, hors preuve de gate

Le système doit disposer d’un **vrai bureau graphique** et d’une session utilisateur avec D-Bus. La préparation système peut être faite lors du provisioning par l’administrateur de la VM ; elle n’est pas la preuve de qualification. L’environnement final doit permettre à l’utilisateur desktop de démarrer le bureau dans le navigateur au moyen d’un accès graphique protégé et temporaire.

L’accès graphique ne doit jamais être ouvert à Internet sans restriction. Limiter toute règle entrante à l’adresse IP publique courante de l’opérateur, employer un mot de passe unique, puis supprimer cette règle à la fin. Ne déposer aucun fichier de clé privée dans le dépôt ForgeLocal ni dans le dossier de preuve.

Avant de récupérer le dépôt, ouvrir la session graphique du compte desktop et vérifier que le coffre attendu est bien présent : `Secret Service`, D-Bus de session actif et keyring déverrouillé. Les prérequis seront de nouveau contrôlés automatiquement par le script officiel.

## Phase C — Préflight non sensible

Dans le terminal du **compte desktop graphique**, sans `sudo`, depuis un clone de travail du dépôt, exécuter :

```bash
printf 'euid=%s\n' "$EUID"
printf 'dbus=%s\n' "${DBUS_SESSION_BUS_ADDRESS:+PRESENT}"
printf 'xdg_runtime=%s\n' "${XDG_RUNTIME_DIR:+PRESENT}"
printf 'arch=%s\n' "$(uname -m)"
printf 'os='
. /etc/os-release && printf '%s %s\n' "$ID" "$VERSION_ID"
if [[ -f /.dockerenv ]] || grep -Eq '/(docker|containerd|kubepods|lxc)/' /proc/1/cgroup 2>/dev/null; then
  printf 'container=DETECTED\n'
else
  printf 'container=not_detected\n'
fi
command -v secret-tool >/dev/null && printf 'secret_tool=PRESENT\n' || printf 'secret_tool=ABSENT\n'
```

Les conditions minimales sont `euid` différent de zéro, `dbus=PRESENT`, `xdg_runtime=PRESENT`, `arch=x86_64`, `os=ubuntu 24.04`, `container=not_detected` et `secret_tool=PRESENT`. Une seule différence stoppe la procédure et laisse le gate `PENDING`.

## Phase D — Matrice SystemVault officielle

Après le préflight positif et le déverrouillage du coffre, exécuter sans `sudo` :

```bash
cd /chemin/vers/ForgeLocal
chmod 0755 scripts/run-systemvault-native-gate.sh scripts/check-systemvault-anti-leak.sh
OUT_DIR="$PWD/systemvault-native-evidence" \
FORGELOCAL_VAULT_SERVICE="ForgeLocal.Back01.ReleaseCandidate" \
scripts/run-systemvault-native-gate.sh
```

Le résultat n’est admissible que si `systemvault-matrix.json` contient `true` pour `created_key`, `read_key`, `restart_read`, `created_secret`, `read_secret`, `deleted` et `absent_verified`. Le script refuse déjà root, un conteneur et l’absence de D-Bus ; **ne pas contourner ces refus**.

## Phase E — Cas manuels, anti-fuite et export

Suivre exactement les sections « Révocation externe », « Coffre verrouillé ou permissions insuffisantes » et « Contrôle anti-fuite intégré » du runbook `SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md`. La sentinelle doit être nouvelle, non productive, stockée dans un fichier `0600` hors des données ForgeLocal et supprimée à la fin. Sa valeur ne doit jamais apparaître à l’écran, dans un argument, un historique shell ou un fichier de preuve.

Seuls les éléments suivants peuvent être copiés hors de la VM :

| Fichier | Contrôle attendu |
|---|---|
| `systemvault-host-context.env` | Ubuntu 24.04, `x86_64`, Secret Service, classe de session non root et sans sudo. |
| `systemvault-matrix.json` | Tous les booléens requis à `true`. |
| `SYSTEMVAULT_NATIVE_GATE_STATUS` | Matrice passée et cas manuels explicitement documentés. |
| `systemvault-anti-leak.json` | `anti_leak: true`, sans valeur sentinelle. |

Calculer les SHA-256 des quatre fichiers exportés, verser les copies assainies dans `validation_back01_integration/final/`, puis mettre à jour le gate uniquement avec les métadonnées complètes et une revue indépendante. À ce stade, le statut public reste `PUBLIC_RELEASE_BLOCKED` : les quatre autres gates demeurent obligatoires.

## Phase F — Nettoyage et sécurité

Après export vérifié des preuves, supprimer la sentinelle, supprimer l’accès graphique public et détruire l’instance ainsi que son volume de démarrage. Oracle indique que les instances Always Free inactives peuvent être récupérées ; cette VM doit être considérée comme une machine de qualification éphémère, jamais comme un poste de production [1].

## Références

[1] [Oracle — Always Free Resources](https://docs.oracle.com/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm)

[2] [Oracle — Cloud Free Tier](https://www.oracle.com/cloud/free/)

[3] [Oracle — Free Tier documentation](https://docs.oracle.com/iaas/Content/FreeTier/freetier.htm)

[4] `release/back01-minimal/SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md` et `scripts/run-systemvault-native-gate.sh`, dépôt ForgeLocal, consultés le 2026-08-14.
