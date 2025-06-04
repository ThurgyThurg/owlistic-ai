-- Script to fix duplicate note_id entries in ai_enhanced_notes table
-- This script keeps only the most recent entry for each note_id

-- First, identify and display duplicates
SELECT 
    note_id, 
    COUNT(*) as duplicate_count,
    array_agg(created_at ORDER BY created_at DESC) as created_dates
FROM ai_enhanced_notes
GROUP BY note_id
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

-- Create a temporary table with the IDs to keep (most recent per note_id)
CREATE TEMP TABLE ai_enhanced_notes_to_keep AS
SELECT DISTINCT ON (note_id) 
    note_id,
    created_at,
    updated_at
FROM ai_enhanced_notes
ORDER BY note_id, updated_at DESC, created_at DESC;

-- Delete duplicates, keeping only the most recent entry
DELETE FROM ai_enhanced_notes
WHERE (note_id, created_at) NOT IN (
    SELECT note_id, created_at 
    FROM ai_enhanced_notes_to_keep
);

-- Verify no duplicates remain
SELECT 
    note_id, 
    COUNT(*) as count
FROM ai_enhanced_notes
GROUP BY note_id
HAVING COUNT(*) > 1;

-- Drop the temporary table
DROP TABLE ai_enhanced_notes_to_keep;

-- Now the unique constraint can be safely added
-- This will be done by the migration, but you can manually run:
-- ALTER TABLE ai_enhanced_notes ADD CONSTRAINT ai_enhanced_notes_note_id_unique UNIQUE (note_id);