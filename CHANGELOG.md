# Changelog

All notable changes should be documented here. This project follows semantic version tags in the form `vX.Y.Z`.

## v2.1.12 - 2026-08-04

### Fixed

- Validated Playwright Bind endpoints before registering browser sessions so external clients no longer receive unusable WebSocket endpoints after a launch, crash, or handshake failure.
- Added structured endpoint health failure codes and diagnostic logs with session, profile, runtime, endpoint, profile directory, user-data directory, executable path, timeout, and retry context.

## v2.1.11 - 2026-07-27

### Fixed

- Made MCP `screenshot` URL delivery work from stdio when `public_base_url` is configured by sharing temporary screenshot artifacts through the main HTTP service.

## v2.1.10 - 2026-07-27

### Fixed

- Added temporary unauthenticated MCP screenshot URLs so remote HTTP agents can fetch image bytes directly instead of parsing base64 image blocks.
- Documented `BROWSEFORGE_PUBLIC_BASE_URL` Docker configuration and the screenshot URL delivery contract for remote agents.

## v2.1.9 - 2026-07-20

### Fixed

- Added a validated MCP default web profile so `web_search` and `web_explore` can start without an explicit profile ID while preserving agent-capable Chromium runtime checks.
- Bumped Docker, Linux server, release, and platform documentation references for the `v2.1.9` release tag.

## v2.1.8 - 2026-07-19

### Fixed

- Pointed BrowseForge to BrowseForge Chromium runtime `v0.1.6-alpha.0` for the BrowserLeaks Fonts crash-fix runtime release.
- Bumped Docker, Linux server, release, and platform documentation references for the `v2.1.8` release tag.

## v2.1.7 - 2026-07-16

### Fixed

- Pointed BrowseForge to BrowseForge Chromium runtime `v0.1.5-alpha.0` after the detector evidence hardening release.
- Bumped Docker, Linux server, release, and platform documentation references for the `v2.1.7` release tag.

## v2.1.6 - 2026-07-16

### Fixed

- Bumped Docker, Linux server, release, and platform documentation references for the `v2.1.6` release tag.
- Aligned BrowseForge Chromium Docker/native headed sessions with the native persona screen contract, including DPR switches, coherent KasmVNC geometry defaults, and smoke-test assertions for screen/window, locale, timezone, WebDriver, and WebGL surfaces.
- Captured sanitized Pixelscan screen/window evidence so detector reports can distinguish true fingerprint mismatches from missing viewport evidence.

## v2.1.5 - 2026-07-16

### Fixed

- Preserved `DISPLAY`, `HOME`, and `LIBGL_ALWAYS_SOFTWARE` when launching BrowseForge Chromium so GHCR Docker sessions can use the KasmVNC X display instead of failing with a missing-X-server error.
- Bumped Docker and Linux server references for the `v2.1.5` release tag.

## v2.1.4 - 2026-07-16

### Changed

- Published GHCR release images as native `linux/amd64` and `linux/arm64` manifests backed by BrowseForge Chromium runtime `v0.1.4-alpha.0`.
- Updated Docker and Linux server references for the `v2.1.4` release tag.
- Kept the runtime release gate strict: Docker preinstall verifies the matching linux runtime asset before image build.

### Fixed

- Added native BrowseForge Chromium `linux/arm64` download mapping and runtime asset checks so ARM containers no longer fall back to x64/emulation assets.
- Preserved coherent native persona metadata for BrowseForge Chromium launch on Linux, macOS, and Windows runtime packages.

## v2.1.2 - 2026-07-13

### Changed

- Updated BrowseForge to the maintained `github.com/mxschmitt/playwright-go` driver package.
- Preinstalled the Playwright control driver in GHCR Docker images so container startup does not depend on runtime downloads from npm or nodejs.org.
- Bumped Docker, Linux server, release, and platform documentation references to the `v2.1.2` release tag.
- Decoupled BrowseForge release support from browser runtime release support so unsupported runtime/platform pairs are skipped and disabled at runtime instead of blocking BF startup.
- Kept Camoufox runtime selection on `v135.0.1-beta.24` because it publishes complete browser binaries for the supported native platforms, and updated BrowseForge Chromium runtime selection to `v0.1.2-alpha.0`.
- Added a shared BrowseForge Chromium release-asset checker plus release-preflight, workflow, and Docker build override support for staged runtime asset roots, and fail release gates before Docker build when the selected runtime version lacks `linux-x64` or `linux-arm64` release assets.

### Fixed

