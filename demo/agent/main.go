package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	loadConfig()
	if os.Getenv("ANTHROPIC_API_KEY") == "" && os.Getenv("ANTHROPIC_AUTH_TOKEN") == "" {
		fmt.Fprintln(os.Stderr, "未找到 API 凭据。请设置以下任一方式：")
		fmt.Fprintln(os.Stderr, "  1. 环境变量: export ANTHROPIC_API_KEY=sk-ant-...")
		fmt.Fprintln(os.Stderr, "  2. 配置文件: ~/.nautikit/config")
		fmt.Fprintln(os.Stderr, "  3. 项目配置: ./nautikit-config")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "配置文件格式：")
		fmt.Fprintln(os.Stderr, "  ANTHROPIC_API_KEY=sk-ant-...")
		fmt.Fprintln(os.Stderr, "  ANTHROPIC_BASE_URL=https://api.anthropic.com   # 可选，自定义端点")
		os.Exit(1)
	}

	model := resolveModel(os.Getenv("NAUTIKIT_MODEL"))
	fmt.Printf("model: %s\n", model)

	// 1. start MCP server
	bin := findBinary()
	fmt.Printf("mcp: connected to %s\n\n", filepath.Base(bin))

	client := mcp.NewClient(&mcp.Implementation{Name: "nautikit-agent", Version: "0.1.0"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.Command(bin)}, nil)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	// 2. discover tools
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	tools, err := session.ListTools(ctx2, &mcp.ListToolsParams{})
	cancel2()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list tools: %v\n", err)
		os.Exit(1)
	}

	anthropicTools := make([]anthropic.ToolUnionParam, 0, len(tools.Tools))
	for _, t := range tools.Tools {
		at := convertTool(t)
		anthropicTools = append(anthropicTools, at)
	}
	fmt.Printf("tools: %d loaded\n\n", len(anthropicTools))

	// 3. LLM client
	llm := anthropic.NewClient()
	sysPrompt := "你是一个任务管理助手。你可以使用提供的工具来帮用户创建和查看任务。请用中文回复。"

	history := []anthropic.MessageParam{}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("输入任务指令（如 '创建任务：买牛奶，高优先级'），输入 /quit 退出\n")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "/quit" {
			break
		}

		history = append(history, anthropic.NewUserMessage(
			anthropic.NewTextBlock(input),
		))

		// 4. call LLM
		ctxLLM, cancelLLM := context.WithTimeout(context.Background(), 60*time.Second)
		msg, err := llm.Messages.New(ctxLLM, anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: 1024,
			System: []anthropic.TextBlockParam{
				{Text: sysPrompt},
			},
			Messages: history,
			Tools:    anthropicTools,
		})
		cancelLLM()
		if err != nil {
			fmt.Fprintf(os.Stderr, "llm error: %v\n", err)
			continue
		}

		// 5. process response blocks
		toolCallsMade := false

		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				fmt.Printf("\n%s\n", block.Text)

			case "tool_use":
				toolCallsMade = true
				var args map[string]any
				json.Unmarshal(block.Input, &args)

				fmt.Printf("\n🔧 %s(%v)\n", block.Name, prettyArgs(args))

				ctxTool, cancelTool := context.WithTimeout(context.Background(), 5*time.Second)
				result, err := session.CallTool(ctxTool, &mcp.CallToolParams{
					Name:      block.Name,
					Arguments: args,
				})
				cancelTool()

				var resultText string
				if err != nil {
					resultText = fmt.Sprintf("error: %v", err)
				} else if len(result.Content) > 0 {
					if tc, ok := result.Content[0].(*mcp.TextContent); ok {
						resultText = tc.Text
					}
				}
				fmt.Printf("   → %s\n", strings.ReplaceAll(resultText, "\n", "\n   "))

				history = append(history, anthropic.MessageParam{
					Role: anthropic.MessageParamRoleAssistant,
					Content: []anthropic.ContentBlockParamUnion{
						anthropic.NewToolUseBlock(block.ID, args, block.Name),
					},
				})
				history = append(history, anthropic.NewUserMessage(
					anthropic.NewToolResultBlock(block.ID, resultText, false),
				))
			}
		}

		if !toolCallsMade {
			blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.Content))
			for _, block := range msg.Content {
				if block.Type == "text" {
					blocks = append(blocks, anthropic.NewTextBlock(block.Text))
				}
			}
			if len(blocks) > 0 {
				history = append(history, anthropic.MessageParam{
					Role:    anthropic.MessageParamRoleAssistant,
					Content: blocks,
				})
			}
		}

		if len(history) > 20 {
			history = history[len(history)-20:]
		}
		fmt.Println()
	}
}

