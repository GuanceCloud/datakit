# DataKit CLI 全量迁移到 Cobra 实施计划

> **面向执行代理：** 本计划是当前分支的主计划。执行时以本文件为进度事实来源，使用复选框 `- [ ]` 跟踪进度。

**目标：** 将 `datakit` CLI 的主入口、帮助系统和顶层命令分发彻底收口到 Cobra，保留主要命令能力，统一为 `--help` 风格，并对明显适合的命令做子命令化规范。

**架构：** `cmd/datakit/cmd/root.go` 作为唯一 CLI 根入口；`cmd/datakit/main.go` 只负责设置构建信息并执行 Cobra；旧 `internal/cmds/parse_flags.go` 不再承担主分发职责；`internal/cmds` 中保留可复用执行逻辑；首批对子命令化价值最高的 `service` 做结构规范。

**技术栈：** Go、Cobra、testify、现有 `cmd/datakit/cmd/*` 命令树、`internal/cmds` 执行逻辑

---

## 文件结构

### 重点修改文件

- `cmd/datakit/main.go`
  - 删除旧入口分流，只执行 Cobra
- `cmd/datakit/cmd/root.go`
  - 统一根命令、默认启动逻辑、构建信息注入
- `cmd/datakit/core/core.go`
  - 抽出可被 Cobra 调用的启动入口，不再依赖旧 ParseFlags
- `cmd/datakit/cmd/run/run.go`
  - 让 `run` 成为真实执行路径
- `cmd/datakit/cmd/service/service.go`
  - 从历史 flags 迁移为子命令
- `cmd/datakit/cmd/importcmd/import.go`
  - 补齐真实导入逻辑
- `internal/cmds/import.go`
  - 导出真实导入执行入口
- `internal/cmds/parse_flags.go`
  - 删除或退役旧主分发逻辑
- `internal/export/doc/zh/*`
  - 更新旧 `help` 和旧命令形态文档
- `internal/export/doc/en/*`
  - 更新旧 `help` 和旧命令形态文档
- `dockerfiles/*`
  - 更新依赖旧 CLI 形态的调用

### 受影响测试

- `cmd/datakit/cmd/root_test.go`
- `cmd/datakit/cmd/run/run_test.go`
- `cmd/datakit/cmd/service/service_test.go`
- `cmd/datakit/cmd/importcmd/import_test.go`
- 其他命令包测试

### 文档文件

- `docs/specs/2026-04-20-datakit-cli-cobra-migration-spec.md`
- `docs/plans/2026-04-20-datakit-cli-cobra-migration.md`

## 阶段 1：入口切换到 Cobra

**涉及文件：**
- 修改：`cmd/datakit/main.go`
- 修改：`cmd/datakit/cmd/root.go`
- 修改：`cmd/datakit/core/core.go`
- 修改：`cmd/datakit/cmd/run/run.go`

- [x] **Step 1: 移除 `main.go` 中的命令分流**

实现目标：

- 删除 `shouldUseCobraCommand()` 之类的特判逻辑
- `main.go` 只设置构建信息并执行 Cobra

- [x] **Step 2: 将构建信息注入从旧入口迁到 Cobra 根命令**

实现目标：

- `ReleaseVersion`
- `InputsReleaseType`
- `Lite`
- `ELinker`

都能在 Cobra 路径下被正确消费

- [x] **Step 3: 提供可由 Cobra 调用的主启动函数**

实现目标：

- `core` 层提供显式启动入口
- 启动过程不再依赖 `cmds.ParseFlags()`
- 容器模式通过参数显式传递，而不是依赖旧全局 flag 分发

- [x] **Step 4: 根命令接管默认启动行为**

实现目标：

- `datakit` 裸执行时，仍然启动主程序
- 默认行为不能退化成只打印帮助

- [x] **Step 5: `run` 命令接入真实执行路径**

实现目标：

- `datakit run` 与默认启动走同一主路径
- `datakit run --container` 继续生效

- [x] **Step 6: 禁用 Cobra 默认 `help` 子命令**

实现目标：

- 只保留 `--help` 形态
- `datakit help` / `datakit help <command>` 不再作为支持接口

## 阶段 2：命令树规范化

**涉及文件：**
- 修改：`cmd/datakit/cmd/service/service.go`
- 修改：`cmd/datakit/cmd/service/service_test.go`
- 按需修改：`cmd/datakit/cmd/tool/tool.go`
- 按需修改：`cmd/datakit/cmd/debug/debug.go`
- 按需修改：`cmd/datakit/cmd/check/check.go`

- [x] **Step 1: 先完成 `service` 子命令化**

目标命令：

```bash
datakit service start
datakit service stop
datakit service restart
datakit service uninstall
datakit service reinstall
```

- [x] **Step 2: 删除 `service` 根命令上的历史互斥动作 flags**

实现目标：

- 不再暴露 `--start` / `--stop` / `--restart` / `--uninstall` / `--reinstall`
- 保留与动作无关的公共 flags

- [x] **Step 3: 盘点其他顶层命令的子命令化机会**

盘点对象：

- `tool`
- `debug`
- `check`

预期结果：

- 明确哪些命令本次直接拆
- 哪些命令暂时先保留 flags 形态，但已完成 Cobra 收口

## 阶段 3：补齐不完整 Cobra 命令

**涉及文件：**
- 修改：`cmd/datakit/cmd/importcmd/import.go`
- 修改：`internal/cmds/import.go`
- 修改：相关测试

