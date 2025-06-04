package main

import (
	"log"
	"os"

	"owlistic-notes/owlistic/config"
	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
)

func main() {
	// Load configuration
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Setup(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Find and fix duplicates
	if err := fixDuplicates(db); err != nil {
		log.Fatalf("Failed to fix duplicates: %v", err)
	}

	log.Println("Duplicate fix completed successfully")
}

func fixDuplicates(db *database.Database) error {
	// Find all duplicate note_ids
	type DuplicateInfo struct {
		NoteID uuid.UUID
		Count  int64
	}

	var duplicates []DuplicateInfo
	err := db.DB.Table("ai_enhanced_notes").
		Select("note_id, COUNT(*) as count").
		Group("note_id").
		Having("COUNT(*) > 1").
		Scan(&duplicates).Error

	if err != nil {
		return err
	}

	if len(duplicates) == 0 {
		log.Println("No duplicates found in ai_enhanced_notes table")
		return nil
	}

	log.Printf("Found %d note_ids with duplicates", len(duplicates))

	// Process each duplicate
	for _, dup := range duplicates {
		log.Printf("Processing duplicates for note_id: %s (count: %d)", dup.NoteID, dup.Count)

		// Get all entries for this note_id, ordered by updated_at DESC
		var entries []models.AIEnhancedNote
		err := db.DB.Where("note_id = ?", dup.NoteID).
			Order("updated_at DESC, created_at DESC").
			Find(&entries).Error

		if err != nil {
			log.Printf("Error fetching entries for note_id %s: %v", dup.NoteID, err)
			continue
		}

		if len(entries) <= 1 {
			continue
		}

		// Keep the first (most recent) entry, delete the rest
		toKeep := entries[0]
		log.Printf("Keeping entry created at %s, updated at %s", 
			toKeep.CreatedAt.Format("2006-01-02 15:04:05"),
			toKeep.UpdatedAt.Format("2006-01-02 15:04:05"))

		// Delete the older entries
		for i := 1; i < len(entries); i++ {
			entry := entries[i]
			log.Printf("Deleting duplicate entry created at %s", 
				entry.CreatedAt.Format("2006-01-02 15:04:05"))

			err := db.DB.Where("note_id = ? AND created_at = ?", 
				entry.NoteID, entry.CreatedAt).
				Delete(&models.AIEnhancedNote{}).Error

			if err != nil {
				log.Printf("Error deleting duplicate entry: %v", err)
			}
		}
	}

	// Verify no duplicates remain
	var remainingDuplicates []DuplicateInfo
	err = db.DB.Table("ai_enhanced_notes").
		Select("note_id, COUNT(*) as count").
		Group("note_id").
		Having("COUNT(*) > 1").
		Scan(&remainingDuplicates).Error

	if err != nil {
		return err
	}

	if len(remainingDuplicates) > 0 {
		log.Printf("WARNING: %d note_ids still have duplicates", len(remainingDuplicates))
		for _, dup := range remainingDuplicates {
			log.Printf("  - note_id: %s (count: %d)", dup.NoteID, dup.Count)
		}
	} else {
		log.Println("All duplicates successfully removed")
	}

	return nil
}