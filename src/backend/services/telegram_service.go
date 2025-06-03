package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"owlistic-notes/owlistic/models"
)

type TelegramService struct {
	db              *gorm.DB
	bot             *tgbotapi.BotAPI
	aiService       *AIService
	calendarService *CalendarService
	allowedChatID   int64
}

type MessageIntent struct {
	Type        string                 `json:"type"`        // "calendar", "task", "project", "note"
	Confidence  float64                `json:"confidence"`  // 0.0 to 1.0
	ExtractedData map[string]interface{} `json:"extracted_data"`
	Reasoning   string                 `json:"reasoning"`
}

type CalendarEvent struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DateTime    time.Time `json:"date_time"`
	Duration    int       `json:"duration"` // minutes
}

func NewTelegramService(db *gorm.DB, aiService *AIService) (*TelegramService, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID environment variable not set")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TELEGRAM_CHAT_ID: %w", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	// Initialize calendar service (optional - will be nil if not configured)
	calendarService, err := NewCalendarService(db)
	if err != nil {
		log.Printf("Calendar service not available: %v", err)
		calendarService = nil
	}

	log.Printf("Telegram bot authorized on account %s", bot.Self.UserName)

	return &TelegramService{
		db:              db,
		bot:             bot,
		aiService:       aiService,
		calendarService: calendarService,
		allowedChatID:   chatID,
	}, nil
}

// StartListening starts the Telegram bot polling loop
func (ts *TelegramService) StartListening() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := ts.bot.GetUpdatesChan(u)

	log.Printf("Telegram bot listening for messages...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Check if message is from allowed chat
		if update.Message.Chat.ID != ts.allowedChatID {
			log.Printf("Ignoring message from unauthorized chat: %d", update.Message.Chat.ID)
			continue
		}

		go ts.handleMessage(update.Message)
	}

	return nil
}

// handleMessage processes incoming Telegram messages
func (ts *TelegramService) handleMessage(message *tgbotapi.Message) {
	ctx := context.Background()
	
	// Get the default user (you might want to implement user mapping)
	userID, err := ts.getDefaultUserID(ctx)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		ts.sendMessage("Sorry, I couldn't identify your user account. Please contact an administrator.")
		return
	}

	// Classify the message intent using AI
	intent, err := ts.classifyMessage(ctx, message.Text)
	if err != nil {
		log.Printf("Failed to classify message: %v", err)
		ts.sendMessage("Sorry, I had trouble understanding your message. Please try again.")
		return
	}

	// Handle the message based on its intent
	response := ts.handleMessageByIntent(ctx, userID, message.Text, intent)
	ts.sendMessage(response)
}

// classifyMessage uses AI to determine the intent of a message
func (ts *TelegramService) classifyMessage(ctx context.Context, messageText string) (*MessageIntent, error) {
	prompt := fmt.Sprintf(`Analyze this message and determine the user's intent. Classify it as one of these types:

1. "calendar" - Adding an event, meeting, appointment, or time-based activity
2. "task" - Creating a to-do item, reminder, or action item
3. "project" - Starting a complex goal that needs to be broken down into steps
4. "note" - General information, thoughts, or miscellaneous content

Message: "%s"

Return a JSON response with this exact structure:
{
  "type": "calendar|task|project|note",
  "confidence": 0.95,
  "extracted_data": {
    "title": "extracted title or description",
    "details": "additional context or details",
    "date_time": "extracted date/time if applicable (ISO format)",
    "duration": "extracted duration in minutes if applicable"
  },
  "reasoning": "Brief explanation of why this classification was chosen"
}

Focus on keywords like:
- Calendar: "meeting", "appointment", "at 3pm", "tomorrow", "schedule", "call", dates/times
- Task: "need to", "remember to", "todo", "buy", "call", "email", "fix", action verbs
- Project: "want to build", "learning", "create", "implement", "develop", complex goals
- Note: general thoughts, ideas, information without clear action

Be confident in your classification. If unsure between task and calendar, prefer task.`, messageText)

	response, err := ts.aiService.callAnthropic(ctx, prompt, 500)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}

	var intent MessageIntent
	if err := json.Unmarshal([]byte(response), &intent); err != nil {
		// Fallback classification if JSON parsing fails
		return ts.fallbackClassification(messageText), nil
	}

	return &intent, nil
}

