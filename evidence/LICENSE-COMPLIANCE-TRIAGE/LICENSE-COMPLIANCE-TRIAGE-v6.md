# LICENSE-COMPLIANCE-TRIAGE — lot séparé

**Décision :** `LICENSE_UNKNOWN_RETAINED_PENDING_OFFICIAL_SOURCE_REVIEW`

**Base contrôlée :** audit licences publié au commit `fb9d20c7ffb241104eede582eee9cbd7a3bbc267`. La branche est indépendante du V6 gelé et ne modifie ni code, ni dépendances, ni wrappers, ni gates.

L’inventaire CycloneDX propre courant contient **749 composants**, dont **3 composants avec licence identifiée** et **746 composants `UNKNOWN`**. Le rapport historique mentionnait 744/741 ; cette différence est conservée et expliquée dans `V6_LICENSE_COMPARE.log`, sans écraser l’inventaire courant.

La matrice individuelle `V6_LICENSE_COMPLIANCE_MATRIX.md` fournit pour chaque ligne le composant, la version, le PURL ou la source à consulter, l’état de licence, l’obligation TBD, la décision `OPEN_REQUIRES_SOURCE_REVIEW` et le propriétaire Legal/OSS. Lorsqu’une licence officielle n’est pas suffisamment établie par le package, sa version et sa source canonique, la valeur reste `UNKNOWN`. Aucune licence n’est déduite du nom du package, d’un transitive connu ou d’une licence supposée.

## Politique de distribution préalable

Aucune release ou distribution publique ne doit inclure un composant `UNKNOWN` sans revue Legal/OSS documentée ou décision d’exception approuvée. La revue doit rattacher le composant à sa source officielle, à la version exacte distribuée, au texte de licence et aux obligations de notice, attribution, copyleft, brevet ou redistribution applicables. Le manifeste de distribution doit ensuite reprendre la décision et les notices correspondantes. Tant que la revue n’est pas complète, la gate de publication reste bloquée.

Le lot ne lance aucun runtime, ne démarre pas T28/T29/T39/T40/T41/T42, Camoufox, proxy réel, cookies, données utilisateur ou SystemVault natif, et ne constitue pas une release.

**Owner :** Legal / OSS review. **Condition de levée :** chaque `UNKNOWN` est résolu par source officielle et version correspondante, ou fait l’objet d’une exception de distribution approuvée individuellement.
