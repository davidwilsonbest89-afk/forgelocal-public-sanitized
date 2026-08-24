# T38 — Session lifecycle local et redacted

**Statut :** `T38_SESSION_LIFECYCLE_APPROVED_VERIFIABLE_LOCAL`
**Prérequis :** T37 distant `c0ad051c77e153b3ec4435fbc7ff98e30b96969b`.

T38 fournit un registre mémoire strictement local de transitions de session : `QUEUED`, `STARTING`, `RUNNING`, `STOPPING`, `STOPPED` et `FAILED`. Les snapshots sont triés par clé et ne contiennent qu’une clé de session contrôlée, un état fermé et un code de raison redacted.

| Domaine | Règle |
|---|---|
| Accès | Mémoire locale uniquement, lecture des snapshots et transitions explicites |
| Redaction | Aucun cookie, secret, URL, User-Agent, chemin ou attribut d’hôte |
| Déterminisme | Snapshot trié par `SessionKey` |
| Fail-closed | Clé invalide, état inconnu, transition inconnue ou raison vide rejetés |
| Non-exécution | Aucun navigateur, Camoufox, proxy, réseau, runtime ou SystemVault |

T38 ne prouve pas qu’une session réelle a démarré ou s’est arrêtée. Il s’agit d’un suivi projeté de lifecycle, sans autorisation d’exécution ni de release.
