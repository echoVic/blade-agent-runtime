# 数据模型设计

## 目录结构

存档目录位于用户主目录 `~/.bar/`，按项目组织：

```
~/.bar/
└── projects/
    └── <project_name>-<hash4>/     # 如 my-project-a3f2
        ├── config.yaml             # 项目配置
        ├── state.json              # 全局状态（当前 active task）
        ├── tasks/                  # 任务数据
        │   └── <task_id>/
        │       ├── task.json       # 任务元信息
        │       ├── ledger.jsonl    # 操作日志（JSONL 格式）
        │       └── artifacts/      # 产物文件
        │           ├── 0001.patch  # Step 1 的 diff
        │           ├── 0001.output # Step 1 的输出
        │           ├── 0002.patch
        │           ├── 0002.output
        │           └── ...
        └── workspaces/             # Git Worktree 目录
            └── <task_id>/          # 每个 task 一个 worktree
```

---

## 配置文件

### `config.yaml`

项目配置文件，`bar init` 时创建，位于 `~/.bar/projects/<project>-<hash4>/config.yaml`。

```yaml
version: 1

git:
  default_base: main
  branch_prefix: bar/

policy:
  enabled: false
  path: .bar/policy.yaml

hooks:
  pre_run: []
  post_run: []

output:
  color: true
  verbose: false
```

**字段说明：**

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `version` | int | 配置版本 | 1 |
| `git.default_base` | string | 默认基准分支 | main |
| `git.branch_prefix` | string | 分支名前缀 | bar/ |
| `policy.enabled` | bool | 是否启用 policy 检查 | false |
| `policy.path` | string | policy 文件路径 | .bar/policy.yaml |
| `hooks.pre_run` | []string | run 前执行的命令 | [] |
| `hooks.post_run` | []string | run 后执行的命令 | [] |
| `output.color` | bool | 是否启用彩色输出 | true |
| `output.verbose` | bool | 是否启用详细输出 | false |

---

### `state.json`

项目状态文件，记录当前 active task，位于 `~/.bar/projects/<project>-<hash4>/state.json`。

```json
{
  "version": 1,
  "active_task_id": "abc123",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `version` | int | 状态版本 |
| `active_task_id` | string | 当前激活的任务 ID（可为空） |
| `updated_at` | string | 最后更新时间（ISO 8601） |

---

## Task 数据模型

### `task.json`

任务元信息文件。

```json
{
  "id": "abc123",
  "name": "fix-null-pointer",
  "repo_root": "/Users/xxx/my-project",
  "base_ref": "main",
  "branch": "bar/fix-null-pointer-abc123",
  "workspace_path": ".bar/workspaces/abc123",
  "status": "active",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "closed_at": null,
  "metadata": {
    "description": "Fix null pointer exception in main.go",
    "tags": ["bug", "critical"]
  }
}
```

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | ✅ | 唯一标识符（nanoid 或 UUID 短格式） |
| `name` | string | ✅ | 任务名称（用户指定） |
| `repo_root` | string | ✅ | 仓库根目录绝对路径 |
| `base_ref` | string | ✅ | 基准分支/commit |
| `branch` | string | ✅ | worktree 分支名 |
| `workspace_path` | string | ✅ | worktree 相对路径 |
| `status` | string | ✅ | 状态：active / closed |
| `created_at` | string | ✅ | 创建时间（ISO 8601） |
| `updated_at` | string | ✅ | 最后更新时间 |
| `closed_at` | string | ❌ | 关闭时间（可为 null） |
| `metadata` | object | ❌ | 用户自定义元数据 |

**Go 结构体：**

```go
type Task struct {
    ID            string            `json:"id"`
    Name          string            `json:"name"`
    RepoRoot      string            `json:"repo_root"`
    BaseRef       string            `json:"base_ref"`
    Branch        string            `json:"branch"`
    WorkspacePath string            `json:"workspace_path"`
    Status        TaskStatus        `json:"status"`
    CreatedAt     time.Time         `json:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at"`
    ClosedAt      *time.Time        `json:"closed_at,omitempty"`
    Metadata      map[string]any    `json:"metadata,omitempty"`
}

type TaskStatus string

