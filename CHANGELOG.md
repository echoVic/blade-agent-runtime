# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.20] - 2026-02-04

### Changed
- Version bump


## [0.0.19] - 2026-02-04

### Documentation
- 更新发布脚本和变更日志格式

### Chore
- 修复 CHANGELOG 生成脚本兼容 macOS bash 3.x


## [0.0.18] - 2026-02-04

### Documentation
- 更新发布脚本和变更日志格式

### Chore
- 修复 CHANGELOG 生成脚本的正则匹配


## [0.0.17] - 2026-02-04

### Documentation
- 更新 CHANGELOG 和增强发布脚本


## [0.0.16] - 2026-02-04

### Added
- 增强发布脚本：自动从 git commit 生成 CHANGELOG
- 更新文档：architecture.md、data-model.md、cli.md 路径改为 `~/.bar`


## [0.0.15] - 2026-02-04

### Changed
- **存档目录迁移**: 从项目内 `.bar/` 迁移到用户主目录 `~/.bar/projects/<project>-<hash4>/`
- 发布脚本增强：支持 `patch/minor/major` 语义化版本命令

### Added
- `internal/util/path/path_test.go` 单元测试
- `CHANGELOG.md` 变更日志

### Removed
- `checkGitignore()` 函数

## [0.0.14] - 2026-02-04

### Changed
- **存档目录迁移**: 从项目内 `.bar/` 迁移到用户主目录 `~/.bar/projects/<project>-<hash4>/`

### Added
- `internal/util/path/path_test.go` 单元测试

### Removed
- `checkGitignore()` 函数

## [0.0.13] - 2026-02-01

### Changed
- Web UI 改进

## [0.0.12] - 2026-02-01

### Fixed
- 修复 wrapped 命令退出时关闭 Web UI

## [0.0.11] - 2026-02-01

### Fixed
- 修复 wrapped 命令退出后保持 Web UI 运行

## [0.0.10] - 2026-02-01

### Changed
- `bar wrap` 默认启用 Web UI

### Fixed
- 修复 `bar wrap --ui` 的 URL 格式

## [0.0.9] - 2026-02-01

### Added
- `bar wrap` 新增 `--ui` 参数

## [0.0.8] - 2026-02-01

### Changed
- Web UI v2 重新设计（Split View + Monaco Editor）

## [0.0.7] - 2026-02-01

### Added
- Web UI 任务审计界面

## [0.0.6] - 2026-02-01

### Fixed
- 修复 release 脚本中 cmd/bar 文件强制添加

## [0.0.5] - 2026-02-01

### Added
- `bar wrap` 自动初始化并创建任务

## [0.0.4] - 2026-02-01

### Added
- `bar update` 命令
- `bar version` 命令

## [0.0.3] - 2026-02-01

### Added
- 任务启动时自动初始化 BAR

## [0.0.2] - 2026-02-01

### Added
- `bar wrap` 命令，支持交互式 Agent
- curl 安装脚本
- GitHub Pages 短链接安装

## [0.0.1] - 2026-02-01

### Added
- 初始实现 Blade Agent Runtime (BAR)
- 核心功能：任务管理、diff/apply、workspace 管理
- CLI 命令：init, task, run, diff, apply, rollback, status, log
- Policy 引擎
- 单元测试
