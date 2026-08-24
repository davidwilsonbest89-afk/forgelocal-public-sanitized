# T40 — Intégration sous gates

**Statut :** `BLOCKED`
**Parent :** T39 `08061f1e55a0bba01a4a28df66dc852e8f345ade`

T40 ne fusionne pas les lots non autorisés dans un runtime. Il définit uniquement une vérification documentaire des prérequis : branches et commits T31–T38 vérifiés, T39 bloqué, et gates permanentes conservées. Toute intégration exécutable dépend d’une décision produit pour T28/T29/T39.

| Précondition | État |
|---|---|
| T31–T38 localement vérifiables | Oui, avec preuves de branches séparées |
| T39 secret import/export | `BLOCKED` |
| SystemVault natif | `NOT_TESTED` et interdit |
| Camoufox/proxy/cookie/runtime | Non exécutés |
| Release | `PUBLIC_RELEASE_BLOCKED` |

Aucune modification de configuration ne lève `camoflox_execution_authorized=false`, `t08_authorized=false` ou `release_authorized=false`.
