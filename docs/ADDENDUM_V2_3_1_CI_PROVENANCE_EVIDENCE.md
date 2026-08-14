# Addendum v2.3.1 — Exécution CI et preuve de provenance

**Identifiant :** `FL-ADD-2.3.1-20260814`
**Statut :** complément opérationnel de l’addendum v2.3.
**Objet :** rendre obligatoire le contrôle de provenance avant les tests Core et archiver le registre JSON validé avec son empreinte.

## Contrat CI

Le workflow `.github/workflows/ci.yml` installe Node 22 avant tout contrôle de provenance, puis exécute obligatoirement :

```text
node scripts/check-component-rights.mjs
node scripts/test-component-rights.mjs
node scripts/create-component-rights-evidence.mjs
```

Un échec du registre, de ses fixtures ou de la génération de preuve arrête le job avant les tests Core. La preuve ne constitue pas une autorisation de release ; son champ `release_status` doit rester `PUBLIC_RELEASE_BLOCKED` tant que les gates publics ne sont pas levés.

## Artefact archivé

Après validation, le pipeline archive `component-provenance-<commit>` pour 90 jours. L’artefact contient exclusivement :

| Fichier | Contenu |
|---|---|
| `component-rights-register.json` | Copie exacte du registre JSON contrôlé durant le job. |
| `component-provenance-evidence.json` | Identifiant de registre, SHA-256, commit source et commandes de contrôle exécutées. |

Les emails d’autorisation, tokens, secrets, valeurs proxy, chemins d’hôte et données personnelles restent exclus de l’artefact. Une preuve n’est archivée qu’après le succès des contrôles de provenance.

## Invariants

Cet ajout ne modifie aucun artefact BACK-01, ne réactive pas le pilote, n’active aucun runtime ni aucune mutation UI, et ne modifie pas `SCAN_BLOCKED_UNKNOWN`, SystemVault, les cinq gates publics ou `PUBLIC_RELEASE_BLOCKED`.
