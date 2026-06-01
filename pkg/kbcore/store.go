package kbcore

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	embedder Embedder
)

func Init(path string) error {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		path = filepath.Join(home, ".nautikit", "data.db")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	var err error
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err := db.AutoMigrate(&Document{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	embedder = NewEmbedder()
	return nil
}

func Close() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying db: %w", err)
	}
	return sqlDB.Close()
}

// Ingest embeds content and stores a new document.
func Ingest(content string, metadata map[string]any) (Document, error) {
	vec, err := embedder.Embed(content)
	if err != nil {
		return Document{}, fmt.Errorf("embed: %w", err)
	}

	summary := content
	if len([]rune(summary)) > 200 {
		summary = string([]rune(summary)[:200])
	}

	doc := Document{
		Content:   content,
		Summary:   summary,
		Embedding: vec,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	db.Create(&doc)
	return doc, nil
}

// Search embeds the query and returns top-k documents by cosine similarity.
func Search(query string, k int, threshold float64, filter map[string]any) ([]Document, error) {
	qvec, err := embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	var docs []Document
	db.Find(&docs)

	type scored struct {
		doc   Document
		score float64
	}
	var candidates []scored

	for _, d := range docs {
		// metadata filter
		if !matchFilter(d.Metadata, filter) {
			continue
		}
		score := cosine(qvec, d.Embedding)
		if score >= threshold {
			candidates = append(candidates, scored{doc: d, score: score})
		}
	}

	// sort descending
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	if k > len(candidates) {
		k = len(candidates)
	}

	out := make([]Document, k)
	for i := 0; i < k; i++ {
		out[i] = candidates[i].doc
	}
	return out, nil
}

func matchFilter(meta map[string]any, filter map[string]any) bool {
	if filter == nil {
		return true
	}
	for k, v := range filter {
		mv, ok := meta[k]
		if !ok {
			return false
		}
		if fmt.Sprint(mv) != fmt.Sprint(v) {
			return false
		}
	}
	return true
}

func cosine(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func ListDocuments(filter map[string]any) []Document {
	var docs []Document
	db.Find(&docs)
	if filter == nil {
		return docs
	}
	out := make([]Document, 0, len(docs))
	for _, d := range docs {
		if matchFilter(d.Metadata, filter) {
			out = append(out, d)
		}
	}
	return out
}

func DeleteDocument(id uint) error {
	return db.Delete(&Document{}, id).Error
}

func ClearDocuments() error {
	return db.Where("1 = 1").Delete(&Document{}).Error
}
