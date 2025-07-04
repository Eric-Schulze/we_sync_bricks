# Build commands
build:
	go build -o bin/main

clean:
	rm -rf bin/ tmp/

# Development commands
dev:
	$(shell go env GOPATH)/bin/air

run:
	go run main.go

# CSS build targets
build-css:
	npx tailwindcss -i ./web/static/css/src/main.css -o ./web/static/css/dist/main.css --watch

build-css-prod:
	npx tailwindcss -i ./web/static/css/src/main.css -o ./web/static/css/dist/main.css --minify

# Database commands
fix_db_permissions:
	chmod 700 "/Users/Eric/Documents/Dev/we_sync_bricks/data"

check_db_status:
	@pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" status > /dev/null 2>&1

start_db: fix_db_permissions
	@if pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" status > /dev/null 2>&1; then \
		echo "PostgreSQL server is already running"; \
	else \
		echo "Starting PostgreSQL server..."; \
		pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" -l logfile start; \
	fi

connect_db:
	PGPASSWORD=postgres psql -U postgres -d postgres -h localhost -p 5432

stop_db:
	@if pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" status > /dev/null 2>&1; then \
		echo "Stopping PostgreSQL server..."; \
		pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" stop; \
	else \
		echo "PostgreSQL server is not running"; \
	fi

restart_db: stop_db start_db

status_db:
	@pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" status

migrate_db:
	PGPASSWORD=postgres go run cmd/migrate/main.go

# Database backup and restore commands
dump_db:
	@echo "Creating database dump..."
	@mkdir -p backups
	@PGPASSWORD=postgres pg_dump -U postgres -h localhost -p 5432 -d postgres --clean --if-exists --create > backups/db_dump_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Database dumped to backups/db_dump_$(shell date +%Y%m%d_%H%M%S).sql"

dump_db_data_only:
	@echo "Creating data-only database dump..."
	@mkdir -p backups
	@PGPASSWORD=postgres pg_dump -U postgres -h localhost -p 5432 -d postgres --data-only > backups/db_data_dump_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Data dump created in backups/"

dump_db_schema_only:
	@echo "Creating schema-only database dump..."
	@mkdir -p backups
	@PGPASSWORD=postgres pg_dump -U postgres -h localhost -p 5432 -d postgres --schema-only > backups/db_schema_dump_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Schema dump created in backups/"

drop_db:
	@echo "WARNING: This will completely destroy the database!"
	@echo "Are you sure? Press Ctrl+C to cancel, or Enter to continue..."
	@read confirmation
	@echo "Stopping database server..."
	@$(MAKE) stop_db
	@echo "Removing database data directory..."
	@rm -rf "/Users/Eric/Documents/Dev/we_sync_bricks/data"
	@echo "Database dropped successfully"

drop_db_force:
	@echo "Forcefully dropping database without confirmation..."
	@echo "Stopping database server..."
	@$(MAKE) stop_db
	@echo "Removing database data directory..."
	@rm -rf "/Users/Eric/Documents/Dev/we_sync_bricks/data"
	@echo "Database dropped successfully"

create_postgres_user:
	@echo "Creating postgres superuser with password..."
	@if pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" status > /dev/null 2>&1; then \
		psql -U postgres -d postgres -c "ALTER USER postgres PASSWORD 'postgres';" 2>/dev/null || echo "Password set successfully"; \
		echo "Postgres user updated with password successfully"; \
	else \
		echo "Error: Database server is not running. Start it first with 'make start_db'"; \
		exit 1; \
	fi

create_postgres_user_force:
	@echo "Creating postgres superuser with password..."
	@psql -U postgres -d postgres -c "DROP USER IF EXISTS postgres;" 2>/dev/null || echo "User postgres doesn't exist yet"
	@psql -U postgres -d postgres -c "CREATE USER postgres WITH SUPERUSER PASSWORD 'postgres';" 2>/dev/null || echo "User postgres already exists"
	@psql -U postgres -d postgres -c "ALTER USER postgres PASSWORD 'postgres';" 2>/dev/null || echo "Password already set"
	@echo "Postgres user created/updated with password successfully"

