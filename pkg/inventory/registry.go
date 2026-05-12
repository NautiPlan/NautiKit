package inventory

import "github.com/modelcontextprotocol/go-sdk/mcp"

type Inventory struct {
	tools []ServerTool
}

func New() *Inventory {
	return &Inventory{}
}

func (inv *Inventory) Add(st ServerTool) {
	inv.tools = append(inv.tools, st)
}

func (inv *Inventory) All() []ServerTool {
	return inv.tools
}

func (inv *Inventory) RegisterAll(server *mcp.Server) {
	for _, st := range inv.tools {
		server.AddTool(st.Tool, st.HandlerFunc)
	}
}