- Kept BrowseForge Chromium launch identity coherent on native Linux ARM64 by deriving UA, UA-CH architecture, `navigator.platform`, and `Accept-Language` from the resolved runtime platform/locale instead of blindly reusing incompatible pool values.
- Populated BrowseForge Chromium native persona JSON from the same canonical launch persona used for command-line fingerprint switches, including browser, platform, locale, hardware, screen, GPU, WebRTC, and storage fields.
- Rejected invalid BrowseForge Chromium proxy region labels before launch and clamped screen available dimensions so emitted switches and native JSON stay coherent.

## v2.1.1 - 2026-07-13

### Changed

- Prepared GHCR Docker publishing for native `linux/amd64` and `linux/arm64` manifests.
- Updated Docker runtime builds to select BrowseForge release ZIPs and KasmVNC packages from the BuildKit target architecture.
- Bumped Docker, Linux server, release, and platform documentation references to the `v2.1.1` release tag.

### Fixed

- Added BrowseForge Chromium downloader support for `linux/arm64` so ARM64 containers install the native `linux-arm64` runtime asset instead of failing during browser preinstall.

## v2.1.0 - 2026-07-12

### Changed

- Updated release and Docker references for the `v2.1.0` release tag.
- Kept Camoufox plus BrowseForge Chromium as the GHCR/Docker preinstall default while retaining CloakBrowser as a manual/custom runtime.

### Added

- Verified BrowseForge can download and install the released BrowseForge Chromium `v0.1.0-alpha.0` runtime asset from GitHub Releases.

## v2.0.0 - 2026-07-07

### Changed

- Replaced the public profile browser-family contract with explicit runtime providers. Profile create/update APIs, MCP profile tools, workflows, dashboard forms, and profile storage now use `runtime_id` values such as `camoufox` and `cloakbrowser`.
- Moved runtime binary paths and CloakBrowser launch/fingerprint settings under `runtimes.<id>` configuration, with `default_runtime_id` for UI defaults and `/api/runtimes` metadata.
- Updated Docker, Linux server, release, and platform documentation for the `v2.0.0` release tag.

### Added

- Added `BrowseForge migrate profiles --from v1 --to v2 [--apply]` to migrate legacy profile JSON safely, including original-file `.v1.bak` backups for every rewritten profile.
- Added runtime capability metadata so API, MCP web sessions, dashboard, and browser manager code gate behavior on provider capabilities instead of hard-coded engine strings.

### Fixed

- Rejected deprecated `engine` profile create/update inputs at REST and MCP boundaries, rejected non-string or disabled `runtime_id` updates, and prevented dashboard selection of disabled runtimes before metadata loads.
- Preserved anti-detection ownership of managed CloakBrowser fingerprint flags, Camoufox WebGL normalization, and large `CAMOU_CONFIG` chunking after the runtime-provider refactor.

## v1.10.2 - 2026-07-05

### Fixed

- Strengthened CloakBrowser safe GPU fallback for Windows VM environments by adding GPU sandbox bypass and in-process GPU launch flags. This covers hosts where `--disable-gpu` alone still fails with `GPU process isn't usable`.

## v1.10.1 - 2026-07-05

### Added

- Added opt-in CloakBrowser launch compatibility settings for Windows VM environments, including safe GPU launch flags, isolated runtime cache, transient cache repair, and sanitized extra Chromium args.

### Fixed

- CloakBrowser launches can now retry once with a safe GPU fallback after GPU/cache startup failures such as `GPU process isn't usable` or `Unable to create cache`, without changing the first-launch anti-detection behavior by default.
- Safe GPU fallback restarts the Playwright driver cleanly, drops stale sessions after the restart, and avoids repeated manager-level retries once the fallback path has already been exhausted.

## v1.9.0 - 2026-07-01

### Added

- Added agent-ready CLI commands for runtime status, dashboard opening, MCP client config generation, browser engine status/install, full filesystem backups, metadata backups, and restore workflows.
- Added wait-aware CLI smoke checks so local and container deployments can block until REST or MCP endpoints are ready.
- Added local quickstart, cloud deployment, agent integration, and developer integration guides.

### Changed

- Docker release images now preinstall BrowseForge-managed browser engines during image build by default.
- Container startup now seeds or updates `/app/browsers` from the image when the mounted browser cache is missing or its engine version differs, while keeping tokens, profiles, data, logs, and backups on host-mounted paths.
- Docker and Linux server documentation now include persistent `/app/backups` mounts and explicit browser-cache upgrade behavior.

## v1.8.1 - 2026-06-30

### Changed

- Playwright Go now uses the upstream community `v0.6000.0` release instead of the temporary `nczz/playwright-go` integration fork.

