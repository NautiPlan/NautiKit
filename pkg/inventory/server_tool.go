package inventory

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerTool struct {
	Tool        *mcp.Tool
	HandlerFunc func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
