# Observation externe — provenance PPA XtraDeb

**Consulté le :** 14 août 2026

La page publique Launchpad de la PPA XtraDeb Apps indique la PPA `xtradeb/apps` et expose une clé de signature `4096R/5301FA4FD93244FBC6F6149982BB6851C64F6880`. La même empreinte est présente dans la configuration APT locale, dans `/etc/apt/trusted.gpg.d/xtradeb-apps.gpg`.

La PPA publiée pour Ubuntu Noble est présentée sur : <https://launchpad.net/~xtradeb/+archive/ubuntu/apps?field.series_filter=noble>. Le dépôt public associé est : <https://ppa.launchpadcontent.net/xtradeb/apps/ubuntu>.

La page consultée affichait une version Chromium plus récente (`151.0.7922.137-1xtradeb1.2404.1`) que le runtime QA local figé (`151.0.7922.71-1xtradeb1.2404.1`). Cette divergence confirme qu’une mise à niveau ne doit jamais être implicite : le paquet exact validé doit être archivé avec son hash et sa chaîne de signature, ou le candidat plus récent doit suivre une nouvelle validation complète.

> Cette note est une observation de provenance publique. Elle ne constitue pas, à elle seule, une vérification cryptographique de la signature du paquet `.deb` exact.

## Référence

1. [Launchpad — PPA XtraDeb Apps pour Ubuntu Noble](https://launchpad.net/~xtradeb/+archive/ubuntu/apps?field.series_filter=noble)

## Disponibilité observée lors de la qualification

La liste publique des publications Chromium pour Ubuntu Noble ne présentait plus que la publication courante `151.0.7922.137-1xtradeb1.2404.1`. La version QA historique `151.0.7922.71-1xtradeb1.2404.1` n’y figurait plus. Côté index APT local, le candidat résolu au moment de la qualification était `151.0.7922.108-1xtradeb1.2404.1`.

Le paquet candidat `151.0.7922.108-1xtradeb1.2404.1` a donc été traité comme un **nouveau candidat de runtime**, sans modifier le statut du pilote BACK-01 ni autoriser une publication publique. Sa capture locale a conservé le `.deb`, `InRelease`, `Packages.gz`, le keyring public et les contrôles de checksum correspondants.

> Une divergence ultérieure entre la publication Launchpad et le candidat observé dans l’index APT n’autorise aucune mise à niveau implicite. La version exacte réellement testée doit rester verrouillée dans les preuves.

2. [Launchpad — publications Chromium de la PPA XtraDeb](https://launchpad.net/~xtradeb/+archive/ubuntu/apps/+packages?field.name_filter=chromium&field.status_filter=published&field.series_filter=noble)
