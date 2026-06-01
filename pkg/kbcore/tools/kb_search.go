package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/kbcore"
)

var kbSearchInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"query":     {Type: "string", Description: "Search query (required)"},
		"k":         {Type: "number", Description: "Number of results to return (default 5)"},
		"threshold": {Type: "number", Description: "Minimum cosine similarity threshold (default 0.5)"},
		"filter":    {Type: "object", Description: "Optional metadata filter (e.g. {\"category\": \"semantic\"})"},
	},
	Required: []string{"query"},
}

func KBSearch() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "kb_search",
			Description: "Search the knowledge base by vector similarity. Returns top-k documents above the threshold, optionally filtered by metadata.",
			InputSchema: kbSearchInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var raw map[string]any
			if err := json.Unmarshal(req.Params.Arguments, &raw); err != nil {
				return nil, err
			}

			query, _ := raw["query"].(string)

			k := 5
			if v, ok := raw["k"].(float64); ok {
				k = int(v)
			}

			threshold := 0.5
			if v, ok := raw["threshold"].(float64); ok {
				threshold = v
			}

			var filter map[string]any
			if v, ok := raw["filter"].(map[string]any); ok {
				filter = v
			}

			docs, err := kbcore.Search(query, k, threshold, filter)
			if err != nil {
				return nil, err
			}

			b, _ := json.MarshalIndent(docs, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}
