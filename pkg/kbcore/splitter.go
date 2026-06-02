package kbcore

import (
	"context"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/schema"
)

func newSplitter() (document.Transformer, error) {
	size, overlap := chunkCfg()
	return recursive.NewSplitter(context.Background(), &recursive.Config{
		ChunkSize:   size,
		OverlapSize: overlap,
		Separators:  []string{"\n\n", "\n", "。", ".", " ", ""},
	})
}

func splitDocs(text string, s document.Transformer) ([]string, error) {
	docs, err := s.Transform(context.Background(), []*schema.Document{{
		Content: text,
	}})
	if err != nil {
		return nil, err
	}
	pieces := make([]string, len(docs))
	for i, d := range docs {
		pieces[i] = d.Content
	}
	return pieces, nil
}
