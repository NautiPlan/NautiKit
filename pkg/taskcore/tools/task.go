package tools

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/taskcore"
)

var taskCreateInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"plan_id":     {Type: "string", Description: "Plan ID this task belongs to"},
		"title":       {Type: "string", Description: "Task title"},
		"description": {Type: "string", Description: "Task description"},
		"date":        {Type: "string", Description: "Scheduled date, e.g. 2026-05-28"},
		"priority":    {Type: "string", Description: "Task priority: high, medium, or low"},
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
				PlanID      string `json:"plan_id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Date        string `json:"date"`
				Priority    string `json:"priority"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if args.Priority == "" {
				args.Priority = "medium"
			}

			var planID uint
			if args.PlanID != "" {
				id, err := strconv.ParseUint(args.PlanID, 10, 64)
				if err != nil {
					return nil, err
				}
				planID = uint(id)
			}

			t := taskcore.AddTask(taskcore.Task{
				PlanID:      planID,
				Title:       args.Title,
				Description: args.Description,
				Date:        args.Date,
				Priority:    args.Priority,
			})

			b, _ := json.MarshalIndent(t, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}

var taskListInputSchema = &jsonschema.Schema{
	Type: "object",
	Properties: map[string]*jsonschema.Schema{
		"plan_id": {Type: "string", Description: "Filter tasks by plan ID (optional)"},
	},
}

func TaskList() inventory.ServerTool {
	return inventory.ServerTool{
		Tool: &mcp.Tool{
			Name:        "task_list",
			Description: "List tasks, optionally filtered by plan",
			InputSchema: taskListInputSchema,
		},
		HandlerFunc: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				PlanID string `json:"plan_id"`
			}
			json.Unmarshal(req.Params.Arguments, &args)

			var planID uint
			if args.PlanID != "" {
				id, err := strconv.ParseUint(args.PlanID, 10, 64)
				if err != nil {
					return nil, err
				}
				planID = uint(id)
			}

			tasks := taskcore.ListTasks(planID)
			b, _ := json.MarshalIndent(tasks, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			}, nil
		},
	}
}
