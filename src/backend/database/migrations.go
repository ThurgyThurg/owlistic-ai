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
		&models.AIProject{},
		&models.AITaskEnhancement{},
		&models.ChatMemory{},
		// Calendar models
		&models.GoogleCalendarCredentials{},
		&models.CalendarEvent{},
		&models.CalendarSync{},
		&models.CalendarReminder{},
		&models.CalendarAttendee{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

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
