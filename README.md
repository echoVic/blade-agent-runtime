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

## 安装

### 一键安装（推荐）

```bash
curl -fsSL https://echovic.github.io/blade-agent-runtime/install.sh | sh
```

### 使用 Go 安装

```bash
go install github.com/echoVic/blade-agent-runtime/cmd/bar@latest
```

### 自定义安装

```bash
# 指定安装目录
BAR_INSTALL_DIR=/usr/local/bin curl -fsSL https://echovic.github.io/blade-agent-runtime/install.sh | sh

# 指定版本
BAR_VERSION=v0.0.1 curl -fsSL https://echovic.github.io/blade-agent-runtime/install.sh | sh
```

## 快速开始

```bash
cd your-repo

# 创建任务（自动初始化，无需 bar init）
bar task start fix-null-pointer

# 方式 1：包装交互式 agent（推荐）
bar wrap -- claude           # Claude Code
bar wrap -- aider            # Aider
bar wrap -- cursor .         # Cursor

# 方式 2：运行一次性命令
bar run -- npm install lodash
bar run -- sh -c 'echo "hello" > test.txt'

# 查看 agent 做了什么
bar diff

# 满意就应用到主分支
bar apply --message "feat: fixed by AI"

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
| `bar task switch <id>` | 切换到指定任务 |
| `bar task close` | 关闭当前任务并删除 worktree |
| `bar task close --keep` | 关闭任务但保留 worktree |
| `bar task close --delete` | 关闭任务并删除所有记录 |
| `bar task close --force` | 强制关闭（忽略未提交更改） |
| `bar wrap -- <cmd>` | 包装交互式命令，自动启动 Web UI |
| `bar wrap --no-ui -- <cmd>` | 包装命令但不启动 Web UI |
| `bar run -- <cmd>` | 执行一次性命令并记录 |
| `bar diff` | 查看当前变更 |
| `bar diff --format json` | JSON 格式输出 |
| `bar apply --message "msg"` | 应用变更到主分支 |
| `bar rollback` | 回滚变更 |
| `bar status` | 查看当前状态 |
| `bar log` | 查看操作日志 |
| `bar log --format markdown` | Markdown 格式日志 |
| `bar update` | 更新到最新版本 |
| `bar update --check` | 检查是否有新版本 |
| `bar version` | 显示版本信息 |
| `bar ui` | 启动 Web UI 审计界面 |
| `bar ui -p 3000` | 指定端口启动 Web UI |

## Web UI
BAR 提供了一个 Web UI 用于审计任务和查看操作日志：

```bash
bar ui              # 启动 Web UI (默认端口 8080)
bar ui -p 3000      # 指定端口
bar ui --no-open    # 不自动打开浏览器
```

**主要功能 (v0.0.8)**：
- **现代化深色主题**：基于 Zinc 色系，Vercel/Linear 风格。
- **Split View (分栏视图)**：左侧 Timeline，右侧 Monaco Editor Diff 视图。
- **智能跳转**：首页自动跳转到当前活跃任务。
- **实时 Diff**：选中步骤即可查看代码变更，支持语法高亮。
- **WebSocket**：实时更新任务状态和日志。

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
