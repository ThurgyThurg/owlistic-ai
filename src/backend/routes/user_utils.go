package routes

import (
	"log"

	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
)

// getSingleUserID returns the first user ID for single-user systems
func getSingleUserID(db *database.Database) uuid.UUID {
	var user models.User
	if err := db.DB.First(&user).Error; err != nil {
		log.Printf("Warning: No users found in database for single-user mode")
		// Return the intended single-user UUID as fallback
		singleUserUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
		return singleUserUUID
	}
	return user.ID
}