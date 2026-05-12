package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/NautiKit/NautiKit/pkg/inventory"
	"github.com/NautiKit/NautiKit/pkg/taskcore"
)

func main() {
	inv := inventory.New()
	inv.Add(taskcore.Echo())
	inv.Add(taskcore.TaskCreate())
	inv.Add(taskcore.TaskList())

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
