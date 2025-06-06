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
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChromaService handles all interactions with ChromaDB for vector embeddings
type ChromaService struct {
	baseURL    string
	httpClient *http.Client
	db         *gorm.DB
}

// ChromaCollection represents a collection in ChromaDB
type ChromaCollection struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaConfiguration for collection settings
type ChromaConfiguration struct {
	HNSW *HNSWConfig `json:"hnsw,omitempty"`
}

// HNSWConfig for index configuration
type HNSWConfig struct {
	Space           string `json:"space,omitempty"`           // "l2", "cosine", or "ip"
	EFConstruction  int    `json:"ef_construction,omitempty"`  // default: 100
	EFSearch        int    `json:"ef_search,omitempty"`        // default: 100
	MaxNeighbors    int    `json:"max_neighbors,omitempty"`    // default: 16
	NumThreads      int    `json:"num_threads,omitempty"`      // default: CPU count
	BatchSize       int    `json:"batch_size,omitempty"`       // default: 100
	SyncThreshold   int    `json:"sync_threshold,omitempty"`   // default: 1000
	ResizeFactor    float32 `json:"resize_factor,omitempty"`   // default: 1.2
}

// ChromaAddRequest for adding documents
type ChromaAddRequest struct {
	Documents  []string                 `json:"documents,omitempty"`
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
	IDs        []string                 `json:"ids"`
}

// ChromaQueryRequest for querying collections
type ChromaQueryRequest struct {
	QueryEmbeddings [][]float64              `json:"query_embeddings,omitempty"`
	QueryTexts      []string                 `json:"query_texts,omitempty"`
	NResults        int                      `json:"n_results"`
	Where           map[string]interface{}   `json:"where,omitempty"`
	WhereDocument   map[string]interface{}   `json:"where_document,omitempty"`
	Include         []string                 `json:"include,omitempty"`
}

// ChromaQueryResponse from query operations
type ChromaQueryResponse struct {
	IDs        [][]string                   `json:"ids"`
	Embeddings [][][]float64                `json:"embeddings,omitempty"`
	Documents  [][]string                   `json:"documents,omitempty"`
	Metadatas  [][]map[string]interface{}   `json:"metadatas,omitempty"`
	Distances  [][]float64                  `json:"distances,omitempty"`
}

