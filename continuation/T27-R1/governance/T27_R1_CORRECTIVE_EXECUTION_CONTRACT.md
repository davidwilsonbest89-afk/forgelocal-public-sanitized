# Contrat d’exécution correctif — T27-R1 amendé

> **Gate d’entrée :** ce contrat ne devient exécutable que si `t26-simulated-proxy-provider-2026-08-20` résout à `930003ca95a934fd996c94ae897693ffb6be21fb`, si le bundle/sidecar passe, si un clone neuf reproduit le HEAD et si `git fsck --full` est vert. Sinon : `BLOCKED_BASELINE_MISMATCH`, aucune écriture.

## AM-02 — Périmètre Playwright

« Navigateur réel interdit » désigne les runtimes produits et profils métier ForgeLocal. Playwright Chromium headless est autorisé uniquement pour tester le Dashboard sur loopback, sans profil ForgeLocal, session métier ou qualification runtime.

## CR-01 — Runtime et Camoufox

Camoufox est désactivé par défaut. Defaults, CLI, API, MCP, workflow, téléchargement et launcher retournent `CAMOUFOX_EXECUTION_NOT_AUTHORIZED` quand `camoflox_execution_authorized=false`. Les tests emploient une ProcessFactory sentinelle qui ne doit jamais être appelée.

Un runtime est refusé si son registre est absent, illisible, stale, failed ou incohérent. La décision lie runtime ID, version, hash binaire, OS/architecture et politique produit. Hash et exécutabilité sont vérifiés immédiatement avant `ProcessFactory.Start`; un changement est refusé/audité. Aucun binaire runtime ne peut être démarré dans les tests.

## CR-02 — Workflow et loopback

La route workflow est supprimée ou déplacée sous `/api/v1/workflows/run`, avec Bearer, IP distante loopback et origine locale. Elle retourne `WORKFLOW_EXECUTION_NOT_AUTHORIZED` tant que l’automation n’est pas autorisée. Aucun workflow n’est interprété ou exécuté.

Le mode MVP refuse avant `net.Listen` les binds `0.0.0.0`, `::`, IP non-loopback et localhost dont une résolution est non-loopback. Docker ne remplace jamais loopback par une adresse publique. Les routes tardives héritent auth/IP/origine. Les tests réseau externes utilisent namespace ou conteneur séparé; modifier seulement `RemoteAddr` n’est pas une preuve réseau.

## CR-03 — Logs, cookies et MCP

Le collecteur est absent des builds qualifiés. En développement, il applique avant buffer une allowlist d’en-têtes sûrs, ne collecte aucun body et subit une seconde redaction serveur. Authorization, Cookie et Set-Cookie ne sont jamais copiés en mémoire.

Les endpoints et outils MCP de cookies bruts sont fail-closed, désactivés et non activables par configuration utilisateur. Ils retournent uniquement des projections `configured`/`disabled`, jamais cookie, token ou `secret_ref`. Les arguments proxy n’acceptent que des références de secret validées. T25 reste limité aux fixtures synthétiques `fixture:` et aux digest redacted.

## CR-04 — Dashboard et supply chain

Critical est interdit. Une high temporaire exige non-atteignabilité prouvée dans le build distribué, propriétaire, ticket, contrôles compensatoires, échéance et revalidation. Le lockfile ne change que pour une mise à niveau explicitement approuvée; SBOM avant/après est obligatoire.

`refreshCoreSnapshot` ne lit le Core qu’après bootstrap autorisé. `AutomationPanel` ne peut déclencher une boucle ; polling borné et annulable. Les tests Playwright/axe inspectent console, réseau et traces à la recherche d’une sentinelle complète et de ses préfixes 8/12/16.

## CR-05 à CR-09 — Gates suivants

CR-05 impose API `/api/v1`, OpenAPI, alias legacy protégés ou `404`/`410`, et comparaison routeur vivant/OpenAPI. CR-06 est documentaire : schéma/machine d’état externe seulement, sans écrire de store. CR-07 requiert une autorisation écrite distincte. CR-08 est `NOT_TESTED` en attente d’environnement natif. CR-09 traite CI, provenance, SBOM/notices, `SECURITY.md`, inventaires et publication ; il ne publie rien.

## Toolchain et preuve

`TOOLCHAIN.lock` est normatif. Toute divergence produit `NOT_VERIFIED_TOOLCHAIN_MISMATCH`. Chaque CR crée `BASELINE_DISCOVERY_RAW.log`, exécute tests dynamiques négatifs, `-race`, vet, build, scans, SBOM, bundle, clone neuf, sidecar, manifeste, extraction et copie canonique. Timeout ou log manquant : `NOT_VERIFIED_TOOL_TIMEOUT`, jamais PASS.

## Invariants

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent actifs. Aucun runtime, Camoufox, proxy réel, fournisseur réel, cookie réel, secret réel, migration utilisateur ou release n’est autorisé.
