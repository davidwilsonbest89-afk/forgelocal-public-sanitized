# GOSEC-R7 — périmètre de risque initial

## Autorité

La branche R7 a été créée depuis le HEAD distant R6 découvert dynamiquement : `3656dbad4bfef0381e1f9d837271d293ecffe292`. La baseline R6-C vérifiée depuis un clone neuf utilise le JSON `evidence/GOSEC_R6/R6_C/gosec_r6c_after_postcommit.json`, SHA-256 `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a`. Le nouveau scan source-only R7 produit exactement le même SHA-256 et la même distribution.

| Règle | Findings R7 |
|---|---:|
| G101 | 1 |
| G115 | 3 |
| G204 | 5 |
| G302 | 5 |
| G304 | 0 |
| G305 | 1 |
| G404 | 17 |
| G703 | 7 |
| G704 | 7 |
| **Total** | **46** |

## Périmètre fermé

R7 peut analyser les 46 findings courants par fichier, fonction, ligne, entrée, chemin exécuté, actif, préconditions, impacts CIA, plateforme, garde, test négatif, résultat Gosec, classification, action et priorité. Les 11 G304 et les deux G703 supprimés par R6-A restent uniquement dans l’historique et ne doivent pas être recréés.

Aucun code ne doit être modifié avant la fin de la matrice et des tests d’atteignabilité. Une correction n’est admissible que si une atteignabilité ou un impact réel est démontré. Un garde existant ne clôt pas un finding Gosec; la classification correcte peut rester `MITIGATED_CONTROL_SCANNER_OPEN`.

## Données et environnements autorisés

Les tests sont limités au loopback, aux services synthétiques, aux profils jetables, aux bases temporaires et aux tokens temporaires redacted. Aucun compte, cookie, secret, proxy commercial, site externe ou donnée utilisateur ne doit être utilisé. Windows/macOS, Camoufox, SystemVault natif et Docker/Buildx ne doivent pas être simulés.

## Invariants

T28 reste fermé fonctionnellement; T29 ne démarre pas; T31–T38 ne sont pas modifiés; les packages R6-A/B/C ne sont pas recréés ni réécrits; `evidence/SMOKE_INTEGRATED_PROXY/` reste local et non suivi. Les invariants restent `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false` et `FORGELOCAL_PRODUCTION_READY=false`.
