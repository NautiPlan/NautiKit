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
- [x] 计划生成 Skill（自然语言 → Plan → Task）
- [ ] 动态优先级计算
- [ ] 计划生成含知识库召回

### MCP 2 · 知识库（RAG）

- [ ] 文档入库、分块、向量化
- [ ] 混合检索（向量 + 全文）
- [ ] 上下文召回

### MCP 3 · 信息检索

- [ ] 通用 Web 搜索
- [ ] arXiv 论文检索
- [ ] GitHub 仓库搜索

### MCP 4 · GitHub 存储

- [ ] Git 版本化存储
- [ ] 文件类型过滤
