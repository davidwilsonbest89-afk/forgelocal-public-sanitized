# Documentation Language Audit

BrowseForge uses English as the default documentation language and supports Traditional Chinese through `.zh-TW.md` companion files.

## Current Policy

- Unsuffixed public documentation files are canonical English documents.
- Traditional Chinese translations use `.zh-TW.md`.
- Public entry points should include language-switch links when a localized companion exists.
- Historical research and planning notes can remain in their original language until they become public onboarding material.

## Public Entry Points

These files are expected to stay English-first:

| File | Status |
|------|--------|
| `README.md` | English canonical with `README.zh-TW.md` companion. |
| `API.md` | English canonical with `API.zh-TW.md` companion. |
| `docker/README.md` | English canonical with `docker/README.zh-TW.md` companion. |
| `docs/README.md` | English documentation index. |
| `docs/cli.md` | English CLI reference. |
| `docs/local-quickstart.md` | English local setup guide. |
| `docs/cloud-deployment.md` | English cloud deployment guide. |
| `docs/agent-integration.md` | English agent integration guide. |
| `docs/developer-integration.md` | English developer integration guide. |
| `docs/linux-server.md` | English canonical with zh-TW companion. |
| `docs/platform-support.md` | English canonical with zh-TW companion. |
| `docs/dual-browser-architecture.md` | English canonical with zh-TW companion. |
| `docs/playwright-patches.md` | English canonical with zh-TW companion. |
| `docs/release.md` | English release process. |
| `docs/i18n.md` | English i18n policy. |

## Known Non-English Historical Notes

These files are not treated as public onboarding entry points yet. They should be translated or split into English canonical plus `.zh-TW.md` companion files before being promoted into the public documentation path:

| File | Current role |
|------|--------------|
| `docs/humanize-research.md` | Historical behavior-simulation research note. |
| `docs/spike-results.md` | Historical spike result notes. |
| `docs/wbs.md` | Historical work breakdown structure. |
| `docs/execution.md` | Historical execution plan. |
| `docs/phase2-fork-plan.md` | Historical fork planning note. |
| `docs/plan.md` | Historical project plan. |
| `docs/webgl-fingerprint-strategy.md` | Browser-runtime research and strategy notes. |

## Guardrail

Run:

```bash
npm run check-doc-language
```

The guardrail scans public English entry points for CJK text. It intentionally does not scan `.zh-TW.md` files or historical planning notes.
