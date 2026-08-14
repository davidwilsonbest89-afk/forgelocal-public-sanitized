# BrowseForge Documentation

This directory contains public product documentation and project planning notes.

## ForgeLocal product reference

The internal French product reference is versioned separately from the public documentation. The current canonical documents are the [ForgeLocal specification v1.0](CAHIER_DES_CHARGES_FORGELOCAL.md), its [v2.1 addendum on audited Camoflox modules and the read-only dashboard](ADDENDUM_V2_1_CAMOFLOX_READONLY_DASHBOARD.md), and the [v2.3 addendum on component provenance](ADDENDUM_V2_3_COMPONENT_PROVENANCE.md). The [canonical JSON component-rights register](component-rights-register.json) is enforced by CI; its Markdown view is available in [COMPONENT_RIGHTS_REGISTER.md](COMPONENT_RIGHTS_REGISTER.md). None of these documents authorizes publication of the frozen BACK-01 RC.

## Public Documentation

These documents should remain understandable for international users:

- [Release Process](release.md)
- [CLI Reference](cli.md)
- [Local Quickstart](local-quickstart.md)
- [Cloud Deployment](cloud-deployment.md)
- [Agent Integration](agent-integration.md)
- [Developer Integration](developer-integration.md)
- [Platform Support](platform-support.md)
- [Linux Server Deployment](linux-server.md)
- [Dual-Browser Anti-Detection Architecture](dual-browser-architecture.md)
- [Playwright Patch Status](playwright-patches.md)
- [Internationalization](i18n.md)
- [Documentation Language Audit](documentation-language-audit.md)
- [Docker Runtime](../docker/README.md)

Traditional Chinese translations are available for:

- [Docker Runtime zh-TW](../docker/README.zh-TW.md)
- [Platform Support zh-TW](platform-support.zh-TW.md)
- [Linux Server Deployment zh-TW](linux-server.zh-TW.md)
- [Dual-Browser Anti-Detection Architecture zh-TW](dual-browser-architecture.zh-TW.md)
- [Playwright Patch Status zh-TW](playwright-patches.zh-TW.md)

## Technical Reference

These documents describe implementation strategy and browser-runtime behavior. They may include research details and historical notes:

- [Architecture Guide](architecture.md)
- [Agent Prompt Guide](agent-prompt-guide.md)
- [Anti-Detection Mechanisms](anti-detection.md)
- [WebGL Fingerprint Strategy](webgl-fingerprint-strategy.md)
- [Humanize Research](humanize-research.md)
- [Spike Results](spike-results.md)

## Planning Archive

These files are retained as historical planning material and may mix languages:

- [Execution Plan](execution.md)
- [Phase 2 Fork Plan](phase2-fork-plan.md)
- [Project Plan](plan.md)
- [Work Breakdown Structure](wbs.md)

For user-facing entry points, prefer the root [README](../README.md), [API reference](../API.md), and [support guide](../SUPPORT.md).