## v1.7.7 - 2026-05-15

### Fixed

- Playwright driver compatibility patching is now format tolerant, so Firefox/Camoufox page errors without source locations no longer crash the driver in Docker builds.
- BrowseForge startup now fails fast if Playwright driver installation or patching fails instead of silently running an unpatched driver.

## v1.7.6 - 2026-05-15

### Fixed

- Playwright 1.60 driver crashes caused by Firefox/Camoufox page errors without source locations no longer poison later browser launches.
- Chromium/CloakBrowser launches now recover from stale Playwright driver protocol failures even when dead sessions remain in memory.

## v1.7.5 - 2026-05-15

### Fixed

- Firefox/Camoufox profile launches now clean stale profile locks before startup and automatically restart the Playwright driver once after recoverable protocol EOF errors.
- Spike tests no longer attempt to download Playwright-managed browsers during `go test ./...`.

## v1.7.4 - 2026-05-15

### Fixed

- Dashboard version status now remains visible after language initialization or locale changes instead of being overwritten by the translated connecting label.

## v1.7.3 - 2026-05-15

### Fixed

- Release binaries now report the tag version through the REST API, doctor output, and MCP initialize response.
- Release workflow verifies the packaged Linux binary version before publishing assets.

## v1.7.2 - 2026-05-15

### Added

- Guarded release scripts and release workflow asset checks.
- Docker release build hardening with pinned release artifact selection and KasmVNC checksum verification.
- Community governance, support, security, and contribution documentation.
- Initial application i18n policy and locale structure.
- English-first README and API reference with Traditional Chinese counterparts.
- English public docs for platform support, Linux server deployment, and Playwright patch status.
- i18n coverage checker for Dashboard and WebExtension locale key parity.
- Marketing-oriented product positioning, audience, trust, and deployment messaging in README.
- Dual-browser anti-detection architecture documentation in English and Traditional Chinese.
- Opt-in CloakBrowser runtime spike harness for the Playwright Bind endpoint path.

### Changed

- Docker documentation recommends pinning version tags for production deployments.
- Replaced remaining early Camoufox-only tool naming in local scripts and clarified current dual-browser fingerprint behavior.
- Release preflight runs the CloakBrowser Bind spike when a local binary is available and can enforce it with `REQUIRE_CLOAKBROWSER=1`.
- Release preflight keeps base Go tests separate from explicit browser runtime spike gates.

## v1.7.0 - 2026-05-15

### Added

- Playwright 1.60 integration through the project fork.
- MCP Streamable HTTP authentication with Bearer tokens.
- Camoufox runtime spike coverage for the Playwright Bind endpoint path.

### Changed

- Removed the previous Playwright 1.59.1 hotfix path and now uses the Playwright 1.60 `browser.Bind()` endpoint directly.
- Improved startup, token, browser-download, profile-store, backup/restore, and session request error handling.

### Upgrade Notes

- External Playwright clients should use Playwright 1.60.x for `browserType.connect()`.
- Existing `config.json`, `data/.api-token`, and `profiles/` remain compatible.
- MCP HTTP clients must send `Authorization: Bearer <token>`.

## 2026-08-24 — Validation indépendante T28–T42

La validation indépendante en sandbox réelle a été poursuivie jusqu’à T42. T31, T32, T33, T34, T35, T36, T37 et T38 sont `APPROVED_VERIFIABLE_LOCAL` avec preuves séquentielles publiées sur GitHub ; T39, T40, T41 et T42 sont `BLOCKED` pour raisons de dépendances produit, d’environnement natif ou de release. T28 et T29 restent également bloqués malgré la validation documentaire de leurs archives. Aucun runtime réel, Camoufox, proxy, cookie, SystemVault natif, migration utilisateur ou release n’a été exécuté. Les archives originales et les paquets rejetés pour taille restent conservés.

## 2026-08-24 — Correction de traçabilité T28–T42

Après revue indépendante, une branche dédiée ajoute uniquement des corrections de preuve : reconstructions postérieures explicitement nommées pour les baseline logs absents, sidecars compagnons portables relatifs, réconciliation T30 et métadonnées T42 corrigées. Les sidecars historiques sont conservés inchangés. Une qualification depuis un clone neuf confirme les tests race, vet, build, diff-check et fsck ; Gitleaks cumulatif et Gosec historiques restent présentés honnêtement avec leurs codes de sortie et leurs différentiels. Aucun code produit ni gate permanent n’est modifié.


## 2026-08-24 — Livraison gelée de la prévalidation humaine T00–T42

