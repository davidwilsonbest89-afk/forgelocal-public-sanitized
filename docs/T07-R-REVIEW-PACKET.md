# T07-R — Paquet de revue indépendante redacted

Ce paquet prépare uniquement une **vérification indépendante de complétude** pour le candidat Camoflox. Il ne contient pas le code Camoflox, l’archive source, une valeur d’alerte, un secret, une clé, un token ou une licence privée brute. Il ne constitue pas une décision `PASS`, n’autorise aucun portage et ne permet pas de commencer T08.

> Le statut de travail reste `T07_PROVENANCE_BLOCKED_PENDING_EVIDENCE`. `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, la suspension du pilote et le gel de T08 demeurent inchangés.

## Pièces du paquet

| Pièce | Destinataire | Finalité | Contenu interdit |
|---|---|---|---|
| `t07-r-camoflox-owner-attestation.prepared.json` | Détenteur des droits puis relectrice | Attestation redacted à compléter et à confirmer | Source, archive, valeur d’alerte, secret, clé, token |
| `T07-R-OWNER-COMPLETION-GUIDE.md` | Détenteur des droits | Remplissage des champs manquants avec références contrôlées | Documents de preuve bruts dans Git |
| `T07-R-INDEPENDENT-REVIEW-RESPONSE.md` | Relectrice indépendante | Retour redacted avec les six couvertures de revue et la décision de triage | Décision de déblocage ou contenu sensible |
| `T07-R-MESSAGE-TO-REVIEWER.md` | Détenteur des droits | Message prêt à envoyer à `@hajarbenmlih91-cloud` | Données privées non nécessaires |

L’attestation préparée est conservée exclusivement sous `/home/ubuntu/forgelocal-private-evidence/t07-r-inbox/`. Ce répertoire privé est exclu du dépôt Git. Il convient de partager avec la relectrice uniquement une copie redacted du JSON, après complétion par le détenteur des droits.

## Chaîne de contrôle attendue

| Étape | Responsable | Résultat attendu | Effet sur T07 |
|---:|---|---|---|
| 1 | Détenteur des droits | Renseigne les références et sa décision de triage, sans contenu sensible | Attestation encore non validée |
| 2 | Relectrice indépendante | Vérifie les six domaines et inscrit ses booléens JSON réels | Attestation prête pour contrôle de complétude uniquement |
| 3 | ForgeLocal | Exécute le validateur sur le JSON redacted reçu | Aucune décision `PASS` automatique |
| 4 | Revue autorisée | Évalue les preuves et les gates `PROV-01`, `PROV-04`, `PROV-06` | T07 reste bloqué jusqu’à décision explicite |

La concordance des décisions de triage est obligatoire. Si l’une des décisions manque, est invalide ou diffère de l’autre, le résultat opérationnel est `UNKNOWN` et T07 demeure bloqué. Une redistribution `not_granted` n’empêche pas l’audit passif mais impose `future_distribution=blocked`.

## Document de référence interne

Les exigences de complétude sont décrites dans [`T07-R-CHECKLIST.md`](./T07-R-CHECKLIST.md). Le contrôle automatisé redacted est [`validate-t07-r-attestation.mjs`](../scripts/validate-t07-r-attestation.mjs). Ces références définissent un contrôle de forme et de cohérence ; elles ne remplacent pas l’examen humain des preuves sous contrôle d’accès.
