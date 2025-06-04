# Fixing Duplicate note_id Entries in ai_enhanced_notes Table

## Problem Description

The `ai_enhanced_notes` table may contain duplicate entries for the same `note_id`, which violates the intended one-to-one relationship between notes and their AI enhancements. This can cause the error:

```
pq: duplicate key value violates unique constraint "ai_enhanced_notes_pkey"
```

## Solution

We've implemented several fixes:

1. **Updated the database migrations** to add a unique constraint on `note_id`
2. **Created scripts** to clean up existing duplicates
3. **Modified the model** to ensure GORM creates the constraint properly

## Steps to Fix Existing Duplicates

### Option 1: Run the Go Script (Recommended)

From the backend directory, run:

```bash
cd src/backend
go run cmd/fix_duplicates/main.go
```

This script will:
- Find all duplicate `note_id` entries
- Keep the most recently updated entry for each `note_id`
- Delete older duplicate entries
- Verify that all duplicates have been removed

### Option 2: Run the SQL Script

If you have direct database access, you can run:

```bash
psql -h localhost -U $DB_USER -d $DB_NAME -f scripts/fix_ai_enhanced_notes_duplicates.sql
```

This SQL script performs the same cleanup operations directly in the database.

## What the Migration Does

The updated migration (`runManualMigrations` function in `migrations.go`) will:

1. Add a unique constraint on `note_id` (if it doesn't already exist)
2. Create indexes for better query performance:
   - Index on `note_id`
   - Index on `processing_status`
   - Composite index on `(note_id, processing_status)`

## Prevention

The unique constraint will prevent future duplicates from being created. The application code already uses `ON CONFLICT` clauses for upserts, so it will properly update existing records instead of creating duplicates.

## Verification

After running the fix, you can verify no duplicates remain by running:

```sql
SELECT note_id, COUNT(*) as count
FROM ai_enhanced_notes
GROUP BY note_id
HAVING COUNT(*) > 1;
```

This query should return no rows if all duplicates have been successfully removed.

## When to Run

You should run the duplicate cleanup script:
- Before restarting the application with the new migration
- If you encounter duplicate key errors
- As part of database maintenance

The migration will automatically run when the application starts, but it requires that duplicates be removed first for the unique constraint to be added successfully.