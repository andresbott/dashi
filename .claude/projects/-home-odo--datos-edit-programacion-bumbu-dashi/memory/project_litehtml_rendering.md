---
name: Image rendering uses litehtml
description: Dashboard image rendering uses litehtml which does not support flexbox CSS — only use float/table layouts in static templates
type: project
---

Image mode rendering uses litehtml (not a full browser engine). litehtml does NOT support flexbox CSS (`display: flex`, `flex-wrap`, etc.).

**Why:** litehtml is a lightweight HTML renderer with limited CSS support. Flexbox properties are silently ignored, causing layout breakage in image mode.

**How to apply:** When writing or editing CSS in `master.html` or widget HTML templates, only use float-based or table-based layouts. Never use `display: flex`, `flex-wrap`, `align-items`, `justify-content` on layout containers that affect widget positioning.
