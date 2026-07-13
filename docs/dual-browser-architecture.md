# Dual-Browser Anti-Detection Architecture

BrowseForge now treats dual-browser support as the core product contract. The early Camoufox-first research remains useful background, but the current runtime is a unified anti-detection workspace for Firefox/Camoufox and Chromium/CloakBrowser profiles.

## Current Runtime Contract

Every v2 profile has one runtime identity:

- `runtime_id` identifies the concrete provider. Built-in values are `camoufox` and `cloakbrowser`.
- Browser family is runtime metadata (`firefox` or `chromium`), not a profile field.
- v1 profiles that only contain `engine` must be migrated with `BrowseForge migrate profiles --from v1 --to v2 --apply`.

The REST API, MCP server, dashboard, workflow runner, session store, backups, and Playwright connection model use the same runtime-provider contract. Clients must send `runtime_id`; `engine` was removed from profile create/update APIs. Provider-specific behavior is isolated in runtime providers, the browser manager, and the browser download layer.

## Firefox/Camoufox Path

Camoufox is the Firefox-family runtime. BrowseForge launches one isolated browser process per profile, with a profile-specific data directory and a native `CAMOU_CONFIG` payload.

Camoufox receives the full fingerprint object from the BrowseForge fingerprint pool. That object may include navigator, screen, window, font, canvas seed, timezone, WebRTC, and complete WebGL fields. WebGL is intentionally all-or-nothing: if BrowseForge does not have a complete WebGL profile, it removes partial renderer/vendor fields and lets Camoufox generate a consistent native result.

This path is best for profiles that need explicit, inspectable fingerprint fields and Camoufox's Firefox-level native masking behavior.

## Chromium/CloakBrowser Path

CloakBrowser is the Chromium-family runtime. BrowseForge launches one isolated browser process per profile with a profile-specific user data directory and native CloakBrowser flags.

CloakBrowser primarily uses `fingerprint_seed` as the identity source. The seed drives the native fingerprint surface inside the Chromium runtime. BrowseForge may also pass selected explicit fields, such as timezone, locale, WebRTC, platform, and fonts, when they are available on the profile.

New Chromium profiles automatically receive a seed when one is not provided. This is intentionally different from the Firefox path: the seed is the primary identity contract, while explicit fields are overrides.

## Playwright Control

Both engines expose a Playwright-compatible Bind endpoint through the project Playwright 1.61.1 integration. External Playwright 1.61.x clients connecting to Camoufox should create pages with `noViewport` because Camoufox v135's Juggler protocol does not accept Playwright's newer viewport `isMobile` field.

The current release gate includes a Camoufox Bind runtime spike. BrowseForge also keeps an opt-in CloakBrowser Bind spike (`CLOAKBROWSER_SPIKE=1`) for machines that provide a verified CloakBrowser binary. Both tests launch a persistent profile, bind a Playwright endpoint, connect a second Playwright client, and open a page through that endpoint.

## Historical Material

Some older files describe a Camoufox-only or per-container setter design. Those documents are planning archive, not the current product contract. The current implementation does not depend on one shared browser with per-container spoofing. It uses one isolated runtime process per profile, which is simpler to reason about and matches both browser families.

Treat these as historical unless they are explicitly refreshed:

- `docs/plan.md`
- `docs/execution.md`
- `docs/wbs.md`
- `docs/phase2-fork-plan.md`
- `docs/spike-results.md`
- `extension/lib/fingerprint-injector.js`

## Current Gaps

- Run release preflight with `REQUIRE_CLOAKBROWSER=1` on release machines that provide a verified CloakBrowser binary.
- Keep public docs English-first with Traditional Chinese counterparts for release-critical workflows.
- Continue replacing old internal names such as `camoufoxmulti` and `cmfx` when they are not part of backward-compatible persisted data.
- Register future providers as runtime-provider changes; do not encode a new fork or vendor as another public `engine` value unless the browser family itself changes.