const (
    TaskStatusActive TaskStatus = "active"
    TaskStatusClosed TaskStatus = "closed"
)
```

---

## Step 数据模型

### `ledger.jsonl`

操作日志文件，使用 JSONL 格式（每行一个 JSON 对象）。

**为什么用 JSONL？**
1. **追加写入友好**：新 step 直接 append，不需要读取-修改-写入
2. **流式处理**：可以逐行读取，不需要一次性加载整个文件
3. **日志语义**：Ledger 本质上是"日志"，JSONL 天然契合
4. **调试友好**：`cat ledger.jsonl | jq .` 即可查看

**示例内容：**

```jsonl
{"step_id":"0001","kind":"run","cmd":["claude","analyze the codebase"],"cwd":".","started_at":"2024-01-15T10:01:00Z","ended_at":"2024-01-15T10:01:30Z","exit_code":0,"diff_stat":{"files":0,"additions":0,"deletions":0},"artifacts":{}}
{"step_id":"0002","kind":"run","cmd":["claude","fix the null pointer in main.go"],"cwd":".","started_at":"2024-01-15T10:02:00Z","ended_at":"2024-01-15T10:03:00Z","exit_code":0,"diff_stat":{"files":3,"additions":15,"deletions":5},"artifacts":{"patch":"artifacts/0002.patch","output":"artifacts/0002.output"}}
{"step_id":"0003","kind":"apply","mode":"commit","commit_sha":"abc1234","started_at":"2024-01-15T10:04:00Z","ended_at":"2024-01-15T10:04:05Z"}
{"step_id":"0004","kind":"rollback","target":"base","started_at":"2024-01-15T10:05:00Z","ended_at":"2024-01-15T10:05:02Z"}
```

---

### Step 类型

#### 1. Run Step（执行命令）

```json
{
  "step_id": "0002",
  "kind": "run",
  "cmd": ["claude", "fix the null pointer in main.go"],
  "cwd": ".",
  "env": {
    "CLAUDE_API_KEY": "***"
  },
  "started_at": "2024-01-15T10:02:00Z",
  "ended_at": "2024-01-15T10:03:00Z",
  "duration_ms": 60000,
  "exit_code": 0,
  "diff_stat": {
    "files": 3,
    "additions": 15,
    "deletions": 5,
    "file_list": ["main.go", "utils.go", "config.go"]
  },
  "artifacts": {
    "patch": "artifacts/0002.patch",
    "output": "artifacts/0002.output"
  },
  "policy_events": []
}
```

#### 2. Apply Step（应用变更）

```json
{
  "step_id": "0003",
  "kind": "apply",
  "mode": "commit",
  "commit_sha": "abc1234def5678",
  "commit_message": "fix: null pointer exception in main.go",
  "target_branch": "main",
  "started_at": "2024-01-15T10:04:00Z",
  "ended_at": "2024-01-15T10:04:05Z",
  "duration_ms": 5000
}
```

#### 3. Rollback Step（回滚）

```json
{
  "step_id": "0004",
  "kind": "rollback",
  "target": "base",
  "target_step": null,
  "hard": false,
  "started_at": "2024-01-15T10:05:00Z",
  "ended_at": "2024-01-15T10:05:02Z",
  "duration_ms": 2000
}
```

---

### Step 字段说明

**通用字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `step_id` | string | ✅ | 步骤 ID（格式：0001, 0002, ...） |
| `kind` | string | ✅ | 类型：run / apply / rollback |
| `started_at` | string | ✅ | 开始时间（ISO 8601） |
| `ended_at` | string | ✅ | 结束时间 |
| `duration_ms` | int | ❌ | 耗时（毫秒） |

**Run Step 特有字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `cmd` | []string | ✅ | 执行的命令 |
| `cwd` | string | ✅ | 工作目录（相对于 worktree） |
| `env` | object | ❌ | 环境变量（敏感信息会脱敏） |
| `exit_code` | int | ✅ | 退出码 |
| `diff_stat` | object | ✅ | diff 统计 |
| `artifacts` | object | ✅ | 产物文件路径 |
| `policy_events` | []object | ❌ | policy 检查事件 |

**Apply Step 特有字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `mode` | string | ✅ | 模式：commit / merge |
| `commit_sha` | string | ✅ | commit SHA |
| `commit_message` | string | ✅ | commit 消息 |
| `target_branch` | string | ✅ | 目标分支 |

**Rollback Step 特有字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | ✅ | 目标：base / step |
| `target_step` | string | ❌ | 目标 step ID（当 target=step 时） |
| `hard` | bool | ✅ | 是否硬回滚 |

---

### Go 结构体

```go
type Step struct {
    StepID    string    `json:"step_id"`
    Kind      StepKind  `json:"kind"`
    StartedAt time.Time `json:"started_at"`
    EndedAt   time.Time `json:"ended_at"`
    DurationMs int64    `json:"duration_ms,omitempty"`

    // Run step fields
    Cmd          []string          `json:"cmd,omitempty"`
    Cwd          string            `json:"cwd,omitempty"`
    Env          map[string]string `json:"env,omitempty"`
    ExitCode     *int              `json:"exit_code,omitempty"`
    DiffStat     *DiffStat         `json:"diff_stat,omitempty"`
    Artifacts    *Artifacts        `json:"artifacts,omitempty"`
    PolicyEvents []PolicyEvent     `json:"policy_events,omitempty"`

    // Apply step fields
    Mode          string `json:"mode,omitempty"`
    CommitSHA     string `json:"commit_sha,omitempty"`
    CommitMessage string `json:"commit_message,omitempty"`
    TargetBranch  string `json:"target_branch,omitempty"`

    // Rollback step fields
    Target     string `json:"target,omitempty"`
    TargetStep string `json:"target_step,omitempty"`
    Hard       *bool  `json:"hard,omitempty"`
}