// fallbackClassification provides simple rule-based classification as backup
func (ts *TelegramService) fallbackClassification(text string) *MessageIntent {
	text = strings.ToLower(text)
	
	// Calendar keywords
	calendarKeywords := []string{"meeting", "appointment", "at ", "pm", "am", "tomorrow", "today", "schedule", "call at", "zoom"}
	// Task keywords  
	taskKeywords := []string{"need to", "remember to", "todo", "buy", "call", "email", "fix", "remind me"}
	// Project keywords
	projectKeywords := []string{"want to build", "learning", "create", "implement", "develop", "project", "goal"}

	calendarScore := 0
	taskScore := 0
	projectScore := 0

	for _, keyword := range calendarKeywords {
		if strings.Contains(text, keyword) {
			calendarScore++
		}
	}
	
	for _, keyword := range taskKeywords {
		if strings.Contains(text, keyword) {
			taskScore++
		}
	}
	
	for _, keyword := range projectKeywords {
		if strings.Contains(text, keyword) {
			projectScore++
		}
	}

	intent := &MessageIntent{
		ExtractedData: map[string]interface{}{
			"title": text,
		},
		Confidence: 0.7,
	}

	if calendarScore > taskScore && calendarScore > projectScore {
		intent.Type = "calendar"
		intent.Reasoning = "Contains calendar-related keywords"
	} else if taskScore > projectScore {
		intent.Type = "task"
		intent.Reasoning = "Contains task-related keywords"
	} else if projectScore > 0 {
		intent.Type = "project"
		intent.Reasoning = "Contains project-related keywords"
	} else {
		intent.Type = "note"
		intent.Reasoning = "No specific action keywords detected"
	}

	return intent
}

// handleMessageByIntent processes the message based on its classified intent
func (ts *TelegramService) handleMessageByIntent(ctx context.Context, userID uuid.UUID, messageText string, intent *MessageIntent) string {
	switch intent.Type {
	case "calendar":
		return ts.handleCalendarEvent(ctx, userID, messageText, intent)
	case "task":
		return ts.handleTask(ctx, userID, messageText, intent)
	case "project":
		return ts.handleProject(ctx, userID, messageText, intent)
	case "note":
		return ts.handleNote(ctx, userID, messageText, intent)
	default:
		return ts.handleNote(ctx, userID, messageText, intent) // Default to note
	}
}

// handleCalendarEvent creates a calendar event
func (ts *TelegramService) handleCalendarEvent(ctx context.Context, userID uuid.UUID, messageText string, intent *MessageIntent) string {
	// Check if calendar service is available and user has calendar access
	if ts.calendarService == nil {
		return ts.handleCalendarEventFallback(ctx, userID, messageText, intent)
	}

	if !ts.calendarService.HasCalendarAccess(ctx, userID) {
		return "📅 Calendar event detected, but you haven't connected your Google Calendar yet.\n\n" +
			"Use `/api/v1/calendar/oauth/authorize` to connect your calendar, then try again.\n\n" +
			"For now, I'll save this as a task:\n\n" + ts.handleCalendarEventFallback(ctx, userID, messageText, intent)
	}

	// Extract calendar event details from AI
	title := messageText
	if extractedTitle, ok := intent.ExtractedData["title"].(string); ok && extractedTitle != "" {
		title = extractedTitle
	}

	// Parse date/time from extracted data or use AI to extract it
	startTime, endTime, allDay := ts.parseEventDateTime(intent.ExtractedData, messageText)

	// Create calendar event request
	request := CalendarEventRequest{
		Title:       title,
		Description: fmt.Sprintf("Created from Telegram: %s", messageText),
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      allDay,
		CalendarID:  "primary", // Use primary calendar
	}

	// Create the calendar event
	event, err := ts.calendarService.CreateEvent(ctx, userID, request)
	if err != nil {
		log.Printf("Failed to create calendar event: %v", err)
		return "❌ Sorry, I couldn't create your calendar event. " + err.Error() + "\n\n" +
			"Falling back to task creation:\n\n" + ts.handleCalendarEventFallback(ctx, userID, messageText, intent)
	}

	// Format response
	timeStr := ""
	if allDay {
		timeStr = fmt.Sprintf("📅 %s", startTime.Format("January 2, 2006"))
	} else {
		timeStr = fmt.Sprintf("📅 %s at %s", startTime.Format("January 2, 2006"), startTime.Format("3:04 PM"))
	}

	return fmt.Sprintf("📅 Calendar event created successfully!\n\n"+
		"**%s**\n%s\n\n"+
		"✅ Added to your Google Calendar\n"+
		"🔗 Event ID: %s", event.Title, timeStr, event.ID)
}

