package index

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/codeforge/tui/internal/provider"
)

type VectorDoc struct {
	Path      string    `json:"path"`
	Snippet   string    `json:"snippet"`
	Embedding []float32 `json:"embedding"`
}

type SemanticIndex struct {
	mu   sync.RWMutex
	docs []VectorDoc
	root string
	prov provider.EmbeddingProvider
}

func NewSemanticIndex(root string, prov provider.EmbeddingProvider) *SemanticIndex {
	return &SemanticIndex{
		root: root,
		prov: prov,
	}
}

func (s *SemanticIndex) indexPath() string {
	return filepath.Join(s.root, ".codeforge", "index", "vectors.json")
}

func (s *SemanticIndex) Load() error {
	b, err := os.ReadFile(s.indexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(b, &s.docs)
}

func (s *SemanticIndex) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.indexPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(s.docs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.indexPath(), b, 0644)
}

// cosineSimilarity calculates cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

type SemanticHit struct {
	Path    string
	Snippet string
	Score   float64
}

// Search performs semantic search by embedding the query and comparing cosine similarity.
func (s *SemanticIndex) Search(ctx context.Context, query string, limit int) ([]SemanticHit, error) {
	if s.prov == nil {
		return nil, fmt.Errorf("no embedding provider configured")
	}
	if limit <= 0 {
		limit = 5
	}

	emb, err := s.prov.EmbedTexts(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(emb) == 0 || len(emb[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}
	qVec := emb[0]

	s.mu.RLock()
	defer s.mu.RUnlock()

	var hits []SemanticHit
	for _, d := range s.docs {
		score := cosineSimilarity(qVec, d.Embedding)
		if score > 0.4 { // basic threshold
			hits = append(hits, SemanticHit{
				Path:    d.Path,
				Snippet: d.Snippet,
				Score:   score,
			})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

// AddDocs embeds and adds new documents to the index.
func (s *SemanticIndex) AddDocs(ctx context.Context, snippets map[string]string) error {
	if s.prov == nil {
		return fmt.Errorf("no embedding provider configured")
	}

	var paths []string
	var texts []string
	for p, text := range snippets {
		paths = append(paths, p)
		texts = append(texts, text)
	}

	// Batch processing (naive)
	embeddings, err := s.prov.EmbedTexts(ctx, texts)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i, emb := range embeddings {
		// Replace existing path
		replaced := false
		for j, ex := range s.docs {
			if ex.Path == paths[i] {
				s.docs[j].Snippet = texts[i]
				s.docs[j].Embedding = emb
				replaced = true
				break
			}
		}
		if !replaced {
			s.docs = append(s.docs, VectorDoc{
				Path:      paths[i],
				Snippet:   texts[i],
				Embedding: emb,
			})
		}
	}
	return nil
}
