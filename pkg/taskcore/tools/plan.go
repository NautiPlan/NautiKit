package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/taskcore"
)

var planCreateInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"title":       {Type: "string", Description: "Plan title"},
		"description": {Type: "string", Description: "Plan description"},
	},
	Required: []string{"title"},
}

func PlanCreate() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "plan_create",
			Description: "Create a new plan",
			InputSchema: planCreateInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}

			p := taskcore.AddPlan(taskcore.Plan{Title: args.Title, Description: args.Description})

			b, _ := json.MarshalIndent(p, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}

var planListInputSchema = &jsonschema.Schema{
	Type:       "object",
	Properties: map[string]*jsonschema.Schema{},
}

func PlanList() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "plan_list",
			Description: "List all plans",
			InputSchema: planListInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			plans := taskcore.ListPlans()
			b, _ := json.MarshalIndent(plans, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}