La branche `audit/t00-t42-prehuman-validation` livre le dossier complet sous `evidence/PREHUMAN_T00_T42/`. Le commit d’artefacts est `cf280858b345e2fd566d391590f23d8cfa6bbe6d`. Le ZIP `forgelocal-t28-t42-prehuman-validation.zip` a le SHA-256 `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80`, et le bundle `forgelocal-t28-t42-prehuman-validation.bundle` a le SHA-256 `14ef76cb68e7f64ff49fdc649cbcf96c5c69b0e9c410c5824a0592b7e33d1d14`. La décision est `T00_T42_PREHUMAN_VALIDATION_DELIVERED_PENDING_INDEPENDENT_REVIEW`.

Cette livraison fige uniquement la chaîne de preuve. Elle ne modifie pas le code produit, ne démarre aucun lot bloqué, ne lève aucune gate et ne constitue ni une release ni une autorisation produit.

Le journal brut de conservation ZIP/bundle a ensuite été ajouté dans le commit `cd6a95c66c029adf2140784491b06b8d9bf64fce`, head publié de la branche au moment du gel. Le bundle est un delta nécessitant la baseline ; sa réhydratation seeded avec `69411e65c880d168832a65fc8475cc97d562a9ad` a réussi avec checkout de `6ae02e4ceed239b9310fbf3fccb1b5170117251e` et `git fsck --full` à zéro.

## 2026-08-24 — Finalisation T00–T42-PREHUMAN-FINDINGS-FINALIZATION

La checklist finale a été complétée sans modification du code produit, des tests métier, de la configuration de lint ou des gates. Les 13 findings GolangCI-Lint ont été extraits individuellement avec chemins, lignes, messages bruts, lots/rattachements prudents, risques, propriétaires et conditions de levée. Douze lignes concernent des fichiers inchangés entre la baseline et HEAD et restent classées comme différentiel scanner/contexte non réconcilié; le finding SA9003 est rattaché à T38 et reste une exception de test ouverte. Les 36 findings Staticcheck, les 6 misconfigurations Trivy et l’inventaire de licences sont également détaillés dans l’addendum.

Playwright/T10 est documenté comme `NOT_APPLICABLE_UNDER_CURRENT_GATES`, avec preuve de commande, CWD, UTC, sortie brute et code de sortie; aucun Core, token, navigateur réel, Camoufox, proxy réel, cookie réel, SystemVault natif ou release n’a été exécuté. Le signal Gitleaks cumulatif `APi=REDACTED` conserve `SCAN_BLOCKED_UNKNOWN`; Gosec conserve 194 findings baseline/head sans nouveau différentiel.

Le wrapper append-only `forgelocal-t00-t42-prehuman-final-review-wrapper-v2.zip` rassemble le ZIP historique intact, l’addendum et les nouvelles preuves. Le ZIP historique et son hash `5c586895ea9b096ee529207ea57640227c5cb663c77c8d3aa77036258528fd80` sont inchangés. La sortie attendue est `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW`; elle ne constitue ni une approbation produit, ni une release, ni une levée de gate.

## 2026-08-24 — Correction de code des 13 findings GolangCI-Lint

Après analyse senior, les 13 findings ont été traités par deux commits de code : `6ee0840a7b264343be3840998df2a8903b511722` et `e0c9352710eb3710eaf0ea5d71614f2731a7051c`. Les retours de `srv.Shutdown`, `cmd.Start`, écritures réseau, copies bidirectionnelles et rollbacks transactionnels sont désormais traités explicitement; le test T38 ne contient plus de branche vide. Des tests de non-régression couvrent les écritures partielles/échouées, les copies échouées et le cycle transactionnel backup.

La qualification post-correctif depuis un clone neuf confirme zéro des 13 findings ciblés. Les findings Staticcheck/GolangCI-Lint non ciblés, Gosec historique, le signal cumulatif Gitleaks `APi=REDACTED`, les misconfigurations Trivy historiques et le blocage Playwright par configuration protégée restent documentés honnêtement. Cette correction ne lève aucune gate et ne constitue pas une release.

## 2026-08-24 — Publication code-fixed finale T00–T42

Le correctif senior des 13 findings GolangCI-Lint réellement défectueux est publié avec deux commits de code, des tests de non-régression et une qualification depuis un clone neuf. Le wrapper V3 code-fixed conserve le ZIP historique intact et rassemble le mapping, les logs, les SBOM, les scans et la preuve Playwright bloquée. Les findings historiques non ciblés, Gitleaks `APi=REDACTED`, les gates permanentes et l’absence de Core/token réel restent explicitement maintenus.

