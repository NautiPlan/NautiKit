package kbcore

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	emb      embedding.Embedder
	splitter document.Transformer
)

func Init(database *gorm.DB) error {
	db = database

	if err := db.AutoMigrate(&Document{}, &Chunk{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	var err error
	emb, err = newEmbedder()
	if err != nil {
		return fmt.Errorf("init embedder: %w", err)
	}

	splitter, err = newSplitter()
	if err != nil {
		return fmt.Errorf("init splitter: %w", err)
	}

	return nil
}

// Close is a no-op when sharing the database with taskcore.
// The database is closed by taskcore.Close().
func Close() error {
	return nil
}

func chunkCfg() (size, overlap int) {
	size = 512
	overlap = 64
	if v := os.Getenv("NAUTIKIT_CHUNK_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			size = n
		}
	}
	if v := os.Getenv("NAUTIKIT_CHUNK_OVERLAP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			overlap = n
		}
	}
	return
}

func Ingest(content string, metadata map[string]any) (Document, error) {
	doc := Document{
		Content:   content,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	db.Create(&doc)

	pieces, err := splitDocs(content, splitter)
	if err != nil {
		return doc, fmt.Errorf("split: %w", err)
	}

	vecs, err := emb.EmbedStrings(context.Background(), pieces)
	if err != nil {
		return doc, fmt.Errorf("embed: %w", err)
	}

	for i, piece := range pieces {
		db.Create(&Chunk{
			DocumentID: doc.ID,
			Index:      i,
			Content:    piece,
			Embedding:  vecs[i],
		})
	}

	return doc, nil
}

func Search(query string, k int, threshold float64, filter map[string]any) ([]Document, error) {
	vecs, err := emb.EmbedStrings(context.Background(), []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	qvec := vecs[0]

	docScores := make(map[uint]float64)
	var chunks []Chunk
	db.Find(&chunks)
	for _, c := range chunks {
		score := cosine(qvec, c.Embedding)
		if score < threshold {
			continue
		}
		if score > docScores[c.DocumentID] {
			docScores[c.DocumentID] = score
		}
	}

	type docScore struct {
		doc   Document
		score float64
	}
	var list []docScore
	for docID, score := range docScores {
		var d Document
		if db.First(&d, docID).Error != nil {
			continue
		}
		if !matchFilter(d.Metadata, filter) {
			continue
		}
		list = append(list, docScore{doc: d, score: score})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].score > list[j].score
	})

	if k > len(list) {
		k = len(list)
	}
	out := make([]Document, k)
	for i := 0; i < k; i++ {
		out[i] = list[i].doc
	}
	return out, nil
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
	db.Where("document_id = ?", id).Delete(&Chunk{})
	return db.Delete(&Document{}, id).Error
}

func ClearDocuments() error {
	db.Where("1 = 1").Delete(&Chunk{})
	db.Where("1 = 1").Delete(&Document{})
	return nil
}

// --- helpers ---

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
