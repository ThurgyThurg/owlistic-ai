package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	db            *gorm.DB
	anthropicKey  string
	openaiKey     string
	chromaBaseURL string
	httpClient    *http.Client
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

type OpenAIEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func NewAIService(db *gorm.DB) *AIService {
	return &AIService{
		db:            db,
		anthropicKey:  os.Getenv("ANTHROPIC_API_KEY"),
		openaiKey:     os.Getenv("OPENAI_API_KEY"),
		chromaBaseURL: os.Getenv("CHROMA_BASE_URL"),
		httpClient:    &http.Client{Timeout: 30 * time.Second},
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
	errChan := make(chan error, 4)

	// Generate title if empty
	go func() {
		if note.Title == "" {
			title, err := ai.generateTitle(content)
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
		summary, err := ai.generateSummary(content)
		if err != nil {
			errChan <- err
			return
		}
		summaryChan <- summary
	}()

	// Extract tags
	go func() {
		tags, err := ai.extractTags(content, note.Title)
		if err != nil {
			errChan <- err
			return
		}
		tagsChan <- tags
	}()

	// Generate embeddings
	go func() {
		embedding, err := ai.createEmbedding(fmt.Sprintf("%s\n\n%s", note.Title, content))
		if err != nil {
			errChan <- err
			return
		}
		embeddingChan <- embedding
	}()

	// Collect results
	var finalTitle string
	var summary string
	var aiTags []string
	var embeddings []float64

	for i := 0; i < 4; i++ {
		select {
		case title := <-titleChan:
			finalTitle = title
		case s := <-summaryChan:
			summary = s
		case tags := <-tagsChan:
			aiTags = tags
		case emb := <-embeddingChan:
			embeddings = emb
		case err := <-errChan:
			log.Printf("AI processing error: %v", err)
			// Continue with partial results
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Update or create AI-enhanced note record
	aiNote := models.AIEnhancedNote{
		Note:             note,
		Summary:          summary,
		AITags:           aiTags,
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

func (ai *AIService) generateTitle(content string) (string, error) {
	prompt := fmt.Sprintf(`Generate a concise, descriptive title (max 10 words) for this content:

%s

Title:`, content[:min(500, len(content))])

	response, err := ai.callAnthropic(prompt, 50)
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(strings.Trim(response, `"`))
	if len(title) > 100 {
		title = title[:100]
	}

	return title, nil
}

func (ai *AIService) generateSummary(content string) (string, error) {
	if len(content) < 200 {
		return content, nil
	}

	prompt := fmt.Sprintf(`Provide a concise 2-3 sentence summary of this content:

%s

Summary:`, content)

	return ai.callAnthropic(prompt, 150)
}

func (ai *AIService) extractTags(content, title string) ([]string, error) {
	prompt := fmt.Sprintf(`Extract 3-7 relevant tags from this content. Tags should be:
- Single words or short phrases (2-3 words max)
- Lowercase
- Descriptive and useful for categorization

Title: %s
Content: %s

Return only the tags as a JSON array. Example: ["productivity", "machine learning", "project planning"]

Tags:`, title, content[:min(800, len(content))])

	response, err := ai.callAnthropic(prompt, 100)
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

func (ai *AIService) createEmbedding(text string) ([]float64, error) {
	req := OpenAIEmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: text[:min(8000, len(text))],
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ai.openaiKey)

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embResp OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return embResp.Data[0].Embedding, nil
}

func (ai *AIService) findRelatedNotes(ctx context.Context, embedding []float64, excludeID uuid.UUID) ([]uuid.UUID, error) {
	// This would integrate with ChromaDB or implement similarity search
	// For now, return empty slice - this needs vector database integration
	return []uuid.UUID{}, nil
}

func (ai *AIService) callAnthropic(prompt string, maxTokens int) (string, error) {
	req := AnthropicRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: maxTokens,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ai.anthropicKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := ai.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", err
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

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
	queryEmbedding, err := ai.createEmbedding(query)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
