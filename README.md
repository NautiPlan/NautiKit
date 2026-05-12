# NautiKit

> 面向 AI 任务规划的 MCP 工具集与 Skill 定义 — 即插即用，兼容任意 Agent。

NautiKit 是 [NautiPlan](https://github.com/NautiPlan/NautiPlan) 的工具层重构，将任务管理、知识检索、信息搜索等能力封装为标准 MCP Server，并提供配套 Skill 定义。

## 模块

### MCP 1 · 任务核心

基于本地 SQLite 的任务与计划管理。

- 任务增删改查
- 自然语言 → 结构化任务生成
- 动态优先级计算与排序
- 生成计划时自动从知识库召回相关上下文

### MCP 2 · 知识库（RAG）

基于本地向量数据库的个人知识检索。

- 文档入库、更新、删除
- 本地向量检索
- 检索结果作为上下文传递给上游 Agent，参与计划制定

### MCP 3 · 信息检索

通用与垂直领域搜索。

- 通用 Web 搜索（无需专用 API）
- 垂直领域检索：arXiv 论文、GitHub 开源项目（可扩展）

### MCP 4 · GitHub 存储 _(附加模块)_

- 大文件与隐私文件分级处理
- 基于文件类型与敏感度的选择性推送/拉取
- 作为知识库的独立存储后端运行

## Skills

Skill 定义上游 Agent 应如何编排上述 MCP 工具，以提示词 + 工作流的形式提供。

| Skill      | 描述                                      |
| ---------- | ----------------------------------------- |
| 任务生成   | 知识库召回 + 可选搜索 → 输出结构化计划    |
| 优先级重排 | 基于更新的上下文对现有任务重新排序        |
| 知识入库   | 处理文档 → 分块 → 向量化 → 存储           |
| 领域调研   | arXiv / GitHub 检索 → 整理摘要 → 关联任务 |

## 当前实现

一期完成了最小 MCP 框架 + ReAct Agent 原型。

### 项目结构

```
NautiKit/
├── cmd/nautikit/main.go          # MCP Server 入口，stdio 模式
├── pkg/
│   ├── inventory/
│   │   ├── server_tool.go        # ServerTool 类型（Tool + Handler）
│   │   └── registry.go           # Inventory（Add / All / RegisterAll）
│   └── taskcore/
│       ├── models.go             # Task 结构体
│       ├── store.go              # 内存存储（sync.RWMutex）
│       └── tools.go              # echo, task_create, task_list
├── demo/
│   ├── agent/main.go             # ReAct Agent（LLM 决策 + MCP 工具调用）
│   └── TEST_GUIDE.md             # 测试指南
├── build/nautikit                # Server 二进制
├── Makefile
├── go.mod / go.sum
└── .gitignore
```

### MCP Server

实现了 3 个工具，通过 stdio 传输，兼容任意 MCP Agent：

| 工具 | 输入 | 描述 |
|------|------|------|
| `echo` | `message` (必填) | Echo 回显，验证链路 |
| `task_create` | `title` (必填), `priority` (选填) | 创建任务，内存存储 |
| `task_list` | 无 | 列出所有任务 |

### Demo Agent

基于 Anthropic API 的 ReAct Agent，通过 `CommandTransport` 启动 MCP Server 子进程，自动发现工具并让 LLM 决策调用。

```bash
# 配置 API Key（三选一）
export ANTHROPIC_API_KEY=sk-ant-...            # 环境变量
echo 'ANTHROPIC_API_KEY=sk-ant-...' > ~/.nautikit/config   # 用户配置
echo 'ANTHROPIC_API_KEY=sk-ant-...' > nautikit-config       # 项目配置

# 可选配置
export NAUTIKIT_MODEL=opus    # 模型: sonnet (默认), opus, haiku, 或完整 ID
export ANTHROPIC_BASE_URL=https://api.anthropic.com   # 自定义端点

# 构建并运行
go build -o build/nautikit ./cmd/nautikit/
go build -o demo/agent/demo-agent ./demo/agent/
./demo/agent/demo-agent
```

交互示例：

```
> 创建三个任务：买牛奶(高)、写周报(中)、看书(低)

🔧 task_create(title=买牛奶, priority=high)
   → {"id":"task-1","title":"买牛奶","priority":"high","done":false}
🔧 task_create(title=写周报, priority=medium)
   → {"id":"task-2","title":"写周报","priority":"medium","done":false}
🔧 task_create(title=看书, priority=low)
   → {"id":"task-3","title":"看书","priority":"low","done":false}

已为你创建了三个任务：买牛奶、写周报、看书
```

### 架构

```
User 输入 → LLM (Claude) 决策 → 调用 MCP 工具 → nautikit 子进程 → 返回结果 → LLM 继续推理
```

详细测试用例见 [demo/TEST_GUIDE.md](demo/TEST_GUIDE.md)。
