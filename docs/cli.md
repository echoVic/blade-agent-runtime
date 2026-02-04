# CLI 命令设计

## 命令概览

```bash
bar <command> [subcommand] [flags] [args]
```

| 命令 | 说明 | v0 状态 |
|------|------|---------|
| `bar init` | 初始化 BAR | ✅ |
| `bar task start` | 创建新任务 | ✅ |
| `bar task list` | 列出所有任务 | ✅ |
| `bar task switch` | 切换当前任务 | ✅ |
| `bar task close` | 关闭任务 | ✅ |
| `bar run` | 执行命令 | ✅ |
| `bar diff` | 查看变更 | ✅ |
| `bar apply` | 应用变更 | ✅ |
| `bar rollback` | 回滚变更 | ✅ |
| `bar status` | 查看状态 | ✅ |
| `bar log` | 查看日志 | ✅ |
| `bar policy check` | 检查策略 | v0.2 |
| `bar pr` | 生成 PR | v0.2 |

---

## 详细命令说明

### `bar init`

在当前 git 仓库中初始化 BAR。

```bash
bar init [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--force` | 强制重新初始化 | false |

**行为:**
1. 检查当前目录是否是 git 仓库
2. 在 `~/.bar/projects/<project>-<hash4>/` 创建目录结构
3. 创建默认配置文件

**示例:**
```bash
cd my-project
bar init
# Output: Initialized BAR in ~/.bar/projects/my-project-a3f2/
```

**错误情况:**
- 不在 git 仓库中：`Error: not a git repository`
- 已初始化：`Error: BAR already initialized (use --force to reinitialize)`

---

### `bar task start`

创建一个新的任务。

```bash
bar task start <name> [flags]
```

**Args:**
| Arg | 说明 | 必填 |
|-----|------|------|
| `name` | 任务名称 | ✅ |

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--base` | 基准分支/commit | 当前 HEAD |
| `--no-switch` | 创建后不切换到该任务 | false |

> **设计决策**：`--base` 默认使用当前 HEAD，最符合用户预期（用户通常在想要的分支上执行命令）。

**行为:**
1. 生成唯一的 task ID（nanoid，8 字符）
2. 创建 git worktree
3. 创建分支：`bar/<name>-<short_id>`
4. 初始化 task.json 和 ledger.jsonl
5. 设置为当前 active task

**示例:**
```bash
bar task start fix-null-pointer
# Output:
# Created task: fix-null-pointer (id: abc123)
# Branch: bar/fix-null-pointer-abc123
# Switched to task: fix-null-pointer

