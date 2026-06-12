package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	openaimodel "github.com/cloudwego/eino-ext/components/model/openai"
	mcptool "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func loadEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

func main() {
	loadEnv(".env")
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "DEEPSEEK_API_KEY not set")
		os.Exit(1)
	}
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	modelName := os.Getenv("DEEPSEEK_MODEL")
	if modelName == "" {
		modelName = "deepseek-chat"
	}
	serverCmd := os.Getenv("NAUTIKIT_SERVER_CMD")
	if serverCmd == "" {
		serverCmd = "go"
	}
	serverArgs := strings.Fields(os.Getenv("NAUTIKIT_SERVER_ARGS"))
	if len(serverArgs) == 0 {
		serverArgs = []string{"run", "./cmd/nautikit/"}
	}

	ctx := context.Background()

	mcpCli, err := client.NewStdioMCPClient(serverCmd, os.Environ(), serverArgs...)
	if err != nil {
		panic(fmt.Errorf("create MCP client: %w", err))
	}
	defer mcpCli.Close()

	_, err = mcpCli.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "nautikit-agent",
				Version: "0.1.0",
			},
		},
	})
	if err != nil {
		panic(fmt.Errorf("initialize MCP: %w", err))
	}

	tools, err := mcptool.GetTools(ctx, &mcptool.Config{Cli: mcpCli})
	if err != nil {
		panic(fmt.Errorf("get MCP tools: %w", err))
	}

	chatModel, err := openaimodel.NewChatModel(ctx, &openaimodel.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
		Timeout: 60 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
		MaxStep: 12,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("agent (model: %s, tools from MCP server)\n", modelName)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") {
			break
		}

		msg, err := agent.Generate(ctx, []*schema.Message{schema.UserMessage(line)})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Println(msg.Content)
	}
}
