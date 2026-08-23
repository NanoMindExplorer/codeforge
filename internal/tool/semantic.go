package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codeforge/tui/internal/index"
	"github.com/codeforge/tui/internal/provider"
)

var globalSemanticIndex *index.SemanticIndex

// SemanticSearch queries the semantic vector index (Phase 4).
type SemanticSearch struct {
	WorkDir string
}

func (s *SemanticSearch) Name() string { return "semantic_search" }
func (s *SemanticSearch) Description() string {
	return `Search the project using vector embeddings and cosine similarity.
Finds code by conceptual meaning rather than exact keywords.`
}

func (s *SemanticSearch) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{"type": "string", "description": "Natural language query"},
			"limit": map[string]any{"type": "integer", "description": "Max results (default 5)"},
		},
		"required": []string{"query"},
	}
}

type semanticInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (s *SemanticSearch) Execute(input json.RawMessage) Result {
	var in semanticInput
	if err := json.Unmarshal(input, &in); err != nil {
		return Result{Error: err.Error()}
	}
	if strings.TrimSpace(in.Query) == "" {
		return Result{Error: "query required"}
	}
	
	if globalSemanticIndex == nil {
		// Initialize with default Ollama provider
		prov := provider.NewOllamaEmbedProvider("", "")
		idx := index.NewSemanticIndex(s.WorkDir, prov)
		if err := idx.Load(); err != nil {
			return Result{Error: fmt.Sprintf("failed to load semantic index: %v", err)}
		}
		globalSemanticIndex = idx
	}

	ctx := context.Background()
	hits, err := globalSemanticIndex.Search(ctx, in.Query, in.Limit)
	if err != nil {
		return Result{Error: fmt.Sprintf("semantic search failed: %v", err)}
	}
	
	if len(hits) == 0 {
		return Result{Success: true, Output: "No semantic matches found for: " + in.Query}
	}
	
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Semantic matches for %q:\n\n", in.Query))
	for _, h := range hits {
		b.WriteString(fmt.Sprintf("-- %s (Score: %.2f) --\n", h.Path, h.Score))
		b.WriteString(h.Snippet)
		b.WriteString("\n\n")
	}
	return Result{Success: true, Output: b.String()}
}

// SemanticIndexBuild builds or updates the semantic vector index.
type SemanticIndexBuild struct {
	WorkDir string
}

func (s *SemanticIndexBuild) Name() string { return "build_semantic_index" }
func (s *SemanticIndexBuild) Description() string {
	return `Build or update the semantic vector index for the project by generating embeddings.`
}

func (s *SemanticIndexBuild) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{},
	}
}

func (s *SemanticIndexBuild) Execute(input json.RawMessage) Result {
	if globalSemanticIndex == nil {
		prov := provider.NewOllamaEmbedProvider("", "")
		idx := index.NewSemanticIndex(s.WorkDir, prov)
		_ = idx.Load()
		globalSemanticIndex = idx
	}
	
	// Just use the existing AST indexer to get samples and embed them
	built, err := index.Build(s.WorkDir)
	if err != nil {
		return Result{Error: fmt.Sprintf("failed to parse workspace: %v", err)}
	}
	
	snippets := make(map[string]string)
	for _, d := range built.Docs() {
		// Only take files that have symbols/sample
		if len(d.Sample) > 0 {
			snippets[d.Path] = d.Sample
		}
	}
	
	if len(snippets) == 0 {
		return Result{Success: true, Output: "No code files found to embed."}
	}
	
	ctx := context.Background()
	if err := globalSemanticIndex.AddDocs(ctx, snippets); err != nil {
		return Result{Error: fmt.Sprintf("embedding failed (ensure ollama is running): %v", err)}
	}
	
	if err := globalSemanticIndex.Save(); err != nil {
		return Result{Error: fmt.Sprintf("saving index failed: %v", err)}
	}
	
	return Result{Success: true, Output: fmt.Sprintf("Semantic index built successfully. Embedded %d files.", len(snippets))}
}
