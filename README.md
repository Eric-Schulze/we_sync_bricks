# WE Sync Bricks

A LEGO marketplace synchronization service that integrates with BrickLink and BrickOwl APIs.

## 🚀 Quick Start

```bash
# Start database
make start_db

# Run migrations
make migrate_db

# Start development server
make dev
```

## 📚 Documentation

All project documentation is located in the [`docs/`](docs/) directory:

- **[Getting Started Guide](docs/CLAUDE.md)** - Setup, architecture, and development
- **[BrickLink Colors Sync](docs/README_COLORS_SYNC.md)** - How to sync BrickLink colors
- **[OAuth Architecture](docs/OAUTH_CLIENT_ARCHITECTURE.md)** - API integration details
- **[Migration History](docs/MIGRATION_SQUASH_LOG.md)** - Database migration changes

## 🔧 Key Features

- **BrickLink Integration** - OAuth-based API client with caching
- **Partial Minifig Tracking** - Track missing parts for minifigures
- **Color Sync System** - Download and sync BrickLink colors
- **Profile Management** - OAuth credential management
- **Web Interface** - HTML templates with HTMX for dynamic content

## 🗄️ Database

PostgreSQL with automated migrations. See [docs/CLAUDE.md](docs/CLAUDE.md) for database setup and management commands.

## 📖 More Information

For detailed information about development, architecture, and usage, see the [documentation directory](docs/).