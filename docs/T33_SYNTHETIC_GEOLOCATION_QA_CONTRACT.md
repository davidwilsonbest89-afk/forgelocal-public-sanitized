# T33 — QA géolocalisation synthétique

**Statut :** `T33_SYNTHETIC_GEOLOCATION_QA_APPROVED_VERIFIABLE_LOCAL`
**Baseline fonctionnelle :** T32 distant `d7279e81dd724ba2278a65838bc65aaa16912007`
**Mode :** fixtures synthétiques uniquement, sans position réelle.

## Périmètre fermé

T33 fournit un validateur pur pour des fixtures de coordonnées synthétiques. Il refuse les valeurs hors des bornes latitude `[-90, 90]`, longitude `[-180, 180]`, ainsi que `NaN` et `Inf`. Les résultats sont triés par nom de fixture et ne contiennent jamais les coordonnées.

Le package n’a aucune dépendance réseau, navigateur, runtime, profil, cookie, proxy, persistance ou stockage. Il ne reverse-geocode pas, ne consulte aucun fournisseur externe et ne déduit aucune position réelle.

## Résultat redacted

Chaque résultat contient seulement le nom contrôlé, `PASS` ou `FAIL`, le mode `SYNTHETIC_FIXTURES_ONLY` et une raison fixe pour un refus de fixture. Les valeurs latitude/longitude, adresses, villes, pays et traces de fournisseur sont exclues.

## Critères d’acceptation

| Critère | Attendu |
|---|---|
| Bornes | Coordonnées valides acceptées ; hors bornes refusées |
| Numérique | `NaN` et `Inf` refusés fail-closed |
| Déterminisme | Fixtures triées par nom |
| Redaction | Coordonnées absentes du JSON de résultat |
| Isolement | Aucun réseau ou runtime réel |
| Qualification | Tests race ciblés, puis suite Go globale, vet, build, format et Gitleaks |

Ce lot ne constitue pas une géolocalisation réelle et ne lève aucun gate de release ou de runtime. T34 peut commencer après publication et vérification du commit T33.