// handleCalendarEventFallback creates a task when calendar integration isn't available
func (ts *TelegramService) handleCalendarEventFallback(ctx context.Context, userID uuid.UUID, messageText string, intent *MessageIntent) string {
	title := messageText
	if extractedTitle, ok := intent.ExtractedData["title"].(string); ok && extractedTitle != "" {
		title = extractedTitle
	}

	// Get or create a default notebook for Telegram messages
	notebook, err := ts.getOrCreateTelegramNotebook(ctx, userID)
	if err != nil {
		log.Printf("Failed to get/create Telegram notebook: %v", err)
		return "❌ Sorry, I couldn't create your calendar event. Please try again."
	}

	// Create a note for the calendar event
	note := models.Note{
		UserID:     userID,
		NotebookID: notebook.ID,
		Title:      fmt.Sprintf("📅 Calendar: %s", title),
		Tags:       pq.StringArray{"telegram", "calendar", "event"},
	}

	if err := ts.db.WithContext(ctx).Create(&note).Error; err != nil {
		log.Printf("Failed to create calendar note: %v", err)
		return "❌ Sorry, I couldn't create your calendar event. Please try again."
	}

	task := models.Task{
		UserID:      userID,
		NoteID:      note.ID,
		Title:       title,
		Description: fmt.Sprintf("Calendar event from Telegram: %s", messageText),
		IsCompleted: false,
		Metadata: models.TaskMetadata{
			"source":           "telegram",
			"intent":           "calendar",
			"original_message": messageText,
			"confidence":       intent.Confidence,
			"reasoning":        intent.Reasoning,
			"extracted_data":   intent.ExtractedData,
		},
	}

	if err := ts.db.WithContext(ctx).Create(&task).Error; err != nil {
		log.Printf("Failed to create calendar task: %v", err)
		return "❌ Sorry, I couldn't save your calendar event. Please try again."
	}

	return fmt.Sprintf("📅 Calendar event saved as task: \"%s\"\n📝 Note ID: %s\n\n⚠️ Connect your Google Calendar for full calendar integration!", task.Title, note.ID)
}

// parseEventDateTime extracts and parses date/time information from the AI extracted data
func (ts *TelegramService) parseEventDateTime(extractedData map[string]interface{}, messageText string) (startTime, endTime time.Time, allDay bool) {
	now := time.Now()
	
	// Try to get parsed datetime from extracted data
	if dateTimeStr, ok := extractedData["date_time"].(string); ok && dateTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, dateTimeStr); err == nil {
			startTime = parsed
		}
	}

	// If no specific time was extracted, try to parse from duration
	if startTime.IsZero() {
		// Default to tomorrow at a reasonable time if no specific time
		startTime = time.Date(now.Year(), now.Month(), now.Day()+1, 14, 0, 0, 0, now.Location())
		
		// Check for time indicators in the message
		msg := strings.ToLower(messageText)
		if strings.Contains(msg, "today") {
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location())
		} else if strings.Contains(msg, "tomorrow") {
			startTime = time.Date(now.Year(), now.Month(), now.Day()+1, 14, 0, 0, 0, now.Location())
		}
		
		// Try to extract time from common patterns
		if strings.Contains(msg, "morning") {
			startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 9, 0, 0, 0, startTime.Location())
		} else if strings.Contains(msg, "afternoon") {
			startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 14, 0, 0, 0, startTime.Location())
		} else if strings.Contains(msg, "evening") {
			startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 18, 0, 0, 0, startTime.Location())
		}
	}

	// Try to get duration from extracted data
	duration := 60 // Default 1 hour
	if durationStr, ok := extractedData["duration"].(string); ok && durationStr != "" {
		if parsed, err := strconv.Atoi(durationStr); err == nil {
			duration = parsed
		}
	}

	// Set end time
	endTime = startTime.Add(time.Duration(duration) * time.Minute)

	// Determine if it's all day (no specific time mentioned)
	allDay = !strings.Contains(strings.ToLower(messageText), "at ") && 
		!strings.Contains(strings.ToLower(messageText), "pm") && 
		!strings.Contains(strings.ToLower(messageText), "am") &&
		startTime.Hour() == 14 && startTime.Minute() == 0 // Default time we set

	return startTime, endTime, allDay
}