type StepKind string

const (
    StepKindRun      StepKind = "run"
    StepKindApply    StepKind = "apply"
    StepKindRollback StepKind = "rollback"
)

type DiffStat struct {
    Files     int      `json:"files"`
    Additions int      `json:"additions"`
    Deletions int      `json:"deletions"`
    FileList  []string `json:"file_list,omitempty"`
}

type Artifacts struct {
    Patch  string `json:"patch,omitempty"`
    Output string `json:"output,omitempty"`
}

type PolicyEvent struct {
    Rule    string `json:"rule"`
    Action  string `json:"action"`
    Matched string `json:"matched"`
}
```

---

## Artifacts 文件

### `<step_id>.patch`

Git diff 格式的 patch 文件。

```diff
diff --git a/main.go b/main.go
index abc1234..def5678 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,10 @@ func main() {
     config := loadConfig()
+    if config == nil {
+        log.Fatal("config is nil")
+        return
+    }
     server := newServer(config)
```

### `<step_id>.output`

命令输出文件，包含 stdout 和 stderr。

```
=== STDOUT ===
Analyzing codebase...
Found 3 potential issues.
Fixing null pointer in main.go...
Done.

=== STDERR ===
Warning: deprecated API usage in utils.go
```

---

## Policy 文件

### `policy.yaml`

Policy 文件位于项目根目录 `.bar/policy.yaml`（注意：这是项目内的配置，不在 `~/.bar` 中）。

```yaml
version: 1

rules:
  - name: no-rm-rf-root
    pattern: "rm\\s+-rf\\s+/"
    action: block
    reason: "Dangerous: recursive delete from root"

  - name: no-rm-rf-home
    pattern: "rm\\s+-rf\\s+~"
    action: block
    reason: "Dangerous: recursive delete home directory"

  - name: no-disk-write
    pattern: ">\\s*/dev/sd"
    action: block
    reason: "Dangerous: write to disk device"

  - name: warn-sudo
    pattern: "^sudo\\s+"
    action: warn
    reason: "Warning: running with elevated privileges"
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `version` | int | 配置版本 |
| `rules[].name` | string | 规则名称 |
| `rules[].pattern` | string | 正则表达式 |
| `rules[].action` | string | 动作：block / warn / log |
| `rules[].reason` | string | 原因说明 |

---

## ID 生成策略

### Task ID

使用 nanoid 或 UUID 短格式，长度 8 字符。

```go
import "github.com/jaevor/go-nanoid"

func generateTaskID() string {
    gen, _ := nanoid.Standard(8)
    return gen()
}
// Output: "abc12345"
```

### Step ID

使用递增的 4 位数字，格式：`0001`, `0002`, ...

```go
func generateStepID(lastStepID string) string {
    if lastStepID == "" {
        return "0001"
    }
    num, _ := strconv.Atoi(lastStepID)
    return fmt.Sprintf("%04d", num+1)
}
```

---

## 数据迁移

### 版本兼容性

所有配置文件和数据文件都包含 `version` 字段，用于未来的数据迁移。

```go
func migrateTaskJSON(data []byte) (*Task, error) {
    var raw map[string]any
    json.Unmarshal(data, &raw)
    
    version := int(raw["version"].(float64))
    switch version {
    case 1:
        return parseTaskV1(data)
    default:
        return nil, fmt.Errorf("unsupported task version: %d", version)
    }
}
```

---

## 并发安全

### JSONL 追加写入

JSONL 格式天然支持追加写入，但需要注意：

1. **单进程写入**：v0 假设同一时间只有一个 `bar run` 在执行
2. **原子写入**：每次写入一行完整的 JSON，避免部分写入
3. **文件锁**：可选，用于防止多个 bar 实例同时操作

```go
func appendToLedger(path string, step *Step) error {
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    data, err := json.Marshal(step)
    if err != nil {
        return err
    }
    
    // 确保原子写入（一行完整的 JSON + 换行符）
    _, err = f.Write(append(data, '\n'))
    return err
}
```
