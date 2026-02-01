# 开发路线图

## 总览

| 阶段 | 时间 | 目标 | 核心交付 |
|------|------|------|----------|
| Week 1 | 第 1 周 | 跑起来 | init, task start, status, run |
| Week 2 | 第 2 周 | 工程级差异 | diff, patch 产物, ledger |
| Week 3 | 第 3 周 | 可回滚/可应用 | apply, rollback |
| Week 4 | 第 4 周 | 可恢复 + Policy | resume, policy, log |
| Week 5-6 | 第 5-6 周 | 集成验证 | blade-code 接入, 文档完善 |

---

## Week 1：跑起来（价值闭环最小化）

### 目标
能创建隔离区 + 在隔离区跑命令

### 交付物

| 命令 | 说明 | 优先级 |
|------|------|--------|
| `bar init` | 初始化 .bar/ 目录 | P0 |
| `bar task start <name>` | 创建 task + worktree + 分支 | P0 |
| `bar status` | 显示当前状态 | P0 |
| `bar run -- <cmd>` | 在 worktree 中执行命令 | P0 |

### 技术任务

```
[ ] 项目初始化
    [ ] go mod init
    [ ] 引入 cobra CLI 框架
    [ ] 基础目录结构

[ ] bar init
    [ ] 检测 git repo
    [ ] 创建 .bar/ 目录结构
    [ ] 生成默认 config.yaml
    [ ] 更新 .gitignore

[ ] bar task start
    [ ] 生成 task ID (nanoid)
    [ ] 创建 git worktree
    [ ] 创建 task.json
    [ ] 创建空 ledger.jsonl
    [ ] 更新 state.json (active task)

[ ] bar status
    [ ] 读取 state.json
    [ ] 读取 task.json
    [ ] 检测 worktree 状态 (clean/dirty)

[ ] bar run
    [ ] 解析 -- 后的命令
    [ ] 在 worktree 目录执行
    [ ] 捕获 stdout/stderr
    [ ] 返回退出码
```

### 验收标准

```bash
# 1. 初始化
cd my-project
bar init
# ✅ 创建 .bar/ 目录

# 2. 创建任务
bar task start fix-bug
# ✅ 创建 worktree
# ✅ 创建分支 bar/fix-bug-xxx

# 3. 查看状态
bar status
# ✅ 显示 active task 信息

# 4. 运行命令
bar run -- echo "hello"
# ✅ 在 worktree 中执行
# ✅ 输出 "hello"

# 5. 手动验证
cd .bar/workspaces/xxx
touch test.txt
cd -
bar status
# ✅ 显示 dirty 状态
```

---

## Week 2：工程级差异（diff 是第一张门票）

### 目标
任何外部 agent 运行后，能输出 diff + patch

### 交付物

| 命令 | 说明 | 优先级 |
|------|------|--------|
| `bar diff` | 显示当前变更 | P0 |
| `bar diff --stat` | 显示统计信息 | P1 |
| `bar diff --output` | 导出 patch 文件 | P1 |

### 技术任务

```
[ ] Diff Engine
    [ ] 调用 git diff 生成 patch
    [ ] 解析 diff stat (files/add/del)
    [ ] 保存 patch 到 artifacts/

[ ] Ledger 写入
    [ ] bar run 执行后写入 step
    [ ] 记录 diff_stat
    [ ] 记录 artifacts 路径

[ ] bar diff 命令
    [ ] 显示当前 diff
    [ ] --stat 只显示统计
    [ ] --output 导出文件
    [ ] --step 查看历史 step 的 diff

[ ] 输出保存
    [ ] 保存 stdout/stderr 到 artifacts/
    [ ] 格式化输出文件
```

### 验收标准

```bash
# 1. 运行 agent 修改文件
bar run -- claude "add a hello function to main.go"

# 2. 查看 diff
bar diff
# ✅ 显示 git diff 格式的变更

bar diff --stat
# ✅ 显示 "2 files changed, 10 insertions(+), 2 deletions(-)"

# 3. 导出 patch
bar diff --output changes.patch
# ✅ 生成 patch 文件

# 4. 验证 ledger
cat .bar/tasks/xxx/ledger.jsonl
# ✅ 包含 step 记录
# ✅ 包含 diff_stat
# ✅ 包含 artifacts 路径

# 5. 验证 artifacts
ls .bar/tasks/xxx/artifacts/
# ✅ 0001.patch
# ✅ 0001.output
```

---

## Week 3：可回滚/可应用（从"玩具"变"敢用"）

### 目标
能安全地应用变更到主分支，也能随时回滚

### 交付物

| 命令 | 说明 | 优先级 |
|------|------|--------|
| `bar apply` | 应用变更到主分支 | P0 |
| `bar rollback --base` | 回滚到初始状态 | P0 |
| `bar task list` | 列出所有任务 | P1 |
| `bar task close` | 关闭任务 | P1 |

### 技术任务

```
[ ] Apply Engine
    [ ] 在 worktree 创建 commit
    [ ] 切换到主分支
    [ ] Cherry-pick commit
    [ ] 处理冲突（v0: 直接报错）

[ ] Rollback Engine
    [ ] git reset --hard
    [ ] git clean -fd
    [ ] 记录 rollback step

[ ] bar apply 命令
    [ ] --message 自定义消息
    [ ] --no-close 应用后不关闭
    [ ] 记录 apply step

[ ] bar rollback 命令
    [ ] --base 回滚到初始
    [ ] --hard 强制回滚

[ ] Task 管理
    [ ] bar task list
    [ ] bar task close
    [ ] bar task switch
```

