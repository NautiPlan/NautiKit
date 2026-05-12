# demo-agent 测试指南

## 前置条件

- Go 1.23+
- 以下任一 Anthropic API 凭据：
  - API Key（从 [console.anthropic.com](https://console.anthropic.com) 获取）
  - Auth Token

## 1. 配置

### 方式 A：环境变量

```bash
export ANTHROPIC_API_KEY=sk-ant-...
export ANTHROPIC_BASE_URL=https://api.anthropic.com    # 可选，默认即为此值
```

### 方式 B：用户级配置文件

```bash
mkdir -p ~/.nautikit
cat > ~/.nautikit/config << 'EOF'
# NautiKit Agent 配置
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_BASE_URL=https://api.anthropic.com
EOF
```

### 方式 C：项目级配置文件

在 `NautiKit/` 根目录创建 `nautikit-config`（已在 `.gitignore` 中）：

```bash
# 标准 Anthropic API
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_BASE_URL=https://api.anthropic.com

# 或者使用代理 / 自定义端点
# ANTHROPIC_API_KEY=sk-ant-...
# ANTHROPIC_BASE_URL=https://your-proxy.example.com
```

### 配置项说明

| 键 | 必需 | 默认值 | 说明 |
|----|------|--------|------|
| `ANTHROPIC_API_KEY` | 是* | - | API Key，以 `sk-ant-` 开头 |
| `ANTHROPIC_AUTH_TOKEN` | 是* | - | Auth Token，与 API Key 二选一 |
| `ANTHROPIC_BASE_URL` | 否 | `https://api.anthropic.com` | 自定义 API 端点 |
| `NAUTIKIT_MODEL` | 否 | `sonnet` | 模型选择（见下方） |

> \* API Key 和 Auth Token 至少提供一个

### 优先级

**环境变量 > ~/.nautikit/config > ./nautikit-config**

环境变量中已设置的键，配置文件中的同名键会被忽略。

### 模型选择

`NAUTIKIT_MODEL` 支持简写和完整 ID：

| 简写 | 完整模型 ID |
|------|-----------|
| `sonnet`（默认） | `claude-sonnet-4-6` |
| `opus` | `claude-opus-4-7` |
| `haiku` | `claude-haiku-4-5` |
| 任意字符串 | 直接作为 model ID 传递（如 `claude-sonnet-4-5-20250929`） |

启动时会打印当前使用的模型：

```
model: claude-sonnet-4-6
api: sk-ant-api...xxxxx (endpoint: https://api.anthropic.com)
mcp: connected to nautikit
tools: 3 loaded
```

## 2. 构建

```bash
cd /path/to/NautiKit

go build -o build/nautikit ./cmd/nautikit/
go build -o demo/agent/demo-agent ./demo/agent/
```

## 3. 运行

```bash
./demo/agent/demo-agent
```

启动时会显示连接的 API 端点：

```
api: sk-ant-api...xxxxx (endpoint: https://api.anthropic.com)
mcp: connected to nautikit
tools: 3 loaded
```

## 4. 测试用例

### 用例 1：基础任务创建

```
> 创建三个任务：买牛奶（高优先级）、写周报（中优先级）、看书（低优先级）
```

**预期行为**：
- LLM 调用 `task_create` 三次
- 控制台输出 `🔧 task_create(title=xxx, priority=xxx)`
- 返回三个 task-1、task-2、task-3 的 JSON
- LLM 最终用中文总结创建结果

### 用例 2：查看任务列表

```
> 帮我看看现在有哪些任务
```

**预期行为**：
- LLM 调用 `task_list`
- 返回所有已创建任务的 JSON 数组

### 用例 3：混合操作（创建 + 查询）

```
> 新建任务：修复登录 bug（高优先级），然后告诉我所有高优先级任务有哪些
```

**预期行为**：
- `task_create` 创建新任务
- `task_list` 列出所有任务
- LLM 从结果中筛选高优先级任务并回复

### 用例 4：echo 测试

```
> 跟我说"Hello World"
```

**预期行为**：
- LLM 可能调用 `echo` 工具
- 返回对应的消息

### 用例 5：闲聊（不触发工具）

```
> 你好，介绍一下你自己
```

**预期行为**：
- LLM 直接以文本回复，不调用任何工具

## 5. 验证要点

| 检查项 | 如何验证 |
|--------|----------|
| MCP 连接 | 启动时显示 `tools: 3 loaded` |
| 工具发现 | LLM 能正确选择 `task_create` / `task_list` / `echo` |
| 工具调用 | 控制台出现 `🔧 tool_name(args)` 和 `→ result` |
| 上下文保持 | 创建任务后马上查询，能看到刚创建的任务 |
| 错误处理 | 输入无意义指令时 LLM 能优雅回复 |
| 自定义端点 | 配置 `ANTHROPIC_BASE_URL` 后看到对应的 endpoint 信息 |

## 6. 调试技巧

查看 MCP Server 日志（stderr）：
```bash
./demo/agent/demo-agent 2> server.log &
tail -f server.log
```

手动测试 MCP Server（不经过 LLM）：
```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}\n' | ./build/nautikit
```
