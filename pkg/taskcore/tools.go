package taskcore

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
)

// ---------- echo ----------

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

// ---------- task_create ----------

var taskCreateInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"title":    {Type: "string", Description: "Task title"},
		"priority": {Type: "string", Description: "Task priority: high, medium, or low"},
	},
	Required: []string{"title"},
}

func TaskCreate() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "task_create",
			Description: "Create a new task",
			InputSchema: taskCreateInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Title    string `json:"title"`
				Priority string `json:"priority"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if args.Priority == "" {
				args.Priority = "medium"
			}

			t := AddTask(Task{Title: args.Title, Priority: args.Priority})

			b, _ := json.MarshalIndent(t, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}

// ---------- task_list ----------

var taskListInputSchema = &jsonschema.Schema{
	Type:       "object",
	Properties: map[string]*jsonschema.Schema{},
}

func TaskList() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "task_list",
			Description: "List all tasks",
			InputSchema: taskListInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tasks := ListTasks()
			b, _ := json.MarshalIndent(tasks, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}
