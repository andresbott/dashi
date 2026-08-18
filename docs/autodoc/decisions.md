# Dashi — Architecture Decisions

## File-Based Storage, No Database
**Date:** 2026 (project inception)
**Context:** Needed persistent storage for a self-hosted dashboard app
**Decision:** JSON files on disk with in-memory index, no database
**Why:** Maximizes portability and simplicity. Users can back up by copying a
folder, edit JSON by hand, and version-control their dashboards. No database
setup or migration overhead for a single-user app.

## Dual Rendering: Interactive + Image
**Date:** 2026 (project inception)
**Context:** Dashboards needed to serve both web browsers and e-ink displays
**Decision:** Same dashboard data drives two rendering paths — Vue SPA
(interactive) and server-side HTML→PNG (image via litehtml-go)
**Why:** E-ink devices need a static PNG at a fixed resolution. Sharing the same
data model and widget definitions avoids maintaining two separate systems.

## Widget Registry Pattern (Backend + Frontend)
**Date:** 2026 (project inception)
**Context:** Multiple widget types need to render in both modes
**Decision:** Backend registry maps type string → `StaticRenderer` function.
Frontend registry maps type string → Vue component + config component.
**Why:** Decouples widget implementation from dashboard rendering. Adding a new
widget doesn't require changes to core dashboard code — just register it.

## RawJSON Widget Config
**Date:** 2026 (project inception)
**Context:** Each widget type has different configuration needs
**Decision:** Widget config stored as `json.RawMessage` (Go) / opaque JSON (TS),
parsed by each widget individually
**Why:** Avoids a central config schema that must be updated for every widget.
Each widget owns its config format. Tradeoff: no shared validation.

## Embedded Frontend in Go Binary
**Date:** 2026 (project inception)
**Context:** Deployment simplicity for self-hosted app
**Decision:** Vue build output copied to `app/spa/files/ui/` and embedded via
`go:embed`
**Why:** Single binary deployment — no separate web server, no static file
hosting, no Docker multi-stage complexity. Users download one file.

## Snake-Case Folder Names for Dashboards
**Date:** 2026
**Context:** Dashboard storage needs human-readable folder names
**Decision:** Folder name derived from dashboard name via Unicode normalization +
snake_case conversion. Collision detection appends `_2`, `_3`, etc.
**Why:** Readable on disk for manual inspection/backup. Folder name is
independent of the 6-char random ID, so renaming a dashboard doesn't move files.

## Two HTTP Servers (App + Observability)
**Date:** 2026
**Context:** Metrics and health endpoints shouldn't be exposed on the public port
**Decision:** Separate server on different port for observability
**Why:** Standard pattern for self-hosted apps. Allows firewall rules to expose
only the app port while keeping metrics internal.

## Self-Contained Widget Modules
**Date:** 2026-05-01
**Context:** Adding a widget previously touched 12 files across 8
directories; app-layer warmup functions hard-coded widget type strings
and config shapes.
**Decision:** Each widget is a `widgets.Module` implementation living in
its own backend folder (`internal/widgets/<type>/`) + mirrored frontend
folder (`webui/src/widgets/<type>/`). Central wiring iterates a
`[]widgets.Module` slice; a generic `widgets.CollectConfigs` helper
replaces per-widget warmup scanners.
**Why:** Collapses the integration surface to two folders + two one-line
registrations. Widget-specific knowledge (handlers, warmup, config
parsing) lives with the widget. The app layer has no widget-specific
imports beyond the modules slice itself.
