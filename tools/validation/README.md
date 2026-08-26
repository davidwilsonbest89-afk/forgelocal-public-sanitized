# Runner Playwright de validation locale

`integrated_smoke_runner.mjs` exécute le parcours Dashboard → création de profils → configuration et affectation de proxies synthétiques → session Core → navigation via proxy → refus fail-closed après arrêt du proxy → isolation et révocation.

Le runner est strictement limité à des destinations HTTP loopback et à un proxy synthétique local. Il utilise le Chromium système autorisé par le protocole de validation ; il ne constitue pas une validation Camoufox, SystemVault natif, Docker/Buildx, proxy commercial, cookie réel ou site externe.

## Exécution

Depuis la racine du dépôt, installer/présenter le `playwright-core` local utilisé par le Dashboard, exporter les variables de `integrated_smoke_runner.env.example` avec des valeurs temporaires du répertoire de données de test, puis exécuter :

```sh
node tools/validation/integrated_smoke_runner.mjs > /tmp/forgelocal-smoke-redacted.json
```

Le token Core et les credentials proxy ne doivent jamais être écrits dans le fichier de sortie. Le runner journalise uniquement des indicateurs de présence, des statuts et des erreurs nettoyées. Les secrets sont fournis au processus par l’environnement et ne sont pas codés dans le dépôt.

Les variables `SMOKE_ALPHA_PROXY_*` et `SMOKE_BETA_PROXY_*` servent au navigateur Chromium ; les proxies synthétiques du parcours Dashboard restent configurés sans secret dans le registre local. `SMOKE_BAD_PROXY_EXPECTED_*` sert à fabriquer une tentative volontairement incorrecte en ajoutant le suffixe `-wrong` au client.

Le résultat final doit conserver `external_forward_observed=false`. Une occurrence de forward externe force l’échec du verdict strict. Les services temporaires doivent être arrêtés par le bloc `finally`, puis les ports loopback doivent être vérifiés séparément.

Le runner est une preuve d’exécution locale reproductible, pas une preuve de readiness de production.
