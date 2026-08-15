# T07 — Qualification de provenance Camoflox

**État :** `T07_PROVENANCE_BLOCKED_PENDING_EVIDENCE`. Cette qualification est passive. Elle n’importe, ne lance et ne porte aucun code de l’archive Camoflox.

L’archive fournie `camoflox-FINAL.zip` est identifiée par l’empreinte SHA-256 `dcf668d463bccd9a3469a0dcb909f447c4d7672f3322ab4680a004b3ee4851c2`. La référence privée `PRIVATE-RIGHTS-CAMOFLOX-001` déclare une autorisation limitée à l’usage interne et à la modification. Cette autorisation ne rend pas l’archive intégrable : un commit source exact, une licence racine et la classification de l’alerte de sécurité restent requis.

| Module étudié | Empreinte source | Décision T07 | Limite |
|---|---|---|---|
| `lib/concurrency.js` | `b055a3e1c995c3dddca054aa90ce2c0b8ff660237bf96b1f2b168dd5a36085d7` | Revue conceptuelle seulement | Aucun code copié ou porté ; une spécification Go indépendante ne peut être envisagée qu’après T07. |
| `lib/global-action-limiter.js` | `a5f6838be5731ff64836f013b3b27fa280874de43afd817829de92052a487d05` | Revue conceptuelle seulement | Aucun état, queue ou décision de lancement n’est repris. |

Les modules `lib/profile-launch-queue.js`, `lib/process-isolation.js` et `lib/browser-lifecycle.js` sont hors périmètre T07. Ils concernent respectivement la queue, les locks/ports et le cycle de vie runtime, qui ne peuvent être étudiés pour un éventuel contrat ForgeLocal qu’après l’audit T07, dans un lot distinct.

## Contrôles PROV-01 à PROV-07

| Contrôle | Résultat T07 | Justification |
|---|---|---|
| `PROV-01` — droits sur révision exacte | **PARTIEL** | L’archive a une empreinte exacte et une autorisation privée redacted, mais aucun commit source vérifiable. |
| `PROV-02` — dépendances et assets | **PASSIF / PARTIEL** | `package-lock.json` est présent ; la SBOM SPDX compte 710 packages, tandis que l’inventaire verrouillé dénombre 8 dépendances runtime et 4 dépendances de développement. Aucun paquet n’a été installé. |
| `PROV-03` — Core unique | **PASS** | Les candidats restent non intégrés ; Node/Electron, runtime, queue, locks et ports sont exclus. |
| `PROV-04` — sécurité | **BLOQUÉ** | Gitleaks 8.18.4 signale une alerte `generic-api-key` à `tests/smoke.test.js:24`; la valeur n’est pas reproduite et la classification reste `UNKNOWN`. |
| `PROV-05` — fiabilité | **NON ÉVALUÉ** | Les tests et scripts source ne sont pas exécutés ; aucune conclusion de fiabilité n’est prétendue. |
| `PROV-06` — SBOM et notices | **PARTIEL** | Une SBOM de provenance est disponible, mais l’archive ne contient aucune licence racine et les licences transitives n’ont pas été confirmées. |
| `PROV-07` — valeur produit testable | **NON ÉVALUÉ** | Aucun portage Go, aucun runtime et aucune écriture UI ne sont autorisés dans T07. |

La SBOM de provenance [`T07_CAMOFLOX_PROVENANCE_SBOM.spdx.json`](T07_CAMOFLOX_PROVENANCE_SBOM.spdx.json) et l’[inventaire verrouillé des dépendances](T07_CAMOFLOX_DEPENDENCY_INVENTORY.json) proviennent exclusivement des métadonnées `package.json` et `package-lock.json` de l’archive. Ils ne constituent ni une autorisation de dépendance ni une certification de licence.

## Source publique distincte

Le projet public [Camoufox][1] est distinct de l’archive privée Camoflox : il affiche une licence MPL-2.0, alors que l’archive Camoflox auditée ne fournit ni une licence racine ni un commit public pouvant établir ce lien. Camoufox n’est donc ni sélectionné, ni téléchargé, ni intégré dans T07.

> La réussite de T07 requiert au minimum la classification indépendante de l’alerte `generic-api-key`, une provenance de commit vérifiable et des notices de licence complètes. Jusque-là, l’état de l’archive est **non intégrable**.

## Références

[1]: https://github.com/daijro/camoufox "Dépôt public officiel Camoufox"
