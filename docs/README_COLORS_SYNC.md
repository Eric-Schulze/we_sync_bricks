# BrickLink Colors Sync

This system allows you to download and sync BrickLink colors to your local database. The colors can be updated at any time using the make commands.

## Usage

### Sync Colors from BrickLink API
```bash
make sync-colors
```
This downloads all colors from the BrickLink `/colors` endpoint and syncs them to your database.

### Show Current Sync Information
```bash
make sync-colors-info
```
Displays information about the current state of colors in your database, including total count and last sync time.

### Verbose Sync (for debugging)
```bash
make sync-colors-verbose
```
Same as `sync-colors` but with detailed debug logging.

## API Response Mapping

The BrickLink colors API returns data in this format:
```json
{
    "meta": {
        "description": "OK",
        "message": "OK", 
        "code": 200
    },
    "data": [
        {
            "color_id": 1,
            "color_name": "White",
            "color_code": "FFFFFF", 
            "color_type": "Solid"
        }
    ]
}
```

This gets mapped to your existing database schema:
- `color_id` → `bricklink_id`
- `color_name` → `name` 
- `color_code` → `code`
- `color_type` → `type`
- `rebrickable_id` is set to `NULL` (not available from BrickLink API)

## Database Table Structure

The system works with your `colors` table (defined in `00001_initial_migration.sql`):
```sql
CREATE TABLE colors (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bricklink_id INTEGER NOT NULL,
    rebrickable_id INTEGER NOT NULL, 
    name TEXT NOT NULL,
    code TEXT,
    type TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
```

**Constraint Behavior:**
- `bricklink_id`: Must be unique across all colors
- `rebrickable_id`: Can be NULL for BrickLink-only colors, but must be unique when not NULL
- This allows multiple BrickLink colors to coexist without rebrickable mappings

**Timestamp Behavior:**
- `created_at`: Set automatically when a color is first inserted
- `updated_at`: Set to current timestamp whenever a color is updated during sync

## How It Works

1. **API Client**: Added `GetColors()` method to BrickLink catalog client
2. **Response Parsing**: `ParseColorsResponse()` converts JSON to Go structs
3. **Data Mapping**: `BricklinkColor.ToDatabase()` maps API response to database format
4. **Sync Service**: `ColorsSyncService` handles the database operations (insert/update)
5. **Command Tool**: Standalone command that can be run independently
6. **Make Commands**: Easy-to-use make targets for common operations

## Advanced Usage

You can also run the sync command directly with options:
```bash
# Run with specific user ID
go run cmd/sync-colors/main.go -user 123

# Show info only
go run cmd/sync-colors/main.go -info

# Verbose mode
go run cmd/sync-colors/main.go -v
```

## Integration

The colors sync system integrates with your existing:
- Database connection system
- BrickLink client manager
- OAuth credential management 
- Logging system

Colors will be automatically available for use in your partial minifig parts system and other features that reference BrickLink color IDs.