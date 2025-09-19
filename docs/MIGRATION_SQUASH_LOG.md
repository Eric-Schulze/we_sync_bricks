# Migration Squash Log

**Date**: 2025-08-01

## What Was Squashed

The following migrations were consolidated into `00001_initial_migration.sql`:

### Original Migrations (DELETED)
1. **00002_add_timestamps_to_colors.sql** - Added `created_at` and `updated_at` columns
2. **00003_allow_null_rebrickable_id.sql** - Made `rebrickable_id` nullable and updated constraints

### Result
All changes were merged into the initial migration to create a clean, single migration that:

1. **Creates colors table with correct structure**:
   - `rebrickable_id INTEGER` (nullable from the start)
   - `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
   - `updated_at TIMESTAMP`

2. **Sets up proper constraints**:
   - Unique constraint on `bricklink_id`
   - Partial unique index on `rebrickable_id` (ignores NULLs)

3. **Adds performance indexes**:
   - Index on `created_at`
   - Index on `updated_at`

4. **Includes column documentation**:
   - Comments explaining the purpose of each column

## Benefits

- **Cleaner migration history**: Single migration instead of 3 separate ones
- **Easier deployment**: New environments get correct schema from start
- **No migration dependencies**: All colors table changes in one place
- **Better documentation**: Comments and constraints defined upfront

## Remaining Migrations

These migrations were **NOT** affected and remain separate:
- `003_create_user_oauth_credentials.sql`
- `004_add_partial_minifig_fields.sql`

## Verification

✅ All builds pass
✅ Migration system works correctly  
✅ BrickLink colors sync system functions properly
✅ Database constraints work as expected