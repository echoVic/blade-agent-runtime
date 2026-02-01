# 第三方 CLI Agent 集成指南

> 本文档描述如何使用 BAR 包装任意第三方 CLI Agent（如 Claude Code、Cursor CLI、Aider 等）。

## 基本用法

BAR 的设计目标是 **Agent 无关**，任何能在命令行运行的 agent 都可以通过 `bar run` 包装：

```bash
# Claude Code
bar run -- claude "fix the bug in main.go"

# Aider
bar run -- aider --message "add unit tests"

# Cursor CLI (如果有)
bar run -- cursor-cli "refactor this function"

# 甚至普通命令
bar run -- npm run build
bar run -- make test
```

## 工作原理

当你执行 `bar run -- <command>` 时，BAR 会：

1. **切换工作目录**：将 cwd 设置为当前 task 的 worktree
2. **注入环境变量**：设置 `BAR_*` 环境变量
3. **执行命令**：运行你指定的命令
4. **捕获输出**：记录 stdout/stderr
5. **生成 diff**：对比执行前后的文件变化
6. **记录到 ledger**：写入 step 日志

## 环境变量

BAR 会注入以下环境变量，agent 可以选择性使用：

| 变量 | 说明 | 示例 |
|------|------|------|
| `BAR_ACTIVE` | 是否在 BAR 环境中 | `true` |
| `BAR_TASK_ID` | 当前 task ID | `abc123` |
| `BAR_TASK_NAME` | 当前 task 名称 | `fix-null-pointer` |
| `BAR_WORKSPACE` | worktree 绝对路径 | `/path/to/.bar/workspaces/abc123` |
| `BAR_BASE_REF` | 基准分支 | `main` |
| `BAR_REPO_ROOT` | 原始 repo 根目录 | `/path/to/my-project` |

## 常见 Agent 配置

### Claude Code

```bash
# 基本用法
bar run -- claude "your prompt here"

# 带上下文
bar run -- claude --context "src/" "fix all TypeScript errors"

# 交互模式（不推荐，无法捕获完整输出）
bar run -- claude --interactive
```

### Aider

```bash
# 基本用法
bar run -- aider --message "add error handling"

# 指定文件
bar run -- aider --file src/main.go --message "add logging"

# 使用特定模型
bar run -- aider --model gpt-4 --message "optimize this function"
```

### GitHub Copilot CLI

```bash
# 如果 Copilot 提供 CLI
bar run -- gh copilot suggest "how to fix this error"
```

### 自定义脚本

```bash
# 你自己的 agent 脚本
bar run -- python my_agent.py --task "refactor"

# Shell 脚本
bar run -- ./scripts/auto-fix.sh
```

## 最佳实践

### 1. 使用非交互模式

尽量使用 agent 的非交互模式，这样 BAR 可以完整捕获输出：

```bash
# ✅ 好：非交互
bar run -- claude "fix the bug"

# ⚠️ 不推荐：交互模式
bar run -- claude --interactive
```

### 2. 指定明确的任务

给 agent 明确的任务描述，方便后续审计：

```bash
# ✅ 好：明确的任务
bar run -- claude "fix null pointer exception in UserService.java line 42"

# ❌ 不好：模糊的任务
bar run -- claude "fix it"
```

### 3. 分步执行

复杂任务分成多个 step，方便回滚到特定状态：

```bash
# Step 1: 分析
bar run -- claude "analyze the codebase and identify issues"

# Step 2: 修复
bar run -- claude "fix the identified issues"

# Step 3: 测试
bar run -- npm test

# 如果测试失败，可以回滚到 step 2
bar rollback --step 0002
```

### 4. 设置超时

对于可能长时间运行的 agent，设置超时：

```bash
bar run --timeout 10m -- long-running-agent
```

## 故障排除

### Agent 无法找到文件

确保 agent 使用相对路径，或者检查 `BAR_WORKSPACE` 环境变量：

```bash
# 检查当前 worktree
bar status

# 手动进入 worktree 验证
cd $(bar status --json | jq -r '.workspace_path')
ls -la
```

### 输出被截断

某些 agent 的输出可能很长，检查 artifacts 文件：

```bash
# 查看完整输出
cat .bar/tasks/<task_id>/artifacts/0001.output
```

### Agent 修改了 .bar/ 目录

这是不允许的，BAR 会忽略 `.bar/` 目录的变更。如果 agent 尝试修改，diff 中不会显示。

## 高级用法

### 传递环境变量

```bash
# 传递 API key
bar run --env OPENAI_API_KEY=xxx -- my-agent

# 传递多个变量
bar run --env KEY1=val1 --env KEY2=val2 -- my-agent
```

### 指定工作目录

```bash
# 在 worktree 的子目录中执行
bar run --cwd src/ -- npm test
```

### 不记录到 ledger

```bash
# 临时命令，不记录
bar run --no-record -- ls -la
```
