# NautiKit

> 面向 AI 任务规划的 MCP 工具集 — 即插即用，兼容任意 MCP Agent。

NautiKit 是 [NautiPlan](https://github.com/NautiPlan/NautiPlan) 的工具层重构，将任务管理、知识检索等能力封装为标准 MCP Server。

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
│   │   ├── store.go                  # SQLite 持久化（GORM） + DB() 共享入口
│   │   └── tools/
│   │       ├── echo.go               # echo 回显工具
│   │       ├── task.go               # task CRUD 工具
│   │       └── plan.go               # plan CRUD 工具
│   └── kbcore/
│       ├── models.go                 # Document + Chunk 结构体
│       ├── store.go                  # 向量存储与检索（全量 cosine）
│       ├── embedder.go               # Eino OpenAI Embedder 初始化
│       ├── splitter.go               # Eino Recursive Splitter 初始化
│       └── tools/
│           ├── kb_ingest.go          # kb_ingest 工具
│           ├── kb_search.go          # kb_search 工具
│           └── kb_manage.go          # kb_list / kb_delete / kb_clear 工具
├── skills/
│   └── plan-generation.md            # 计划生成 Skill 定义
├── build/                            # 构建输出
├── Makefile
├── go.mod / go.sum
└── README.md
```

## MCP 工具

| 工具          | 参数                                                            | 描述                                     |
| ------------- | --------------------------------------------------------------- | ---------------------------------------- |
| `echo`        | `message` (必填)                                                | 回显测试，验证 MCP 链路                  |
| `task_create` | `title` (必填), `plan_id`, `description`, `date`, `priority`    | 创建任务                                 |
| `task_list`   | `plan_id` (选填)                                                | 列出任务，可按计划过滤                   |
| `task_update` | `id` (必填), `title`, `description`, `date`, `priority`, `done` | 更新任务，只更新传入字段                 |
| `task_delete` | `id` (必填)                                                     | 删除任务                                 |
| `plan_create` | `title` (必填), `description`                                   | 创建计划                                 |
| `plan_list`   | 无                                                              | 列出所有计划                             |
| `plan_get`    | `id` (必填)                                                     | 查看单个计划及其所有任务                 |
| `plan_delete` | `id` (必填)                                                     | 删除计划及其中所有任务                   |
| `kb_ingest`   | `content` (必填), `metadata` (选填)                             | 文档入库：分块 → 向量化 → 存储           |
| `kb_search`   | `query` (必填), `k` (默认5), `threshold` (默认0.5), `filter`    | 向量检索：embed → cosine 全量扫描 → 过滤 |
| `kb_list`     | `filter` (选填)                                                 | 列出所有文档，可按 metadata 过滤         |
| `kb_delete`   | `id` (必填)                                                     | 删除单个文档及其所有 chunks              |
| `kb_clear`    | 无                                                              | 清空知识库（不可逆）                     |

## Skill

| Skill           | 说明                                                              |
| --------------- | ----------------------------------------------------------------- |
| plan-generation | 自然语言 → Plan → Task，拆解、分配日期和优先级（由 LLM 实时决定） |

使用方式：复制到 `.claude/skills/<名称>/SKILL.md`，重启 Claude Code 后 `/skills` 可见。

## 规划中

| 模块                | 说明                                                     |
| ------------------- | -------------------------------------------------------- |
| 信息检索 (MCP 3)    | 通用 Web 搜索、arXiv 论文、GitHub 开源项目等垂直领域检索 |
| GitHub 存储 (MCP 4) | 大文件与隐私文件分级处理,作为知识库的独立存储后端        |

## 环境变量

| 变量                          | 用途                   | 默认值                      |
| ----------------------------- | ---------------------- | --------------------------- |
| `NAUTIKIT_EMBEDDING_API_KEY`  | Embedding API 密钥     | 无（必填）                  |
| `NAUTIKIT_EMBEDDING_BASE_URL` | Embedding API 地址     | `https://api.openai.com/v1` |
| `NAUTIKIT_EMBEDDING_MODEL`    | Embedding 模型名       | `text-embedding-3-small`    |
| `NAUTIKIT_CHUNK_SIZE`         | 分块字符数上限（rune） | `512`                       |
| `NAUTIKIT_CHUNK_OVERLAP`      | 块间重叠字符数         | `64`                        |

## 依赖

| 依赖                                                     | 用途                 | 许可证     |
| -------------------------------------------------------- | -------------------- | ---------- |
| [Eino](https://github.com/cloudwego/eino)                | Embedding + 文本分块 | Apache 2.0 |
| [go-sdk](https://github.com/modelcontextprotocol/go-sdk) | MCP 协议实现         | MIT        |
| [GORM](https://gorm.io)                                  | SQLite ORM           | MIT        |

## 构建与运行

```bash
# 构建
go build -o build/nautikit ./cmd/nautikit/

# 运行 MCP Server（stdio 模式）
./build/nautikit
```
