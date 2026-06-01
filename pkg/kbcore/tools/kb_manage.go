package tools

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/kbcore"
)

var kbListInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"filter": {Type: "object", Description: "Optional metadata filter (e.g. {\"category\": \"semantic\"})"},
	},
}

func KBList() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "kb_list",
			Description: "List all documents in the knowledge base, optionally filtered by metadata. Returns id, summary, metadata — no embedding vectors.",
			InputSchema: kbListInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var raw map[string]any
			json.Unmarshal(req.Params.Arguments, &raw)

			var filter map[string]any
			if v, ok := raw["filter"].(map[string]any); ok {
				filter = v
			}

			docs := kbcore.ListDocuments(filter)
			b, _ := json.MarshalIndent(docs, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}

var kbDeleteInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"id": {Type: "string", Description: "Document ID to delete (required)"},
	},
	Required: []string{"id"},
}

func KBDelete() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "kb_delete",
			Description: "Delete a single document from the knowledge base by ID.",
			InputSchema: kbDeleteInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}

			id, err := strconv.ParseUint(args.ID, 10, 64)
			if err != nil {
				return nil, err
			}

			if err := kbcore.DeleteDocument(uint(id)); err != nil {
				return nil, err
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}},
			}, nil
		},
	}
}

var kbClearInputSchema = &jsonschema.Schema{
	Type:       "object",
	Properties: map[string]*jsonschema.Schema{},
}

func KBClear() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "kb_clear",
			Description: "Delete all documents from the knowledge base. Irreversible.",
			InputSchema: kbClearInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if err := kbcore.ClearDocuments(); err != nil {
				return nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "cleared"}},
			}, nil
		},
	}
}
