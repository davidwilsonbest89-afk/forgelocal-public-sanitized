# T28–T42 — correction de portabilité et traçabilité

**Statut de cette correction :** `T28_T42_CLOSURE_EVIDENCE_CORRECTION_READY_FOR_INDEPENDENT_REVIEW` après requalification complète.

Cette branche ne modifie aucun code produit T31–T38 et ne prétend pas réparer une autorisation produit. Les sidecars historiques restent inchangés. Les fichiers `*.portable.sha256` sont des compagnons additionnels contenant uniquement le hash et le nom relatif de l’archive ou du bundle ; ils sont vérifiés depuis le répertoire distribué.

Les fichiers `BASELINE_RECONSTRUCTION_POSTHOC_RAW.log` de T31 à T38 sont explicitement des reconstructions postérieures. Ils ne sont pas présentés comme des `BASELINE_DISCOVERY_RAW.log` contemporains et ne remplacent aucune preuve originale absente. Chaque journal contient UTC, CWD, commandes, sorties, codes de sortie, baseline, tag et versions d’outils.

T30 est ramené à `PENDING_REMOTE_EVIDENCE_RECONCILIATION` : son commit `cbf3a502b3fd37c48798ec67a3a6d4edd5d4a5fb` et ses hashes d’archive sont identifiés, mais aucune branche GitHub dont le head soit ce commit n’a été trouvée. T42 conserve le chemin réel `evidence/T42_DELIVERY.md`, le commit documentaire `0e2b45a0b05ca7e03fd7bb514027c87d618a957f` et le commit de clôture `6489af39a4ac4f91f9f7dc1435f10b2bd10dfdc0`.

Le signal cumulatif Gitleaks `APi=REDACTED` reste conservé comme `SCAN_BLOCKED_UNKNOWN`. Les lots T28, T29, T39, T40, T41 et T42 restent `BLOCKED`. Les gates permanentes ne sont pas levées et aucun runtime réel, Camoufox, proxy, cookie, SystemVault natif, migration utilisateur ou release n’est exécuté.