init_db:
	@echo "Initializing new PostgreSQL database..."
	@initdb -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" -U postgres --auth-local=trust --auth-host=md5
	@echo "Starting database to create user..."
	@pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" -l logfile start
	@sleep 2
	@$(MAKE) create_postgres_user_force
	@echo "Restarting database to apply authentication settings..."
	@pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" restart
	@sleep 2
	@echo "Database initialized successfully"

recreate_db: drop_db init_db
	@echo "Database recreated successfully"
	@echo "Testing database connection..."
	@$(MAKE) test_db_connection
	@echo "Showing database users:"
	@$(MAKE) show_users
	@echo "Run 'make migrate_db' to apply schema migrations"

recreate_db_force: drop_db_force init_db
	@echo "Database recreated successfully"
	@echo "Testing database connection..."
	@$(MAKE) test_db_connection
	@echo "Showing database users:"
	@$(MAKE) show_users
	@echo "Run 'make migrate_db' to apply schema migrations"

setup_db_complete: recreate_db migrate_db
	@echo "Complete database setup finished!"
	@echo "Database is ready for development"

setup_db_complete_force: recreate_db_force migrate_db
	@echo "Complete database setup finished!"
	@echo "Database is ready for development"

load_db:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make load_db FILE=path/to/dump.sql"; \
		echo "Available dumps in backups/:"; \
		ls -la backups/ 2>/dev/null || echo "No backups directory found"; \
		exit 1; \
	fi
	@if [ ! -f "$(FILE)" ]; then \
		echo "Error: File $(FILE) not found"; \
		exit 1; \
	fi
	@echo "Loading database from $(FILE)..."
	@echo "WARNING: This will replace current database content!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read confirmation
	@PGPASSWORD=postgres psql -U postgres -h localhost -p 5432 -d postgres < "$(FILE)"
	@echo "Database loaded successfully from $(FILE)"

# Quick backup before dangerous operations
backup_before_drop: dump_db
	@echo "Backup created before dropping database"

# Debug and test commands
test_db_connection:
	@echo "Testing database connection..."
	@echo "Trying connection with password..."
	@PGPASSWORD=postgres psql -U postgres -h localhost -p 5432 -d postgres -c "SELECT current_user, version();" || echo "Connection failed"

show_pg_hba:
	@echo "Current pg_hba.conf settings:"
	@cat "/Users/Eric/Documents/Dev/we_sync_bricks/data/pg_hba.conf" | grep -v "^#" | grep -v "^$$"

show_users:
	@echo "Current database users:"
	@psql -U postgres -d postgres -c "SELECT usename, usesuper, passwd IS NOT NULL as has_password FROM pg_user;" 2>/dev/null || echo "Could not connect to database"

show_current_user:
	@echo "Current connection info:"
	@psql -U postgres -d postgres -c "SELECT current_user, current_database();" 2>/dev/null || echo "Could not connect to database"

debug_connection:
	@echo "=== Debug Connection Information ==="
	@echo "Trying connection as postgres user..."
	@psql -U postgres -d postgres -c "SELECT current_user, version();" 2>&1 || echo "Failed to connect as postgres"
	@echo "Trying connection without specifying user..."
	@psql -d postgres -c "SELECT current_user, version();" 2>&1 || echo "Failed to connect without user"
	@echo "Checking if database is running..."
	@pg_ctl -D "/Users/Eric/Documents/Dev/we_sync_bricks/data" status || echo "Database not running"

# Testing commands
test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

# Docker commands (if needed)
docker-build:
	docker build -t we_sync_bricks .

docker-run:
	docker compose up --build

docker-down:
	docker compose down

# Combined commands
setup: tidy fmt vet
	@echo "Setup complete"

all: clean fmt vet test build
	@echo "All tasks completed"

.PHONY: build clean dev run build-css build-css-prod fix_db_permissions check_db_status start_db connect_db stop_db restart_db status_db migrate_db dump_db dump_db_data_only dump_db_schema_only drop_db drop_db_force create_postgres_user create_postgres_user_force init_db recreate_db recreate_db_force setup_db_complete setup_db_complete_force load_db backup_before_drop test_db_connection show_pg_hba show_users show_current_user debug_connection test fmt vet tidy docker-build docker-run docker-down setup all