# GOSEC-R7 — réconciliation du manifest

Le package `forgelocal-gosec-r7-final-v1` et son manifest historique sont conservés tels quels. La vérification publique du 26 août 2026 a confirmé le contenu réel du bundle : il contient la ref d’évidence `24537ac09191ca4e1472ae48e167a0fa30691eaf` et requiert le parent source `3656dbad4bfef0381e1f9d837271d293ecffe292`.

Le champ `bundle_scope` du manifest v1 contient une erreur rédactionnelle : il indiquait `delta_from_24537ac...` alors que le scope réel et le nom du bundle sont `delta_from_3656dba..._to_24537ac...`. Cette note ne réécrit ni le raw, ni le ZIP, ni le TAR, ni le bundle v1.

Le fichier compagnon `R7_FINAL_MANIFEST_V2.txt` et le package v2 portent la formulation corrigée. Les hashes du package v1 restent ceux publiés dans ses sidecars et la vérification v1 reste PASS.
