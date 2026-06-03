# DataKit CLI 全量迁移到 Cobra 规格说明

**日期：** 2026-04-20  
**Issue ID：** issue-id-not-set  
**主题：** 将 `datakit` CLI 全量迁移到 Cobra，并统一为 Cobra 风格

## 背景

当前 `datakit` CLI 处于双轨状态：

- `cmd/datakit/main.go` 仍以旧入口为主，只对 `completion` 做 Cobra 特判
- `internal/cmds/parse_flags.go` 仍承担顶层命令分发、帮助输出和参数解析
- `cmd/datakit/cmd/*` 已存在一批 Cobra 命令，但多数只是外层包装，内部仍通过全局 flag 回填再调用旧执行逻辑

这会带来几个问题：

- 帮助系统分裂，`completion` 与其他命令行为不一致
- 顶层命令发现不统一，`datakit --help` 无法作为唯一事实来源
- 旧 `help`、旧 `FlagSet`、新 Cobra 命令树同时存在，维护成本高
- 某些命令的 Cobra 包装并不完整，例如 `import` 仍是占位实现
- 明显适合子命令化的命令仍维持历史 flag 驱动形式，难以形成清晰的命令树

本次变更的目标不是继续在旧框架外补一个 Cobra 外壳，而是把 `datakit` 的 CLI 主入口、帮助系统和命令树彻底收口到 Cobra。

## 目标

- `datakit` 主入口永久切换为 Cobra
- 所有顶层命令都从同一棵 Cobra 命令树进入
- 统一帮助入口为 `--help` 形态
- 不再兼容 `datakit help xxx`
- 顶层命令保留，但允许顺手做命令结构规范化
- 明显适合子命令化的命令改造成子命令模型
- 删除旧 `parse_flags` 的主分发职责
- 保留旧业务逻辑，但不保留旧 CLI 框架

## 非目标

- 本次不要求所有内部执行逻辑都迁出 `internal/cmds`
- 本次不追求保留旧帮助文案、旧错误文案、旧 usage 排版
- 本次不承诺保留所有历史 flag 形态
- 本次不处理 `cmd/upgrader` 的 CLI 体系

## 用户接口

迁移完成后，CLI 形态统一为：

```bash
datakit
datakit --help
datakit completion --help
datakit service --help
datakit service start
datakit service stop
datakit run --container
```

### 顶层命令集合

本次迁移保留以下顶层命令能力：

- `completion`
- `import`
- `run`
- `check`
- `debug`
- `dql`
- `pipeline`
- `version`
- `service`
- `monitor`
- `install`
- `tool`

### 帮助入口

统一帮助入口为：

```bash
datakit --help
datakit <command> --help
datakit <command> <subcommand> --help
```

不再支持，也不再保留 Cobra 默认 `help` 子命令：

```bash
datakit help
datakit help <command>
```

### 默认启动行为

裸执行：

```bash
datakit
```

仍然必须保持“启动 DataKit 主程序”的语义，而不是仅输出帮助。

### 命令结构规范化

对明显适合做层级化的命令，允许顺手规范化为子命令模型。首批明确规范化的命令为：

```bash
datakit service start
datakit service stop
datakit service restart
datakit service uninstall
datakit service reinstall
```

## 架构约束

### 根入口

`cmd/datakit/main.go` 必须直接执行 Cobra root command，不再进行命令分流。

### Root Command

`cmd/datakit/cmd/root.go` 作为唯一根命令，负责：

- 注册所有顶层命令
- 统一默认启动逻辑
- 统一帮助、错误和 usage 风格
- 注入构建时版本信息
- 禁用或隐藏默认 `help` 子命令，确保帮助入口只保留 `--help`

### 命令边界

命令层职责划分如下：

- `cmd/datakit/cmd/*`
  - Cobra 命令定义
  - 参数声明
  - 参数校验
  - 子命令组织
  - 调用执行层
- `internal/cmds`
  - 保留有业务价值的执行逻辑
  - 不再承担 CLI 主分发、help 体系、顶层 `FlagSet` 框架职责

### 旧框架处理原则

以下内容属于旧 CLI 骨架，应退出主路径：

- `internal/cmds/parse_flags.go` 中的 `printHelp`
- `internal/cmds/parse_flags.go` 中的 `runHelpFlags`
- `internal/cmds/parse_flags.go` 中的 `doParseAndRunFlags`
- `cmd/datakit/main.go` 中的 Cobra/旧入口分流逻辑

### 业务逻辑复用原则

允许 Cobra 命令继续复用已有执行逻辑，但要求：

- 不再依赖旧的顶层分发入口
- 不再通过旧 help/usage 体系暴露给用户
- 优先收敛成明确的可调用函数，而不是继续堆叠全局状态和占位包装
- 不能把“通过 Cobra 命令设置 `internal/cmds` 全局 flag 别名再调用旧函数”视为最终完成态；这类桥接只能作为临时中间态

## 兼容性策略

本次迁移采取“能力兼容，交互规范化”的策略：

- 保留主要命令能力
- 保留主要命令名
- 不保留旧帮助格式
- 不保留 `datakit help xxx`
- 允许错误文案和 usage 结构变化
- 允许把历史 flags 改造成子命令

## 关键行为要求

### 默认运行

- `datakit` 默认启动主程序
- `datakit run` 作为显式运行入口，应与默认运行保持一致的主执行路径
- `datakit run --container` 应继续支持容器模式启动

### Service 命令

- `service` 必须子命令化
- 根 `service` 命令不再承载 `--start`、`--stop`、`--restart`、`--uninstall`、`--reinstall` 这类历史互斥 flag 入口
- `service` 根命令仍可保留通用 flags，例如 `--log`

### Import 命令

- `import` 的 Cobra 实现必须接入真实导入逻辑
- 不允许继续保留 placeholder 行为

### Completion 命令

- `completion` 仍保留并继续作为 Cobra 原生命令体系的一部分
- `completion` 必须出现在根 help 中
- `completion --help` 与其他命令帮助风格保持一致

### 文档与打包

- 所有关联文档都要更新为新 CLI 形态，至少包括 `internal/export/doc/zh/*` 与 `internal/export/doc/en/*` 中引用旧 `help` 或旧命令形态的内容
- `dockerfiles` 下凡是直接调用 `datakit` 命令且依赖旧 CLI 形态的文件，都必须同步更新

## 验收标准

迁移完成后，至少满足：

- `datakit --help` 展示全部顶层命令
- `datakit completion --help` 正常工作
- `datakit service --help` 正常工作
- `datakit service start` 等子命令存在
- `datakit` 默认启动 DataKit 主程序
- `datakit run --container` 正常进入运行路径
- `import` 不再是占位实现
- `datakit help` / `datakit help <command>` 不再作为支持接口存在
- 代码主路径上不再调用 `internal/cmds.ParseFlags()`
- `cmd/datakit/main.go` 中不再保留命令分流逻辑
- 所有关联用户文档与 Dockerfile 中的旧 CLI 用法已更新

## 替代关系

本规格说明替代“仅围绕 shell completion 做局部改造”的方向。  
之前的 `docs/specs/2026-04-14-shell-completion-spec.md` 仍可作为 completion 细节参考，但不再代表当前分支的总体目标。