### 验收标准

```bash
# 1. 运行 agent
bar run -- claude "fix the bug"

# 2. 查看 diff 确认
bar diff

# 3. 应用变更
bar apply --message "fix: resolve null pointer"
# ✅ 在主分支创建 commit
# ✅ 任务关闭

# 4. 验证
git log -1
# ✅ 显示新 commit

# --- 回滚场景 ---

# 1. 创建新任务
bar task start experiment

# 2. 运行 agent
bar run -- claude "try something crazy"

# 3. 不满意，回滚
bar rollback --base
# ✅ worktree 恢复到初始状态

bar status
# ✅ 显示 clean 状态
```

---

## Week 4：可恢复 + 最小 Policy

### 目标
中断后能继续；危险命令能拦；能导出报告

### 交付物

| 命令 | 说明 | 优先级 |
|------|------|--------|
| `bar log` | 查看操作日志 | P0 |
| `bar log --format markdown` | 导出报告 | P1 |
| Policy 检查 | 拦截危险命令 | P1 |

### 技术任务

```
[ ] Resume 能力
    [ ] bar status 定位 active task
    [ ] 断开重开后 ledger 不丢
    [ ] bar run 自动恢复上下文

[ ] Policy Engine (v0 简化版)
    [ ] 加载 policy.yaml
    [ ] 正则匹配检查
    [ ] block/warn/log 三种动作
    [ ] 记录 policy_events

[ ] bar log 命令
    [ ] 表格格式显示
    [ ] --step 查看详情
    [ ] --format json
    [ ] --format markdown
    [ ] --output 导出文件

[ ] 报告生成
    [ ] Markdown 模板
    [ ] 包含 task 信息
    [ ] 包含所有 step
    [ ] 包含 diff 统计
```

### 验收标准

```bash
# 1. Resume 测试
bar task start long-task
bar run -- sleep 5
# Ctrl+C 中断

bar status
# ✅ 显示 active task

bar run -- echo "continue"
# ✅ 继续在同一个 task 中执行

# 2. Policy 测试
echo 'rules:
  - pattern: "rm -rf /"
    action: block
    reason: "Dangerous"' > .bar/policy.yaml

bar run -- rm -rf /
# ✅ 被拦截
# ✅ 显示 "Policy violation: Dangerous"

# 3. Log 测试
bar log
# ✅ 显示所有 step 的表格

bar log --step 0001
# ✅ 显示详细信息

bar log --format markdown --output report.md
# ✅ 生成 Markdown 报告
```

---

## Week 5-6：集成验证（可选）

### 目标
证明 BAR 是真正的基建，能接入实际项目

### 交付物

| 任务 | 说明 | 优先级 |
|------|------|--------|
| blade-code 接入 | 让 blade-code 的执行落在 BAR 中 | P1 |
| 文档完善 | 完善 README 和使用文档 | P1 |
| Demo 脚本 | 提供可运行的 demo | P2 |

### 技术任务

```
[ ] blade-code 集成
    [ ] 分析 blade-code 执行模型
    [ ] 设计集成方案
    [ ] 实现集成代码
    [ ] 测试验证

[ ] 文档
    [ ] 完善 README
    [ ] 添加 GIF 演示
    [ ] 编写 integration 文档
    [ ] 添加 FAQ

[ ] Demo
    [ ] demo-claude-code 脚本
    [ ] demo-blade-code 脚本
    [ ] 录制演示视频
```

### 验收标准

```bash
# 1. blade-code 集成
bar task start feature
blade-code "implement user auth"
# ✅ 变更发生在 worktree
bar diff
# ✅ 能看到 blade-code 的变更
bar apply
# ✅ 能应用到主分支

# 2. Demo 可运行
cd examples/demo-claude-code
./run.sh
# ✅ 完整演示流程
```

---

## 成功指标

### 自我验证（Week 4 结束时）

| 指标 | 目标 |
|------|------|
| 自己愿意用 | 每天在真实 repo 中使用 |
| 替代原流程 | 不再直接跑 agent，而是用 bar run |
| 无 bug 阻塞 | 核心流程无严重 bug |

### 用户验证（Week 6 结束时）

| 指标 | 目标 |
|------|------|
| 试用用户 | 至少 3 个工程师愿意试用 |
| 正向反馈 | 有人说"回滚/审计"是他们用的原因 |
| GitHub star | 100+（可选，不强求） |

---

## 风险与应对

| 风险 | 概率 | 影响 | 应对 |
|------|------|------|------|
| git worktree 兼容性问题 | 中 | 高 | 明确最低 git 版本要求（2.20+） |
| 二进制文件 diff | 低 | 中 | v0 跳过二进制，只处理文本 |
| apply 时冲突 | 中 | 中 | v0 遇到冲突直接报错 |
| 用户不愿意多一层操作 | 中 | 高 | 强调价值，降低使用门槛 |
| Agent 厂商自己做类似功能 | 低 | 高 | 专注差异化（本地 + 轻量 + 通用） |

---

## 里程碑检查点

### Week 2 检查点

- [ ] 能在隔离区跑命令
- [ ] 能看到 diff
- [ ] 有 ledger 记录

**如果未达成**：暂停，分析原因，调整计划

### Week 4 检查点

- [ ] 能 apply 到主分支
- [ ] 能 rollback
- [ ] 自己愿意每天用

**如果未达成**：考虑是否继续

### Week 6 检查点

- [ ] 有 3+ 用户试用
- [ ] 收到正向反馈
- [ ] 文档完善

**如果未达成**：复盘，决定下一步