- [x] **Step 1: 修复 `import` 占位实现**

实现目标：

- `import` 直接调用真实导入逻辑
- 不再错误调用其他命令逻辑

- [x] **Step 2: 逐个审计现有 Cobra 命令是否仍依赖旧占位行为**

目标命令：

- `import`
- `run`
- `service`
- 其他顶层命令

- [x] **Step 3: 确认 `import`、`run`、`service` 已不再通过旧全局 flag 回填执行**

预期结果：

- 这三个命令都直接调用显式执行函数
- 不再依赖 `parse_flags.go` 导出的别名 flag 指针

## 阶段 4：逐命令去除旧全局 flag 桥接

**涉及文件：**
- 修改：`cmd/datakit/cmd/dql/dql.go`
- 修改：`cmd/datakit/cmd/pipeline/pipeline.go`
- 修改：`cmd/datakit/cmd/monitor/monitor.go`
- 修改：`cmd/datakit/cmd/install/install.go`
- 修改：`cmd/datakit/cmd/check/check.go`
- 修改：`cmd/datakit/cmd/debug/debug.go`
- 修改：`cmd/datakit/cmd/tool/tool.go`
- 修改：`cmd/datakit/cmd/version/version.go`
- 按需修改：`internal/cmds/*`

- [x] **Step 1: 盘点所有仍通过 `cmds.FlagXxx` 回填的 Cobra 命令**

输出要求：

- 列出命令清单
- 标明每个命令当前依赖的旧全局变量和旧执行函数

- [x] **Step 2: 为每个命令定义最终完成态**

完成态要求：

- Cobra 直接调用显式执行函数
- 不再依赖 `parse_flags.go` 导出的别名 flag 指针
- 不接受“暂时保留全局 flag 回填”作为收尾状态

- [x] **Step 3: 逐命令迁移**

优先顺序：

1. `version` 已完成
2. `install` 已完成
3. `check` 已完成
4. `debug` 已完成
5. `monitor` 已完成
6. `dql` 已完成
7. `pipeline` 已完成
8. `tool` 已完成

- [x] **Step 4: 删除不再需要的旧导出 flag 别名**

实现目标：

- `internal/cmds/parse_flags.go` 中仅为 Cobra 包装服务的导出 flag 别名清理完成

## 阶段 5：移除旧 CLI 主分发

**涉及文件：**
- 修改：`internal/cmds/parse_flags.go`
- 按需修改：`cmd/datakit/core/core.go`
- 按需修改：其他依赖文件

- [x] **Step 1: 删除旧 help 分发路径**

实现目标：

- 移除 `printHelp`
- 移除 `runHelpFlags`
- 移除 `doParseAndRunFlags`
- 不再支持 `datakit help xxx`

- [x] **Step 2: 删除旧 ParseFlags 主入口调用**

实现目标：

- 主路径上不再调用 `cmds.ParseFlags()`
- CLI 行为完全由 Cobra 控制

- [x] **Step 3: 清理仅服务于旧 CLI 骨架的导出 flag 别名**

实现目标：

- 删除所有不再被 Cobra 命令使用的旧导出 flag 别名
- `parse_flags.go` 不再作为任何 Cobra 命令的依赖源

## 阶段 6：帮助与行为验证

**涉及文件：**
- 修改：相关测试

- [x] **Step 1: 补充根命令和帮助行为测试**

覆盖：

- `datakit --help`
- `datakit completion --help`
- `datakit service --help`
- `datakit help`
- `datakit help service`

- [x] **Step 2: 补充 `service` 子命令结构测试**

覆盖：

- `start`
- `stop`
- `restart`
- `uninstall`
- `reinstall`

- [x] **Step 3: 补充 `run` 与默认启动路径测试**

覆盖：

- 根命令默认执行路径
- `run --container` 路径

- [x] **Step 4: 运行聚焦测试**

执行：

```bash
go test ./cmd/datakit/cmd/... ./cmd/datakit/core ./internal/cmds/...
```

预期：命令树、帮助行为和关键执行路径相关测试通过

## 阶段 7：文档与打包同步

**涉及文件：**
- 修改：`internal/export/doc/zh/*`
- 修改：`internal/export/doc/en/*`
- 修改：`dockerfiles/*`

- [x] **Step 1: 搜索所有旧 CLI 形态引用**

执行：

```bash
rg -n "datakit help|--start|--stop|--restart|--uninstall|--reinstall" internal/export/doc dockerfiles
```

- [x] **Step 2: 更新用户文档**

实现目标：

- 所有文档示例统一为 `--help`
- `service` 示例统一为子命令形态

- [x] **Step 3: 更新 Dockerfile 和构建脚本**

实现目标：

- 凡是依赖旧 CLI 形态的地方都改成新命令形态

## 阶段 8：收尾

- [x] **Step 1: 更新本计划中的完成状态**

- [x] **Step 2: 确认旧 shell completion 计划的地位**

预期结果：

- 旧 `docs/specs/2026-04-14-shell-completion-spec.md`
- 旧 `docs/plans/2026-04-14-shell-completion.md`

保留为历史背景，但不再作为当前总目标

- [x] **Step 3: 形成最终变更说明**

至少包含：

- 入口已全量切换到 Cobra
- `help` 形态变更为 `--help`
- 默认 `help` 子命令已禁用
- `service` 已子命令化
- `import` 已接入真实执行逻辑