// ---------- helpers ----------

// loadConfig reads Anthropic settings from config files and sets them as env vars.
// Env vars already set take precedence over file values.
// Config file format: KEY=VALUE, one per line. Lines starting with # are comments.
//
// Supported keys (all optional):
//
//	ANTHROPIC_API_KEY     - API key (sk-ant-...)
//	ANTHROPIC_AUTH_TOKEN  - Auth token (alternative to API key)
//	ANTHROPIC_BASE_URL    - Custom API endpoint (default: https://api.anthropic.com)
//	NAUTIKIT_MODEL        - Model: sonnet (default), opus, haiku, or full ID
func loadConfig() {
	// env vars already set → skip file loading for those keys
	envKeys := map[string]bool{}
	for _, k := range []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL", "NAUTIKIT_MODEL"} {
		if os.Getenv(k) != "" {
			envKeys[k] = true
		}
	}

	// read from config files (last wins)
	var files []string
	if home, err := os.UserHomeDir(); err == nil {
		files = append(files, filepath.Join(home, ".nautikit", "config"))
	}
	files = append(files, "nautikit-config")

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if !envKeys[key] && value != "" {
				os.Setenv(key, value)
			}
		}
	}

	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		if baseURL != "" {
			fmt.Printf("api: %s (endpoint: %s)\n", maskKey(apiKey), baseURL)
		} else {
			fmt.Printf("api: %s\n", maskKey(apiKey))
		}
	}
}

// resolveModel maps short names to Anthropic model IDs.
// Supports: sonnet, opus, haiku (case-insensitive).
// Full model IDs like "claude-sonnet-4-6" pass through unchanged.
// Default (empty string) → "claude-sonnet-4-6".
func resolveModel(name string) anthropic.Model {
	switch strings.ToLower(name) {
	case "opus":
		return anthropic.ModelClaudeOpus4_7
	case "haiku":
		return anthropic.ModelClaudeHaiku4_5
	case "sonnet", "":
		return anthropic.ModelClaudeSonnet4_6
	default:
		return anthropic.Model(name)
	}
}

func maskKey(k string) string {
	if len(k) <= 15 {
		return "***"
	}
	return k[:10] + "..." + k[len(k)-5:]
}

func convertTool(t *mcp.Tool) anthropic.ToolUnionParam {
	var props any
	var required []string

	if schema, ok := t.InputSchema.(map[string]any); ok {
		if p, ok := schema["properties"]; ok {
			props = p
		}
		if r, ok := schema["required"].([]any); ok {
			required = make([]string, len(r))
			for i, v := range r {
				required[i] = fmt.Sprint(v)
			}
		}
	}

	return anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        t.Name,
			Description: anthropic.String(t.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type:       constant.Object("object"),
				Properties: props,
				Required:   required,
			},
		},
	}
}

func prettyArgs(args map[string]any) string {
	parts := make([]string, 0, len(args))
	for k, v := range args {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, ", ")
}

func findBinary() string {
	candidates := []string{
		"./build/nautikit",
		"../build/nautikit",
		"../../build/nautikit",
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return "nautikit"
}
