<!-- todo:guide — managed by todo; this block is rewritten on save. Docs: https://github.com/candy-tools/todo
This file is a todo list managed by "todo", a terminal TODO app:
https://github.com/candy-tools/todo

todo watches this file and reloads it automatically when it changes on disk, so
you — human or agent — can edit it directly in any editor. Keep to this format
so todo can parse what you write:

  # Heading           Headings ("#" to "######") are categories; they nest by
                      heading level.
  - [ ] Open task     A "- [ ]" line is an open task; "- [x]" marks it done.
  - [/] In progress   "- [/]" flags a task in progress, "- [>]" defers it.
  - [x] Done task     Tasks must live under a category heading.
    - [ ] Subtask     Indent by two spaces to nest a subtask under a task.
    Description text  An indented, non-checkbox line is the task's description.

Notes for editors:
- Text above the first heading (this block included) is preserved on save.
- todo rewrites the file into the canonical form above on every change, so any
  other free-form markdown placed between items is not kept.

Prefer the CLI over editing by hand: it writes this file in the exact format
above. `todo add`, `todo done|progress|defer|reopen`, `todo edit`, `todo rm`
and `todo list --json` cover the common actions — run `todo --help`.
-->

# Backlog

- [x] migrate code as HA addon
  ha-addon/, editor via ingress + viewer host port
- [x] use enhanced dither
  per-dashboard dither mode: plain / auto / preset
- [x] rework widgets as isolated component
  internal/widgets/{name} + webui/src/widgets/{name}, see docs/agents/widgets.md
- [x] remove preview fetaure
  preview dashboards removed, see docs/agents/features.md
- [x] recreate themes
  internal/themes + internal/widgetthemes
- [x] make custom preview dialog for eink dispalayus
  "E-ink preview" in the dashboard editor, image dashboards
- [x] custom themes
  themes editor, widget theme upload
- [ ] json not passed directly to backend
- [x] make sure mobile view works
  viewer stacks rows below 640px; admin drawer below 768px; dialogs fit; editor grid scrolls sideways
- [x] classify main sections in the admin SPA (e.g. notes/images = user content, backgrounds/themes = theming)
  Theming: themes, background images, widget themes; Content: notes, images
- [x] uploaded themes only become usable by widgets/renderers after a backend restart (icons/fonts registered once at startup in app/router/main.go)
  upload/delete re-sync PNG fonts via widgetthemes Store.OnChange
- [x] LITEhtmk cannot real css vars, => explore if we watnto use a css library to transform modern css into plain css e.g. something simialr to scss
  explored in docs/project/litehtml-css-vars.md: the vendored litehtml already supports var(); no CSS transform library needed

# Release

- [x] Home Assistant integration
  delivered as a prebuilt HA add-on image; see ha-addon/ and docs/agents/releasing.md

# Vuejs

- [x] autosave
  dashboard editor, dashboard settings and theme editor autosave (useAutosave / useDashboardDraft)
- [x] Grid placement without placeholders

# Backedn

- [ ] review new userauth capabilites for auth, e.g. PAT

# theming

- [x] background editor
  a background is a self contained unit 
- [x] themes define fonts etc
  themes define fonts, colors, icons etc
- [x] theme still bakces in background, now that we have moved that as sepearete entity
  cleanup
- [x] remove background images from content
