package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChromaServiceIntegration tests the ChromaDB service integration
// Note: This requires ChromaDB to be running (use docker-compose up chroma)
func TestChromaServiceIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Initialize service
	chromaService := NewChromaService("http://localhost:8001", nil)
	ctx := context.Background()

	// Test collection name
	testCollection := fmt.Sprintf("test_collection_%d", time.Now().Unix())

	// Clean up after test
	defer func() {
		chromaService.DeleteCollection(ctx, testCollection)
	}()

	t.Run("CreateCollection", func(t *testing.T) {
		config := &ChromaConfiguration{
			HNSW: &HNSWConfig{
				Space:          "cosine",
				EFConstruction: 100,
				EFSearch:       50,
			},
		}

		err := chromaService.CreateCollection(ctx, testCollection, config)
		assert.NoError(t, err)
	})

	t.Run("AddDocuments", func(t *testing.T) {
		ids := []string{"doc1", "doc2", "doc3"}
		documents := []string{
			"Machine learning is a subset of artificial intelligence",
			"Deep learning uses neural networks with multiple layers",
			"Natural language processing helps computers understand human language",
		}
		metadatas := []map[string]interface{}{
			{"topic": "ml", "difficulty": "beginner"},
			{"topic": "dl", "difficulty": "advanced"},
			{"topic": "nlp", "difficulty": "intermediate"},
		}

		err := chromaService.AddDocuments(ctx, testCollection, ids, documents, metadatas)
		assert.NoError(t, err)
	})

	t.Run("QueryByText", func(t *testing.T) {
		queries := []string{"neural networks"}
		results, err := chromaService.QueryByText(ctx, testCollection, queries, 2, nil)
		
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.IDs, 1)        // One query
		assert.Len(t, results.IDs[0], 2)     // Two results
		
		// Check that deep learning doc is in results (should be most relevant)
		foundDeepLearning := false
		for _, id := range results.IDs[0] {
			if id == "doc2" {
				foundDeepLearning = true
				break
			}
		}
		assert.True(t, foundDeepLearning, "Expected deep learning document in results")
	})

	t.Run("QueryWithMetadataFilter", func(t *testing.T) {
		queries := []string{"learning"}
		where := map[string]interface{}{
			"difficulty": "beginner",
		}
		
		results, err := chromaService.QueryByText(ctx, testCollection, queries, 3, where)
		
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.IDs, 1)
		
		// Should only get the ML document (beginner difficulty)
		if len(results.IDs[0]) > 0 {
			assert.Equal(t, "doc1", results.IDs[0][0])
		}
	})

	t.Run("UpdateDocuments", func(t *testing.T) {
		ids := []string{"doc1"}
		documents := []string{"Machine learning and AI are transforming technology"}
		metadatas := []map[string]interface{}{
			{"topic": "ml", "difficulty": "beginner", "updated": true},
		}

		err := chromaService.UpdateDocuments(ctx, testCollection, ids, documents, metadatas)
		assert.NoError(t, err)
	})

	t.Run("GetDocuments", func(t *testing.T) {
		ids := []string{"doc1", "doc2"}
		results, err := chromaService.GetDocuments(ctx, testCollection, ids)
		
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.IDs, 2)
		assert.Len(t, results.Documents, 2)
		assert.Len(t, results.Metadatas, 2)
		
		// Check updated metadata
		for i, id := range results.IDs {
			if id == "doc1" {
				assert.Equal(t, true, results.Metadatas[i]["updated"])
			}
		}
	})

	t.Run("CountDocuments", func(t *testing.T) {
		count, err := chromaService.CountDocuments(ctx, testCollection)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		err := chromaService.DeleteDocuments(ctx, testCollection, []string{"doc3"})
		assert.NoError(t, err)

		// Verify deletion
		count, err := chromaService.CountDocuments(ctx, testCollection)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("UpsertDocuments", func(t *testing.T) {
		ids := []string{"doc3", "doc4"}
		documents := []string{
			"Reinforcement learning trains agents through rewards",
			"Computer vision helps machines interpret visual data",
		}
		metadatas := []map[string]interface{}{
			{"topic": "rl", "difficulty": "advanced"},
			{"topic": "cv", "difficulty": "intermediate"},
		}

		err := chromaService.UpsertDocuments(ctx, testCollection, ids, documents, metadatas)
		assert.NoError(t, err)

		// Verify count
		count, err := chromaService.CountDocuments(ctx, testCollection)
		require.NoError(t, err)
		assert.Equal(t, 4, count)
	})
}

// TestNoteIDConversion tests the ID conversion helpers
func TestNoteIDConversion(t *testing.T) {
	noteID := uuid.New()
	
	// Test conversion to ChromaDB ID
	chromaID := NoteIDToChromaID(noteID)
	assert.Equal(t, fmt.Sprintf("note_%s", noteID.String()), chromaID)
	
	// Test conversion back to note ID
	convertedID, err := ChromaIDToNoteID(chromaID)
	assert.NoError(t, err)
	assert.Equal(t, noteID, convertedID)
	
	// Test invalid ChromaDB ID
	_, err = ChromaIDToNoteID("invalid_id")
	assert.Error(t, err)
}

// BenchmarkChromaQuery benchmarks ChromaDB query performance
func BenchmarkChromaQuery(b *testing.B) {
	// Skip if not in benchmark mode
	if testing.Short() {
		b.Skip("Skipping benchmark")
	}

	chromaService := NewChromaService("http://localhost:8001", nil)
	ctx := context.Background()
	testCollection := fmt.Sprintf("bench_collection_%d", time.Now().Unix())

	// Setup
	chromaService.CreateCollection(ctx, testCollection, nil)
	defer chromaService.DeleteCollection(ctx, testCollection)

	// Add test documents
	ids := make([]string, 100)
	documents := make([]string, 100)
	metadatas := make([]map[string]interface{}, 100)
	
	for i := 0; i < 100; i++ {
		ids[i] = fmt.Sprintf("doc_%d", i)
		documents[i] = fmt.Sprintf("This is test document number %d with some content about various topics", i)
		metadatas[i] = map[string]interface{}{
			"index": i,
			"type":  "test",
		}
	}
	
	chromaService.AddDocuments(ctx, testCollection, ids, documents, metadatas)

	// Benchmark queries
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queries := []string{fmt.Sprintf("document number %d", i%100)}
		chromaService.QueryByText(ctx, testCollection, queries, 5, nil)
	}
}