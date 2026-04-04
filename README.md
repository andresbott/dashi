# dashi

A user-defined landing page dashboard application. Users can create personalized dashboards with configurable widgets such as weather, bookmarks, stock tickers, and more.

## Tech Stack

- **Backend:** Go (Cobra CLI, Gorilla Mux, GORM, slog)
- **Frontend:** Vue.js 3 (Composition API, PrimeVue, Pinia, Vue Query, Vite)
- **Structure:** Monorepo

## Project Structure

```
dashi/
├── main.go                    # Entry point
├── go.mod                     # Go module definition
├── Makefile                   # Build, test, and dev automation
├── .goreleaser.yaml           # Multi-platform release builds
├── .golangci.yaml             # Linter configuration
├── config.yaml                # App config (gitignored, generate with `dashi config`)
├── .env.example               # Environment variable template
│
├── app/                       # Application layer
│   ├── cmd/                   # CLI commands (start, version, config)
│   ├── metainfo/              # Build metadata (version, commit, date)
│   ├── router/                # HTTP routing and API endpoints
│   │   └── handlers/          # Request handlers by domain
│   └── spa/                   # Embedded SPA serving
│       └── files/ui/          # Built frontend assets (generated)
│
├── internal/                  # Business logic (not importable externally)
│   ├── dashboard/             # Dashboard CRUD and layout management
│   └── widgets/               # Widget registry and implementations
│
├── webui/                     # Vue.js frontend
│   ├── package.json           # NPM dependencies and scripts
│   ├── vite.config.js         # Vite build configuration
│   ├── vitest.config.js       # Test runner configuration
│   ├── tsconfig.json          # TypeScript configuration
│   ├── index.html             # HTML entry point
│   └── src/
│       ├── main.ts            # App bootstrap
│       ├── App.vue            # Root component
│       ├── router/            # Vue Router configuration
│       ├── store/             # Pinia stores (user, UI state)
│       ├── composables/       # Vue Query hooks and custom composables
│       ├── components/        # Shared UI components
│       ├── views/             # Page-level components
│       ├── utils/             # Helper functions
│       └── assets/            # Styles and static assets
│
├── zarf/                      # Deployment artifacts
│   ├── e2e/                   # End-to-end browser tests
│   └── pkg/                   # Package scripts (deb, systemd)
│
├── docs/                      # Documentation
│
└── .github/workflows/         # CI/CD pipelines
    ├── test.yml               # Go tests + UI tests + build
    ├── golangci-lint.yml      # Code quality linting
    ├── license-check.yml      # Dependency license compliance
    └── release.yml            # GoReleaser multi-platform release
```

## Development

### Prerequisites

- Go 1.25.4+
- Node.js 22+
- golangci-lint (for linting)

### Running

```bash
# Start backend (debug mode, built-in defaults)
make run

# Build frontend and start backend with embedded UI
make run-ui

# Frontend dev server (hot reload, proxies API to backend)
cd webui && npm run dev
```

### Testing

```bash
make test          # Go unit tests
make ui-test       # Vue.js unit tests
make lint          # Go linter
make coverage      # Enforce 70% coverage on internal packages
make verify        # Run all checks
```

### Building

```bash
make build         # Build binary for current OS/arch (includes frontend)
make package-ui    # Build frontend and embed in Go package
```

### Configuration

Generate a default config file:

```bash
dashi config
```

Configuration is loaded in order (last wins):
1. Built-in defaults
2. `.env` file (optional)
3. `config.yaml` (optional)
4. Environment variables (prefix: `DASHI_`)
