package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"owlistic-notes/owlistic/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ZettelkastenAIService handles AI-powered analysis for the Zettelkasten system
type ZettelkastenAIService struct {
	db        *gorm.DB
	aiService *AIService
}

// NewZettelkastenAIService creates a new ZettelkastenAIService
func NewZettelkastenAIService(db *gorm.DB, aiService *AIService) *ZettelkastenAIService {
	return &ZettelkastenAIService{
		db:        db,
		aiService: aiService,
	}
}

// AITaggingResult represents the result of AI tagging analysis
type AITaggingResult struct {
	Tags []TagSuggestion `json:"tags"`
	Summary string `json:"summary"`
	Confidence float64 `json:"confidence"`
}

// TagSuggestion represents a suggested tag with metadata
type TagSuggestion struct {
	Name string `json:"name"`
	Category string `json:"category"`
	Confidence float64 `json:"confidence"`
	Reason string `json:"reason"`
	Color string `json:"color,omitempty"`
}

// ConnectionSuggestion represents a suggested connection between nodes
type ConnectionSuggestion struct {
	TargetNodeID uuid.UUID `json:"target_node_id"`
	TargetTitle string `json:"target_title"`
	ConnectionType string `json:"connection_type"`
	Strength float64 `json:"strength"`
	Reason string `json:"reason"`
	Confidence float64 `json:"confidence"`
}

// AIConnectionResult represents the result of AI connection analysis
type AIConnectionResult struct {
	Connections []ConnectionSuggestion `json:"connections"`
	Insights []string `json:"insights"`
	Confidence float64 `json:"confidence"`
}

// AnalyzeContentForTags uses AI to suggest tags for content
func (zai *ZettelkastenAIService) AnalyzeContentForTags(ctx context.Context, content, title, nodeType string) (*AITaggingResult, error) {
	prompt := fmt.Sprintf(`Analyze the following %s content and suggest semantic tags that would be useful for a Zettelkasten knowledge management system.

Title: %s

Content: %s

Please provide:
1. 3-8 relevant tags that capture the key concepts, topics, and themes
2. A brief summary (1-2 sentences)
3. Your confidence level (0.0 to 1.0)

For each tag, include:
- name: The tag name (lowercase, use underscores for spaces)
- category: The type of tag ("concept", "topic", "project", "person", "methodology", "technology", etc.)
- confidence: How confident you are this tag is relevant (0.0 to 1.0)
- reason: Brief explanation for why this tag is relevant
- color: A hex color that would be appropriate for this tag category

Respond with valid JSON in this format:
{
  "tags": [
    {
      "name": "tag_name",
      "category": "concept",
      "confidence": 0.9,
      "reason": "This tag captures the main theme because...",
      "color": "#3b82f6"
    }
  ],
  "summary": "Brief summary of the content",
  "confidence": 0.8
}`, nodeType, title, content)

	response, err := zai.aiService.callAnthropic(ctx, prompt, 1000)
	if err != nil {
		return nil, fmt.Errorf("AI tagging analysis failed: %w", err)
	}

	var result AITaggingResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("Failed to parse AI tagging response: %v", err)
		// Fallback: extract basic tags from content
		return zai.extractBasicTags(content, title, nodeType), nil
	}

	return &result, nil
}

// AnalyzeConnections uses AI to suggest connections between nodes
func (zai *ZettelkastenAIService) AnalyzeConnections(ctx context.Context, sourceNode *models.ZettelNode, candidateNodes []models.ZettelNode) (*AIConnectionResult, error) {
	if len(candidateNodes) == 0 {
		return &AIConnectionResult{
			Connections: []ConnectionSuggestion{},
			Insights: []string{},
			Confidence: 1.0,
		}, nil
	}

	// Build candidate descriptions
	candidateDescriptions := make([]string, len(candidateNodes))
	for i, node := range candidateNodes {
		candidateDescriptions[i] = fmt.Sprintf("ID: %s, Type: %s, Title: %s, Summary: %s, Tags: %s",
			node.ID.String(), node.NodeType, node.Title, node.Summary, strings.Join(node.GetTagNames(), ", "))
	}

	prompt := fmt.Sprintf(`Analyze potential connections between the source node and candidate nodes in a Zettelkasten knowledge graph.

Source Node:
- Type: %s
- Title: %s
- Summary: %s
- Tags: %s

Candidate Nodes:
%s

For each candidate that should be connected, determine:
1. connection_type: "related", "depends_on", "references", "contradicts", or "supports"
2. strength: Connection strength from 0.1 to 1.0
3. reason: Why these nodes should be connected
4. confidence: Your confidence in this connection (0.0 to 1.0)

Only suggest meaningful connections (minimum confidence 0.6). Provide insights about patterns you notice.

Respond with valid JSON:
{
  "connections": [
    {
      "target_node_id": "uuid-here",
      "target_title": "Target Title",
      "connection_type": "related",
      "strength": 0.8,
      "reason": "Both discuss similar concepts and could benefit from cross-referencing",
      "confidence": 0.9
    }
  ],
  "insights": [
    "Pattern or insight about the knowledge graph structure"
  ],
  "confidence": 0.8
}`, sourceNode.NodeType, sourceNode.Title, sourceNode.Summary, strings.Join(sourceNode.GetTagNames(), ", "), strings.Join(candidateDescriptions, "\n"))

	response, err := zai.aiService.callAnthropic(ctx, prompt, 1000)
	if err != nil {
		return nil, fmt.Errorf("AI connection analysis failed: %w", err)
	}

	var result AIConnectionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("Failed to parse AI connection response: %v", err)
		// Fallback: simple tag-based connections
		return zai.extractBasicConnections(sourceNode, candidateNodes), nil
	}

	return &result, nil
}

