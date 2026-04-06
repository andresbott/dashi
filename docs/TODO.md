# TODO

## Security

### Stored XSS via bookmark widget

The backend is a pass-through for dashboard JSON — no validation or sanitization on widget configs.
`BookmarkWidget.vue` binds `config.url` directly to `:href`, which means a crafted API request
can persist a `javascript:` URI that executes when a user clicks the link.

**Fix options:**
- **Frontend:** validate URLs in BookmarkWidget — reject anything not starting with `http://`, `https://`, or `/`
- **Backend:** validate widget configs by type before persisting
- **Both** (defense in depth)

Also audit any future widgets that use config values in `href`, `src`, or `class` bindings.
