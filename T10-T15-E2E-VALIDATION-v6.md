# T10–T15 — validation E2E Dashboard complète post-V6

**Décision technique :** `T10_T15_E2E_EXECUTED_PASS_PENDING_INDEPENDENT_REVIEW`

Ce lot est séparé de la baseline V6 et ne modifie pas le Core, le Dashboard métier, les wrappers historiques ou les gates. Il exécute les suites Playwright principales T10 et T15 contre un Core compilé depuis le commit gelé, un Dashboard Vite loopback, une base temporaire et des fixtures synthétiques.

## Exécution réelle

| Contrôle | Résultat |
|---|---|
| Core | compilé avec Go `go1.25.13`, base temporaire, loopback `127.0.0.1:19280` |
| Dashboard | Vite sur `127.0.0.1:3000`, HTTP 200 vérifié |
| Playwright | 7 tests, `--workers=1`, séquentiel, 15,6 s |
| T15 | W1 session/listing, W2 navigation fail-closed, W3 contenu/capture redacted, W4 fermeture, W5 panneau Dashboard : pass |
| T10 | création valide, listing redacted, port invalide refusé, profil inexistant refusé, unlink mémoire seule, origine hors loopback refusée : pass |
| Données | fixture `file://` synthétique ; aucune donnée utilisateur ou cookie réel |
| Runtime | Chromium système local utilisé pour l’E2E ; Camoufox jamais démarré ; aucun workflow de production |
| Token | valeur temporaire en mémoire/fichier `0600`, non journalisée, redaction appliquée aux logs |
| Cleanup | token, base et répertoire temporaire supprimés ; ports 19280 et 3000 fermés ; aucun processus temporaire résiduel |

Le résultat brut redacted rapporte `7 passed (15.6s)`. Les assertions positives et négatives sont celles des fichiers de test versionnés `tests/proxies-t10.spec.ts` et `tests/automation-t15.spec.ts`, notamment refus d’URL externes, refus de port hors bornes, absence de credential dans le listing, stockage navigateur vide, projection de digest et fermeture de session.

## Limites et statut

Le passage des tests est une validation technique locale, non une approbation produit et non une release. T10 et T15 ne lèvent aucune gate globale. La revue indépendante doit confirmer la reproductibilité du bundle, du clone neuf et des logs redacted. L’exécution n’utilise ni Camoufox, ni proxy réel, ni cookie réel, ni SystemVault natif, ni migration, ni T28/T29/T39/T40/T41/T42.

**Owner suivant :** revue indépendante E2E / mainteneurs Dashboard. **Condition de clôture :** acceptation indépendante du run et de la preuve de cleanup ; toute extension fonctionnelle demeure soumise aux décisions produit séparées.
