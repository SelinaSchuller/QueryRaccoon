# QueryRaccoon — Planning Document

A cross-platform desktop database management app for PostgreSQL, MySQL, SQLite, and MSSQL.
Inspired by DBeaver's features and Beekeeper Studio's UI.

---

## Stack

| Layer              | Technology                                                                        |
| ------------------ | --------------------------------------------------------------------------------- |
| Desktop framework  | [Wails v2](https://wails.io) — Go backend + web frontend → native `.app` / `.exe` |
| Backend language   | Go                                                                                |
| Frontend framework | Svelte (plain, no SvelteKit — desktop apps don't need URL routing)                |
| Build tool         | Vite                                                                              |
| Styling            | Tailwind CSS                                                                      |
| UI primitives      | bits-ui (headless Svelte components: dialogs, dropdowns, tooltips)                |
| Icons              | lucide-svelte                                                                     |

---

## Database Drivers (Go)

| Database   | Package                           | Notes                                                      |
| ---------- | --------------------------------- | ---------------------------------------------------------- |
| PostgreSQL | `github.com/jackc/pgx/v5`         | Best-in-class PG driver                                    |
| MySQL      | `github.com/go-sql-driver/mysql`  | Handles MySQL 5.7+, 8.x, MariaDB                           |
| SQLite     | `modernc.org/sqlite`              | Pure Go, no CGO — required for cross-platform Wails builds |
| MSSQL      | `github.com/microsoft/go-mssqldb` | Microsoft's official Go driver                             |

---

## Frontend Libraries

| Feature               | Library                                                                    |
| --------------------- | -------------------------------------------------------------------------- |
| SQL editor            | CodeMirror 6 (`@codemirror/lang-sql`, `@codemirror/autocomplete`)          |
| Data grid             | TanStack Table v8 + TanStack Virtual (headless, handles large result sets) |
| ER diagrams           | @xyflow/svelte (Svelte Flow) + @dagrejs/dagre for auto-layout              |
| EXPLAIN visualization | D3.js (custom tree/flame graph)                                            |
| Date formatting       | date-fns                                                                   |

---

## Key Go Packages

| Purpose                        | Package                         |
| ------------------------------ | ------------------------------- |
| SSH tunneling                  | `golang.org/x/crypto/ssh`       |
| OS keychain (password storage) | `github.com/zalando/go-keyring` |
| Unique IDs                     | `github.com/google/uuid`        |
| Concurrency helpers            | `golang.org/x/sync/errgroup`    |

---

## Platforms

- **macOS** — `.app` bundle, code signed + notarized via `xcrun notarytool`
- **Windows** — `.exe` + NSIS installer, code signed via `signtool`

---

## Features

### Connection Management

- Save, edit, delete connections
- Connection groups / folders
- SSH tunnel support
- SSL/TLS certificate configuration
- Passwords stored in OS keychain (Keychain on Mac, Credential Manager on Windows)

### Schema Browser (left sidebar tree)

- Connection → Database → Schema → Tables / Views / Procedures / Functions / Triggers
- Lazy-loaded (children fetched only when a node is expanded)
- Right-click context menu: Open, New Query, Edit Table, Drop, Refresh

### Query Editor

- SQL syntax highlighting (dialect-aware per connection)
- Autocomplete: keywords, table names, column names (from live schema)
- Multi-tab support
- Run selected text or full editor content
- Query cancellation
- Transaction control (Begin / Commit / Rollback)
- Find & replace

### Data Grid

- View table data with sorting, filtering, pagination
- Inline cell editing with staged changes (apply / discard)
- Column resize and reorder
- Row virtualization (handles large result sets without freezing)
- Type-aware cell rendering (timestamps, JSON, NULL, binary)

### Schema Editor

- Create / alter / drop tables
- Add / edit / remove columns (type, nullable, default, constraints)
- Index manager (create / drop, unique constraints)
- Foreign key editor
- DDL preview before applying changes

### ER Diagram

- Auto-layout via dagre
- Pan, zoom, minimap
- Export as PNG

### EXPLAIN / Execution Plan

- Parse EXPLAIN (ANALYZE) output
- Visual tree + flame graph

### Import / Export

- Export table data: CSV, JSON
- SQL dump export
- CSV / JSON import

### Database Objects

- View and edit stored procedures, views, triggers
- Source shown in CodeMirror editor tab

### User Management

- List users and roles
- Create / edit / drop users
- Grant / revoke permissions

### Query History

- All executed queries logged with timestamp, connection, duration
- Searchable, re-runnable from sidebar

### App

- Dark and light theme
- Keyboard shortcuts (Cmd+T new tab, Cmd+Enter run query, Cmd+Shift+F format SQL)
- Settings page (default query limit, font size, editor preferences)
- Connection export / import (`.qrconn` files for sharing between machines)

---

## Project Structure

```
QueryRaccoon/
├── main.go
├── wails.json
├── go.mod
│
├── internal/
│   ├── connections/       # Connection manager, SSH tunnels, SSL, storage
│   ├── drivers/           # Driver interface + per-DB implementations
│   │   ├── driver.go      # Common interface all drivers implement
│   │   ├── postgres/
│   │   ├── mysql/
│   │   ├── sqlite/
│   │   └── mssql/
│   ├── query/             # Query execution, history, formatting
│   ├── schema/            # Schema introspection, DDL generation
│   ├── dataops/           # Import / export
│   ├── users/             # User and role management
│   └── storage/           # Local SQLite for app state (saved connections, history)
│
├── bindings/              # Wails-exposed services (called from Svelte via generated JS)
│   ├── connection_service.go
│   ├── query_service.go
│   ├── schema_service.go
│   ├── data_service.go
│   ├── export_service.go
│   └── user_service.go
│
└── frontend/
    └── src/
        ├── lib/
        │   ├── wailsjs/       # Auto-generated by Wails (do not edit)
        │   ├── api/           # Typed wrappers around wailsjs
        │   ├── stores/        # Svelte stores: connections, tabs, schema tree, theme
        │   └── components/
        │       ├── layout/    # AppShell, Sidebar, TabBar, StatusBar
        │       ├── connections/
        │       ├── schema-browser/
        │       ├── editor/    # QueryEditor (CodeMirror), ResultsPane, ExplainViewer
        │       ├── data-grid/
        │       ├── schema-editor/
        │       ├── er-diagram/
        │       ├── objects/   # Procedures, views, triggers
        │       ├── users/
        │       └── shared/    # Button, Modal, ContextMenu, Splitter, etc.
        └── App.svelte
```

---

## Build Phases

### Phase 1 — Foundation

Running app that connects to a DB and executes queries.

- Wails project scaffold, Go module, Svelte + Tailwind wired up
- App layout: sidebar, tab bar, status bar
- PostgreSQL connection (host/port/user/pass, no SSH yet)
- Query editor (CodeMirror) + basic data grid
- Passwords saved to OS keychain

### Phase 2 — Core Features

Useful for daily work.

- MySQL, SQLite, MSSQL drivers added
- Schema tree with lazy loading
- Data grid: sorting, filtering, pagination, inline editing, virtual rows
- Multi-tab support
- Query history
- SQL autocomplete from live schema
- Transaction control + query cancellation

### Phase 3 — Schema Management

Manage schema without touching the CLI.

- Table / column / index / FK editor with DDL preview
- SSH tunnel + SSL support in connection dialog
- Context menus on tree nodes

### Phase 4 — Advanced Features

Power user capabilities.

- ER diagram viewer
- EXPLAIN visualization
- Import / export (CSV, JSON, SQL dump)
- Stored procedures, views, triggers editor
- User / role management

### Phase 5 — Polish & Distribution

Shippable product.

- Dark / light theme + settings page
- Keyboard shortcuts
- Mac code signing + notarization
- Windows installer
- Auto-updater (GitHub Releases)
- Connection file import / export