La livraison reste `T00_T42_PREHUMAN_VALIDATION_FINALIZED_PENDING_INDEPENDENT_REVIEW_WITH_CODE_FIXES`; elle ne constitue pas une release.

## 2026-08-24 — Self-validation v4 synthetic E2E

Ajout append-only de la chaîne de preuves `SELF_VALIDATION_V4`, incluant baseline brute, réhydratation LFS ciblée, contrôles d’artefacts, qualification Go/Dashboard, scans, SBOM CycloneDX/SPDX, inventaire de licences, classification individuelle des findings, E2E Playwright synthétique loopback et preuve de cleanup. Le statut est `T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW`; aucune gate n’est levée et aucune release n’est autorisée.

## 2026-08-24 — Livraison v4 append-only

La branche `audit/t00-t42-self-validation-synthetic-e2e` est préparée avec les commits `ad41afd71498fa5dda8eacc6a6ae0b47dbc865fd` et `c905f884ad9a84228985add6b7f77391e12b7b03`. Le wrapper V4 a le SHA-256 `429e683472b484076938d71428b5f52e1eb28794da1ad92526b670aa152f706b` et le bundle delta le SHA-256 `0e4159703d453d8bea37617fe8e89460026b0a3118e57257d471af1777fd743e`. Le statut reste `T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW`; aucune gate n’est levée et aucune release n’est autorisée.

## 2026-08-24 — Référence distante finale de la self-validation v4

La branche audit/t00-t42-self-validation-synthetic-e2e est publiée au HEAD 5e174dba6dddc35865f5bd943383d988ea12170c. Le wrapper V4 f6544091783c2a4d4694d4b5f02c5dd5f0c70d22dab5efbb90abfd81418019bc et le bundle delta 10059c3c610d5a1b1ade88f936c8bb52ed893741a596326d8ba532f6f415e2fe sont vérifiés par sidecars. Le statut reste T00_T42_SELF_VALIDATION_WITH_SYNTHETIC_E2E_COMPLETE_PENDING_INDEPENDENT_REVIEW, sans levée de gate ni release.

## 2026-08-24 — HEAD publié v4 synchronisé

Le HEAD publié final de la branche audit/t00-t42-self-validation-synthetic-e2e est b4a04e4b9b489c22f3a86986c6faa1cbb9bf77c5. Les preuves et hashes sont synchronisés ; le statut de self-validation reste en attente de revue indépendante et aucune release n’est autorisée.


## 2026-08-25 — T00–T42 V6 findings remediation

La qualification V6 append-only part du HEAD V5 `b34fa5c02ff20144abfb5d240db1c67ad1f038f9`. Elle corrige les deux findings GolangCI-Lint obligatoires, les violations Axe `color-contrast`/`meta-viewport` et les deux advisories `golang.org/x/mod` avec tests ciblés et commits séparés. Les 18 findings Semgrep sont qualifiés individuellement comme usages `crypto/rand`, et les findings Gitleaks historiques restent redacted et classés sans réécriture d’historique.

Les tests Go shuffle/race, vet, build, govulncheck, Grype sur SBOM CycloneDX/SPDX propres et l’E2E Playwright/Axe loopback passent. Les diagnostics Staticcheck/GolangCI-Lint historiques, les résultats OSV liés à la limite de détection de patch toolchain, les misconfigurations Docker Trivy, les licences inconnues, la limite Gitleaks de plage et les objets LFS historiques indisponibles restent ouverts et documentés. La sortie est `T00_T42_V6_FINDINGS_REMEDIATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`; aucune gate n’est levée et aucune release n’est autorisée.


## 2026-08-25 — Publication distante V6 vérifiée

La branche [audit/t00-t42-v6-findings-remediation](https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized/tree/audit/t00-t42-v6-findings-remediation) est publiée au HEAD distant `8e26bfb0c8bf6e92c09d645dd84ec854320c01f9`. Un clone neuf correspond exactement à ce SHA ; `git fsck --full` retourne 0 avant et après le fetch LFS ciblé des deux artefacts V6. Le wrapper V6 est `ce722915d70e0aa528927b753c6f18efa5706fc9fa8703ef6f449b6728a5fab6` et le bundle delta est `ad4484e795b80eb5b7655228012e695dc4b260d43057477a97ae145d164614c2`.

Le bundle exige b34fa5c et cible le commit de contenu fc08045 ; cette distinction avec le commit de packaging est volontaire et documentée. Le statut reste `T00_T42_V6_FINDINGS_REMEDIATION_COMPLETE_PENDING_INDEPENDENT_REVIEW`, sans levée de gate ni release.