bar task start experiment --base develop --no-switch
# Output:
# Created task: experiment (id: def456)
# Branch: bar/experiment-def456
```

---

### `bar task list`

列出所有任务。

```bash
bar task list [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--all` | 显示已关闭的任务 | false |
| `--format` | 输出格式 (table/json) | table |

**示例:**
```bash
bar task list
# Output:
# ID       NAME              STATUS   CREATED              STEPS
# abc123   fix-null-pointer  active   2024-01-15 10:00:00  3
# def456   experiment        active   2024-01-15 11:00:00  1
# * = current task

bar task list --format json
# Output: [{"id":"abc123","name":"fix-null-pointer",...}]
```

---

### `bar task switch`

切换当前任务。

```bash
bar task switch <task_id|name>
```

**示例:**
```bash
bar task switch abc123
# Output: Switched to task: fix-null-pointer (abc123)

bar task switch fix-null-pointer
# Output: Switched to task: fix-null-pointer (abc123)
```

---

### `bar task close`

关闭任务（默认删除 worktree，保留记录）。

```bash
bar task close [task_id] [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--keep` | 保留 worktree 不删除 | false |
| `--delete` | 同时删除所有记录 | false |
| `--force` | 强制关闭（即使有未提交变更） | false |

> **设计决策**：默认删除 worktree 以节省空间，使用 `--keep` 保留以便调试。

**示例:**
```bash
bar task close
# Output: Closed task: fix-null-pointer (abc123)
# Worktree deleted: .bar/workspaces/abc123

bar task close --keep
# Output: Closed task: fix-null-pointer (abc123)
# Worktree kept: .bar/workspaces/abc123

bar task close abc123 --delete
# Output: Deleted task: fix-null-pointer (abc123)
```

---

### `bar run`

在当前任务的隔离区中执行命令。

```bash
bar run [flags] -- <command> [args...]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--task` | 指定任务（默认当前任务） | active task |
| `--timeout` | 超时时间 | 0 (无限) |
| `--no-record` | 不记录到 ledger | false |
| `--env` | 额外环境变量 | - |

**行为:**
1. 获取当前 active task
2. 检查 policy（如果启用）
3. 在 worktree 目录中执行命令
4. 捕获 stdout/stderr（透传 stdin/stdout，支持交互）
5. 生成 diff
6. 记录到 ledger

> **设计决策**：v0 采用透传 stdin/stdout 模式，用户可以与 agent 交互，但输出捕获可能不完整。

**示例:**
```bash
# 运行 Claude Code
bar run -- claude "fix the null pointer exception in main.go"

# 运行任意命令
bar run -- npm test

# 带超时
bar run --timeout 5m -- long-running-agent

# 传递环境变量
bar run --env API_KEY=xxx -- my-agent
```

**输出:**
```
Running: claude "fix the null pointer exception in main.go"
Workspace: .bar/workspaces/abc123

[Agent output here...]

Step 0004 completed (exit code: 0)
Duration: 45.2s
Files changed: 3 (+15, -5)
```

---

### `bar diff`

查看当前变更。

```bash
bar diff [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--step` | 查看特定 step 的 diff | latest |
| `--stat` | 只显示统计信息 | false |
| `--output` | 输出到文件 | - |
| `--format` | 输出格式 (patch/json) | patch |

**示例:**
```bash
# 查看当前 diff
bar diff

# 只看统计
bar diff --stat
# Output:
# 3 files changed, 15 insertions(+), 5 deletions(-)
#  main.go      | 10 +++++-----
#  utils.go     |  5 +++++
#  config.go    |  5 +++++

# 查看特定 step 的 diff
bar diff --step 0002

# 导出 patch
bar diff --output changes.patch
```

---

### `bar apply`

将变更应用到主分支。

```bash
bar apply [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--message` | Commit 消息 | 自动生成 |
| `--mode` | 应用模式 (commit/merge) | commit |
| `--no-close` | 应用后不关闭任务 | false |

**行为 (commit 模式):**
1. 在 worktree 分支上创建 commit
2. 切换到主分支
3. Cherry-pick commit
4. 关闭任务（除非 --no-close）

**示例:**
```bash
bar apply
# Output:
# Committed: abc1234 "fix: null pointer exception in main.go"
# Applied to: main
# Task closed: fix-null-pointer

bar apply --message "feat: add user authentication" --no-close
# Output:
# Committed: def5678 "feat: add user authentication"
# Applied to: main
# Task still active: fix-null-pointer
```

---

### `bar rollback`

回滚变更。

```bash
bar rollback [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--step` | 回滚到特定 step | - |
| `--base` | 回滚到初始状态 | false |
| `--hard` | 硬回滚（丢弃所有变更） | false |

**示例:**
```bash
# 回滚到初始状态
bar rollback --base
# Output: Rolled back to base state

# 回滚到特定 step
bar rollback --step 0002
# Output: Rolled back to step 0002

# 硬回滚（丢弃未记录的变更）
bar rollback --hard --base
# Output: Hard rolled back to base state (discarded uncommitted changes)
```

---

### `bar status`

查看当前状态。

```bash
bar status [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--format` | 输出格式 (text/json) | text |

**示例:**
```bash
bar status
# Output:
# BAR Status
# ──────────────────────────────
# Repository:  /path/to/my-project
# Active Task: fix-null-pointer (abc123)
# Workspace:   .bar/workspaces/abc123
# Branch:      bar/fix-null-pointer-abc123
# Base:        main
# Status:      dirty (3 files changed)
# Steps:       4
# Last Step:   0004 (run) - 2 minutes ago
```

---

### `bar log`

查看操作日志。

```bash
bar log [flags]
```

**Flags:**
| Flag | 说明 | 默认值 |
|------|------|--------|
| `--step` | 查看特定 step 详情 | - |
| `--limit` | 显示最近 N 条 | 10 |
| `--format` | 输出格式 (table/json/markdown) | table |
| `--output` | 输出到文件 | - |

**示例:**
```bash
bar log
# Output:
# STEP   KIND      COMMAND                          DURATION  EXIT  FILES
# 0004   run       claude "fix null pointer..."     45.2s     0     3 (+15, -5)
# 0003   run       npm test                         12.1s     1     0
# 0002   run       claude "add logging"             30.5s     0     2 (+20, -0)
# 0001   run       claude "initial analysis"        15.3s     0     0

bar log --step 0004
# Output:
# Step 0004
# ──────────────────────────────
# Kind:     run
# Command:  claude "fix the null pointer exception in main.go"
# Started:  2024-01-15 10:05:00
# Ended:    2024-01-15 10:05:45
# Duration: 45.2s
# Exit:     0
# Files:    3 (+15, -5)
# 
# Files Changed:
#   main.go      | 10 +++++-----
#   utils.go     |  5 +++++
#   config.go    |  5 +++++

bar log --format markdown --output report.md
# Output: Saved to report.md
```

---

## 全局 Flags

所有命令都支持以下全局 flags：

| Flag | 说明 | 默认值 |
|------|------|--------|
| `--help, -h` | 显示帮助 | - |
| `--version, -v` | 显示版本 | - |
| `--verbose` | 详细输出 | false |
| `--quiet, -q` | 静默模式 | false |
| `--config` | 指定配置文件 | .bar/config.yaml |

---

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2 | 参数错误 |
| 3 | 未初始化 |
| 4 | 无 active task |
| 5 | Policy 违规 |
| 126 | 命令执行失败 |
| 127 | 命令不存在 |

---

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `BAR_HOME` | BAR 数据目录 | ~/.bar |
| `BAR_CONFIG` | 配置文件路径 | ~/.bar/projects/<project>/config.yaml |
| `BAR_VERBOSE` | 详细输出 | false |
| `BAR_NO_COLOR` | 禁用颜色输出 | false |

---

## Shell 自动补全

BAR 支持 Bash、Zsh、Fish 和 PowerShell 的自动补全功能。

### 启用补全

**Zsh** (添加到 `~/.zshrc`)：
```bash
source <(bar completion zsh)
```

**Bash** (添加到 `~/.bashrc`)：
```bash
source <(bar completion bash)
```

**Fish**：
```fish
bar completion fish | source
```

**PowerShell**：
```powershell
bar completion powershell | Out-String | Invoke-Expression
```

### 支持的补全

| 命令 | 补全内容 |
|------|----------|
| `bar task switch <TAB>` | 任务 ID 和名称 |
| `bar task close <TAB>` | 任务 ID 和名称 |
| `bar log --step <TAB>` | Step ID |
| `bar diff --step <TAB>` | Step ID |
| `bar rollback --step <TAB>` | Step ID |

**示例:**
```bash
$ bar task switch <TAB>
abc123  -- fix-bug (active)
def456  -- add-feature (active)
fix-bug -- ID: abc123

$ bar log --step <TAB>
0001  -- run: echo hello
0002  -- run: npm test
0003  -- apply
```

---

## 错误提示

BAR 提供友好的错误提示，包含问题描述和解决建议：

```
❌ BAR is not initialized in this repository.

💡 Run 'bar init' or 'bar task start <name>' to get started.
```

### 常见错误

| 错误 | 说明 | 解决方法 |
|------|------|----------|
| `BAR is not initialized` | 当前仓库未初始化 BAR | 运行 `bar init` |
| `No active task` | 没有活动任务 | 运行 `bar task start <name>` |
| `Task not found` | 指定的任务不存在 | 运行 `bar task list` 查看 |
| `Step not found` | 指定的 Step 不存在 | 运行 `bar log` 查看 |
| `Workspace has uncommitted changes` | 工作区有未提交更改 | 使用 `--force` 或先提交 |
| `Command blocked by policy` | 命令被策略拦截 | 检查 policy 配置 |
| `Not a git repository` | 当前目录不是 Git 仓库 | 运行 `git init` |
