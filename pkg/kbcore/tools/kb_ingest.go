package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/kbcore"
)

var kbIngestInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"content":  {Type: "string", Description: "Document content to ingest (required)"},
		"metadata": {Type: "object", Description: "Optional metadata (e.g. {\"category\": \"semantic\", \"task_type\": \"research\"})"},
	},
	Required: []string{"content"},
}

func KBIngest() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "kb_ingest",
			Description: "Ingest a document into the knowledge base: embed and store. Supports optional metadata for filtering.",
			InputSchema: kbIngestInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Content  string         `json:"content"`
				Metadata map[string]any `json:"metadata"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}

			doc, err := kbcore.Ingest(args.Content, args.Metadata)
			if err != nil {
				return nil, err
			}

			b, _ := json.MarshalIndent(doc, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}
