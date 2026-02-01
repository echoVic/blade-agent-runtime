# Blade Agent Runtime (BAR)

> A local execution runtime for AI agents: isolated workspace + step ledger + diff/rollback + resumable tasks.

**一句话价值**：让任何 agent（你自己的或第三方 CLI）在隔离区里干活，并且**每一步可审计、可回滚、可恢复**。

## 为什么需要 BAR？

当你使用 Claude Code、Cursor、Copilot 等 AI Coding Agent 时：

| 痛点 | BAR 的解决方案 |
|------|---------------|
| Agent 直接修改主 repo，改错了难以回滚 | 在 git worktree 隔离区执行，随时 rollback |
| 不知道 Agent 到底改了什么 | 每一步都有 diff 记录，完整审计日志 |
| 中断后无法恢复 | Task 状态持久化，支持 resume |
| 想试多种方案对比 | 多个 task 并行，互不干扰 |

## 快速开始

```bash
# 安装（Go 1.21+）
go install github.com/user/blade-agent-runtime/cmd/bar@latest

# 在你的项目中初始化
cd your-repo
bar init

# 创建一个任务（自动创建隔离的 worktree）
bar task start fix-null-pointer

# 在隔离区运行任何 AI agent
bar run -- claude "fix the null pointer exception in main.go"

# 查看 agent 做了什么
bar diff

# 满意就应用到主分支
bar apply

# 不满意就回滚
bar rollback
```

## 核心概念

### Task（任务）
一个独立的工作单元，包含：
- 隔离的 git worktree
- 独立的分支 `bar/<task-name>`
- 完整的操作日志（ledger）

### Step（步骤）
Task 中的每一个操作，包括：
- `run`: 执行外部命令/agent
- `apply`: 将变更应用到主分支
- `rollback`: 回滚到某个状态

### Ledger（账本）
记录所有 step 的日志，包含：
- 执行的命令
- 开始/结束时间
- diff 统计
- stdout/stderr

## CLI 命令

| 命令 | 说明 |
|------|------|
| `bar init` | 初始化 BAR（创建 `.bar/` 目录） |
| `bar task start <name>` | 创建新任务 |
| `bar task list` | 列出所有任务 |
| `bar run -- <cmd>` | 在当前任务的隔离区执行命令 |
| `bar diff` | 查看当前变更 |
| `bar apply` | 应用变更到主分支 |
| `bar rollback` | 回滚变更 |
| `bar status` | 查看当前状态 |
| `bar log` | 查看操作日志 |

## 目录结构

```
your-repo/
├── .bar/                    # BAR 数据目录
│   ├── tasks/
│   │   └── <task_id>/
│   │       ├── task.json    # 任务元信息
│   │       ├── ledger.jsonl # 操作日志
│   │       └── artifacts/   # diff/output 文件
│   └── workspaces/
│       └── <task_id>/       # git worktree
└── ... (your code)
```

## v0 非目标

为了保持专注，v0 **不做**：

- ❌ 多 agent 协作
- ❌ 云端 / SaaS / 账号体系
- ❌ UI（桌面端）
- ❌ 通用插件市场
- ❌ 深度 syscall 级沙箱
- ❌ Windows 完整支持（先 Mac/Linux）

## 文档

- [架构设计](docs/architecture.md)
- [CLI 详细说明](docs/cli.md)
- [数据模型](docs/data-model.md)
- [开发路线图](docs/roadmap.md)

## License

MIT
