# T07-R — Chaîne de liaison de provenance privée

**État :** `PENDING_EXTERNAL_EVIDENCE`. Ce document organise la remédiation du blocage T07. Il ne sélectionne aucun module, ne porte aucun code et n’autorise pas T08.

## Principe

Une source privée ne doit pas être publiée pour être vérifiable. En revanche, une relecture autorisée doit pouvoir relier, sans ambiguïté et sans secret dans Git, la révision source étudiée à l’éventuelle décision future de ForgeLocal.

> La preuve complète reste dans un espace d’accès contrôlé. Git ne contient que les identifiants non sensibles, les hashes, les statuts et les décisions redacted nécessaires à une revue indépendante autorisée.

## Chaîne obligatoire

| Maillon | Donnée exigée | Stockage autorisé | État actuel |
|---|---|---|---|
| 1. Révision privée attestée | Identifiant de dépôt ou d’espace privé, commit exact **ou** snapshot immuable attesté | Espace privé contrôlé ; identifiant redacted dans Git | PENDING |
| 2. Snapshot étudié | SHA-256, date d’acquisition, arborescence ou manifeste d’archive | Hash et manifeste redacted dans Git ; archive hors Git | Archive hashée ; lien source PENDING |
| 3. Droits | Détenteur, portée : usage interne, modification, redistribution éventuelle, obligations | Document privé ; identifiant, date et revue redacted dans Git | Référence limitée existante ; portée complète PENDING |
| 4. Licence et notices | Licence racine ou accord attesté, notices des dépendances sélectionnées | Source/document privé ; références redacted et SBOM dans Git | PENDING |
| 5. Triage sécurité | Ticket redacted, deux décisions indépendantes et preuve de rotation si nécessaire | Ticket privé ; identifiant et résultat sans valeur dans Git | `UNKNOWN` / BLOCKED |
| 6. Décision de module | `porter`, `réimplémenter` ou `écarter`, pour **un seul** module hashé | Registre Git après levée des P0 | PENDING |
| 7. Éventuel lot futur | Commit ForgeLocal, tests et archive de preuves | Git et archive redacted après autorisation distincte | Interdit avant approbation T07 |

## Conditions de cohérence

La révision source attestée et le snapshot étudié doivent désigner le même contenu. Le détenteur ou le mandataire qui confirme les droits doit être identifié dans la preuve privée. Les deux relecteurs autorisés vérifient indépendamment le hash, la portée des droits et le résultat de triage ; leurs identités peuvent être consignées sous leurs identifiants GitHub, sans adresse e-mail ni donnée personnelle additionnelle.

Un changement de commit, de snapshot, de hash, de licence, de module ou de résultat de triage invalide toute décision antérieure. Il exige un nouveau scan du snapshot et une nouvelle archive T07.

## Interdictions maintenues

Il est interdit de déposer l’archive privée, le contenu source, une clé, un token, une capture contenant la valeur détectée ou le document de droits brut dans le dépôt ForgeLocal. Il est également interdit de démarrer T08, de créer une queue, un lock, un port, un runtime, un lancement ou un portage Go tant que `PROV-01`, `PROV-04` et `PROV-06` ne sont pas **PASS**.
