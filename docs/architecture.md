# 架构设计

## 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer (bar)                          │
│  init | task | run | diff | apply | rollback | status | log     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Core Layer                               │
│  ┌─────────┐ ┌───────────┐ ┌────────┐ ┌───────┐ ┌────────┐     │
│  │  Task   │ │ Workspace │ │ Ledger │ │ Diff  │ │ Policy │     │
│  │ Manager │ │  Manager  │ │Manager │ │Engine │ │ Engine │     │
│  └─────────┘ └───────────┘ └────────┘ └───────┘ └────────┘     │
│  ┌─────────┐ ┌───────────┐                                      │
│  │  Apply  │ │   Exec    │                                      │
│  │ Engine  │ │  Runner   │                                      │
│  └─────────┘ └───────────┘                                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Adapter Layer                              │
│         ┌─────────────┐              ┌─────────────┐            │
│         │ Git Adapter │              │ FS Adapter  │            │
│         │ (worktree,  │              │ (file ops)  │            │
│         │  diff, etc) │              │             │            │
│         └─────────────┘              └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Storage Layer                              │
│    ~/.bar/projects/<project>-<hash4>/                            │
│    ├── state.json           # 全局状态                           │
│    ├── config.yaml          # 配置文件                           │
│    ├── tasks/<task_id>/                                         │
│    │   ├── task.json        # Task 元信息                        │
│    │   ├── ledger.jsonl     # Step 日志（追加写入）               │
│    │   └── artifacts/       # Diff/Output 文件                   │
│    └── workspaces/<task_id>/ # Git Worktree                     │
└─────────────────────────────────────────────────────────────────┘
```

## 核心模块职责

### 1. Task Manager (`internal/core/task`)

**职责**：管理 Task 的生命周期

```go
type TaskManager interface {
    Create(name string, baseRef string) (*Task, error)
    Get(taskID string) (*Task, error)
    List() ([]*Task, error)
    GetActive() (*Task, error)
    SetActive(taskID string) error
    Close(taskID string) error
    Delete(taskID string) error
}
```

**状态机**：
```
                    ┌─────────┐
         create     │         │    close
    ─────────────▶  │ active  │ ─────────────▶ closed
                    │         │
                    └─────────┘
                         │
                         │ delete
                         ▼
                    ┌─────────┐
                    │ deleted │
                    └─────────┘
```

### 2. Workspace Manager (`internal/core/workspace`)

**职责**：管理 Git Worktree 的创建和销毁

```go
type WorkspaceManager interface {
    Create(taskID string, baseRef string) (string, error)  // 返回 worktree 路径
    Delete(taskID string) error
    GetPath(taskID string) string
    IsClean(taskID string) (bool, error)
    Reset(taskID string) error  // git reset --hard + clean
}
```

**实现细节**：
- 使用 `git worktree add` 创建隔离工作区
- 分支命名：`bar/<task-name>-<short-id>`
- Worktree 路径：`~/.bar/projects/<project>-<hash4>/workspaces/<task_id>`

### 3. Ledger Manager (`internal/core/ledger`)

**职责**：管理 Step 日志的读写

```go
type LedgerManager interface {
    Append(taskID string, step *Step) error
    List(taskID string) ([]*Step, error)
    GetLast(taskID string) (*Step, error)
    GetByID(taskID string, stepID string) (*Step, error)
}
```

**存储格式**：JSONL（每行一个 JSON 对象）

```jsonl
{"step_id":"0001","kind":"run","cmd":["claude","fix bug"],...}
{"step_id":"0002","kind":"apply","mode":"commit",...}
```

### 4. Diff Engine (`internal/core/diff`)

**职责**：生成和管理 diff

```go
type DiffEngine interface {
    Generate(taskID string) (*DiffResult, error)
    GenerateFromStep(taskID string, stepID string) (*DiffResult, error)
    SavePatch(taskID string, stepID string, patch []byte) error
    GetPatch(taskID string, stepID string) ([]byte, error)
}

type DiffResult struct {
    Files     int
    Additions int
    Deletions int
    Patch     []byte
    FileList  []string
}
```

### 5. Apply Engine (`internal/core/apply`)

**职责**：将 worktree 变更应用到主分支

```go
type ApplyEngine interface {
    Commit(taskID string, message string) (string, error)  // 返回 commit SHA
    Merge(taskID string, message string) error             // v0.2
    CherryPick(taskID string, commits []string) error      // v0.2
}
```

**v0 只实现 Commit**：
1. 在 worktree 分支上创建 commit
2. 切换到主分支
3. Cherry-pick 或 merge

### 6. Policy Engine (`internal/core/policy`)

**职责**：检查命令是否安全

```go
type PolicyEngine interface {
    Check(cmd []string) (*PolicyResult, error)
    LoadPolicy(path string) error
}

