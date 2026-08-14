# Observation externe — provenance PPA XtraDeb

**Consulté le :** 14 août 2026

La page publique Launchpad de la PPA XtraDeb Apps indique la PPA `xtradeb/apps` et expose une clé de signature `4096R/5301FA4FD93244FBC6F6149982BB6851C64F6880`. La même empreinte est présente dans la configuration APT locale, dans `/etc/apt/trusted.gpg.d/xtradeb-apps.gpg`.

La PPA publiée pour Ubuntu Noble est présentée sur : <https://launchpad.net/~xtradeb/+archive/ubuntu/apps?field.series_filter=noble>. Le dépôt public associé est : <https://ppa.launchpadcontent.net/xtradeb/apps/ubuntu>.

La page consultée affichait une version Chromium plus récente (`151.0.7922.137-1xtradeb1.2404.1`) que le runtime QA local figé (`151.0.7922.71-1xtradeb1.2404.1`). Cette divergence confirme qu’une mise à niveau ne doit jamais être implicite : le paquet exact validé doit être archivé avec son hash et sa chaîne de signature, ou le candidat plus récent doit suivre une nouvelle validation complète.

> Cette note est une observation de provenance publique. Elle ne constitue pas, à elle seule, une vérification cryptographique de la signature du paquet `.deb` exact.

## Référence

1. [Launchpad — PPA XtraDeb Apps pour Ubuntu Noble](https://launchpad.net/~xtradeb/+archive/ubuntu/apps?field.series_filter=noble)
