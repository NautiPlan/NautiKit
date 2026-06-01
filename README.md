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
│   └── taskcore/
│       ├── models.go                 # Task、Plan 结构体
│       ├── store.go                  # SQLite 持久化（GORM）
│       └── tools/
│           ├── echo.go               # echo 回显工具
│           ├── task.go               # task CRUD 工具
│           └── plan.go               # plan CRUD 工具
├── skills/
│   └── plan-generation.md            # 计划生成 Skill 定义
├── demo/agent/                       # Demo Agent（独立 CLI）
├── build/                            # 构建输出
├── .mcp.json                         # MCP Server 配置（Claude Code 用）
├── DigimonGPT-guide.md               # 可演进 Agent 记忆系统架构指南
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

## Skill

Skill 定义在 `skills/` 目录，描述如何编排 MCP 工具完成复杂任务：

| Skill               | 说明                                                              |
| ------------------- | ----------------------------------------------------------------- |
| plan-generation     | 自然语言 → Plan → Task，拆解、分配日期和优先级（由 LLM 实时决定） |
| knowledge-ingestion | （规划中）校验浓缩 → 向量入库                                     |
| memory-recall       | （规划中）复杂度感知的动态检索                                    |
| pattern-abstraction | （规划中）从历史任务中提炼通用决策模式                            |

使用方式：复制到 `.claude/skills/<名称>/SKILL.md`，重启 Claude Code 后 `/skills` 可见。

## 存储

- SQLite 持久化（GORM），数据库文件：`~/.nautikit/data.db`
- `plan_delete` 在事务中先删任务再删计划，保证数据一致性

## 构建与运行

```bash
# 构建
go build -o build/nautikit ./cmd/nautikit/

# 运行 MCP Server（stdio 模式）
./build/nautikit

```

## 规划模块

### MCP 1 · 任务核心

- [x] 任务 CRUD（create / list / update / delete）
- [x] 计划 CRUD（create / list / get / delete），删除计划级联删任务
- [x] SQLite 持久化
- [x] 计划生成 Skill（自然语言 → Plan → Task，含 LLM 实时确定优先级）
- [ ] 动态优先级计算（公式化算法，如 urgency + importance + recency_bonus，暂不列入当前迭代）

### MCP 2 · 知识库（RAG）

参考 [DigimonGPT-guide.md](DigimonGPT-guide.md) 的架构设计，MCP 工具提供基础操作，Skill 负责智能编排。

**MCP 工具（4 个）：**

| 工具        | 描述                                                                    |
| ----------- | ----------------------------------------------------------------------- |
| `kb_ingest` | 文档入库：分块、向量化、存储，支持 metadata 标签                        |
| `kb_search` | 向量检索：cosine similarity 排序，可配 k 值、相似度阈值和 metadata 过滤 |
| `kb_delete` | 删除指定文档                                                            |
| `kb_list`   | 列出文档（按 metadata 过滤）                                            |

**配套 Skill（3 个）：**

| Skill               | 编排逻辑                                            |
| ------------------- | --------------------------------------------------- |
| knowledge-ingestion | 校验结果质量 → 浓缩摘要 → `kb_ingest`               |
| memory-recall       | 复杂度评估 → 动态 k → `kb_search`（替代固定 top-N） |
| pattern-abstraction | 攒够 N 个同类任务 → LLM 抽象 pattern → `kb_ingest`  |

**落地顺序：** knowledge-ingestion → memory-recall → pattern-abstraction

### MCP 3 · 信息检索

- [ ] 通用 Web 搜索
- [ ] arXiv 论文检索
- [ ] GitHub 仓库搜索

### MCP 4 · GitHub 存储

- [ ] Git 版本化存储
- [ ] 文件类型过滤
