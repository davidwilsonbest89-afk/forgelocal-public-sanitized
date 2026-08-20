# ForgeLocal

ForgeLocal est un gestionnaire local-first de profils et d’environnements de navigation, destiné aux tests de compatibilité, à l’assurance qualité et aux workflows d’automatisation expressément autorisés. Le Core Go est l’unique source de vérité ; le Dashboard React est un client local.

## État de cette publication

Cette publication correspond à la lignée locale T23, au commit de reprise `df969533cdd446be41d868bebea2c8a106f543d5`. Elle inclut les fondations Notes/Custom Fields, Templates, History, Archive/Restore et les contrats documentaires associés. T24 n’est pas inclus.

Cette publication ne promet ni anonymat, ni absence de détection, ni contournement de mécanismes tiers. Elle exclut le vol de données, le credential/cookie theft, la fraude, le CAPTCHA bypass, le contournement anti-bot/anti-fraude, l’accès non autorisé et le contournement d’un mécanisme de sécurité.

## Démarrage local

```bash
GOTOOLCHAIN=local /usr/local/go/bin/go test -count=1 -race ./...
GOTOOLCHAIN=local /usr/local/go/bin/go vet ./...
GOTOOLCHAIN=local /usr/local/go/bin/go build ./...
```

Le Dashboard est dans `forge-dashboard/`. Consulter `docs/local-quickstart.md`, `docs/architecture.md` et le CDC inclus pour les conventions du projet.

## Limites connues

Cette publication publique ne contient aucun document de provenance privé, archive privée T07, secret, cookie, session, attestation privée ou artefact de continuité privé. Les gates de release restent `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN` et `NATIVE_SYSTEMVAULT_NOT_TESTED`.

## Licence

Voir [LICENSE](LICENSE).
