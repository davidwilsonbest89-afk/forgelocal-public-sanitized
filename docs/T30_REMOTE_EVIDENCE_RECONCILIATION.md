# T30 — réconciliation de la preuve distante

**Verdict corrigé :** `PENDING_REMOTE_EVIDENCE_RECONCILIATION`
**Commit T30 qualifié :** `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb`
**Commit GitHub accessible :** [cbf3a502](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/commit/cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb)
**Baseline :** tag `t00-t27-complete-20260820`, commit `69411e65c880d168832a65fc8475cc97d562a9ad`

## Preuve disponible

L’archive T30-R3 fournie et vérifiée dans la sandbox contient le head `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb`, un bundle SHA-256 `c4f514ffe4bc24c3adfefff3cfb3a6b07db4a4e19c4c764c16bc9a395c867f14` et un ZIP SHA-256 `c07321fbdf5f16948484264cf9677831cea6f3fd53ee54c0e72273ebea36304d`. Le replay global indépendant a ensuite été exécuté avec succès dans l’environnement spacieux de cette passation, mais le kit T30 canonique n’a pas été rattaché à une branche GitHub dédiée dans la clôture finale.

La recherche distante des branches a été effectuée avec l’API GitHub ; aucune branche dont le head soit `cbf3a502`, `5d100a0d` ou `4f415ebd` n’a été trouvée. Le commit reste directement accessible, mais cette absence empêche de déclarer la chaîne distante complète comme réconciliée.

## Décision fail-closed

T30 est donc maintenu en `PENDING_REMOTE_EVIDENCE_RECONCILIATION`, et non `APPROVED_VERIFIABLE_LOCAL`, jusqu’à ce qu’un propriétaire rattache explicitement le bundle, le ZIP, les sidecars et le kit de qualification à une branche GitHub vérifiable. Aucune archive originale n’est remplacée et aucune autorisation de release n’est déduite de ce document.
