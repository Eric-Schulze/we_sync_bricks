# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

- **Development (Live Reload)**: `make dev` - Uses Air for automatic rebuilding on file changes
- **Build**: `make build` or `go build -o bin/main`
- **Run**: `make run` or `go run main.go`
- **Test**: `go test ./...` or `make test`
- **Format**: `go fmt ./...` or `make fmt`
- **Vet**: `go vet ./...` or `make vet`
- **Tidy modules**: `go mod tidy` or `make tidy`
- **Clean**: `make clean` - Remove build artifacts

### Live Reloading with Air

The project uses [Air](https://github.com/air-verse/air) for live reloading during development:

- **Start development server**: `make dev`
- **Configuration**: `.air.toml` (watches Go, HTML, CSS, JS, YAML files)
- **Automatic rebuilds**: Changes to code or templates trigger immediate rebuilds
- **Build output**: Temporary binaries are stored in `tmp/` directory

## Architecture Overview

This is a LEGO marketplace synchronization service that integrates with BrickLink and BrickOwl APIs. The application follows a modular architecture with clear separation of concerns:

### Core Structure
- **Entry Point**: `main.go` delegates to `internal/init/start.go`
- **Configuration**: Uses Viper for config management with `internal/init/config.yaml`
- **Web Server**: Chi router on port 4000 with HTML templates in `web/templates/`
- **Database**: PostgreSQL with GORM ORM and migration system

### Service Layer Architecture
- **Services**: Located in `internal/common/services/`
  - `bricklink/`: BrickLink API integration (inventory, orders, catalog)
  - `brickowl/`: BrickOwl API integration
  - `db/`: Database services with PostgreSQL
  - `auth/`: Authentication middleware
- **Models**: Common data models in `internal/common/models/`
- **Domain Services**: Top-level services in root directories:
  - `orders/`: Order management service
  - `inventory/`: Inventory management service  
  - `partial_minifigs/`: Partial minifigure tracking service

### Key Features
- OAuth integration for external APIs
- Database migrations system
- Template-based web interface
- RESTful API endpoints
- Middleware for logging and authentication

### Database Configuration
Database settings are in `internal/init/config.yaml`:
- Host: localhost:5432
- Default database: postgres
- Migrations located in `internal/common/services/db/migrations/`

#### Database Commands
- **Start database**: `make start_db` (automatically fixes permissions & checks if running)
- **Stop database**: `make stop_db` (checks if running before stopping)
- **Restart database**: `make restart_db` (full restart sequence)
- **Check database status**: `make status_db` (shows server status and PID)
- **Connect to database**: `make connect_db` (opens psql connection)
- **Fix permissions**: `make fix_db_permissions` (manual permission fix)

#### Smart Database Management
The database commands are intelligent and handle common scenarios:
- `start_db` checks if PostgreSQL is already running before attempting to start
- `stop_db` checks if PostgreSQL is running before attempting to stop
- Both commands provide clear feedback about the current state
- Permissions are automatically fixed on each start attempt

#### PostgreSQL Permissions Issue
PostgreSQL requires strict permissions on the data directory. If you encounter:
```
FATAL: data directory has invalid permissions
DETAIL: Permissions should be u=rwx (0700) or u=rwx,g=rx (0750)
```

This is automatically fixed when using `make start_db`, or manually run:
```bash
chmod 700 data/
```

## Development Notes

The project uses Go modules with dependencies for:
- Chi router for HTTP routing
- GORM for database ORM
- Viper for configuration
- OAuth1 for API authentication
- PostgreSQL driver

The application serves both web pages (templates) and API endpoints, with static assets served from `web/static/`.