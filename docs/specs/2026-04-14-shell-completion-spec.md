# DataKit Shell Completion 规格说明

**日期：** 2026-04-14  
**Issue ID：** issue-id-not-set  
**主题：** 统一采用 `datakit completion` / `datakit completion <bash|powershell|fish|zsh>` 作为补全入口

## 背景

当前 DataKit 存在多套 shell completion 入口：

- `cmd/datakit/cmd/completion.go` 中的 `datakit completion [bash|powershell]`
- `cmd/datakit/cmd/tool/tool.go` 中的 `datakit tool --setup-completer-script` 和 `datakit tool --completer-script`
- `datakit-completer.sh` 与 `internal/cmds/completer.go` 中维护的手写 bash 补全脚本

这会带来几个问题：

- 用户入口分散，不清楚应该使用哪一种方式
- 手写静态脚本容易和 Cobra 当前命令树漂移
- 新增命令和 flag 需要额外维护脚本，长期成本高
- 当前 Cobra completion 只显式支持 `bash` 和 `powershell`，还没有纳入更完整的 shell 范围

本次需求变更后，目标不再是新增 `datakit tool --complete` 包装层，而是直接采用 Cobra 原生命令：

```bash
datakit completion
datakit completion <bash|powershell|fish|zsh>
```

同时该命令不再只是向 `stdout` 输出脚本，而是默认自动识别当前 shell 并安装到本机；用户也可以显式指定 shell 覆盖自动识别结果。安装完成后需要打印安装路径与生效提示。

## 目标

- 将 `datakit completion` / `datakit completion <bash|powershell|fish|zsh>` 作为唯一推荐补全入口
- 将 shell 支持扩展为 `bash`、`powershell`、`fish`、`zsh`
- 默认自动识别当前 shell 并直接安装到本机
- 安装成功后打印“已安装到哪里”以及如何生效
- 允许用户显式指定 shell，覆盖自动识别结果
- 保留输出脚本能力，便于用户审阅或手工安装
- 使用 Cobra 当前命令树生成补全脚本，避免静态脚本漂移
- 明确废弃 `tool` 下的旧 completer flags 和手写补全脚本

## 非目标

- 本次不通过 `tool` 子命令提供 completion 安装器
- 本次不重构与补全无关的子命令行为
- 本次不保留手写补全脚本作为长期事实来源

## 当前状态概览

### Cobra completion

`cmd/datakit/cmd/completion.go` 当前已支持：

- `datakit completion bash`
- `datakit completion powershell`

该命令直接基于 Cobra root command 生成脚本，是最适合作为统一事实来源的实现路径。

### `tool` 下的遗留 completer flag

`cmd/datakit/cmd/tool/tool.go` 当前仍暴露：

- `--setup-completer-script`
- `--completer-script`

这两者依赖 `internal/cmds/completer.go` 的旧逻辑，仅覆盖 Linux/bash，并且脚本内容是静态维护的。

### 手写脚本

`datakit-completer.sh` 代表的是旧方案。随着命令迁移到 Cobra，这份脚本不再适合作为主要维护对象。

## 用户接口

新的唯一推荐入口为：

```bash
datakit completion
datakit completion bash
datakit completion powershell
datakit completion fish
datakit completion zsh
```

默认行为：

- 自动识别当前 shell
- 根据 shell 类型生成对应 completion 脚本
- 自动写入本机标准安装位置
- 成功后打印目标路径和生效提示

### shell 自动识别与显式覆盖

推荐行为：

- `datakit completion`
  - 自动识别当前 shell 并安装
- `datakit completion bash`
  - 忽略自动识别结果，强制按 bash 安装
- `datakit completion fish`
  - 忽略自动识别结果，强制按 fish 安装
- `datakit completion zsh`
  - 忽略自动识别结果，强制按 zsh 安装
- `datakit completion powershell`
  - 忽略自动识别结果，强制按 PowerShell 安装

自动识别建议优先读取：

