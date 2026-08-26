# GOSEC-R4 — Lot D

## Périmètre

Le lot R4-D couvre les subprocess GUI/OS et les flux réseau/WebSocket encore signalés par Gosec. Aucun navigateur natif, Camoufox, Darwin GUI, proxy réel, cookie réel ou compte externe n’a été exécuté.

## Résultat

Le scan R4-C final rattache 5 findings G204 et 7 findings G704 au périmètre R4-D. Les contrôles applicatifs sont présents et testés localement, mais les findings restent visibles dans le scanner et ne sont pas déclarés clos.

| Groupe | Résultat | Explication |
|---|---:|---|
| G204 | 5 ouverts | Les commandes OS sont des lanceurs littéraux (`open`, `xdg-open`, `rundll32`) ou `xattr` avec timeout. Le chemin Darwin natif n’a pas été exécuté. |
| G704 | 7 ouverts | Les cibles CLI et WebSocket sont validées loopback avant connexion, avec ports, schémas, chemins, userinfo, query, fragment et redirections contrôlés. |

Les tests disponibles couvrent les limites positives et négatives des URLs CLI, le refus des hôtes externes et des redirections, ainsi que la validation stricte de l’endpoint WebSocket Playwright. Ils ne remplacent pas une exécution native Darwin ou une validation de production.

## Verdict du lot

```text
GOSEC_R4_D_MITIGATED_WITH_OPEN_FINDINGS
GOSEC_R4_PARTIAL_ENVIRONMENT_UNAVAILABLE
```

La matrice détaillée est `R4_D_CLASSIFICATION.tsv`. Aucun `nosec`, `nolint`, skip global ou allowlist n’a été ajouté.
