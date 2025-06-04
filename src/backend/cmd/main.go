package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"owlistic-notes/owlistic/broker"
	"owlistic-notes/owlistic/config"
	"owlistic-notes/owlistic/database"
	"owlistic-notes/owlistic/middleware"
	"owlistic-notes/owlistic/models"
	"owlistic-notes/owlistic/routes"
	"owlistic-notes/owlistic/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	db, err := database.Setup(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize producer
	err = broker.InitProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize producer: %v", err)
	}
	defer broker.CloseProducer()

	// Initialize all service instances properly with database
	// Initialize authentication service
	authService := services.NewAuthService(cfg.JWTSecret, cfg.JWTExpirationHours)
	services.AuthServiceInstance = authService

	// Initialize user service with auth service dependency
	userService := services.NewUserService(authService)
	services.UserServiceInstance = userService

	// Properly initialize other service instances with the database
	services.NoteServiceInstance = services.NewNoteService()
	services.NotebookServiceInstance = services.NewNotebookService()
	services.BlockServiceInstance = services.NewBlockService()
	services.TaskServiceInstance = services.NewTaskService()
	services.TrashServiceInstance = services.NewTrashService()

	// Initialize single user for single-user mode
	if err := initializeSingleUser(db, cfg); err != nil {
		log.Printf("Warning: Failed to initialize single user: %v", err)
	}

	// Initialize eventHandler service with the database
	eventHandlerService := services.NewEventHandlerService(db)
	services.EventHandlerServiceInstance = eventHandlerService

	webSocketService := services.NewWebSocketService(db)
	webSocketService.SetJWTSecret([]byte(cfg.JWTSecret))
	services.WebSocketServiceInstance = webSocketService

	// Initialize BlockTaskSyncHandler service with the database
	syncHandler := services.NewSyncHandlerService(db)

	// Start event-based services
	log.Println("Starting event handler service...")
	eventHandlerService.Start()
	defer eventHandlerService.Stop()

	log.Println("Starting WebSocket service...")
	webSocketService.Start(cfg)
	defer webSocketService.Stop()

	// Start block-task sync handler
	log.Println("Starting Block-Task Sync Handler...")
	syncHandler.Start(cfg)
	defer syncHandler.Stop()

	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORSMiddleware(cfg.AppOrigins))

	// Create public API groups
	publicGroup := router.Group("/api/v1")

	// Register public routes (no auth required)
	routes.RegisterAuthRoutes(publicGroup, db, authService)
	routes.RegisterPublicUserRoutes(publicGroup, db, userService, authService)
	
	// Register AI routes on public group for single-user mode
	aiRoutes := routes.NewAIRoutes(db.DB)
	aiRoutes.RegisterRoutes(publicGroup)
	
	// Register Agent Orchestrator routes on public group for single-user mode
	orchestratorRoutes := routes.NewAgentOrchestratorRoutes(db.DB)
	orchestratorRoutes.RegisterRoutes(publicGroup)

	// Register core routes on public group for single-user mode
	routes.RegisterNoteRoutes(publicGroup, db, services.NoteServiceInstance)
	routes.RegisterTaskRoutes(publicGroup, db, services.TaskServiceInstance)
	routes.RegisterNotebookRoutes(publicGroup, db, services.NotebookServiceInstance)
	routes.RegisterBlockRoutes(publicGroup, db, services.BlockServiceInstance)
	routes.RegisterTrashRoutes(publicGroup, db, services.TrashServiceInstance)

	// Create protected API group with auth middleware (for future multi-user features)
	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(middleware.AuthMiddleware(authService))

	// Register remaining protected API routes
	routes.RegisterProtectedUserRoutes(protectedGroup, db, userService, authService)
	routes.RegisterRoleRoutes(protectedGroup, db, services.RoleServiceInstance)

	// Register WebSocket routes with consistent auth middleware
	wsGroup := router.Group("/ws")
	wsGroup.Use(middleware.AuthMiddleware(authService))
	routes.RegisterWebSocketRoutes(wsGroup, webSocketService)


	// Register Calendar routes on protected group
	calendarRoutes, err := routes.NewCalendarRoutes(db.DB)
	if err != nil {
		log.Printf("Failed to initialize calendar routes: %v", err)
		log.Printf("Calendar functionality will not be available")
	} else {
		calendarRoutes.RegisterRoutes(protectedGroup)
		calendarRoutes.RegisterPublicRoutes(publicGroup)
		log.Println("Calendar routes registered successfully")
	}

	// Register Zettelkasten routes on public group for single-user mode
	zettelRoutes, err := routes.NewZettelkastenRoutes(db.DB)
	if err != nil {
		log.Printf("Failed to initialize Zettelkasten routes: %v", err)
		log.Printf("Zettelkasten functionality will not be available")
	} else {
		zettelRoutes.RegisterRoutes(publicGroup)
		log.Println("Zettelkasten routes registered successfully")
	}

	// Initialize Telegram service and routes
	aiService := services.NewAIService(db.DB)
	telegramService, err := services.NewTelegramService(db.DB, aiService)
	if err != nil {
		log.Printf("Failed to initialize Telegram service: %v", err)
		log.Printf("Telegram bot will not be available")
	} else {
		// Register Telegram routes
		telegramRoutes := routes.NewTelegramRoutes(db.DB, telegramService)
		telegramRoutes.RegisterRoutes(protectedGroup)

		// Start Telegram bot listening in background
		go func() {
			if err := telegramService.StartListening(); err != nil {
				log.Printf("Telegram bot stopped: %v", err)
			}
		}()
		log.Println("Telegram bot started and listening for messages...")
	}

	// Register debug routes for monitoring events
	routes.SetupDebugRoutes(router, db)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		os.Exit(0)
	}()

	log.Printf("API server is running on http://0.0.0.0:%v", cfg.AppPort)
	log.Printf("Access from other devices: http://YOUR_COMPUTER_IP:%v", cfg.AppPort)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", cfg.AppPort), router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initializeSingleUser creates the single user if it doesn't exist
func initializeSingleUser(db *database.Database, cfg config.Config) error {
	singleUserUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000001")
	if err != nil {
		return fmt.Errorf("invalid single user UUID: %w", err)
	}

	// Check if user already exists
	var existingUser models.User
	err = db.DB.Where("id = ?", singleUserUUID).First(&existingUser).Error
	if err == nil {
		// User already exists
		log.Printf("Single user already exists: %s", existingUser.Email)
		return nil
	}

	// User doesn't exist, create it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.UserPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		ID:           singleUserUUID,
		Username:     cfg.UserUsername,
		Email:        cfg.UserEmail,
		PasswordHash: string(hashedPassword),
		DisplayName:  cfg.UserUsername,
		Preferences:  make(map[string]interface{}),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create single user: %w", err)
	}

	log.Printf("Single user created successfully: %s (%s)", user.Email, user.Username)
	return nil
}
