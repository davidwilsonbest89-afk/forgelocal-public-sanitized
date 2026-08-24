# T34 — Diagnostics matériel read-only et redacted

**Statut :** `T34_HARDWARE_PROJECTION_APPROVED_VERIFIABLE_LOCAL`
**Prérequis :** T33 distant `693632791041fde14db14ec8982b8bff1060a8d3`
**Mode :** projection déclarative, sans sondage de l’hôte.

## Périmètre fermé

T34 expose un package pur qui retourne une liste déterministe de capacités matérielles projetées : `cpu`, `memory`, `storage`, `gpu`, `display`, `network-adapters` et `thermal`. Chaque capacité est `UNSUPPORTED` avec la note fixe `host observation not implemented`.

Le package ne lit ni `/proc`, ni `/sys`, ni `cpuinfo`, ni `meminfo`, ni numéro de série, ni hostname, ni adresse MAC, ni identifiant machine. Il n’utilise aucun réseau, navigateur, runtime, cookie, proxy, stockage ou secret.

## Invariants de redaction

La réponse contient seulement la version de diagnostic, le mode `PROJECTED_METADATA_ONLY`, les noms contractuels, l’état et une note fixe. Aucune valeur matérielle, mesure, température, identifiant ou chemin hôte ne peut être déduite du résultat. L’état `UNSUPPORTED` est conservé fail-closed et ne devient pas `PASS` par défaut.

## Critères d’acceptation

| Critère | Attendu |
|---|---|
| Couverture | Les sept capacités contractuelles sont présentes |
| Déterminisme | Ordre et valeurs identiques à chaque appel |
| Redaction | Aucun identifiant ou chemin hôte dans le JSON |
| Isolement | Aucun accès système, réseau ou runtime |
| Qualification | Tests race ciblés, suite globale, vet, build, format et Gitleaks |

T34 reste une projection locale non invasive et ne constitue pas une qualification matérielle réelle. T35 peut commencer après publication et vérification du commit T34.
