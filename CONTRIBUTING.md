# Contributing to BAR

感谢你对 Blade Agent Runtime 的关注！欢迎贡献代码、报告问题或提出建议。

## 开发环境

### 前置要求

- Go 1.21+
- Git

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/echoVic/blade-agent-runtime.git
cd blade-agent-runtime

# 编译
go build -o bin/bar ./cmd/bar

# 运行测试
go test ./...

# 格式化代码
gofmt -w ./cmd ./internal
```

## 提交代码

### 分支命名

- `feat/xxx` - 新功能
- `fix/xxx` - Bug 修复
- `docs/xxx` - 文档更新
- `refactor/xxx` - 重构

### Commit 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>: <description>

[optional body]
```

类型：
- `feat` - 新功能
- `fix` - Bug 修复
- `docs` - 文档
- `refactor` - 重构
- `test` - 测试
- `chore` - 构建/工具

示例：
```
feat: add JSON output format for diff command
fix: handle empty ledger in log command
docs: update installation instructions
```

### Pull Request

1. Fork 仓库
2. 创建分支 `git checkout -b feat/my-feature`
3. 提交更改 `git commit -m "feat: add my feature"`
4. 推送分支 `git push origin feat/my-feature`
5. 创建 Pull Request

### PR 检查清单

- [ ] 代码通过 `go test ./...`
- [ ] 代码通过 `gofmt` 格式化
- [ ] 添加了必要的测试
- [ ] 更新了相关文档

## 项目结构

```
blade-agent-runtime/
├── cmd/bar/           # CLI 入口
├── internal/
│   ├── adapters/git/  # Git 适配层
│   ├── core/          # 核心逻辑
│   │   ├── task/      # 任务管理
│   │   ├── workspace/ # Worktree 管理
│   │   ├── ledger/    # 操作日志
│   │   ├── diff/      # Diff 生成
│   │   ├── apply/     # 应用变更
│   │   ├── policy/    # 策略引擎
│   │   └── config/    # 配置管理
│   └── util/          # 工具函数
├── docs/              # GitHub Pages
├── docs-md/           # 设计文档
└── scripts/           # 脚本
```

## 报告问题

请在 [GitHub Issues](https://github.com/echoVic/blade-agent-runtime/issues) 提交问题，包含：

- 问题描述
- 复现步骤
- 期望行为
- 实际行为
- 环境信息（OS、Go 版本）

## 联系方式

- GitHub: [@echoVic](https://github.com/echoVic)
- Website: [echovic.com](https://echovic.com)

## License

贡献的代码将遵循 [MIT License](LICENSE)。
