# Project: gocost-web

## Project Overview
`gocost-web` is a web application designed to track monthly expenses. It is built using the **Go** programming language, following **Clean Architecture** principles. The frontend utilizes a modern stack comprising **Templ**, **Tailwind CSS**, **HTMX**, and **Alpine.js**.

### Architecture
The project follows a standard Go project layout with a strong separation of concerns:
*   **`cmd/web/`**: Web application entry point.
*   **`cmd/cli/`**: CLI application entry point (used for migrations and other tasks).
*   **`internal/domain/`**: Core business logic and entities (Expense, User, Income, Tracking).
*   **`internal/usecase/`**: Business use cases and repository interfaces.
*   **`internal/interfaces/web/`**: Web layer (HTTP handlers, routers, middleware, forms, views).
*   **`internal/infrastructure/`**: Infrastructure implementations (Config via `viper`, Storage via `sqlite`).
*   **`internal/platform/`**: Cross-cutting concerns (Money, Security/Hasher, Identifiers).
*   **`ui/`**: Frontend resources (Templ components, static assets, Tailwind configuration).
*   **`migrations/`**: SQL database migrations managed by `goose`.

### Tech Stack
*   **Backend:** Go 1.25+
    *   **Router:** Standard `net/http` mux with `justinas/alice` for middleware chaining.
    *   **Database:** SQLite with `mattn/go-sqlite3`.
    *   **Migrations:** `goose` (embedded via `internal/infrastructure/storage/sqlite/connection.go`).
    *   **Config:** `spf13/viper` for environment configuration.
    *   **Logging:** `log/slog`.
    *   **Templating:** `a-h/templ`.
*   **Frontend:**
    *   **Styling:** Tailwind CSS v4.
    *   **Interactivity:** HTMX (server-driven interactions) & Alpine.js (client-side state).
*   **Tooling:** `Make`, `Air` (live reload), `Docker`, `direnv`.

## Building and Running

The project uses a `Makefile` to manage common tasks.

### Prerequisites
*   Go 1.25+
*   Node.js & npm (for Tailwind CSS)
*   `direnv` (recommended for environment management)
*   `1Password CLI` (optional, for secret injection via `op inject`)

### Key Commands

| Command | Description |
| :--- | :--- |
| `make init` | Initialize project: download mods, generate templ files, setup `.envrc`, install npm packages. |
| `make dev` | **Primary Dev Command:** Runs server, Templ, and Tailwind watchers in parallel with live reload. |
| `make build/web` | Compiles the web server binary to `bin/server`. |
| `make build/cli` | Compiles the CLI binary to `bin/gocost`. |
| `go run ./cmd/cli migrate` | Applies database migrations. |
| `make test` | Runs the Go test suite. |
| `make check` | Runs `go vet` and `staticcheck`. |
| `make vuln` | Runs `govulncheck`. |
| `make docker/run` | Runs the application in a Docker container. |

## Development Conventions

### Development Tips
*   **Environment Variables:** Local development uses `.envrc`. If you have 1Password CLI, use `make secrets` to populate it from `envrc.template`.
*   **Live Reload:** Always use `make dev` for a smooth development experience. It handles Templ regeneration and CSS building automatically.
*   **Database:** All database changes MUST be done via migration files in `migrations/`. Use `go run ./cmd/cli migrate` to apply them.
*   **Dependency Injection:** Dependencies are wired in `main.go`. `ApplicationContext` and `UseCases` are injected into Handlers.

### Testing Instructions
*   **Library:** Use `github.com/stretchr/testify/assert` (or `require`) for assertions.
*   **Patterns:** Follow **Arrange-Act-Assert**. Use `t.Run()` for subtests and table-driven tests.
*   **Running Tests:** Run `make test` frequently. To run a specific test: `go test ./internal/... -run TestName`.
*   **Race Conditions:** Use `make test/race` to check for data races.

### Code Style & PRs
*   **Linting:** Run `make check` before committing.
*   **Vulnerabilities:** Run `make vuln` to ensure no known vulnerabilities are introduced.
*   **Formatting:** Use `go fmt`.
*   **Commits:** Prefer clear, concise messages focusing on "why".

### Security
*   **Secrets:** NEVER commit secrets or API keys. Use `envrc.template` for placeholders and `.envrc` (ignored by git) for actual values.
*   **CSRF:** The project uses `justinas/nosurf` for CSRF protection. Ensure forms include the CSRF token.