// handleTask creates a new task
func (ts *TelegramService) handleTask(ctx context.Context, userID uuid.UUID, messageText string, intent *MessageIntent) string {
	title := messageText
	if extractedTitle, ok := intent.ExtractedData["title"].(string); ok && extractedTitle != "" {
		title = extractedTitle
	}

	// Get or create a default notebook for Telegram tasks
	notebook, err := ts.getOrCreateTelegramNotebook(ctx, userID)
	if err != nil {
		log.Printf("Failed to get/create Telegram notebook: %v", err)
		return "❌ Sorry, I couldn't create your task. Please try again."
	}

	// Create a note for the task
	note := models.Note{
		UserID:     userID,
		NotebookID: notebook.ID,
		Title:      fmt.Sprintf("Task: %s", title),
		Tags:       pq.StringArray{"telegram", "task"},
	}

	if err := ts.db.WithContext(ctx).Create(&note).Error; err != nil {
		log.Printf("Failed to create task note: %v", err)
		return "❌ Sorry, I couldn't create your task. Please try again."
	}

	task := models.Task{
		UserID:      userID,
		NoteID:      note.ID,
		Title:       title,
		Description: fmt.Sprintf("Task from Telegram: %s", messageText),
		IsCompleted: false,
		Metadata: models.TaskMetadata{
			"source":           "telegram",
			"intent":           "task",
			"original_message": messageText,
			"confidence":       intent.Confidence,
			"reasoning":        intent.Reasoning,
			"extracted_data":   intent.ExtractedData,
		},
	}

	if err := ts.db.WithContext(ctx).Create(&task).Error; err != nil {
		log.Printf("Failed to create task: %v", err)
		return "❌ Sorry, I couldn't create your task. Please try again."
	}

	return fmt.Sprintf("✅ Task created: \"%s\"\n📝 Note ID: %s", task.Title, note.ID)
}

// handleProject creates an AI project with task breakdown
func (ts *TelegramService) handleProject(ctx context.Context, userID uuid.UUID, messageText string, intent *MessageIntent) string {
	title := messageText
	if extractedTitle, ok := intent.ExtractedData["title"].(string); ok && extractedTitle != "" {
		title = extractedTitle
	}

	// Use AI to break down the project
	breakdown, err := ts.aiService.BreakDownTask(ctx, title, messageText, 8)
	if err != nil {
		log.Printf("Failed to break down project: %v", err)
		return "❌ Sorry, I couldn't break down your project. Please try again."
	}

	// Create the AI project with notebook integration
	project := models.AIProject{
		UserID:      userID,
		Name:        title,
		Description: fmt.Sprintf("Project from Telegram: %s", messageText),
		Status:      "active",
		AITags:      pq.StringArray{"telegram", "project"},
		AIMetadata: models.AIMetadata{
			"source":           "telegram",
			"intent":           "project",
			"original_message": messageText,
			"confidence":       intent.Confidence,
			"reasoning":        intent.Reasoning,
			"breakdown":        breakdown,
		},
	}

	// Create notebook and notes for the project
	notebookID, noteIDs, err := ts.aiService.CreateProjectNotebook(ctx, userID, title, project.Description, breakdown)
	if err != nil {
		log.Printf("Failed to create project notebook: %v", err)
	} else {
		project.NotebookID = notebookID
		project.RelatedNoteIDs = models.UUIDArray(noteIDs)
	}

	if err := ts.db.WithContext(ctx).Create(&project).Error; err != nil {
		log.Printf("Failed to create project: %v", err)
		return "❌ Sorry, I couldn't create your project. Please try again."
	}

	stepsCount := 0
	if steps, ok := breakdown["steps"].([]interface{}); ok {
		stepsCount = len(steps)
	}

	response := fmt.Sprintf("🚀 Project created: \"%s\"\n📊 Broken down into %d steps", project.Name, stepsCount)
	if project.NotebookID != nil {
		response += fmt.Sprintf("\n📓 Notebook ID: %s", *project.NotebookID)
	}
	
	return response
}

