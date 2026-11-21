# TODO

## Security

### Stored XSS via bookmark widget

**Status: fixed (2026-11-21), option "Both".** Bookmark URLs must be empty,
`http(s)://…` or a single-slash `/path`; the dashboard store rejects anything
else on create/update/zip import (400), the server renderer and the editor
preview render unsafe stored URLs inert, and the editor form refuses them.
See `docs/agents/architecture.md` (security model) and `docs/agents/widgets.md`
(config validation, sink audit). Kept below as the original report.

The backend is a pass-through for dashboard JSON — no validation or sanitization on widget configs.
`BookmarkWidget.vue` binds `config.url` directly to `:href`, which means a crafted API request
can persist a `javascript:` URI that executes when a user clicks the link.

**Fix options:**
- **Frontend:** validate URLs in BookmarkWidget — reject anything not starting with `http://`, `https://`, or `/`
- **Backend:** validate widget configs by type before persisting
- **Both** (defense in depth)

Also audit any future widgets that use config values in `href`, `src`, or `class` bindings.
