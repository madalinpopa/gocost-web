# Project: gocost-web

## Project Overview
`gocost-web` is a web application designed to track monthly expenses. It is built using the **Go** programming language, following **Clean Architecture** principles. The frontend utilizes a modern stack comprising **Templ**, **Tailwind CSS**, **HTMX**, and **Alpine.js**.

### Architecture
The project follows a standard Go project layout with a strong separation of concerns:
*   **`cmd/web/`**: Application entry point (`main.go`).
*   **`internal/domain/`**: Contains core business logic and entities (e.g., `Expense`, `User`).
*   **`internal/usecase/`**: Defines business use cases and repository interfaces (currently scaffolding).
*   **`internal/interfaces/web/`**: The web layer, including HTTP handlers, routers, middleware, and response helpers.
*   **`internal/infrastructure/`**: Infrastructure implementations such as Configuration (`viper`) and Storage (`sqlite`).
*   **`ui/`**: Frontend resources including Templ components, static assets, and Tailwind configuration.
*   **`migrations/`**: SQL database migrations managed by `goose`.

### Tech Stack
*   **Backend:** Go (1.25+)
    *   **Router:** Standard `net/http` mux with `justinas/alice` for middleware chaining.
    *   **Database:** SQLite with `mattn/go-sqlite3`.
    *   **Config:** `spf13/viper` for environment configuration.
    *   **Logging:** `log/slog`.
    *   **Templating:** `a-h/templ`.
*   **Frontend:**
    *   **Styling:** Tailwind CSS v4.
    *   **Interactivity:** HTMX (server-driven interactions) & Alpine.js (client-side state).
*   **Tooling:** `Make`, `Air` (live reload), `Docker`.

## Building and Running

The project uses a `Makefile` to manage common tasks.

### Prerequisites
*   Go 1.25+
*   Node.js & npm (for Tailwind CSS)
*   `direnv` (recommended for environment management)

### Key Commands

| Command | Description |
| :--- | :--- |
| `make init` | Initialize the project: download Go mods, install npm packages, generate templ files, and setup `.envrc`. |
| `make dev` | **Primary Dev Command:** Runs the server, Templ watcher, and Tailwind watcher in parallel with live reload (Air). |
| `make build/web` | Compiles the web server binary to `bin/server`. |
| `make test` | Runs the Go test suite (`./internal...`). |
| `make db/migrate` | *Note: Inferred from context, check Makefile.* Applies database migrations. |
| `make docker/run` | Runs the application in a Docker container. |

## Development Conventions

### Code Structure
*   **Dependency Injection:** The `main.go` file wires up the application. `ApplicationContext` (config, logger, decoder) and `UseCases` are injected into Handlers.
*   **Configuration:** Environment variables are loaded via `viper`. Local development uses `.envrc` (generated from `envrc.template`).
*   **Database:** All database changes must be done via migration files in `migrations/`.

### Frontend Workflow
1.  **Templ:** UI components are written in `.templ` files. The `make dev` command watches these files and regenerates the corresponding Go code automatically.
2.  **Tailwind:** CSS is defined in `ui/static/css/input.css`. The build process generates `ui/static/css/output.css`.
3.  **HTMX:** Used for dynamic partial page updates. Check `ui/templates/layouts/base.templ` for global HTMX configuration.

### Testing
*   **Library:** Use `github.com/stretchr/testify/assert` (or `require`) for assertions instead of standard `if` checks.
*   **Structure:** Follow the **Arrange-Act-Assert** pattern within your tests.
*   **Subtests:** Use `t.Run()` for table-driven tests or distinct scenarios.

**Example:**
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
    t.Run("successful scenario", func(t *testing.T) {
        // Arrange
        input := "valid input"
        expected := "result"

        // Act
        result, err := MyFunction(input)

        // Assert
        assert.NoError(t, err)
        assert.Equal(t, expected, result)
    })
}
```
