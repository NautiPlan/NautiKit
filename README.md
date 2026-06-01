# NautiKit

> 面向 AI 任务规划的 MCP 工具集 — 即插即用，兼容任意 MCP Agent。

NautiKit 是 [NautiPlan](https://github.com/NautiPlan/NautiPlan) 的工具层重构，将任务管理、知识检索、信息搜索等能力封装为标准 MCP Server。

## 在 Claude Code 中使用

### 1. 配置 MCP Server

项目根目录创建 `.mcp.json`：

```json
{
  "mcpServers": {
    "nautikit": {
      "command": "./build/nautikit"
    }
  }
}
```

重启 Claude Code，输入 `/mcp` 确认 `nautikit` 已连接。

### 2. 安装 Skill

```bash
mkdir -p .claude/skills/plan-generation
cp skills/plan-generation.md .claude/skills/plan-generation/SKILL.md
```

重启后 `/skills` 即可看到 `plan-generation`，输入 `/plan-generation` 调用。

---

## 项目结构

```
NautiKit/
├── cmd/nautikit/main.go              # MCP Server 入口，stdio 模式
├── pkg/
│   ├── inventory/
│   │   ├── server_tool.go            # ServerTool 类型（Tool + HandlerFunc）
│   │   └── registry.go               # Inventory（Add / RegisterAll）
│   ├── taskcore/
│   │   ├── models.go                 # Task、Plan 结构体
│   │   ├── store.go                  # SQLite 持久化（GORM）
│   │   └── tools/
│   │       ├── echo.go               # echo 回显工具
│   │       ├── task.go               # task CRUD 工具
│   │       └── plan.go               # plan CRUD 工具
│   └── kbcore/
│       ├── models.go                 # Document 结构体
│       ├── store.go                  # 向量存储与检索
│       ├── embedder.go               # Embedder 接口 + OpenAI 兼容实现
│       └── tools/
│           ├── kb_ingest.go          # kb_ingest 工具
│           └── kb_search.go          # kb_search 工具
├── skills/
│   └── plan-generation.md            # 计划生成 Skill 定义
├── demo/agent/                       # Demo Agent（独立 CLI）
├── build/                            # 构建输出
├── Makefile
├── go.mod / go.sum
└── README.md
```

## MCP 工具

通过 stdio 传输，兼容任意 MCP Agent：

| 工具          | 参数                                                            | 描述                     |
| ------------- | --------------------------------------------------------------- | ------------------------ |
| `echo`        | `message` (必填)                                                | 回显测试，验证 MCP 链路  |
| `task_create` | `title` (必填), `plan_id`, `description`, `date`, `priority`    | 创建任务                 |
| `task_list`   | `plan_id` (选填)                                                | 列出任务，可按计划过滤   |
| `task_update` | `id` (必填), `title`, `description`, `date`, `priority`, `done` | 更新任务，只更新传入字段 |
| `task_delete` | `id` (必填)                                                     | 删除任务                 |
| `plan_create` | `title` (必填), `description`                                   | 创建计划                 |
| `plan_list`   | 无                                                              | 列出所有计划             |
| `plan_get`    | `id` (必填)                                                     | 查看单个计划及其所有任务 |
| `plan_delete` | `id` (必填)                                                     | 删除计划及其中所有任务   |
| `kb_ingest`   | `content` (必填), `metadata` (选填, JSON)                       | 文档入库：向量化、存储   |
| `kb_search`   | `query` (必填), `k` (默认5), `threshold` (默认0.5), `filter`    | 向量检索：cosine + 过滤  |

## Skill

Skill 定义在 `skills/` 目录，描述如何编排 MCP 工具完成复杂任务：

| Skill               | 说明                                                              |
| ------------------- | ----------------------------------------------------------------- |
| plan-generation     | 自然语言 → Plan → Task，拆解、分配日期和优先级（由 LLM 实时决定） |
| knowledge-ingestion | （规划中）校验浓缩 → 向量入库                                     |
| memory-recall       | （规划中）复杂度感知的动态检索                                    |
| memory-management   | （规划中）记忆淘汰，评估重要性与时效性，防止记忆膨胀              |
| pattern-abstraction | （规划中）从历史任务中提炼通用决策模式                            |

使用方式：复制到 `.claude/skills/<名称>/SKILL.md`，重启 Claude Code 后 `/skills` 可见。

## 环境变量

| 变量                          | 用途               | 默认值                                 |
| ----------------------------- | ------------------ | -------------------------------------- |
| `NAUTIKIT_EMBEDDING_BASE_URL` | Embedding API 地址 | `https://api.openai.com/v1/embeddings` |
| `NAUTIKIT_EMBEDDING_MODEL`    | Embedding 模型名   | `text-embedding-3-small`               |
| `NAUTIKIT_EMBEDDING_API_KEY`  | Embedding API 密钥 | 无（必填）                             |

## 构建与运行

```bash
# 构建
go build -o build/nautikit ./cmd/nautikit/

# 运行 MCP Server（stdio 模式）
./build/nautikit

```
