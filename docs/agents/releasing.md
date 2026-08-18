# Releasing — tags, goreleaser, the frontend-embed trap

Releases are goreleaser builds triggered by pushing a `v*.*.*` tag
(`.github/workflows/release.yml`).

## Mechanics

    make tag version="v1.2.3"

`make tag` refuses unless the working tree is clean and you are on `main`
(`check-git-clean`, `check-branch`). Note it does **not** run `make verify` —
CI test workflows run on push, but the tag target itself only checks git
state; run `make verify` yourself before tagging.

The Release workflow then:
1. `release-linux` (ubuntu-22.04 — pinned for glibc compatibility, PR #3):
   builds the frontend (`make package-ui`) and the viewer stylesheet
   (`make viewer-css`), then goreleaser for linux amd64 (v1–v4 microarchs)
   + arm64 and windows amd64, creates the GitHub release, plus a `.deb`
   via nfpm.
2. `release-darwin` (needs #1): appends macOS artifacts using
   `.goreleaser-darwin.yaml`.

Version metadata is injected via ldflags into `app/metainfo` (Version,
BuildTime, ShaVer).

## Traps

- **CGO is required** (`CGO_ENABLED=1`, litehtml-go) — cross-compiles need the
  toolchains the workflow installs (aarch64-gnu, mingw-w64). A plain
  `GOOS=... go build` on your machine will not reproduce release binaries.
- **The embedded UI is a build artifact.** `app/spa/files/ui/` is populated by
  `make package-ui` from `webui/dist/` (built with `VITE_BASE="/ui"`). Both
  directories are committed but generated — never hand-edit them; a Go-only
  change still ships whatever UI was last packaged, so rebuild the UI when
  frontend code changed.
- **The viewer stylesheet is the same trap, one layer down.**
  `internal/dashboard/browser/assets/viewer.css` is Tailwind output produced by
  `make viewer-css` and embedded into the binary. It is committed, so a release
  builds fine without regenerating it — and would silently ship a stale
  stylesheet if `page.html` or any widget `browser.html` changed since the last
  local run. Both release jobs therefore run `make viewer-css` after
  `make package-ui`. Rerun it locally whenever you touch a browser template
  (see [viewer.md](viewer.md)).
- Local snapshot build: `make build` (goreleaser `--snapshot --single-target`,
  includes the frontend). `make clean` removes `dist/`.
- Changelog excludes `docs:` and `test:` commit prefixes — use conventional
  prefixes (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `trivial:`) as the
  history does.

See [testing.md](testing.md) for the verify gate to run before tagging.
