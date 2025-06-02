package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIService struct {
	db             *gorm.DB
	anthropicKey   string
	anthropicModel string
	chromaBaseURL  string
	httpClient     *http.Client
}

type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type ChromaEmbeddingRequest struct {
	Input []string `json:"input"`
}

type ChromaEmbeddingResponse struct {
	Data [][]float64 `json:"data"`
}

func NewAIService(db *gorm.DB) *AIService {
	// Set default models if not specified
	anthropicModel := os.Getenv("ANTHROPIC_MODEL")
	if anthropicModel == "" {
		anthropicModel = "claude-3-5-sonnet-20241022" // Default Anthropic model
	}

	// Clean the API key of any whitespace
	anthropicKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))

	return &AIService{
		db:             db,
		anthropicKey:   anthropicKey,
		anthropicModel: anthropicModel,
		chromaBaseURL:  os.Getenv("CHROMA_BASE_URL"),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// ProcessNoteWithAI enhances a note with AI-generated metadata
func (ai *AIService) ProcessNoteWithAI(ctx context.Context, noteID uuid.UUID) error {
	// Get the note
	var note models.Note
	if err := ai.db.WithContext(ctx).First(&note, noteID).Error; err != nil {
		return fmt.Errorf("failed to find note: %w", err)
	}

	// Get note content (combine title and blocks content)
	content := ai.extractNoteContent(&note)

	// Generate AI enhancements concurrently
	titleChan := make(chan string, 1)
	summaryChan := make(chan string, 1)
	tagsChan := make(chan []string, 1)
	embeddingChan := make(chan []float64, 1)
	actionStepsChan := make(chan []string, 1)
	learningItemsChan := make(chan []string, 1)
	errChan := make(chan error, 6)

	// Generate title if empty
	go func() {
		if note.Title == "" {
			title, err := ai.generateTitle(ctx, content)
			if err != nil {
				errChan <- err
				return
			}
			titleChan <- title
		} else {
			titleChan <- note.Title
		}
	}()

	// Generate summary
	go func() {
		summary, err := ai.generateSummary(ctx, content)
		if err != nil {
			errChan <- err
			return
		}
		summaryChan <- summary
	}()

	// Extract tags
	go func() {
		tags, err := ai.extractTags(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		tagsChan <- tags
	}()

	// Generate embeddings
	go func() {
		embedding, err := ai.createEmbeddingWithChroma(fmt.Sprintf("%s\n\n%s", note.Title, content))
		if err != nil {
			errChan <- err
			return
		}
		embeddingChan <- embedding
	}()

	// Generate actionable steps
	go func() {
		steps, err := ai.extractActionableSteps(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		actionStepsChan <- steps
	}()

	// Generate learning items
	go func() {
		items, err := ai.extractLearningItems(ctx, content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		learningItemsChan <- items
	}()

	// Collect results
	var finalTitle string
	var summary string
	var aiTags []string
	var embeddings []float64
	var actionSteps []string
	var learningItems []string
	
	completed := 0
	errors := 0

	for completed < 6 {
		select {
		case title := <-titleChan:
			finalTitle = title
			completed++
		case s := <-summaryChan:
			summary = s
			completed++
		case tags := <-tagsChan:
			aiTags = tags
			completed++
		case emb := <-embeddingChan:
			embeddings = emb
			completed++
		case steps := <-actionStepsChan:
			actionSteps = steps
			completed++
		case items := <-learningItemsChan:
			learningItems = items
			completed++
		case err := <-errChan:
			log.Printf("AI processing error: %v", err)
			errors++
			completed++ // Count error as completed to avoid infinite loop
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	// If all operations failed, return error
	if errors >= 6 {
		return fmt.Errorf("all AI processing operations failed")
	}

	// Update or create AI-enhanced note record
	aiNote := models.AIEnhancedNote{
		Note:             note,
		Summary:          summary,
		AITags:           aiTags,
		ActionSteps:      actionSteps,
		LearningItems:    learningItems,
		Embeddings:       embeddings,
		ProcessingStatus: "completed",
		LastProcessedAt:  &[]time.Time{time.Now()}[0],
	}

	// Update original note title if it was generated
	if note.Title == "" && finalTitle != "" {
		if err := ai.db.WithContext(ctx).Model(&note).Update("title", finalTitle).Error; err != nil {
			log.Printf("Failed to update note title: %v", err)
		}
	}

	// Find related notes using embeddings
	if len(embeddings) > 0 {
		relatedNotes, err := ai.findRelatedNotes(ctx, embeddings, noteID)
		if err != nil {
			log.Printf("Failed to find related notes: %v", err)
		} else {
			aiNote.RelatedNoteIDs = relatedNotes
		}
	}

	// Add action steps and learning items as blocks to the note
	if err := ai.addAIInsightsToNote(ctx, noteID, actionSteps, learningItems); err != nil {
		log.Printf("Failed to add AI insights to note: %v", err)
		// Don't fail the whole process if this fails
	}

	// Save AI enhancement
	return ai.db.WithContext(ctx).Save(&aiNote).Error
}

func (ai *AIService) extractNoteContent(note *models.Note) string {
	var content strings.Builder

	// Load blocks
	var blocks []models.Block
	ai.db.Where("note_id = ?", note.ID).Find(&blocks)

	for _, block := range blocks {
		// Extract text content from block based on type
		if textContent := ai.extractBlockText(block); textContent != "" {
			content.WriteString(textContent)
			content.WriteString("\n")
		}
	}

	return content.String()
}

func (ai *AIService) generateTitle(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf(`Generate a concise, descriptive title (max 10 words) for this content:

%s

Title:`, content[:min(500, len(content))])

	response, err := ai.callAnthropic(ctx, prompt, 50)
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(strings.Trim(response, `"`))
	if len(title) > 100 {
		title = title[:100]
	}

	return title, nil
}

func (ai *AIService) generateSummary(ctx context.Context, content string) (string, error) {
	if len(content) < 200 {
		return content, nil
	}

	prompt := fmt.Sprintf(`Provide a concise 2-3 sentence summary of this content:

%s

Summary:`, content)

	return ai.callAnthropic(ctx, prompt, 150)
}

func (ai *AIService) extractTags(ctx context.Context, content, title string) ([]string, error) {
	prompt := fmt.Sprintf(`Extract 3-7 relevant tags from this content. Tags should be:
- Single words or short phrases (2-3 words max)
- Lowercase
- Descriptive and useful for categorization

Title: %s
Content: %s

Return only the tags as a JSON array. Example: ["productivity", "machine learning", "project planning"]

Tags:`, title, content[:min(800, len(content))])

	response, err := ai.callAnthropic(ctx, prompt, 100)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var tags []string
	if err := json.Unmarshal([]byte(response), &tags); err != nil {
		// Fallback: split by commas
		parts := strings.Split(response, ",")
		for _, part := range parts {
			tag := strings.ToLower(strings.Trim(strings.Trim(part, "[]\"'"), " "))
			if tag != "" && len(tag) > 1 {
				tags = append(tags, tag)
			}
		}
	}

	// Limit to 7 tags
	if len(tags) > 7 {
		tags = tags[:7]
	}

	return tags, nil
}

func (ai *AIService) createEmbeddingWithChroma(text string) ([]float64, error) {
	if ai.chromaBaseURL == "" {
		log.Printf("Chroma base URL not configured, using simple embedding fallback")
		return ai.createSimpleEmbedding(text), nil
	}

	log.Printf("Attempting to create embedding for text length: %d", len(text))

	// Temporarily disable Chroma due to API issues, use simple fallback
	log.Printf("Chroma API has compatibility issues, using simple embedding fallback")
	return ai.createSimpleEmbedding(text), nil

	// TODO: Re-enable once Chroma API issues are resolved
	// Ensure the note_embeddings collection exists with proper embedding function
	if err := ai.ensureEmbeddingCollection(); err != nil {
		log.Printf("Failed to ensure embedding collection: %v", err)
		return ai.createSimpleEmbedding(text), nil // Fallback instead of error
	}

	// Create a temporary document to get its embedding
	tempDocID := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	
	// Add document to collection (this will generate the embedding using Chroma's embedding function)
	addPayload := map[string]interface{}{
		"documents": []string{text[:min(8000, len(text))]},
		"ids":       []string{tempDocID},
	}

	addJSON, err := json.Marshal(addPayload)
	if err != nil {
		log.Printf("Failed to marshal add payload: %v", err)
		return nil, err
	}

	addReq, err := http.NewRequest("POST", ai.chromaBaseURL+"/api/v1/collections/note_embeddings/add", bytes.NewBuffer(addJSON))
	if err != nil {
		log.Printf("Failed to create add request: %v", err)
		return nil, err
	}
	addReq.Header.Set("Content-Type", "application/json")

	log.Printf("Adding document to Chroma collection...")
	addResp, err := ai.httpClient.Do(addReq)
	if err != nil {
		log.Printf("Failed to add document to Chroma: %v", err)
		return nil, err
	}
	defer addResp.Body.Close()

	if addResp.StatusCode != 200 && addResp.StatusCode != 201 {
		log.Printf("Chroma add request failed with status: %d", addResp.StatusCode)
		return nil, fmt.Errorf("chroma add request failed with status: %d", addResp.StatusCode)
	}

	// Get the document with its embedding
	getPayload := map[string]interface{}{
		"ids":     []string{tempDocID},
		"include": []string{"embeddings"},
	}

	getJSON, err := json.Marshal(getPayload)
	if err != nil {
		log.Printf("Failed to marshal get payload: %v", err)
		return nil, err
	}

	getReq, err := http.NewRequest("POST", ai.chromaBaseURL+"/api/v1/collections/note_embeddings/get", bytes.NewBuffer(getJSON))
	if err != nil {
		log.Printf("Failed to create get request: %v", err)
		return nil, err
	}
	getReq.Header.Set("Content-Type", "application/json")

	log.Printf("Getting embedding from Chroma...")
	getResp, err := ai.httpClient.Do(getReq)
	if err != nil {
		log.Printf("Failed to get embedding from Chroma: %v", err)
		return nil, err
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != 200 {
		log.Printf("Chroma get request failed with status: %d", getResp.StatusCode)
		return nil, fmt.Errorf("chroma get request failed with status: %d", getResp.StatusCode)
	}

	var getResult struct {
		Embeddings [][]float64 `json:"embeddings"`
	}

	if err := json.NewDecoder(getResp.Body).Decode(&getResult); err != nil {
		log.Printf("Failed to decode Chroma response: %v", err)
		return nil, err
	}

	// Clean up the temporary document
	deletePayload := map[string]interface{}{
		"ids": []string{tempDocID},
	}
	deleteJSON, _ := json.Marshal(deletePayload)
	deleteReq, err := http.NewRequest("POST", ai.chromaBaseURL+"/api/v1/collections/note_embeddings/delete", bytes.NewBuffer(deleteJSON))
	if err == nil {
		deleteReq.Header.Set("Content-Type", "application/json")
		ai.httpClient.Do(deleteReq)
	}

	if len(getResult.Embeddings) == 0 || len(getResult.Embeddings[0]) == 0 {
		log.Printf("No embedding data returned from Chroma")
		return nil, fmt.Errorf("no embedding data returned from Chroma")
	}

	log.Printf("Successfully generated embedding with dimension: %d", len(getResult.Embeddings[0]))
	return getResult.Embeddings[0], nil
}

// createSimpleEmbedding creates a simple deterministic embedding as fallback
func (ai *AIService) createSimpleEmbedding(text string) []float64 {
	textBytes := []byte(text[:min(1000, len(text))])
	embedding := make([]float64, 384) // Standard sentence-transformer dimension
	
	// Create a simple hash-based embedding (deterministic)
	for i := 0; i < len(embedding); i++ {
		sum := 0
		for j, b := range textBytes {
			sum += int(b) * (i + j + 1)
		}
		embedding[i] = float64(sum%1000-500) / 500.0 // Normalize to [-1, 1]
	}
	
	log.Printf("Generated simple embedding with dimension: %d", len(embedding))
	return embedding
}

// ensureEmbeddingCollection creates the note_embeddings collection if it doesn't exist
func (ai *AIService) ensureEmbeddingCollection() error {
	log.Printf("Checking if note_embeddings collection exists...")
	
	// First, let's try to list all collections to see what endpoint works
	listReq, err := http.NewRequest("GET", ai.chromaBaseURL+"/api/v1/collections", nil)
	if err != nil {
		log.Printf("Failed to create collections list request: %v", err)
		return err
	}

	listResp, err := ai.httpClient.Do(listReq)
	if err != nil {
		log.Printf("Failed to list collections: %v", err)
		return err
	}
	defer listResp.Body.Close()
	
	log.Printf("Collections list response status: %d", listResp.StatusCode)

	// Try to get the specific collection
	getReq, err := http.NewRequest("GET", ai.chromaBaseURL+"/api/v1/collections/note_embeddings", nil)
	if err != nil {
		log.Printf("Failed to create collection check request: %v", err)
		return err
	}

	resp, err := ai.httpClient.Do(getReq)
	if err != nil {
		log.Printf("Failed to check collection existence: %v", err)
		return err
	}
	resp.Body.Close()

	// If collection exists (status 200), return success
	if resp.StatusCode == 200 {
		log.Printf("Collection note_embeddings already exists")
		return nil
	}

	log.Printf("Collection doesn't exist (status: %d), creating it...", resp.StatusCode)

	// Create collection with default embedding function (all-MiniLM-L6-v2)
	collectionPayload := map[string]interface{}{
		"name": "note_embeddings",
		"metadata": map[string]string{
			"description": "Embeddings for note similarity search",
		},
		// Chroma will use the default all-MiniLM-L6-v2 embedding function
	}

	collectionJSON, err := json.Marshal(collectionPayload)
	if err != nil {
		log.Printf("Failed to marshal collection payload: %v", err)
		return err
	}

	log.Printf("Attempting to create collection with payload: %s", string(collectionJSON))

	collectionReq, err := http.NewRequest("POST", ai.chromaBaseURL+"/api/v1/collections", bytes.NewBuffer(collectionJSON))
	if err != nil {
		log.Printf("Failed to create collection request: %v", err)
		return err
	}
	collectionReq.Header.Set("Content-Type", "application/json")

	createResp, err := ai.httpClient.Do(collectionReq)
	if err != nil {
		log.Printf("Failed to create collection: %v", err)
		return err
	}
	defer createResp.Body.Close()

	log.Printf("Collection creation response status: %d", createResp.StatusCode)
	
	if createResp.StatusCode != 200 && createResp.StatusCode != 201 {
		// Read the response body to see what the error is
		bodyBytes, _ := io.ReadAll(createResp.Body)
		log.Printf("Collection creation error response: %s", string(bodyBytes))
		return fmt.Errorf("collection creation failed with status: %d: %s", createResp.StatusCode, string(bodyBytes))
	}

	log.Printf("Successfully created note_embeddings collection")
	return nil
}

func (ai *AIService) findRelatedNotes(ctx context.Context, embedding []float64, excludeID uuid.UUID) ([]uuid.UUID, error) {
	if ai.chromaBaseURL == "" {
		// Chroma not configured, skip related notes
		return []uuid.UUID{}, nil
	}

	// Store the current note's embedding in Chroma
	if err := ai.storeEmbeddingInChroma(ctx, excludeID.String(), embedding); err != nil {
		log.Printf("Failed to store embedding in Chroma: %v", err)
		// Don't fail the whole process
	}

	// Query for similar embeddings
	relatedIDs, err := ai.queryChromaForSimilar(ctx, embedding, excludeID.String(), 5)
	if err != nil {
		log.Printf("Failed to query Chroma for similar embeddings: %v", err)
		return []uuid.UUID{}, nil
	}

	// Convert string IDs back to UUIDs
	var result []uuid.UUID
	for _, idStr := range relatedIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			result = append(result, id)
		}
	}

	return result, nil
}

func (ai *AIService) callAnthropic(ctx context.Context, prompt string, maxTokens int) (string, error) {
	log.Printf("Making Anthropic API call with model: %s, maxTokens: %d", ai.anthropicModel, maxTokens)
	
	if ai.anthropicKey == "" {
		log.Printf("ERROR: Anthropic API key is empty")
		return "", fmt.Errorf("anthropic API key not configured")
	}
	
	// Debug: Log the API key format (first and last few characters only for security)
	if len(ai.anthropicKey) > 10 {
		log.Printf("API key format: %s...%s (length: %d)", ai.anthropicKey[:8], ai.anthropicKey[len(ai.anthropicKey)-4:], len(ai.anthropicKey))
	}

	req := AnthropicRequest{
		Model:     ai.anthropicModel,
		MaxTokens: maxTokens,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Failed to marshal Anthropic request: %v", err)
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create Anthropic request: %v", err)
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ai.anthropicKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	log.Printf("Sending request to Anthropic API...")
	log.Printf("Request URL: %s", httpReq.URL)
	log.Printf("Request headers: Content-Type=%s, x-api-key=%s..., anthropic-version=%s", 
		httpReq.Header.Get("Content-Type"),
		ai.anthropicKey[:20],
		httpReq.Header.Get("anthropic-version"))
	log.Printf("Request body: %s", string(jsonData)[:min(200, len(jsonData))])
	
	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Anthropic API request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("Anthropic API response status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Anthropic API error response: %s", string(bodyBytes))
		return "", fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		log.Printf("Failed to decode Anthropic response: %v", err)
		return "", err
	}

	if len(anthropicResp.Content) == 0 {
		log.Printf("Anthropic response has no content")
		return "", fmt.Errorf("no content in response")
	}

	log.Printf("Anthropic API call successful, response length: %d", len(anthropicResp.Content[0].Text))
	return anthropicResp.Content[0].Text, nil
}

// extractBlockText extracts readable text from a block's content
func (ai *AIService) extractBlockText(block models.Block) string {
	if block.Content == nil {
		return ""
	}

	// For most block types, look for "text" field in content
	if text, ok := block.Content["text"].(string); ok {
		return text
	}

	// For other content structures, try to extract meaningful text
	if content, ok := block.Content["content"].(string); ok {
		return content
	}

	return ""
}

// SearchNotesBySimilarity performs semantic search using embeddings
func (ai *AIService) SearchNotesBySimilarity(ctx context.Context, query string, limit int) ([]models.AIEnhancedNote, error) {
	// Generate embedding for query
	queryEmbedding, err := ai.createEmbeddingWithChroma(query)
	if err != nil {
		return nil, err
	}

	// This would use vector similarity search with ChromaDB or pgvector
	// For now, return empty slice - needs vector database integration
	// When implemented, you would use queryEmbedding for similarity search
	_ = queryEmbedding // Suppress unused variable warning

	var results []models.AIEnhancedNote
	return results, nil
}

// extractActionableSteps analyzes content and extracts actionable steps
func (ai *AIService) extractActionableSteps(ctx context.Context, content, title string) ([]string, error) {
	prompt := fmt.Sprintf(`Analyze this content and extract 3-7 specific, actionable steps that could be taken based on the information. Each step should be:
- Concrete and actionable (start with a verb)
- Specific enough to be implemented
- Relevant to the content's main topic
- Practical and achievable

Title: %s
Content: %s

Return only the actionable steps as a JSON array. Example: ["Research available frameworks", "Set up development environment", "Create project timeline"]

Actionable Steps:`, title, content[:min(1000, len(content))])

	response, err := ai.callAnthropic(ctx, prompt, 200)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var steps []string
	if err := json.Unmarshal([]byte(response), &steps); err != nil {
		// Fallback: split by lines and clean up
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			step := strings.TrimSpace(strings.Trim(line, "-•*\"'[]"))
			if step != "" && len(step) > 5 {
				steps = append(steps, step)
			}
		}
	}

	// Limit to 7 steps
	if len(steps) > 7 {
		steps = steps[:7]
	}

	return steps, nil
}

// extractLearningItems analyzes content and extracts learning opportunities
func (ai *AIService) extractLearningItems(ctx context.Context, content, title string) ([]string, error) {
	prompt := fmt.Sprintf(`Analyze this content and identify 3-7 specific learning opportunities, concepts, or knowledge gaps that could be explored further. Each item should be:
- A specific topic, concept, or skill to learn
- Relevant to understanding or applying the content better
- Educational and knowledge-building focused
- Distinct from actionable steps (these are for learning, not doing)

Title: %s
Content: %s

Return only the learning items as a JSON array. Example: ["Docker containerization principles", "REST API design patterns", "Database indexing strategies"]

Learning Items:`, title, content[:min(1000, len(content))])

	response, err := ai.callAnthropic(ctx, prompt, 200)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var items []string
	if err := json.Unmarshal([]byte(response), &items); err != nil {
		// Fallback: split by lines and clean up
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			item := strings.TrimSpace(strings.Trim(line, "-•*\"'[]"))
			if item != "" && len(item) > 5 {
				items = append(items, item)
			}
		}
	}

	// Limit to 7 items
	if len(items) > 7 {
		items = items[:7]
	}

	return items, nil
}

// addAIInsightsToNote adds action steps and learning items as blocks to the note
func (ai *AIService) addAIInsightsToNote(ctx context.Context, noteID uuid.UUID, actionSteps, learningItems []string) error {
	// Get the note to retrieve the user ID
	var note models.Note
	if err := ai.db.WithContext(ctx).First(&note, noteID).Error; err != nil {
		return fmt.Errorf("failed to find note: %w", err)
	}
	// Check if AI insights already exist in the note to avoid duplicates
	var existingBlocks []models.Block
	if err := ai.db.WithContext(ctx).Where("note_id = ? AND (content->>'text' LIKE ? OR content->>'text' LIKE ?)", 
		noteID, "%## 🎯 Action Steps%", "%## 💡 Learning Opportunities%").Find(&existingBlocks).Error; err != nil {
		return fmt.Errorf("failed to check for existing AI insights: %w", err)
	}

	// If AI insights already exist, don't add them again
	if len(existingBlocks) > 0 {
		log.Printf("AI insights already exist for note %s, skipping", noteID)
		return nil
	}

	// Get the current highest order number for blocks in this note
	var maxOrder float64
	if err := ai.db.WithContext(ctx).Model(&models.Block{}).
		Where("note_id = ?", noteID).
		Select("COALESCE(MAX(\"order\"), 0)").
		Scan(&maxOrder).Error; err != nil {
		return fmt.Errorf("failed to get max order: %w", err)
	}

	currentOrder := maxOrder + 1

	// Add Action Steps section if we have any
	if len(actionSteps) > 0 {
		// Add header block
		headerBlock := models.Block{
			UserID:  note.UserID,
			NoteID:  noteID,
			Type:    "heading",
			Order:   currentOrder,
			Content: map[string]interface{}{
				"text":  "## 🎯 Action Steps",
				"level": 2,
			},
		}
		if err := ai.db.WithContext(ctx).Create(&headerBlock).Error; err != nil {
			return fmt.Errorf("failed to create action steps header: %w", err)
		}
		currentOrder++

		// Add each action step as a task block
		for i, step := range actionSteps {
			stepBlock := models.Block{
				UserID: note.UserID,
				NoteID: noteID,
				Type:   "task",
				Order:  currentOrder,
				Content: map[string]interface{}{
					"text":      fmt.Sprintf("%d. %s", i+1, step),
					"completed": false,
				},
			}
			if err := ai.db.WithContext(ctx).Create(&stepBlock).Error; err != nil {
				return fmt.Errorf("failed to create action step block: %w", err)
			}
			currentOrder++
		}
	}

	// Add Learning Opportunities section if we have any
	if len(learningItems) > 0 {
		// Add header block
		headerBlock := models.Block{
			UserID:  note.UserID,
			NoteID:  noteID,
			Type:    "heading",
			Order:   currentOrder,
			Content: map[string]interface{}{
				"text":  "## 💡 Learning Opportunities",
				"level": 2,
			},
		}
		if err := ai.db.WithContext(ctx).Create(&headerBlock).Error; err != nil {
			return fmt.Errorf("failed to create learning opportunities header: %w", err)
		}
		currentOrder++

		// Add each learning item as a bullet point
		for _, item := range learningItems {
			itemBlock := models.Block{
				UserID: note.UserID,
				NoteID: noteID,
				Type:   "bulleted-list",
				Order:  currentOrder,
				Content: map[string]interface{}{
					"text": item,
				},
			}
			if err := ai.db.WithContext(ctx).Create(&itemBlock).Error; err != nil {
				return fmt.Errorf("failed to create learning item block: %w", err)
			}
			currentOrder++
		}
	}

	return nil
}

// storeEmbeddingInChroma stores an embedding in the Chroma vector database
func (ai *AIService) storeEmbeddingInChroma(ctx context.Context, noteID string, embedding []float64) error {
	// Create collection if it doesn't exist
	collectionPayload := map[string]interface{}{
		"name": "note_embeddings",
		"metadata": map[string]string{
			"description": "Embeddings for note similarity search",
		},
	}

	collectionJSON, _ := json.Marshal(collectionPayload)
	collectionReq, err := http.NewRequestWithContext(ctx, "POST", ai.chromaBaseURL+"/api/v1/collections", bytes.NewBuffer(collectionJSON))
	if err != nil {
		return err
	}
	collectionReq.Header.Set("Content-Type", "application/json")
	
	// Try to create collection (will fail if exists, which is fine)
	ai.httpClient.Do(collectionReq)

	// Add the embedding
	addPayload := map[string]interface{}{
		"embeddings": [][]float64{embedding},
		"documents":  []string{noteID}, // Use noteID as document
		"ids":        []string{noteID},
		"metadatas":  []map[string]string{{"note_id": noteID}},
	}

	addJSON, err := json.Marshal(addPayload)
	if err != nil {
		return err
	}

	addReq, err := http.NewRequestWithContext(ctx, "POST", ai.chromaBaseURL+"/api/v1/collections/note_embeddings/add", bytes.NewBuffer(addJSON))
	if err != nil {
		return err
	}
	addReq.Header.Set("Content-Type", "application/json")

	resp, err := ai.httpClient.Do(addReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// queryChromaForSimilar finds similar embeddings in Chroma
func (ai *AIService) queryChromaForSimilar(ctx context.Context, embedding []float64, excludeID string, limit int) ([]string, error) {
	queryPayload := map[string]interface{}{
		"query_embeddings": [][]float64{embedding},
		"n_results":        limit + 1, // Get one extra in case we need to exclude current note
	}

	queryJSON, err := json.Marshal(queryPayload)
	if err != nil {
		return nil, err
	}

	queryReq, err := http.NewRequestWithContext(ctx, "POST", ai.chromaBaseURL+"/api/v1/collections/note_embeddings/query", bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, err
	}
	queryReq.Header.Set("Content-Type", "application/json")

	resp, err := ai.httpClient.Do(queryReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		IDs [][]string `json:"ids"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Extract IDs and exclude the current note
	var relatedIDs []string
	if len(result.IDs) > 0 {
		for _, id := range result.IDs[0] {
			if id != excludeID && len(relatedIDs) < limit {
				relatedIDs = append(relatedIDs, id)
			}
		}
	}

	return relatedIDs, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