// AnalyzeKnowledgeGaps identifies potential gaps in the knowledge graph
func (zai *ZettelkastenAIService) AnalyzeKnowledgeGaps(ctx context.Context, nodes []models.ZettelNode, tags []models.ZettelTag) (*models.ZettelAnalysis, error) {
	// Build overview of the knowledge graph
	nodesByType := make(map[string]int)
	tagCategories := make(map[string]int)
	
	for _, node := range nodes {
		nodesByType[node.NodeType]++
	}
	
	for _, tag := range tags {
		tagCategories[tag.Category]++
	}

	// Create a summary of the current state
	summary := fmt.Sprintf(`Current Knowledge Graph State:
- Total Nodes: %d
- Notes: %d, Tasks: %d, Projects: %d  
- Total Tags: %d
- Tag Categories: %v

Recent Nodes (sample):`, 
		len(nodes), nodesByType["note"], nodesByType["task"], nodesByType["project"], 
		len(tags), tagCategories)

	// Add sample of recent nodes
	for i, node := range nodes {
		if i >= 10 { // Limit to first 10
			break
		}
		summary += fmt.Sprintf("\n- %s: %s (Tags: %s)", node.NodeType, node.Title, strings.Join(node.GetTagNames(), ", "))
	}

	prompt := fmt.Sprintf(`Analyze this Zettelkasten knowledge graph and identify potential knowledge gaps, missing connections, and recommendations for improvement.

%s

Please identify:
1. Missing topic areas that would strengthen the knowledge base
2. Isolated nodes that might benefit from more connections  
3. Underrepresented categories or themes
4. Opportunities for creating more structured knowledge paths
5. Recommendations for better organization

Provide actionable insights and specific recommendations.

Respond with valid JSON:
{
  "analysis_type": "gap",
  "title": "Knowledge Gap Analysis",
  "description": "Analysis of potential improvements to the knowledge graph",
  "insights": [
    "Specific insight about the current state"
  ],
  "recommendations": [
    "Actionable recommendation for improvement"
  ],
  "confidence": 0.8
}`, summary)

	response, err := zai.aiService.callAnthropic(ctx, prompt, 1000)
	if err != nil {
		return nil, fmt.Errorf("AI gap analysis failed: %w", err)
	}

	var analysisData struct {
		AnalysisType    string   `json:"analysis_type"`
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		Insights        []string `json:"insights"`
		Recommendations []string `json:"recommendations"`
		Confidence      float64  `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(response), &analysisData); err != nil {
		return nil, fmt.Errorf("failed to parse AI gap analysis response: %w", err)
	}

	// Create ZettelAnalysis model
	analysis := &models.ZettelAnalysis{
		AnalysisType:    analysisData.AnalysisType,
		Title:           analysisData.Title,
		Description:     analysisData.Description,
		Insights:        analysisData.Insights,
		Recommendations: analysisData.Recommendations,
		Confidence:      analysisData.Confidence,
		AffectedNodeIDs: []uuid.UUID{}, // Could be populated with specific node IDs
	}

	return analysis, nil
}

// extractBasicTags provides fallback tag extraction when AI fails
func (zai *ZettelkastenAIService) extractBasicTags(content, title, nodeType string) *AITaggingResult {
	tags := []TagSuggestion{}
	
	// Add basic type-based tag
	tags = append(tags, TagSuggestion{
		Name:       nodeType,
		Category:   "type",
		Confidence: 1.0,
		Reason:     fmt.Sprintf("Content is of type %s", nodeType),
		Color:      "#6b7280",
	})

	// Extract some basic keywords (simplified)
	words := strings.Fields(strings.ToLower(content + " " + title))
	wordCount := make(map[string]int)
	
	for _, word := range words {
		if len(word) > 3 { // Only consider words longer than 3 characters
			wordCount[word]++
		}
	}

	// Add most frequent words as basic tags
	for word, count := range wordCount {
		if count > 1 && len(tags) < 5 {
			tags = append(tags, TagSuggestion{
				Name:       word,
				Category:   "keyword",
				Confidence: 0.5,
				Reason:     fmt.Sprintf("Frequent word in content (appears %d times)", count),
				Color:      "#10b981",
			})
		}
	}

	return &AITaggingResult{
		Tags:       tags,
		Summary:    fmt.Sprintf("A %s titled '%s'", nodeType, title),
		Confidence: 0.5,
	}
}

// extractBasicConnections provides fallback connection extraction when AI fails
func (zai *ZettelkastenAIService) extractBasicConnections(sourceNode *models.ZettelNode, candidateNodes []models.ZettelNode) *AIConnectionResult {
	connections := []ConnectionSuggestion{}
	sourceTags := sourceNode.GetTagNames()

	for _, candidate := range candidateNodes {
		if candidate.ID == sourceNode.ID {
			continue
		}

		candidateTags := candidate.GetTagNames()
		sharedTags := 0

		for _, sourceTag := range sourceTags {
			for _, candidateTag := range candidateTags {
				if sourceTag == candidateTag {
					sharedTags++
				}
			}
		}

		// Simple connection based on shared tags
		if sharedTags > 0 {
			strength := float64(sharedTags) / float64(len(sourceTags)+len(candidateTags)-sharedTags)
			if strength > 0.2 { // Minimum threshold
				connections = append(connections, ConnectionSuggestion{
					TargetNodeID:   candidate.ID,
					TargetTitle:    candidate.Title,
					ConnectionType: "related",
					Strength:       strength,
					Reason:         fmt.Sprintf("Shares %d tags", sharedTags),
					Confidence:     0.6,
				})
			}
		}
	}

	return &AIConnectionResult{
		Connections: connections,
		Insights:    []string{"Basic tag-based connections identified"},
		Confidence:  0.6,
	}
}