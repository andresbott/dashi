<!-- todo:guide — managed by todo; this block is rewritten on save. Docs: https://github.com/andresbott/todo
This file is a todo list managed by "todo", a terminal TODO app:
https://github.com/andresbott/todo

todo watches this file and reloads it automatically when it changes on disk, so
you — human or agent — can edit it directly in any editor. Keep to this format
so todo can parse what you write:

  # Heading           Headings ("#" to "######") are categories; they nest by
                      heading level.
  - [ ] Open task     A "- [ ]" line is an open task; "- [x]" marks it done.
  - [x] Done task     Tasks must live under a category heading.
    - [ ] Subtask     Indent by two spaces to nest a subtask under a task.
    Description text  An indented, non-checkbox line is the task's description.

Notes for editors:
- Text above the first heading (this block included) is preserved on save.
- todo rewrites the file into the canonical form above on every change, so any
  other free-form markdown placed between items is not kept.
-->

* migrate code as HA addon — DONE (ha-addon/, editor via ingress + viewer host port)
* use enhanced dither
* rework widgets as isolated component
* remove preview fetaure
* recreate themes
* make custom preview dialog for eink dispalayus
* custom themes
* json not passed directly to backend
* make sure mobile view works
* classify main sections in the admin SPA (e.g. notes/images = user content, backgrounds/themes = theming)
* uploaded themes only become usable by widgets/renderers after a backend restart (icons/fonts registered once at startup in app/router/main.go)

- [] LITEhtmk cannot real css vars, => explore if we watnto use a css library to transform modern css into plain css e.g. something simialr to scss

# Release

- [x] Home Assistant integration
  delivered as a prebuilt HA add-on image; see ha-addon/ and docs/agents/releasing.md

# Vuejs

- [ ] autosave
- [x] Grid placement without placeholders

# Backedn

- [ ] review new userauth capabilites for auth, e.g. PAT

# theming

- [ ] background editor
  a background is a self contained unit 
- [ ] themes define fonts etc
  themes define fonts, colors, icons etc
- [ ] theme still bakces in background, now that we have moved that as sepearete entity
  cleanup
