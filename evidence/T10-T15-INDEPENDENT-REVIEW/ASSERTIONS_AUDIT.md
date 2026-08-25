# Assertions audited from published T10/T15 kit

| Assertion | Evidence checked |
|---|---|
| T10 valid creation | `T10_VALID_PROXY_CREATED: PASS server_validated` |
| T10 invalid port refused | `T10_INVALID_PORT_REFUSED: PASS no_write_on_rejection` |
| T10 phantom profile refused | `T10_ASSIGN_REQUIRES_CORE_PROFILE: PASS explicit_refusal_no_ghost_binding correlated` |
| T10 off-loopback writes refused | `T10_OFFLOOPBACK_REFUSED: PASS no_write_path_origin_offloopback` |
| Redacted listing / no credential value | `T10_LISTING_REDACTED: PASS no_credential_value_in_ui` |
| T15 external navigation refused | T15 W2 test name in published Playwright log/source |
| Browser credential absence / digest projection | T15 W3/W5 test names in published Playwright log/source |
| T15 session close | T15 W4 test name in published Playwright log/source |
| Dashboard automation panel | T15 W5 test name in published Playwright log/source |
| Sequential execution | `Running 7 tests using 1 worker` and `7 passed (15.6s)` |
