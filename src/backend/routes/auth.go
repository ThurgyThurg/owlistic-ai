package routes

import (
	"net/http"

	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/models"
	"owlistic-notes/owlistic/services"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(group *gin.RouterGroup, db *database.Database, authService services.AuthServiceInterface) {
	group.POST("/login", func(c *gin.Context) { Login(c, db, authService) })
}

func Login(c *gin.Context, db *database.Database, authService services.AuthServiceInterface) {
	var loginInput models.UserLoginInput
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := authService.Login(db, loginInput.Email, loginInput.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get user information for complete response
	var user models.User
	if err := db.DB.Where("email = ?", loginInput.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information"})
		return
	}

	// Return complete login response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"userId":  user.ID.String(),
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"profile_pic":  user.ProfilePic,
		},
	})
}
