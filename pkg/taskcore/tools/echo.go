package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
)

var echoInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"message": {Type: "string", Description: "The message to echo back"},
	},
	Required: []string{"message"},
}

func Echo() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "echo",
			Description: "Echo back the input message",
			InputSchema: echoInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: args.Message}},
			}, nil
		},
	}
}
