# T31 — Contrat des diagnostics Canvas, WebGL et Audio

**Statut :** `T31_CONTRACT_AND_PROJECTED_CONTROLS_IMPLEMENTED_PENDING_REVIEW`
**Baseline :** `t00-t27-complete-20260820` / `69411e65c880d168832a65fc8475cc97d562a9ad`
**Mode :** lecture seule, projection redacted, sans lancement de navigateur ni runtime.

## Périmètre fermé

T31 ajoute uniquement des contrôles stables à la projection de diagnostic d’environnement existante : `canvas-2d`, `canvas-webgl`, `webgl2`, `audio-context` et `offline-audio-context`. Ces contrôles signalent explicitement `UNSUPPORTED` avec la note fixe `runtime observation not implemented`.

Le contrat n’autorise aucun calcul, hash, comparaison ou collecte d’une valeur Canvas, d’un fournisseur WebGL, d’un renderer, d’une extension WebGL, d’un AudioContext, d’un AudioBuffer, d’une fréquence, d’un signal ou d’une empreinte audio. Il ne retourne ni User-Agent, ni adresse locale, ni chemin, ni identifiant de runtime.

## Invariants de sécurité

La réponse reste `PROJECTED_METADATA_ONLY`. Aucun navigateur, Camoufox, proxy réel, cookie, session, port runtime ou système audio n’est démarré. Les noms de contrôles sont des identifiants de contrat, non des observations. L’absence de prise en charge est un état explicite et stable ; elle ne peut pas être transformée en `PASS` par défaut.

Le diagnostic conserve l’identifiant de profil validé et le verdict agrégé existants. Les erreurs de profil inconnu ou d’identifiant invalide restent refusées avant toute observation. L’API reste loopback-only, authentifiée et redacted selon le contrat T30.

## Contrôles et critères

| Contrôle | État T31 | Observation autorisée | Observation interdite |
|---|---|---|---|
| `canvas-2d` | `UNSUPPORTED` | Présence du contrôle et note fixe | Pixels, dimensions, hash, bruit, valeur de rendu |
| `canvas-webgl` | `UNSUPPORTED` | Présence du contrôle et note fixe | Renderer, vendor, extensions, paramètres, hash |
| `webgl2` | `UNSUPPORTED` | Présence du contrôle et note fixe | Contexte réel, capacités, shader, résultat |
| `audio-context` | `UNSUPPORTED` | Présence du contrôle et note fixe | Fréquence, latence, contexte, signal, empreinte |
| `offline-audio-context` | `UNSUPPORTED` | Présence du contrôle et note fixe | AudioBuffer, rendu offline, hash, bruit |

## Tests requis et réalisés

Les tests T31 vérifient la présence des cinq contrôles, leur ordre déterministe, leur état `UNSUPPORTED`, leur note fixe et l’absence de valeurs brutes dans le JSON sérialisé. Les tests API hérités vérifient en outre l’authentification, le mode projeté, le refus de profil inconnu et l’absence de chaînes sensibles.

La qualification de la branche T31 exécute `go test -count=1 -race ./...`, `go vet ./...`, `go build ./...`, `git diff --check` et Gitleaks sur le diff binaire exact depuis la baseline. Aucune qualification de navigateur, de WebGL, d’audio ou de runtime réel n’est revendiquée.

## Sortie et gate

T31 peut être accepté comme incrément local de projection et de contrat si les tests et scans ci-dessus restent verts. Il ne lève aucun gate de release, de provenance, de SystemVault ou de Camoufox. T32 peut commencer après publication de la branche T31 et vérification de son commit distant.
