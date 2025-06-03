package routes

import (
	"errors"
	"net/http"

	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterNotebookRoutes(group *gin.RouterGroup, db *database.Database, notebookService services.NotebookServiceInterface) {
	// Collection endpoints with query parameters
	group.GET("/notebooks", func(c *gin.Context) { GetNotebooks(c, db, notebookService) })
	group.POST("/notebooks", func(c *gin.Context) { CreateNotebook(c, db, notebookService) })

	// Resource-specific endpoints
	group.GET("/notebooks/:id", func(c *gin.Context) { GetNotebookById(c, db, notebookService) })
	group.PUT("/notebooks/:id", func(c *gin.Context) { UpdateNotebook(c, db, notebookService) })
	group.DELETE("/notebooks/:id", func(c *gin.Context) { DeleteNotebook(c, db, notebookService) })
}

func GetNotebooks(c *gin.Context, db *database.Database, notebookService services.NotebookServiceInterface) {
	// Extract query parameters
	params := make(map[string]interface{})

	// Get user ID from context (added by AuthMiddleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		// For single-user systems, use the first user in the database
		userIDInterface = getSingleUserID(db)
	}
	// Convert user ID to string and add to params
	params["user_id"] = userIDInterface.(uuid.UUID).String()

	// Extract other query parameters
	if name := c.Query("name"); name != "" {
		params["name"] = name
	}

	notebooks, err := notebookService.GetNotebooks(db, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notebooks)
}

func CreateNotebook(c *gin.Context, db *database.Database, notebookService services.NotebookServiceInterface) {
	var notebookData map[string]interface{}
	if err := c.ShouldBindJSON(&notebookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add user ID from context to notebook data
	userIDInterface, exists := c.Get("userID")
	if !exists {
		// For single-user systems, use the first user in the database
		// This is more flexible than requiring a specific UUID
		userIDInterface = getSingleUserID(db)
	}
	notebookData["user_id"] = userIDInterface.(uuid.UUID).String()

	notebook, err := notebookService.CreateNotebook(db, notebookData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, notebook)
}

func GetNotebookById(c *gin.Context, db *database.Database, notebookService services.NotebookServiceInterface) {
	id := c.Param("id")

	// Create params map for permissions check
	params := make(map[string]interface{})

	// Add user ID from context to params
	userIDInterface, exists := c.Get("userID")
	if !exists {
		// For single-user systems, use the first user in the database
		userIDInterface = getSingleUserID(db)
	}
	params["user_id"] = userIDInterface.(uuid.UUID).String()

	notebook, err := notebookService.GetNotebookById(db, id, params)
	if err != nil {
		if errors.Is(err, services.ErrNotebookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notebook)
}

func UpdateNotebook(c *gin.Context, db *database.Database, notebookService services.NotebookServiceInterface) {
	id := c.Param("id")
	var notebookData map[string]interface{}
	if err := c.ShouldBindJSON(&notebookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create params map for permissions check
	params := make(map[string]interface{})

	// Add user ID from context to params
	userIDInterface, exists := c.Get("userID")
	if !exists {
		// For single-user systems, use the first user in the database
		userIDInterface = getSingleUserID(db)
	}
	params["user_id"] = userIDInterface.(uuid.UUID).String()

	notebook, err := notebookService.UpdateNotebook(db, id, notebookData, params)
	if err != nil {
		if errors.Is(err, services.ErrNotebookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notebook)
}

func DeleteNotebook(c *gin.Context, db *database.Database, notebookService services.NotebookServiceInterface) {
	id := c.Param("id")

	// Create params map for permissions check
	params := make(map[string]interface{})

	// Add user ID from context to params
	userIDInterface, exists := c.Get("userID")
	if !exists {
		// For single-user systems, use the first user in the database
		userIDInterface = getSingleUserID(db)
	}
	params["user_id"] = userIDInterface.(uuid.UUID).String()

	if err := notebookService.DeleteNotebook(db, id, params); err != nil {
		if errors.Is(err, services.ErrNotebookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}
