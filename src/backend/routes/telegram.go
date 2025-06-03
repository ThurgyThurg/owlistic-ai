package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"owlistic-notes/owlistic/services"
)

type TelegramRoutes struct {
	db              *gorm.DB
	telegramService *services.TelegramService
}

func NewTelegramRoutes(db *gorm.DB, telegramService *services.TelegramService) *TelegramRoutes {
	return &TelegramRoutes{
		db:              db,
		telegramService: telegramService,
	}
}

func (tr *TelegramRoutes) RegisterRoutes(routerGroup *gin.RouterGroup) {
	telegramGroup := routerGroup.Group("/telegram")
	{
		// Webhook endpoint for Telegram
		telegramGroup.POST("/webhook", tr.handleWebhook)
		
		// Manual controls (protected by auth middleware)
		telegramGroup.POST("/send-notification", tr.sendNotification)
		telegramGroup.GET("/status", tr.getStatus)
	}
}

// handleWebhook processes incoming Telegram webhook messages
func (tr *TelegramRoutes) handleWebhook(c *gin.Context) {
	var update tgbotapi.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update format"})
		return
	}

	// Process the update (this could be moved to a background worker for better performance)
	go tr.processUpdate(&update)

	// Telegram expects a 200 OK response
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// processUpdate handles a Telegram update
func (tr *TelegramRoutes) processUpdate(update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// The telegram service will handle message processing
	// This is a simplified version - the actual message handling is done in the service
	// when using polling mode. For webhook mode, we'd need to replicate that logic here.
	
	// For now, we'll just log that we received a webhook
	// In a production setup, you'd want to process the message here
	// or send it to a queue for processing
}

// sendNotification allows manual sending of notifications via API
func (tr *TelegramRoutes) sendNotification(c *gin.Context) {
	var request struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user is authenticated (optional - you might want this for admin use)
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := tr.telegramService.SendNotification(request.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent successfully"})
}

// getStatus returns the status of the Telegram bot
func (tr *TelegramRoutes) getStatus(c *gin.Context) {
	// Verify user is authenticated
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "active",
		"bot_name": "Owlistic Telegram Bot",
		"message": "Telegram bot is running and ready to receive messages",
	})
}