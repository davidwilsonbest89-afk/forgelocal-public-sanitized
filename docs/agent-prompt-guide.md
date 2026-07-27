# Agent Prompt Guide

Use this guide when connecting an AI agent to BrowseForge MCP. It is written to be copied into an agent system or developer prompt.

## Recommended Tool Policy

BrowseForge controls persistent browser profiles and isolated agent pages. Prefer stable, observable browser automation over guessing.

1. Start by identifying the right profile. Use `list_profiles` when `profile_id` is unknown.
2. Diagnose before retrying. Use `doctor_profile` when a profile, browser, or session state is unclear.
3. Observe before acting. Use `get_page_state` after opening a browser, navigating, switching tabs, or recovering from a failed action.
4. Wait for conditions, not time. Use `wait_for` after `navigate`, `click`, form submit, SPA transitions, or downloads that expose page status. Do not use fixed sleeps when a selector state can be observed.
5. Use dedicated input tools before JavaScript. Prefer `form_fill`, `select_option`, `check`, `press_key`, `click`, and `type_text`. Use `evaluate` only when no dedicated tool fits.
6. Reuse agent sessions. For `web_search` and `web_explore`, keep the returned `session_id` and pass it to follow-up tools such as `web_extract`, `wait_for`, `get_page_state`, and `screenshot`.
7. Clean up isolated agent sessions. Use `destroy_session` after a search/explore workflow when the page is no longer needed. Do not close the whole profile browser unless the task is complete or explicitly requested.
8. Use downloads tools for exported files. After an export action, call `list_downloads`, then `read_download` for small files or report the file path for large files.
9. Save visual evidence when useful. For HTTP MCP, use `screenshot` with URL delivery and fetch the temporary unauthenticated `screenshot_url` before `expires_at`; for stdio/default clients, keep `delivery=image` unless the client cannot handle image blocks.
10. Keep profile state safe. Do not directly modify profile directories or browser-data. Use profile, cookie, download, and session tools.

## Common Sequences

### General Page Task

1. `open_browser`
2. `get_page_state`
3. `navigate`
4. `wait_for` a page-specific visible selector
5. `get_page_state`
6. Use `click`, `form_fill`, `select_option`, `check`, or `press_key`
7. `wait_for` the next meaningful selector or state
8. `web_extract`, `get_content`, or `screenshot`

### Search And Inspect

1. `web_search` with a Chromium profile
2. Save `session_id`
3. Choose a result URL from top-level `results` or `raw_fallback`
4. `web_explore` with the same `session_id`
5. `web_extract` or `get_page_state` with the same `session_id`
6. `destroy_session`

### Login Or Form Flow

1. `doctor_profile`
2. `open_browser`
3. `navigate`
4. `wait_for` the login form selector
5. `form_fill`
6. `press_key` or `click` submit
7. `wait_for` a dashboard/authenticated selector
8. `get_page_state`

### Export Or Download Flow

1. Trigger the export with `click` or `press_key`
2. Use `wait_for` if the page shows a completion indicator
3. Call `list_downloads`
4. Call `read_download` only for small files that need inspection
5. Keep the file path for larger artifacts

## Avoid

- Do not use `evaluate` for normal clicks, typing, dropdowns, checkboxes, page state, or structured extraction.
- Do not create a new profile when an existing profile contains required cookies or identity state.
- Do not create repeated agent sessions for one research task; reuse `session_id`.
- Do not close the profile browser to clean up an agent page; use `destroy_session`.
- Do not assume navigation is complete just because `navigate` returned; wait for a selector that proves the app is ready.
