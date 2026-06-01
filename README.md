# NautiKit

> 面向 AI 任务规划的 MCP 工具集 — 即插即用，兼容任意 MCP Agent。

NautiKit 是 [NautiPlan](https://github.com/NautiPlan/NautiPlan) 的工具层重构，将任务管理、知识检索、信息搜索等能力封装为标准 MCP Server。

## 当前实现

一期完成了最小 MCP 框架，包含 Plan + Task 数据模型与基础 CRUD 工具。

### 项目结构

```
NautiKit/
├── cmd/nautikit/main.go              # MCP Server 入口，stdio 模式
├── pkg/
│   ├── inventory/
│   │   ├── server_tool.go            # ServerTool 类型（Tool + HandlerFunc）
│   │   └── registry.go               # Inventory（Add / RegisterAll）
│   └── taskcore/
│       ├── models.go                 # Task、Plan 结构体
│       ├── store.go                  # 内存存储 + JSON 文件持久化
│       └── tools/
│           ├── echo.go               # echo 回显工具
│           ├── task.go               # task_create, task_list
│           └── plan.go               # plan_create, plan_list
├── build/                            # 构建输出
├── Makefile
├── go.mod / go.sum
└── README.md
```

### 数据模型

```
Plan ──1:N──> Task（Task 通过 plan_id 归属 Plan，date 字段表示安排在哪天）
```

```go
type Plan struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    CreatedAt   string `json:"created_at"`
}

type Task struct {
    ID          string `json:"id"`
    PlanID      string `json:"plan_id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Date        string `json:"date"`     // "2026-05-29"
    Priority    string `json:"priority"` // "high" | "medium" | "low"
    Done        bool   `json:"done"`
}
```

### MCP 工具

通过 stdio 传输，兼容任意 MCP Agent：

| 工具          | 参数                                                         | 描述                    |
| ------------- | ------------------------------------------------------------ | ----------------------- |
| `echo`        | `message` (必填)                                             | 回显测试，验证 MCP 链路 |
| `task_create` | `title` (必填), `plan_id`, `description`, `date`, `priority` | 创建任务                |
| `task_list`   | `plan_id` (选填)                                             | 列出任务，可按计划过滤  |
| `plan_create` | `title` (必填), `description`                                | 创建计划                |
| `plan_list`   | 无                                                           | 列出所有计划            |

### 存储

- 内存存储，通过 `sync.RWMutex` 保证并发安全
- JSON 文件持久化：`~/.nautikit/data.json`，每次写操作后同步落盘
- 后续可迁移至 SQLite

### 构建与运行

```bash
# 构建
go build -o build/nautikit ./cmd/nautikit/

# 运行 MCP Server（stdio 模式）
./build/nautikit

# 手动测试
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./build/nautikit
```

## 规划模块

### MCP 1 · 任务核心

- [x] 任务创建与查询
- [ ] 任务更新、删除
- [ ] 自然语言 → 结构化任务生成
- [ ] 动态优先级计算
- [ ] 计划生成（含知识库召回）

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
