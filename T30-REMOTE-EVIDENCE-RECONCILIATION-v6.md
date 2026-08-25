# T30-REMOTE-EVIDENCE-RECONCILIATION — lot séparé

**Décision :** `T30_REMOTE_EVIDENCE_RECONCILED_PENDING_INDEPENDENT_REVIEW`

T30 est maintenant rattaché à une chaîne de preuves GitHub précise, sans lancer T30 ni aucun runtime. La référence canonique est la branche [`audit/t00-t42-v6-findings-remediation`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/audit/t00-t42-v6-findings-remediation), dont le commit de packaging publié est `727e94bcf2bf40bc9388ee2df82bdb241962e5b2`. Le tag annoté [`t00-t42-v6-local-qualified-2026-08-25`](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/t00-t42-v6-local-qualified-2026-08-25) pointe sur le commit de gel `999374d99b7996504ba91e421850a2fe84afb78d`.

## Références d’artefacts

| Élément | Référence | Hash SHA-256 / contrôle |
|---|---|---|
| ZIP du gel V6 | `evidence/forgelocal-t00-t42-v6-local-qualified-freeze.zip` | `eb61ca0a42aad8afc8ba3a6088855acfd61dae0887bde339fad371e46feee264` |
| Bundle delta du gel | `evidence/forgelocal-t00-t42-self-validation-v6-freeze.delta.bundle` | `7a324a93cffa2be2447371f92b3f3a1365d00bcc512fd606bb4bd39ea1384d9c` |
| Objet du tag annoté | `t00-t42-v6-local-qualified-2026-08-25` | `310c10e7541bbec7828226ce1049df56da11ce0b` |
| Sidecars | `*.portable.sha256` correspondants | validés dans `V6_FINAL_PUBLICATION_VERIFY.log` |
| Manifeste | manifeste interne non auto-référentiel du ZIP | `sha256sum -c` et `unzip -t` réussis |
| Clone public | branche V6 publiée | HEAD `727e94b...`, `git fsck --full=0` |

La vérification publique a confirmé le SHA distant de branche, le tag et sa cible, le téléchargement LFS ciblé des artefacts, les hashes, l’extraction ZIP, la vérification du bundle, Gitleaks et le fsck final. Le log complet est archivé dans ce lot. La qualification reste en attente de revue indépendante ; elle ne constitue ni release ni levée de gate.

## Statut T30

Le statut initial `PENDING_REMOTE_EVIDENCE_RECONCILIATION` est remplacé pour cette chaîne documentaire par `T30_REMOTE_EVIDENCE_RECONCILED_PENDING_INDEPENDENT_REVIEW`. Les gates de produit et de publication restent inchangées. Aucun test T30 opérationnel, runtime ciblé, navigateur, proxy, cookie, donnée utilisateur, SystemVault natif, migration, Camoufox, T28, T29, T39, T40, T41 ou T42 n’est démarré.

**Owner :** responsable des preuves distantes / revue indépendante. **Condition de levée :** acceptation indépendante de la chaîne d’artefacts et de la distinction gel V6 versus release.
