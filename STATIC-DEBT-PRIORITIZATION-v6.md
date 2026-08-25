# STATIC-DEBT-PRIORITIZATION — lot séparé

**Décision :** `STATIC_DEBT_INDIVIDUALLY_TRIAGED_NO_CODE_CHANGE`

**Base contrôlée :** audit documentaire publié au commit `fb9d20c7ffb241104eede582eee9cbd7a3bbc267`. La branche reste indépendante du V6 gelé et ne change ni Core, ni Dashboard métier, ni wrappers historiques, ni dépendances.

Le dossier contient **34 findings Staticcheck** et **89 findings GolangCI-Lint**, soit 123 entrées individuelles. Pour chaque entrée, `V6_STATIC_DEBT_TRIAGE.md` conserve la règle, le fichier, la ligne, le message, l’impact, la sévérité, le correctif possible, le test de non-régression, le propriétaire et le lot futur envisagé.

Le triage ne vaut pas correction. Aucun finding historique n’est présenté comme résolu et aucun code n’est modifié ici. Les défauts de sécurité ou de fiabilité élevés doivent être revus par un mainteneur Go, corrigés dans un lot séparé, puis couverts par des tests ciblés et une requalification complète. Les autres entrées restent un backlog priorisé afin de ne pas contaminer la baseline V6 gelée.

Le lot ne lance aucun runtime, ne démarre pas T28/T29/T39/T40/T41/T42, Camoufox, proxy réel, cookies, données utilisateur ou SystemVault natif, et ne constitue pas une release.

**Owner :** mainteneurs Go. **Condition de levée :** correction ou justification approuvée par entrée, test de non-régression réussi et nouveau scan archivé.