// ChromaGetResponse from get operations
type ChromaGetResponse struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	Documents  []string                 `json:"documents,omitempty"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// NewChromaService creates a new ChromaDB service
func NewChromaService(baseURL string, db *gorm.DB) *ChromaService {
	if baseURL == "" {
		// Try different ChromaDB configurations
		if chromaURL := os.Getenv("CHROMA_BASE_URL"); chromaURL != "" {
			baseURL = chromaURL
		} else {
			baseURL = "http://192.168.0.11:8000" // Default to external ChromaDB server
		}
	}
	
	return &ChromaService{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second}, // Reduced to minimize context goroutines
		db:         db,
	}
}

// CreateCollection creates a new collection with specified configuration
func (cs *ChromaService) CreateCollection(ctx context.Context, name string, config *ChromaConfiguration) error {
	// Collection configuration for optimal note search  
	payload := map[string]interface{}{
		"name": name,
		"get_or_create": true, // Use get_or_create to avoid conflicts
		"metadata": map[string]interface{}{
			"description": "Owlistic AI note embeddings collection",
			"created":     time.Now().Format(time.RFC3339),
		},
	}
	
	// Add configuration if provided
	if config != nil {
		configMap := map[string]interface{}{}
		if config.HNSW != nil {
			hnswMap := map[string]interface{}{
				"space": config.HNSW.Space,
			}
			if config.HNSW.EFConstruction > 0 {
				hnswMap["ef_construction"] = config.HNSW.EFConstruction
			}
			if config.HNSW.EFSearch > 0 {
				hnswMap["ef_search"] = config.HNSW.EFSearch
			}
			if config.HNSW.MaxNeighbors > 0 {
				hnswMap["max_neighbors"] = config.HNSW.MaxNeighbors
			}
			configMap["hnsw"] = hnswMap
		}
		
		if len(configMap) > 0 {
			payload["configuration"] = configMap
		}
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal create collection request: %w", err)
	}
	
	// Use ChromaDB v2 API with default tenant and database
	url := cs.getCollectionsURL()
	log.Printf("Creating ChromaDB collection at URL: %s", url)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	log.Printf("Sending request with payload: %s", string(jsonData))
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	log.Printf("ChromaDB response: status=%d, body=%s", resp.StatusCode, string(body))
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Log the error but don't fail completely - AI features can work without vector search
		log.Printf("ChromaDB collection creation failed (status %d): %s", resp.StatusCode, string(body))
		log.Printf("This likely means ChromaDB is using a different API version or is not properly configured")
		log.Printf("Vector search will be disabled, but all other AI features will continue to work")
		return fmt.Errorf("failed to create collection, status %d: %s", resp.StatusCode, string(body))
	}
	
	log.Printf("Successfully created ChromaDB collection: %s", name)
	return nil
}

// GetOrCreateCollection gets an existing collection or creates it if it doesn't exist
func (cs *ChromaService) GetOrCreateCollection(ctx context.Context, name string, config *ChromaConfiguration) error {
	// For v2 API, we need to ensure tenant and database exist first
	tenant := "default_tenant" 
	database := "default_database"
	
	// Create tenant if needed
	if err := cs.ensureTenant(ctx, tenant); err != nil {
		log.Printf("Warning: Could not ensure tenant exists: %v", err)
	}
	
	// Create database if needed
	if err := cs.ensureDatabase(ctx, tenant, database); err != nil {
		log.Printf("Warning: Could not ensure database exists: %v", err)
	}
	
	// Create the collection
	return cs.CreateCollection(ctx, name, config)
}

// DeleteCollection deletes a collection and all its data
func (cs *ChromaService) DeleteCollection(ctx context.Context, name string) error {
	url := cs.getCollectionURL(name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete collection, status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// AddDocuments adds documents to a collection (embeddings will be generated automatically)
func (cs *ChromaService) AddDocuments(ctx context.Context, collectionName string, ids []string, documents []string, metadatas []map[string]interface{}) error {
	payload := ChromaAddRequest{
		IDs:       ids,
		Documents: documents,
		Metadatas: metadatas,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal add request: %w", err)
	}
	
	url := cs.getCollectionURL(collectionName) + "/add"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add documents, status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// UpdateDocuments updates existing documents in a collection
func (cs *ChromaService) UpdateDocuments(ctx context.Context, collectionName string, ids []string, documents []string, metadatas []map[string]interface{}) error {
	payload := ChromaAddRequest{
		IDs:       ids,
		Documents: documents,
		Metadatas: metadatas,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}
	
	url := cs.getCollectionURL(collectionName) + "/update"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update documents: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update documents, status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// UpsertDocuments inserts or updates documents in a collection
func (cs *ChromaService) UpsertDocuments(ctx context.Context, collectionName string, ids []string, documents []string, metadatas []map[string]interface{}) error {
	// Ensure collection exists first
	if err := cs.GetOrCreateCollection(ctx, collectionName, nil); err != nil {
		return fmt.Errorf("failed to ensure collection exists: %w", err)
	}
	
	// Check which documents already exist
	existingDocs, err := cs.GetDocuments(ctx, collectionName, ids)
	if err != nil {
		// If get fails, assume none exist and try to add all
		log.Printf("Failed to get existing documents (collection might be empty), trying to add all: %v", err)
		return cs.AddDocuments(ctx, collectionName, ids, documents, metadatas)
	}
	
	// Separate existing and new documents
	existingIDsMap := make(map[string]bool)
	for _, id := range existingDocs.IDs {
		existingIDsMap[id] = true
	}
	
	var newIDs, updateIDs []string
	var newDocs, updateDocs []string
	var newMetas, updateMetas []map[string]interface{}
	
	for i, id := range ids {
		if existingIDsMap[id] {
			updateIDs = append(updateIDs, id)
			updateDocs = append(updateDocs, documents[i])
			updateMetas = append(updateMetas, metadatas[i])
		} else {
			newIDs = append(newIDs, id)
			newDocs = append(newDocs, documents[i])
			newMetas = append(newMetas, metadatas[i])
		}
	}
	
	// Add new documents
	if len(newIDs) > 0 {
		log.Printf("Adding %d new documents to ChromaDB", len(newIDs))
		if err := cs.AddDocuments(ctx, collectionName, newIDs, newDocs, newMetas); err != nil {
			return fmt.Errorf("failed to add new documents: %w", err)
		}
	}
	
	// Update existing documents
	if len(updateIDs) > 0 {
		log.Printf("Updating %d existing documents in ChromaDB", len(updateIDs))
		if err := cs.UpdateDocuments(ctx, collectionName, updateIDs, updateDocs, updateMetas); err != nil {
			return fmt.Errorf("failed to update existing documents: %w", err)
		}
	}
	
	log.Printf("Successfully upserted %d documents (%d new, %d updated)", len(ids), len(newIDs), len(updateIDs))
	return nil
}

// DeleteDocuments deletes documents from a collection by IDs
func (cs *ChromaService) DeleteDocuments(ctx context.Context, collectionName string, ids []string) error {
	payload := map[string]interface{}{
		"ids": ids,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}
	
	url := cs.getCollectionURL(collectionName) + "/delete"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete documents, status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// QueryByText queries the collection using text (will be embedded automatically)
func (cs *ChromaService) QueryByText(ctx context.Context, collectionName string, queryTexts []string, nResults int, where map[string]interface{}) (*ChromaQueryResponse, error) {
	payload := ChromaQueryRequest{
		QueryTexts: queryTexts,
		NResults:   nResults,
		Where:      where,
		Include:    []string{"documents", "metadatas", "distances"},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}
	
	url := cs.getCollectionURL(collectionName) + "/query"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query collection, status %d: %s", resp.StatusCode, string(body))
	}
	
	var result ChromaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}
	
	return &result, nil
}

// QueryByEmbedding queries the collection using pre-computed embeddings
func (cs *ChromaService) QueryByEmbedding(ctx context.Context, collectionName string, queryEmbeddings [][]float64, nResults int, where map[string]interface{}) (*ChromaQueryResponse, error) {
	payload := ChromaQueryRequest{
		QueryEmbeddings: queryEmbeddings,
		NResults:        nResults,
		Where:           where,
		Include:         []string{"documents", "metadatas", "distances"},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}
	
	url := cs.getCollectionURL(collectionName) + "/query"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query collection, status %d: %s", resp.StatusCode, string(body))
	}
	
	var result ChromaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}
	
	return &result, nil
}

// GetDocuments gets documents by IDs
func (cs *ChromaService) GetDocuments(ctx context.Context, collectionName string, ids []string) (*ChromaGetResponse, error) {
	payload := map[string]interface{}{
		"ids":     ids,
		"include": []string{"documents", "metadatas", "embeddings"},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get request: %w", err)
	}
	
	url := cs.getCollectionURL(collectionName) + "/get"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get documents, status %d: %s", resp.StatusCode, string(body))
	}
	
	var result ChromaGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}
	
	return &result, nil
}

// CountDocuments returns the number of documents in a collection
func (cs *ChromaService) CountDocuments(ctx context.Context, collectionName string) (int, error) {
	url := cs.getCollectionURL(collectionName) + "/count"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to count documents, status %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}
	
	count, ok := result["count"]
	if !ok {
		return 0, fmt.Errorf("count not found in response")
	}
	
	return count, nil
}

// Helper function to convert note ID to ChromaDB document ID
func NoteIDToChromaID(noteID uuid.UUID) string {
	return fmt.Sprintf("note_%s", noteID.String())
}

// ensureTenant creates a tenant if it doesn't exist
func (cs *ChromaService) ensureTenant(ctx context.Context, tenantName string) error {
	payload := map[string]interface{}{
		"name": tenantName,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant request: %w", err)
	}
	
	url := cs.baseURL + "/api/v2/tenants"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create tenant request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Tenant might already exist, which is fine
		log.Printf("Tenant creation response: status=%d, body=%s", resp.StatusCode, string(body))
	}
	
	return nil
}

// getCollectionURL returns the v2 API URL for a collection
func (cs *ChromaService) getCollectionURL(collectionName string) string {
	tenant := "default_tenant"
	database := "default_database" 
	return fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s", cs.baseURL, tenant, database, collectionName)
}

// getCollectionsURL returns the v2 API URL for collections
func (cs *ChromaService) getCollectionsURL() string {
	tenant := "default_tenant"
	database := "default_database"
	return fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections", cs.baseURL, tenant, database)
}

// ensureDatabase creates a database if it doesn't exist
func (cs *ChromaService) ensureDatabase(ctx context.Context, tenantName, databaseName string) error {
	payload := map[string]interface{}{
		"name": databaseName,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal database request: %w", err)
	}
	
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases", cs.baseURL, tenantName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create database request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Database might already exist, which is fine
		log.Printf("Database creation response: status=%d, body=%s", resp.StatusCode, string(body))
	}
	
	return nil
}

