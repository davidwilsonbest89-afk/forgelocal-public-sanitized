# Registre de clôture canonique T27-R1

| Élément | Empreinte SHA-256 |
|---|---|
| Manifeste du kit `SHA256SUMS` | `5e5defbb24b3314bde1c43802b3c81c900603823cac82ab702eb1ec7d3021dab` |
| Bundle CR-01 LFS | `ab57d11f59fa0d13db7ded91e0a468413a094104ae8778761df919d103cbd271` |
| Kit de preuve CR-01 LFS | `b1cdcf44d081ef6e30f2f5524156765c88d1dc2a9e24a08048cad025935d741b` |

Le kit comprend 50 fichiers, sa chaîne CR-01 à CR-05, la gouvernance, le verrou de toolchain et la clôture SBOM/provenance CR-09. Les artefacts CR-01 de plus de 100 Mo sont suivis par Git LFS.

La reprise exige la vérification de `SHA256SUMS`, des sidecars puis des bundles dans l’ordre défini par `evidence/CHAIN.md`.
