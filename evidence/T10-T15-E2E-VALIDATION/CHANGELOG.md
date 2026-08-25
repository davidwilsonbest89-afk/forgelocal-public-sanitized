# CHANGELOG — T10-T15-E2E-VALIDATION

## 2026-08-25

Baseline brute post-gel créée avant exécution. Une première tentative a échoué avant Playwright car la compilation avait été lancée hors du répertoire module. Une seconde tentative a atteint le Core mais a déclenché un téléchargement de runtime non voulu ; elle a été arrêtée et ses temporaires nettoyés. L’orchestrateur a ensuite été corrigé pour compiler depuis le module, pointer vers `/usr/bin/chromium`, interdire l’auto-update et gérer les groupes de processus.

Le run final a exécuté 7 tests en séquence : 7 pass en 15,6 secondes. Le token temporaire, la base, le run root et les processus Core/Vite ont été supprimés ; les ports 19280 et 3000 sont fermés. Aucun code produit n’a été modifié.
