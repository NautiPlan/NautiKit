package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/kbcore"
	kbtools "github.com/NautiKit/NautiKit/pkg/kbcore/tools"
	"github.com/NautiKit/NautiKit/pkg/taskcore"
	"github.com/NautiKit/NautiKit/pkg/taskcore/tools"
)

func main() {

	if err := taskcore.Init(""); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer taskcore.Close()

	if err := kbcore.Init(""); err != nil {
		log.Fatalf("初始化知识库失败: %v", err)
	}
	defer kbcore.Close()

	inv := inventory.New()
	inv.Add(tools.Echo())
	inv.Add(tools.TaskCreate())
	inv.Add(tools.TaskList())
	inv.Add(tools.TaskUpdate())
	inv.Add(tools.TaskDelete())
	inv.Add(tools.PlanCreate())
	inv.Add(tools.PlanList())
	inv.Add(tools.PlanGet())
	inv.Add(tools.PlanDelete())
	inv.Add(kbtools.KBIngest())
	inv.Add(kbtools.KBSearch())
	inv.Add(kbtools.KBList())
	inv.Add(kbtools.KBDelete())
	inv.Add(kbtools.KBClear())

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "NautiKit",
		Version: "0.1.0",
	}, nil)
	inv.RegisterAll(server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	log.SetOutput(os.Stderr)
	log.Println("NautiKit MCP Server starting (stdio mode)")

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