1. 当前进程环境变量，例如 `SHELL`
2. 平台信息与父进程特征
3. PowerShell/Windows 场景下的宿主信息

如果自动识别失败，命令必须报错，并提示用户显式指定 shell。

### 可选打印模式

允许用户只输出脚本，不执行安装：

```bash
datakit completion --print
datakit completion bash --print
datakit completion powershell --print
datakit completion fish --print
datakit completion zsh --print
```

`--print` 行为：

- 将脚本输出到 `stdout`
- 不写入任何本机文件

### 可选目标路径

允许用户指定安装路径：

```bash
datakit completion --path /tmp/datakit-auto-completion
datakit completion bash --path /tmp/datakit-bash-completion
```

### 覆盖行为

如果目标文件已存在：

- 默认失败并提示目标路径
- 使用 `--force` 时允许覆盖

## CLI 规格

### 命令形式

命令形态为：

```bash
datakit completion
datakit completion <shell>
```

其中 `<shell>` 为可选参数；如果提供，其值必须是以下之一：

- `bash`
- `powershell`
- `fish`
- `zsh`

### Flags

为 `completion` 命令增加以下 flags：

- `--print`
  - 仅输出脚本，不安装
- `--path <file>`
  - 安装到指定路径
- `--force`
  - 允许覆盖已存在目标

### 参数校验

- shell 参数最多只能传一个
- 传入未知 shell 时返回错误
- 未传 shell 参数时进入自动识别流程
- help 文案中必须展示所有受支持 shell
- 如果同时出现自动识别结果和显式 shell 参数，必须以显式参数为准

## 安装规则

### 自动识别规则

自动识别应遵循以下原则：

1. 如果用户显式指定 shell，则以显式参数为准
2. 如果用户未显式指定，则尝试自动识别
3. 如果自动识别结果不明确，则报错并提示用户显式指定 shell

自动识别结果最终只允许落到以下四种 shell 之一：

- `bash`
- `powershell`
- `fish`
- `zsh`

实现要求：

- 自动识别逻辑必须独立封装，便于单元测试
- 自动识别结果应做归一化处理，例如把 `/bin/bash` 归一为 `bash`
- 无法可靠识别时，不允许猜测安装目标

### Bash

默认安装目标优先级：

1. `--path` 指定路径
2. `/usr/share/bash-completion/completions/datakit`
3. `/etc/bash_completion.d/datakit`
4. `~/.local/share/bash-completion/completions/datakit`

行为要求：

- 优先尝试系统路径
- 系统路径不可写时回退到用户目录
- 成功后打印最终安装位置

### Zsh

默认安装目标优先级：

1. `--path` 指定路径
2. `${fpath[1]}/_datakit` 等标准 zsh completion 目录
3. 用户目录下可用的 zsh completion 路径

行为要求：

- 输出文件名应符合 zsh 习惯，例如 `_datakit`
- 成功后打印最终安装位置

### Fish

默认安装目标优先级：

1. `--path` 指定路径
2. `~/.config/fish/completions/datakit.fish`

行为要求：

- 父目录不存在时自动创建
- 成功后打印最终安装位置

### PowerShell

默认安装目标优先级：

1. `--path` 指定路径
2. 当前用户 profile 对应的 completion 文件路径或 profile 脚本路径

行为要求：

- 根据平台和当前用户环境选择合适目标
- 父目录不存在时自动创建
- 必要时避免重复插入 profile 加载片段
- 成功后打印最终安装位置

## 生成策略

所有 completion 脚本必须由 Cobra 当前命令树实时生成。

建议实现方式：

- 在 `cmd/datakit/cmd/completion.go` 或共享 helper 中统一处理 shell 分发
- `bash` 使用 `GenBashCompletionV2`
- `powershell` 使用 `GenPowerShellCompletionWithDesc`
- `fish` 使用 Cobra 对应 fish 生成接口
- `zsh` 使用 Cobra 对应 zsh 生成接口

这样可以保证以下信息只有一份事实来源：

