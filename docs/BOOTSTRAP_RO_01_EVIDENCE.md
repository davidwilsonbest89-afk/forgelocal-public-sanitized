# BOOTSTRAP-RO-01 — Preuves de validation locale lecture seule

> **État provisoire :** les constats ci-dessous sont assainis. Aucun code de bootstrap, token de session, Bearer principal, en-tête `Authorization`, secret proxy ou chemin utilisateur n’est consigné dans ce document.

## Contexte contrôlé

Le 15 août 2026, le Core ForgeLocal a été démarré dans un répertoire temporaire distinct de tout artefact RC, sur `127.0.0.1`, avec `serve --no-runtime`. Ce mode a empêché toute installation, activation, qualification ou exécution de runtime, y compris Camoufox.

| Élément | Constat assaini |
|---|---|
| Binding du Core | `127.0.0.1:18765` |
| Dashboard local vérifié | `http://127.0.0.1:3001` |
| Mode runtime | Tous les runtimes désactivés ; aucune qualification effectuée |
| Prévol CORS local | `204`, origine exacte loopback acceptée |
| Prévol CORS distant | `403 CORS_ORIGIN_NOT_ALLOWED` |
| Binaire exercé | SHA-256 `56f852e97c89f70a0e4664fbd28a69c5515b7068ca9df79f5da444f080fd1113` |
| Outil Go | `1.25.13`, avec `GOTOOLCHAIN=local` |

## Matrice d’acceptation

| # | Contrôle | Preuve assainie | Résultat |
|---:|---|---|---|
| 1 | Code à usage unique, format et TTL | La CLI a émis un code conforme à 64 caractères hexadécimaux avec TTL de 600 secondes ; la valeur n’a jamais été imprimée. | **PASS** |
| 2 | Échange depuis loopback | `POST /api/v1/readonly/session/bootstrap` sur `127.0.0.1` a retourné une session courte de 900 secondes. | **PASS** |
| 3 | Rejeu du code | Le même code a retourné `401 INVALID_BOOTSTRAP_CODE` après le premier échange. | **PASS** |
| 4 | Expiration réelle | Après une attente supérieure au TTL de 600 secondes sur le binaire final, l’échange a retourné `401`. | **PASS** |
| 5 | Refus hors loopback | L’échange tenté via l’adresse IPv4 non loopback de contrôle a retourné `403 LOOPBACK_REQUIRED`. | **PASS** |
| 6 | Token uniquement mémoire | Le client TypeScript a validé l’absence dans l’URL et les stockages persistants ; l’observation navigateur a confirmé `localStorage=0` et `sessionStorage=0`. | **PASS** |
| 7 | Bearer absent des logs | Les sorties Core observées contiennent seulement méthode, chemin, source, statut, taille et durée ; aucun en-tête `Authorization`, Bearer, code ou token n’y apparaît. | **PASS** |
| 8 | Invalidation après `401` | Le harness client a forcé une réponse `401` : le client a effacé sa fermeture mémoire et a restauré l’état déconnecté. | **PASS** |
| 9 | Lectures redacted | `health`, `summary` et `profiles` ont répondu avec le token court ; les champs interdits contrôlés par le harness sont absents. | **PASS** |
| 10 | Aucune écriture | Une écriture de profil avec token court a été refusée ; le nombre de profils est resté `0 → 0`. Aucun runtime n’a été installé, activé ou lancé. | **PASS** |

## Observations navigateur locales

Le dashboard a été ouvert depuis une origine loopback réelle. Avant toute saisie de code, l’inspection a confirmé une URL sans token ni fragment de session, avec `localStorage` et `sessionStorage` vides. Le formulaire de connexion utilise un champ `autocomplete="off"` et le code ne peut être conservé qu’en état React transitoire.

Le prévol vers le Core a répondu avec les en-têtes CORS limités à l’origine `http://127.0.0.1:3000` durant le contrôle programmatique. La requête réelle de validation client a ensuite atteint exclusivement les endpoints lecture seule autorisés : échange bootstrap, `summary` et `profiles`.

## Journal Core observé

Les lignes de journal du Core vues pendant le parcours ne contiennent que la méthode, le chemin, l’adresse source, le statut, la taille et la durée. Elles n’affichent aucun en-tête `Authorization`, Bearer, code de bootstrap ou token de session. La valeur de configuration du token n’est pas imprimée : seul son chemin de stockage local est signalé au démarrage.

## Contrôles de régression et hygiène

Les commandes ciblées Core, CLI et CI ont réussi, notamment `go test -race` sur le broker lecture seule, la vérification des invariants de provenance et les sept scénarios de séparation Sécurité/Release. Le build TypeScript du dashboard a également réussi.

Le scan Gitleaks imposé avec `protect --staged` n’a trouvé aucune fuite parmi les changements mis en scène ; un second scan réel du delta BOOTSTRAP-RO-01 — 11 fichiers Core et dashboard, y compris les fichiers non suivis — a également produit zéro résultat. Ce contrôle de delta ne reclassifie pas l’alerte historique `generic-api-key` du RC gelé.

## Décision

> **BOOTSTRAP_RO_APPROVED** — Les dix contrôles du contrat local lecture seule sont validés sur un Core réellement démarré sur loopback, avec une session mémoire courte, sans mutation et sans activation de runtime.

Cette décision n’autorise ni lancement de Camoufox, ni mutation UI, ni SystemVault, ni modification du RC BACK-01. Le statut de release reste **`PUBLIC_RELEASE_BLOCKED`**, avec `SCAN_BLOCKED_UNKNOWN`, le pilote suspendu et les cinq gates publics en attente.
