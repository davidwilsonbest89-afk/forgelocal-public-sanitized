# T24-SR — Contrat de remédiation des dépendances

## Objectif fermé

Supprimer les versions de dépendances signalées par l’audit indépendant T24, sans modifier le comportement métier T24, les routes, les modèles, le Dashboard, le runtime, les proxies ou les secrets.

| Module | Version baseline | Version cible | Justification |
|---|---:|---:|---|
| `github.com/go-chi/chi/v5` | `v5.1.0` | `v5.3.1` | Correctifs Chi postérieurs à la version vulnérable relevée par Govulncheck. |
| `golang.org/x/net` | `v0.53.0` | `v0.58.0` | Dernière version disponible compatible avec Go `1.25`, au-delà des versions de correction indiquées par les avis observés. |

## Autorisé

La modification de `go.mod`, `go.sum` et de versions indirectes strictement imposées par le solveur de modules est autorisée. Aucun code de production ne doit changer, sauf incompatibilité de compilation directement causée par ces mises à niveau ; dans ce cas, l’écart doit être documenté et le travail arrêté avant tout élargissement.

## Interdit

T25, nouvelles fonctionnalités, Dashboard, extension de l’API, runtime, Camoufox, proxy réel, secrets réels, coffre natif, migration de données, release et changement de gate sont hors périmètre.

## Qualification requise

`go mod tidy`, diff de dépendances, tests T24 sous `-race`, suite Go globale `-race`, vet, build, diff-check, Gitleaks delta/extraction, Trivy secrets, Gosec delta, Govulncheck ciblé et OSV-Scanner du module sont requis. Un SBOM doit confirmer les versions finales. Le package de preuves doit passer manifestes, bundle, clone neuf et `git fsck --full`.
