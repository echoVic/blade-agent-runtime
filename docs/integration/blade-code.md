# Blade Code 集成指南

> 本文档描述如何将 blade-code 与 BAR 集成，使 blade-code 的文件操作发生在隔离的 worktree 中。

## 集成目标

1. blade-code 的所有文件写入发生在 BAR 的 worktree 中
2. 用户可以通过 `bar diff` 查看 blade-code 的变更
3. 用户可以通过 `bar apply` 将变更应用到主分支
4. 用户可以通过 `bar rollback` 回滚 blade-code 的变更

## 集成方案

### 方案 A：环境变量注入

BAR 在执行 `bar run` 时注入环境变量，blade-code 读取这些变量来确定工作目录。

```bash
# BAR 注入的环境变量
BAR_WORKSPACE=/path/to/repo/.bar/workspaces/abc123
BAR_TASK_ID=abc123
BAR_ACTIVE=true
```

blade-code 检测到 `BAR_ACTIVE=true` 时，将所有文件操作重定向到 `BAR_WORKSPACE`。

### 方案 B：包装脚本

用户通过 `bar run` 包装 blade-code 命令：

```bash
bar run -- blade-code "implement feature X"
```

BAR 会：
1. 切换到 worktree 目录
2. 执行 blade-code
3. 捕获输出
4. 记录 diff

### 方案 C：直接集成（推荐）

blade-code 内置 BAR 支持，检测到 `.bar/` 目录时自动使用 active task 的 worktree。

```go
// blade-code 中的集成代码
func getWorkingDirectory() string {
    // 检查是否在 BAR 管理的 repo 中
    barDir := findBarDir()
    if barDir == "" {
        return getCurrentDir()
    }
    
    // 读取 active task
    state := readState(barDir)
    if state.ActiveTaskID == "" {
        return getCurrentDir()
    }
    
    // 返回 worktree 路径
    return filepath.Join(barDir, "workspaces", state.ActiveTaskID)
}
```

## 实现步骤

### Step 1: BAR 侧准备

确保 BAR 提供以下能力：

```go
// 获取当前 active task 的 worktree 路径
func GetActiveWorkspace() (string, error)

// 检查是否在 BAR 管理的 repo 中
func IsBarRepo() bool

// 获取 BAR 状态
func GetState() (*State, error)
```

### Step 2: blade-code 侧集成

```go
package main

import (
    "os"
    "path/filepath"
)

func init() {
    // 在启动时检查 BAR 环境
    if workspace := os.Getenv("BAR_WORKSPACE"); workspace != "" {
        // 切换工作目录到 worktree
        os.Chdir(workspace)
    }
}
```

### Step 3: 测试验证

```bash
# 1. 初始化 BAR
cd my-project
bar init

# 2. 创建任务
bar task start feature-x

# 3. 运行 blade-code
bar run -- blade-code "implement user authentication"

# 4. 验证变更在 worktree 中
bar diff
# 应该显示 blade-code 的变更

# 5. 应用变更
bar apply
```

## 注意事项

1. **路径处理**：blade-code 需要正确处理相对路径和绝对路径
2. **文件监听**：如果 blade-code 使用文件监听，需要监听 worktree 而非主 repo
3. **Git 操作**：blade-code 的 git 操作需要在 worktree 中执行

## 未来扩展

- 支持 blade-code 直接调用 BAR API
- 支持 blade-code 创建/切换 task
- 支持 blade-code 读取 ledger 历史