- 根命令
- 子命令
- flags
- 帮助描述

## 安装结果输出

安装成功后必须打印高信号结果信息，至少包括：

- shell 类型
- 实际安装目标路径
- 当前会话如何立即生效
- 如果本次是自动识别得到 shell，应打印识别结果

示例：

```text
completion for bash installed to /usr/share/bash-completion/completions/datakit
reload your shell or run: source /usr/share/bash-completion/completions/datakit
```

如果安装到了用户目录，也应打印对应用户目录路径。

## 兼容与废弃策略

### `datakit completion`

`datakit completion` 将成为正式推荐入口，不再视为过渡命令，也不需要废弃提示。

### 遗留 `tool` completer flags

以下 flag 不再保留兼容，应直接移除：

- `--setup-completer-script`
- `--completer-script`

预期行为：

- 从代码中删除对应入口
- 同步删除相关测试断言与文档示例
- 通过 breaking change changelog 明确通知用户改用 `datakit completion`

### 手写脚本与旧 completer 逻辑

以下内容不应继续作为主维护路径：

- `datakit-completer.sh`
- `internal/cmds/completer.go` 中静态脚本相关逻辑

处理策略：

- 在实现阶段直接删除
- 同步清理所有引用
- 通过 breaking change changelog 明确通知用户新的补全入口

## 错误处理

以下场景必须返回清晰错误：

- 未传 shell 参数
- 传入多个参数
- 传入不支持的 shell
- Cobra 生成脚本失败
- 无可写安装目标
- 目标已存在但未传 `--force`

错误信息应明确告诉用户：

- 可用 shell 列表
- 当前失败的目标路径
- 自动识别失败还是安装失败
- 是否可以改用 `--path` 或 `--print`

## 测试要求

### 单元测试

需要覆盖：

- `completion` 命令的 `ValidArgs`
- 参数个数校验
- shell 自动识别逻辑
- 显式 shell 覆盖自动识别
- 四种 shell 的分发逻辑
- 未知 shell 错误
- 各 shell 的安装路径选择
- `--print` 模式
- `--force` 覆盖逻辑
- 安装成功提示内容

### 命令级测试

需要验证：

- `datakit completion` 能自动识别 shell 并安装到正确路径
- `datakit completion bash` 会安装到正确路径或可写回退路径
- `datakit completion powershell` 会安装到正确路径
- `datakit completion fish` 会安装到正确路径
- `datakit completion zsh` 会安装到正确路径
- `datakit completion <shell> --print` 会输出脚本且不写文件

### 回归重点

需要确保生成结果中包含当前已迁移到 Cobra 的代表性命令，例如：

- `dql`
- `pipeline`
- `service`
- `monitor`
- `install`
- `debug`
- `tool`
- `check`
- `version`
- `import`
- `run`

## 文档更新要求

需要更新用户文档，至少说明：

- 标准入口为 `datakit completion` 和 `datakit completion <shell>`
- 当前支持 `bash`、`powershell`、`fish`、`zsh`
- 默认会自动识别当前 shell 并安装到本机
- 用户可以显式指定 shell 覆盖自动识别
- 安装完成后会打印目标路径
- `--print` 可用于只输出脚本
- 旧 completer flags 与静态脚本已被移除
- 需要在 breaking change changelog 中说明迁移方式

不得将任何 `internal/export` 下的现有手册内容迁移到 `docs`。

## 验收标准

- `datakit completion` / `datakit completion <bash|powershell|fish|zsh>` 成为唯一推荐入口
- 四种 shell 均支持
- 默认行为是自动识别当前 shell 并安装到本机，而不是只输出到 stdout
- 用户可以显式指定 shell 覆盖自动识别
- 安装成功后会打印实际安装位置
- `--print` 可输出脚本且不写文件
- 生成内容来自 Cobra，而不是手写静态脚本
- 遗留 completer 入口和静态脚本被直接移除
- breaking change changelog 包含迁移说明
- 自动化测试覆盖四种 shell、安装路径和参数校验