// handleNote creates a miscellaneous note
func (ts *TelegramService) handleNote(ctx context.Context, userID uuid.UUID, messageText string, intent *MessageIntent) string {
	title := messageText
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	// Get or create a default notebook for Telegram notes
	notebook, err := ts.getOrCreateTelegramNotebook(ctx, userID)
	if err != nil {
		log.Printf("Failed to get/create Telegram notebook: %v", err)
		return "❌ Sorry, I couldn't create your note. Please try again."
	}

	note := models.Note{
		UserID:     userID,
		NotebookID: notebook.ID,
		Title:      title,
		Tags:       pq.StringArray{"telegram", "note"},
	}

	if err := ts.db.WithContext(ctx).Create(&note).Error; err != nil {
		log.Printf("Failed to create note: %v", err)
		return "❌ Sorry, I couldn't create your note. Please try again."
	}

	// Create a text block with the message content
	block := models.Block{
		UserID:  userID,
		NoteID:  note.ID,
		Type:    "paragraph",
		Order:   1.0,
		Content: map[string]interface{}{
			"text": messageText,
		},
	}

	if err := ts.db.WithContext(ctx).Create(&block).Error; err != nil {
		log.Printf("Failed to create note block: %v", err)
		// Note was created, so don't fail completely
	}

	// Trigger AI processing for the note
	go func() {
		if err := ts.aiService.ProcessNoteWithAI(context.Background(), note.ID); err != nil {
			log.Printf("Failed to process note with AI: %v", err)
		}
	}()

	return fmt.Sprintf("📝 Note created: \"%s\"\n🤖 AI processing started for enhanced insights\n📝 Note ID: %s", note.Title, note.ID)
}

// getOrCreateTelegramNotebook gets or creates a default notebook for Telegram messages
func (ts *TelegramService) getOrCreateTelegramNotebook(ctx context.Context, userID uuid.UUID) (*models.Notebook, error) {
	var notebook models.Notebook
	
	// Try to find existing Telegram notebook
	err := ts.db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, "📱 Telegram Messages").First(&notebook).Error
	if err == nil {
		return &notebook, nil
	}

	// Create new Telegram notebook
	notebook = models.Notebook{
		UserID:      userID,
		Name:        "📱 Telegram Messages",
		Description: "Notes, tasks, and projects created via Telegram bot",
	}

	if err := ts.db.WithContext(ctx).Create(&notebook).Error; err != nil {
		return nil, fmt.Errorf("failed to create Telegram notebook: %w", err)
	}

	return &notebook, nil
}

// getDefaultUserID gets the first user ID (you might want to implement proper user mapping)
func (ts *TelegramService) getDefaultUserID(ctx context.Context) (uuid.UUID, error) {
	var user models.User
	if err := ts.db.WithContext(ctx).First(&user).Error; err != nil {
		return uuid.Nil, fmt.Errorf("no user found: %w", err)
	}
	return user.ID, nil
}

// sendMessage sends a message to the configured Telegram chat
func (ts *TelegramService) sendMessage(text string) {
	msg := tgbotapi.NewMessage(ts.allowedChatID, text)
	msg.ParseMode = "Markdown"
	
	if _, err := ts.bot.Send(msg); err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
	}
}

// SendNotification sends a notification to Telegram (can be used by other services)
func (ts *TelegramService) SendNotification(message string) error {
	msg := tgbotapi.NewMessage(ts.allowedChatID, message)
	msg.ParseMode = "Markdown"
	
	_, err := ts.bot.Send(msg)
	return err
}