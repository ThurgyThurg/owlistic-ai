package database

import (
	"log"

	"owlistic-notes/owlistic/config"
	"owlistic-notes/owlistic/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RunMigrations runs database migrations to ensure tables are up to date
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Add all models that should be migrated
	err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Notebook{},
		&models.Note{},
		&models.Block{},
		&models.Task{},
		&models.Event{},
		// AI Enhancement models
		&models.AIEnhancedNote{},
		&models.AIAgent{},
		&models.AIAgentStep{},
		&models.AIProject{},
		&models.AITaskEnhancement{},
		&models.ChatMemory{},
		// Calendar models
		&models.GoogleCalendarCredentials{},
		&models.CalendarEvent{},
		&models.CalendarSync{},
		&models.CalendarReminder{},
		&models.CalendarAttendee{},
		// Zettelkasten models
		&models.ZettelNode{},
		&models.ZettelEdge{},
		&models.ZettelTag{},
		&models.ZettelGraph{},
		&models.ZettelAnalysis{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	// Run additional manual migrations for constraints
	if err := runManualMigrations(db); err != nil {
		log.Printf("Manual migrations failed: %v", err)
		return err
	}

	return nil
}

// runManualMigrations runs manual SQL migrations for constraints and indexes
func runManualMigrations(db *gorm.DB) error {
	log.Println("Running manual migrations for constraints...")

	// Add unique constraint on note_id in ai_enhanced_notes table
	// This ensures one-to-one relationship between notes and AI enhancements
	if err := db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint 
				WHERE conname = 'ai_enhanced_notes_note_id_unique'
			) THEN
				ALTER TABLE ai_enhanced_notes 
				ADD CONSTRAINT ai_enhanced_notes_note_id_unique 
				UNIQUE (note_id);
			END IF;
		END $$;
	`).Error; err != nil {
		log.Printf("Failed to add unique constraint on ai_enhanced_notes.note_id: %v", err)
		// Don't fail the migration if constraint already exists
	}

	// Add index on note_id for better query performance
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ai_enhanced_notes_note_id 
		ON ai_enhanced_notes(note_id);
	`).Error; err != nil {
		log.Printf("Failed to add index on ai_enhanced_notes.note_id: %v", err)
	}

	// Add index on processing_status for finding pending notes
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ai_enhanced_notes_processing_status 
		ON ai_enhanced_notes(processing_status);
	`).Error; err != nil {
		log.Printf("Failed to add index on ai_enhanced_notes.processing_status: %v", err)
	}

	// Add composite index for user queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ai_enhanced_notes_user_status 
		ON ai_enhanced_notes(note_id, processing_status);
	`).Error; err != nil {
		log.Printf("Failed to add composite index on ai_enhanced_notes: %v", err)
	}

	// Add indexes for ai_agents table (agent chains use agent_type = 'agent_chain')
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ai_agents_agent_type 
		ON ai_agents(agent_type);
	`).Error; err != nil {
		log.Printf("Failed to add index on ai_agents.agent_type: %v", err)
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ai_agents_user_status 
		ON ai_agents(user_id, status);
	`).Error; err != nil {
		log.Printf("Failed to add index on ai_agents user_status: %v", err)
	}

	log.Println("Manual migrations completed")
	return nil
}

// SetupSingleUser creates or updates the single user from environment variables
func SetupSingleUser(db *gorm.DB, cfg config.Config) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.UserPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Check if any user exists
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount == 0 {
		// Create new user
		user := models.User{
			Username:     cfg.UserUsername,
			Email:        cfg.UserEmail,
			PasswordHash: string(hashedPassword),
		}

		if err := db.Create(&user).Error; err != nil {
			return err
		}

		log.Printf("Created single user with email: %s", cfg.UserEmail)
	} else {
		// Update existing user (assuming there's only one)
		var user models.User
		if err := db.First(&user).Error; err != nil {
			return err
		}

		// Update user details
		user.Username = cfg.UserUsername
		user.Email = cfg.UserEmail
		user.PasswordHash = string(hashedPassword)

		if err := db.Save(&user).Error; err != nil {
			return err
		}

		log.Printf("Updated single user with email: %s", cfg.UserEmail)
	}

	return nil
}
