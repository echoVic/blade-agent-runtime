# 存档目录迁移：从 `.bar` 到 `~/.bar`

## 背景

原设计将存档目录放在项目内的 `.bar/`，存在以下问题：
1. 每个项目都需要在 `.gitignore` 中添加 `.bar/`
2. 不符合常见 CLI 工具的惯例（如 `~/.npm`, `~/.cargo`, `~/.docker`）
3. workspaces 使用 git worktree 会在项目内创建大量副本

## 设计决策

### 方案对比

| 方案 | 结构 | 优点 | 缺点 |
|------|------|------|------|
| A: Hash 映射 | `~/.bar/repos/<hash>/` | 绝对唯一 | 不直观，需要额外映射文件 |
| B: 项目名分组 | `~/.bar/projects/<name>/` | 直观可读 | 同名项目冲突 |
| C: 项目名+Hash | `~/.bar/projects/<name>-<hash4>/` | 直观且唯一 | 名称稍长 |

**最终选择方案 C**：`<项目名>-<路径SHA256前4位>`，兼顾可读性和唯一性。

## 新目录结构

```
~/.bar/
├── config.yaml                          # 全局配置（可选）
└── projects/
    └── <project_name>-<hash4>/          # 如 "blade-agent-runtime-a3f2"
        ├── state.json
        ├── tasks/
        │   └── <task_id>/
        │       ├── task.json
        │       ├── ledger.jsonl
        │       └── artifacts/
        └── workspaces/
            └── <task_id>/
```

**命名规则**: `<项目目录名>-<路径SHA256前4位>`
- 例如: `/Users/bytedance/Documents/GitHub/blade-agent-runtime` → `blade-agent-runtime-a3f2`

## 修改文件清单

| 文件 | 修改内容 |
|------|----------|
| `internal/util/path/path.go` | 新增 `GlobalBarDir()` 返回 `~/.bar`；新增 `ProjectID(repoRoot)` 生成 `name-hash4` |
| `cmd/bar/root.go` | 修改 `initApp()` 使用新的 barDir 路径 |
| `cmd/bar/init.go` | 移除 `.gitignore` 检查逻辑 |
| `internal/core/config/model.go` | 调整默认 policy 路径 |

## 具体修改

### 1. `internal/util/path/path.go`
```go
func GlobalBarDir() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".bar")
}

func ProjectID(repoRoot string) string {
    name := filepath.Base(repoRoot)
    hash := sha256.Sum256([]byte(repoRoot))
    return fmt.Sprintf("%s-%x", name, hash[:2]) // 前4位hex
}

func BarDir(repoRoot string) string {
    return filepath.Join(GlobalBarDir(), "projects", ProjectID(repoRoot))
}
```

### 2. `cmd/bar/init.go`
- 删除 `checkGitignore()` 函数及其调用

### 3. `internal/util/path/path_test.go`
新增单元测试覆盖：
- `TestGlobalBarDir` - 验证返回 `~/.bar`
- `TestProjectID` - 验证格式为 `<项目名>-<4位hash>`
- `TestProjectID_Uniqueness` - 验证不同路径生成不同 ID
- `TestProjectID_Consistency` - 验证相同路径生成相同 ID
- `TestBarDir` - 验证完整路径结构
- `TestFindRepoRoot` - 验证 git 仓库根目录查找
- `TestEnsureDir` - 验证递归创建目录

## 迁移影响

### 用户侧
- 无需再修改 `.gitignore`
- 所有项目的存档统一管理在 `~/.bar/`
- 可通过 `ls ~/.bar/projects/` 查看所有项目

### 开发侧
- `BarDir()` 函数签名不变，调用方无需修改
- 新增 `GlobalBarDir()` 和 `ProjectID()` 供其他模块使用