type PolicyResult struct {
    Allowed bool
    Reason  string
    Rule    string
}
```

**v0 规则（简单 blocklist）**：
```yaml
# .bar/policy.yaml
version: 1
rules:
  - pattern: "rm -rf /"
    action: block
    reason: "Dangerous: recursive delete from root"
  - pattern: "rm -rf ~"
    action: block
    reason: "Dangerous: recursive delete home directory"
  - pattern: "> /dev/sda"
    action: block
    reason: "Dangerous: write to disk device"
```

### 7. Exec Runner (`internal/core/exec`)

**职责**：执行外部命令并捕获输出

```go
type ExecRunner interface {
    Run(ctx context.Context, cmd []string, opts *RunOptions) (*RunResult, error)
}

type RunOptions struct {
    Cwd     string            // 工作目录（worktree 路径）
    Env     map[string]string // 环境变量
    Timeout time.Duration     // 超时时间
}

type RunResult struct {
    ExitCode int
    Stdout   []byte
    Stderr   []byte
    Duration time.Duration
}
```

## 数据流

### 典型流程：`bar run -- claude "fix bug"`

```
1. CLI 解析命令
   │
   ▼
2. 获取当前 active task
   │
   ▼
3. Policy Engine 检查命令是否安全
   │
   ├─ 不安全 ──▶ 返回错误
   │
   ▼
4. Exec Runner 在 worktree 中执行命令
   │
   ├─ 捕获 stdout/stderr
   │
   ▼
5. Diff Engine 生成 diff
   │
   ▼
6. Ledger Manager 追加 step 记录
   │
   ├─ 写入 ledger.jsonl
   ├─ 保存 patch 到 artifacts/
   ├─ 保存 output 到 artifacts/
   │
   ▼
7. 返回执行结果
```

## 错误处理策略

### 原则

1. **Fail Fast**：遇到错误立即返回，不尝试自动恢复
2. **状态一致**：任何操作要么完全成功，要么完全失败
3. **可恢复**：失败后用户可以通过 `bar rollback` 恢复

### 错误类型

```go
type BarError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

const (
    ErrNotInitialized    ErrorCode = "NOT_INITIALIZED"
    ErrNoActiveTask      ErrorCode = "NO_ACTIVE_TASK"
    ErrTaskNotFound      ErrorCode = "TASK_NOT_FOUND"
    ErrWorkspaceNotClean ErrorCode = "WORKSPACE_NOT_CLEAN"
    ErrPolicyViolation   ErrorCode = "POLICY_VIOLATION"
    ErrGitOperation      ErrorCode = "GIT_OPERATION"
    ErrExecFailed        ErrorCode = "EXEC_FAILED"
)
```

## 并发模型

### v0 简化模型

- **单 task 单进程**：同一时间只有一个 `bar run` 在执行
- **无锁设计**：JSONL 追加写入天然支持
- **文件锁**：可选，用于防止多个 bar 实例同时操作

### 未来扩展

- 支持多个 task 并行（不同 worktree）
- 支持同一 task 内的并行 step（需要更复杂的 ledger 设计）

## 扩展点

### 1. 自定义 Policy

用户可以在项目根目录的 `.bar/policy.yaml` 中定义自己的规则。

### 2. 自定义 Hooks

```yaml
# .bar/config.yaml
hooks:
  pre_run:
    - "echo 'Starting agent...'"
  post_run:
    - "notify-send 'Agent finished'"
```

### 3. 输出格式

```bash
bar diff --format=json
bar log --format=markdown
```

## 设计决策

### 技术选型

| 决策 | 选择 | 理由 |
|------|------|------|
| Git 操作 | `go-git` 库 | 纯 Go 实现，不依赖系统 git，跨平台更稳定 |
| ID 生成 | `nanoid` (8 字符) | 比 UUID 更短、更可读，适合 CLI 场景 |

### 行为决策

| 决策 | 选择 | 理由 |
|------|------|------|
| `bar run` 交互模式 | 透传 stdin/stdout | 支持用户与 agent 交互，但输出捕获可能不完整 |
| `bar task close` worktree | 默认删除，`--keep` 保留 | 节省空间，同时保留调试能力 |
| `bar task start` 默认 base | 当前 HEAD | 最符合用户预期，用户通常在目标分支上执行 |

## 依赖

### 外部依赖

| 依赖 | 用途 | 版本要求 |
|------|------|----------|
| Git | Worktree 管理（go-git 内部使用） | >= 2.20 |

### Go 依赖

| 库 | 用途 |
|----|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `github.com/spf13/viper` | 配置管理 |
| `github.com/go-git/go-git/v5` | Git 操作 |
| `github.com/jaevor/go-nanoid` | ID 生成 |
| `gopkg.in/yaml.v3` | YAML 解析 |

## 测试策略

### 单元测试

- 每个 Manager 都有对应的 mock
- 测试核心逻辑，不依赖真实 git repo

### 集成测试

- 在临时目录创建真实 git repo
- 测试完整流程

### E2E 测试

- 使用 `examples/` 中的脚本
- 验证 CLI 行为符合预期
